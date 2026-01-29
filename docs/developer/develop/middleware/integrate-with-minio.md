---
outline: [2, 3]
description: Learn how to integrate your app with MinIO service in Olares.
---
# Integrate with MinIO

Use Olares MinIO middleware by declaring it in `OlaresManifest.yaml`, then mapping the injected values to your container environment variables.

## Install MinIO service

Install the MinIO service from Market.

1. Open Market from Launchpad and search for "MinIO".
2. Click **Get**, then **Install**, and wait for the installation to complete.

Once installed, the service and its connection details will appear in the Middleware list in Control Hub.

## Configure `OlaresManifest.yaml`

In `OlaresManifest.yaml`, add the required middleware configuration.

- Use the `username` field to specify the MinIO access key.
- Use the `buckets` field to request one or more buckets. Each bucket name is used as the key in `.Values.minio.buckets`.

:::info Example
```yaml
middleware:
  minio:
    username: miniouser
    buckets:
      - name: mybucket
```
:::

## Inject environment variables

In your deployment YAML, map the injected `.Values.minio.*` fields to the environment variables your app uses.



:::info Example
```yaml
containers:
  - name: my-app
    # For MinIO, the corresponding values are as follows
    env:
      # Construct the endpoint using host and port
      - name: MINIO_ENDPOINT
        value: "{{ .Values.minio.host }}:{{ .Values.minio.port }}"

      - name: MINIO_PORT
        value: "{{ .Values.minio.port }}"

      - name: MINIO_ACCESS_KEY
        value: "{{ .Values.minio.username }}"

      - name: MINIO_SECRET_KEY
        value: "{{ .Values.minio.password }}"

      # Bucket name
      # The bucket name you configured in OlaresManifest (e.g., 'mybucket')
      - name: MINIO_BUCKET
        value: "{{ .Values.minio.buckets.mybucket }}"
```
:::

## MinIO Values reference

MinIO Values are predefined environment variables injected into `values.yaml` during deployment. They are system-managed and not user-editable.

| Key  | Type  | Description  |
|--|--|--|
| `.Values.minio.host` | String | MinIO service host |
| `.Values.minio.port` | Number | MinIO service port |
| `.Values.minio.username` | String | MinIO access key |
| `.Values.minio.password` | String | MinIO secret key |
| `.Values.minio.buckets` | Map<String,String> | The requested bucket name is used as the key. <br>For example, if you request `mybucket`, the value is available at `.Values.minio.buckets.mybucket`. |