package response

import (
	"fmt"
	"net/http"
	"strings"

	"bytetrade.io/web3os/bfl/internal/log"
	"github.com/emicklei/go-restful/v3"
)

var SuccessMsg = "success"

var UnexpectedError = "unexpected error"

type Header struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Header

	Data any `json:"data,omitempty"` // data field, optional, object or list
}

func (s Header) Error() string {
	return fmt.Sprintf("ServiceError[%v]:%v", s.Code, s.Message)
}

func NewHeader(code int, msg ...string) Header {
	var text string
	if len(msg) > 0 {
		text = strings.Join(msg, ", ") + "."
	} else {
		text = http.StatusText(code)
	}
	return Header{Code: code, Message: text}
}

func errHandle(code int, w *restful.Response, err error) {
	var text string

	switch e := err.(type) {
	case TokenValidationError:
		code = TokenInvalidErrorCode
		text = e.Error()
	default:
		if strings.HasPrefix(e.Error(), "Unauthorized:") {
			code = TokenInvalidErrorCode
		} else {
			code = 1
		}
		text = err.Error()
	}

	log.Errorf("%+v", err)

	// code is all to 200
	w.WriteHeaderAndEntity(http.StatusOK, Header{
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
	errHandle(http.StatusInternalServerError, w, err)
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
