---
outline: [2, 3]
description: 了解如何在 Olares 集群中启用 OpenTelemetry 自动注入，并在 Jaeger 中查看 Trace 数据。
---
# 查看 OpenTelemetry 数据

本文介绍如何在 Olares 集群中为服务启用 OpenTelemetry 自动注入，并通过 Jaeger 查看 Trace 数据。


## 前提条件

- 目标服务以 Kubernetes 工作负载形式运行，例如 Deployment、StatefulSet 或 DaemonSet。
- 你可以使用 `kubectl` 访问 Olares 集群。
- 你可以向目标服务发送请求流量。只有存在流量时才会生成 Trace 数据。

## 安装 Jaeger

Jaeger 用于可视化 Trace 数据，需要通过应用市场安装。

1. 从启动台打开应用市场，搜索“Jaeger”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

## 应用 Trace 配置

在启用自动注入之前，需要先准备 Trace 后端配置。

1. 点击 [`otc.yaml`](https://cdn.olares.com/common/otc.yaml) 下载配置文件。
2. 将该文件上传到你的 Olares 主机。
3. 进入文件所在目录并执行：
    ```bash
    kubectl apply -f otc.yaml
    ```

## 配置服务接入

要启用 OpenTelemetry 自动注入，需要在工作负载的 **Pod 模板**中添加指定的 **annotations**。

自动注入完全由 annotations 触发，不需要修改任何业务代码。

:::info 服务接入配置规则
- 必须在 `.spec.template.metadata.annotations` 下添加 annotations，而不是在顶层 metadata 中。
- 修改 Pod 模板后，Pod 会通过滚动更新的方式重新创建，注入才会生效。
:::

:::tip 保存修改
使用 `kubectl edit` 编辑完成后，保存并退出编辑器。大多数情况下，Kubernetes 会自动触发 Pod 的滚动更新。
:::

### BFL 服务（StatefulSet）

1. 编辑 StatefulSet：
    ```bash
    kubectl edit sts -n user-space-<olaresid> bfl
    ```
2. 在 `.spec.template.metadata.annotations` 下添加：
    ```yaml
    spec:
      template:
        metadata:
          labels:
            tier: bfl
          # 在这里添加 annotations 
          annotations:
            instrumentation.opentelemetry.io/inject-go: "olares-instrumentation"
            instrumentation.opentelemetry.io/go-container-names: "api"    
            instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/bfl-api"
            instrumentation.opentelemetry.io/inject-nginx: "olares-instrumentation"
            instrumentation.opentelemetry.io/inject-nginx-container-names: "ingress"    
    ```

### ChartRepo（Deployment）

1. 编辑 Deployment：
    ```bash
    kubectl edit deploy -n os-framework chartrepo-deployment
    ```
2. 在 `.spec.template.metadata.annotations` 下添加：
    ```yaml
    spec:
      template:
        metadata:
          labels:
            app: chartrepo
            io.bytetrade.app: "true"
          # 在这里添加 annotations 
          annotations:
            instrumentation.opentelemetry.io/inject-go: "olares-instrumentation"
            instrumentation.opentelemetry.io/go-container-names: "chartrepo"    
            instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/root/app" 
    ```

### Olares 应用（Deployment）

1. 编辑 Deployment：
    ```bash
    kubectl edit deploy -n user-space-<olaresid> olares-app-deployment
    ```
2. 在 `.spec.template.metadata.annotations` 下添加：
    ```yaml
    spec:
      template:
        metadata:
          labels:
            app: olares-app
            io.bytetrade.app: "true"
          # 在这里添加 annotations 
          annotations:
            instrumentation.opentelemetry.io/inject-nodejs: "olares-instrumentation"
            instrumentation.opentelemetry.io/nodejs-container-names: "user-service"    
            instrumentation.opentelemetry.io/inject-nginx: "olares-instrumentation"
            instrumentation.opentelemetry.io/inject-nginx-container-names: "olares-app"
    ```

### Files 服务（DaemonSet）

1. 编辑 DaemonSet：
    ```bash
    kubectl edit ds -n os-framework files
    ```
2. 在 `.spec.template.metadata.annotations` 下添加：
    ```yaml
    spec:
      template:
        metadata:
          labels:
            app: files
          # 在这里添加 annotations 
          annotations:
            instrumentation.opentelemetry.io/inject-nginx: "olares-instrumentation"
            instrumentation.opentelemetry.io/inject-nginx-container-names: "nginx"    
            instrumentation.opentelemetry.io/inject-go: "olares-instrumentation"
            instrumentation.opentelemetry.io/go-container-names: "gateway,files,uploader" 
            instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/filebrowser"
    ```

### Market（Deployment）

1. 编辑 Deployment：
    ```bash
    kubectl edit deploy -n os-framework market-deployment
    ```
2. 在 `.spec.template.metadata.annotations` 下添加：
    ```yaml
    spec:
      template:
        metadata:
          labels:
            app: appstore
            io.bytetrade.app: "true"
          # 在这里添加 annotations 
          annotations:
            instrumentation.opentelemetry.io/inject-go: "olares-instrumentation"
            instrumentation.opentelemetry.io/go-container-names: "appstore-backend"    
            instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/opt/app/market"
    ```

### System server（Deployment）

1. 编辑 Deployment：
    ```bash
    kubectl edit deploy -n user-system-<olaresid> system-server
    ```
2. 在 `.spec.template.metadata.annotations` 下添加：
    ```yaml
    spec:
      template:
        metadata:
          labels:
            app: systemserver
          # 在这里添加 annotations 
          annotations:
            instrumentation.opentelemetry.io/go-container-names: "system-server"
            instrumentation.opentelemetry.io/inject-go: "olares-instrumentation:"
            instrumentation.opentelemetry.io/otel-go-auto-target-exe: "/system-server"
    ```

## 在 Jaeger 中查看 Trace 数据

:::info Trace 数据显示延迟说明
在 Pod 滚动更新完成后，Trace 数据可能需要 1–5 分钟才会显示。请确保服务已接收到请求流量。
:::

1. 从启动台打开 Jaeger。
2. 在 **Service** 下拉列表中选择对应的服务。
3. 点击 **Find Traces** 查看 Trace 数据。
    ![查看 Trace 数据](/images/developer/develop/middleware/mw-jaeger.png#bordered)