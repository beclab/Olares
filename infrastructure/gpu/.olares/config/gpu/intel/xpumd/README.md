# xpumd

![Version: 2.0.0](https://img.shields.io/badge/Version-2.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)
A Helm chart for Intel(R) XPUM Daemon
**Homepage:** <https://github.com/intel/xpumanager/>

## Pre-conditions

Kubernetes cluster with either Intel GPU plugin[^1] or Intel GPU (DRA) resource driver[^2] installed
and node(s) with Intel GPUs supported by the XPUMD image driver stack
([Level-Zero driver release notes](https://github.com/intel/compute-runtime/releases)).

NOTE: Enable NFD ([Node Feature Discovery](https://kubernetes-sigs.github.io/node-feature-discovery/))
/ GPU node labeling when installing either GPU plugin or DRA driver, to avoid `Pending` daemon
pods on cluster nodes that do not have Intel GPUs.

For stub-driver or other non-GPU test deployments, the chart can also run with
`gpuAccess=none`, which skips GPU resource requests entirely.

[^1]: [Intel GPU plugin](https://intel.github.io/intel-device-plugins-for-kubernetes/cmd/gpu_plugin/README.html)
[^2]: [Intel GPU resource driver](https://github.com/intel/intel-resource-drivers-for-kubernetes/tree/main/doc/gpu#readme)

## Installation

Specify which type of Intel [GPU monitoring access](#other-values) is used in the cluster:

* `dra`: Intel GPU (DRA) resource driver[^2] claim for monitoring Intel GPUs
* `plugin`: Intel GPU plugin[^1] (v0.36.0 or newer) resource for monitoring Intel GPUs
* `i915`: Intel GPU plugin (legacy) resource for monitoring Intel GPUs supported by the `i915` kernel driver
* `xe`: Intel GPU plugin (legacy) resource for monitoring Intel GPUs supported by the `xe` kernel driver
* `none`: Do not request GPU resources. Useful for stub-driver testing on generic clusters.

For example:

```bash
MONITOR=dra
```

Create namespace for the daemon:

```bash
kubectl create ns intel-xpumd
```

If using DRA, label namespace for [admin access](https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/)
(needed for GPU monitoring):

```bash
kubectl label ns intel-xpumd resource.kubernetes.io/admin-access=true
```

Install the chart:

```bash
helm install xpumd oci://ghcr.io/intel/xpumanager/charts/xpumd \
  --version 0.0.0-latest \
  --set gpuAccess=$MONITOR \
  --namespace intel-xpumd
```

If cluster GPU nodes are labeled (with NFD), add option limiting daemon pods to Intel GPU nodes:

```bash
  --set-string 'nodeSelector.intel\.feature\.node\.kubernetes\.io/gpu=true'
```

### Optional functionality

By default, only `intel_xpu_info` exporter (with local unix / gRPC socket endpoint) for GPU status is enabled.
To enable other exporters, the corresponding values must be set.

To enable also Prometheus exporter endpoint, Prometheus monitoring of that
and Grafana dashboard for the collected metrics, add options:

```bash
  --set config.service.pipelines.metrics.exporters="{intel_xpu_info,prometheus}" \
  --set prometheus.release=prometheus-stack \
  --set prometheus.monitor=true \
  --set grafana.dashboards=true \
  --set grafana.namespace=monitoring
```

NOTE: above requires Prometheus to be installed to the cluster beforehand, otherwise install fails.
See [monitoring](../../docs/MONITORING.md) and [Prometheus exporter README](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/prometheusexporter#readme).

To enable (only) OTLP gRPC exporter i.e. data pushing to specified OTel collector service, add options:

```bash
  --set config.exporters.otlp.endpoint="otel-collector:4317" \
  --set config.service.pipelines.metrics.exporters="{otlp}"
```

See [OTEL_STACK](../../docs/OTEL_STACK.md) on how to setup OTel collector, and
[OTLP gRPC exporter README](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter#readme) or
[OTLP HTTP exporter README](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter#readme)
on how to configure OTel exporter for gRPC or HTTP transport.

### Using private GitHub registry

If image is in a private GitHub registry to which you have access, Helm login to ghcr.io:

```bash
helm registry login ghcr.io -u $GITHUB_USERNAME
```

Create an image pull secret in your Kubernetes cluster:

```bash
kubectl create secret docker-registry ghcr-secret \
  --namespace=intel-xpumd \
  --docker-server=ghcr.io/intel \
  --docker-username=$GITHUB_USERNAME \
  --docker-password=$GITHUB_PERSONAL_ACCESS_TOKEN
```

And add following option to chart install:

```bash
  --set imagePullSecrets[0].name=ghcr-secret
```

## Values

### XPUM Daemon Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| config.exporters | object |   | [Configuration for exporters](https://opentelemetry.io/docs/collector/configuration/#exporters) |
| config.exporters.intel_xpu_info | object | `{}` | Override configuration for the Intel XPU info exporter. Should only be used for advanced use cases. See `intel_xpu_info` README for details. |
| config.extensions | object | `{}` | [Configuration for extensions](https://opentelemetry.io/docs/collector/configuration/#extensions) |
| config.processors | object | `{"intel_xpu_status":{}}` | [Configuration for processors](https://opentelemetry.io/docs/collector/configuration/#processors) |
| config.processors.intel_xpu_status | object | `{}` | Override configuration for the Intel XPU processor. Should only be used for advanced use cases. See `intel_xpu_status` README for details. |
| config.receivers | object |   | [Configuration for receivers](https://opentelemetry.io/docs/collector/configuration/#receivers) |
| config.receivers.intel_crashlog | object | `{}` | Override configuration for the Intel Crashlog receiver. Should only be used for advanced use cases. See `intel_crashlog` README for details. |
| config.receivers.intel_xpu | object |   | Configuration for the Intel XPU receiver. |
| config.receivers.intel_xpu.collection_interval | string | `"5s"` | Metrics data collection interval. Must be at least twice the sampling_interval. |
| config.receivers.intel_xpu.initial_delay | string | `"1s"` | Initial start delay for metrics collection, any non positive value is assumed to be immediately. |
| config.receivers.intel_xpu.metrics | object | `{}` | Configuration for enabling/disabling individual metrics. |
| config.receivers.intel_xpu.sampling_interval | string | `"1s"` | Sampling interval for the high-frequency metrics. |
| config.receivers.intel_xpu.timeout | int | `0` | Metrics collection timeout. |
| config.service | object | `{"pipelines":{"logs":{"exporters":["intel_xpu_info"],"receivers":["intel_xpu","intel_crashlog"]},"metrics":{"exporters":["intel_xpu_info"],"processors":["intel_xpu_status"],"receivers":["intel_xpu"]}},"telemetry":{"logs":{"disable_stacktrace":true,"level":"info"}}}` | [Configuration for service](https://opentelemetry.io/docs/collector/configuration/#service) |

### Other Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | [Affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) for the pods |
| config.receivers.intel_xpu.fail_on_sysman_init_error | bool | `false` | Whether to fail collector startup if the L0 Sysman API cannot be initialized (e.g. no GPU driver or device access). |
| crashlog.directory | string | `"/var/log/crashlog"` | Host directory to watch for collecting GPU crash logs. Use `*` to match all files. |
| extraEnv | list | `[]` | Extra environment variables for the xpumd container |
| extraVolumeMounts | list | `[]` | Additional volume mounts for the xpumd container |
| extraVolumes | list | `[]` | Additional volumes for the xpumd pod |
| fullnameOverride | string | `""` | Override the fully qualified app name |
| gpuAccess | string | `"dra"` | method for requesting monitoring access to Intel GPUs: `dra` (K8s DRA GPU driver), `plugin` (K8s GPU plugin), `i915` / `xe` (old K8s GPU plugin KMD-based resource names), `none` (no GPU resource request, useful for stub-driver testing) |
| grafana.dashboards | bool | `false` | Install XPUMD dashboard(s) for Grafana |
| grafana.namespace | string | `"monitoring"` | Namespace for Grafana install. Needed when Grafana dashboard auto-loader sidecar `searchNamespace: ALL` option is not used |
| image.pullPolicy | string | `"Always"` | Image pull policy |
| image.repository | string | `"ghcr.io/intel/xpumanager/xpumd"` | Image repository |
| image.tag | string | `""` | Image tag, defaults to Chart.AppVersion |
| imagePullSecrets | list | `[]` | [Image pull secrets](https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod) |
| initContainers | list | `[]` | Init containers to run |
| nameOverride | string | `""` | Override the chart name |
| nodeSelector | object | `{}` | Node selector for pod placement |
| podAnnotations | object | `{}` | Annotations to add to the pod |
| podSecurityContextOverride | object | `{}` | [Pod security context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-pod). NOTE: security settings control to what GPU metrics container has access to |
| prometheus.monitor | bool | `false` | Add Prometheus service monitor for XPUMD, requires Prometheus to be installed. overrides `service.create` |
| prometheus.release | string | `"prometheus-stack"` | Helm release name for Helm ('kube-prometheus') chart that installed Prometheus. (May be) needed for Prometheus to detect daemon serviceMonitor |
| resourcesOverride | object | `{}` | [Resource requests and limits](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/) of the container. NOTE: overrides `gpuAccess` setting |
| securityContextOverride | object | `{}` | [Container security context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-container). NOTE: security settings control to what GPU metrics container has access to |
| service.create | bool | `false` | Create service |
| service.port | int | `8080` | Service port |
| service.type | string | `"ClusterIP"` | Service type |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Create service account |
| serviceAccount.name | string | `""` | The name of the service account to use. If not specified and create is true, a name is generated using the fully qualified app name. |
| tolerations | list | `[]` | [Tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) for the pods |
