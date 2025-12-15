package lldap

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"github.com/beclab/lldap-client/pkg/cache/memory"
	lclient "github.com/beclab/lldap-client/pkg/client"
	lconfig "github.com/beclab/lldap-client/pkg/config"
	"github.com/beclab/lldap-client/pkg/generated"
	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

	"fmt"
	"github.com/go-resty/resty/v2"
	"k8s.io/klog"
	"kubesphere.io/api/iam/v1alpha2"
	"net/http"
	"time"
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

func createGraphClient(url, token string) graphql.Client {
	httpClient := http.Client{
		Transport: &authedTransport{
			key:     token,
			wrapped: http.DefaultTransport,
		},
	}
	return graphql.NewClient(url+"/api/graphql", &httpClient)
}

type LLdapOperator struct {
	*v1alpha2.LLdapProvider
	*kubernetes.Clientset
	LLdapClient *lclient.Client
	restClient  *resty.Client
}

func New(lldap *v1alpha2.LLdapProvider) (*LLdapOperator, error) {
	operator := &LLdapOperator{
		LLdapProvider: lldap,
		restClient:    resty.New().SetTimeout(5 * time.Second),
	}

	klog.Infof("init operator clientset...")
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	operator.Clientset, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	klog.Infof("init operator clientset sucess...")

	bindUsername, err := operator.getCredentialVal("lldap-ldap-user-dn")
	if err != nil {
		return nil, err
	}
	klog.V(0).Infof("bindUsername=%v", bindUsername)
	bindPassword, err := operator.getCredentialVal("lldap-ldap-user-pass")
	if err != nil {
		return nil, err
	}
	klog.V(0).Infof("bindPassword=%v", bindPassword)
	klog.V(0).Infof("lldap.URL=%v", lldap.URL)
	operator.LLdapClient, err = lclient.New(&lconfig.Config{
		Host:       lldap.URL,
		Username:   bindUsername,
		Password:   bindPassword,
		TokenCache: memory.New(),
	})
	if err != nil {
		return nil, err
	}

	return operator, nil
}

func (l *LLdapOperator) getCredentialVal(key string) (string, error) {
	if l.CredentialsSecret != nil {
		secret, err := l.Clientset.CoreV1().Secrets(l.CredentialsSecret.Namespace).Get(context.TODO(), l.CredentialsSecret.Name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		if value, ok := secret.Data[key]; ok {
			return string(value), nil
		}
	}
	return "", fmt.Errorf("can not find credentialval for key %s", key)
}

func (l *LLdapOperator) GetUser(name string) (*User, error) {
	userDetails, err := l.LLdapClient.Users().Get(context.TODO(), name)
	if err != nil {
		return nil, err
	}

	groups := make([]string, 0)
	for _, g := range userDetails.Groups {
		groups = append(groups, g.DisplayName)
	}

	user := &User{
		Id:          userDetails.Id,
		DisplayName: userDetails.DisplayName,
		Groups:      groups,
	}
	return user, nil
}

func (l *LLdapOperator) GetUserList() ([]User, error) {
	klog.V(0).Infof("lldapclient1111: %v", l)

	klog.V(0).Infof("lldapclient: %v", l.LLdapClient)
	groupList, err := l.LLdapClient.Groups().List(context.TODO())
	klog.V(0).Infof("GetUserList, err=%v", err)
	if err != nil {
		return nil, err
	}

	users := l.filter(groupList)
	return users, nil
}

func (l *LLdapOperator) filter(groups []generated.GetGroupListGroupsGroup) []User {
	users := make([]User, 0)
	userMap := make(map[string]User)
	klog.Infof("filter: group:%v, user:%v", l.GroupWhitelist, l.UserBlacklist)
	for _, group := range groups {
		if len(l.GroupWhitelist) != 0 && !funk.Contains(l.GroupWhitelist, group.DisplayName) {
			continue
		}
		for _, user := range group.Users {
			if len(l.UserBlacklist) != 0 && funk.Contains(l.UserBlacklist, user.Id) {
				continue
			}
			if v, ok := userMap[user.Id]; ok {
				v.Groups = append(v.Groups, group.DisplayName)
			} else {
				userMap[user.Id] = User{
					Id:          user.Id,
					Email:       user.Email,
					DisplayName: user.DisplayName,
					Groups:      []string{group.DisplayName},
				}
			}
		}
	}

	for _, user := range userMap {
		users = append(users, user)
	}
	return users
}
