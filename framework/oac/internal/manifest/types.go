package manifest

import (
	v1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	apimanifest "github.com/beclab/api/manifest"
)

const (
	APIVersionV1 = "v1"
	APIVersionV2 = "v2"
	APIVersionV3 = "v3"
)

// All schema types are direct aliases onto github.com/beclab/api/manifest
// (and the related v1alpha1 packages). Aliasing — rather than redeclaring —
// keeps the public API (oac.AppConfiguration etc.) and the internal
// validation/parsing pipeline operating on a single concrete type, so values
// flow freely without conversion.
type (
	AppMetaData         = apimanifest.AppMetaData
	AppConfiguration    = apimanifest.AppConfiguration
	AppSpec             = apimanifest.AppSpec
	Hardware            = apimanifest.Hardware
	CpuConfig           = apimanifest.CpuConfig
	GpuConfig           = apimanifest.GpuConfig
	SupportClient       = apimanifest.SupportClient
	Permission          = apimanifest.Permission
	ProviderPermission  = apimanifest.ProviderPermission
	Policy              = apimanifest.Policy
	Dependency          = apimanifest.Dependency
	Conflict            = apimanifest.Conflict
	Options             = apimanifest.Options
	ResetCookie         = apimanifest.ResetCookie
	AppScope            = apimanifest.AppScope
	WsConfig            = apimanifest.WsConfig
	Upload              = apimanifest.Upload
	OIDC                = apimanifest.OIDC
	Chart               = apimanifest.Chart
	Provider            = apimanifest.Provider
	SpecialResource     = apimanifest.SpecialResource
	ResourceRequirement = apimanifest.ResourceRequirement
	ResourceMode        = apimanifest.ResourceMode
	Middleware          = apimanifest.Middleware
	Database            = apimanifest.Database
	PostgresConfig      = apimanifest.PostgresConfig
	ArgoConfig          = apimanifest.ArgoConfig
	MinioConfig         = apimanifest.MinioConfig
	Bucket              = apimanifest.Bucket
	RabbitMQConfig      = apimanifest.RabbitMQConfig
	VHost               = apimanifest.VHost
	ElasticsearchConfig = apimanifest.ElasticsearchConfig
	Index               = apimanifest.Index
	RedisConfig         = apimanifest.RedisConfig
	MongodbConfig       = apimanifest.MongodbConfig
	MariaDBConfig       = apimanifest.MariaDBConfig
	MySQLConfig         = apimanifest.MySQLConfig
	ClickHouseConfig    = apimanifest.ClickHouseConfig
	NatsConfig          = apimanifest.NatsConfig
	Subject             = apimanifest.Subject
	Export              = apimanifest.Export
	Ref                 = apimanifest.Ref
	RefSubject          = apimanifest.RefSubject
	PermissionNats      = apimanifest.PermissionNats
)

type (
	Entrance    = v1alpha1.Entrance
	ServicePort = v1alpha1.ServicePort
	TailScale   = v1alpha1.TailScale
	ACL         = v1alpha1.ACL
)

type (
	AppEnvVar          = sysv1alpha1.AppEnvVar
	EnvVarSpec         = sysv1alpha1.EnvVarSpec
	ValueFrom          = sysv1alpha1.ValueFrom
	EnvValueOptionItem = sysv1alpha1.EnvValueOptionItem
)
