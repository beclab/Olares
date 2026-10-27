package router

import (
	"bytes"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

var installedFixture = []installedApp{
	{AppName: "wise", Title: "Wise", State: "running"},
	{AppName: "mins", Title: "Mins", State: "stopped"},
	{AppName: "llamacppqwen3", Title: "Qwen3 8B", State: "running", Shared: true},
}

// A state nobody uses is refused with the vocabulary this Olares does use.
// The states come from the platform rather than from Router, so a list spelled
// out in this tree would be a guess that goes stale.
func TestAStateNothingIsInIsRefusedWithTheOnesThatExist(t *testing.T) {
	kept, states := filterInstalledApps(installedFixture, "sleeping")
	if len(kept) != 0 {
		t.Fatalf("a state nothing is in matched %d applications", len(kept))
	}
	if strings.Join(states, ",") != "running,stopped" {
		t.Errorf("the states that exist were not collected: %v", states)
	}
}

func TestFilteringByStateKeepsOnlyThatState(t *testing.T) {
	kept, _ := filterInstalledApps(installedFixture, "RUNNING")
	if len(kept) != 2 {
		t.Fatalf("case-insensitive matching kept %d of 2", len(kept))
	}
	all, _ := filterInstalledApps(installedFixture, "")
	if len(all) != len(installedFixture) {
		t.Errorf("an empty filter dropped rows: %d of %d", len(all), len(installedFixture))
	}
}

// The table's APPLICATION column is the name --caller-app takes, so the
// rendering has to print the app name rather than the title. A title is
// written to be read and often carries spaces; pasting one into --caller-app
// works only by the fallback that also matches it.
func TestTheColumnIsTheNameTheFilterTakes(t *testing.T) {
	var buf bytes.Buffer
	if err := renderInstalledApps(&buf, installedFixture); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"llamacppqwen3", "Qwen3 8B", "stopped", "caller-app"} {
		if !strings.Contains(out, want) {
			t.Errorf("the table does not carry %q:\n%s", want, out)
		}
	}
}

func TestNoInstalledApplicationIsSaidRatherThanShownAsAnEmptyTable(t *testing.T) {
	var buf bytes.Buffer
	if err := renderInstalledApps(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "no application is installed") {
		t.Errorf("an empty directory rendered as %q", buf.String())
	}
}

// The directory is its own route. The caller_app dimension of the summary
// reads the same table but answers a different question, and spelling this one
// under /spend would reach a report instead.
func TestTheDirectoryIsItsOwnRoute(t *testing.T) {
	if epInstalledApps != consoleAPI+"/installed-apps" {
		t.Errorf("the directory route is spelled %q", epInstalledApps)
	}
}

// A draft probe is refused before the request when it has no type: the type is
// what selects the probe recipe, and without one there is nothing to run.
func TestADraftProbeNeedsToKnowWhatItIsProbing(t *testing.T) {
	err := runProviderValidateDraft(t.Context(), cmdutil.NewFactory(), draftProbe{}, nil, "", "")
	if err == nil || !strings.Contains(err.Error(), "--type") {
		t.Fatalf("a draft with no type was not refused for the type: %v", err)
	}
}

// The two forms of the verb take different arguments, and mixing them is a
// caller expecting the probe to honour something it does not: a saved provider
// is probed with what it stores, and a draft has no name.
func TestTheTwoFormsOfValidateRefuseEachOthersArguments(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"draft with a name", []string{"--draft", "--type", "openai", "openai-prod"}, "no name to give"},
		{"saved with a type", []string{"openai-prod", "--type", "openai"}, "belong to --draft"},
		{"neither", nil, "name the provider"},
	}
	for _, c := range cases {
		cmd := newProviderValidateCommand(cmdutil.NewFactory())
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs(c.args)
		err := cmd.Execute()
		if err == nil {
			t.Errorf("%s was accepted", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: the refusal does not say %q: %v", c.name, c.want, err)
		}
	}
}

// The draft route is a hyphen rather than /providers/validate, because a
// static segment cannot sit beside the /providers/:id wildcard.
func TestTheDraftProbeIsNotUnderTheProviderWildcard(t *testing.T) {
	if strings.HasPrefix(epProviderValidateDraft, epProviders+"/") {
		t.Errorf("the draft probe is spelled %q, under the wildcard that cannot hold it",
			epProviderValidateDraft)
	}
	if epProviderValidateDraft != consoleAPI+"/provider-validate" {
		t.Errorf("the draft probe is spelled %q", epProviderValidateDraft)
	}
}
