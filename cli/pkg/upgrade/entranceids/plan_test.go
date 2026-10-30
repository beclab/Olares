package entranceids

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	app "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sys "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	fakeapi "github.com/beclab/api/pkg/generated/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakekube "k8s.io/client-go/kubernetes/fake"
	kt "k8s.io/client-go/testing"
)

func legacy(name, user, prefix string) app.Application {
	return app.Application{ObjectMeta: metav1.ObjectMeta{Name: name, UID: types.UID(name)}, Spec: app.ApplicationSpec{Appid: "abcd1234", Owner: user, Entrances: []app.Entrance{{Name: "web"}}, SharedEntrances: []app.Entrance{{Name: "shared"}}, Settings: map[string]string{"customDomain": fmt.Sprintf(`{"web":{"third_level_domain":%q,"third_party_domain":"external.example","cert":"keep"}}`, prefix), "defaultThirdLevelDomainConfig": `[{"entranceName":"web","thirdLevelDomain":"same-yaml"}]`}}}
}
func prefixOf(t *testing.T, s map[string]string) string {
	t.Helper()
	var d map[string]map[string]string
	if err := json.Unmarshal([]byte(s["customDomain"]), &d); err != nil {
		t.Fatal(err)
	}
	if d["web"]["cert"] != "keep" || d["web"]["third_party_domain"] != "external.example" {
		t.Fatal("lost unrelated domain fields", d)
	}
	return d["web"]["third_level_domain"]
}
func TestPlanConflictPolicyAndAllEntrances(t *testing.T) {
	a, b, c := legacy("a", "alice", "dup"), legacy("b", "bob", "dup"), legacy("c", "carol", "dup")
	a.Spec.Settings["addedEntrances"] = `[{"name":"dev-8080","type":"dev","port":8080}]`
	b.Spec.UserSettings = map[string]map[string]string{"dave": {"addedEntrances": `[{"name":"dev-9000","type":"dev","port":9000}]`}}
	p, err := Build([]app.Application{a, b, c}, "alice", "cluster")
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range p.Changes {
		if change.After.Entrances[0].ID == "" || change.After.SharedEntrances[0].ID == "" {
			t.Fatal("missing IDs", change)
		}
		want := ""
		if change.Name == "a" {
			want = "dup"
		}
		if got := prefixOf(t, change.After.Settings); got != want {
			t.Fatalf("%s: got %q want %q", change.Name, got, want)
		}
		if change.After.Settings["defaultThirdLevelDomainConfig"] != change.Before.Settings["defaultThirdLevelDomainConfig"] {
			t.Fatal("changed YAML")
		}
	}
	added, _ := app.ParseAddedEntrances(p.Changes[0].After.Settings)
	if added[0].ID != "abcd1234-8080" {
		t.Fatal(added)
	}
	added, _ = app.ParseAddedEntrances(p.Changes[1].After.UserSettings["dave"])
	if added[0].ID != app.NewDevEntranceID("abcd1234", "dave", "alice", 9000) {
		t.Fatal(added)
	}
	p, err = Build([]app.Application{b, c}, "alice", "cluster")
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range p.Changes {
		if prefixOf(t, change.After.Settings) != "" {
			t.Fatal("nonowner duplicate survived")
		}
	}
}
func clients(apps ...app.Application) (*fakeapi.Clientset, *fakekube.Clientset) {
	term := &sys.Terminus{ObjectMeta: metav1.ObjectMeta{Name: "terminus", UID: "cluster"}, Spec: sys.TerminusSpec{LocalName: "alice"}}
	objects := []runtime.Object{term}
	for i := range apps {
		objects = append(objects, &apps[i])
	}
	return fakeapi.NewSimpleClientset(objects...), fakekube.NewSimpleClientset()
}
func TestResumeImmutablePlanAndSkipCompleted(t *testing.T) {
	cs, kube := clients(legacy("a", "bob", "dup"), legacy("b", "carol", "dup"))
	journal := filepath.Join(t.TempDir(), "plan.json")
	fail := true
	cs.PrependReactor("update", "applications", func(action kt.Action) (bool, runtime.Object, error) {
		if action.(kt.UpdateAction).GetObject().(*app.Application).Name == "b" && fail {
			return true, nil, fmt.Errorf("interrupted")
		}
		return false, nil, nil
	})
	if err := RunLocked(context.Background(), cs, kube, journal); err == nil {
		t.Fatal("expected interruption")
	}
	original, err := os.ReadFile(journal)
	if err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(journal)
	if info.Mode().Perm() != 0600 {
		t.Fatal("backup permissions", info.Mode())
	}
	first, _ := cs.AppV1alpha1().Applications().Get(context.Background(), "a", metav1.GetOptions{})
	if prefixOf(t, first.Spec.Settings) != "" {
		t.Fatal("first app not migrated")
	}
	fail = false
	// Resume must not list again, otherwise remaining singleton b would keep dup.
	cs.PrependReactor("list", "applications", func(kt.Action) (bool, runtime.Object, error) { return true, nil, fmt.Errorf("unexpected rescan") })
	if err := RunLocked(context.Background(), cs, kube, journal); err != nil {
		t.Fatal(err)
	}
	second, _ := cs.AppV1alpha1().Applications().Get(context.Background(), "b", metav1.GetOptions{})
	if prefixOf(t, second.Spec.Settings) != "" {
		t.Fatal("recomputed conflict group")
	}
	raw, _ := os.ReadFile(journal)
	if string(raw) != string(original) {
		t.Fatal("mutated saved plan")
	}
	// Completion must short-circuit before reading backup or applications.
	if err := os.Remove(journal); err != nil {
		t.Fatal(err)
	}
	cs.PrependReactor("get", "applications", func(kt.Action) (bool, runtime.Object, error) { return true, nil, fmt.Errorf("unexpected get") })
	if err := RunLocked(context.Background(), cs, kube, journal); err != nil {
		t.Fatal(err)
	}
}
func TestResumeRejectsDriftAndPrunedIDs(t *testing.T) {
	for _, mode := range []string{"drift", "pruned"} {
		t.Run(mode, func(t *testing.T) {
			a := legacy("a", "alice", "")
			cs, kube := clients(a)
			journal := filepath.Join(t.TempDir(), "plan.json")
			p, err := Build([]app.Application{a}, "alice", "cluster")
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(p)
			if err = saveExclusive(journal, raw); err != nil {
				t.Fatal(err)
			}
			if mode == "drift" {
				a.Spec.Entrances[0].Host = "changed"
				_, _ = cs.AppV1alpha1().Applications().Update(context.Background(), &a, metav1.UpdateOptions{})
			} else {
				cs.PrependReactor("update", "applications", func(action kt.Action) (bool, runtime.Object, error) {
					obj := action.(kt.UpdateAction).GetObject().(*app.Application).DeepCopy()
					obj.Spec.Entrances[0].ID = ""
					return true, obj, nil
				})
			}
			err = RunLocked(context.Background(), cs, kube, journal)
			if err == nil {
				t.Fatal("accepted", mode)
			}
			if mode == "drift" && !strings.Contains(err.Error(), "drifted") {
				t.Fatal(err)
			}
			marker, _ := kube.CoreV1().ConfigMaps(markerNamespace).Get(context.Background(), Migration, metav1.GetOptions{})
			if marker != nil && marker.Data["status"] == "complete" {
				t.Fatal("marked failed migration complete")
			}
		})
	}
}

func TestMissingOrChangedJournalCannotReplan(t *testing.T) {
	for _, mode := range []string{"missing", "changed"} {
		t.Run(mode, func(t *testing.T) {
			cs, kube := clients(legacy("a", "bob", "dup"), legacy("b", "carol", "dup"))
			cs.PrependReactor("update", "applications", func(kt.Action) (bool, runtime.Object, error) { return true, nil, fmt.Errorf("interrupted") })
			journal := filepath.Join(t.TempDir(), "plan.json")
			if err := RunLocked(context.Background(), cs, kube, journal); err == nil {
				t.Fatal("expected failure")
			}
			if mode == "missing" {
				_ = os.Remove(journal)
			} else {
				f, err := os.OpenFile(journal, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = f.WriteString(" ")
				_ = f.Close()
			}
			cs.PrependReactor("list", "applications", func(kt.Action) (bool, runtime.Object, error) {
				t.Fatal("attempted to recompute a partial plan")
				return true, nil, nil
			})
			if err := RunLocked(context.Background(), cs, kube, journal); err == nil {
				t.Fatal("accepted changed/missing journal")
			}
		})
	}
}
