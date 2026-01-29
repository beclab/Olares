---
outline: [2, 3]
description: Learn how to enable OpenTelemetry auto-instrumentation in an Olares cluster and view trace data in Jaeger.
---
# View OpenTelemetry data

This guide walks you through enabling OpenTelemetry auto-instrumentation for services running in an Olares cluster and viewing trace data in Jaeger.

## Prerequisites

- Your target service runs as a Kubernetes workload (Deployment, StatefulSet, or DaemonSet).
- You have access to run `kubectl` against the Olares cluster.
- You can generate some traffic to the target service. Trace data is generated only when traffic exists.

## Install Jaeger

Jaeger is used to visualize trace data. Install Jaeger from Market.

1. Open Market from Launchpad and search for "Jaeger".
2. Click **Get**, then **Install**, and wait for the installation to complete.

## Apply tracing configuration

Prepare the tracing backend configuration before enabling auto-instrumentation.

1. Click [`otc.yaml`](https://cdn.olares.com/common/otc.yaml) to download the configuration file.
2. Upload the file to your Olares host.
3. In the directory containing the file, apply it:
    ```bash
    kubectl apply -f otc.yaml
    ```

## Configure service access

To enable OpenTelemetry auto-instrumentation, add specific **annotations** to the **Pod template** of your workload.

Auto-instrumentation is triggered entirely by annotations. No code changes are required.

:::info Rules for service access configuration
- Add annotations under `.spec.template.metadata.annotations` (Pod template, not top-level metadata).
- Pods will be recreated (rollout) for injection to take effect.
:::

:::tip Saving changes
After you finish editing with `kubectl edit`, save and exit the editor. Kubernetes will roll out updated Pods automatically in most cases.
:::

### BFL service (StatefulSet)

1. Edit the StatefulSet:
    ```bash
    kubectl edit sts -n user-space-<olaresid> bfl
    ```
2. Under `.spec.template.metadata.annotations`, add:
```yaml
spec:
  template:
    metadata:
      labels:
        tier: bfl
      # Locate here and add annotations  
      annotations:
        instrumentation.opentelemetry.io/inject-go: "olares-instrumentation"
        instrumentation.opentelemetry.io/go-container-names: "api"    
        instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/bfl-api"
        instrumentation.opentelemetry.io/inject-nginx: "olares-instrumentation"
        instrumentation.opentelemetry.io/inject-nginx-container-names: "ingress"    
```

### ChartRepo (Deployment)

1. Edit the Deployment:
    ```bash
    kubectl edit deploy -n os-framework chartrepo-deployment
    ```
2. Under `.spec.template.metadata.annotations`, add:
```yaml
spec:
  template:
    metadata:
      labels:
        app: chartrepo
        io.bytetrade.app: "true"
      # Locate here and add annotations  
      annotations:
        instrumentation.opentelemetry.io/inject-go: "olares-instrumentation"
        instrumentation.opentelemetry.io/go-container-names: "chartrepo"    
        instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/root/app" 
```

### Olares app (Deployment)

1. Edit the Deployment:
    ```bash
    kubectl edit deploy -n user-space-<olaresid> olares-app-deployment
    ```
2. Under `.spec.template.metadata.annotations`, add:
```yaml
spec:
  template:
    metadata:
      labels:
        app: olares-app
        io.bytetrade.app: "true"
      # Locate here and add annotations  
      annotations:
        instrumentation.opentelemetry.io/inject-nodejs: "olares-instrumentation"
        instrumentation.opentelemetry.io/nodejs-container-names: "user-service"    
        instrumentation.opentelemetry.io/inject-nginx: "olares-instrumentation"
        instrumentation.opentelemetry.io/inject-nginx-container-names: "olares-app"
```

### Files (DaemonSet)

1. Edit the DaemonSet:
    ```bash
    kubectl edit ds -n os-framework files
    ```
2. Under `.spec.template.metadata.annotations`, add:
```yaml
spec:
  template:
    metadata:
      labels:
        app: files
      # Locate here and add annotations  
      annotations:
        instrumentation.opentelemetry.io/inject-nginx: "olares-instrumentation"
        instrumentation.opentelemetry.io/inject-nginx-container-names: "nginx"    
        instrumentation.opentelemetry.io/inject-go: "olares-instrumentation"
        instrumentation.opentelemetry.io/go-container-names: "gateway,files,uploader" 
        instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/filebrowser"
```

### Market (Deployment)

1. Edit the Deployment:
    ```bash
    kubectl edit deploy -n os-framework market-deployment
    ```
2. Under `.spec.template.metadata.annotations`, add:
```yaml
spec:
  template:
    metadata:
      labels:
        app: appstore
        io.bytetrade.app: "true"
      # Locate here and add annotations  
      annotations:
        instrumentation.opentelemetry.io/inject-go: "olares-instrumentation"
        instrumentation.opentelemetry.io/go-container-names: "appstore-backend"    
        instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/opt/app/market"
```

### System server (Deployment)

1. Edit the Deployment:
    ```bash
    kubectl edit deploy -n user-system-<olaresid> system-server
    ```
2. Under `.spec.template.metadata.annotations`, add:
```yaml
spec:
  template:
    metadata:
      labels:
        app: systemserver
      # Locate here and add annotations  
      annotations:
        instrumentation.opentelemetry.io/go-container-names: "system-server"
        instrumentation.opentelemetry.io/inject-go: "olares-instrumentation:"
        instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/system-server"
```

## View traces in Jaeger

:::info Traces may appear with a delay
After rollout, traces may take 1–5 minutes to appear. Make sure the service receives traffic.
:::

Generate traffic to the service.

1. Open Jaeger from Launchpad.
2. Select the service name from the **Service** dropdown.
3. Click **Find Traces** to view trace data.
    ![View traces](/images/developer/develop/middleware/mw-jaeger.png#bordered)
