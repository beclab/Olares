package infisical

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId         string `json:"userId"`
	AuthTokenType  string `json:"authTokenType"`
	TokenVersionId string `json:"tokenVersionId"`
	AccessVersion  *int   `json:"accessVersion,omitempty"`
	RefreshVersion *int   `json:"refreshVersion,omitempty"`
	OrganizationId string `json:"organizationId"`
	AuthMethod     string `json:"authMethod"`
}

type tokenIssuer struct {
	kubeconfig    *rest.Config
	getUserAndPwd func(ctx context.Context) (user string, pwd string, err error)
}

func NewTokenIssuer(kubeconfig *rest.Config) *tokenIssuer {
	return &tokenIssuer{kubeconfig: kubeconfig}
}

func (t *tokenIssuer) WithUserAndPwd(f func(ctx context.Context) (user string, pwd string, err error)) *tokenIssuer {
	t.getUserAndPwd = f
	return t
}

func (t *tokenIssuer) IssueInfisicalToken(next func(c *fiber.Ctx) error) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		// get user email from ctx
		email, ok := c.Context().UserValueBytes([]byte(constants.UserEmailCtxKey)).(string)
		if !ok {
			return c.JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": "auth user email is invalid",
				"data":    nil,
			})
		}

		ctx := c.UserContext()
		user, session, err := t.getUserFromInfisicalPostgres(ctx, email)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": fmt.Sprintf("get user from infisical error, %s, %s", err.Error(), email),
				"data":    nil,
			})
		}
		c.Context().SetUserValueBytes(constants.UserCtxKey, user)
		c.Context().SetUserValueBytes(constants.UserOrganizationIdCtxKey, *user.OrgId)
		uid := user.UserID
		klog.Info("get user id, ", uid)

		authKey, _, err := t.getJwtSecret(ctx)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": fmt.Sprintf("get user jwt key error, %s", err.Error()),
				"data":    nil,
			})
		}

		refreshToken, err := t.issueToken(uid, authKey, 10*24*time.Hour, "refreshToken", *session.ID, *user.OrgId)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": fmt.Sprintf("unable to sign refresh token, %s", err.Error()),
				"data":    nil,
			})
		}
		c.Context().SetUserValueBytes(constants.UserRefreshTokenCtxKey, refreshToken)

		// client := resty.New().SetTimeout(10 * time.Second)
		// type Token struct {
		// 	Token string `json:"token"`
		// }

		// res, err := client.R().
		// 	SetCookie(&http.Cookie{Name: "jid", Value: refreshToken}).
		// 	SetResult(&Token{}).
		// 	Post(InfisicalAddr + "/api/v1/auth/token")

		// if err != nil {
		// 	klog.Error("refresh access token err, ", err)
		// 	return c.JSON(fiber.Map{
		// 		"code":    http.StatusUnauthorized,
		// 		"message": fmt.Sprintf("unable to sign auth token, %s", err.Error()),
		// 		"data":    nil,
		// 	})
		// }

		// if res.StatusCode() != http.StatusOK {
		// 	err = errors.New(string(res.Body()))
		// 	klog.Error("refresh access token return code err, ", res.StatusCode(), ", ", err)

		// 	return c.JSON(fiber.Map{
		// 		"code":    http.StatusUnauthorized,
		// 		"message": fmt.Sprintf("unable to sign auth token, %s", err.Error()),
		// 		"data":    nil,
		// 	})
		// }

		// authToken := res.Result().(*Token).Token
		authToken, err := t.issueToken(uid, authKey, 10*24*time.Hour, "accessToken", *session.ID, *user.OrgId)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusUnauthorized,
				"message": fmt.Sprintf("unable to sign auth token, %s", err.Error()),
				"data":    nil,
			})
		}
		c.Context().SetUserValueBytes(constants.UserAuthTokenCtxKey, authToken)
		// klog.Info("get user token, ", authToken)

		return next(c)
	}
}

func (t *tokenIssuer) getUserFromInfisicalMongoDB(ctx context.Context, email string) (*UserMDB, error) {
	user, password, err := t.getUserAndPwd(ctx)
	if err != nil {
		return nil, err
	}
	mongo := MongoClient{
		User:     user,
		Password: password,
		Database: InfisicalDBName,
		Addr:     InfisicalDBAddr,
	}

	return mongo.GetUser(ctx, email)
}

func (t *tokenIssuer) getUserFromInfisicalPostgres(ctx context.Context, email string) (*UserEncryptionKeysPG, *AuthTokenSessionsPG, error) {
	user, password, err := t.getUserAndPwd(ctx)
	if err != nil {
		return nil, nil, err
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, InfisicalDBAddr, InfisicalDBName)

	pg, err := NewClient(dsn)
	if err != nil {
		return nil, nil, err
	}

	defer pg.Close()

	userEnc, err := pg.GetUser(ctx, email)
	if err != nil {
		klog.Error("get user encrypted key error, ", err)
		return nil, nil, err
	}

	if userEnc == nil {
		klog.Error("get user encrypted key is nil, ", email)
		return nil, nil, fmt.Errorf("user %s not found", email)
	}

	session, err := pg.GetUserTokenSession(ctx, userEnc.UserID, "localhost", "tapr-sidecar")
	if err != nil {
		klog.Error("get user token session error, ", err)
		return nil, nil, err
	}

	return userEnc, session, nil
}

func (t *tokenIssuer) getJwtSecret(ctx context.Context) (authKey string, refreshKey string, err error) {
	client, err := kubernetes.NewForConfig(t.kubeconfig)
	if err != nil {
		return "", "", err
	}

	backendSecret, err := client.CoreV1().Secrets(InfisicalNamespace).Get(ctx, "infisical-backend", metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	authKey = string(backendSecret.Data["JWT_AUTH_SECRET"])
	refreshKey = string(backendSecret.Data["JWT_REFRESH_SECRET"])

	return authKey, refreshKey, nil
}

func (t *tokenIssuer) issueToken(userId string, key string, expireIn time.Duration, tokenType, sessionId, orgId string) (string, error) {
	var (
		av, rv *int
	)

	switch tokenType {
	case "accessToken":
		av = pointer.Int(1)
	case "refreshToken":
		rv = pointer.Int(1)
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims{
		UserId:         userId,
		AuthTokenType:  tokenType,
		TokenVersionId: sessionId,
		AccessVersion:  av,
		RefreshVersion: rv,
		OrganizationId: orgId,
		AuthMethod:     "email",
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(expireIn).Unix(),
		},
	}).SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return token, nil
}
