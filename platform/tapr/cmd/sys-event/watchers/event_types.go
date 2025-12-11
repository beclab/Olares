package watchers

type PermissionRequire struct {
	Group    string   `json:"group"`
	DataType string   `json:"dataType"`
	Version  string   `json:"version"`
	Ops      []string `json:"ops"`
}

type AccessTokenRequest struct {
	AppKey    string            `json:"app_key"`
	Timestamp int64             `json:"timestamp"`
	Token     string            `json:"token"`
	Perm      PermissionRequire `json:"perm"`
}

type Header struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Header

	Data any `json:"data,omitempty"` // data field, optional, object or list
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
}

type AccessTokenResp struct {
	Header

	Data AccessToken `json:"data,omitempty"`
}

// event types

type Event struct {
	Type    string    `json:"type"`
	Version string    `json:"version"`
	Data    EventData `json:"data"`
}

type EventData struct {
	Message string      `json:"msg"`
	Payload interface{} `json:"payload"`
}
