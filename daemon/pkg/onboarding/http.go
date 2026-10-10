package onboarding

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
)

// Handler serves only onboarding routes. Basic authentication uses the initial
// password for this attempt, not the user's subsequent login password.
func Handler(service func(context.Context) (*Service, error)) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /onboarding", func(w http.ResponseWriter, r *http.Request) {
		var request CreateRequest
		if err := decode(w, r, &request); err != nil {
			respond(w, 400, err.Error(), nil)
			return
		}
		if err := request.validate(); err != nil {
			respond(w, 400, err.Error(), nil)
			return
		}
		s, err := service(r.Context())
		if err != nil {
			respond(w, 503, "system is not ready for onboarding", nil)
			return
		}
		status, err := s.Create(r.Context(), request)
		result(w, status, err)
	})
	mux.HandleFunc("GET /onboarding", func(w http.ResponseWriter, r *http.Request) {
		s, err := service(r.Context())
		if err != nil {
			respond(w, 503, "system is not ready for onboarding", nil)
			return
		}
		status, err := s.Status(r.Context())
		if err != nil {
			result(w, Status{}, err)
			return
		}
		if status.State == Waiting {
			result(w, status, nil)
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok {
			result(w, Status{}, ErrIdentity)
			return
		}
		if err := s.Authorize(r.Context(), username, password); err != nil {
			result(w, Status{}, err)
			return
		}
		status, err = s.Status(r.Context())
		result(w, status, err)
	})
	mux.HandleFunc("PUT /onboarding/password", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Password string `json:"password"`
		}
		if err := decode(w, r, &request); err != nil {
			respond(w, 400, err.Error(), nil)
			return
		}
		if len(request.Password) < 8 || len(request.Password) > 72 {
			respond(w, 400, "password must contain 8 to 72 bytes", nil)
			return
		}
		s, err := service(r.Context())
		if err != nil {
			respond(w, 503, "system is not ready for onboarding", nil)
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok {
			result(w, Status{}, ErrIdentity)
			return
		}
		status, err := s.Reset(r.Context(), username, password, request.Password)
		result(w, status, err)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

func decode(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return errors.New("invalid request body")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("invalid request body")
	}
	return nil
}

func result(w http.ResponseWriter, status Status, err error) {
	if err == nil {
		respond(w, 200, "", status)
		return
	}
	switch {
	case errors.Is(err, ErrPasswordUnchanged):
		respond(w, 400, err.Error(), nil)
	case errors.Is(err, ErrIdentity):
		respond(w, 401, err.Error(), nil)
	case errors.Is(err, ErrClosed):
		respond(w, 410, err.Error(), nil)
	case errors.Is(err, ErrConflict), errors.Is(err, ErrNotReady):
		respond(w, 409, err.Error(), nil)
	case apierrors.IsNotFound(err):
		respond(w, 404, "onboarding has not started", nil)
	default:
		klog.Error("onboarding operation failed: ", err)
		respond(w, 503, "onboarding operation failed; retry or check olaresd logs", nil)
	}
}

func respond(w http.ResponseWriter, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data,omitempty"`
	}{code, message, data})
}
