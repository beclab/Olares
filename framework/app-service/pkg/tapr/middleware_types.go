package tapr

import "github.com/beclab/api/manifest"

// Middleware-related structs are aliased to the shared manifest package in
// github.com/beclab/api so that the configs describing app middleware needs
// live in a single place and are compatible across packages.
type (
	Middleware          = manifest.Middleware
	Database            = manifest.Database
	PostgresConfig      = manifest.PostgresConfig
	ArgoConfig          = manifest.ArgoConfig
	MinioConfig         = manifest.MinioConfig
	Bucket              = manifest.Bucket
	RabbitMQConfig      = manifest.RabbitMQConfig
	VHost               = manifest.VHost
	ElasticsearchConfig = manifest.ElasticsearchConfig
	Index               = manifest.Index
	RedisConfig         = manifest.RedisConfig
	MongodbConfig       = manifest.MongodbConfig
	MariaDBConfig       = manifest.MariaDBConfig
	MySQLConfig         = manifest.MySQLConfig
	ClickHouseConfig    = manifest.ClickHouseConfig
	NatsConfig          = manifest.NatsConfig
	Subject             = manifest.Subject
	Export              = manifest.Export
	Ref                 = manifest.Ref
	RefSubject          = manifest.RefSubject
	// Permission was renamed to PermissionNats in the manifest package; keep
	// the shorter local name here for backwards compatibility.
	Permission = manifest.PermissionNats
)
