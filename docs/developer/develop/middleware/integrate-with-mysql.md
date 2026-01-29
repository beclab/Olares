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

:::info Example
```yaml
middleware:
  mysql:
    username: mysqlclient
    databases:
      - name: aaa
```
:::

## Inject environment variables

In your deployment YAML, map the injected `.Values.mysql.*` fields to the environment variables your app uses.

:::info Example
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
      # The database name you configured in OlaresManifest (e.g., 'aaa')
      - name: MDB_DB
        value: "{{ .Values.mysql.databases.aaa }}"
```
:::

## MySQL Values reference

MySQL Values are predefined environment variables injected into `values.yaml` during deployment. They are system-managed and not user-editable.

| Key  | Type  | Description  |
|--|--|--|
| `.Values.mysql.host` | String | MySQL database host |
| `.Values.mysql.port` | Number | MySQL database port |
| `.Values.mysql.username` | String | MySQL database username |
| `.Values.mysql.password` | String | MySQL database password |
| `.Values.mysql.databases` | Map<String,String> | The requested database name is used as the key. <br/>For example, if you request `aaa`, the value is available at `.Values.mysql.databases.aaa`. |