package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
	"github.com/gofiber/fiber/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

type userClient struct {
}

func (u *userClient) GetUserOrganizationId(token string) (string, error) {

	url := infisical.InfisicalAddr + "/api/v2/users/me/organizations"

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&Organizations{}).
		Get(url)

	if err != nil {
		klog.Error("get user organiztions error, ", err)
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("get user organiztions error, ", string(resp.Body()))
		return "", errors.New(string(resp.Body()))
	}

	orgs := resp.Result().(*Organizations)
	if len(orgs.Items) == 0 {
		return "", errors.New("user doesn't has organizations")
	}

	return orgs.Items[0].Id, nil
}

func (u *userClient) GetUserPrivateKey(user *infisical.UserEncryptionKeysPG, password string) (string, error) {
	return infisical.DecryptUserPrivateKeyHelper(user, password)
}

type userController struct {
	Clientset     func() *Clientset
	DynamicClient func() *dynamic.DynamicClient
}

func (user *userController) CreateUser(
	getPostgresUserAndPwd func(context.Context) (user, pwd string, err error),
	c *fiber.Ctx) error {
	klog.Info("received user create event")

	pguser, password, err := getPostgresUserAndPwd(c.Context())
	if err != nil {
		klog.Error("get postgres user and password error, ", err)
		return err
	}

	postUserInfo := &struct {
		Name  string `json:"name"`
		Role  string `json:"role"`
		Email string `json:"email"`
	}{}

	err = c.BodyParser(postUserInfo)
	if err != nil {
		klog.Error("body parser error,  ", err)
		return err
	}

	var client *infisical.PostgresClient

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", pguser, password, infisical.InfisicalDBAddr, infisical.InfisicalDBName)

	// try and wait for infisical postgres to connect
	func() {
		for {
			if client, err = infisical.NewClient(dsn); err != nil {
				klog.Info("connecting infisical postres error, ", err, ".  Waiting ... ")
				time.Sleep(time.Second)
			} else {
				return
			}
		}
	}()
	defer client.Close()

	// init user
	u, err := client.GetUser(c.Context(), postUserInfo.Email)
	if err != nil {
		klog.Error("get user from infisical error,  ", err)
		return err
	}

	if u == nil {
		err = infisical.InsertKsUserToPostgres(c.Context(), client, postUserInfo.Name, postUserInfo.Email, infisical.Password)
		if err != nil {
			klog.Error("init user error, ", err)
			return err
		}
	}

	return nil

}

func (user *userController) DeleteUser(
	getPostgresUserAndPwd func(context.Context) (user, pwd string, err error),
	c *fiber.Ctx) error {
	klog.Info("received user create event")

	pguser, password, err := getPostgresUserAndPwd(c.Context())
	if err != nil {
		klog.Error("get postgres user and password error, ", err)
		return err
	}

	postUserInfo := &struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{}

	if err := c.BodyParser(postUserInfo); err != nil {
		klog.Error("body parser error,  ", err)
		return err
	}

	var client *infisical.PostgresClient

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", pguser, password, infisical.InfisicalDBAddr, infisical.InfisicalDBName)

	// try and wait for infisical postgres to connect
	func() {
		for {
			if client, err = infisical.NewClient(dsn); err != nil {
				klog.Info("connecting infisical postres error, ", err, ".  Waiting ... ")
				time.Sleep(time.Second)
			} else {
				return
			}
		}
	}()
	defer client.Close()

	// find user
	u, err := client.GetUser(c.Context(), postUserInfo.Email)
	if err != nil {
		klog.Error("get user from infisical error,  ", err)
		return err
	}

	if u != nil {
		if err := infisical.DeleteUserFromPostgres(c.Context(), client, u.UserID); err != nil {
			klog.Error("delete user from postgres error,  ", err)
			return err
		}

		// delete the workspace and secret of the workspace
		token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
		if token == nil {
			return errors.New("no auth token")
		}

		username := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)
		orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)

		workspaces, err := user.getWorkspaces(c.UserContext(), token.(string), orgId, username)
		if err != nil {
			klog.Warning("list app permissions error, ", err, ", ", username)
		}

		if len(workspaces) == 0 {
			klog.Warning("no workspaces found for user,  ", username)
			return nil
		}

		for _, workspace := range workspaces {
			err := user.Clientset().workspaceClient.DeleteWorkspace(token.(string), workspace.workspaceId)
			if err != nil {
				klog.Warning(err)
			}
		}
	}

	return nil
}

func (u *userController) getWorkspaces(ctx context.Context, token, orgId, username string) ([]*struct{ workspaceName, workspaceId string }, error) {
	namespace := fmt.Sprintf("user-system-%s", username)
	permList, err := u.DynamicClient().Resource(appPermResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list app permissions error, ", err, ", ", namespace)
		return nil, err
	}

	if len(permList.Items) == 0 {
		return nil, nil
	}

	var workspaces []*struct{ workspaceName, workspaceId string }
	var errs []error
	for _, data := range permList.Items {
		var appperm ApplicationPermission
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(data.Object, &appperm)
		if err != nil {
			klog.Error("convert application permission error, ", err)
			return nil, err
		}

		for _, pr := range appperm.Spec.Permission {
			if pr.DataType == "secret" &&
				pr.Group == "secret.infisical" &&
				pr.Version == "v1" {
				if len(pr.Ops) > 0 {
					// assume one workspace per app
					op := DecodeOps(pr.Ops[0])
					if workspaceName, ok := op.Params["workspace"]; ok {
						userWorkspace := UserWorkspaceName(workspaceName, username)
						workspaceId, err := u.Clientset().GetWorkspace(token, orgId, userWorkspace)
						if err != nil {
							if !u.Clientset().IsNotFound(err) {
								return nil, err
							}
							errs = append(errs, err)
						}

						w := &struct {
							workspaceName string
							workspaceId   string
						}{
							userWorkspace,
							workspaceId,
						}

						workspaces = append(workspaces, w)
					}
				} // end of if ops
			} // end of if secret
		} // end of perms loop

	} // end of items loop

	if len(errs) > 0 {
		return workspaces, utilerrors.NewAggregate(errs)
	}

	return workspaces, nil
}
