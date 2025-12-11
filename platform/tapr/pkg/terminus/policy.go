package terminus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"bytetrade.io/web3os/tapr/pkg/constants"
	"github.com/emicklei/go-restful"
	"github.com/go-resty/resty/v2"
	"k8s.io/klog/v2"
)

var (
	bflUrl = "bfl"
)

func init() {
	envBfl := os.Getenv("BFL")
	if envBfl != "" {
		bflUrl = envBfl
	}
}

func UpdatePolicy(ctx context.Context, uri, app, entrance, token, policy string) error {
	client := resty.New().SetTimeout(2 * time.Second)

	policyUrl := fmt.Sprintf("http://%s/bfl/settings/v1alpha1/applications/%s/setup/policy", bflUrl, app)
	resp, err := client.R().
		SetHeader(constants.AuthorizationTokenKey, token).
		SetHeader(restful.HEADER_Accept, restful.MIME_JSON).
		SetResult(&Response{Data: &map[string]*ApplicationSettingsPolicy{}}).
		Get(policyUrl)

	if err != nil {
		klog.Error("get app current policy error, ", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New(string(resp.Body()))
	}

	respData := resp.Result().(*Response)
	var currentPolicy map[string]*ApplicationSettingsPolicy
	var updatePolicy *ApplicationSettingsPolicy
	if respData.Code != 0 {
		klog.Error(string(respData.Message))

		// new policy
		updatePolicy = &ApplicationSettingsPolicy{
			DefaultPolicy: oneFactor,
			SubPolicies: []*ApplicationSettingsSubPolicy{
				{
					URI:    uri,
					Policy: policy,
				},
			},
		}
	} else {
		currentPolicy = *respData.Data.(*map[string]*ApplicationSettingsPolicy)
		found := false
		ok := false
		if updatePolicy, ok = currentPolicy[entrance]; !ok {
			updatePolicy = &ApplicationSettingsPolicy{
				DefaultPolicy: oneFactor,
				SubPolicies:   []*ApplicationSettingsSubPolicy{},
			}
		}

		for i, sub := range updatePolicy.SubPolicies {
			if sub.URI == uri {
				klog.Info("found uri policy, and update", uri)
				updatePolicy.SubPolicies[i].Policy = policy
				updatePolicy.SubPolicies[i].Duration = 0
				updatePolicy.SubPolicies[i].OneTime = false
				found = true
				break
			}
		}

		if !found {
			klog.Info("append a new policy for, ", uri)
			if updatePolicy.SubPolicies == nil {
				updatePolicy.SubPolicies = make([]*ApplicationSettingsSubPolicy, 0)
			}

			updatePolicy.SubPolicies = append(updatePolicy.SubPolicies, &ApplicationSettingsSubPolicy{
				URI:    uri,
				Policy: policy,
			})
		}

	}

	updatePolicyUrl := fmt.Sprintf("http://%s/bfl/settings/v1alpha1/applications/%s/%s/setup/policy", bflUrl, app, entrance)
	resp, err = client.R().
		SetHeader(constants.AuthorizationTokenKey, token).
		SetHeader(restful.HEADER_Accept, restful.MIME_JSON).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetResult(&Response{Data: &ApplicationSettingsPolicy{}}).
		SetBody(updatePolicy).
		Post(updatePolicyUrl)

	if err != nil {
		klog.Error("update app policy error, ", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New(string(resp.Body()))
	}

	respData = resp.Result().(*Response)
	if respData.Code != 0 {
		return errors.New(string(respData.Message))
	}

	return nil
}
