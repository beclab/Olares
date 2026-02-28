---
outline: [2, 3]
description: Olares 在应用部署期间注入到 `application values.yaml` 中的预定义运行时变量。
---

# 使用预定义运行时变量

在应用部署过程中，Olares 会向应用的 `values.yaml` 注入系统管理的运行时变量（Helm Values）。

这些预定义运行时变量为应用提供以下运行时上下文：

- 用户与身份信息
- 存储路径  
- 集群元数据  
- 应用依赖  
- 中间件连接凭证  

:::info
这些变量由 Olares 系统管理，无法被用户编辑。

由于它们属于 Helm 值（通过 `.Values.*` 访问），不会自动出现在容器环境中。如需在容器中使用，请在模板中显式映射。
:::

## 使用示例

在应用中使用这些变量时，可在 Helm 模板（例如 `deployment.yaml`）中引用：

```yaml
# 示例：将 Olares 运行时变量映射为容器环境变量
spec:
  containers:
    - name: my-app
      env:
        - name: APP_USER
          value: "{{ .Values.bfl.username }}"
        - name: DB_HOST
          value: "{{ .Values.postgres.host }}"
```

## 变量参考

下表中的类型表示 Helm values 的数据类型，并不对应 `OlaresManifest.yaml` 中 `envs.type` 的枚举定义。

### 用户与身份信息
| 变量 | 类型 | 说明 |
| -- | -- | -- |
| `.Values.bfl.username` | String | 当前用户名 |
| `.Values.user.zone`    | String | 当前用户的域名 |
| `.Values.admin`        | String | 管理员用户名 |

### 应用与系统信息

| 变量 | 类型 | 说明 |
| -- | -- | -- |
| `.Values.domain` | Map<String,String> | 应用入口地址映射，格式为 `entry_name => URL` |
| `.Values.sysVersion` | String | 系统版本号 |
| `.Values.deviceName` | String  | 设备名称 |
| `.Values.downloadCdnURL` | String | 系统资源下载使用的 CDN 地址 |

### 存储路径

Olares 会注入预定义存储路径，供应用进行数据存储与缓存。

| 变量 | 类型 | 说明 |
| -- | -- | -- |
| `.Values.userspace.appData` | String | 应用可用的集群存储路径，目录为 `/Data/<appname>` |
| `.Values.userspace.appCache` | String | 应用可用的节点本地缓存路径，目录为 `/Cache/<appname>` |
| `.Values.userspace.userData` | String | 用户数据目录，路径为 `/Files/Home/` |
| `.Values.sharedlib` | String | 用户外部存储目录，路径为 `/Files/External/<devicename>/` |

### 集群硬件信息

集群硬件信息会在部署时注入到 `values.yaml` 中。

| 变量 | 类型 | 说明 |
| -- | -- | -- |
| `.Values.cluster.arch` | String | 集群 CPU 架构（例如 `amd64`）。不支持混合架构集群。 |
| `.Values.nodes` | List\<NodeInfo> | 节点硬件元数据列表，注入在 `values["nodes"]`下。 |

`values["nodes"]` 的值为 `NodeInfo` 对象列表。

**`NodeInfo` 结构示例**：

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
| `.Values.deps` | Map<String,Value> | 当前依赖应用的地址和端口信息 |
| `.Values.svcs` | Map<String,Value> | 依赖应用暴露的服务及其端口信息 |

对于 `.Values.deps`，诸如格式为：
- `<entry_name>_host`
- `<entry_name>_port`

**示例**：

```json
{
  "aserver_host": "aserver-svc.<namespace>",
  "aserver_port": 80
}
```
对于 `.Values.svcs`，注入格式为：
- `<service_name>_host`
- `<service_name>_port`

`_port` 为列表类型。如果服务暴露多个端口，将全部包含在列表中。

**示例**：
```json
{
  "aserver-svc_host": "aserver-svc.<namespace>",
  "aserver-svc_port": [80]
}
```

### 中间件变量

仅当在 `OlaresManifest.yaml` 的 `middleware` 部分声明时，才会注入对应的中间件变量。

PostgreSQL 与 Redis 默认预装。
其他中间件服务（MongoDB、MinIO、RabbitMQ、MySQL、MariaDB、Elasticsearch）需在使用前安装。

