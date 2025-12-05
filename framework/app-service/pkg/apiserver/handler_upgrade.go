package apiserver

type ResultResponse struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"message"`
}
