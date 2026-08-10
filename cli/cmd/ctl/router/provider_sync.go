package router

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli router provider sync-models <name|id>`
// POST /console/api/providers/:id/sync-models
//
// For an upstream that publishes its own model list — Ollama, a generic
// OpenAI-compatible endpoint, a model application installed on this Olares —
// the list is the upstream's to own. This mirrors it: models the upstream has
// and Router does not are added, models Router has and the upstream no longer
// serves are removed, and models on both sides are left exactly as they are, so
// an enable/disable or a quota you set survives every re-sync.
//
// One guard is worth relying on: an upstream that answers with zero models
// prunes nothing. A blank list from a starting-up application must not wipe a
// working catalog.

type syncModelsResult struct {
	Created    int    `json:"created"`
	Deleted    int    `json:"deleted"`
	Unchanged  int    `json:"unchanged"`
	ModelCount int    `json:"model_count"`
	ProbeURL   string `json:"probe_url"`
	SyncedAt   string `json:"synced_at"`
}

func newProviderSyncModelsCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "sync-models <name|id>",
		Short: "mirror an upstream's live model list into Router",
		Long: `Mirror what an upstream currently serves into Router's model list.

Use this for an upstream whose catalog is its own: Ollama, a generic
OpenAI-compatible endpoint, or a model application on this Olares. For a vendor
with a published catalog — OpenAI, Anthropic, Gemini — the models are chosen
from that catalog instead, with "provider models add".

What a sync does:

  a model the upstream serves and Router lacks     is added, as a chat model
                                                   with no capabilities recorded
  a model Router has and the upstream dropped      is removed, along with the
                                                   defaults, settings and quotas
                                                   pointing at it
  a model on both sides                            is left untouched, so your
                                                   enable/disable and quota edits
                                                   survive

An upstream that answers with an empty list changes nothing, on purpose: an
application that is still starting must not be able to erase a working catalog.

Removal is real. A model that comes back on the next sync returns as a fresh
row, without the settings the old one carried.

Example:
  olares-cli router provider sync-models ollama-local
`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runProviderSyncModels(c.Context(), f, args[0], output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runProviderSyncModels(ctx context.Context, f *cmdutil.Factory, ref, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	found, err := resolveProvider(ctx, pc, ref)
	if err != nil {
		return err
	}
	var res syncModelsResult
	path := consoleAPI + "/providers/" + url.PathEscape(found.ID) + "/sync-models"
	if err := pc.router.doJSON(ctx, "POST", path, nil, &res); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, res)
	}
	if _, err := fmt.Fprintf(os.Stdout,
		"%s now mirrors %d models from %s: %d added, %d removed, %d unchanged (at %s)\n",
		found.Name, res.ModelCount, nonEmpty(res.ProbeURL),
		res.Created, res.Deleted, res.Unchanged, nonEmpty(res.SyncedAt)); err != nil {
		return err
	}
	if res.ModelCount == 0 {
		_, err := fmt.Fprintf(os.Stdout,
			"The upstream reported no models, so nothing was removed. "+
				"Check that it is running and answering: `olares-cli router provider validate %s`\n",
			found.Name)
		return err
	}
	return nil
}
