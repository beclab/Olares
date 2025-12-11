package apitools

import (
	"fmt"
	"net/http"

	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/constants"
	"bytetrade.io/web3os/system-server/pkg/utils"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
)

type BaseHandler struct {
}

func (h *BaseHandler) GetUser(req *restful.Request) (string, error) {
	user := req.Request.Header.Get(constants.BflUserKey)
	if user == "" {
		err := restful.NewError(http.StatusUnauthorized, "User not found in request header")
		klog.Error(err)
		return "", err
	}

	return user, nil
}

func (h *BaseHandler) Validate(req *restful.Request, resp *restful.Response) (ok bool, username string) {
	account, err := h.GetUser(req)
	if err != nil {
		klog.Error(err)
		api.HandleUnauthorized(resp, req, err)
		return
	}

	var namespace string
	// get provider of user
	if isSA, saNamespace, _ := utils.IsServiceAccount(account); isSA {
		var isUserNamespace bool
		if isUserNamespace, username = utils.IsUserNamespace(saNamespace); !isUserNamespace {
			err := fmt.Errorf("user is not found in namespace %s", saNamespace)
			klog.Error(err)
			api.HandleUnauthorized(resp, req, err)
			return
		} // end of service account check

		namespace = saNamespace
	} else {
		username = account
		namespace = "user-system-" + username
	}

	klog.Infof("User %s is a system user in namespace %s", username, namespace)
	if constants.MyNamespace != namespace && constants.MyUserspace != namespace {
		api.HandleUnauthorized(resp, req, fmt.Errorf("invalid user, %s", username))
		return
	}

	ok = true
	return
}
