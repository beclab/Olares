---
outline: [2, 3]
description: 了解如何在 Olares 中通过 Bytebase 查看并管理 Elasticsearch 数据。
---
# 查看 Elasticsearch 数据

Elasticsearch 需要先在应用市场安装后才能使用。本文将介绍安装流程，并说明如何在 Olares 中访问并管理 Elasticsearch 数据。

## 安装 Elasticsearch 服务

在建立连接之前，需要先通过应用市场安装 Elasticsearch 服务。

1. 从启动台打开应用市场，搜索“Elasticsearch”。
2. 点击**获取**，然后点击**安装**，并等待安装完成。

安装完成后，Elasticsearch 服务及其连接信息将显示在控制面板的中间件列表中。

## 获取连接信息

在建立连接之前，需要从控制面板获取 Elasticsearch 的连接信息。

1. 从启动台打开控制面板。
2. 在左侧导航栏中找到中间件，并选择 **Elasticsearch**。
3. 记录信息面板中的以下信息：
    - **主机**：用于在 Bytebase 中建立连接。
    - **用户**：用于在 Bytebase 中建立连接。
    - **密码**：用于在 Bytebase 中建立连接。

    ![Elasticsearch 详情](/images/zh/manual/developer/mw-es-details.png){width=60% style="margin-left:0"}

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

### 创建 Elasticsearch 实例

1. 使用管理员账号登录 Bytebase。
2. 在左侧导航栏中选择**实例**，然后点击 <b>+ 添加实例</b>。
3. 选择 **Elasticsearch** 作为数据库类型。
4. 使用控制面板中获取的信息填写连接配置：
    - **Host 或 Socket**：填写`主机`地址，不包含端口。
    - **端口**：保持默认值，通常为`9200`。
    - **用户名**：填写`用户`。
    - **密码**：填写`密码`。
5. 点击**测试连接**验证连接是否成功，然后点击**创建**。

在 Bytebase 中创建实例并不会新建数据库。
实例创建完成后，你可以使用相应的客户端工具对数据进行查看和管理。