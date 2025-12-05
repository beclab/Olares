package tapr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"bytetrade.io/web3os/app-service/pkg/constants"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// MiddlewareType represents the type of middleware.
type MiddlewareType string

// describes the type of middleware support.
const (
	// TypePostgreSQL indicates the middleware is postgresql.
	TypePostgreSQL MiddlewareType = "postgres"
	// TypeMongoDB indicates the middleware is mongodb.
	TypeMongoDB MiddlewareType = "mongodb"
	// TypeRedis indicates the middleware is redis.
	TypeRedis MiddlewareType = "redis"
	// TypeNats indicates the middleware is nats
	TypeNats MiddlewareType = "nats"

	// TypeMinio indicates the middleware is minio
	TypeMinio MiddlewareType = "minio"

	// TypeRabbitMQ indicates the middleware is rabbitmq
	TypeRabbitMQ MiddlewareType = "rabbitmq"

	// TypeElasticsearch indicates the middleware is elasticsearch
	TypeElasticsearch MiddlewareType = "elasticsearch"

	// TypeMariaDB indicates the middleware is mariadb
	TypeMariaDB MiddlewareType = "mariadb"

	// TypeMySQL indicates the middleware is mysql
	TypeMySQL MiddlewareType = "mysql"
)

func (mr MiddlewareType) String() string {
	return string(mr)
}

// MiddlewareReq represents a request for a middleware.
type MiddlewareReq struct {
	App          string         `json:"app"`
	AppNamespace string         `json:"appNamespace"`
	Namespace    string         `json:"namespace"`
	Middleware   MiddlewareType `json:"middleware"`
}

// MetaInfo represents middleware meta info.
type MetaInfo struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// MiddlewareRequestInfo contains information for middlewarerequest.
type MiddlewareRequestInfo struct {
	MetaInfo
	App       MetaInfo   `json:"app"`
	UserName  string     `json:"username,omitempty"`
	Password  string     `json:"password"`
	Type      string     `json:"type"`
	Databases []Database `json:"databases,omitempty"`
	Buckets   []Bucket   `json:"buckets,omitempty"`
}

type MiddlewareRequestResp struct {
	MiddlewareRequestInfo
	Host         string            `json:"host"`
	Port         int32             `json:"port"`
	Indexes      map[string]string `json:"indexes"`
	Databases    map[string]string `json:"databases"`
	Buckets      map[string]string `json:"buckets"`
	Vhosts       map[string]string `json:"vhosts"`
	Subjects     map[string]string `json:"subjects"`
	Refs         map[string]string `json:"refs"`
	BucketPrefix string            `json:"bucketPrefix"`
	IndexPrefix  string            `json:"indexPrefix"`
}

type Resp struct {
	Code int                    `json:"code"`
	Data *MiddlewareRequestResp `json:"data"`
}

