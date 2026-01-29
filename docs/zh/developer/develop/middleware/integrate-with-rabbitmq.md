---
outline: [2, 3]
description: 了解如何在 Olares 中将你的应用接入 RabbitMQ 服务。
---
# 集成 RabbitMQ

通过在 `OlaresManifest.yaml` 中声明 RabbitMQ 中间件，并将系统注入的配置值映射到容器的环境变量中，即可在应用中使用 Olares 提供的 RabbitMQ 服务。

## 安装 RabbitMQ 服务

通过应用市场安装 RabbitMQ 服务。

1. 从启动台打开应用市场，搜索“RabbitMQ”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

安装完成后，RabbitMQ 服务及其连接信息将显示在控制面板的中间件列表中。

## 配置 `OlaresManifest.yaml`

在 `OlaresManifest.yaml` 中添加所需的中间件配置。

- 使用 `username` 字段指定 RabbitMQ 用户。
- 使用 `vhosts` 字段申请一个或多个虚拟主机（vhost）。每个 vhost 名称将作为键注入到 `.Values.rabbitmq.vhosts` 中。

:::info 示例
```yaml
middleware:
  rabbitmq:
    username: rabbitmquser
    vhosts:
      - name: aaa
```
:::

## 注入环境变量

在应用的部署 YAML 中，将系统注入的 `.Values.rabbitmq.*` 字段映射为应用所使用的环境变量。

:::info 示例
```yaml
containers:
  - name: my-app
    # RabbitMQ 对应的注入值如下
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
      # 在 OlaresManifest 中配置的 vhost 名称（例如：'aaa'）
      - name: RABBITMQ_VHOST
        value: "{{ .Values.rabbitmq.vhosts.aaa }}"
```
:::

## 构建 RabbitMQ 连接 URI

完成环境变量配置后，可以在应用代码中读取这些变量，并构建 RabbitMQ 的连接字符串。

以下是使用环境变量构建 AMQP URL 的示例（Go 语言）：

```Go
// 读取环境变量
user := os.Getenv("RABBITMQ_USER")
password := os.Getenv("RABBITMQ_PASSWORD")
vhost := os.Getenv("RABBITMQ_VHOST")
host := os.Getenv("RABBITMQ_HOST")
portMQ := os.Getenv("RABBITMQ_PORT")

// 构建 AMQP 连接字符串
// 格式: amqp://user:password@host:port/vhost
url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, portMQ, vhost)
```

## RabbitMQ Values 参考

RabbitMQ Values 是在部署过程中自动注入到 `values.yaml` 中的预定义变量，由系统统一管理，用户不可手动修改。

| 键  | 类型  | 说明  |
|--|--|--|
| `.Values.rabbitmq.host` | String | RabbitMQ 服务地址 |
| `.Values.rabbitmq.port` | Number | RabbitMQ 服务端口 |
| `.Values.rabbitmq.username` | String | RabbitMQ 用户名 |
| `.Values.rabbitmq.password` | String | RabbitMQ 密码 |
| `.Values.rabbitmq.vhosts` | Map<String,String> | 以申请的 vhost 名作为键。<br/>例如申请 `aaa`，可通过 `.Values.rabbitmq.vhosts.aaa` 获取对应的值。 |