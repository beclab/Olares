---
outline: [2, 3]
description: Reference for predefined runtime values injected into application values.yaml during Olares deployment.
---

# Use predefined runtime values

During application deployment, Olares injects system-managed values into the application's `values.yaml`.

These predefined runtime values provide runtime context for:

- User and identity information  
- Storage paths  
- Cluster metadata  
- Application dependency endpoints  
- Middleware connection credentials  

:::info
These values are managed by the Olares system and cannot be edited.

Because they are Helm values (accessed via `.Values.*`), they are not automatically available inside containers. Map them explicitly in your templates if needed.
:::

## Usage example

To use these values in your application, reference them inside your Helm templates (e.g., `deployment.yaml`):

```yaml
# Example: mapping Olares runtime values to container environment variables
spec:
  containers:
    - name: my-app
      env:
        - name: APP_USER
          value: "{{ .Values.bfl.username }}"
        - name: DB_HOST
          value: "{{ .Values.postgres.host }}"
```

## Value reference

The "Type" column below describes the data type in Helm values and does not correspond to the `envs.type` enum in `OlaresManifest.yaml`.

### User and identity
| Value | Type | Description |
| -- | -- | -- |
| `.Values.bfl.username` | String | Current username. |
| `.Values.user.zone`    | String | Current user's domain.   |
| `.Values.admin`        | String | Administrator username. |

### Application and system information

| Value | Type | Description |
| -- | -- | -- |
| `.Values.domain` | Map<String,String> | Application entrance URLs in the format `entry_name => URL`. |
| `.Values.sysVersion` | String | Current system version. |
| `.Values.deviceName` | String  | Device name. |
| `.Values.downloadCdnURL` | String | CDN address for system resource downloads. |

### Storage paths

Olares injects predefined storage paths that applications can use for data storage and caching.

| Value | Type   | Description |
| -- | -- | -- |
| `.Values.userspace.appData` | String | Cluster storage path available to the application. Directory: `/Data/<appname>`. |
| `.Values.userspace.appCache` | String | Node-local cache path available to the application. Directory: `/Cache/<appname>`. |
| `.Values.userspace.userData` | String | User data directory. Path: `/Files/Home/`. |
| `.Values.sharedlib` | String | User external storage directory. Path: `/Files/External/<devicename>/`. |

### Cluster hardware metadata

Cluster hardware information is injected into `values.yaml` at deployment time.

| Value | Type | Description |
| -- | -- | -- |
| `.Values.cluster.arch` | String | Cluster CPU architecture (e.g., `amd64`). Mixed-architecture clusters are not supported. |
| `.Values.nodes` | List\<NodeInfo> | List of node hardware metadata objects injected under `values["nodes"]`. |

The value of `values["nodes"]` is a list of `NodeInfo` objects.

**Example `NodeInfo` structure**:

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

### Application dependencies

When an application declares a dependency on another application in `OlaresManifest.yaml`, Olares injects connection information into `values.yaml`.

| Value | Type | Description |
| -- | -- | -- |
| `.Values.deps` | Map<String,Value> | Host and port of the dependent application. |
| `.Values.svcs` | Map<String,Value> | Services and ports exposed by the dependent application. |

For `.Values.deps`, Olares injects:
- `<entry_name>_host`
- `<entry_name>_port`

**Example**:

```json
{
  "aserver_host": "aserver-svc.<namespace>",
  "aserver_port": 80
}
```
For `.Values.svcs`, Olares injects:
- `<service_name>_host`
- `<service_name>_port`

The `_port` value is a list. If a service exposes multiple ports, all ports are included.

**Example**:
```json
{
  "aserver-svc_host": "aserver-svc.<namespace>",
  "aserver-svc_port": [80]
}
```

### Middleware values

Middleware values are injected only when declared in the `middleware` section of `OlaresManifest.yaml`.

PostgreSQL and Redis are preinstalled. Other middleware (MongoDB, MinIO, RabbitMQ, MySQL, MariaDB, Elasticsearch) must be installed before use.

