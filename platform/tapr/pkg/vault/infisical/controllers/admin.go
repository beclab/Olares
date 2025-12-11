package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
	"github.com/gofiber/fiber/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

type adminController struct {
	Clientset     func() *Clientset
	DynamicClient func() *dynamic.DynamicClient
}

type secretData struct {
	Key   string `json:"Key"`
	Value string `json:"Value,omitempty"`
}

const (
	ParamWorkspace = "workspace"
	ParamApp       = "appid"
)

var appPermResource = schema.GroupVersionResource{
	Group:    "sys.bytetrade.io",
	Version:  "v1alpha1",
	Resource: "applicationpermissions",
}

var ErrorNoPerm error = errors.New("app didn't has perm on infisical secret")

func (a *adminController) CheckAppSecretPerm(c *fiber.Ctx) error {
	app := c.Params(ParamApp)
	if app == "" {
		return errors.New("app name is empty")
	}

	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	username := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)
	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)

	w, err := a.getAppSecretWorkspaces(c.UserContext(), token.(string), orgId, userPrivateKey, username, app)
	if err != nil && err != ErrorNoPerm {
		klog.Error("list app permissions error, ", err, ", ", app)

		if w == nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": "check app secret permission error, " + err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"code": StatusOK,
		"data": map[string]bool{
			"permission": err != ErrorNoPerm && len(w) > 0,
		},
	})
}

func (a *adminController) ListAppSecret(c *fiber.Ctx) error {
	app := c.Params(ParamApp)
	if app == "" {
		return errors.New("app name is empty")
	}

	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	username := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)
	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)

	workspaces, err := a.getAppSecretWorkspaces(c.UserContext(), token.(string), orgId, userPrivateKey, username, app)
	if err != nil {
		klog.Error("list app permissions error, ", err, ", ", app)
		if workspaces == nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": "get app workspace key error, " + err.Error(),
			})
		}
	}

	if len(workspaces) == 0 {
		return c.JSON(fiber.Map{
			"code": StatusOK,
			"data": []interface{}{},
		})
	}

	var secrets []*struct{ Workspace, Key string }
	for _, w := range workspaces {
		if w.workspaceId == "" {
			continue
		}

		workspaceSecrets, err := a.Clientset().ListSecretInWorkspace(token.(string), w.workspaceId, w.projectKey, DefaulSecretEnv)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": "list secres in app workspace error, " + err.Error(),
			})
		}

		for _, s := range workspaceSecrets {
			secrets = append(secrets, &struct {
				Workspace string
				Key       string
			}{w.workspaceName, s.name})
		}
	}

	return c.JSON(fiber.Map{
		"code": StatusOK,
		"data": secrets,
	})
}

func (a *adminController) CreateAppSecret(c *fiber.Ctx) error {
	app := c.Params(ParamApp)
	if app == "" {
		return errors.New("app name is empty")
	}

	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	body := c.Request().Body()
	var secret secretData
	err := json.Unmarshal(body, &secret)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "send data invalid, " + err.Error(),
		})
	}

	user := c.Context().UserValueBytes(constants.UserCtxKey).(*infisical.UserEncryptionKeysPG)
	username := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)
	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)

	workspaces, err := a.getAppSecretWorkspaces(c.UserContext(), token.(string), orgId, userPrivateKey, username, app)
	if err != nil {
		klog.Warning("list app permissions error, ", err, ", ", app)
	}

	if len(workspaces) == 0 {
		return c.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": "app has not permission to create secret",
		})
	}

	defaultWorkspace := workspaces[0]
	if defaultWorkspace.workspaceId == "" {
		defaultWorkspace.workspaceId, err = a.Clientset().CreateWorkspace(user, token.(string), orgId, defaultWorkspace.workspaceName, userPrivateKey)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": "create user workspace error, " + err.Error(),
			})
		}

		defaultWorkspace.projectKey, err = a.Clientset().GetWorkspaceKey(token.(string), defaultWorkspace.workspaceId, userPrivateKey)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": "get user workspace key error, " + err.Error(),
			})
		}

	}
	err = a.Clientset().CreateSecretInWorkspace(user, token.(string),
		defaultWorkspace.workspaceId, defaultWorkspace.projectKey,
		secret.Key, secret.Value, DefaulSecretEnv,
	)

	if err != nil {
		klog.Error("create secret error, ", err, ", ", secret.Key, ", ", secret.Value)
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "create secret error, " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    StatusOK,
		"message": "create secret success",
	})
}

