/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package im

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	resources "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type IdentityManagementInterface interface {
	CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
	ListUsers(query *query.Query) (*api.ListResult, error)
	DeleteUser(username string) error
	UpdateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error)
	DescribeUser(username string) (*iamv1alpha2.User, error)
	ModifyPassword(username string, password string) error
	ListLLdapUsers(query *query.Query) (*api.ListResult, error)
	ListLLdapGroups(query *query.Query) (*api.ListResult, error)
}

func NewOperator(ksClient kubesphere.Interface, userGetter resources.Interface) IdentityManagementInterface {
	im := &imOperator{
		ksClient:   ksClient,
		userGetter: userGetter,
	}
	return im
}

const syncToLLdapKey = "iam.kubesphere.io/sync-to-lldap"

type LLdapUser struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Groups   []string `json:"groups"`
}

type imOperator struct {
	ksClient   kubesphere.Interface
	userGetter resources.Interface
}

// UpdateUser returns user information after update.
func (im *imOperator) UpdateUser(new *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	old, err := im.fetch(new.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	// keep encrypted password and user status
	status := old.Status
	// only support enable or disable
	if new.Status.State == iamv1alpha2.UserDisabled || new.Status.State == iamv1alpha2.UserActive {
		status.State = new.Status.State
		status.LastTransitionTime = &metav1.Time{Time: time.Now()}
	}
	new.Status = status
	updated, err := im.ksClient.IamV1alpha2().Users().Update(context.Background(), new, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return ensurePasswordNotOutput(updated), nil
}

func (im *imOperator) fetch(username string) (*iamv1alpha2.User, error) {
	obj, err := im.userGetter.Get("", username)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	user := obj.(*iamv1alpha2.User).DeepCopy()
	return user, nil
}

func (im *imOperator) ModifyPassword(username string, password string) error {
	user, err := im.fetch(username)
	if err != nil {
		klog.Error(err)
		return err
	}
	_, err = im.ksClient.IamV1alpha2().Users().Update(context.Background(), user, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (im *imOperator) ListUsers(query *query.Query) (result *api.ListResult, err error) {
	result, err = im.userGetter.List("", query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	items := make([]interface{}, 0)
	for _, item := range result.Items {
		user := item.(*iamv1alpha2.User)
		out := ensurePasswordNotOutput(user)
		items = append(items, out)
	}
	result.Items = items
	return result, nil
}
func (im *imOperator) ListLLdapUsers(query *query.Query) (*api.ListResult, error) {
	result := new(api.ListResult)
	users, err := im.userGetter.List("", query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	items := make([]interface{}, 0)
	for _, item := range users.Items {
		user := item.(*iamv1alpha2.User)
		// filter user that do not sync to lldap
		if user.Annotations[syncToLLdapKey] != "true" {
			continue
		}
		if len(user.Spec.Groups) == 0 {
			user.Spec.Groups = make([]string, 0)
		}

		out := LLdapUser{Username: user.Name, Email: user.Spec.Email, Groups: user.Spec.Groups}
		items = append(items, out)
	}
	result.Items = items
	result.TotalItems = len(items)
	return result, nil
}

func (im *imOperator) ListLLdapGroups(query *query.Query) (*api.ListResult, error) {
	result := new(api.ListResult)
	users, err := im.userGetter.List("", query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	items := make([]interface{}, 0)
	for _, item := range users.Items {
		user := item.(*iamv1alpha2.User)
		// filter user that do not sync to lldap
		if user.Annotations[syncToLLdapKey] != "true" {
			continue
		}
		for i := range user.Spec.Groups {
			items = append(items, user.Spec.Groups[i])
		}
	}
	result.Items = items
	result.TotalItems = len(items)
	return result, nil
}

func (im *imOperator) DescribeUser(username string) (*iamv1alpha2.User, error) {
	obj, err := im.userGetter.Get("", username)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	user := obj.(*iamv1alpha2.User)
	return ensurePasswordNotOutput(user), nil
}

func (im *imOperator) DeleteUser(username string) error {
	return im.ksClient.IamV1alpha2().Users().Delete(context.Background(), username, *metav1.NewDeleteOptions(0))
}

func (im *imOperator) CreateUser(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	user, err := im.ksClient.IamV1alpha2().Users().Create(context.Background(), user, metav1.CreateOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return user, nil
}

func ensurePasswordNotOutput(user *iamv1alpha2.User) *iamv1alpha2.User {
	out := user.DeepCopy()
	// ensure encrypted password will not be output
	return out
}
