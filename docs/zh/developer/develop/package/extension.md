# Olares 对 Helm 进行的字段扩展

在安装过程中，Olares 使用 [OlaresManifest.yaml](manifest.md) 中的配置将扩展字段注入到应用程序中。

这些扩展字段信息，可以在模版中直接引用，而无需在 values.yaml 中事先定义 （如果 values.yaml 中有相同的定义，会被系统覆盖）。


- 用户信息

  | 变量                   | 类型     | 说明         |
  |----------------------|--------|------------|
  | .Values.bfl.username | String | 当前安装应用的用户名 |
  | .Values.user.zone    | String | 当前用户的域名    |

- 地址信息

  | 变量             | 类型                 | 说明                                   |
  |----------------|--------------------|--------------------------------------|
  | .Values.domain | Map<String,String> | 应用定义的入口地址 URL, 每个 Entry 为：入口名 => URL |

- 存储信息
  | 变量 | 类型 | 说明 |
  | -------------- | ------------------ | ---------------------------------------------------- |
  | .Values.userspace.appData | String | 应用可用的集群存储地址 |
  | .Values.userspace.appCache | String | 应用可用的本地节点缓存地址 |
  | .Values.userspace.userData | String | 用户的数据存储目录 |

- 集群信息
  | 变量 | 类型 | 说明 |
  | -------------- | ------------------ | ---------------------------------------------------- |
  | .Values.cluster.arch | String | 集群 CPU 架构   |

  目前不支持多平台（AMD64 和 ARM）集群。

- 应用依赖
  | 变量 | 类型 | 说明 |
  | -------------- | ------------------ | ---------------------------------------------------- |
  | .Values.deps | Map<String, Value> | 应用所依赖的应用当前的地址和端口 |
  | .Values.svcs | Map<String, Value> | 应用所依赖的应用其他 service 和端口 |

 当应用程序设置对另一个应用程序的依赖关系时，这将通过 `deps` 参数传递。例如，如果一个应用程序设置了对另一个名为“A-Server”的应用程序的依赖关系，并且 A-Server 将条目名称设置为条目主机的 aserver 配置为 `'aserver-svc'`，值将类似于：
```
  {
    "aserver_host": "aserver-svc.<A-Server namespce>",
    "aserver_port": 80
  }
  ```
  同时，`svcs`会传入 A-Server 的所有 service
  ```
  {
    "aserver-svc_host": "aserver-svc.<A-Server namespce>",
    "aserver-svc_port": [80]    # 如果 service 有多个端口会一起传入
  }
  ```

- 数据库信息

  | 变量                         | 类型                 | 说明                                                                                 |
  |----------------------------|--------------------|------------------------------------------------------------------------------------|
  | .Values.postgres.host      | String             | PostgreSQL 数据库地址                                                                   |
  | .Values.postgres.port      | Number             | PostgreSQL 数据库端口                                                                   |
  | .Values.postgres.username  | String             | PostgreSQL 数据库用户名                                                                  |
  | .Values.postgres.password  | String             | PostgreSQL 数据库密码                                                                   |
  | .Values.postgres.databases | Map<String,String> | PostgreSQL 数据库名称，以申请数据库名作为 key，例如， 申请 app_db, 变量为.Values.postgres.databases.app_db |
  | .Values.mongo.host         | String             | MongoDB 数据库地址                                                                      |
  | .Values.mongo.port         | Number             | MongoDB 数据库端口                                                                      |
  | .Values.mongo.username     | String             | MongoDB 数据库用户名                                                                     |
  | .Values.mongo.password     | String             | MongoDB 数据库密码                                                                      |
  | .Values.mongo.databases    | Map<String,String> | MongoDB 数据库名称，以申请数据库名作为 key，例如， 申请 app_db, 变量为.Values.mongo.databases.app_db       |
  | .Values.redis.host         | String             | Redis 数据库地址                                                                        |
  | .Values.redis.port         | Number             | Redis 数据库端口                                                                        |
  | .Values.redis.password     | String             | Redis 数据库密码                                                                        |
  | .Values.redis.namespaces   | Map<String,String> | Redis 命名空间名称，以申请命名空间作为 key，例如， 申请 app_ns, 变量为.Values.redis.namespaces.app_ns       |