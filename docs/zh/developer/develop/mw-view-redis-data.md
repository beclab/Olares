---
outline: [2, 3]
description: 了解如何在 Olares 中查看并管理 Redis 数据。
---
# 查看 Redis 数据

Redis 服务在 Olares 中默认可用。本文将介绍如何在 Olares 中访问并管理 Redis 数据。

## 获取连接信息

在建立连接之前，需要从控制面板获取 Redis 的连接信息。

1. 从启动台打开控制面板。
2. 在左侧导航栏中找到中间件，并选择 **Redis**。
3. 记录信息面板中的以下信息：
    - **主机**：用于在 Bytebase 中建立连接。
    - **密码**：用于 CLI 和 Bytebase。

    ![连接信息](/images/zh/manual/developer/mw-redis-details.png#bordered){width=60% style="margin-left:0"}

## 通过 CLI 访问

你可以使用 Olares 提供的终端直接访问数据库容器。

1. 在控制面板左侧导航栏底部，打开 **Olares 终端**。
2. 进入 Redis 容器。容器名称是固定的。

    ```bash
    kubectl exec -it -n os-platform kvrocks-0 -- sh 
    ```
3. 连接 Redis 数据库：

    ```bash
    redis-cli -p 6666 -a <your password from control-hub>
    ```

## Bytebase 可视化管理

Bytebase 提供图形化界面，用于数据库管理和结构变更。

### 前提条件

:::info 需要 MongoDB 应用
Bytebase 使用 MongoDB 存储元数据。在安装 Bytebase 之前，需要先安装 MongoDB。
:::

1. 打开应用市场，搜索“MongoDB”。
2. 点击**获取**，然后点击**安装**，并等待服务运行。
3. MongoDB 安装完成后，在应用市场搜索“Bytebase”。
4. 点击**获取**，然后点击**安装**。

### 初始设置

首次启动 Bytebase 时，需要配置一个管理员账号。

:::tip
请妥善保存该账号信息。只有管理员账号才有权限创建新的数据库实例。
:::

1. 从启动台打开 Bytebase。
2. 按照界面提示，使用邮箱和密码完成管理员账号的设置。

### 创建 Redis 实例

1. 使用管理员账号登录 Bytebase。
2. 在左侧导航栏中选择**实例**，然后点击 <b>+ 添加实例</b>。
3. 选择 **Redis** 作为数据库类型。
4. 使用控制面板中获取的信息填写连接配置：
    - **Host 或 Socket**：填写**主机**地址，不包含端口。
    - **端口**：保持默认值，通常为`6379`。
    - **用户名**：留空。
    - **密码**：填写从控制面板获取的**密码**。
5. 点击**测试连接**验证连接是否成功，然后点击**创建**。

在 Bytebase 中创建实例并不会新建数据库。实例创建完成后，你可以使用相应的客户端工具对数据进行查看和管理。
