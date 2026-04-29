package overview

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// sectionKeys is the canonical iteration order for the sections
// envelope. Pinned both for the JSON consumer (so map-order is
// stable) and for the table renderer (so the human-facing output
// scrollback is deterministic).
var sectionKeys = []string{"physical", "user", "ranking"}

// RunDefault is the cmd-side entry point for `dashboard overview`
// without a subverb. Fans out the three SECTIONS in parallel; per-
// section failures populate Meta.Error on that section without
// aborting the whole envelope (mirrors the SPA's "partial degradation
// is fine, surface it" behaviour). Watch-loop semantics are NOT
// applied here — the default action is one-shot only because the
// per-leaf cadences differ and a unified --watch-interval would lie
// to consumers.
func RunDefault(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	now := time.Now()
	env := BuildSectionsEnvelope(ctx, c, cf, now)
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	return WriteSectionsTable(os.Stdout, env)
}

// BuildSectionsEnvelope fans out the three section builders in
// parallel, attaches the parent envelope, and returns a single
// Envelope with Sections populated. Per-section errors land on the
// section's own Meta.Error / Meta.ErrorKind; the parent envelope
// itself never fails (callers can tell the difference by walking
// Sections[*].Meta.Error).
func BuildSectionsEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) pkgdashboard.Envelope {
	type sectionResult struct {
		key string
		env pkgdashboard.Envelope
	}
	results := make(chan sectionResult, len(sectionKeys))

	go func() {
		env, err := BuildPhysicalEnvelope(ctx, c, cf, now)
		if err != nil {
			env.Kind = pkgdashboard.KindOverviewPhysical
			env.Meta.Error = err.Error()
			env.Meta.ErrorKind = pkgdashboard.ClassifyTransportErr(err)
		}
		results <- sectionResult{"physical", env}
	}()
	go func() {
		env, err := BuildUserEnvelope(ctx, c, cf, cf.User, now)
		if err != nil {
			env.Kind = pkgdashboard.KindOverviewUser
			env.Meta.Error = err.Error()
			env.Meta.ErrorKind = pkgdashboard.ClassifyTransportErr(err)
		}
		results <- sectionResult{"user", env}
	}()
	go func() {
		env, err := BuildRankingEnvelope(ctx, c, cf, "desc", now)
		if err != nil {
			env.Kind = pkgdashboard.KindOverviewRanking
			env.Meta.Error = err.Error()
			env.Meta.ErrorKind = pkgdashboard.ClassifyTransportErr(err)
		}
		results <- sectionResult{"ranking", env}
	}()

	out := map[string]pkgdashboard.Envelope{}
	for i := 0; i < len(sectionKeys); i++ {
		r := <-results
		r.env.Meta.FetchedAt = time.Now().In(cf.Timezone.Time()).Format(time.RFC3339)
		out[r.key] = r.env
	}
	return pkgdashboard.Envelope{
		Kind:     pkgdashboard.KindOverview,
		Meta:     pkgdashboard.NewMeta(time.Now().In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Sections: out,
	}
}

// WriteSectionsTable lays out the three sections back-to-back in
// table mode. Section banners use a leading "==" so a human scanning
// the scrollback can locate them by simple pattern match. A section
// with Meta.Error populated emits a "(error: …)" placeholder rather
// than its table — keeping the scrollback aligned even when one
// fetch failed.
func WriteSectionsTable(w io.Writer, env pkgdashboard.Envelope) error {
	for _, key := range sectionKeys {
		section, ok := env.Sections[key]
		if !ok {
			continue
		}
		fmt.Fprintf(w, "== %s ==\n", strings.ToUpper(key))
		if section.Meta.Error != "" {
			fmt.Fprintf(w, "(error: %s)\n\n", section.Meta.Error)
			continue
		}
		switch section.Kind {
		case pkgdashboard.KindOverviewPhysical:
			if err := WritePhysicalTable(w, section); err != nil {
				return err
			}
		case pkgdashboard.KindOverviewUser:
			if err := WriteUserTable(w, section); err != nil {
				return err
			}
		case pkgdashboard.KindOverviewRanking:
			if err := WriteRankingTable(w, section); err != nil {
				return err
			}
		}
		fmt.Fprintln(w)
	}
	return nil
}
