module bytetrade.io/web3os/bfl

go 1.23.0

require (
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/beclab/api v0.0.0-20250707125039-268cccf7e81c
	github.com/beclab/lldap-client v0.0.11
	github.com/emicklei/go-restful v2.16.0+incompatible
	github.com/emicklei/go-restful-openapi/v2 v2.9.1
	github.com/emicklei/go-restful/v3 v3.8.0
	github.com/enobufs/go-nats v0.0.1
	github.com/go-logr/logr v1.4.1
	github.com/go-openapi/spec v0.20.7
	github.com/go-resty/resty/v2 v2.16.2
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/json-iterator/go v1.1.12
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/common v0.37.0
	github.com/spf13/cobra v1.5.0
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.19.0
	k8s.io/api v0.26.2
	k8s.io/apimachinery v0.26.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.130.1
	k8s.io/utils v0.0.0-20221107191617-1a15be271d1d
	sigs.k8s.io/controller-runtime v0.10.0
)

require (
	github.com/Khan/genqlient v0.7.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-logr/zapr v0.4.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/stun v0.3.2 // indirect
	github.com/pion/transport v0.8.8 // indirect
	github.com/pion/turn v1.3.5 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vektah/gqlparser/v2 v2.5.20 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.27.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.22.1 // indirect
	k8s.io/component-base v0.22.1 // indirect
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0
	k8s.io/api => k8s.io/api v0.21.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.2
	k8s.io/client-go => k8s.io/client-go v0.21.2
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.9.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7
)
