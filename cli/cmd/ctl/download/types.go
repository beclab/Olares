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

// TorrentInspectResult is POST /api/download/torrent/inspect data.
type TorrentInspectResult struct {
	InfoHash     string               `json:"info_hash"`
	Name         string               `json:"name"`
	Mode         string               `json:"mode"` // "single" | "multi"
	PieceLength  int64                `json:"piece_length"`
	NumPieces    int64                `json:"num_pieces"`
	TotalSize    int64                `json:"total_size"`
	Files        []TorrentInspectFile `json:"files"`
	Trackers     []string             `json:"trackers"`
	Comment      string               `json:"comment,omitempty"`
	CreatedBy    string               `json:"created_by,omitempty"`
	CreationDate int64                `json:"creation_date,omitempty"`
}

type TorrentInspectFile struct {
	Index  int    `json:"index"` // 1-based, aria2 select-file index
	Path   string `json:"path"`
	Length int64  `json:"length"`
}

// TorrentLiveStats is GET /api/download/<id>/torrent data.
type TorrentLiveStats struct {
	DownloadSpeed int64   `json:"download_speed"`
	UploadSpeed   int64   `json:"upload_speed"`
	UploadedBytes int64   `json:"uploaded_bytes"`
	ShareRatio    float64 `json:"share_ratio"`
	Connections   int64   `json:"connections"`
	NumSeeders    int64   `json:"num_seeders"`
	PiecesHave    int64   `json:"pieces_have"`
	NumPieces     int64   `json:"num_pieces"`
	VerifiedBytes int64   `json:"verified_length"`
	ETASeconds    int64   `json:"eta_seconds"`
	IsSeeding     bool    `json:"is_seeding"`
}

// TorrentPeers is GET /api/download/<id>/torrent/peers data.
type TorrentPeers struct {
	Peers []TorrentPeer `json:"peers"`
}

type TorrentPeer struct {
	PeerID        string  `json:"peer_id"`
	IP            string  `json:"ip"`
	Port          int     `json:"port"`
	DownloadSpeed int64   `json:"download_speed"`
	UploadSpeed   int64   `json:"upload_speed"`
	Progress      float64 `json:"progress"` // 0..1
	AmChoking     bool    `json:"am_choking"`
	PeerChoking   bool    `json:"peer_choking"`
	Seeder        bool    `json:"seeder"`
}

// TorrentInspectReq is POST /api/download/torrent/inspect body.
type TorrentInspectReq struct {
	TorrentFileB64 string `json:"torrent_file_b64"`
}

// SetTorrentFilesReq is PUT /api/download/<id>/torrent/files body.
// Selected is the full 1-based index list (not a delta); empty slice = all files.
type SetTorrentFilesReq struct {
	Selected []int `json:"selected"`
}

// SetTorrentFilesResult is the response data for PUT .../torrent/files.
type SetTorrentFilesResult struct {
	TaskID   int64 `json:"task_id"`
	Selected []int `json:"selected"`
}

// SeedControlResult is the response data for POST .../seed/stop|resume.
type SeedControlResult struct {
	TaskID int64  `json:"task_id"`
	Status string `json:"status"`
}

// FileExistsData is GET /api/url/file-exists data.
type FileExistsData struct {
	Exists       bool   `json:"exists"`
	ConflictPath string `json:"conflict_path,omitempty"`
}

// FileCheckResult mirrors GET /api/download/file_check, whose success body is
// top-level {code, exist} with NO data wrapper (note: "exist", not "exists").
type FileCheckResult struct {
	Exist bool `json:"exist"`
}

// RemoveActionResult is the -o json body for mutating verbs whose server
// response carries no payload (file remove, cookies delete). It mirrors, in
// machine-readable form, the plain-text acknowledgement printed in table mode
// so `--output json` is honoured consistently with the other verbs. Exactly one
// of Path / Domain is set, depending on the verb.
type RemoveActionResult struct {
	Removed bool   `json:"removed"`
	Path    string `json:"path,omitempty"`
	Domain  string `json:"domain,omitempty"`
}

// SyncResult holds GET /api/download/sync's success body, whose wire shape is
// the top-level {code, list, has_more} envelope (same "list" slot as the list
// endpoint, plus has_more) — NOT a {data:{items,next_cursor}} object. Rows come
// back in (updated_at, id) ascending order.
//
// The server uses a composite cursor (updated_at, id) via the `since` /
// `since_id` query params and does NOT echo a next_cursor: the client derives
// the next cursor from the last returned row (see NextCursor). Items maps from
// the envelope's "list" and is populated by doMutate's *SyncResult special case.
type SyncResult struct {
	Items   []DownloadTask `json:"items"`
	HasMore bool           `json:"has_more"`
}

// NextCursor returns the composite (updated_at, id) cursor to pass as
// --since / --since-id on the next page: the last returned row's values. It
// yields the zero time and 0 when the page is empty.
func (r SyncResult) NextCursor() (time.Time, int64) {
	if len(r.Items) == 0 {
		return time.Time{}, 0
	}
	last := r.Items[len(r.Items)-1]
	return last.UpdatedAt, last.ID
}

// CookieSummary is a row of GET /api/integration/cookies. It never carries the
// plaintext cookie (only whether one is stored).
type CookieSummary struct {
	Domain    string `json:"domain"`
	Provider  string `json:"provider"`
	HasCookie bool   `json:"has_cookie"`
	UpdatedAt int64  `json:"updated_at"` // unix seconds
}

// CookieListResult is the GET /api/integration/cookies success body
// (top-level {code, total, list} of CookieSummary).
type CookieListResult struct {
	List  []CookieSummary `json:"list"`
	Total int64           `json:"total"`
}

// UpsertCookieReq is PUT /api/integration/cookies body. Cookie is the full
// Netscape cookies.txt text.
type UpsertCookieReq struct {
	Domain   string `json:"domain"`
	Provider string `json:"provider,omitempty"`
	Cookie   string `json:"cookie"`
}

// RetrieveCookieReq is POST /api/integration/cookies/retrieve body.
type RetrieveCookieReq struct {
	Domain   string `json:"domain"`
	Provider string `json:"provider,omitempty"`
}

// RetrieveCookieResult is the retrieve response data; Cookie (Netscape text)
// is only populated when Found is true.
type RetrieveCookieResult struct {
	Domain    string `json:"domain"`
	Found     bool   `json:"found"`
	Cookie    string `json:"cookie,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

// IntegrationHealth is GET /api/integration/healthz data.
type IntegrationHealth struct {
	Healthy   bool              `json:"healthy"`
	Providers map[string]string `json:"providers"` // provider -> "ok" | "error: ..."
}

// systemSettingAria2MaxConcurrent is the only download-server system
// setting key today (manager models.SystemSettingAria2MaxConcurrent):
// aria2 max-concurrent-downloads, a bare JSON integer in [1, 16].
const systemSettingAria2MaxConcurrent = "aria2_max_concurrent"

// SystemSettings is the GET /api/system/settings snapshot (and the
// object echoed back by PUT). It mirrors the manager's typed response,
// which currently exposes a single global knob; the value is server
// authoritative (defaults + range clamping happen server-side).
type SystemSettings struct {
	Aria2MaxConcurrent int `json:"aria2_max_concurrent"`
}

// SystemSettingUpdateReq is the PUT /api/system/settings body. The
// manager applies exactly one key/value pair per request (a bad value
// for one knob must not clobber the others), so the CLI mirrors that
// shape rather than sending a whole-object patch. Value is whatever
// JSON the key expects (a bare integer for aria2_max_concurrent).
type SystemSettingUpdateReq struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}
