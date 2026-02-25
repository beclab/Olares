---
outline: [2, 3]
description: 了解如何在 Olares 中安装 RabbitMQ，并通过 RabbitMQ Dashboard 管理 RabbitMQ 资源。
---
# 查看 RabbitMQ 数据

本文介绍如何在 Olares 中安装 RabbitMQ 服务，并通过 RabbitMQ 仪表盘管理数据。

## 安装 RabbitMQ 服务

在使用 RabbitMQ 之前，需从应用市场安装 RabbitMQ 服务。

1. 从启动台打开应用市场并搜索“RabbitMQ”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

安装完成后，RabbitMQ 服务及其连接详情将显示在控制面板的中间件列表中。

## 安装 RabbitMQ 仪表盘

RabbitMQ 仪表盘依赖于 RabbitMQ 服务，只有在 RabbitMQ 服务就绪后才能安装。

1. 在应用市场中搜索“RabbitMQ 仪表盘”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

## 获取连接信息

在建立连接前，需从控制面板获取 RabbitMQ 连接详情。

1. 从启动台打开控制面板。
2. 在左侧导航栏中找到中间件，并选择 **Rabbitmq**。
3. 记录信息面板中的以下信息：
    - **用户**：用于连接 RabbitMQ 仪表盘。
    - **密码**：用于连接 RabbitMQ 仪表盘。

    ![RabbitMQ details](/images/zh/manual/developer/mw-rabbitmq-details.png#bordered){width=60% style="margin-left:0"}

## RabbitMQ 仪表盘可视化管理

RabbitMQ 仪表盘提供图形化界面，用于查看和管理 RabbitMQ 资源。

1. 从启动台打开 RabbitMQ 仪表盘。
2. 在登录界面，输入从控制面板获取的**用户**和**密码**。

登录成功后，即可进入管理界面，查看并管理 RabbitMQ 资源。