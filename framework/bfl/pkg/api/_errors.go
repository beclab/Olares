package api

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/utils"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

type ErrorType = string

const (
	ErrorInternalServerError ErrorType = "internal_server_error"
	ErrorInvalidGrant                  = "invalid_grant"
	ErrorBadRequest                    = "bad_request"
	ErrorUnknown                       = "unknown_error"
	ErrorIamOperator                   = "iam_operator"
)

// Avoid emitting errors that look like valid HTML. Quotes are okay.
var sanitizer = strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;")

func HandleInternalError(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusInternalServerError, response, req, err)
}

// HandleBadRequest writes http.StatusBadRequest and log error
func HandleBadRequest(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusBadRequest, response, req, err)
}

func HandleNotFound(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusNotFound, response, req, err)
}

func HandleForbidden(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusForbidden, response, req, err)
}

func HandleUnauthorized(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusUnauthorized, response, req, err)
}

func HandleTooManyRequests(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusTooManyRequests, response, req, err)
}

func HandleConflict(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusConflict, response, req, err)
}

func HandleError(response *restful.Response, req *restful.Request, err error) {
	var statusCode int
	switch t := err.(type) {
	case errors.APIStatus:
		statusCode = int(t.Status().Code)
	case restful.ServiceError:
		statusCode = t.Code
	default:
		statusCode = http.StatusInternalServerError
	}
	handle(statusCode, response, req, err)
}

func handle(statusCode int, resp *restful.Response, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(2)
	klog.Errorf("%s:%d %v", fn, line, err)

	var errType ErrorType
	var errDesc string

	if t, ok := err.(Error); ok {
		resp.WriteHeaderAndEntity(statusCode, t)
		return
	}

	switch statusCode {
	case http.StatusBadRequest:
		errType = ErrorBadRequest
	case http.StatusUnauthorized, http.StatusForbidden:
		errType = ErrorInvalidGrant
	case http.StatusInternalServerError:
		errType = ErrorInternalServerError
	default:
		errType = ErrorUnknown
	}
	errDesc = err.Error()
	resp.WriteHeaderAndEntity(statusCode, Error{
		ErrorType:        errType,
		ErrorDescription: errDesc,
	})
}

type Error struct {
	ErrorType        string `json:"error_type"`
	ErrorDescription string `json:"error_description"`
}

func (e Error) Error() string {
	return utils.PrettyJSON(e)
}

func NewError(t string, errs ...string) Error {
	var desc string
	if len(errs) > 0 {
		desc = errs[0]
	}
	return Error{ErrorType: t, ErrorDescription: desc}
}

func ErrorWithMessage(err error, message string) error {
	return fmt.Errorf("%v: %v", message, err.Error())
}

type ErrorMessage struct {
	Message string `json:"message"`
}

func (e ErrorMessage) Error() string {
	return e.Message
}

var ErrorNone = ErrorMessage{Message: "success"}
