---
outline: [2, 3]
description: Learn how to integrate your app with the built-in Redis service in Olares.
---
# Integrate with Redis

Use Olares Redis middleware by declaring it in `OlaresManifest.yaml`, then mapping the injected values to your container environment variables.

:::info Redis installed
Redis service has been installed by default.
:::

## Configure `OlaresManifest.yaml`

In `OlaresManifest.yaml`, add the required Redis middleware configuration.

- Use the `password` field to specify the Redis access password.
- Use the `namespace` field to request a Redis namespace.

**Example**
```yaml
middleware:
  redis:
    password: password
    namespace: db0
```

## Map to environment variables

In your deployment YAML, map the injected `.Values.redis.*` fields to the container environment variables your app requires.

**Example**
```yaml
containers:
  - name: my-app
    env:
      # Host
      - name: REDIS_HOST
        value: {{ .Values.redis.host }}

      # Port
      # Quote the value to ensure it's treated as a string
      - name: REDIS_PORT
        value: "{{ .Values.redis.port }}"

      # Password
      # Quote the value to handle special characters correctly
      - name: REDIS_PASSWORD
        value: "{{ .Values.redis.password }}"

      # Namespace
      # NOTE: Replace <namespace> with the actual namespace defined in OlaresManifest (e.g., db0)
      - name: REDIS_NAMESPACE
        value: {{ .Values.redis.namespaces.<namespace> }}
```

## Redis values reference

Redis values are predefined runtime values injected into `values.yaml` during deployment. They are system-managed and not user-editable.

| Value | Type | Description |
| --- | --- | --- |
| `.Values.redis.host` | String | Redis host. |
| `.Values.redis.port` | Number | Redis port. |
| `.Values.redis.password` | String | Redis password. |
| `.Values.redis.namespaces` | Map\<String,String> | Requested namespaces, keyed by namespace name. For example, a request for `app_ns` is available at `.Values.redis.namespaces.app_ns`. |