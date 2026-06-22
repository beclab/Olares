package market

type AppQueryInfo struct {
	AppID          string `json:"appid"`
	SourceDataName string `json:"sourceDataName"`
}

type AppEnvVar struct {
	EnvName string `json:"envName" yaml:"envName"`
	Value   string `json:"value,omitempty" yaml:"value,omitempty"`
}

type AppEntrance struct {
	Name  string `json:"name"`
	Title string `json:"title"`
}

type InstallRequest struct {
	Source  string      `json:"source"`
	AppName string      `json:"app_name"`
	Version string      `json:"version"`
	Sync    bool        `json:"sync"`
	Envs    []AppEnvVar `json:"envs,omitempty"`
}

type CloneRequest struct {
	Source    string        `json:"source"`
	AppName   string        `json:"app_name"`
	Title     string        `json:"title"`
	Sync      bool          `json:"sync"`
	Envs      []AppEnvVar   `json:"envs,omitempty"`
	Entrances []AppEntrance `json:"entrances,omitempty"`
	// TemplateClone marks an instance created from a template app (no
	// installable body), mirroring the SPA's onClone() which sets
	// templateClone:true for templateOnly apps (1.12.6+). omitempty keeps
	// the 1.12.5 wire byte-identical (the flag is only ever set on 1.12.6).
	TemplateClone bool `json:"templateClone,omitempty"`
}

type UninstallRequest struct {
	Sync       bool `json:"sync"`
	All        bool `json:"all"`
	DeleteData bool `json:"deleteData"`
}

// UpgradeRequest is the payload for /apps/{name}/upgrade. Deliberately
// does NOT carry env vars: the Market SPA's upgradeApp() sends only
// {app_name, source, version} (see InstallButton.vue / useAppAction.ts).
// Existing env values are preserved server-side from the previous
// install. Use `olares-cli market env` to update them out-of-band.
type UpgradeRequest struct {
	Source  string `json:"source"`
	AppName string `json:"app_name"`
	Version string `json:"version"`
	Sync    bool   `json:"sync"`
}

// OperationResult is the structured output for mutating commands.
//
// FinalState / FinalOpType are populated only by --watch flows once a
// terminal classification has been reached; both use omitempty so JSON
// emitted by non-watch invocations stays byte-identical to the previous
// release. They duplicate State so scripts can distinguish "the latest
// state we observed" from "the state the watcher classified as terminal"
// (e.g. when failures surface a Reason that already moved the row on).
type OperationResult struct {
	App         string `json:"app"`
	TargetApp   string `json:"targetApp,omitempty"`
	Operation   string `json:"operation"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	Source      string `json:"source,omitempty"`
	Version     string `json:"version,omitempty"`
	User        string `json:"user,omitempty"`
	State       string `json:"state,omitempty"`
	Progress    string `json:"progress,omitempty"`
	FinalState  string `json:"finalState,omitempty"`
	FinalOpType string `json:"finalOpType,omitempty"`
}

type SourceStateData struct {
	Type           string           `json:"type"`
	AppStateLatest []AppStateLatest `json:"app_state_latest"`
}

type AppStateLatest struct {
	// Version is the version recorded on this per-user state row —
	// the chart the user picked to install or upgrade to, regardless
	// of whether that operation has completed. The marketplace backend
	// exposes it as a sibling of `status`, NOT inside it, so we have
	// to surface it at this level. The legacy `app` tree never needed
	// it; the per-user `market list --mine` view does because the
	// row's version otherwise has to be derived from the catalog,
	// which is wrong as soon as the catalog's latest moves ahead of
	// the user's row (the most common case) or when an upgrade is
	// mid-flight (the user picked vN+1 but vN is still running).
	Version string    `json:"version,omitempty"`
	Status  AppStatus `json:"status"`
}

type AppStatus struct {
	Name     string `json:"name"`
	RawName  string `json:"rawAppName"`
	Title    string `json:"title,omitempty"`
	State    string `json:"state"`
	OpType   string `json:"opType,omitempty"`
	CfgType  string `json:"cfgType,omitempty"`
	Progress string `json:"progress,omitempty"`
	Message  string `json:"message,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type SourceInfoData struct {
	Type          string              `json:"type"`
	AppInfoLatest []AppInfoLatestItem `json:"app_info_latest"`
}

type AppInfoLatestItem struct {
	Type          string                 `json:"type"`
	Timestamp     int64                  `json:"timestamp"`
	Version       string                 `json:"version,omitempty"`
	AppSimpleInfo map[string]interface{} `json:"app_simple_info"`
}

type MarketDataResponse struct {
	UserData  *UserDataFiltered `json:"user_data"`
	UserID    string            `json:"user_id"`
	Timestamp int64             `json:"timestamp"`
}

type UserDataFiltered struct {
	Sources map[string]*SourceInfoData `json:"sources"`
	Hash    string                     `json:"hash"`
}

type MarketStateResponse struct {
	UserData  *UserDataStateFiltered `json:"user_data"`
	UserID    string                 `json:"user_id"`
	Timestamp int64                  `json:"timestamp"`
}

