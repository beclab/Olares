package response

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
)

var successMsg = "success"

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

func errHandle(code int, w *restful.Response, err error) {
	var text string

	switch e := err.(type) {
	case TokenValidationError:
		code = tokenInvalidErrorCode
		text = e.Error()
	default:
		code = 1
		text = err.Error()
	}

	// code is all to 200
	w.WriteHeaderAndEntity(http.StatusOK, Header{
		Code:    code,
		Message: text,
	})
}

func HandleForbidden(w *restful.Response, err error) {
	errHandle(http.StatusForbidden, w, err)
}

func HandleError(w *restful.Response, err error) {
	errHandle(http.StatusInternalServerError, w, err)
}

// Success writes data to response with http.StatusOK.
func Success(w *restful.Response, v any) {
	w.WriteHeaderAndEntity(http.StatusOK, Response{
		Header: Header{
			Code:    0,
			Message: successMsg,
		},
		Data: v,
	})
}

// SuccessNoData writes to response with http.StatusOK but no data.
func SuccessNoData(w *restful.Response) {
	w.WriteHeaderAndEntity(http.StatusOK, Response{
		Header: Header{
			Code:    0,
			Message: successMsg,
		},
	})
}
