package response

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"olares.com/backup-server/pkg/util/log"
)

var SuccessMsg = "success"

type Header struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Header

	Data any `json:"data,omitempty"` // data field, optional, object or list
}

func errHandle(statusCode int, w *restful.Response, err error) {
	var code int
	var text string

	switch e := err.(type) {
	case TokenError: // capture custom error type
		code = TokenInvalidErrCode
		text = e.Error()
	default:
		code = 1
		text = err.Error()
	}

	log.Errorf("%v", err)

	w.WriteHeaderAndEntity(statusCode, Header{
		Code:    code,
		Message: text,
	})
}

func HandleBadRequest(w *restful.Response, err error) {
	errHandle(http.StatusBadRequest, w, err)
}

func HandleNotFound(w *restful.Response, err error) {
	errHandle(http.StatusNotFound, w, err)
}

func HandleInternalError(w *restful.Response, err error) {
	errHandle(http.StatusInternalServerError, w, err)
}

func HandleForbidden(w *restful.Response, err error) {
	errHandle(http.StatusForbidden, w, err)
}

func HandleUnauthorized(w *restful.Response, err error) {
	errHandle(http.StatusUnauthorized, w, err)
}

func HandleTooManyRequests(w *restful.Response, err error) {
	errHandle(http.StatusTooManyRequests, w, err)
}

func HandleConflict(w *restful.Response, err error) {
	errHandle(http.StatusConflict, w, err)
}

func HandleError(w *restful.Response, err error) {
	// errHandle(http.StatusInternalServerError, w, err)
	w.WriteHeaderAndEntity(http.StatusOK, Response{
		Header: Header{
			Code:    1,
			Message: err.Error(),
		},
	})
}

func Error(w *restful.Response, v any, err error) {
	w.WriteHeaderAndEntity(http.StatusOK, Response{
		Header: Header{
			Code:    0,
			Message: err.Error(),
		},
		Data: v,
	})
}

func Success(w *restful.Response, v any) {
	w.WriteHeaderAndEntity(http.StatusOK, Response{
		Header: Header{
			Code:    0,
			Message: SuccessMsg,
		},
		Data: v,
	})
}

func SuccessNoData(w *restful.Response) {
	w.WriteHeaderAndEntity(http.StatusOK, Response{
		Header: Header{
			Code:    0,
			Message: SuccessMsg,
		},
	})
}
