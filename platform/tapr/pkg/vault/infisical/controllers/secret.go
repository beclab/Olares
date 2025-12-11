package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
	infisical_crypto "bytetrade.io/web3os/tapr/pkg/vault/infisical/crypto"
	"github.com/emicklei/go-restful"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type secretController struct {
	Clientset func() *Clientset
}

func (s *secretController) CreateSecret(c *fiber.Ctx) error {
	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	body := c.Request().Body()
	var secret Secret
	err := json.Unmarshal(body, &secret)
	if err != nil {
		klog.Error("decode request error: ", err)
		return err
	}

	reqWorkspaceName := c.Query("workspace")
	if reqWorkspaceName == "" {
		return c.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "invalid request, none of the workspace param",
		})
	}

	user := c.Context().UserValueBytes(constants.UserCtxKey).(*infisical.UserEncryptionKeysPG)
	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)
	userName := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)

	userWorkspace := UserWorkspaceName(reqWorkspaceName, userName)
	workspaceId, err := s.Clientset().GetWorkspace(token.(string), orgId, userWorkspace)
	if err != nil {
		if !s.Clientset().IsNotFound(err) {
			return c.JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": "get user workspace error, " + err.Error(),
			})
		}

		workspaceId, err = s.Clientset().CreateWorkspace(user, token.(string), orgId, userWorkspace, userPrivateKey)
		if err != nil {
			return c.JSON(fiber.Map{
				"code":    http.StatusInternalServerError,
				"message": "create user workspace error, " + err.Error(),
			})
		}
	}

	projectKey, err := s.Clientset().GetWorkspaceKey(token.(string), workspaceId, userPrivateKey)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get user workspace key error, " + err.Error(),
		})
	}

	err = s.Clientset().CreateSecretInWorkspace(user, token.(string),
		workspaceId, projectKey,
		secret.Name, secret.Value, secret.Environment)

	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "create user secret error, " + err.Error(),
		})
	}

	return nil
}

func (s *secretController) RetrieveSecret(c *fiber.Ctx) error {
	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	body := c.Request().Body()
	var secret Secret
	err := json.Unmarshal(body, &secret)
	if err != nil {
		klog.Error("decode request error: ", err)
		return err
	}

	reqWorkspaceName := c.Query("workspace")
	if reqWorkspaceName == "" {
		return c.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "invalid request, none of the workspace param",
		})
	}

	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)
	userName := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)

	userWorkspace := UserWorkspaceName(reqWorkspaceName, userName)

	workspaceId, err := s.Clientset().GetWorkspace(token.(string), orgId, userWorkspace)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get user workspace error, " + err.Error(),
		})
	}

	projectKey, err := s.Clientset().GetWorkspaceKey(token.(string), workspaceId, userPrivateKey)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get user workspace key error, " + err.Error(),
		})
	}

	var resData Secret
	resData.Environment = secret.Environment
	resData.Name, resData.Value, err = s.Clientset().GetSecretInWorkspace(token.(string), workspaceId, projectKey,
		secret.Environment, secret.Name)

	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "retrieve secret key error, " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    StatusOK,
		"message": "",
		"data":    resData,
	})
}

func (s *secretController) ListSecret(c *fiber.Ctx) error {
	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	body := c.Request().Body()
	var secret Secret
	err := json.Unmarshal(body, &secret)
	if err != nil {
		klog.Error("decode request error: ", err)
		return err
	}

	reqWorkspaceName := c.Query("workspace")
	if reqWorkspaceName == "" {
		return c.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "invalid request, none of the workspace param",
		})
	}

	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)
	userName := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)

	userWorkspace := UserWorkspaceName(reqWorkspaceName, userName)

	workspaceId, err := s.Clientset().GetWorkspace(token.(string), orgId, userWorkspace)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get user workspace error, " + err.Error(),
		})
	}

	projectKey, err := s.Clientset().GetWorkspaceKey(token.(string), workspaceId, userPrivateKey)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get user workspace key error, " + err.Error(),
		})
	}

	var resData []*Secret
	list, err := s.Clientset().ListSecretInWorkspace(token.(string), workspaceId, projectKey,
		secret.Environment)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "list secret keys error, " + err.Error(),
		})
	}

	for _, nv := range list {
		resData = append(resData, &Secret{
			Environment: secret.Environment,
			Name:        nv.name,
			Value:       nv.value,
		})
	}

	return c.JSON(fiber.Map{
		"code":    StatusOK,
		"message": "",
		"data":    resData,
	})
}

