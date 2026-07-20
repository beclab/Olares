# Database

The Olares system provides three most popular data storage cluster for all APPs, covering `RDS`, `NoSQL`, and `Cache` data storage use cases.

## RDS

The system has deployed **PostgreSQL** and provides two types of databases.

- **Standalone PostgreSQL**, providing the most commonly used `RDS` database layer functions.
- **Distributed PostgreSQL** extension, powered by **Citus**. Provides the ability to horizontally scale the database.

When setting up a **PostgreSQL** database, you can specify the type of database to be used in [OlaresManifest.yaml](../package/manifest.md#middleware).

```yaml
middleware:
postgres:
  username: postgres
  databases:
    - name: db
      distributed: true # Whether the database is distributed in the cluster.
```

If you use **Citus**, **Olares** will automatically shard the database tables and perform rebalancing during the horizontal scaling of **PostgreSQL** replicas.

## NoSQL

The NoSQL cluster is not deployed by default in Olares, but it can be easily installed from the Market. To set up a NoSQL cluster, the administrator needs to install the [**MongoDB**](https://market.olares.com/middleware/mongodb) middleware. Once installed, the [Percona Operator for MongoDB](https://github.com/percona/percona-server-mongodb-operator) automatically manages the **MongoDB** cluster. Users can then horizontally scale **MongoDB** cluster replicas, as well as perform backup and restore operations on databases.

You can specify detailed configuration for MongoDB in [OlaresManifest.yaml](../package/manifest.md#middleware) as follows:

```yaml
middleware:
  mongodb:
    username: mongodb
    databases:
      - name: db0
      - name: db1
options:
  dependencies:
    - name: olares
      type: system
      version: ">=1.6.0-0"
    - name: mongodb
      version: ">=6.0.0-0"
      type: middleware
```

## Cache

In terms of the Cache cluster, Olares uses Redis Cluster. The cluster is managed by a customized [Redis Cluster Operator](https://github.com/beclab/redis-cluster-operator) to achieve cloud nativeness. It enables us to scale replicas horizontally in a convenient and effective manner.

To ensure **data isolation** between users and apps in the **Redis cluster**, the **Olares** system has added a **Redis cluster proxy**. It isolates data based on the `namespace`. This operation is transparent, meaning app developers typically do not need to be aware of it.

Additionally, this proxy simplifies the process of connecting to clusters. It eliminates the need to switch from a **standalone Redis Client** to a **Redis Cluster client** in the app, thus simplifying app code modifications.

```
middleware:
  redis:
    password: password
    namespace: db0
```

:::info NOTE
Since Olares uses the Redis Cluster version, developers need to understand the usage restrictions of Redis Cluster in detail when using it.
:::
