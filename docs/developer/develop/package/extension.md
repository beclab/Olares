# Extensions field to Helm in Olares

During installation, Olares injects extended field into the APP, using the configuration from [OlaresManifest.yaml](manifest.md).

The information from these extended fields can be directly referenced in the template without being defined in values.yaml. If there are definitions in values.yaml with the same name, the system value will overwrite them.


- User Information

  | Value             | Type   | Description         |
  | -------------------- | ------ | --------------------- |
  | .Values.bfl.username | String | Username of the currently installing APP |
  | .Values.user.zone    | String | Domain of the current user          |

- Domain Information

  | Value       | Type               | Description                               |
  | -------------- | ------------------ | ---------------------------------------------------- |
  | .Values.domain | Map<String,String> | Define the Entry for the App, using the entrance name as the key and the URL as the value.  |

- Storage Information
  | Value       | Type               | Description 
  | -------------- | ------------------ | ---------------------------------------------------- |
  | .Values.userspace.appData | String | Cluster storage address available to the APP  |
  | .Values.userspace.appCache | String | Local node cache address available to the APP  |
  | .Values.userspace.userData | String | storage directory of User's data |

- Cluster Information
  | Value       | Type               | Description 
  | -------------- | ------------------ | ---------------------------------------------------- |
  | .Values.cluster.arch | String | Cluster CPU architecture   |

  Multi-platform (AMD64 and ARM) cluster is not supported for now.

- Application Dependencies
  | Value       | Type               | Description 
  | -------------- | ------------------ | ---------------------------------------------------- |
  | .Values.deps | Map<String, Value> | Current address and port of the applications the APP depends on  |
  | .Values.svcs | Map<String, Value> | Other services and ports of the applications the APP depends on  |

  When an application sets dependencies on another application, this will be passed through the `deps` parameter. For instance, if an application, sets a dependency on another application called **'A-Server'**, and **'A-Server'** sets the entry name as **'aserver'** with the entry host configured as **'aserver-svc'**, the value will be like:
  ```
  {
    "aserver_host": "aserver-svc.<A-Server namespce>",
    "aserver_port": 80
  }
  ```
  At the same time, `svcs` will pass in all services of A-Server.
  ```
  {
    "aserver-svc_host": "aserver-svc.<A-Server namespce>",
    "aserver-svc_port": [80]    # If there are multiple ports in the service, they will be passed in together.
  }
  ```

- Database Information

  | Value                       | Type               | Description                                                                                                       |
  | -------------------------- | ------------------ | ---------------------------------------------------------------------------------------------------------- |
  | .Values.postgres.host      | String             | PostgreSQL database host address                                                                                      |
  | .Values.postgres.port      | Number             | PostgreSQL database port                                                                                      |
  | .Values.postgres.username  | String             | PostgreSQL database username                                                                                    |
  | .Values.postgres.password  | String             | PostgreSQL database password                                                                                      |
  | .Values.postgres.databases | Map<String,String> | PostgreSQL database name. Use the configured database name as the key. For instance, if it's configured as 'app_db', the variable would be `.Values.postgres.databases.app_db`. |
  | .Values.mongo.host         | String             | MongoDB database host address                                                                                            |
  | .Values.mongo.port         | Number             | MongoDB database port                                                                                         |
  | .Values.mongo.username     | String             | MongoDB database username                                                                                       |
  | .Values.mongo.password     | String             | MongoDB database username                                                                                         |
  | .Values.mongo.databases    | Map<String,String> | MongoDB database name. Use the configured database name as the key. For instance, if it's configured as 'app_db', the variable would be `.Values.mongo.databases.app_db `      |
  | .Values.redis.host         | String             | Redis database host address                                                                                              |
  | .Values.redis.port         | Number             | Redis database port                                                                                           |
  | .Values.redis.password     | String             | Redis database username                                                                                           |
  | .Values.redis.namespaces   | Map<String,String> | Redis namespace. Use the configured namespace as the key. For instance, if it's configured as 'app_ns', the variable would be `.Values.redis.namespaces.app_ns`|