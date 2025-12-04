package tapr

// Middleware describe middleware config.
type Middleware struct {
	Postgres      *PostgresConfig      `yaml:"postgres,omitempty"`
	Redis         *RedisConfig         `yaml:"redis,omitempty"`
	MongoDB       *MongodbConfig       `yaml:"mongodb,omitempty"`
	Nats          *NatsConfig          `yaml:"nats,omitempty"`
	Minio         *MinioConfig         `yaml:"minio,omitempty"`
	RabbitMQ      *RabbitMQConfig      `yaml:"rabbitmq,omitempty"`
	Elasticsearch *ElasticsearchConfig `yaml:"elasticsearch,omitempty"`
	MariaDB       *MariaDBConfig       `yaml:"mariadb,omitempty"`
	MySQL         *MySQLConfig         `yaml:"mysql,omitempty"`
	Argo          *ArgoConfig          `yaml:"argo,omitempty"`
}

// Database specify database name and if distributed.
type Database struct {
	Name        string   `yaml:"name" json:"name"`
	Extensions  []string `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	Scripts     []string `yaml:"scripts,omitempty" json:"scripts,omitempty"`
	Distributed bool     `yaml:"distributed,omitempty" json:"distributed,omitempty"`
}

// PostgresConfig contains fields for postgresql config.
type PostgresConfig struct {
	Username  string     `yaml:"username" json:"username"`
	Password  string     `yaml:"password,omitempty" json:"password,omitempty"`
	Databases []Database `yaml:"databases" json:"databases"`
}

type ArgoConfig struct {
	Required bool `yaml:"required" json:"required"`
}

type MinioConfig struct {
	Username              string   `yaml:"username" json:"username"`
	Password              string   `yaml:"password" json:"password"`
	Buckets               []Bucket `yaml:"buckets" json:"buckets"`
	AllowNamespaceBuckets bool     `yaml:"allowNamespaceBuckets" json:"allowNamespaceBuckets"`
}

type Bucket struct {
	Name string `json:"name"`
}

type RabbitMQConfig struct {
	Username string  `yaml:"username" json:"username"`
	Password string  `yaml:"password" json:"password"`
	VHosts   []VHost `yaml:"vhosts" json:"vhosts"`
}

type VHost struct {
	Name string `json:"name"`
}

type ElasticsearchConfig struct {
	Username              string  `yaml:"username" json:"username"`
	Password              string  `yaml:"password" json:"password"`
	Indexes               []Index `yaml:"indexes" json:"indexes"`
	AllowNamespaceIndexes bool    `yaml:"allowNamespaceIndexes" json:"allowNamespaceIndexes"`
}

type Index struct {
	Name string `json:"name"`
}

// RedisConfig contains fields for redis config.
type RedisConfig struct {
	Password  string `yaml:"password,omitempty" json:"password"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

// MongodbConfig contains fields for mongodb config.
type MongodbConfig struct {
	Username  string     `yaml:"username" json:"username"`
	Password  string     `yaml:"password,omitempty" json:"password"`
	Databases []Database `yaml:"databases" json:"databases"`
}

// MariaDBConfig contains fields for mariadb config.
type MariaDBConfig struct {
	Username  string     `yaml:"username" json:"username"`
	Password  string     `yaml:"password,omitempty" json:"password"`
	Databases []Database `yaml:"databases" json:"databases"`
}

// MySQLConfig contains fields for mysql config.
type MySQLConfig struct {
	Username  string     `yaml:"username" json:"username"`
	Password  string     `yaml:"password,omitempty" json:"password"`
	Databases []Database `yaml:"databases" json:"databases"`
}

type NatsConfig struct {
	Username string    `yaml:"username" json:"username"`
	Password string    `yaml:"password,omitempty" json:"password,omitempty"`
	Subjects []Subject `yaml:"subjects" json:"subjects"`
	Refs     []Ref     `yaml:"refs" json:"refs"`
}

type Subject struct {
	Name string `yaml:"name" json:"name"`
	// Permissions indicates the permission that app can perform on this subject
	Permission Permission   `yaml:"permission" json:"permission"`
	Export     []Permission `yaml:"export" json:"export"`
}

type Export struct {
	AppName string `yaml:"appName" json:"appName"`
	Pub     string `yaml:"pub" json:"pub"`
	Sub     string `yaml:"sub" json:"sub"`
}

type Ref struct {
	AppName string `yaml:"appName" json:"appName"`
	// option for ref app in user-space-<>, user-system-<>, os-system
	AppNamespace string       `yaml:"appNamespace" json:"appNamespace"`
	Subjects     []RefSubject `yaml:"subjects" json:"subjects"`
}

type RefSubject struct {
	Name string   `yaml:"name" json:"name"`
	Perm []string `yaml:"perm" json:"perm"`
}

type Permission struct {
	AppName string `yaml:"appName,omitempty" json:"appName,omitempty"`
	// default is deny
	Pub string `yaml:"pub" json:"pub"`
	Sub string `yaml:"sub" json:"sub"`
}
