---
outline: [2, 3]
description: Learn how to integrate your app with the built-in PostgreSQL service in Olares.
---
# Integrate with PostgreSQL

Use Olares PostgreSQL middleware by declaring it in `OlaresManifest.yaml`, then wiring the injected values into your container environment variables.

:::info PosgreSQL installed
PostgreSQL service has been installed by default.
:::

## Configure `OlaresManefest.yaml`

In `OlaresManifest.yaml`, add the required middleware configuration.

- Use the `scripts` field to specify scripts that should be executed after the database is created. 
- Use the `extensions` field to add the corresponding extension in the database.

:::info Variable injection in scripts
The OS provides two variables, `$databasename` and `$dbusername`, which will be replaced by Olares Application Runtime when the command is executed.
:::

:::info Example
```yaml
middleware:
  postgres:
    username: immich
    databases:
    - name: immich
      extensions:
      - vectors
      - earthdistance
      scripts:
      - BEGIN;                                           
      - ALTER DATABASE $databasename SET search_path TO "$user", public, vectors;
      - ALTER SCHEMA vectors OWNER TO $dbusername;
      - COMMIT;
```
:::

## Inject environment variables

In your deployment YAML, map the injected `.Values.postgres.*` fields to the environment variables your app uses.

:::info Example
```yaml
containers:
  - name: my-app
    env:
      # The database name you configured in OlaresManifest, specified in middleware.postgres.databases[i].name
      # NOTE: Replace <dbname> with the actual name defined in the Manifest (e.g., immich)
      - name: DB_POSTGRESDB_DATABASE
        value: {{ .Values.postgres.databases.<dbname> }}
      
      # Host
      - name: DB_POSTGRESDB_HOST
        value: {{ .Values.postgres.host }}
      
      # Port
      - name: DB_POSTGRESDB_PORT
        value: "{{ .Values.postgres.port }}"
      
      # Username
      - name: DB_POSTGRESDB_USER
        value: {{ .Values.postgres.username }}
      
      # Password
      - name: DB_POSTGRESDB_PASSWORD
        value: {{ .Values.postgres.password }}
```
:::

## PostgreSQL Values reference

PostgreSQL Values are predefined environment variables injected into `values.yaml` during deployment. They are system-managed and not user-editable.
| Key  | Type  | Description  |
|--|--|--|
| `.Values.postgres.host` | String  | PostgreSQL database host |
| `.Values.postgres.port` | Number | PostgreSQL database port |
| `.Values.postgres.username`  | String | PostgreSQL database username |
| `.Values.postgres.password`  | String | PostgreSQL database password |
| `.Values.postgres.databases` | Map<String,String> | The requested database name is used as the key. <br>For example, if you request `app_db`, the value is available at `.Values.postgres.databases.app_db`|