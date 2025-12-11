package controllers

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"bytetrade.io/web3os/tapr/pkg/utils"
	"bytetrade.io/web3os/tapr/pkg/vault/infisical"
	infisical_crypto "bytetrade.io/web3os/tapr/pkg/vault/infisical/crypto"
	"github.com/emicklei/go-restful"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

const WORKSPACE_KEY_SIZE_BYTES = 16

type workspaceController struct {
	Clientset func() *Clientset
}

type workspaceClient struct {
	mu sync.Mutex
}

func (w *workspaceClient) CreateWorkspace(user *infisical.UserEncryptionKeysPG, token, orgId, workspace, password string) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	workspaceId, err := w.GetWorkspace(token, orgId, workspace)
	if err == nil {
		// already exists
		return workspaceId, nil
	}

	if !w.IsNotFound(err) {
		return "", err
	}

	url := infisical.InfisicalAddr + "/api/v2/workspace"

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(map[string]*Workspace{}).
		SetBody(fiber.Map{
			"projectName": workspace,
			//			"organizationId": orgId,
		}).
		Post(url)

	if err != nil {
		klog.Error("create workspace error, ", err)
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("create workspace response error, ", resp.StatusCode(), ", ", string(resp.Body()))
		return "", errors.New(string(resp.Body()))
	}

	result := resp.Result().(*map[string]*Workspace)
	workspaceId = (*result)["project"].Id

	// unnecessary in v2
	// upload key to workspace to finish creating
	// err = w.uploadKeyToWorkspace(user, token, workspaceId, password)
	// if err != nil {
	// 	return "", err
	// }

	return workspaceId, nil
}

func (w *workspaceClient) GetWorkspace(token, orgId, workspace string) (string, error) {
	url := infisical.InfisicalAddr + "/api/v2/organizations/" + orgId + "/workspaces"

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&Workspaces{}).
		Get(url)

	if err != nil {
		klog.Error("get user workspace error, ", err)
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("get user workspace error, ", string(resp.Body()))
		return "", errors.New(string(resp.Body()))
	}

	workspaces := resp.Result().(*Workspaces)
	if len(workspaces.Items) == 0 {
		return "", errors.New("not found the workspaces of user org")
	}

	for _, w := range workspaces.Items {
		if w.Name == workspace {
			return w.Id, nil
		}
	}

	return "", fmt.Errorf("not found the workspace: %s", workspace)

}

func (w *workspaceClient) ListWorkspace(token, orgId string) ([]string, error) {
	url := infisical.InfisicalAddr + "/api/v2/organizations/" + orgId + "/workspaces"

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&Workspaces{}).
		Get(url)

	if err != nil {
		klog.Error("list user workspace error, ", err)
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("list user workspace error, ", string(resp.Body()))
		return nil, errors.New(string(resp.Body()))
	}

	workspaces := resp.Result().(*Workspaces)
	if len(workspaces.Items) == 0 {
		return nil, errors.New("not found the workspaces of user org")
	}

	var ret []string
	for _, w := range workspaces.Items {
		ret = append(ret, w.Name)
	}

	return ret, nil
}

func (w *workspaceClient) DeleteWorkspace(token, workspaceId string) error {
	url := infisical.InfisicalAddr + "/api/v1/workspace/" + workspaceId

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&Workspaces{}).
		Delete(url)

	if err != nil {
		klog.Error("delete user workspace error, ", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("delete user workspace error, ", string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil

}

func (w *workspaceClient) IsNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), "not found")
}

func (w *workspaceClient) uploadKeyToWorkspace(user *infisical.UserEncryptionKeysPG, token, workspaceId, privateKey string) error {
	key := make([]byte, WORKSPACE_KEY_SIZE_BYTES)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return err
	}

	ciphertext, nonce, err := infisical_crypto.EncryptAssymmetric(utils.Hex(key), user.PublicKey, privateKey)
	if err != nil {
		klog.Error("encrypt workspace key error, ", err)
		return err
	}

	url := infisical.InfisicalAddr + "/api/v1/key/" + workspaceId

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetBody(fiber.Map{
			"key": fiber.Map{
				"userId":       user.UserID,
				"encryptedKey": ciphertext,
				"nonce":        nonce,
			},
		}).
		Post(url)

	if err != nil {
		klog.Error("upload key error, ", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("upload key response error, ", string(resp.Body()))
		return errors.New(string(resp.Body()))
	}

	return nil
}

func (w *workspaceClient) GetWorkspaceKey(token, workspaceId, privateKey string) (string, error) {
	url := infisical.InfisicalAddr + fmt.Sprintf("/api/v2/workspace/%s/encrypted-key", workspaceId)

	client := NewHttpClient()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&EncryptedKey{}).
		Get(url)

	if err != nil {
		klog.Error("get user workspace key error, ", err)
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		klog.Error("get user workspace key response error, ", string(resp.Body()))
		return "", errors.New(string(resp.Body()))
	}

	key := resp.Result().(*EncryptedKey)

	plainKey, err := infisical_crypto.DecryptAsymmetric(key.Encryptedkey, key.Nonce, key.Sender.PublicKey, privateKey)
	if err != nil {
		klog.Error("decrypt project key error, ", err)
		return "", err
	}

	return plainKey, nil
}

// infisical shares the workspace with user in the same organization
// so we need to generate a unique workspace name for each user
func UserWorkspaceName(workflowName, user string) string {
	return fmt.Sprintf("%s-%s", workflowName, user)
}
