package router

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// The catalog's two right-hand columns are the only place a reader learns what
// can be done with a row, and each of the four answers sends them somewhere
// different: install it, clone it into an instance, upgrade or remove what is
// already here, or nothing at all because another source holds the name.
func TestCatalogSaysWhatEachRowTakesAndWhereItStands(t *testing.T) {
	items := []marketApp{
		{AppName: "llamacppllmbasev3", Title: map[string]string{"en-US": "llama.cpp base"}, TemplateOnly: true},
		{AppName: "embeddinggemmav3", Title: map[string]string{"en-US": "EmbeddingGemma"}},
		{
			AppName: "qwen3v3",
			Install: marketInstallState{Installed: true, Status: "downloading", ProviderID: "p-1"},
		},
		{
			AppName: "fromanothersource",
			Source:  "market.example",
			Install: marketInstallState{TakenBySource: "market.olares"},
		},
	}

	var out bytes.Buffer
	if err := renderCatalog(&out, items); err != nil {
		t.Fatalf("renderCatalog: %v", err)
	}
	got := out.String()

	for _, want := range []string{
		"llamacppllmbasev3", "llama.cpp base", "clone", "market clone",
		"install",
		"downloading", "upgrade, uninstall",
		"taken by market.olares", "one app name occupies one",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("catalog output is missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "(?)") {
		t.Errorf("nothing here is source-ambiguous, so no row should be marked uncertain:\n%s", got)
	}
}

// An install nobody could attribute to a source is reported against every copy
// of the name, so the row has to say the status might be a sibling's. Left
// unsaid, a reader takes it for this row's own.
func TestCatalogMarksAnUnattributableInstallUncertain(t *testing.T) {
	var out bytes.Buffer
	err := renderCatalog(&out, []marketApp{{
		AppName: "qwen3v3",
		Install: marketInstallState{
			Installed: true, Status: "running", ProviderID: "p-1", SourceAmbiguous: true,
		},
	}})
	if err != nil {
		t.Fatalf("renderCatalog: %v", err)
	}
	got := out.String()
	for _, want := range []string{"running (?)", "may belong to a sibling"} {
		if !strings.Contains(got, want) {
			t.Errorf("catalog output is missing %q:\n%s", want, got)
		}
	}
}

// Each note under the table describes a row the reader can see. A catalog of
// ordinary applications must carry none of them.
func TestCatalogExplainsOnlyWhatItShows(t *testing.T) {
	var out bytes.Buffer
	if err := renderCatalog(&out, []marketApp{{AppName: "embeddinggemmav3"}}); err != nil {
		t.Fatalf("renderCatalog: %v", err)
	}
	got := out.String()
	for _, unwanted := range []string{"market clone", "taken by", "sibling"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("catalog output should not mention %q:\n%s", unwanted, got)
		}
	}
}

// A title arrives as a whole locale map, and an application translated into
// nothing at all still has a name to print.
func TestAnUntranslatedAppIsStillNamed(t *testing.T) {
	app := marketApp{AppName: "qwen3v3"}
	if got := app.title(); got != "qwen3v3" {
		t.Errorf("title of an untranslated app: got %q want the app name", got)
	}
	app.Title = map[string]string{"zh-CN": "千问"}
	if got := app.title(); got != "千问" {
		t.Errorf("title with only one locale: got %q want that locale", got)
	}
}

var twoSources = []marketApp{
	{AppName: "qwen3v3", Source: "market.olares", Version: "1.0.0"},
	{AppName: "qwen3v3", Source: "market.example", Version: "0.9.0"},
	{AppName: "embeddinggemmav3", Source: "market.olares"},
}

// Two sources publishing one name are two different builds, and only one copy
// can be installed. Picking for someone installs software they did not choose,
// so the refusal names the sources instead.
func TestInstallWillNotChooseBetweenTwoSources(t *testing.T) {
	_, err := pickCatalogApp(twoSources, "qwen3v3", "")
	if err == nil {
		t.Fatal("an app published twice was resolved without --source")
	}
	for _, want := range []string{"market.example", "market.olares", "--source"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal is missing %q: %s", want, err)
		}
	}
}

func TestInstallTakesTheNamedSource(t *testing.T) {
	app, err := pickCatalogApp(twoSources, "qwen3v3", "market.example")
	if err != nil {
		t.Fatalf("pickCatalogApp with a named source: %v", err)
	}
	if app.Version != "0.9.0" {
		t.Errorf("resolved the wrong copy: got version %q want 0.9.0", app.Version)
	}
	if _, err := pickCatalogApp(twoSources, "qwen3v3", "market.nowhere"); err == nil {
		t.Fatal("a source that publishes nothing was accepted")
	}
}

// A name published once needs no choice, and the source it came from is sent
// with the install so the Market cannot pick a different copy in between.
func TestInstallNeedsNoSourceForAnUncontestedName(t *testing.T) {
	app, err := pickCatalogApp(twoSources, "embeddinggemmav3", "")
	if err != nil {
		t.Fatalf("pickCatalogApp: %v", err)
	}
	if app.Source != "market.olares" {
		t.Errorf("source: got %q want market.olares", app.Source)
	}
}