func (s *secretController) DeleteSecret(c *fiber.Ctx) error {
	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	body := c.Request().Body()
	var secret Secret
	err := json.Unmarshal(body, &secret)
	if err != nil {
		klog.Error("decode request error: ", err)
		return err
	}

	reqWorkspaceName := c.Query("workspace")
	if reqWorkspaceName == "" {
		return c.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "invalid request, none of the workspace param",
		})
	}

	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)
	userName := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)

	userWorkspace := UserWorkspaceName(reqWorkspaceName, userName)

	workspaceId, err := s.Clientset().GetWorkspace(token.(string), orgId, userWorkspace)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get user workspace error, " + err.Error(),
		})
	}

	projectKey, err := s.Clientset().GetWorkspaceKey(token.(string), workspaceId, userPrivateKey)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get user workspace key error, " + err.Error(),
		})
	}

	err = s.Clientset().DeleteSecretInWorkspace(token.(string), workspaceId, projectKey,
		secret.Environment, secret.Name)

	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "delete secret error, " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    StatusOK,
		"message": "delete secret succeed",
	})
}

func (s *secretController) UpdateSecret(c *fiber.Ctx) error {
	token := c.Context().UserValueBytes(constants.UserAuthTokenCtxKey)
	if token == nil {
		return errors.New("no auth token")
	}

	body := c.Request().Body()
	var secret Secret
	err := json.Unmarshal(body, &secret)
	if err != nil {
		klog.Error("decode request error: ", err)
		return err
	}

	reqWorkspaceName := c.Query("workspace")
	if reqWorkspaceName == "" {
		return c.JSON(fiber.Map{
			"code":    http.StatusBadRequest,
			"message": "invalid request, none of the workspace param",
		})
	}

	user := c.Context().UserValueBytes(constants.UserCtxKey).(*infisical.UserEncryptionKeysPG)
	orgId := c.Context().UserValueBytes([]byte(constants.UserOrganizationIdCtxKey)).(string)
	userPrivateKey := c.Context().UserValueBytes([]byte(constants.UserPrivateKeyCtxKey)).(string)
	userName := c.Context().UserValueBytes([]byte(constants.UsernameCtxKey)).(string)

	userWorkspace := UserWorkspaceName(reqWorkspaceName, userName)

	workspaceId, err := s.Clientset().GetWorkspace(token.(string), orgId, userWorkspace)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get user workspace error, " + err.Error(),
		})
	}

	projectKey, err := s.Clientset().GetWorkspaceKey(token.(string), workspaceId, userPrivateKey)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "get user workspace key error, " + err.Error(),
		})
	}

	err = s.Clientset().UpdateSecretInWorkspace(user, token.(string),
		workspaceId, projectKey,
		secret.Name, secret.Value, secret.Environment)

	if err != nil {
		return c.JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"message": "update user secret error, " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    StatusOK,
		"message": "update secret succeed",
	})
}

type secretClient struct {
}

