package controllers

import (
	"time"

	"github.com/go-resty/resty/v2"
)

type Clientset struct {
	workspaceClient
	userClient
	secretClient
}

func NewClientset() *Clientset {
	return &Clientset{}
}

func NewHttpClient() *resty.Client {
	return resty.New().SetTimeout(30 * time.Second)
}
