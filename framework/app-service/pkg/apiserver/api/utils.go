package api

import (
	"net/http"
	"runtime"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Avoid emitting errors that look like valid HTML. Quotes are okay.
var sanitizer = strings.NewReplacer(`&`, "&amp;", `<`, "&lt;", `>`, "&gt;")

// HandleInternalError writes http.StatusInternalServerError and log error.
func HandleInternalError(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusInternalServerError, response, req, err)
}

// HandleBadRequest writes http.StatusBadRequest and log error.
func HandleBadRequest(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusBadRequest, response, req, err)
}

// HandleNotFound writes http.StatusNotFound and log error.
func HandleNotFound(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusNotFound, response, req, err)
}

// HandleForbidden writes http.StatusForbidden and log error.
func HandleForbidden(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusForbidden, response, req, err)
}

// HandleUnauthorized writes http.StatusUnauthorized and log error.
func HandleUnauthorized(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusUnauthorized, response, req, err)
}

// HandleTooManyRequests writes http.StatusTooManyRequests and log error.
func HandleTooManyRequests(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusTooManyRequests, response, req, err)
}

// HandleConflict writes http.StatusConflict and log error.
func HandleConflict(response *restful.Response, req *restful.Request, err error) {
	handle(http.StatusConflict, response, req, err)
}

func HandleFailedCheck(response *restful.Response, checkType string, checkResult any, code int) {
	response.WriteHeaderAndEntity(http.StatusUnprocessableEntity, FailedCheckResponse{Code: code, Data: FailedCheckResponseData{Type: checkType, Data: checkResult}})
}

// HandleError handles the given error by determining the appropriate HTTP status code and performing error handling logic.
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

func handle(statusCode int, response *restful.Response, req *restful.Request, err error) {
	_, fn, line, _ := runtime.Caller(2)
	ctrl.Log.Error(err, "response error", "func", fn, "line", line)
	http.Error(response, sanitizer.Replace(err.Error()), statusCode)
}
