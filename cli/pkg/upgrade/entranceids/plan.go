// Package entranceids implements the one-time upgrade, never runtime backfill.
package entranceids

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	app "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

const Migration = "entrance-ids-v1"

type Change struct {
	Name   string              `json:"name"`
	UID    types.UID           `json:"uid"`
	Before app.ApplicationSpec `json:"before"`
	After  app.ApplicationSpec `json:"after"`
}
type Plan struct {
	Migration    string    `json:"migration"`
	ClusterUID   types.UID `json:"clusterUID"`
	OriginalUser string    `json:"originalUser"`
	Changes      []Change  `json:"changes"`
	Cleared      []string  `json:"cleared"`
}

// Build takes one snapshot. A retry must load this saved plan, never rebuild
// conflict groups after some of their members have already been cleared.
func Build(apps []app.Application, original string, clusterUID types.UID) (*Plan, error) {
	if original == "" || clusterUID == "" {
		return nil, fmt.Errorf("original user and cluster identity are required")
	}
	p := &Plan{Migration: Migration, OriginalUser: original, ClusterUID: clusterUID}
	sort.Slice(apps, func(i, j int) bool { return apps[i].Name < apps[j].Name })
	after := make([]app.Application, 0, len(apps))
	for i := range apps {
		a := apps[i].DeepCopy()
		// Only the upgrade can treat missing saved IDs as unallocated. Existing
		// IDs remain anchors; normal assignment rejects such incomplete records.
		prior := a.DeepCopy()
		prior.Spec.Entrances = withIDs(prior.Spec.Entrances)
		prior.Spec.SharedEntrances = withIDs(prior.Spec.SharedEntrances)
		filterAdded := func(s map[string]string) error {
			es, err := app.ParseAddedEntrances(s)
			if err != nil {
				return err
			}
			if s["addedEntrances"] != "" {
				b, err := json.Marshal(withIDs(es))
				if err != nil {
					return err
				}
				s["addedEntrances"] = string(b)
			}
			return nil
		}
		if err := filterAdded(prior.Spec.Settings); err != nil {
			return nil, err
		}
		for _, s := range prior.Spec.UserSettings {
			if err := filterAdded(s); err != nil {
				return nil, err
			}
		}
		assigned, err := app.AssignApplicationEntranceIDs(a, prior, original)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", a.Name, err)
		}
		after = append(after, *assigned)
	}
	type ref struct {
		appIndex       int
		user, entrance string
		settings       map[string]string
		domains        map[string]map[string]json.RawMessage
	}
	groups := map[string][]ref{}
	for i := range after {
		a := &after[i]
		scan := func(user string, s map[string]string) error {
			if s["customDomain"] == "" {
				return nil
			}
			var domains map[string]map[string]json.RawMessage
			if err := json.Unmarshal([]byte(s["customDomain"]), &domains); err != nil {
				return fmt.Errorf("%s invalid customDomain: %w", a.Name, err)
			}
			for name, fields := range domains {
				var prefix string
				if raw := fields["third_level_domain"]; len(raw) > 0 {
					if err := json.Unmarshal(raw, &prefix); err != nil {
						return err
					}
				}
				prefix = strings.ToLower(strings.TrimSpace(prefix))
				if prefix != "" {
					groups[prefix] = append(groups[prefix], ref{i, user, name, s, domains})
				}
			}
			return nil
		}
		if err := scan(a.Spec.Owner, a.Spec.Settings); err != nil {
			return nil, err
		}
		for user, s := range a.Spec.UserSettings {
			if err := scan(user, s); err != nil {
				return nil, err
			}
		}
	}
	for prefix, refs := range groups {
		if len(refs) < 2 {
			continue
		}
		// Same user's default and overlay for the same entrance represent one
		// logical claim. Multiple original-user entrances cannot be auto-selected.
		ownerClaims := map[string]bool{}
		for _, r := range refs {
			if r.user == original {
				ownerClaims[after[r.appIndex].Name+"/"+r.entrance] = true
			}
		}
		if len(ownerClaims) > 1 {
			return nil, fmt.Errorf("multiple original-user entrances claim custom ID %q", prefix)
		}
		logical := map[string]bool{}
		for _, r := range refs {
			logical[after[r.appIndex].Name+"/"+r.user+"/"+r.entrance] = true
		}
		if len(logical) < 2 {
			continue
		}
		for _, r := range refs {
			if r.user == original {
				continue
			}
			r.domains[r.entrance]["third_level_domain"] = json.RawMessage(`""`)
			b, err := json.Marshal(r.domains)
			if err != nil {
				return nil, err
			}
			r.settings["customDomain"] = string(b)
			p.Cleared = append(p.Cleared, after[r.appIndex].Name+"/"+r.user+"/"+r.entrance+": "+prefix)
		}
	}
	// Only ordinary persisted IDs and Settings claims are checked; YAML aliases
	// are explicitly outside the automatic upgrade conflict policy.
	for i := range after {
		if err := app.ValidateOrdinaryIDClaims(&after[i], after); err != nil {
			return nil, err
		}
	}
	for i := range apps {
		p.Changes = append(p.Changes, Change{Name: apps[i].Name, UID: apps[i].UID, Before: apps[i].Spec, After: after[i].Spec})
	}
	sort.Strings(p.Cleared)
	return p, nil
}
func withIDs(es []app.Entrance) []app.Entrance {
	var out []app.Entrance
	for _, e := range es {
		if e.ID != "" {
			out = append(out, e)
		}
	}
	return out
}
