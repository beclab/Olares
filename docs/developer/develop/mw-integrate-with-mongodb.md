---
outline: [2, 3]
description: Learn how to integrate your app with MongoDB service in Olares.
---
# Integrate with MongoDB

Use Olares MongoDB middleware by declaring it in `OlaresManifest.yaml`, then mapping the injected values to your container environment variables.

## Install MongoDB service

Install the MongoDB service from Market.

1. Open Market from Launchpad and search for "MongoDB".
2. Click **Get**, then **Install**, and wait for the installation to complete.

Once installed, the service and its connection details will appear in the Middleware list in Control Hub.

## Configure `OlaresManifest.yaml`

In `OlaresManifest.yaml`, add the required middleware configuration.

- Use the `username` field to specify the MongoDB database user.
- Use the `databases` field to request one or more databases.
- (Optional) Use the `script` field under each database to specify initialization scripts that are executed after the database is created.

**Example**
```yaml
middleware:
  mongodb:
    username: chromium
    databases:
    - name: chromium
      script:
      - 'db.getSiblingDB("$databasename").myCollection.insertOne({ x: 111 });'
      # Please make sure each line is a complete query.
```

## Inject environment variables

In your deployment YAML, map the injected `.Values.mongodb.*` fields to the environment variables your app uses.

**Example**
```yaml
containers:
  - name: my-app
    # For MongoDB, the corresponding values are as follows
    env:
      - name: MONGODB_HOST
        value: "{{ .Values.mongodb.host }}"

      - name: MONGODB_PORT
        value: "{{ .Values.mongodb.port }}"

      - name: MONGODB_USER
        value: "{{ .Values.mongodb.username }}"

      - name: MONGODB_PASSWORD
        value: "{{ .Values.mongodb.password }}"

      # Database name
      # The database name configured in OlaresManifest (e.g., app_db)
      - name: MONGODB_DATABASE
        value: "{{ .Values.mongodb.databases.app_db }}"
```

## MongoDB Values reference

MongoDB Values are predefined environment variables injected into `values.yaml` during deployment. They are system-managed and not user-editable.

| Key  | Type  | Description  |
|--|--|--|
| `.Values.mongodb.host` | String  | MongoDB database host |
| `.Values.mongodb.port` | Number  | MongoDB database port |
| `.Values.mongodb.username` | String | MongoDB database username |
| `.Values.mongodb.password`  | String | MongoDB database password |
| `.Values.mongodb.databases` | Map<String,String> | The requested database name is used as the key. <br/>For example, if you request `app_db`, the value is available at `.Values.mongodb.databases.app_db`. |