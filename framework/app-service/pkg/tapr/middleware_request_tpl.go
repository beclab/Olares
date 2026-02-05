package tapr

import (
	"bytes"
	"fmt"
	"text/template"
)

const postgresRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-postgres
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: postgres
  postgreSQL:
    databases:
    {{- range $k, $v := .Middleware.Databases }}
    - distributed: {{ $v.Distributed }}
      name: {{ $v.Name }}
      {{- if gt (len $v.Extensions) 0 }}
      extensions:
      {{- range $i, $ext := $v.Extensions }}
      - {{ $ext }}
      {{- end }}
      {{- end }}
      {{- if gt (len $v.Scripts) 0 }}
      scripts:
      {{- range $i, $s := $v.Scripts }}
      - '{{ $s }}'
      {{- end }}
      {{- end }}
      
    {{- end }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-postgres-password
          key: "password"
	 {{- end }}
    user: {{ .Middleware.Username }}
`

const redisRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-redis
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: redis
  redis:
    namespace: {{ .Middleware.Namespace }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-redis-password
          key: "password"
	 {{- end }}
`

const mongodbRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-mongodb
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: mongodb
  mongodb:
    databases:
    {{- range $k, $v := .Middleware.Databases }}
    - name: {{ $v.Name }}
	{{- if gt (len $v.Scripts) 0 }}
      scripts:
      {{- range $i, $s := $v.Scripts }}
      - '{{ $s }}'
      {{- end }}
    {{- end }}
    {{- end }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-mongodb-password
          key: "password"
	 {{- end }}
    user: {{ .Middleware.Username }}
`

const mariadbRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-mariadb
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: mariadb
  mariadb:
    databases:
    {{- range $k, $v := .Middleware.Databases }}
    - name: {{ $v.Name }}
    {{- end }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-mariadb-password
          key: "password"
     {{- end }}
    user: {{ .Middleware.Username }}
`

const mysqlRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-mysql
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: mysql
  mysql:
    databases:
    {{- range $k, $v := .Middleware.Databases }}
    - name: {{ $v.Name }}
    {{- end }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-mysql-password
          key: "password"
     {{- end }}
    user: {{ .Middleware.Username }}
`

const natsRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-nats
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: nats
  nats:
    user: {{ .Middleware.Username }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-nats-password
          key: "password"
     {{- end }}
    {{- if gt (len .Middleware.Subjects) 0 }}
    subjects: 
      {{- range $k, $v := .Middleware.Subjects }}
      - name: {{ $v.Name }}
        permission:
          pub: {{ $v.Permission.Pub }}
          sub: {{ $v.Permission.Sub }}
        {{- if gt (len $v.Export) 0 }}
        export:
          {{- range $ek, $ev := $v.Export }}
        - appName: {{ $ev.AppName }}
          pub: {{ $ev.Pub }}
          sub: {{ $ev.Sub }}
          {{- end }}
        {{- end }}
      {{- end }}
    {{- end }}
    {{- if gt (len .Middleware.Refs) 0 }}
    refs: 
      {{- range $k, $v := .Middleware.Refs }}
      - appName: {{ $v.AppName }}
        subjects:
          {{- range $sk, $sv := $v.Subjects }}
          - name: {{ $sv.Name }}
            perm:
            {{- range $pk, $pv := $sv.Perm }}
            - {{ $pv }}
            {{- end }}
          {{- end }}
      {{- end }}
    {{- else }}
    refs: []
    {{- end }}
`

const minioRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-minio
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: minio
  minio:
    allowNamespaceBuckets: {{ .Middleware.AllowNamespaceBuckets }}
    buckets:
    {{- range $k, $v := .Middleware.Buckets }}
    - name: {{ $v.Name }}
    {{- end }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-minio-password
          key: "password"
	 {{- end }}
    user: {{ .Middleware.Username }}
`

const rabbitmqRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-rabbitmq
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: rabbitmq
  rabbitmq:
    vhosts:
    {{- range $k, $v := .Middleware.VHosts }}
    - name: {{ $v.Name }}
    {{- end }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-rabbitmq-password
          key: "password"
     {{- end }}
    user: {{ .Middleware.Username }}
`

const elasticsearchRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-elasticsearch
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: elasticsearch
  elasticsearch:
    allowNamespaceIndexes: {{ .Middleware.AllowNamespaceIndexes }}
    indexes:
    {{- range $k, $v := .Middleware.Indexes }}
    - name: {{ $v.Name }}
    {{- end }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-elasticsearch-password
          key: "password"
     {{- end }}
    user: {{ .Middleware.Username }}
`

const clickHouseRequest = `apiVersion: apr.bytetrade.io/v1alpha1
kind: MiddlewareRequest
metadata:
  name: {{ .AppName }}-clickhouse
  namespace: {{ .Namespace }}
spec:
  app: {{ .AppName }}
  appNamespace: {{ .AppNamespace }}
  middleware: clickhouse
  clickhouse:
    databases:
    {{- range $k, $v := .Middleware.Databases }}
    - name: {{ $v.Name }}
    {{- end }}
    password:
     {{- if not (eq .Middleware.Password "") }}
      value: {{ .Middleware.Password }}
     {{- else }}
      valueFrom:
        secretKeyRef:
          name: {{ .AppName }}-{{ .Namespace }}-clickhouse-password
          key: "password"
     {{- end }}
    user: {{ .Middleware.Username }}
`

type RequestParams struct {
	MiddlewareType MiddlewareType
	AppName        string
	AppNamespace   string
	Namespace      string
	Username       string
	Password       string
	OwnerName      string
	Middleware     *Middleware
}

func GenMiddleRequest(p RequestParams) ([]byte, error) {
	switch p.MiddlewareType {
	case TypePostgreSQL:
		return genPostgresRequest(p)
	case TypeRedis:
		return genRedisRequest(p)
	case TypeMongoDB:
		return genMongodbRequest(p)
	case TypeNats:
		return genNatsRequest(p)
	case TypeMinio:
		return genMinioRequest(p)
	case TypeRabbitMQ:
		return genRabbitMQRequest(p)
	case TypeElasticsearch:
		return genElasticsearchRequest(p)
	case TypeMariaDB:
		return genMariadbRequest(p)
	case TypeMySQL:
		return genMysqlRequest(p)
	case TypeClickHouse:
		return genClickHouseRequest(p)
	default:
		return []byte{}, fmt.Errorf("unsupported middleware type: %s", p.MiddlewareType)
	}
}

func renderTemplate(tplStr string, data interface{}) ([]byte, error) {
	tpl, err := template.New("tpl").Parse(tplStr)
	if err != nil {
		return []byte{}, err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func genPostgresRequest(p RequestParams) ([]byte, error) {
	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *PostgresConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware: &PostgresConfig{
			Username:  p.Username,
			Password:  p.Password,
			Databases: p.Middleware.Postgres.Databases,
		},
	}
	return renderTemplate(postgresRequest, data)
}

func genRedisRequest(p RequestParams) ([]byte, error) {
	if len(p.Middleware.Redis.Namespace) == 0 {
		return []byte{}, fmt.Errorf("redis requires at least one namespace/database")
	}
	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *RedisConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware: &RedisConfig{
			Password:  p.Password,
			Namespace: p.Middleware.Redis.Namespace,
		},
	}
	return renderTemplate(redisRequest, data)
}

func genMongodbRequest(p RequestParams) ([]byte, error) {
	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *MongodbConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware: &MongodbConfig{
			Username:  p.Username,
			Password:  p.Password,
			Databases: p.Middleware.MongoDB.Databases,
		},
	}
	return renderTemplate(mongodbRequest, data)
}

func genMariadbRequest(p RequestParams) ([]byte, error) {
	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *MariaDBConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware: &MariaDBConfig{
			Username:  p.Username,
			Password:  p.Password,
			Databases: p.Middleware.MariaDB.Databases,
		},
	}
	return renderTemplate(mariadbRequest, data)
}

func genMysqlRequest(p RequestParams) ([]byte, error) {
	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *MySQLConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware: &MySQLConfig{
			Username:  p.Username,
			Password:  p.Password,
			Databases: p.Middleware.MySQL.Databases,
		},
	}
	return renderTemplate(mysqlRequest, data)
}

func genNatsRequest(p RequestParams) ([]byte, error) {
	if p.Middleware.Nats == nil {
		return []byte{}, fmt.Errorf("natsConfig cannot be nil for NATS middleware request")
	}

	p.Middleware.Nats.Username = p.Username
	p.Middleware.Nats.Password = p.Password

	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *NatsConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware:   p.Middleware.Nats,
	}
	return renderTemplate(natsRequest, data)
}

func genMinioRequest(p RequestParams) ([]byte, error) {
	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *MinioConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware: &MinioConfig{
			Username:              p.Username,
			Password:              p.Password,
			Buckets:               p.Middleware.Minio.Buckets,
			AllowNamespaceBuckets: p.Middleware.Minio.AllowNamespaceBuckets,
		},
	}
	return renderTemplate(minioRequest, data)
}

func genRabbitMQRequest(p RequestParams) ([]byte, error) {
	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *RabbitMQConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware: &RabbitMQConfig{
			Username: p.Username,
			Password: p.Password,
			VHosts:   p.Middleware.RabbitMQ.VHosts,
		},
	}
	return renderTemplate(rabbitmqRequest, data)
}

func genElasticsearchRequest(p RequestParams) ([]byte, error) {
	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *ElasticsearchConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware: &ElasticsearchConfig{
			Username:              p.Username,
			Password:              p.Password,
			Indexes:               p.Middleware.Elasticsearch.Indexes,
			AllowNamespaceIndexes: p.Middleware.Elasticsearch.AllowNamespaceIndexes,
		},
	}
	return renderTemplate(elasticsearchRequest, data)
}

func genClickHouseRequest(p RequestParams) ([]byte, error) {
	data := struct {
		AppName      string
		AppNamespace string
		Namespace    string
		Middleware   *ClickHouseConfig
	}{
		AppName:      p.AppName,
		AppNamespace: p.AppNamespace,
		Namespace:    p.Namespace,
		Middleware: &ClickHouseConfig{
			Username:  p.Username,
			Password:  p.Password,
			Databases: p.Middleware.ClickHouse.Databases,
		},
	}
	return renderTemplate(clickHouseRequest, data)
}
