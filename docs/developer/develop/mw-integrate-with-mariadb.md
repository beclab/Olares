---
outline: [2, 3]
description: Learn how to integrate your app with MariaDB service in Olares.
---
# Integrate with MariaDB

Use Olares MariaDB middleware by declaring it in `OlaresManifest.yaml`, then mapping the injected values to your container environment variables.

## Install MariaDB service

Install the MariaDB service from Market.

1. Open Market from Launchpad and search for "MariaDB".
2. Click **Get**, then **Install**, and wait for the installation to complete.

Once installed, the service and its connection details will appear in the Middleware list in Control Hub.

## Configure `OlaresManifest.yaml`

In `OlaresManifest.yaml`, add the required middleware configuration.

- Use the `username` field to specify the database username.
- Use the `databases` field to request one or more databases. Each database name is used as the key in `.Values.mariadb.databases`.

**Example**
```yaml
middleware:
  mariadb:
    username: mariadbclient
    databases:
      - name: aaa
```

## Map to environment variables
In your deployment YAML, map the injected `.Values.mariadb.*` fields to the container environment variables your app requires.

**Example**
```yaml
containers:
  - name: my-app
    # For MariaDB, the corresponding value is as follows
    env:
      - name: MDB_HOST
        value: "{{ .Values.mariadb.host }}"
      
      - name: MDB_PORT
        value: "{{ .Values.mariadb.port }}"
      
      - name: MDB_USER
        value: "{{ .Values.mariadb.username }}"
      
      - name: MDB_PASSWORD
        value: "{{ .Values.mariadb.password }}"
      
      # Database Name
      # The database name configured in OlaresManifest (e.g., aaa)
      - name: MDB_DB
        value: "{{ .Values.mariadb.databases.aaa }}"
```

## MariaDB values reference
MariaDB values are predefined runtime values injected into `values.yaml` during deployment. They are system-managed and not user-editable.

| Value | Type | Description |
| --- | --- | --- |
| `.Values.mariadb.host` | String | MariaDB host. |
| `.Values.mariadb.port` | Number | MariaDB port. |
| `.Values.mariadb.username` | String | MariaDB username. |
| `.Values.mariadb.password` | String | MariaDB password. |
| `.Values.mariadb.databases` | Map\<String,String> | Requested databases, keyed by database name. For example, a request for `app_db` is available at `.Values.mariadb.databases.app_db`. |