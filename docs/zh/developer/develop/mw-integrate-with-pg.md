---
outline: [2, 3]
description: 了解如何在 Olares 中将应用接入内置的 PostgreSQL 服务。
---
# 集成 PostgreSQL

通过在 `OlaresManifest.yaml` 中声明依赖，并将系统注入的配置值映射到容器的环境变量中，即可在应用中使用 Olares 提供的 PostgreSQL 服务。

:::info PostgreSQL 已安装
PostgreSQL 服务在 Olares 中默认已安装。
:::

## 配置 `OlaresManifest.yaml`

在 `OlaresManifest.yaml` 中添加所需的中间件配置。

- 使用 `scripts` 字段指定数据库创建完成后需要执行的 SQL 脚本。
- 使用 `extensions` 字段为数据库添加所需的扩展。

:::info 脚本中的变量注入
系统会提供两个变量 `$databasename` 和 `$dbusername`，在执行脚本时由 Olares 应用运行时自动替换。
:::

**示例**
```yaml
middleware:
  postgres:
    username: immich
    databases:
    - name: immich
      extensions:
      - vectors
      - earthdistance
      scripts:
      - BEGIN;
      - ALTER DATABASE $databasename SET search_path TO "$user", public, vectors;
      - ALTER SCHEMA vectors OWNER TO $dbusername;
      - COMMIT;
```

## 注入环境变量

在应用的部署 YAML 中，将系统注入的 `.Values.postgres.*` 字段映射为应用所使用的环境变量。

**示例**
```yaml
containers:
  - name: my-app
    env:
      # 在 OlaresManifest 中配置的数据库名称，
      # 对应 middleware.postgres.databases[i].name
      # 注意：将 <dbname> 替换为 Manifest 中定义的实际名称（例如 immich）
      - name: DB_POSTGRESDB_DATABASE
        value: {{ .Values.postgres.databases.<dbname> }}
      
      # 主机地址
      - name: DB_POSTGRESDB_HOST
        value: {{ .Values.postgres.host }}
      
      # 端口
      - name: DB_POSTGRESDB_PORT
        value: "{{ .Values.postgres.port }}"
      
      # 用户名
      - name: DB_POSTGRESDB_USER
        value: {{ .Values.postgres.username }}
      
      # 密码
      - name: DB_POSTGRESDB_PASSWORD
        value: {{ .Values.postgres.password }}
```

## PostgreSQL Values 参考
PostgreSQL Values 是在部署过程中由系统自动注入到 `values.yaml` 中的预定义变量。这些值由系统统一管理，用户无法自行修改。

| 键  | 类型  | 说明  |
|--|--|--|
| `.Values.postgres.host` | String  | PostgreSQL 数据库地址 |
| `.Values.postgres.port` | Number | PostgreSQL 数据库端口 |
| `.Values.postgres.username`  | String | PostgreSQL 数据库用户名 |
| `.Values.postgres.password`  | String | PostgreSQL 数据库密码 |
| `.Values.postgres.databases` | Map<String,String> | PostgreSQL 数据库以申请的数据库名作为键。例如，若申请的数据库名为 `app_db`，可通过 `.Values.postgres.databases.app_db`获取对应值。|