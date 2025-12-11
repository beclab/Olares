package app

import aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"

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

type MiddlewareReq struct {
	App          string               `json:"app"`
	AppNamespace string               `json:"appNamespace"`
	Namespace    string               `json:"namespace"`
	Middleware   aprv1.MiddlewareType `json:"middleware"`
}

type Database struct {
	Name        string `json:"name"`
	Distributed bool   `json:"distributed,omitempty"`
}

type Bucket struct {
	Name string `json:"name"`
}

type Vhost struct {
	Name string `json:"name"`
}

type Index struct {
	Name string `json:"name"`
}

type MetaInfo struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type MiddlewareRequestInfo struct {
	MetaInfo
	App       MetaInfo             `json:"app"`
	UserName  string               `json:"username,omitempty"`
	Password  string               `json:"password"`
	Type      aprv1.MiddlewareType `json:"type"`
	Databases []Database           `json:"databases,omitempty"`
	Buckets   []Bucket             `json:"buckets,omitempty"`
	Indexes   []Index              `json:"indexes,omitempty"`
	Vhosts    []Vhost              `json:"vhosts,omitempty"`
	Subjects  []aprv1.Subject      `json:"subjects,omitempty"`
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
	BucketPrefix string            `json:"bucketPrefix,omitempty"`
	IndexPrefix  string            `json:"indexPrefix,omitempty"`
}

type Proxy struct {
	Endpoint string `json:"endpoint"`
	Size     int32  `json:"size"`
}

type MiddlewareClusterResp struct {
	MetaInfo
	Nodes          int32          `json:"nodes"`
	AdminUser      string         `json:"adminUser"`
	Password       string         `json:"password"`
	Mongos         Proxy          `json:"mongos,omitempty"`
	RedisProxy     Proxy          `json:"redisProxy,omitempty"`
	Proxy          Proxy          `json:"proxy,omitempty"`
	MiddlewareType MiddlewareType `json:"type"`
}

type ClusterScaleReq struct {
	MetaInfo
	Middleware aprv1.MiddlewareType `json:"middleware"`
	Nodes      int32                `json:"nodes"`
}

type ClusterChangePwdReq struct {
	MetaInfo
	Middleware aprv1.MiddlewareType `json:"middleware"`

	User     string `json:"user,omitempty"`
	Password string `json:"password"`
}
