// Package fan hosts the business logic for the
// `olares-cli dashboard overview fan` subtree (root + live +
// curve). The cmd-side area
// (cli/cmd/ctl/dashboard/overview/fan/) is a thin shell — it owns
// cobra wiring, persistent-flag binding, and the area-private
// *Client factory; this package owns the fetcher, the gating, the
// envelope shape, and the table renderer for both leaves +
// the sections envelope.
//
// Layout choice: the directory tree mirrors the command tree
// (overview/fan/{root,live,curve}.go). All Olares One-only
// gating happens here — every leaf routes through the same
// pkgdashboard.GateOlaresOne so the SPA's `FanStore.isOlaresOneDevice`
// invariant (`Overview2/ClusterResource.vue:238`) is enforced
// once, in one place.
package fan