func (s *secretClient) CreateSecretInWorkspace(user *infisical.UserEncryptionKeysPG, token, workspaceId, projectKey, secretName, secretValue, env string) error {
	url := infisical.InfisicalAddr + "/api/v3/secrets/" + secretName

	secretKeyCiphertext, secretKeyIV, secretKeyTag, err := infisical_crypto.Encrypt(secretName, projectKey)
	if err != nil {
		klog.Error("encrypt secret key error, ", err)
		return err
	}

	secretValueCiphertext, secretValueIV, secretValueTag, err := infisical_crypto.Encrypt(secretValue, projectKey)
	if err != nil {
		klog.Error("encrypt secret value error, ", err)
		return err
	}

	if env == "" {
		env = DefaulSecretEnv
	}
	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetBody(fiber.Map{
			"workspaceId":           workspaceId,
			"environment":           env,
			"type":                  "shared",
			"secretKeyCiphertext":   secretKeyCiphertext,
			"secretKeyIV":           secretKeyIV,
			"secretKeyTag":          secretKeyTag,
			"secretValueCiphertext": secretValueCiphertext,
			"secretValueIV":         secretValueIV,
			"secretValueTag":        secretValueTag,
		}).
		Post(url)

	if err != nil {
		klog.Error("create secret error, ", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("create secret response error, ", string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}

func (s *secretClient) GetSecretInWorkspace(token, workspaceId, projectKey, env, secretName string) (name, value string, err error) {
	url := infisical.InfisicalAddr + "/api/v3/secrets/" + secretName
	if env == "" {
		env = DefaulSecretEnv
	}

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetQueryParams(map[string]string{
			"environment": env,
			"workspaceId": workspaceId,
			"type":        "shared",
		}).
		SetResult(&map[string]fiber.Map{}).
		Get(url)

	if err != nil {
		klog.Error("get secret error, ", err)
		return "", "", err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("get secret response error, ", string(resp.Body()))
		return "", "", errors.New(string(resp.Body()))
	}

	result := resp.Result().(*map[string]fiber.Map)

	valueOf := func(k string) string {
		if v, ok := (*result)["secret"][k]; !ok {
			return ""
		} else {
			return v.(string)
		}
	}

	name, err = infisical_crypto.Decrypt(
		valueOf("secretKeyCiphertext"),
		valueOf("secretKeyIV"),
		valueOf("secretKeyTag"),
		projectKey,
	)
	if err != nil {
		klog.Error("decrypt secret key error, ", err)
		return "", "", err
	}

	value, err = infisical_crypto.Decrypt(
		valueOf("secretValueCiphertext"),
		valueOf("secretValueIV"),
		valueOf("secretValueTag"),
		projectKey,
	)
	if err != nil {
		klog.Error("decrypt secret key error, ", err)
		return "", "", err
	}

	return
}

func (s *secretClient) ListSecretInWorkspace(token, workspaceId, projectKey, env string) (secrets []*struct{ name, value string }, err error) {
	url := infisical.InfisicalAddr + "/api/v3/secrets/"
	if env == "" {
		env = DefaulSecretEnv
	}

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetQueryParams(map[string]string{
			"environment": env,
			"workspaceId": workspaceId,
			"type":        "shared",
		}).
		SetResult(&map[string][]*fiber.Map{}).
		Get(url)

	if err != nil {
		klog.Error("list secret error, ", err)
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("list secret response error, ", string(resp.Body()))
		return nil, errors.New(string(resp.Body()))
	}

	resultList := resp.Result().(*map[string][]*fiber.Map)

	for _, result := range (*resultList)["secrets"] {
		valueOf := func(k string) string {
			if v, ok := (*result)[k]; !ok {
				return ""
			} else {
				return v.(string)
			}
		}

		name, err := infisical_crypto.Decrypt(
			valueOf("secretKeyCiphertext"),
			valueOf("secretKeyIV"),
			valueOf("secretKeyTag"),
			projectKey,
		)
		if err != nil {
			klog.Error("decrypt secret key error, ", err)
			return nil, err
		}

		value, err := infisical_crypto.Decrypt(
			valueOf("secretValueCiphertext"),
			valueOf("secretValueIV"),
			valueOf("secretValueTag"),
			projectKey,
		)
		if err != nil {
			klog.Error("decrypt secret key error, ", err)
			return nil, err
		}

		secrets = append(secrets, &struct {
			name  string
			value string
		}{name, value})
	}

	return
}

func (s *secretClient) DeleteSecretInWorkspace(token, workspaceId, projectKey, env, secretName string) error {
	url := infisical.InfisicalAddr + "/api/v3/secrets/" + secretName
	if env == "" {
		env = DefaulSecretEnv
	}

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetBody(fiber.Map{
			"environment": env,
			"workspaceId": workspaceId,
			"type":        "shared",
		}).
		Delete(url)

	if err != nil {
		klog.Error("delete secret error, ", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("delete secret response error, ", string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}

func (s *secretClient) UpdateSecretInWorkspace(user *infisical.UserEncryptionKeysPG, token, workspaceId, projectKey, secretName, secretValue, env string) error {
	url := infisical.InfisicalAddr + "/api/v3/secrets/" + secretName

	secretValueCiphertext, secretValueIV, secretValueTag, err := infisical_crypto.Encrypt(secretValue, projectKey)
	if err != nil {
		klog.Error("encrypt secret value error, ", err)
		return err
	}

	if env == "" {
		env = DefaulSecretEnv
	}
	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetBody(fiber.Map{
			"workspaceId":           workspaceId,
			"environment":           env,
			"type":                  "shared",
			"secretValueCiphertext": secretValueCiphertext,
			"secretValueIV":         secretValueIV,
			"secretValueTag":        secretValueTag,
		}).
		Patch(url)

	if err != nil {
		klog.Error("update secret error, ", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("update secret response error, ", string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}