type UserDataStateFiltered struct {
	Sources map[string]*SourceStateData `json:"sources"`
	Hash    string                      `json:"hash"`
}

type AppDisplayInfo struct {
	Name       string   `json:"name"`
	Title      string   `json:"title"`
	Version    string   `json:"version"`
	Source     string   `json:"source"`
	Categories []string `json:"categories,omitempty"`
	// State is populated only by the "my apps" listing path
	// (`market list --mine`). Catalog browsing leaves it empty and
	// `omitempty` keeps the JSON shape byte-identical with the
	// pre-flag release.
	State string `json:"state,omitempty"`
}

// notInstalledStates is the verbatim mirror of the SPA's
// `uninstalledAppStates` set in apps/packages/app/src/constant/config.ts
// (around line 170). That set is what `MarketRemotePage.vue` →
// `appStore.getSourceInstalledApp(sourceId)` consults via
// `uninstalledApp(status)` to decide which rows are HIDDEN from the
// Market UI's "My Terminus" tab.
//
// `market list --mine` is the CLI counterpart of "My Terminus", so we
// deliberately mirror this SPA filter — NOT the broader upstream state
// machine in framework/app-service/pkg/appstate/state_transition.go —
// so the two listings show the exact same set of apps. Note this
// means `--mine` is NOT "已安装应用 / completed installs only": the
// SPA keeps in-flight install rows (`pending`, `downloading`,
// `installing`, plus their `*Canceling` / `*CancelFailed` variants),
// post-install transitional rows (`upgrading`, `resuming`, `stopping`,
// `applyingEnv`, `uninstalling`), and post-install failures
// (`upgradeFailed`, `stopFailed`, `resumeFailed`, `applyEnvFailed`,
// `uninstallFailed`, ...) all visible on My Terminus because they're
// still "the user's apps" — the user clicked something and expects
// to see / monitor / cancel / retry the row. Only the 6 SPA-hidden
// states below — terminal-cancel of the install path and terminal
// `uninstalled` — are NOT part of "my apps".
//
// (The variable name uses "notInstalled" as an internal shorthand for
// "not part of the user's apps per the SPA filter"; `--mine` is the
// UX-facing name. Both refer to the same denylist defined here.)
//
// If the SPA adds or renames an entry in `uninstalledAppStates`, update
// this set to match so the two listings stay in sync.
var notInstalledStates = map[string]struct{}{
	"pendingCanceled":     {}, // APP_STATUS.PENDING.CANCELED
	"downloadingCanceled": {}, // APP_STATUS.DOWNLOAD.CANCELED
	"downloadFailed":      {}, // APP_STATUS.DOWNLOAD.FAILED
	"installFailed":       {}, // APP_STATUS.INSTALL.FAILED
	"installingCanceled":  {}, // APP_STATUS.INSTALL.CANCELED
	"uninstalled":         {}, // APP_STATUS.UNINSTALL.COMPLETED
}

// isInstalledState reports whether a row's `state` value would render
// on the SPA's "My Terminus" tab — i.e. whether it counts as one of
// the user's apps under `market list --mine`. Returns false only for
// the SPA's six `uninstalledAppStates`; every other state — including
// in-flight install rows (`pending` / `downloading` / `installing`
// and their `*Canceling` / `*CancelFailed` variants) plus all
// post-install transitional / failure states (`upgrading`, `stopping`,
// `resuming`, `applyingEnv`, `upgradeFailed`, `stopFailed`,
// `uninstallFailed`, ...) — returns true. The helper name reflects
// the internal "installed" shorthand; the user-facing semantic is
// "is this row part of the user's apps?" (== "show on My Terminus").
func isInstalledState(state string) bool {
	if state == "" {
		return false
	}
	_, missed := notInstalledStates[state]
	return !missed
}

// extractLocalizedString resolves a value that may be a plain string
// or an i18n map (e.g. {"en-US": "Firefox", "zh-CN": "火狐"}).
func extractLocalizedString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case map[string]interface{}:
		for _, locale := range []string{"en-US", "en", "zh-CN"} {
			if s, ok := val[locale].(string); ok && s != "" {
				return s
			}
		}
		for _, s := range val {
			if str, ok := s.(string); ok && str != "" {
				return str
			}
		}
	}
	return ""
}

func extractAppDisplayInfo(item AppInfoLatestItem, sourceName string) *AppDisplayInfo {
	m := item.AppSimpleInfo
	if m == nil {
		return nil
	}

	name, _ := m["app_name"].(string)
	if name == "" {
		name, _ = m["app_id"].(string)
	}
	if name == "" {
		return nil
	}

	title := extractLocalizedString(m["app_title"])
	if title == "" {
		title = extractLocalizedString(m["title"])
	}

	version := item.Version
	if version == "" {
		version, _ = m["app_version"].(string)
	}

	var categories []string
	if cats, ok := m["categories"].([]interface{}); ok {
		for _, c := range cats {
			if s, ok := c.(string); ok && s != "" {
				categories = append(categories, s)
			}
		}
	}

	return &AppDisplayInfo{
		Name:       name,
		Title:      title,
		Version:    version,
		Source:     sourceName,
		Categories: categories,
	}
}
