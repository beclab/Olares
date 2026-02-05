---
outline: [2, 3]
description: 了解如何在 Olares 中将你的应用接入 Elasticsearch 服务。
---
# 集成 Elasticsearch

通过在 `OlaresManifest.yaml` 中声明 Elasticsearch 中间件，并将系统注入的配置值映射到容器的环境变量中，即可在应用中使用 Olares 提供的 Elasticsearch 服务。

## 安装 Elasticsearch 服务

通过应用市场安装 Elasticsearch 服务。

1. 从启动台打开应用市场，搜索“Elasticsearch”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

安装完成后，Elasticsearch 服务及其连接信息将显示在控制面板的中间件列表中。

## 配置 `OlaresManifest.yaml`

在 `OlaresManifest.yaml` 中添加所需的中间件配置。

- 使用 `username` 字段指定 Elasticsearch 用户名。
- 使用 `indexes` 字段申请一个或多个索引。每个索引名称将作为键注入到 `.Values.elasticsearch.indexes` 中。

**示例**
```yaml
middleware:
  elasticsearch:
    username: elasticlient
    indexes:
      - name: aaa
```

## 注入环境变量

在应用的部署 YAML 中，将系统注入的 `.Values.elasticsearch.*` 字段映射为应用所使用的环境变量。

**示例**
```yaml
containers:
  - name: my-app
    env:
      - name: ES_HOST
        value: "{{ .Values.elasticsearch.host }}"

      - name: ES_PORT
        value: "{{ .Values.elasticsearch.port }}"

      - name: ES_USER
        value: "{{ .Values.elasticsearch.username }}"

      - name: ES_PASSWORD
        value: "{{ .Values.elasticsearch.password }}"

      # 索引名称
      # 使用在 OlaresManifest 中配置的索引名称（例如 aaa）
      - name: ES_INDEX
        value: "{{ .Values.elasticsearch.indexes.aaa }}"
```

## Elasticsearch Values 参考

Elasticsearch Values 是在部署过程中由系统自动注入到 `values.yaml` 中的预定义变量。这些值由系统统一管理，用户无法自行修改。

| 键  | 类型  | 说明  |
|--|--|--|
| `.Values.elasticsearch.host` | String | Elasticsearch 服务地址 |
| `.Values.elasticsearch.port` | Number  | Elasticsearch 服务端口 |
| `.Values.elasticsearch.username` | String | Elasticsearch 用户名 |
| `.Values.elasticsearch.password` | String | Elasticsearch 密码 |
| `.Values.elasticsearch.indexes`  | Map<String,String> | 以申请的索引名称作为键。例如申请 `aaa`，可通过 `.Values.elasticsearch.indexes.aaa` 获取对应的值。 |