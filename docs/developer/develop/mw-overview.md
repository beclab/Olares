---
outline: [2, 3]
description: Learn what middleware is in Olares and navigate to access and integration guides for each supported service.
---
# Middleware in Olares

Middleware refers to infrastructure services that sit between your application and the system, providing data storage, messaging, and other common capabilities.

Our middleware documentation is organized into two types of guides:
- **Access and manage data**: Connect to a running service to inspect data and troubleshoot issues.
- **App integrate**: Configure your app to use a middleware service in Olares using `OlaresManifest.yaml`.

## Document types

### Access and manage data

Access and manage data guides explain how to connect to a running middleware service for administration.

Use these guides when you want to:
- Inspect stored data or indexes.
- Run queries or commands.
- Debug application behavior.
- Verify service status.

The access method (e.g., CLI, Dashboard, or Bytebase) depends on the service.

### App integration

App integration guides explain how to connect your application to a middleware service.

Use these guides when you want your application to:
- Declare dependencies in `OlaresManifest.yaml`.
- Request service resources.
- Read system-injected connection values in your application.

Integration is declarative and handled by Olares at deployment time.

## Supported services

### Databases and caching

| Service | Access and manage data | App integration |
| --- | --- | --- |
| Elasticsearch | [Access](./mw-view-es-data.md) | [Integrate](./mw-integrate-with-es.md) |
| MariaDB | [Access](./mw-view-mariadb-data.md) | [Integrate](./mw-integrate-with-mariadb.md) |
| MongoDB | [Access](./mw-view-mongodb-data.md) | [Integrate](./mw-integrate-with-mongodb.md) |
| MySQL | [Access](./mw-view-mysql-data.md) | [Integrate](./mw-integrate-with-mysql.md) |
| PostgreSQL | [Access](./mw-view-pg-data.md) | [Integrate](./mw-integrate-with-pg.md) |
| Redis | [Access](./mw-view-redis-data.md) | [Integrate](./mw-integrate-with-redis.md) |

### Messaging and streaming

| Service | Access and manage data | App integration |
| --- | --- | --- |
| RabbitMQ | [Access](./mw-view-rabbitmq-data.md) | [Integrate](./mw-integrate-with-rabbitmq.md) |
| NATS | [Access](./mw-view-nats-data.md) | — |

### Storage and observability

| Service | Access and manage data | App integration |
| --- | --- | --- |
| MinIO | [Access](./mw-view-minio-data.md) | [Integrate](./mw-integrate-with-minio.md) |
| Grafana | [Access](./mw-view-grafana-data.md) | — |
| OpenTelemetry | [Access](./mw-view-otel-data.md) | — |