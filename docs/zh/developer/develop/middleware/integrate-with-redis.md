---
outline: [2, 3]
description: 了解如何在 Olares 中将应用接入内置的 Redis 服务。
---
# 集成 Redis

通过在 `OlaresManifest.yaml` 中声明依赖，并将系统注入的配置值映射到容器的环境变量中，即可在应用中使用 Olares 提供的 Redis 服务。

:::info Redis 已安装
Redis 服务在 Olares 中默认已安装。
:::

## 配置 `OlaresManifest.yaml`

在 `OlaresManifest.yaml` 中添加所需的中间件配置。

- 使用 `password` 字段指定 Redis 访问密码请求。
- 使用 `namespace` 字段申请 Redis 命名空间。

:::info 示例
```yaml
middleware:
  redis:
    password: password
    namespace: db0
```
:::

## 注入环境变量

在应用的部署 YAML 中，将系统注入的 `.Values.redis.*` 字段映射为应用所使用的环境变量。

:::info 示例
```yaml
containers:
  - name: my-app
    env:
      # 主机地址
      - name: REDIS_HOST
        value: {{ .Values.redis.host }}

      # 端口
      # 建议添加引号，确保作为字符串处理
      - name: REDIS_PORT
        value: "{{ .Values.redis.port }}"

      # 密码
      # 必须添加引号，以防止特殊字符导致解析错误
      - name: REDIS_PASSWORD
        value: "{{ .Values.redis.password }}"

      # 命名空间
      # 注意：请将 <namespace> 替换为 OlaresManifest 中定义的实际命名空间（例如 db0）
      - name: REDIS_NAMESPACE
        value: {{ .Values.redis.namespaces.<namespace> }}
```
:::

## Redis Values 参考

Redis Values 是在部署过程中由系统自动注入到 `values.yaml` 中的预定义变量。这些值由系统统一管理，用户无法自行修改。

| 键  | 类型  | 说明  |
|--|--|--|
| `.Values.redis.host` | String | Redis 数据库地址 |
| `.Values.redis.port` | Number  | Redis 数据库端口 |
| `.Values.redis.password`| String | Redis 数据库密码 |
| `.Values.redis.namespaces` | Map<String, String> | Redis 命名空间名称，以申请命名空间作为键。例如，若申请的命名空间名为 `app_ns`，可通过 `.Values.redis.namespaces.app_ns`获取对应值。 |
