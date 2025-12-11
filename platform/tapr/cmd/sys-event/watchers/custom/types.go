package custom

type CustomEvent struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Message string      `json:"msg"`
	User    string      `json:"user"`
}
