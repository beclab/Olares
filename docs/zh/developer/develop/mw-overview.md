---
outline: [2, 3]
description: 了解 Olares 中的 middleware，并快速导航到各个已支持服务的数据访问与集成指南。
---
# Olares 中的中间件

中间件指位于应用与系统之间的基础设施服务，用于提供数据存储、消息通信等通用能力。

Olares 的中间件指南分为两类：
- **访问与管理数据**：连接到正在运行的服务，用于查看数据和排查问题。
- **应用集成**：通过 `OlaresManifest.yaml` 配置应用，使其使用某个中间件服务。

## 文档类型

### 访问与管理数据

访问与管理数据类指南说明如何连接到正在运行的中间件服务并进行管理操作。

当你需要执行以下操作时，请阅读这类指南：
- 查看已存储的数据或索引
- 执行查询或命令
- 调试应用行为
- 验证服务运行状态

具体的访问方式（例如 CLI、dashboard 或 Bytebase）因服务而异，并会在对应的指南中说明。

### 应用集成

应用集成类指南说明如何将应用连接到某个中间件服务。

当你的应用需要执行以下操作时，请阅读这类指南：
- 在 `OlaresManifest.yaml` 中声明依赖
- 请求所需的服务资源
- 在应用中读取系统注入的连接信息

集成过程是声明式的，并由 Olares 在部署阶段自动完成。

## 支持的服务

### 数据库与缓存

| 服务 | 访问与管理数据 | 应用集成 |
| --- | --- | --- |
| Elasticsearch | [访问](./mw-view-es-data.md) | [集成](./mw-integrate-with-es.md) |
| MariaDB | [访问](./mw-view-mariadb-data.md) | [集成](./mw-integrate-with-mariadb.md) |
| MongoDB | [访问](./mw-view-mongodb-data.md) | [集成](./mw-integrate-with-mongodb.md) |
| MySQL | [访问](./mw-view-mysql-data.md) | [集成](./mw-integrate-with-mysql.md) |
| PostgreSQL | [访问](./mw-view-pg-data.md) | [集成](./mw-integrate-with-pg.md) |
| Redis | [访问](./mw-view-redis-data.md) | [集成](./mw-integrate-with-redis.md) |

### 消息与流处理

| 服务 | 访问与管理数据 | 应用集成 |
| --- | --- | --- |
| RabbitMQ | [访问](./mw-view-rabbitmq-data.md) | [集成](./mw-integrate-with-rabbitmq.md) |
| NATS | [访问](./mw-view-nats-data.md) | — |

### 存储与可观测性

| 服务 | 访问与管理数据 | 应用集成 |
| --- | --- | --- |
| MinIO | [访问](./mw-view-minio-data.md) | [集成](./mw-integrate-with-minio.md) |
| Grafana | [访问](./mw-view-grafana-data.md) | — |
| OpenTelemetry | [访问](./mw-view-otel-data.md) | — |
