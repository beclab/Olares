# device-metrics-exporter-charts

![Version: v1.5.1](https://img.shields.io/badge/Version-v1.5.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.5.1](https://img.shields.io/badge/AppVersion-v1.5.1-informational?style=flat-square)

A Helm chart for AMD Device Metric Exporter

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Praveen Kumar Shanmugam | <prshanmug@amd.com> |  |
| Yan Sun | <yan.sun3@amd.com> |  |
| Shrey Ajmera | <shrey.ajmera@amd.com> |  |

## Requirements

Kubernetes: `>= 1.29.0-0`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Add node affinity for the daemonset of metrics exporter |
| args | list | `[]` | Additional arguments to pass to the exporter container (for testing only) |
| configMap | string | `""` | configMap name for the customizing configs and mount into metrics exporter container |
| hostNetwork | bool | `false` | Use host network for the metrics exporter pods |
| image.initContainerImage | string | `"busybox:1.36"` | metrics exporter initContainer image |
| image.pullPolicy | string | `"Always"` | metrics exporter image pullPolicy |
| image.pullSecrets | string | `""` | metrics exporter image pullSecret name |
| image.repository | string | `"docker.io/rocm/device-metrics-exporter"` | repository URL for the metrics exporter image |
| image.tag | string | `"v1.5.1"` | metrics exporter image tag |
| kubelet | object | `{"podResourceAPISocketPath":"/var/lib/kubelet/pod-resources"}` | kubelet configuration |
| kubelet.podResourceAPISocketPath | string | `"/var/lib/kubelet/pod-resources"` | host path for kubelet pod-resources directory (optional)    - vanilla k8s kubelet path: /var/lib/kubelet/pod-resources    - micro k8s kubelet path: /var/snap/microk8s/common/var/lib/kubelet/pod-resources/    - default to /var/lib/kubelet/pod-resources |
| monitor | object | `{"resources":{"gpu":true,"nic":false}}` | monitoring configuration |
| monitor.resources | object | `{"gpu":true,"nic":false}` | Enable or disable specific resource monitoring components. |
| nodeSelector | object | `{}` | Add node selector for the daemonset of metrics exporter |
| platform | string | `"k8s"` | Specify the platform to deploy the metrics exporter, k8s or openshift |
| podAnnotations | object | `{}` | Add annotations to the pods |
| priorityClassName | string | `""` | Set priorityClassName for the metrics exporter pods (optional) |
| resources | object | `{"limits":{"cpu":"2","memory":"4Gi"},"requests":{"cpu":"500m","memory":"512M"}}` | options for the metrics exporter container - default values are set if not specified |
| resources.limits | object | `{"cpu":"2","memory":"4Gi"}` | Resource limits and requests for the metrics exporter container |
| resources.limits.cpu | string | `"2"` | CPU limit for the metrics exporter container |
| resources.limits.memory | string | `"4Gi"` | Memory limit for the metrics exporter container |
| resources.requests.cpu | string | `"500m"` | CPU request for the metrics exporter container |
| resources.requests.memory | string | `"512M"` | Memory request for the metrics exporter container |
| service.ClusterIP.port | int | `5000` | set port for ClusterIP type service |
| service.NodePort.nodePort | int | `32500` | set nodePort for NodePort type service   |
| service.NodePort.port | int | `5000` | set port for NodePort type service    |
| service.annotations | object | `{}` | Add annotations to the service |
| service.type | string | `"ClusterIP"` | metrics exporter service type, could be ClusterIP or NodePort |
| serviceMonitor | object | `{"attachMetadata":{"node":false},"enabled":false,"honorLabels":true,"honorTimestamps":true,"interval":"30s","labels":{},"metricRelabelings":[],"relabelings":[]}` | ServiceMonitor configuration |
| serviceMonitor.attachMetadata | object | `{"node":false}` | Adds node metadata to discovered targets for node-based filtering |
| serviceMonitor.enabled | bool | `false` | Whether to create a ServiceMonitor resource for Prometheus Operator |
| serviceMonitor.honorLabels | bool | `true` | Honor labels configuration for ServiceMonitor |
| serviceMonitor.honorTimestamps | bool | `true` | Honor timestamps configuration for ServiceMonitor |
| serviceMonitor.interval | string | `"30s"` | Scrape interval for the ServiceMonitor |
| serviceMonitor.labels | object | `{}` | Additional labels for the ServiceMonitor |
| serviceMonitor.metricRelabelings | list | `[]` | Relabeling rules applied to individual scraped metrics |
| serviceMonitor.relabelings | list | `[]` | RelabelConfigs to apply to samples before scraping |
| tolerations | list | `[]` | Add tolerations for deploying metrics exporter on tainted nodes |
| upgradePolicy | object | `{"maxUnavailable":"1","upgradeStrategy":"RollingUpdate"}` | DaemonSet upgrade policy configuration |
| upgradePolicy.maxUnavailable | string | `"1"` | The maximum number of Pods that can be unavailable during the update process Can be an absolute number (e.g., 1, 5) or a percentage (e.g., "25%") Only applicable when upgradeStrategy is RollingUpdate |
| upgradePolicy.upgradeStrategy | string | `"RollingUpdate"` | The type of daemonset upgrade: RollingUpdate or OnDelete |