See [Middleware](/developer/develop/mw-overview.md#supported-services) for installation and configuration details.

#### Elasticsearch

| Value  | Type  | Description  |
|--|--|--|
|`.Values.elasticsearch.host`| String | Elasticsearch service host |
|`.Values.elasticsearch.port`| Number | Elasticsearch service port |
|`.Values.elasticsearch.username`| String | Elasticsearch username |
|`.Values.elasticsearch.password`| String | Elasticsearch password |
|`.Values.elasticsearch.indexes` | Map<String,String> | The requested index name is used<br> as the key. For example, if you request `aaa`, the value is available at `.Values.elasticsearch.indexes.aaa`. |

#### MariaDB

| Value  | Type  | Description  |
|--|--|--|
| `.Values.mariadb.host` | String | MariaDB database host |
| `.Values.mariadb.port` | Number | MariaDB database port |
| `.Values.mariadb.username` | String | MariaDB database username |
| `.Values.mariadb.password` | String | MariaDB database password |
| `.Values.mariadb.databases` | Map<String,String> | The requested database name is used as the key. <br/>For example, if you request `aaa`, the value is available at `.Values.mariadb.databases.aaa`. |

#### MinIO

| Value  | Type  | Description  |
|--|--|--|
| `.Values.minio.host` | String | MinIO service host |
| `.Values.minio.port` | Number | MinIO service port |
| `.Values.minio.username` | String | MinIO access key |
| `.Values.minio.password` | String | MinIO secret key |
| `.Values.minio.buckets` | Map<String,String> | The requested bucket name is used as the key. <br>For example, if you request `mybucket`, the value is available at `.Values.minio.buckets.mybucket`. |

#### MongoDB

| Value  | Type  | Description  |
|--|--|--|
| `.Values.mongodb.host` | String  | MongoDB database host |
| `.Values.mongodb.port` | Number  | MongoDB database port |
| `.Values.mongodb.username` | String | MongoDB database username |
| `.Values.mongodb.password`  | String | MongoDB database password |
| `.Values.mongodb.databases` | Map<String,String> | The requested database name is used as the key. <br/>For example, if you request `app_db`, the value is available at `.Values.mongodb.databases.app_db`. |

#### MySQL

| Value  | Type  | Description  |
|--|--|--|
| `.Values.mysql.host` | String | MySQL database host |
| `.Values.mysql.port` | Number | MySQL database port |
| `.Values.mysql.username` | String | MySQL database username |
| `.Values.mysql.password` | String | MySQL database password |
| `.Values.mysql.databases` | Map<String,String> | The requested database name is used as the key. <br/>For example, if you request `aaa`, the value is available at `.Values.mysql.databases.aaa`. |

#### PostgreSQL

| Value  | Type  | Description  |
|--|--|--|
| `.Values.postgres.host` | String  | PostgreSQL database host |
| `.Values.postgres.port` | Number | PostgreSQL database port |
| `.Values.postgres.username`  | String | PostgreSQL database username |
| `.Values.postgres.password`  | String | PostgreSQL database password |
| `.Values.postgres.databases` | Map<String,String> | The requested database name is used as the key. <br>For example, if you request `app_db`, the value is available at `.Values.postgres.databases.app_db`|

#### RabbitMQ

| Value  | Type  | Description  |
|--|--|--|
| `.Values.rabbitmq.host` | String | RabbitMQ service host |
| `.Values.rabbitmq.port` | Number | RabbitMQ service port |
| `.Values.rabbitmq.username` | String | RabbitMQ username |
| `.Values.rabbitmq.password` | String | RabbitMQ password |
| `.Values.rabbitmq.vhosts` | Map<String,String> | The requested vhost name is used as the key. <br/>For example, if you request `aaa`, the value is available at `.Values.rabbitmq.vhosts.aaa`. |

#### Redis

| Value  | Type  | Description  |
|--|--|--|
| `.Values.redis.host` | String | Redis service host |
| `.Values.redis.port` | Number  | Redis service port |
| `.Values.redis.password`| String | Redis service password |
| `.Values.redis.namespaces` | Map<String,String> | The requested namespace is used as the key. <br>For example, if you request `app_ns`, the value is available at `.Values.redis.namespaces.app_ns`. |