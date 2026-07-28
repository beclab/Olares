package doctor

import (
	"encoding/json"
	"sort"
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// reservedThirdLevelDomains mirrors app-service handler_settings.go.
var reservedThirdLevelDomains = map[string]struct{}{
	"auth":    {},
	"desktop": {},
	"wizard":  {},
}

// DomainIssue is one third-level-domain problem in a user zone.
type DomainIssue struct {
	User       string `json:"user"`
	App        string `json:"app"`
	Entrance   string `json:"entrance"`
	ThirdLevel string `json:"third_level"`
	Issue      string `json:"issue"` // "duplicate" | "reserved"
}

type thirdLevelClaim struct {
	user       string
	objectName string
	app        string
	entrance   string
	thirdLevel string
	shared     bool
}

// ClearThirdLevelOp clears one entrance's third_level_domain.
// Reason is "duplicate" (loser of a keep-one conflict) or "reserved"
// (auth/desktop/wizard — always cleared, nothing is kept).
type ClearThirdLevelOp struct {
	ObjectName   string `json:"object_name"`
	App          string `json:"app"`
	User         string `json:"user"`
	Entrance     string `json:"entrance"`
	ThirdLevel   string `json:"third_level"`
	KeepApp      string `json:"keep_app,omitempty"`
	KeepEntrance string `json:"keep_entrance,omitempty"`
	Shared       bool   `json:"shared"`
	Reason       string `json:"reason"` // "duplicate" | "reserved"
}

func collectThirdLevelClaims(apps []appv1alpha1.Application, users []string) []thirdLevelClaim {
	claims := make([]thirdLevelClaim, 0)
	for _, user := range users {
		user = strings.TrimSpace(user)
		if user == "" {
			continue
		}
		for i := range apps {
			app := &apps[i]
			blob := callerCustomDomainBlob(app, user)
			for entrance, cfg := range parseCustomDomain(blob) {
				tl := strings.TrimSpace(cfg.thirdLevel)
				if tl == "" {
					continue
				}
				claims = append(claims, thirdLevelClaim{
					user:       user,
					objectName: app.Name,
					app:        app.Spec.Name,
					entrance:   entrance,
					thirdLevel: tl,
					shared:     appv1alpha1.IsShared(app),
				})
			}
		}
	}
	return claims
}

// FindThirdLevelDomainIssues audits customDomain.third_level_domain values
// per user zone (Approach 1): duplicates within a zone, and reserved names
// auth/desktop/wizard. users is the set of zone owners to scan (typically
// IAM usernames); empty users yields no issues.
func FindThirdLevelDomainIssues(apps []appv1alpha1.Application, users []string) []DomainIssue {
	claims := collectThirdLevelClaims(apps, users)

	type key struct {
		user   string
		prefix string
	}
	counts := make(map[key]int)
	for _, c := range claims {
		counts[key{user: c.user, prefix: strings.ToLower(c.thirdLevel)}]++
	}

	issues := make([]DomainIssue, 0)
	seen := make(map[string]struct{})
	for _, c := range claims {
		lower := strings.ToLower(c.thirdLevel)
		if _, ok := reservedThirdLevelDomains[lower]; ok {
			iss := DomainIssue{
				User: c.user, App: c.app, Entrance: c.entrance,
				ThirdLevel: c.thirdLevel, Issue: "reserved",
			}
			id := issueID(iss)
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				issues = append(issues, iss)
			}
		}
		if counts[key{user: c.user, prefix: lower}] > 1 {
			iss := DomainIssue{
				User: c.user, App: c.app, Entrance: c.entrance,
				ThirdLevel: c.thirdLevel, Issue: "duplicate",
			}
			id := issueID(iss)
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				issues = append(issues, iss)
			}
		}
	}

	sort.Slice(issues, func(i, j int) bool {
		a, b := issues[i], issues[j]
		if a.User != b.User {
			return a.User < b.User
		}
		if a.App != b.App {
			return a.App < b.App
		}
		if a.Entrance != b.Entrance {
			return a.Entrance < b.Entrance
		}
		return a.Issue < b.Issue
	})
	return issues
}

// PlanForceDedupeClears returns clear ops for --force-dedupe:
//   - reserved (auth/desktop/wizard): clear every matching claim
//   - duplicate: keep the lexicographically first (app, entrance) in each
//     user-zone prefix group; clear the rest
func PlanForceDedupeClears(apps []appv1alpha1.Application, users []string) []ClearThirdLevelOp {
	claims := collectThirdLevelClaims(apps, users)
	ops := make([]ClearThirdLevelOp, 0)
	seen := make(map[string]struct{})
	add := func(op ClearThirdLevelOp) {
		id := strings.Join([]string{op.ObjectName, op.User, op.Entrance, op.Reason}, "\x00")
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ops = append(ops, op)
	}

	for _, c := range claims {
		if _, ok := reservedThirdLevelDomains[strings.ToLower(c.thirdLevel)]; !ok {
			continue
		}
		add(ClearThirdLevelOp{
			ObjectName: c.objectName,
			App:        c.app,
			User:       c.user,
			Entrance:   c.entrance,
			ThirdLevel: c.thirdLevel,
			Shared:     c.shared,
			Reason:     "reserved",
		})
	}

	type key struct {
		user   string
		prefix string
	}
	groups := make(map[key][]thirdLevelClaim)
	for _, c := range claims {
		k := key{user: c.user, prefix: strings.ToLower(c.thirdLevel)}
		groups[k] = append(groups[k], c)
	}

	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			if group[i].app != group[j].app {
				return group[i].app < group[j].app
			}
			return group[i].entrance < group[j].entrance
		})
		keep := group[0]
		// If the kept claim is reserved it is already queued for clear;
		// clear every claim in the group (including "keep") via reserved.
		if _, reservedKeep := reservedThirdLevelDomains[strings.ToLower(keep.thirdLevel)]; reservedKeep {
			continue
		}
		for _, c := range group[1:] {
			if _, reserved := reservedThirdLevelDomains[strings.ToLower(c.thirdLevel)]; reserved {
				continue // already covered by reserved clears
			}
			add(ClearThirdLevelOp{
				ObjectName:   c.objectName,
				App:          c.app,
				User:         c.user,
				Entrance:     c.entrance,
				ThirdLevel:   c.thirdLevel,
				KeepApp:      keep.app,
				KeepEntrance: keep.entrance,
				Shared:       c.shared,
				Reason:       "duplicate",
			})
		}
	}
	sort.Slice(ops, func(i, j int) bool {
		a, b := ops[i], ops[j]
		if a.User != b.User {
			return a.User < b.User
		}
		if a.App != b.App {
			return a.App < b.App
		}
		if a.Entrance != b.Entrance {
			return a.Entrance < b.Entrance
		}
		return a.Reason < b.Reason
	})
	return ops
}

