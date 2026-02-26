---
outline: [2, 3]
description: 了解如何在 Olares 中将你的应用接入 MySQL 服务。
---
# 集成 MySQL

通过在 `OlaresManifest.yaml` 中声明 MySQL 中间件，并将系统注入的配置值映射到容器的环境变量中，即可在应用中使用 Olares 提供的 MySQL 服务。

## 安装 MySQL 服务

通过应用市场安装 MySQL 服务。

1. 从启动台打开应用市场，搜索“MySQL”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

安装完成后，MySQL 服务及其连接信息将显示在控制面板的中间件列表中。

## 配置 `OlaresManifest.yaml`

在 `OlaresManifest.yaml` 中添加所需的中间件配置。

- 使用 `username` 字段指定 MySQL 数据库用户。
- 使用 `databases` 字段申请一个或多个数据库。每个数据库名称将作为键注入到 `.Values.mysql.databases` 中。

**示例**
```yaml
middleware:
  mysql:
    username: mysqlclient
    databases:
      - name: aaa
```

## 映射环境变量

在应用的部署 YAML 中，将系统注入的 `.Values.mysql.*` 字段映射为应用所需的环境变量。

**示例**
```yaml
containers:
  - name: my-app
    # MySQL 对应的注入值如下
    env:
      - name: MDB_HOST
        value: "{{ .Values.mysql.host }}"

      - name: MDB_PORT
        value: "{{ .Values.mysql.port }}"

      - name: MDB_USER
        value: "{{ .Values.mysql.username }}"

      - name: MDB_PASSWORD
        value: "{{ .Values.mysql.password }}"

      # 数据库名称
      # 在 OlaresManifest 中配置的数据库名（例如：aaa）
      - name: MDB_DB
        value: "{{ .Values.mysql.databases.aaa }}"
```

## MySQL 变量参考

MySQL 运行时变量会在部署过程中注入到 `values.yaml` 中。这些变量由系统统一管理，用户无法自行修改。

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.mysql.host` | String | MySQL 数据库地址 |
| `.Values.mysql.port` | Number | MySQL 数据库端口 |
| `.Values.mysql.username` | String | MySQL 数据库用户名 |
| `.Values.mysql.password` | String | MySQL 数据库密码 |
| `.Values.mysql.databases` | Map<String,String> | 以申请的数据库名作为键。<br/>例如申请 `aaa`，可通过 `.Values.mysql.databases.aaa` 获取对应的值。 |