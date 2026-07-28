package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func newResp() (*restful.Response, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)
	return resp, rec
}

func newReq() *restful.Request {
	return restful.NewRequest(httptest.NewRequest(http.MethodGet, "/x", nil))
}

func TestHandleErrorStatusMapping(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"plain error -> 500", errors.New("boom"), http.StatusInternalServerError},
		{"apistatus notfound -> 404", apierrors.NewNotFound(schema.GroupResource{Group: "app", Resource: "applications"}, "x"), http.StatusNotFound},
		{"service error -> code", restful.ServiceError{Code: http.StatusBadRequest, Message: "bad"}, http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, rec := newResp()
			HandleError(resp, newReq(), c.err)
			if rec.Code != c.want {
				t.Errorf("status=%d want %d", rec.Code, c.want)
			}
		})
	}
}

func TestHandleErrorSanitizesAngleBrackets(t *testing.T) {
	resp, rec := newResp()
	HandleError(resp, newReq(), errors.New("<script>alert(1)</script>"))
	body := rec.Body.String()
	// The sanitizer replaces angle brackets with HTML entities; the JSON
	// encoder then escapes the ampersand, so no raw < or > may survive.
	if strings.ContainsAny(body, "<>") {
		t.Errorf("response body contains raw angle brackets: %s", body)
	}
	if !strings.Contains(body, "lt;script") {
		t.Errorf("expected escaped markup, got: %s", body)
	}
}

func TestHandleNotFoundAndBadRequest(t *testing.T) {
	resp, rec := newResp()
	HandleNotFound(resp, newReq(), errors.New("missing"))
	if rec.Code != http.StatusNotFound {
		t.Errorf("HandleNotFound status=%d want 404", rec.Code)
	}

	resp2, rec2 := newResp()
	HandleBadRequest(resp2, newReq(), errors.New("bad"))
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("HandleBadRequest status=%d want 400", rec2.Code)
	}
}