详情请参考[中间件](/zh/developer/develop/mw-overview.md#支持的服务)。

#### Elasticsearch

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.elasticsearch.host` | String | Elasticsearch 服务地址 |
| `.Values.elasticsearch.port` | Number  | Elasticsearch 服务端口 |
| `.Values.elasticsearch.username` | String | Elasticsearch 用户名 |
| `.Values.elasticsearch.password` | String | Elasticsearch 密码 |
| `.Values.elasticsearch.indexes`  | Map<String,String> | 以申请的索引名称作为键。例如申请 `aaa`，可通过 `.Values.elasticsearch.indexes.aaa` 获取对应的值。 |

#### MariaDB

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.mariadb.host` | String | MariaDB 数据库地址 |
| `.Values.mariadb.port` | Number | MariaDB 数据库端口 |
| `.Values.mariadb.username` | String | MariaDB 数据库用户名 |
| `.Values.mariadb.password` | String | MariaDB 数据库密码 |
| `.Values.mariadb.databases` | Map<String,String> | 以申请的数据库名作为键。<br/>例如申请 `aaa`，可通过 `.Values.mariadb.databases.aaa` 获取对应的值。 |

#### MinIO

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.minio.host` | String | MinIO 服务地址 |
| `.Values.minio.port` | Number | MinIO 服务端口 |
| `.Values.minio.username` | String | MinIO 访问密钥（Access Key） |
| `.Values.minio.password` | String | MinIO 密钥（Secret Key） |
| `.Values.minio.buckets` | Map<String,String> | 以申请的存储桶名称作为键。例如申请 `mybucket`，可通过 `.Values.minio.buckets.mybucket` 获取对应的值。 |

#### MongoDB

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.mongodb.host` | String | MongoDB 数据库地址 |
| `.Values.mongodb.port` | Number | MongoDB 数据库端口 |
| `.Values.mongodb.username` | String | MongoDB 数据库用户名 |
| `.Values.mongodb.password` | String  | MongoDB 数据库密码 |
| `.Values.mongodb.databases` | Map<String,String> | 以申请的数据库名作为键。<br/>例如申请 `app_db`，可通过 `.Values.mongodb.databases.app_db` 获取对应的值。 |

#### MySQL

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.mysql.host` | String | MySQL 数据库地址 |
| `.Values.mysql.port` | Number | MySQL 数据库端口 |
| `.Values.mysql.username` | String | MySQL 数据库用户名 |
| `.Values.mysql.password` | String | MySQL 数据库密码 |
| `.Values.mysql.databases` | Map<String,String> | 以申请的数据库名作为键。<br/>例如申请 `aaa`，可通过 `.Values.mysql.databases.aaa` 获取对应的值。 |

#### PostgreSQL

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.postgres.host` | String  | PostgreSQL 数据库地址 |
| `.Values.postgres.port` | Number | PostgreSQL 数据库端口 |
| `.Values.postgres.username`  | String | PostgreSQL 数据库用户名 |
| `.Values.postgres.password`  | String | PostgreSQL 数据库密码 |
| `.Values.postgres.databases` | Map<String,String> | PostgreSQL 数据库以申请的数据库名作为键。例如，若申请的数据库名为 `app_db`，可通过 `.Values.postgres.databases.app_db`获取对应值。|

#### RabbitMQ

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.rabbitmq.host` | String | RabbitMQ 服务地址 |
| `.Values.rabbitmq.port` | Number | RabbitMQ 服务端口 |
| `.Values.rabbitmq.username` | String | RabbitMQ 用户名 |
| `.Values.rabbitmq.password` | String | RabbitMQ 密码 |
| `.Values.rabbitmq.vhosts` | Map<String,String> | 以申请的 vhost 名作为键。<br/>例如申请 `aaa`，可通过 `.Values.rabbitmq.vhosts.aaa` 获取对应的值。 |

#### Redis

| 变量 | 类型 | 说明 |
|--|--|--|
| `.Values.redis.host` | String | Redis 数据库地址 |
| `.Values.redis.port` | Number  | Redis 数据库端口 |
| `.Values.redis.password`| String | Redis 数据库密码 |
| `.Values.redis.namespaces` | Map<String, String> | Redis 命名空间名称，以申请命名空间作为键。例如，若申请的命名空间名为 `app_ns`，可通过 `.Values.redis.namespaces.app_ns`获取对应值。 |