package users

import (
	"context"

	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var gvr = schema.GroupVersionResource{
	Group:    iamV1alpha2.SchemeGroupVersion.Group,
	Version:  iamV1alpha2.SchemeGroupVersion.Version,
	Resource: iamV1alpha2.ResourcesPluralUser,
}

type ResourceUserClient struct {
	c *dynamic_client.ResourceClient[iamV1alpha2.User]
}

func NewResourceUserClient() (*ResourceUserClient, error) {
	ri, err := dynamic_client.NewResourceClient[iamV1alpha2.User](gvr)
	if err != nil {
		return nil, err
	}
	return &ResourceUserClient{c: ri}, nil
}

func NewResourceUserClientOrDie() *ResourceUserClient {
	c, err := NewResourceUserClient()
	if err != nil {
		panic(err)
	}
	return c
}

func (u *ResourceUserClient) Create(ctx context.Context, user iamV1alpha2.User, options metav1.CreateOptions) (*iamV1alpha2.User, error) {
	obj, err := dynamic_client.ToUnstructured(user)
	if err != nil {
		return nil, err
	}

	err = u.c.Create(ctx, &unstructured.Unstructured{Object: obj}, options, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *ResourceUserClient) Delete(ctx context.Context, name string, options metav1.DeleteOptions) error {
	return u.c.Delete(ctx, name, options)
}

func (u *ResourceUserClient) Update(ctx context.Context, user *iamV1alpha2.User, options metav1.UpdateOptions) (*iamV1alpha2.User, error) {
	return u.c.Update(ctx, user, options)
}

func (u *ResourceUserClient) List(ctx context.Context, options metav1.ListOptions) ([]*iamV1alpha2.User, error) {
	return u.c.List(ctx, options)
}

func (u *ResourceUserClient) Get(ctx context.Context, name string, options metav1.GetOptions) (*iamV1alpha2.User, error) {
	return u.c.Get(ctx, name, options)
}
