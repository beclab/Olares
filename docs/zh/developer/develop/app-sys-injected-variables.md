---
outline: [2, 4]
description: Olares 在应用部署期间注入到 `application values.yaml` 中的运行时变量。
---

# 系统注入的运行时变量

在部署时，Olares 会自动向应用的 `values.yaml` 注入由系统管理的变量。这些变量为只读，涵盖用户身份、存储路径、集群元数据、应用依赖以及中间件凭据等信息。

由于它们属于 Helm values，因此不会自动传递到容器内部。如需在容器内部使用，请在部署模板中通过 `env:` 显式映射。

## 在应用中使用

你可以在 Helm 模板（如 `deployment.yaml`）中直接引用这些值。

**示例**：将当前用户名和 Postgres 主机地址传入容器环境变量。

```yaml
# 将系统注入的运行时变量传入容器环境变量
spec:
  containers:
    - name: my-app
      env:
        - name: APP_USER
          value: "{{ .Values.bfl.username }}"
        - name: DB_HOST
          value: "{{ .Values.postgres.host }}"
```

完整变量列表请参见[变量参考](#变量参考)。

## 变量参考

“类型”列描述的是 Helm value 的数据类型，并不对应 `OlaresManifest.yaml` 中的 `type` 字段。

### 用户与身份信息
| 变量 | 类型 | 说明 |
| -- | -- | -- |
| `.Values.bfl.username` | String | 当前用户名。 |
| `.Values.user.zone`    | String | 当前用户的域名。 |
| `.Values.admin`        | String | 管理员用户名。 |

### 应用与系统信息

| 变量 | 类型 | 说明 |
| -- | -- | -- |
| `.Values.domain` | Map<String,String> | 应用入口地址映射，每个条目将入口名称映射为对应的 URL。 |
| `.Values.sysVersion` | String | 系统版本号。 |
| `.Values.deviceName` | String  | 设备名称。 |
| `.Values.downloadCdnURL` | String | 系统资源下载使用的 CDN 地址。 |

### 存储路径

| 变量 | 类型 | 说明 |
| -- | -- | -- |
| `.Values.userspace.appData` | String | 应用的集群存储路径，路径为 `/Data/<appname>`。 |
| `.Values.userspace.appCache` | String | 应用的节点本地缓存路径，路径为 `/Cache/<appname>`。 |
| `.Values.userspace.userData` | String | 用户数据目录，路径为 `/Files/Home/`。 |
| `.Values.sharedlib` | String | 用户外部存储目录，路径为 `/Files/External/<devicename>/`。 |

### 集群硬件信息

集群硬件信息会在部署时注入到 `values.yaml` 中。

| 变量 | 类型 | 说明 |
| -- | -- | -- |
| `.Values.cluster.arch` | String | 集群 CPU 架构（例如 `amd64`）。Olares 目前不支持混合架构组成集群。 |
| `.Values.nodes` | List\<NodeInfo> | 节点硬件元数据列表，注入在 `values["nodes"]`下。 |

`.Values.nodes` 中每个条目结构如下：

```json
[
  {
    "cudaVersion": "12.9",
    "cpu": [
      {
        "coreNumber": 16,
        "arch": "amd64",
        "frequency": 4900000000,
        "model": "151",
        "modelName": "12th Gen Intel(R) Core(TM) i5-12600KF",
        "vendor": "GenuineIntel"
      }
    ],
    "memory": {
      "total": 50351353856
    },
    "gpus": [
      {
        "vendor": "NVIDIA",
        "arch": "Ada Lovelace",
        "model": "4060",
        "memory": 17175674880,
        "modelName": "NVIDIA GeForce RTX 4060 Ti"
      }
    ]
  }
]
```

### 应用依赖

当应用在 `OlaresManifest.yaml` 中声明对其他应用的依赖时，Olares 会将连接信息注入到 `values.yaml`。

| 变量 | 类型 | 说明 |
| -- | -- | -- |
| `.Values.deps` | Map<String,Value> | 每个声明依赖的服务主机地址与端口。键名格式为 `<entry_name>_host` 与 `<entry_name>_port`。|
| `.Values.svcs` | Map<String,Value> | 每个声明依赖的所有服务主机地址与端口。键名格式为 `<service_name>_host` 与 `<service_name>_port`。端口值为列表类型，用于支持多端口服务。 |

**示例**：依赖入口名为 `aserver`，服务名为 `aserver-svc`。

`.Values.deps`：

```json
{
  "aserver_host": "aserver-svc.<namespace>",
  "aserver_port": 80
}
```

`.Values.svcs`：

```json
{
  "aserver-svc_host": "aserver-svc.<namespace>",
  "aserver-svc_port": [80]
}
```

### 中间件变量

仅当在 `OlaresManifest.yaml` 的 `middleware` 部分声明中间件依赖时，才会注入对应变量。

PostgreSQL 与 Redis 为预安装组件。MongoDB、MinIO、RabbitMQ、MySQL 和 MariaDB 需要单独安装后方可使用。

#### MariaDB

安装与配置详情请参见[集成 MariaDB](/zh/developer/develop/mw-integrate-with-mariadb.md)。

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.mariadb.host` | String | MariaDB 主机地址。 |
| `.Values.mariadb.port` | Number | MariaDB 端口。 |
| `.Values.mariadb.username` | String | MariaDB 用户名。 |
| `.Values.mariadb.password` | String | MariaDB 密码。 |
| `.Values.mariadb.databases` | Map<String,String> | 请求的数据库集合，按数据库名为键。<br/>例如申请 `app_db`，可通过 `.Values.mariadb.databases.app_db` 获取对应的值。 |

#### MinIO

安装与配置详情请参见[集成 MinIO](/zh/developer/develop/mw-integrate-with-minio.md)。

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.minio.host` | String | MinIO 服务地址。 |
| `.Values.minio.port` | Number | MinIO 服务端口。 |
| `.Values.minio.username` | String | MinIO 访问密钥。 |
| `.Values.minio.password` | String | MinIO 密钥。 |
| `.Values.minio.buckets` | Map<String,String> | 请求的存储桶集合，按桶名为键。例如申请 `mybucket`，可通过 `.Values.minio.buckets.mybucket` 获取对应的值。 |

#### MongoDB

安装与配置详情请参见[集成 MongoDB](/zh/developer/develop/mw-integrate-with-mongodb.md)。

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.mongodb.host` | String | MongoDB 主机地址。 |
| `.Values.mongodb.port` | Number | MongoDB 端口。 |
| `.Values.mongodb.username` | String | MongoDB 用户名。 |
| `.Values.mongodb.password` | String  | MongoDB 密码。 |
| `.Values.mongodb.databases` | Map<String,String> | 请求的数据库集合，按数据库名为键。<br/>例如申请 `app_db`，可通过 `.Values.mongodb.databases.app_db` 获取对应的值。 |

#### MySQL

安装与配置详情请参见[集成 MySQL](/zh/developer/develop/mw-integrate-with-mysql.md)。

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.mysql.host` | String | MySQL 主机地址。 |
| `.Values.mysql.port` | Number | MySQL端口。 |
| `.Values.mysql.username` | String | MySQL 用户名。 |
| `.Values.mysql.password` | String | MySQL 密码。 |
| `.Values.mysql.databases` | Map<String,String> | 请求的数据库集合，按数据库名为键。<br/>例如申请 `app_db`，可通过 `.Values.mysql.databases.app_db` 获取对应的值。 |

#### PostgreSQL

安装与配置详情请参见[集成 PostgreSQL](/zh/developer/develop/mw-integrate-with-pg.md)。

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.postgres.host` | String  | PostgreSQL 主机地址。 |
| `.Values.postgres.port` | Number | PostgreSQL 端口。 |
| `.Values.postgres.username`  | String | PostgreSQL 用户名。 |
| `.Values.postgres.password`  | String | PostgreSQL 密码。 |
| `.Values.postgres.databases` | Map<String,String> | 请求的数据库集合，按数据库名为键。例如，若申请的数据库名为 `app_db`，可通过 `.Values.postgres.databases.app_db`获取对应值。|

#### RabbitMQ

安装与配置详情请参见[集成 RabbitMQ](/zh/developer/develop/mw-integrate-with-rabbitmq.md)。

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.rabbitmq.host` | String | RabbitMQ 主机地址。 |
| `.Values.rabbitmq.port` | Number | RabbitMQ 端口。 |
| `.Values.rabbitmq.username` | String | RabbitMQ 用户名。 |
| `.Values.rabbitmq.password` | String | RabbitMQ 密码。 |
| `.Values.rabbitmq.vhosts` | Map<String,String> | 请求的虚拟主机集合，按名称为键。<br/>例如申请 `myvhost`，可通过 `.Values.rabbitmq.vhosts.myvhost` 获取对应的值。 |

#### Redis

安装与配置详情请参见[集成 Redis](/zh/developer/develop/mw-integrate-with-redis.md)。

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.redis.host` | String | Redis 主机地址。 |
| `.Values.redis.port` | Number  | Redis 端口。 |
| `.Values.redis.password`| String | Redis 密码。 |
| `.Values.redis.namespaces` | Map<String, String> | 请求的命名空间集合，按名称为键。<br>例如，请求 `app_ns`，可通过 `.Values.redis.namespaces.app_ns`获取对应值。 |