func (a *adminController) DeleteAppSecret(c *fiber.Ctx) error {
	app := c.Params(ParamApp)
	if app == "" {
		return errors.New("app name is empty")
	}

	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	body := c.Request().Body()
	var secret secretData
	err := json.Unmarshal(body, &secret)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "send data invalid, " + err.Error(),
		})
	}

	username := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)
	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)

	workspaces, err := a.getAppSecretWorkspaces(c.UserContext(), token.(string), orgId, userPrivateKey, username, app)
	if err != nil {
		klog.Error("list app permissions error, ", err, ", ", app)
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get app workspace key error, " + err.Error(),
		})
	}

	if len(workspaces) == 0 {
		return c.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": "app has not permission to delete secret",
		})
	}

	defaultWorkspace := workspaces[0]
	err = a.Clientset().DeleteSecretInWorkspace(token.(string),
		defaultWorkspace.workspaceId, defaultWorkspace.projectKey,
		DefaulSecretEnv, secret.Key,
	)

	if err != nil {
		klog.Error("delete secret error, ", err, ", ", secret.Key)
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "delete secret error, " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    StatusOK,
		"message": "delete secret success",
	})
}

func (a *adminController) UpdateAppSecret(c *fiber.Ctx) error {
	app := c.Params(ParamApp)
	if app == "" {
		return errors.New("app name is empty")
	}

	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	body := c.Request().Body()
	var secret secretData
	err := json.Unmarshal(body, &secret)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "send data invalid, " + err.Error(),
		})
	}

	user := c.Context().UserValueBytes(constants.UserCtxKey).(*infisical.UserEncryptionKeysPG)
	username := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)
	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)

	workspaces, err := a.getAppSecretWorkspaces(c.UserContext(), token.(string), orgId, userPrivateKey, username, app)
	if err != nil {
		klog.Error("list app permissions error, ", err, ", ", app)
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get app workspace key error, " + err.Error(),
		})
	}

	if len(workspaces) == 0 {
		return c.JSON(fiber.Map{
			"code":    http.StatusNotFound,
			"message": "app has not permission to delete secret",
		})
	}

	defaultWorkspace := workspaces[0]
	err = a.Clientset().UpdateSecretInWorkspace(user, token.(string),
		defaultWorkspace.workspaceId, defaultWorkspace.projectKey,
		secret.Key, secret.Value, DefaulSecretEnv,
	)

	if err != nil {
		klog.Error("update secret error, ", err, ", ", secret.Key)
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "update secret error, " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    StatusOK,
		"message": "update secret success",
	})
}

func (a *adminController) getAppSecretWorkspaces(ctx context.Context,
	token, orgId, userPrivateKey, username, appName string) ([]*struct{ workspaceName, workspaceId, projectKey string }, error) {
	namespace := fmt.Sprintf("user-system-%s", username)
	permList, err := a.DynamicClient().Resource(appPermResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list app permissions error, ", err, ", ", appName, ", ", namespace)
		return nil, err
	}

	if len(permList.Items) == 0 {
		return nil, ErrorNoPerm
	}

	var workspaces []*struct{ workspaceName, workspaceId, projectKey string }
	var errs []error
	foundPerm := false
	for _, data := range permList.Items {
		var appperm ApplicationPermission
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(data.Object, &appperm)
		if err != nil {
			klog.Error("convert application permission error, ", err)
			return nil, err
		}

		if appperm.Spec.App == appName {
			foundPerm = true
			for _, pr := range appperm.Spec.Permission {
				if pr.DataType == "secret" &&
					pr.Group == "secret.infisical" &&
					pr.Version == "v1" {
					if len(pr.Ops) > 0 {
						// assume one workspace per app
						op := DecodeOps(pr.Ops[0])
						if workspaceName, ok := op.Params["workspace"]; ok {
							userWorkspace := UserWorkspaceName(workspaceName, username)
							w := &struct {
								workspaceName string
								workspaceId   string
								projectKey    string
							}{
								userWorkspace,
								"",
								"",
							}

							workspaces = append(workspaces, w)
							workspaceId, err := a.Clientset().GetWorkspace(token, orgId, userWorkspace)
							if err != nil {
								if !a.Clientset().IsNotFound(err) {
									return nil, err
								}
								errs = append(errs, err)
							}

							if workspaceId != "" {
								projectKey, err := a.Clientset().GetWorkspaceKey(token, workspaceId, userPrivateKey)
								if err != nil {
									return nil, err
								}

								w.workspaceId = workspaceId
								w.projectKey = projectKey
							}
						}
					}
				}
			}
		}
	}

	if len(errs) > 0 {
		return workspaces, utilerrors.NewAggregate(errs)
	}

	if !foundPerm {
		return nil, ErrorNoPerm
	}

	return workspaces, nil
}
