package download

import "time"

// Types mirror download-server manager wire JSON (models.DownloadTask /
// NewDownloadReq / preferences / url.InspectData), not thrift IDL names.

// DownloadTask is the list/info/create response row shape.
type DownloadTask struct {
	ID                int64                  `json:"id"`
	Username          string                 `json:"username"`
	App               string                 `json:"app"`
	URL               string                 `json:"url"`
	DownloadProvider  string                 `json:"download_provider"`
	Status            string                 `json:"status"`
	Path              string                 `json:"path"`
	FileName          string                 `json:"file_name"`
	FileType          string                 `json:"file_type,omitempty"`
	Size              int64                  `json:"size"`
	DownloadedBytes   int64                  `json:"downloaded_bytes"`
	Percent           float32                `json:"percent"`
	Extra             map[string]interface{} `json:"extra,omitempty"`
	ProviderTaskID    string                 `json:"provider_task_id,omitempty"`
	ErrMsg            string                 `json:"err_msg,omitempty"`
	ErrCategory       string                 `json:"err_category,omitempty"`
	RetryCount        int                    `json:"retry_count,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	LiveDownloadSpeed int64                  `json:"live_download_speed,omitempty"`
	LiveUploadSpeed   int64                  `json:"live_upload_speed,omitempty"`
	FileMissing       *bool                  `json:"file_missing,omitempty"`
	IsDir             *bool                  `json:"is_dir,omitempty"`
}

// NewDownloadReq is POST /api/download body. Extra values are strings on
// the wire (thrift map[string]string); the manager promotes a few keys.
type NewDownloadReq struct {
	URL      string            `json:"url"`
	App      string            `json:"app"`
	Path     string            `json:"path,omitempty"`
	FileName string            `json:"file_name,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

// ListResult is GET /api/download/list success body (list + total at top level).
type ListResult struct {
	List  []DownloadTask `json:"list"`
	Total int64          `json:"total"`
}

// RemoveReq is DELETE /api/download/remove body.
type RemoveReq struct {
	TaskID     int64 `json:"task_id"`
	RemoveFlag bool  `json:"remove_flag"`
}

// UserPreference is GET/PUT /api/user/preferences data.
type UserPreference struct {
	Username     string    `json:"username"`
	App          string    `json:"app"`
	YtdlpQuality string    `json:"ytdlp_quality"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

// InspectData is GET /api/url/inspect data.
type InspectData struct {
	Provider           string   `json:"provider"`
	Title              string   `json:"title,omitempty"`
	Thumbnail          string   `json:"thumbnail,omitempty"`
	AvailableQualities []string `json:"available_qualities,omitempty"`
	Error              string   `json:"error,omitempty"`
	ErrorCode          int      `json:"error_code,omitempty"`
	ErrorCategory      string   `json:"error_category,omitempty"`
	Available          *bool    `json:"available,omitempty"`
}
