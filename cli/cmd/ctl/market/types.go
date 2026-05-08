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
}

type UninstallRequest struct {
	Sync       bool `json:"sync"`
	All        bool `json:"all"`
	DeleteData bool `json:"deleteData"`
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
	Status AppStatus `json:"status"`
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
