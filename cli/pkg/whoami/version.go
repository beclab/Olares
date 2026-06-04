package whoami

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/utils"
)

// OlaresInfoEndpoint is the user-service path that proxies BFL's
// /bfl/backend/v1/olares-info handler. It's the dedicated, light-weight
// counterpart to /api/init (the heavy SPA aggregate): same OlaresInfo shape
// under `data`, no extra payload. Both `settings me version` and the
// version-cache (cmdutil.Factory.OlaresBackendVersion) hit it through the
// single FetchOlaresInfo path below so there is exactly one olares-info
// implementation in the CLI — mirroring how user-info flows through
// FetchAndCache.
//
// Backend: user-service/src/init.controller.ts delegates to
// OlaresService.updateOlaresInfo(), which proxies bfl
// /bfl/backend/v1/olares-info (framework/bfl/.../handler.go) and returns a
// Result<OlaresInfo> envelope.
const OlaresInfoEndpoint = "/api/olares-info"

// OlaresInfo mirrors framework/bfl/pkg/apis/backend/v1/model.go's OlaresInfo.
// We model it here (rather than import the bfl Go module, which the CLI does
// not pull in) so callers can `--output json` straight from the struct and so
// version detection has a single typed home.
type OlaresInfo struct {
	OlaresID           string `json:"olaresId"`
	WizardStatus       string `json:"wizardStatus"`
	EnableReverseProxy bool   `json:"enableReverseProxy"`
	TailScaleEnable    bool   `json:"tailScaleEnable"`
	OsVersion          string `json:"osVersion"`
	LoginBackground    string `json:"loginBackground"`
	Avatar             string `json:"avatar"`
	ID                 string `json:"id"`
	UserDID            string `json:"did"`
	Olaresd            string `json:"olaresd"`
	Style              string `json:"style"`
}

// olaresInfoEnvelope decodes the BFL {code,message,data} wrapper. user-service
// returns code 0 for its own returnSucceed path and code 200 for routes that
// proxy through an extra layer — both are success (matching settings/me's
// doMutateEnvelope).
type olaresInfoEnvelope struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    OlaresInfo `json:"data"`
}

// FetchOlaresInfo performs the single GET /api/olares-info round-trip and
// returns the decoded OlaresInfo. This is the one and only place the CLI
// talks to that endpoint; everything that needs the OS version (or any other
// olares-info field) funnels through here.
func FetchOlaresInfo(ctx context.Context, client Doer) (*OlaresInfo, error) {
	if client == nil {
		return nil, errors.New("whoami: nil http client")
	}
	var env olaresInfoEnvelope
	if err := client.DoJSON(ctx, http.MethodGet, OlaresInfoEndpoint, nil, &env); err != nil {
		return nil, err
	}
	switch env.Code {
	case 0, 200:
	default:
		msg := strings.TrimSpace(env.Message)
		if msg == "" {
			msg = fmt.Sprintf("server returned code=%d", env.Code)
		}
		return nil, fmt.Errorf("olares-info: %s", msg)
	}
	return &env.Data, nil
}

// VersionResult is what FetchAndCacheVersion hands back: the raw osVersion
// string, its parsed semver form, and the cache-write bookkeeping (mirrors
// the user-info Result fields callers already know).
type VersionResult struct {
	OsVersion    string
	Version      *semver.Version
	Changed      bool
	WroteToCache bool  // false when caller passed cfg=nil
	RefreshedAt  int64 // Unix-second timestamp written to cache
}

// FetchAndCacheVersion is the version-cache analogue of FetchAndCache: it
// reads olares-info, parses osVersion, and (when cfg is non-nil) atomically
// updates the matching profile's BackendVersion + BackendVersionRefreshedAt.
//
// It is shared by the eager post-login fetch and by
// cmdutil.Factory.OlaresBackendVersion's lazy auto-detect, so a single
// implementation owns "read the version and remember it". `now` is injected
// for testability; pass time.Now in production.
func FetchAndCacheVersion(
	ctx context.Context,
	client Doer,
	cfg *cliconfig.MultiProfileConfig,
	olaresID string,
	now func() time.Time,
) (*VersionResult, error) {
	if olaresID == "" {
		return nil, errors.New("whoami: empty olaresID")
	}
	if now == nil {
		now = time.Now
	}
	info, err := FetchOlaresInfo(ctx, client)
	if err != nil {
		return nil, err
	}
	osVersion := strings.TrimSpace(info.OsVersion)
	if osVersion == "" {
		return nil, errors.New("olares-info: response did not include an osVersion")
	}
	v, err := utils.ParseOlaresVersionString(osVersion)
	if err != nil {
		return nil, fmt.Errorf("parse osVersion %q: %w", osVersion, err)
	}

	res := &VersionResult{OsVersion: osVersion, Version: v, RefreshedAt: now().Unix()}
	if cfg == nil {
		return res, nil
	}
	changed, err := cfg.SetBackendVersion(olaresID, v.Original(), res.RefreshedAt)
	if err != nil {
		return nil, err
	}
	res.Changed = changed
	res.WroteToCache = true
	return res, nil
}
