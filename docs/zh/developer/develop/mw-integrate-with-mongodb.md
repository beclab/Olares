---
outline: [2, 3]
description: 了解如何在 Olares 中将你的应用接入 MongoDB 服务。
---
# 集成 MongoDB

通过在 `OlaresManifest.yaml` 中声明 MongoDB 中间件，并将系统注入的配置值映射到容器的环境变量中，即可在应用中使用 Olares 提供的 MongoDB 服务。

## 安装 MongoDB 服务

通过应用市场安装 MongoDB 服务。

1. 从启动台打开应用市场，搜索“MongoDB”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

安装完成后，MongoDB 服务及其连接信息将显示在控制面板的中间件列表中。

## 配置 `OlaresManifest.yaml`

在 `OlaresManifest.yaml` 中添加所需的中间件配置。

- 使用 `username` 字段指定 MongoDB 数据库用户。
- 使用 `databases` 字段申请一个或多个数据库。
- （可选）在每个数据库下使用 `script` 字段指定初始化脚本，这些脚本将在数据库创建完成后执行。

**示例**
```yaml
middleware:
  mongodb:
    username: chromium
    databases:
    - name: chromium
      script:
      - 'db.getSiblingDB("$databasename").myCollection.insertOne({ x: 111 });'
      # 请确保每一行都是完整的查询语句。
```

## 注入环境变量

在应用的部署 YAML 中，将系统注入的 `.Values.mongodb.*` 字段映射为应用所使用的环境变量。

**示例**
```yaml
containers:
  - name: my-app
    # MongoDB 对应的注入值如下
    env:
      - name: MONGODB_HOST
        value: "{{ .Values.mongodb.host }}"

      - name: MONGODB_PORT
        value: "{{ .Values.mongodb.port }}"

      - name: MONGODB_USER
        value: "{{ .Values.mongodb.username }}"

      - name: MONGODB_PASSWORD
        value: "{{ .Values.mongodb.password }}"

      # 数据库名称
      # 在 OlaresManifest 中配置的数据库名（例如：app_db）
      - name: MONGODB_DATABASE
        value: "{{ .Values.mongodb.databases.app_db }}"
```

## MongoDB Values 参考

MongoDB Values 是在部署过程中自动注入到 `values.yaml` 中的预定义变量，由系统统一管理，用户不可手动修改。

| 键  | 类型  | 说明  |
|--|--|--|
| `.Values.mongodb.host` | String | MongoDB 数据库地址 |
| `.Values.mongodb.port` | Number | MongoDB 数据库端口 |
| `.Values.mongodb.username` | String | MongoDB 数据库用户名 |
| `.Values.mongodb.password` | String  | MongoDB 数据库密码 |
| `.Values.mongodb.databases` | Map<String,String> | 以申请的数据库名作为键。<br/>例如申请 `app_db`，可通过 `.Values.mongodb.databases.app_db` 获取对应的值。 |