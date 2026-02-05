---
outline: [2, 3]
description: Learn how to integrate your app with RabbitMQ service in Olares.
---
# Integrate with RabbitMQ

Use Olares RabbitMQ middleware by declaring it in `OlaresManifest.yaml`, then mapping the injected values to your container environment variables.

## Install RabbitMQ service

Install the RabbitMQ service from Market.

1. Open Market from Launchpad and search for "RabbitMQ".
2. Click **Get**, then **Install**, and wait for the installation to complete.

Once installed, the service and its connection details will appear in the Middleware list in Control Hub.

## Configure `OlaresManifest.yaml`

In `OlaresManifest.yaml`, add the required middleware configuration.

- Use the `username` field to specify the RabbitMQ user.
- Use the `vhosts` field to request one or more virtual hosts (vhosts). Each vhost name is used as the key in `.Values.rabbitmq.vhosts`.

**Example**
```yaml
middleware:
  rabbitmq:
    username: rabbitmquser
    vhosts:
      - name: aaa
```

## Inject environment variables

In your deployment YAML, map the injected `.Values.rabbitmq.*` fields to the environment variables your app uses.

**Example**
```yaml
containers:
  - name: my-app
    # For RabbitMQ, the corresponding values are as follows
    env:
      - name: RABBITMQ_HOST
        value: "{{ .Values.rabbitmq.host }}"

      - name: RABBITMQ_PORT
        value: "{{ .Values.rabbitmq.port }}"

      - name: RABBITMQ_USER
        value: "{{ .Values.rabbitmq.username }}"

      - name: RABBITMQ_PASSWORD
        value: "{{ .Values.rabbitmq.password }}"

      # Vhost
      # The vhost name configured in OlaresManifest (e.g., aaa)
      - name: RABBITMQ_VHOST
        value: "{{ .Values.rabbitmq.vhosts.aaa }}"
```

## Construct a RabbitMQ connection URI

After configuring the environment variables, you can read them in your application code to construct the connection string.

Below is an example of constructing an AMQP URL using the environment variables:

```Go
// Read environment variables
user := os.Getenv("RABBITMQ_USER")
password := os.Getenv("RABBITMQ_PASSWORD")
vhost := os.Getenv("RABBITMQ_VHOST")
host := os.Getenv("RABBITMQ_HOST")
portMQ := os.Getenv("RABBITMQ_PORT")

// Construct AMQP connection string
// Format: amqp://user:password@host:port/vhost
url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, portMQ, vhost)
```

## RabbitMQ Values reference

RabbitMQ Values are predefined environment variables injected into `values.yaml` during deployment. They are system-managed and not user-editable.

| Key  | Type  | Description  |
|--|--|--|
| `.Values.rabbitmq.host` | String | RabbitMQ service host |
| `.Values.rabbitmq.port` | Number | RabbitMQ service port |
| `.Values.rabbitmq.username` | String | RabbitMQ username |
| `.Values.rabbitmq.password` | String | RabbitMQ password |
| `.Values.rabbitmq.vhosts` | Map<String,String> | The requested vhost name is used as the key. <br/>For example, if you request `aaa`, the value is available at `.Values.rabbitmq.vhosts.aaa`. |