// ApplyClearOpsToApp clears third_level_domain for ops targeting this app.
// Shared apps write Spec.UserSettings[user]; per-user apps write Spec.Settings.
// Shared clears always land in the per-user overlay (creating one from the
// effective blob when needed) so global Spec.Settings is not rewritten.
func ApplyClearOpsToApp(app *appv1alpha1.Application, ops []ClearThirdLevelOp) error {
	if app == nil || len(ops) == 0 {
		return nil
	}
	byUser := make(map[string][]ClearThirdLevelOp)
	var settingsOps []ClearThirdLevelOp
	for _, op := range ops {
		if op.ObjectName != "" && op.ObjectName != app.Name {
			continue
		}
		if op.App != "" && op.App != app.Spec.Name {
			continue
		}
		if appv1alpha1.IsShared(app) {
			byUser[op.User] = append(byUser[op.User], op)
		} else {
			settingsOps = append(settingsOps, op)
		}
	}

	for user, userOps := range byUser {
		entrances := make([]string, 0, len(userOps))
		for _, op := range userOps {
			entrances = append(entrances, op.Entrance)
		}
		// Source the base from the exact blob the audit reads
		// (callerCustomDomainBlob → EffectiveSettings(user) for shared),
		// so the cleared overlay matches what was scanned and what
		// app-service's domain handler resolves. Writing it back into
		// UserSettings[user] promotes an inherited global blob into the
		// per-user overlay, leaving global Spec.Settings intact.
		base := callerCustomDomainBlob(app, user)
		updated, err := clearThirdLevelInBlob(base, entrances)
		if err != nil {
			return err
		}
		if app.Spec.UserSettings == nil {
			app.Spec.UserSettings = make(map[string]map[string]string)
		}
		if app.Spec.UserSettings[user] == nil {
			app.Spec.UserSettings[user] = make(map[string]string)
		}
		app.Spec.UserSettings[user]["customDomain"] = updated
	}

	if len(settingsOps) > 0 {
		entrances := make([]string, 0, len(settingsOps))
		for _, op := range settingsOps {
			entrances = append(entrances, op.Entrance)
		}
		base := ""
		if app.Spec.Settings != nil {
			base = app.Spec.Settings["customDomain"]
		}
		updated, err := clearThirdLevelInBlob(base, entrances)
		if err != nil {
			return err
		}
		if app.Spec.Settings == nil {
			app.Spec.Settings = make(map[string]string)
		}
		app.Spec.Settings["customDomain"] = updated
	}
	return nil
}

func clearThirdLevelInBlob(blob string, entrances []string) (string, error) {
	raw := make(map[string]map[string]interface{})
	if blob != "" {
		if err := json.Unmarshal([]byte(blob), &raw); err != nil {
			return "", err
		}
	}
	for _, entrance := range entrances {
		cfg := raw[entrance]
		if cfg == nil {
			cfg = make(map[string]interface{})
			raw[entrance] = cfg
		}
		cfg["third_level_domain"] = ""
		if _, ok := cfg["third_party_domain"]; !ok {
			cfg["third_party_domain"] = ""
		}
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func issueID(i DomainIssue) string {
	return strings.Join([]string{i.User, i.App, i.Entrance, i.ThirdLevel, i.Issue}, "\x00")
}

type entranceCustomDomain struct {
	thirdLevel string
	thirdParty string
}

func parseCustomDomain(blob string) map[string]entranceCustomDomain {
	out := make(map[string]entranceCustomDomain)
	if blob == "" {
		return out
	}
	var raw map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(blob), &raw); err != nil {
		return out
	}
	for entrance, cfg := range raw {
		tl, _ := cfg["third_level_domain"].(string)
		tp, _ := cfg["third_party_domain"].(string)
		out[entrance] = entranceCustomDomain{thirdLevel: tl, thirdParty: tp}
	}
	return out
}

// callerCustomDomainBlob mirrors app-service handler_settings.go.
func callerCustomDomainBlob(app *appv1alpha1.Application, caller string) string {
	if appv1alpha1.IsShared(app) {
		return app.EffectiveSettings(caller)["customDomain"]
	}
	if app.Spec.Owner == caller {
		return app.Spec.Settings["customDomain"]
	}
	return ""
}

// CollectZoneUsers returns unique usernames to scan: explicit users plus
// every Application owner and UserSettings key (covers overlays without
// relying solely on the IAM list).
func CollectZoneUsers(apps []appv1alpha1.Application, explicit []string) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	for _, u := range explicit {
		add(u)
	}
	for i := range apps {
		add(apps[i].Spec.Owner)
		for u := range apps[i].Spec.UserSettings {
			add(u)
		}
	}
	sort.Strings(out)
	return out
}
