package clusterclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ControlHub talks at least three wire shapes; this file consolidates the
// envelope-aware GET helpers so verb code stays dense.
//
// Background — which prefix uses which shape (see
// apps/packages/app/src/apps/controlPanelCommon/network/network.ts and
// apps/packages/app/src/apps/controlPanelCommon/network/middleware.ts for
// the SPA's reference):
//
//   - "/kapis/..."      → KubeSphere paginated list:
//                         {"items": [...], "totalItems": N}  (use ListResponse)
//   - "/api/v1/..."     → K8s native list:
//                         {"kind":"PodList","apiVersion":"v1","items":[...],
//                          "metadata":{"resourceVersion":"..."}}
//   - "/apis/<g>/<v>/..." → K8s native list (same envelope as /api/v1/...)
//   - "/capi/app/detail" → ControlHub custom object (no envelope, decode
//                          straight into a typed struct)
//   - "/middleware/..." → Olares middleware aggregator (object or list,
//                         shape depends on path; treat as raw and decode
//                         into a per-call typed struct)

// ListResponse is the canonical KubeSphere paginated list envelope returned
// by every "/kapis/..." path the SPA consumes (see
// apps/packages/app/src/apps/controlPanelCommon/network/network.ts:117
// `ResourcesResponse`). TotalItems is the unfiltered total — the API does
// pagination server-side via `?limit=&page=` query params, so callers that
// want everything iterate Page until len(Items) < limit OR
// (page-1)*limit + len(Items) >= TotalItems.
//
// T is the per-call element type (corev1.Pod, appsv1.Deployment, ...).
// We intentionally do NOT vendor k8s.io/api just for these structs:
// callers define minimal local structs that capture only the fields they
// render, mirroring how each settings tree area declares its own typed
// view.
type ListResponse[T any] struct {
	Items      []T `json:"items"`
	TotalItems int `json:"totalItems"`
}

// K8sList is the K8s native list envelope returned by `/api/v1/...` and
// `/apis/<group>/<version>/...` paths (see e.g. PodList / DeploymentList
// in k8s.io/api/core/v1). The same shape covers single-resource GETs
// (just decode straight into the per-call typed struct without this
// helper).
//
// Note: K8s native lists don't carry a TotalItems; pagination happens via
// metadata.continue + opaque resourceVersion. We expose only Items + the
// resource version for callers that want to chain --watch via
// resourceVersion-based polling, but most read verbs ignore both.
type K8sList[T any] struct {
	Kind       string         `json:"kind,omitempty"`
	APIVersion string         `json:"apiVersion,omitempty"`
	Items      []T            `json:"items"`
	Metadata   K8sListMetadata `json:"metadata,omitempty"`
}

// K8sListMetadata is the partial metav1.ListMeta shape we care about for
// list responses. `continue` is opaque and forwarded back into the next
// request when paginating. `resourceVersion` is the sequence point future
// --watch polling can resume from.
type K8sListMetadata struct {
	Continue        string `json:"continue,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// GetKubeSphereList fetches a "/kapis/..."-style paginated list and decodes
// it into a ListResponse[T]. Returns the ListResponse so callers can read
// TotalItems even when len(Items) < TotalItems (paginated). On HTTP error
// the underlying *clusterclient.Client surfaces a reformatted message via
// the same path Cluster.DoJSON uses; this helper just decodes.
func GetKubeSphereList[T any](ctx context.Context, c *Client, path string) (*ListResponse[T], error) {
	var out ListResponse[T]
	if err := c.DoJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetK8sList fetches a "/api/v1/..." or "/apis/<g>/<v>/..." K8s native list
// and decodes it into a K8sList[T]. Same error surface as
// GetKubeSphereList.
func GetK8sList[T any](ctx context.Context, c *Client, path string) (*K8sList[T], error) {
	var out K8sList[T]
	if err := c.DoJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetK8sObject fetches a single K8s native object (or any other ControlHub
// path that returns a typed JSON document with no envelope) and decodes
// it into the supplied out parameter. Wraps DoJSON for symmetry with the
// other helpers.
//
// `out` MUST be a pointer to a typed struct or to interface{}; nil is
// rejected to avoid the silent "got bytes but couldn't decode" footgun
// the raw API would otherwise allow.
func GetK8sObject(ctx context.Context, c *Client, path string, out interface{}) error {
	if out == nil {
		return fmt.Errorf("clusterclient: GetK8sObject called with nil out")
	}
	return c.DoJSON(ctx, http.MethodGet, path, nil, out)
}

// GetRaw returns the raw response body for a GET request. Used by verbs
// that need to forward bytes to the user verbatim (`cluster pod yaml`
// converts the JSON body to YAML and prints it; some monitoring paths
// return non-JSON shapes our typed helpers can't model). 401/403 / other
// HTTP errors are still reformatted before this returns.
func GetRaw(ctx context.Context, c *Client, path string) ([]byte, error) {
	return c.DoRaw(ctx, http.MethodGet, path, nil)
}

// DecodeJSON is a tiny convenience wrapper: when a verb has already
// fetched bytes via GetRaw and decided what shape to expect, this avoids
// a one-line `json.Unmarshal` + error wrapping at every call site.
//
// The error wraps the path so the user can tell which response failed to
// decode (helpful when a verb fans out to multiple endpoints).
func DecodeJSON(path string, body []byte, out interface{}) error {
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode %s response: %w", path, err)
	}
	return nil
}

// Patch issues a PATCH against `path` with the supplied Content-Type
// (typically "application/merge-patch+json" or
// "application/strategic-merge-patch+json" — K8s picks the merge
// algorithm by header) and decodes the response into `out` when
// non-nil.
//
// out should be a pointer to the typed struct describing the patched
// object's expected shape (K8s returns the post-patch object on
// success). Pass nil to discard the body when the caller only cares
// about the success/failure outcome.
func Patch[T any](ctx context.Context, c *Client, path, contentType string, body interface{}, out *T) error {
	if out == nil {
		return c.DoJSONWithContentType(ctx, http.MethodPatch, path, body, contentType, nil)
	}
	return c.DoJSONWithContentType(ctx, http.MethodPatch, path, body, contentType, out)
}
