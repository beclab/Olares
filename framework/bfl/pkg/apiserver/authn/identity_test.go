package authn

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bytetrade.io/web3os/bfl/pkg/constants"
)

func TestAuthenticate(t *testing.T) {
	originalUsername := constants.Username
	t.Cleanup(func() { constants.Username = originalUsername })
	constants.Username = "alice"

	tests := []struct {
		name       string
		headerUser string
		headerJWT  string
		cookieJWT  string
		parsedUser string
		parseErr   error
		want       Identity
		wantErr    error
	}{
		{
			name:       "trusted user header",
			headerUser: "alice",
			want:       Identity{Username: "alice"},
		},
		{
			name:       "authorization token",
			headerJWT:  "header-token",
			parsedUser: "alice",
			want:       Identity{Username: "alice", Token: "header-token"},
		},
		{
			name:       "matching trusted header and token",
			headerUser: "alice",
			headerJWT:  "header-token",
			parsedUser: "alice",
			want:       Identity{Username: "alice", Token: "header-token"},
		},
		{
			name:       "header and token users differ",
			headerUser: "alice",
			headerJWT:  "header-token",
			parsedUser: "bob",
			wantErr:    ErrUserMismatch,
		},
		{
			name:       "websocket cookie token",
			cookieJWT:  "cookie-token",
			parsedUser: "alice",
			want:       Identity{Username: "alice", Token: "cookie-token"},
		},
		{
			name:    "missing identity",
			wantErr: ErrUserNotFound,
		},
		{
			name:      "bad token",
			headerJWT: "bad-token",
			parseErr:  errors.New("invalid signature"),
			wantErr:   errors.New("parse token"),
		},
		{
			name:       "different bfl owner",
			headerUser: "bob",
			wantErr:    ErrUserMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://bfl.test/", nil)
			if tt.headerUser != "" {
				req.Header.Set(constants.HeaderBflUserKey, tt.headerUser)
			}
			if tt.headerJWT != "" {
				req.Header.Set(constants.UserAuthorizationTokenKey, tt.headerJWT)
			}
			if tt.cookieJWT != "" {
				req.AddCookie(&http.Cookie{Name: "auth_token", Value: tt.cookieJWT})
			}

			got, err := authenticate(req, func(token string) (string, error) {
				return tt.parsedUser, tt.parseErr
			})
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("authenticate() error = nil, want %v", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) && !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Fatalf("authenticate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("authenticate() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("authenticate() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
