---
outline: [2, 3]
description: Olares data management architecture, explaining file system types, application storage paths and database support. Covers technical specifications of JuiceFS, PostgreSQL, MongoDB and Redis.
---

# Data

User data is usually stored in file systems and databases. Of the two, databases are built based on file systems. Here are our design philosophies with them:

**For file systems**:

  Olares is designed for multi-node clusters. Therefore, developers need to consider the access to the file system when the program is scheduled to different nodes when developing applications. We want to shield these details from developers.

**For databases**:

- For common databases, developers only need to modify the configuration to complete the integration.
- Different users and applications can share physical database instances to save resource overhead.

**For both**:

- Data between different users and different applications are isolated from each other.
- Scalable and highly available.
- Capable of performing unified backup and restore at the system level.

## File System Type

### JuiceFS

Olares OS uses [JuiceFS](https://juicefs.com) as the underlying multi-physical node shared file system solution. In this way, applications can obtain cross-node file access using the simplest HostPath PV method. This allows Pods to be freely scheduled in the cluster.

As for the back-end object storage solution of JuiceFS, we also provide two solutions: S3 and MinIO.

By default, Olares uses the local file system (FS) when installed locally. However, if the `--with-juicefs=true` option is specified when running the [`olares-cli prepare`](../../developer/install/cli/prepare.md) command, JuiceFS will be installed and used. Additionally, a MinIO instance will be set up as the backend storage.

### Local disk

In some application systems, intensive file system read and write operations may occur. These intensive file system read and write operations are often fragmented random reads and writes. In various existing distributed storage cluster solutions, for such intensive fragmented random read and write operations, it is very likely that I/O or CPU consumption will be too high (usually due to high I/O Wait).

The best practice provided by Olares is to make full use of the node's local hard disk as a file buffer. Although the local hard disk of the node has limited capacity, it has high-speed read and write performance because it basically uses SSD hard disk. If the application reads and writes files, it will be buffered on the local hard disk of the node, and then written asynchronously to the distributed file system in batches. This can turn most of the fragmented random reads and writes into a few sequential reads and writes. This greatly improves system I/O efficiency.

## Application Storage Path

For applications, there are 3 different storage paths to deal with different usage scenarios.

### UserData

The `UserData` storage path stores files that change infrequently but require cross-application access, such as documents, photos, and videos.

Applications can obtain access permissions to a directory under the Home directory by applying for [UserData](../../developer/develop/package/manifest.md#userdata) permissions in `OlaresManifest.yaml`. For example, you can request permissions to the Picture directory for PhotoPrism, and permissions to the Downloads directory for qBittorrent and Jellyfin.

### AppData

The `AppData` storage path stores data that does not change frequently but needs to span across nodes. For example, configuration files.

Applications can apply for [AppData](../../developer/develop/package/manifest.md#appdata) permissions in `OlaresManifest.yaml`.

### AppCache

The `AppCache` storage path is allocated for applications that directly operate the disk with good performance. The disadvantage is that it cannot be accessed across nodes. For example, the system database, application log, and cache.

Applications can apply for [AppCache](../../developer/develop/package/manifest.md#appcache) permissions in `OlaresManifest.yaml`.

## [PostgreSQL](../../developer/develop/advanced/database.md#rds)

As one of the most popular open-source relational databases, PostgreSQL has excellent performance and rich plug-in functions. Olares OS deploys PostgreSQL on the system along with the popular Citus distributed database plug-in. At the same time, its cluster is managed through the PG Operator in the TAPR component. Users can easily expand the number of PostgreSQL nodes, and back up or restore data along with the entire Olares system.

If the PostgreSQL database application declared by the developer in the application is Distributed, then Olares will build its database on Citus, allowing the application to fully utilize the capabilities of the distributed PG database.

## [MongoDB](../../developer/develop/advanced/database.md#nosql)

MongoDB, as a representative of NoSQL, has a wide range of application scenarios in the Internet of Things field. By deploying [Percona Operator for MongoDB](https://github.com/percona/percona-server-mongodb-operator), developers have a cloud-native version of MongoDB cluster in Olares.

Like PostgreSQL, Olares also manages MongoDB backup and restore in a unified manner. Users do not need to have any DBA technical capabilities to easily implement functions such as scheduled backup, incremental backup, and fixed-point restore.

## [Redis](../../developer/develop/advanced/database.md#cache)

There is no doubt that Redis can be regarded as the most popular memory cache software currently. It has rich instructions and derives a variety of data types based on Key-Value data. Many systems even use it as KV data storage. Olares OS also deploys a customized [Redis Cluster Operator](https://github.com/beclab/redis-cluster-operator) in the system, providing a cloud-native version of Redis Cluster.

Olares also takes over the backup and restore of Redis Cluster. There is no need for users to provide any separate operation and maintenance operations for Redis Cluster.

In addition, since Redis Cluster itself lacks a data isolation mechanism, Olares OS has also developed a proxy layer tool to implement the `namespace` mechanism of data. This isolation mechanism is completely transparent to developers. Developers do not need to do any special processing of data keys in their code. Data isolation between multiple applications and multiple users can be achieved with simple configuration in application chart.

:::tip
The system uses the Redis Cluster version, which is different from the stand-alone version of Redis. It is recommended to read the official Redis documentation for reference.
:::

## Backup

Backup is the backup and restore module of Olares OS.

It helps users backup the entire Olares to Olares Space, and also supports user-defined storage locations.

Backup operations can be performed daily and weekly. The first backup of each backup plan is a full backup and serves as the first snapshot of the backup plan. Subsequent snapshots are incremental backups.

Backup objects include:

- Kubernetes configuration data, such as user information, application information, etc.
- Database data, such as Redis, MongoDB, PostgreSQL, etc.
- File system data, such as videos, pictures, and various documents uploaded by users through the Files application

The Backup component also has data restoration capabilities. You can download a backup snapshot to a local server or Olares Space to restore a complete Olares by rebuilding Kubernetes, databases, and user personal information.

## Learn more

- User

  [Manage files](../olares/files/index.md)<br>
  [Back up and restore](../../space/backup-restore.md) 

- Developer

  [File upload](/developer/develop/advanced/file-upload.md)<br>