// Apply middlewarerequest, get response and set values.
func Apply(middleware *Middleware, kubeConfig *rest.Config, appName, appNamespace,
	namespace, token, ownerName string, vals map[string]interface{}) error {
	if middleware == nil {
		return nil
	}
	client := resty.New()
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		statusCode := r.StatusCode()
		return statusCode == 404 || statusCode == 429 || statusCode >= 500
	})
	getMiddlewareRequest := func(middlewareType MiddlewareType) (*MiddlewareRequestResp, error) {
		url := "http://middleware-service.os-platform/middleware/v1/request/info"
		request := MiddlewareReq{
			App:          appName,
			AppNamespace: appNamespace,
			Namespace:    namespace,
			Middleware:   middlewareType,
		}
		resp, err := client.SetTimeout(1*time.Second).R().
			SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
			SetAuthToken(token).
			SetHeader(constants.BflUserKey, ownerName).
			SetBody(request).Post(url)
		if err != nil {
			klog.Errorf("Failed to make middleware request middlewareType=%s err=%v", middlewareType, err)
			return nil, err
		}
		if resp.StatusCode() != 200 {
			klog.Errorf("Failed to get middleware request response status=%s body=%s", resp.Status(), resp.String())
			return nil, errors.New(resp.String())
		}
		var middlewareRequestResp Resp
		err = json.Unmarshal(resp.Body(), &middlewareRequestResp)
		if err != nil {
			klog.Errorf("Failed to unmarshal middleware request response err=%v", err)
			return nil, err
		}

		if middlewareRequestResp.Code != 200 {
			return nil, errors.New(resp.String())
		}

		return middlewareRequestResp.Data, nil
	}

	if middleware.Postgres != nil {
		username := fmt.Sprintf("%s_%s_%s", middleware.Postgres.Username, ownerName, appName)
		username = strings.ReplaceAll(username, "-", "_")
		err := process(kubeConfig, appName, appNamespace, namespace, username, TypePostgreSQL, ownerName, middleware)
		if err != nil {
			return err
		}
		resp, err := getMiddlewareRequest(TypePostgreSQL)
		if err != nil {
			return err
		}
		vals["postgres"] = map[string]interface{}{
			"host":      resp.Host,
			"port":      resp.Port,
			"username":  resp.UserName,
			"password":  resp.Password,
			"databases": resp.Databases,
		}
	}

	if middleware.Redis != nil {
		username := ""
		err := process(kubeConfig, appName, appNamespace, namespace, username, TypeRedis, ownerName, middleware)
		if err != nil {
			return err
		}
		resp, err := getMiddlewareRequest(TypeRedis)
		if err != nil {
			return err
		}
		vals["redis"] = map[string]interface{}{
			"host":       resp.Host,
			"port":       resp.Port,
			"password":   resp.Password,
			"namespaces": resp.Databases,
		}
	}

	if middleware.MongoDB != nil {
		username := fmt.Sprintf("%s-%s-%s", middleware.MongoDB.Username, ownerName, appName)
		err := process(kubeConfig, appName, appNamespace, namespace, username, TypeMongoDB, ownerName, middleware)
		if err != nil {
			return err
		}
		resp, err := getMiddlewareRequest(TypeMongoDB)
		if err != nil {
			return err
		}
		vals["mongodb"] = map[string]interface{}{
			"host":      resp.Host,
			"port":      resp.Port,
			"username":  resp.UserName,
			"password":  resp.Password,
			"databases": resp.Databases,
		}
	}
	if middleware.Minio != nil {
		username := fmt.Sprintf("%s-%s-%s", middleware.Minio.Username, ownerName, appName)
		err := process(kubeConfig, appName, appNamespace, namespace, username, TypeMinio, ownerName, middleware)
		if err != nil {
			return err
		}

		resp, err := getMiddlewareRequest(TypeMinio)
		if err != nil {
			return err
		}
		vals["minio"] = map[string]interface{}{
			"host":         resp.Host,
			"port":         resp.Port,
			"username":     resp.UserName,
			"password":     resp.Password,
			"buckets":      resp.Buckets,
			"bucketPrefix": resp.BucketPrefix,
		}
	}
	if middleware.RabbitMQ != nil {
		username := fmt.Sprintf("%s-%s-%s", middleware.RabbitMQ.Username, ownerName, appName)
		err := process(kubeConfig, appName, appNamespace, namespace, username, TypeRabbitMQ, ownerName, middleware)
		if err != nil {
			return err
		}
		resp, err := getMiddlewareRequest(TypeRabbitMQ)
		if err != nil {
			klog.Errorf("failed to get rabbitmq middleware request info %v", err)
			return err
		}
		vals["rabbitmq"] = map[string]interface{}{
			"host":     resp.Host,
			"port":     resp.Port,
			"username": resp.UserName,
			"password": resp.Password,
			"vhosts":   resp.Vhosts,
		}
		klog.Infof("values.rabbitmq: %v", vals["rabbitmq"])
	}

	if middleware.Elasticsearch != nil {
		username := fmt.Sprintf("%s-%s-%s", middleware.Elasticsearch.Username, ownerName, appName)
		err := process(kubeConfig, appName, appNamespace, namespace, username, TypeElasticsearch, ownerName, middleware)
		if err != nil {
			return err
		}
		resp, err := getMiddlewareRequest(TypeElasticsearch)
		if err != nil {
			klog.Errorf("failed to get elasticsearch middleware request info %v", err)
			return err
		}
		vals["elasticsearch"] = map[string]interface{}{
			"host":        resp.Host,
			"port":        resp.Port,
			"username":    resp.UserName,
			"password":    resp.Password,
			"indexes":     resp.Indexes,
			"indexPrefix": resp.IndexPrefix,
		}
		klog.Infof("values.elasticsearch: %v", vals["elasticsearch"])
	}

	if middleware.Nats != nil {
		username := fmt.Sprintf("%s-%s", middleware.Nats.Username, appNamespace)
		err := process(kubeConfig, appName, appNamespace, namespace, username, TypeNats, ownerName, middleware)
		if err != nil {
			return err
		}
		resp, err := getMiddlewareRequest(TypeNats)
		if err != nil {
			return err
		}
		vals["nats"] = map[string]interface{}{
			"host":     resp.Host,
			"port":     resp.Port,
			"username": resp.UserName,
			"password": resp.Password,
			"subjects": resp.Subjects,
			"refs":     resp.Refs,
		}
		klog.Infof("vals[nats]: %v", vals["nats"])
	}

	if middleware.MariaDB != nil {
		klog.Infof("middleware.MariaDB: %#v", middleware.MariaDB)
		username := fmt.Sprintf("%s-%s-%s", middleware.MariaDB.Username, ownerName, appName)
		err := process(kubeConfig, appName, appNamespace, namespace, username, TypeMariaDB, ownerName, middleware)
		if err != nil {
			return err
		}
		resp, err := getMiddlewareRequest(TypeMariaDB)
		if err != nil {
			klog.Errorf("failed to get mariadb middleware request info %v", err)
			return err
		}
		vals["mariadb"] = map[string]interface{}{
			"host":      resp.Host,
			"port":      resp.Port,
			"username":  resp.UserName,
			"password":  resp.Password,
			"databases": resp.Databases,
		}
		klog.Infof("values.mariadb: %v", vals["mariadb"])
	}

	if middleware.MySQL != nil {
		klog.Errorf("middleware.mYSQL: %#v", middleware.MySQL)
		username := fmt.Sprintf("%s-%s-%s", middleware.MySQL.Username, ownerName, appName)
		err := process(kubeConfig, appName, appNamespace, namespace, username, TypeMySQL, ownerName, middleware)
		if err != nil {
			return err
		}
		resp, err := getMiddlewareRequest(TypeMySQL)
		if err != nil {
			klog.Errorf("failed to get mysql middleware request info %v", err)
			return err
		}
		vals["mysql"] = map[string]interface{}{
			"host":      resp.Host,
			"port":      resp.Port,
			"username":  resp.UserName,
			"password":  resp.Password,
			"databases": resp.Databases,
		}
		klog.Infof("values.mysql: %v", vals["mysql"])
	}
	return nil
}

