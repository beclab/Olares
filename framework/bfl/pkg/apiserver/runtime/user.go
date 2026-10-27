package runtime

import (
	"context"
	"fmt"
	"strings"

	"bytetrade.io/web3os/bfl/pkg/constants"
	"github.com/emicklei/go-restful/v3"
)

type userContextKey struct{}

func WithUser(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, userContextKey{}, username)
}

func UserFromContext(ctx context.Context) string {
	username, _ := ctx.Value(userContextKey{}).(string)
	return username
}

func RequestUser(req *restful.Request) (string, error) {
	username := UserFromContext(req.Request.Context())
	if username == "" {
		username, _ = req.Attribute(constants.UserContextAttribute).(string)
	}
	if username == "" {
		username = strings.TrimSpace(req.HeaderParameter(constants.HeaderBflUserKey))
	}
	if username == "" {
		return "", fmt.Errorf("request user is required")
	}
	return username, nil
}

func UserNamespace(username string) string {
	return fmt.Sprintf(constants.UserspaceNameFormat, username)
}