// "There is no such application" and "the Market offers you nothing" send a
// reader somewhere different.
func TestAnUnknownAppNamesWhatTheMarketOffers(t *testing.T) {
	_, err := pickCatalogApp(twoSources, "qwen4", "")
	if err == nil {
		t.Fatal("an unknown application resolved")
	}
	for _, want := range []string{`no model application named "qwen4"`, "qwen3v3", "embeddinggemmav3"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal is missing %q: %s", want, err)
		}
	}
	_, err = pickCatalogApp(nil, "qwen4", "")
	if err == nil || !strings.Contains(err.Error(), "offers none to this account") {
		t.Errorf("an empty catalog should say so: %v", err)
	}
}

// Three rows the Market would refuse or accept-then-fail. Catching them here is
// the difference between an answer and a piece of history to explain.
func TestInstallIsRefusedBeforeTheMarketWouldRefuseIt(t *testing.T) {
	cases := []struct {
		name string
		app  marketApp
		want []string
	}{
		{
			name: "a template",
			app:  marketApp{AppName: "llamacppllmbasev3", TemplateOnly: true},
			want: []string{"engine template", "market clone", "MODEL_SOURCE"},
		},
		{
			name: "a name another source holds",
			app: marketApp{
				AppName: "qwen3v3",
				Install: marketInstallState{TakenBySource: "market.olares"},
			},
			want: []string{"already taken", "market.olares", "uninstalled"},
		},
		{
			name: "already installed here",
			app: marketApp{
				AppName: "qwen3v3",
				Install: marketInstallState{Installed: true, Status: "stopped", ProviderID: "p-1"},
			},
			want: []string{"already installed", "stopped", "app upgrade", "market resume"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := refuseUninstallable(&tc.app)
			if err == nil {
				t.Fatal("the install was allowed")
			}
			for _, want := range tc.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("refusal is missing %q: %s", want, err)
				}
			}
		})
	}
	if err := refuseUninstallable(&marketApp{AppName: "qwen3v3"}); err != nil {
		t.Errorf("an installable application was refused: %v", err)
	}
}

// The same status means opposite things depending on what was asked for, which
// is the whole reason the verdict is keyed by action.
func TestUnreachableEndsAnUninstallAndBreaksAnInstall(t *testing.T) {
	if got := lifecycleFor(marketActionUninstall).verdict("unreachable", "", false, 0); got != "done" {
		t.Errorf("uninstall reaching unreachable: got %q want done", got)
	}
	if got := lifecycleFor(marketActionInstall).verdict("unreachable", "", false, 0); got != "failed" {
		t.Errorf("install reaching unreachable: got %q want failed", got)
	}
}

// An upgrade starts on `running`, so the first poll would otherwise read
// success off the state being replaced — and a short upgrade that finishes
// before the first poll has no departure to observe, which the grace covers.
func TestAnUpgradeWaitsToSeeTheAppLeaveRunning(t *testing.T) {
	lc := lifecycleFor(marketActionUpgrade)
	if got := lc.verdict("running", "", false, time.Second); got != "" {
		t.Errorf("first poll of an upgrade: got %q want no verdict yet", got)
	}
	if got := lc.verdict("running", "", true, time.Second); got != "done" {
		t.Errorf("after the app was seen leaving running: got %q want done", got)
	}
	if got := lc.verdict("running", "", false, marketDepartureGrace+time.Second); got != "done" {
		t.Errorf("after the grace expired: got %q want done", got)
	}
}

// A running application is not ready while it is still fetching the weights it
// serves, and an application whose model failed to load has failed even though
// its container is fine. Neither applies to an uninstall, where the phase is a
// leftover from the app being removed.
func TestAModelStillLoadingHoldsAnInstallOpen(t *testing.T) {
	install := lifecycleFor(marketActionInstall)
	if got := install.verdict("running", "download", false, 0); got != "" {
		t.Errorf("running while downloading a model: got %q want no verdict yet", got)
	}
	if got := install.verdict("running", "ready", false, 0); got != "done" {
		t.Errorf("running with a loaded model: got %q want done", got)
	}
	if got := install.verdict("running", "failed", false, 0); got != "failed" {
		t.Errorf("running with a model that could not load: got %q want failed", got)
	}
	if got := lifecycleFor(marketActionUninstall).verdict("unreachable", "failed", false, 0); got != "done" {
		t.Errorf("uninstall must not read the model phase: got %q want done", got)
	}
}

// A row with no status yet is what the first poll of a freshly created provider
// sees, and it is not an outcome.
func TestNoStatusYetIsNotAVerdict(t *testing.T) {
	if got := lifecycleFor(marketActionInstall).verdict("", "", false, time.Hour); got != "" {
		t.Errorf("a row with no status: got %q want no verdict", got)
	}
}
