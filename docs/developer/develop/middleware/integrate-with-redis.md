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

:::info Example
```yaml
middleware:
  redis:
    password: password
    namespace: db0
```
:::

## Inject environment variables

In your deployment YAML, map the injected `.Values.redis.*` fields to the environment variables your app uses.

:::info Example
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
:::

## Redis Values reference

Redis Values are predefined environment variables injected into `values.yaml` during deployment. They are system-managed and not user-editable.
| Key  | Type  | Description  |
|--|--|--|
| `.Values.redis.host` | String | Redis service host |
| `.Values.redis.port` | Number  | Redis service port |
| `.Values.redis.password`| String | Redis service password |
| `.Values.redis.namespaces` | Map<String, String> | The requested namespace is used as the key. <br>For example, if you request `app_ns`, the value is available at `.Values.redis.namespaces.app_ns`. |
