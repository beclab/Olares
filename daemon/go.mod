module github.com/beclab/Olares/daemon

go 1.24.11

replace (
	bytetrade.io/web3os/backups-sdk => github.com/Above-Os/backups-sdk v0.1.17
	bytetrade.io/web3os/bfl => github.com/beclab/bfl v0.3.36
	github.com/labstack/echo/v4 => github.com/eball/echo/v4 v4.13.4-patch
	k8s.io/api => k8s.io/api v0.34.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.34.0
	k8s.io/client-go => k8s.io/client-go v0.34.0
	kubesphere.io/api => ../../kubesphere-ext/staging/src/kubesphere.io/api/
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.19.6

)

require (
	bytetrade.io/web3os/bfl v0.0.0-00010101000000-000000000000
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/beclab/Olares/cli v0.0.0-20251230161135-5264df60cc33
	github.com/beclab/Olares/framework/app-service v0.0.0-20251225061130-909b7656fd70
	github.com/containerd/containerd v1.7.29
	github.com/distribution/distribution/v3 v3.0.0
	github.com/dustin/go-humanize v1.0.1
	github.com/eball/zeroconf v0.2.2
	github.com/godbus/dbus/v5 v5.1.0
	github.com/gofiber/fiber/v2 v2.52.9
	github.com/google/gopacket v1.1.19
	github.com/hirochachacha/go-smb2 v1.1.0
	github.com/jaypipes/ghw v0.13.0
	github.com/jochenvg/go-udev v0.0.0-20171110120927-d6b62d56d37b
	github.com/joho/godotenv v1.5.1
	github.com/klauspost/cpuid/v2 v2.2.8
	github.com/labstack/echo/v4 v4.0.0-00010101000000-000000000000
	github.com/libp2p/go-netroute v0.2.2
	github.com/mackerelio/go-osstat v0.2.5
	github.com/mdlayher/raw v0.1.0
	github.com/miekg/dns v1.1.55
	github.com/muka/network_manager v0.0.0-20200903202308-ae5ede816e07
	github.com/nxadm/tail v1.4.11
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58
	github.com/pelletier/go-toml v1.9.5
	github.com/pkg/errors v0.9.1
	github.com/rubiojr/go-usbmon v0.0.0-20240513072523-d5cbf336b315
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/shirou/gopsutil/v4 v4.25.7
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/pflag v1.0.7
	github.com/txn2/txeh v1.5.5
	github.com/vishvananda/netlink v1.3.0
	go.opentelemetry.io/otel/trace v1.36.0
	golang.org/x/crypto v0.45.0
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b
	golang.org/x/sys v0.38.0
	k8s.io/api v0.34.1
	k8s.io/apiextensions-apiserver v0.34.0
	k8s.io/apimachinery v0.34.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cri-api v0.34.1
	k8s.io/cri-client v0.34.1
	k8s.io/klog/v2 v2.130.1
	k8s.io/mount-utils v0.31.0
	k8s.io/utils v0.0.0-20250820121507-0af2bda4dd1d
	sigs.k8s.io/controller-runtime v0.22.0
	tinygo.org/x/bluetooth v0.10.0
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230811130428-ced1acdcaa24 // indirect
	github.com/AdamKorcz/go-118-fuzz-build v0.0.0-20230306123547-8075edf89bb0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.13.0 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/cgroups/v3 v3.0.5 // indirect
	github.com/containerd/containerd/api v1.8.0 // indirect
	github.com/containerd/continuity v0.4.4 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/fifo v1.1.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/containerd/ttrpc v1.2.7 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/cyphar/filepath-securejoin v0.5.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/geoffgarside/ber v1.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.21.1 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/jaypipes/pcidb v1.0.1 // indirect
	github.com/jkeiser/iter v0.0.0-20200628201005-c8aa0ae784d1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mdlayher/packet v0.0.0-20220221164757-67998ac0ff93 // indirect
	github.com/mdlayher/socket v0.2.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/signal v0.7.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runc v1.3.0 // indirect
	github.com/opencontainers/runtime-spec v1.2.1 // indirect
	github.com/opencontainers/selinux v1.13.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.23.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/saltosystems/winrt-go v0.0.0-20240509164145-4f7860a3bd2b // indirect
	github.com/soypat/cyw43439 v0.0.0-20240609122733-da9153086796 // indirect
	github.com/soypat/seqs v0.0.0-20240527012110-1201bab640ef // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tinygo-org/cbgo v0.0.4 // indirect
	github.com/tinygo-org/pio v0.0.0-20231216154340-cd888eb58899 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.36.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.36.0 // indirect
	go.opentelemetry.io/otel/sdk v1.36.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/term v0.37.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/genproto v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.72.2 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	howett.net/plist v1.0.0 // indirect
	k8s.io/apiserver v0.34.0 // indirect
	k8s.io/component-base v0.34.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
