---
outline: [2, 3]
description: Olares 数据管理架构说明，阐述文件系统类型、应用存储路径和数据库支持体系。包括 JuiceFS、PostgreSQL、MongoDB 和 Redis 的技术特性。
---

# 数据

数据通常存储在文件系统和数据库中，其中数据库又是建立在文件系统之上的。以下是 Olares 在这两方面的设计理念：

**文件系统方面**：

  Olares 设计用于多节点集群环境。因此在开发应用时，需要考虑程序被调度到不同节点时对文件系统的访问问题。我们致力于对开发者屏蔽这些细节。

**数据库方面**：

- 对于常用数据库，开发者只需修改配置即可完成集成。
- 不同用户和应用可以共享物理数据库实例，以节省资源开销。

**共同特点**：

- 不同用户、不同应用之间的数据相互隔离。
- 可扩展且高可用。
- 能在系统层面进行统一的备份和恢复。

## 文件系统类型

### JuiceFS

Olares 采用 [JuiceFS](https://juicefs.com) 作为底层的多物理节点共享文件系统方案。这样应用可以通过最简单的 HostPath PV 方式获得跨节点的文件访问能力，使 Pod 能够在集群中自由调度。

针对 JuiceFS 的后端对象存储方案，我们提供了 S3 和 MinIO 两种选择。

默认情况下，Olares 在本地安装时使用本地文件系统（FS）。不过，如果在运行 [`olares-cli prepare`](../../developer/install/cli/prepare.md) 命令时指定了 `--with-juicefs=true` 选项，系统就会安装并使用 JuiceFS，同时会搭建一个 MinIO 实例作为后端存储。

### 本地磁盘

某些应用系统中可能会出现密集的文件系统读写操作，这些密集的文件系统读写往往是碎片化的随机读写。在现有的各种分布式存储集群方案中，对于这种密集的碎片化随机读写操作，很容易造成 I/O 或 CPU 消耗过高（通常表现为较高的 I/O Wait）。

Olares 提供的最佳实践是充分利用节点的本地硬盘作为文件缓冲区。虽然节点的本地硬盘容量有限，但由于基本采用 SSD 硬盘，具有较高的读写性能。应用读写文件时会先在节点本地硬盘上进行缓冲，然后批量异步写入分布式文件系统。这样可以将大部分碎片化的随机读写转化为少量的顺序读写，大幅提升系统 I/O 效率。

## 应用存储路径

对于应用来说，有三种不同的存储路径用于处理不同的使用场景。

### UserData

`UserData` 存储路径用于存放变动不频繁但需要跨应用访问的文件，如文档、照片和视频等。

应用可以通过在 `OlaresManifest.yaml` 中申请 [UserData](../../developer/develop/package/manifest.md#userdata) 权限来获取 `Home` 目录下某个目录的访问权限。比如 PhotoPrism 可以申请 `Picture` 目录的权限，qBittorrent 和 Jellyfin 可以申请 `Downloads` 目录的权限。

### AppData

`AppData` 存储路径用于存放变动不频繁但需要跨节点的数据，比如配置文件。

应用可以在 `OlaresManifest.yaml` 中申请 [AppData](../../developer/develop/package/manifest.md#appdata) 权限。

### AppCache

`AppCache` 存储路径分配给需要直接操作磁盘且性能要求较好的应用。比如系统数据库、应用日志和缓存等。缺点是无法跨节点访问。

应用可以在 `OlaresManifest.yaml` 中申请 [AppCache](../../developer/develop/package/manifest.md#appcache) 权限。

## [PostgreSQL](../../developer/develop/advanced/database.md#rds)

作为最受欢迎的开源关系型数据库之一，PostgreSQL 具有出色的性能和丰富的插件功能。Olares 在系统中部署了 PostgreSQL，同时集成了广受欢迎的 Citus 分布式数据库插件。通过 Olares 应用运行时组件中的 PG Operator 进行集群管理，用户可以轻松扩展 PostgreSQL 节点数量，并随整个 Olares 系统进行备份或恢复。

如果开发者在应用中声明的 PostgreSQL 数据库为分布式类型，那么 Olares 会在 Citus 上构建其数据库，让应用充分利用分布式 PG 数据库的能力。

## [MongoDB](../../developer/develop/advanced/database.md#nosql)

MongoDB 作为 NoSQL 的代表，在物联网领域有着广泛的应用场景。通过部署 [Percona Operator for MongoDB](https://github.com/percona/percona-server-mongodb-operator)，开发者在 Olares 中就拥有了云原生版本的 MongoDB 集群。

与 PostgreSQL 一样，Olares 也统一管理 MongoDB 的备份和恢复。用户无需具备任何 DBA 技术能力，就能轻松实现定时备份、增量备份、定点恢复等功能。

## [Redis](../../developer/develop/advanced/database.md#cache)

毫无疑问，Redis 可以说是目前最受欢迎的内存缓存软件。它拥有丰富的指令，并基于 Key-Value 数据衍生出多种数据类型。很多系统甚至将其作为 KV 数据存储使用。Olares 也在系统中部署了定制的 [Redis Cluster Operator](https://github.com/beclab/redis-cluster-operator)，提供云原生版本的 Redis 集群。

Olares 同样接管了 Redis 集群的备份和恢复工作，用户无需为 Redis 集群提供任何单独的运维操作。

此外，由于 Redis 集群本身缺乏数据隔离机制，Olares 还开发了代理层工具来实现数据的 `namespace` 机制。这种隔离机制对开发者来说是完全透明的，开发者无需在代码中对数据键做任何特殊处理，只需在应用 chart 中简单配置即可实现多应用、多用户之间的数据隔离。

:::tip 提示
系统使用的是 Redis 集群版本，与单机版 Redis 有所不同，建议参考 Redis 官方文档。
:::

## 备份

备份是 Olares 的备份和恢复模块。

它帮助用户将整个 Olares 备份到 Olares Space，同时也支持用户自定义存储位置。

备份操作可以按日或按周进行。每个备份计划的第一次备份是全量备份，作为该备份计划的第一个快照。后续快照均为增量备份。

备份对象包括：

- Kubernetes 配置数据，如用户信息、应用信息等
- 数据库数据，如 Redis、MongoDB、PostgreSQL 等
- 文件系统数据，如用户通过文件管理器上传的视频、图片、各类文档等

备份组件还具备数据恢复能力。可以将备份快照下载到本地服务器或 Olares Space，通过重建 Kubernetes、数据库和用户个人信息，恢复出一个完整的 Olares。

## 了解更多

- 用户

  [文件管理](../olares/files/index.md)<br>
  [备份与恢复](../../space/backup-restore.md)

- 开发者

  [文件上传](../../developer/develop/advanced/file-upload.md)<br>