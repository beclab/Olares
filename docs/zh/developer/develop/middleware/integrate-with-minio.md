---
outline: [2, 3]
description: 了解如何在 Olares 中将你的应用接入 MinIO 服务。
---
# 集成 MinIO

通过在 `OlaresManifest.yaml` 中声明 MinIO 中间件，并将系统注入的配置值映射到容器的环境变量中，即可在应用中使用 Olares 提供的 MinIO 服务。

## 安装 MinIO 服务

通过应用市场安装 MinIO 服务。

1. 从启动台打开应用市场，搜索“MinIO”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

安装完成后，MinIO 服务及其连接信息将显示在控制面板的中间件列表中。

## 配置 `OlaresManifest.yaml`

在 `OlaresManifest.yaml` 中添加所需的中间件配置。

- 使用 `username` 字段指定 MinIO 的访问密钥（Access Key）。
- 使用 `buckets` 字段申请一个或多个存储桶。每个存储桶名称将作为键注入到 `.Values.minio.buckets` 中。

:::info 示例
```yaml
middleware:
  minio:
    username: miniouser
    buckets:
      - name: mybucket
```
:::

## 注入环境变量

在应用的部署 YAML 中，将系统注入的 `.Values.minio.*` 字段映射为应用所使用的环境变量。

:::info 示例
```yaml
containers:
  - name: my-app
    # 对于 MinIO，对应的值如下所示
    env:
      # 使用 host 和 port 构建 endpoint
      - name: MINIO_ENDPOINT
        value: "{{ .Values.minio.host }}:{{ .Values.minio.port }}"
      
      - name: MINIO_PORT
        value: "{{ .Values.minio.port }}"
      
      - name: MINIO_ACCESS_KEY
        value: "{{ .Values.minio.username }}"
      
      - name: MINIO_SECRET_KEY
        value: "{{ .Values.minio.password }}"
      
      # 存储桶名称
      # 使用在 OlaresManifest 中配置的存储桶名称（例如 'mybucket'）
      - name: MINIO_BUCKET
        value: "{{ .Values.minio.buckets.mybucket }}"
```
:::

## MinIO Values 参考

MinIO Values 是在部署过程中由系统自动注入到 `values.yaml` 中的预定义变量。这些值由系统统一管理，用户无法自行修改。

| 键  | 类型  | 说明  |
|--|--|--|
| `.Values.minio.host` | String | MinIO 服务地址 |
| `.Values.minio.port` | Number | MinIO 服务端口 |
| `.Values.minio.username` | String | MinIO 访问密钥（Access Key） |
| `.Values.minio.password` | String | MinIO 密钥（Secret Key） |
| `.Values.minio.buckets` | Map<String,String> | 以申请的存储桶名称作为键。例如申请 `mybucket`，可通过 `.Values.minio.buckets.mybucket` 获取对应的值。 |