package helmops

import "context"

// GetDeployedVersion implements ports.WorkloadOps. Legacy call
// site: upgrading_app.go:169 (apputils.GetDeployedReleaseVersion).
// Returns the currently-deployed chart version + revision so the
// upgrade handler can guard against downgrades.
//
// This is a HANDLER-SAFE READ per the port comment — the Runner
// materializes it into domain.World.Workload before Engine.Step
// runs, so handlers see a snapshot rather than making a live call.
//
// Namespace is the release namespace (not necessarily the app
// namespace); release is the helm release name (== appCfg.AppName
// on every path).
func (o *Ops) GetDeployedVersion(ctx context.Context, namespace, release string) (string, int, error) {
	return o.getVersion(ctx, o.kubeConfig, namespace, release)
}
