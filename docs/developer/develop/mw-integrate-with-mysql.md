---
outline: [2, 3]
description: Learn how to integrate your app with MySQL service in Olares.
---
# Integrate with MySQL

Use Olares MySQL middleware by declaring it in `OlaresManifest.yaml`, then mapping the injected values to your container environment variables.

## Install MySQL service

Install the MySQL service from Market.

1. Open Market from Launchpad and search for "MySQL".
2. Click **Get**, then **Install**, and wait for the installation to complete.

Once installed, the service and its connection details will appear in the Middleware list in Control Hub.

## Configure `OlaresManifest.yaml`

In `OlaresManifest.yaml`, add the required middleware configuration.

- Use the `username` field to specify the MySQL database user.
- Use the `databases` field to request one or more databases. Each database name is used as the key in `.Values.mysql.databases`.

**Example**
```yaml
middleware:
  mysql:
    username: mysqlclient
    databases:
      - name: aaa
```

## Map to environment variables

In your deployment YAML, map the injected `.Values.mysql.*` fields to the container environment variables your app requires.

**Example**
```yaml
containers:
  - name: my-app
    # For MySQL, the corresponding values are as follows
    env:
      - name: MDB_HOST
        value: "{{ .Values.mysql.host }}"

      - name: MDB_PORT
        value: "{{ .Values.mysql.port }}"

      - name: MDB_USER
        value: "{{ .Values.mysql.username }}"

      - name: MDB_PASSWORD
        value: "{{ .Values.mysql.password }}"

      # Database name
      # The database name configured in OlaresManifest (e.g., aaa)
      - name: MDB_DB
        value: "{{ .Values.mysql.databases.aaa }}"
```

## MySQL values reference

MySQL values are predefined runtime values injected into `values.yaml` during deployment. They are system-managed and not user-editable.

| Value | Type | Description |
| --- | --- | --- |
| `.Values.mysql.host` | String | MySQL host. |
| `.Values.mysql.port` | Number | MySQL port. |
| `.Values.mysql.username` | String | MySQL username. |
| `.Values.mysql.password` | String | MySQL password. |
| `.Values.mysql.databases` | Map\<String,String> | Requested databases, keyed by database name. For example, a request for `app_db` is available at `.Values.mysql.databases.app_db`. |