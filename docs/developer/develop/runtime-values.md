---
outline: [2, 3]
description: Reference for predefined runtime values injected into application values.yaml during Olares deployment.
---

# Predefined runtime values

At deployment, Olares automatically injects a set of system-managed values into the app's `values.yaml`. These values are read-only and cover user identity, storage paths, cluster metadata, app dependencies, and middleware credentials.

Because they are Helm values, they are not automatically available inside containers. To pass one into a container, map it explicitly under `env:` in your deployment template.

## Use in your app

Reference these values directly in your Helm templates, such as `deployment.yaml`.

**Example**: pass the current username and Postgres host into container environment variables.

```yaml
# Pass predefined runtime values into container environment variables
spec:
  containers:
    - name: my-app
      env:
        - name: APP_USER
          value: "{{ .Values.bfl.username }}"
        - name: DB_HOST
          value: "{{ .Values.postgres.host }}"
```

For the full list of available values, see [Value reference](#value-reference).

## Value reference

The Type column describes the Helm value data type. It does not correspond to the `type` field in `OlaresManifest.yaml`.

### User and identity

| Value | Type | Description |
| --- | --- | --- |
| `.Values.bfl.username` | String | Current username. |
| `.Values.user.zone` | String | Current user's domain. |
| `.Values.admin` | String | Administrator username. |

### Application and system information

| Value | Type | Description |
| --- | --- | --- |
| `.Values.domain` | Map\<String,String> | App entrance URLs. Each entry maps an entrance name to its URL. |
| `.Values.sysVersion` | String | Current Olares system version. |
| `.Values.deviceName` | String | Device name. |
| `.Values.downloadCdnURL` | String | CDN address used for system resource downloads. |

### Storage paths

| Value | Type | Description |
| --- | --- | --- |
| `.Values.userspace.appData` | String | Cluster storage path for the app. Path: `/Data/<appname>`. |
| `.Values.userspace.appCache` | String | Node-local cache path for the app. Path: `/Cache/<appname>`. |
| `.Values.userspace.userData` | String | User's home data directory. Path: `/Files/Home/`. |
| `.Values.sharedlib` | String | User's external storage directory. Path: `/Files/External/<devicename>/`. |

### Cluster hardware metadata

| Value | Type | Description |
| --- | --- | --- |
| `.Values.cluster.arch` | String | Cluster CPU architecture, such as `amd64`. Mixed-architecture clusters are not supported. |
| `.Values.nodes` | List\<NodeInfo> | Hardware metadata for each node in the cluster. |

Each entry in `.Values.nodes` follows this structure:

```json
// Single entry in the .Values.nodes list
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

When an app declares a dependency in `OlaresManifest.yaml`, Olares injects connection information into `values.yaml`.

| Value | Type | Description |
| --- | --- | --- |
| `.Values.deps` | Map\<String,Value> | Main entry host and port for each declared dependency. Keys follow the pattern `<entry_name>_host` and `<entry_name>_port`. |
| `.Values.svcs` | Map\<String,Value> | All service hosts and ports for each declared dependency. Keys follow the pattern `<service_name>_host` and `<service_name>_port`. Port values are lists to support multiple ports per service. |

**Example**: for a dependency with entry name `aserver` and service name `aserver-svc`.

`.Values.deps`:
```json
{
  "aserver_host": "aserver-svc.<namespace>",
  "aserver_port": 80
}
```

`.Values.svcs`:
```json
{
  "aserver-svc_host": "aserver-svc.<namespace>",
  "aserver-svc_port": [80]
}
```

### Middleware values

Middleware values are injected only after you declare the middleware dependency in the `middleware` section of `OlaresManifest.yaml`.

PostgreSQL and Redis are preinstalled. MongoDB, MinIO, RabbitMQ, MySQL, MariaDB, and Elasticsearch must be installed separately before your app can use them.

See [Middleware](/developer/develop/mw-overview.md#supported-services) for installation and configuration details.

#### Elasticsearch

| Value | Type | Description |
| --- | --- | --- |
| `.Values.elasticsearch.host` | String | Elasticsearch service host. |
| `.Values.elasticsearch.port` | Number | Elasticsearch service port. |
| `.Values.elasticsearch.username` | String | Elasticsearch username. |
| `.Values.elasticsearch.password` | String | Elasticsearch password. |
| `.Values.elasticsearch.indexes` | Map\<String,String> | Requested indexes, keyed by index name. For example, a request for `aaa` is available at `.Values.elasticsearch.indexes.aaa`. |

#### MariaDB

| Value | Type | Description |
| --- | --- | --- |
| `.Values.mariadb.host` | String | MariaDB host. |
| `.Values.mariadb.port` | Number | MariaDB port. |
| `.Values.mariadb.username` | String | MariaDB username. |
| `.Values.mariadb.password` | String | MariaDB password. |
| `.Values.mariadb.databases` | Map\<String,String> | Requested databases, keyed by database name. For example, a request for `app_db` is available at `.Values.mariadb.databases.app_db`. |

#### MinIO

| Value | Type | Description |
| --- | --- | --- |
| `.Values.minio.host` | String | MinIO service host. |
| `.Values.minio.port` | Number | MinIO service port. |
| `.Values.minio.username` | String | MinIO access key. |
| `.Values.minio.password` | String | MinIO secret key. |
| `.Values.minio.buckets` | Map\<String,String> | Requested buckets, keyed by bucket name. For example, a request for `mybucket` is available at `.Values.minio.buckets.mybucket`. |

#### MongoDB

| Value | Type | Description |
| --- | --- | --- |
| `.Values.mongodb.host` | String | MongoDB host. |
| `.Values.mongodb.port` | Number | MongoDB port. |
| `.Values.mongodb.username` | String | MongoDB username. |
| `.Values.mongodb.password` | String | MongoDB password. |
| `.Values.mongodb.databases` | Map\<String,String> | Requested databases, keyed by database name. For example, a request for `app_db` is available at `.Values.mongodb.databases.app_db`. |

#### MySQL

| Value | Type | Description |
| --- | --- | --- |
| `.Values.mysql.host` | String | MySQL host. |
| `.Values.mysql.port` | Number | MySQL port. |
| `.Values.mysql.username` | String | MySQL username. |
| `.Values.mysql.password` | String | MySQL password. |
| `.Values.mysql.databases` | Map\<String,String> | Requested databases, keyed by database name. For example, a request for `app_db` is available at `.Values.mysql.databases.app_db`. |

#### PostgreSQL

| Value | Type | Description |
| --- | --- | --- |
| `.Values.postgres.host` | String | PostgreSQL host. |
| `.Values.postgres.port` | Number | PostgreSQL port. |
| `.Values.postgres.username` | String | PostgreSQL username. |
| `.Values.postgres.password` | String | PostgreSQL password. |
| `.Values.postgres.databases` | Map\<String,String> | Requested databases, keyed by database name. For example, a request for `app_db` is available at `.Values.postgres.databases.app_db`. |

#### RabbitMQ

| Value | Type | Description |
| --- | --- | --- |
| `.Values.rabbitmq.host` | String | RabbitMQ host. |
| `.Values.rabbitmq.port` | Number | RabbitMQ port. |
| `.Values.rabbitmq.username` | String | RabbitMQ username. |
| `.Values.rabbitmq.password` | String | RabbitMQ password. |
| `.Values.rabbitmq.vhosts` | Map\<String,String> | Requested vhosts, keyed by vhost name. For example, a request for `myvhost` is available at `.Values.rabbitmq.vhosts.myvhost`. |

#### Redis

| Value | Type | Description |
| --- | --- | --- |
| `.Values.redis.host` | String | Redis host. |
| `.Values.redis.port` | Number | Redis port. |
| `.Values.redis.password` | String | Redis password. |
| `.Values.redis.namespaces` | Map\<String,String> | Requested namespaces, keyed by namespace name. For example, a request for `app_ns` is available at `.Values.redis.namespaces.app_ns`. |
