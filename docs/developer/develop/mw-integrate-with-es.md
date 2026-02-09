---
outline: [2, 3]
description: Learn how to integrate your app with Elasticsearch service in Olares.
---
# Integrate with Elasticsearch

Use Olares Elasticsearch middleware by declaring it in `OlaresManifest.yaml`, then mapping the injected values to your container environment variables.

## Install Elasticsearch service

Install the Elasticsearch service from Market.

1. Open Market from Launchpad and search for "Elasticsearch".
2. Click **Get**, then **Install**, and wait for the installation to complete.

Once installed, the service and its connection details will appear in the Middleware list in Control Hub.

## Configure `OlaresManifest.yaml`

In `OlaresManifest.yaml`, add the required middleware configuration.

- Use the `username` field to specify the Elasticsearch user.
- Use the `indexes` field to request one or more indexes. Each index name is used as the key in `.Values.elasticsearch.indexes`.

**Example**

```yaml
middleware:
  elasticsearch:
    username: elasticlient
    indexes:
      - name: aaa
```

## Inject environment variables

In your deployment YAML, map the injected `.Values.elasticsearch.*` fields to the environment variables your app uses.

**Example**
```yaml
containers:
  - name: my-app
    env:
      - name: ES_HOST
        value: "{{ .Values.elasticsearch.host }}"

      - name: ES_PORT
        value: "{{ .Values.elasticsearch.port }}"

      - name: ES_USER
        value: "{{ .Values.elasticsearch.username }}"

      - name: ES_PASSWORD
        value: "{{ .Values.elasticsearch.password }}"

      # Index name
      # The index name configured in OlaresManifest (for example, aaa)
      - name: ES_INDEX
        value: "{{ .Values.elasticsearch.indexes.aaa }}"
```

## Elasticsearch Values reference

Elasticsearch Values are predefined environment variables injected into `values.yaml` during deployment. They are system-managed and not user-editable.

| Key  | Type  | Description  |
|--|--|--|
|`.Values.elasticsearch.host`| String | Elasticsearch service host |
|`.Values.elasticsearch.port`| Number | Elasticsearch service port |
|`.Values.elasticsearch.username`| String | Elasticsearch username |
|`.Values.elasticsearch.password`| String | Elasticsearch password |
|`.Values.elasticsearch.indexes` | Map<String,String> | The requested index name is used<br> as the key. For example, if you request `aaa`, the value is available at `.Values.elasticsearch.indexes.aaa`. |