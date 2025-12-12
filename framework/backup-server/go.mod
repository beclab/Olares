module olares.com/backup-server

go 1.23.0

toolchain go1.23.7

replace (
	k8s.io/api => k8s.io/api v0.25.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.25.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.25.6
	k8s.io/apiserver => k8s.io/apiserver v0.25.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.25.6
	k8s.io/client-go => k8s.io/client-go v0.25.6
	k8s.io/component-base => k8s.io/component-base v0.25.6
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.60.1
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20220627174259-011e075b9cb8
	k8s.io/kubectl => k8s.io/kubectl v0.24.2
	k8s.io/utils => k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	olares.com/backups-sdk => github.com/Above-Os/backups-sdk v0.1.37
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.12.2
	sigs.k8s.io/json => sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2
	sigs.k8s.io/kustomize/api => sigs.k8s.io/kustomize/api v0.11.4
	sigs.k8s.io/kustomize/kyaml => sigs.k8s.io/kustomize/kyaml v0.13.6
	sigs.k8s.io/structured-merge-diff/v4 => sigs.k8s.io/structured-merge-diff/v4 v4.2.1
	sigs.k8s.io/yaml => sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/alitto/pond/v2 v2.3.2
	github.com/aws/aws-sdk-go-v2 v1.32.3
	github.com/aws/aws-sdk-go-v2/config v1.18.28
	github.com/aws/aws-sdk-go-v2/credentials v1.13.27
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.72
	github.com/aws/aws-sdk-go-v2/service/s3 v1.37.0
	github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/emicklei/go-restful-openapi/v2 v2.9.1
	github.com/emicklei/go-restful/v3 v3.12.1
	github.com/go-openapi/spec v0.20.7
	github.com/go-playground/assert/v2 v2.2.0
	github.com/go-resty/resty/v2 v2.16.5
	github.com/google/uuid v1.6.0
	github.com/lithammer/dedent v1.1.0
	github.com/nats-io/nats.go v1.43.0
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil/v4 v4.25.2
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.37.0
	golang.org/x/text v0.24.0
	k8s.io/api v0.26.2
	k8s.io/apimachinery v0.32.3
	k8s.io/client-go v0.32.1
	k8s.io/code-generator v0.26.1
	k8s.io/klog/v2 v2.130.1
	olares.com/backups-sdk v0.0.0-20250416000000-8290325e5376
	sigs.k8s.io/controller-runtime v0.12.2
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.35 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.14.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.19.3 // indirect
	github.com/aws/smithy-go v1.22.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/ebitengine/purego v0.8.2 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.31.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	golang.org/x/tools v0.26.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.26.1 // indirect
	k8s.io/component-base v0.26.1 // indirect
	k8s.io/gengo v0.0.0-20220902162205-c0856e24416d // indirect
	k8s.io/kube-openapi v0.0.0-20241105132330-32ad38e42d3f // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.2 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