func process(kubeConfig *rest.Config, appName, appNamespace, namespace, username string,
	middlewareType MiddlewareType, ownerName string, middleware *Middleware) error {
	params := RequestParams{
		MiddlewareType: middlewareType,
		AppName:        appName,
		AppNamespace:   appNamespace,
		Namespace:      namespace,
		Username:       username,
		OwnerName:      ownerName,
		Middleware:     middleware,
	}
	password, err := getPassword(middleware, middlewareType)
	if err != nil {
		klog.Error(err)
		return err
	}
	params.Password = password
	request, err := GenMiddleRequest(params)
	if err != nil {
		klog.Errorf("Failed to generate middleware request from template middlewareType=%s err=%v", middlewareType, err)
		return err
	}
	if len(params.Password) == 0 {
		err = CreateOrUpdateSecret(kubeConfig, appName, namespace, middlewareType)
		klog.Errorf("create secret:appname: %s,namespace:%s, middlewareType:%s,err: %v", appName, namespace, middlewareType, err)
		if err != nil {
			return err
		}
	}
	_, err = CreateOrUpdateMiddlewareRequest(kubeConfig, namespace, request)
	if err != nil {
		klog.Errorf("Failed to create or update middleware request middlewareType=%s err=%v", middlewareType, err)
		return err
	}
	return nil
}

func getPassword(middleware *Middleware, middlewareType MiddlewareType) (string, error) {
	switch middlewareType {
	case TypePostgreSQL:
		return middleware.Postgres.Password, nil
	case TypeRedis:
		return middleware.Redis.Password, nil
	case TypeMongoDB:
		return middleware.MongoDB.Password, nil
	case TypeNats:
		return middleware.Nats.Password, nil
	case TypeMinio:
		return middleware.Minio.Password, nil
	case TypeRabbitMQ:
		return middleware.RabbitMQ.Password, nil
	case TypeElasticsearch:
		return middleware.Elasticsearch.Password, nil
	case TypeMariaDB:
		return middleware.MariaDB.Password, nil
	case TypeMySQL:
		return middleware.MySQL.Password, nil
	}
	return "", fmt.Errorf("unsupported middleware type %v", middlewareType)
}
