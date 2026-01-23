---
outline: [2, 3]
description: 了解如何在 Olares 中通过 CLI 或 Bytebase 查看并管理 MongoDB 数据。
---
# 查看 MongoDB 数据

MongoDB 需要先在 Olares 应用市场安装后才能使用。本文将介绍安装流程，并说明如何在 Olares 中访问并管理 MongoDB 数据。

## 安装 MongoDB 服务

在建立连接之前，需要在 Olares 应用市场安装 MongoDB 服务。

1. 从启动台打开 Olares 应用市场，搜索“MongoDB”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

安装完成后，MongoDB 服务及其连接信息将显示在控制面板的中间件列表中。

## 获取连接信息

在建立连接之前，需要从控制面板获取 MongoDB 的连接信息。

1. 从启动台打开控制面板。
2. 在左侧导航栏中找到中间件，并选择 **Mongodb**。
3. 记录信息面板中的以下信息：
    - **Mongos**：控制面板提供的主机地址，用于在 Bytebase 中建立连接。
    - **用户**：用于在 Bytebase 中建立连接。
    - **密码**：用于 CLI 和 Bytebase。

    ![MongoDB details](/images/zh/manual/developer/mw-mongodb-details.png#bordered){width=60% style="margin-left:0"}

## 通过 CLI 访问

你可以使用 Olares 提供的终端直接访问数据库容器。

1. 在控制面板左侧导航栏底部，打开 **Olares 终端**。
2. 获取中间件容器名称：

    ```bash
    kubectl get pods -n os-platform | grep tapr-middleware
    ```
3. 记录以 `tapr-middleware` 开头的 Pod 名称，并进入容器：

    ```bash
    kubectl exec -it -n os-platform <tapr-middleware-pod> -- sh
    ```
4. 使用 `mongosh` 连接 MongoDB：

    ```bash
    mongosh "mongodb://root:<your password from controlhub>@mongodb-mongodb-headless.mongodb-middleware:27017"
    ```

## Bytebase 可视化管理

Bytebase 提供图形化界面，用于数据库管理和结构变更。

### 安装 Bytebase

1. 打开 Olares 应用市场，搜索“Bytebase”。
2. 点击**获取**，然后点击**安装**。

### 初始设置

首次启动 Bytebase 时，需要配置一个管理员账号。

:::tip
请妥善保存该账号信息。只有管理员账号才有权限创建新的数据库实例。
:::

1. 从启动台打开 Bytebase。
2. 按照界面提示，使用邮箱和密码完成管理员账号的设置。

### 创建 MongoDB 实例

1. 使用管理员账号登录 Bytebase。
2. 在左侧导航栏中选择**实例**，然后点击 <b>+ 添加实例</b>。
3. 选择 **MongoDB** 作为数据库类型。
4. 使用控制面板中获取的信息填写连接配置：
    - **Host 或 Socket**：填写 `Mongos` 主机地址，不包含端口。
    - **端口**：保持默认值，通常为`27017`。
    - **用户名**：填写`用户`。
    - **密码**：填写`密码`。
5. 点击**测试连接**验证连接是否成功，然后点击**创建**。

在 Bytebase 中创建实例并不会新建数据库。
实例创建完成后，你可以使用相应的客户端工具对数据进行查看和管理。