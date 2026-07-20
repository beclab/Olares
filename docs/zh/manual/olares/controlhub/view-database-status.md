---
description: 在 Olares 中使用 Bytebase 查看数据库状态。学习添加和管理 PostgreSQL、MongoDB 实例，使用 SQL 编辑器操作数据库。
---
# 在 Olares 中查看数据库状态

你可以使用第三方数据库中间件应用在 Olares 中查看数据库状态。本指南以 Bytebase 为例，演示如何通过中间件访问数据库。

## 准备工作

确保你已从**应用市场**安装 Bytebase。

![安装bytebase](/images/how-to/olares/controlhub/middleware/07.jpg#bordered)

## 添加 PostgreSQL 实例

在 Bytebase 中添加 PostgreSQL 实例：

1. 在 **Bytebase** 中，点击**添加实例**，选择 **PostgreSQL**。
2. 配置实例：
    - **实例名称**：输入 `Olares` 或其他名称。
    - **环境**：选择 `PROD` 或 `TEST`。
    - **主机**、**用户名**、**密码**：根据**控制面板**的**中间件**部分中的信息填写。

   ![配置PostgreSQL](/images/how-to/olares/controlhub/middleware/09.jpg#bordered)

3. 点击**创建**保存更改并连接实例。

现在，你应该能够查看刚刚添加的 PostgreSQL 实例的详细信息。

![PostgreSQL信息](/images/how-to/olares/controlhub/middleware/10.jpg#bordered)

## 添加 MongoDB 实例

在 Bytebase 中添加 MongoDB 实例：

1. 在 **Bytebase** 中，点击**添加实例**，选择 **MongoDB**。
2. 配置实例：
   - **实例名称**：输入 `Olares` 或其他名称。
   :::info 提示
   请勿使用重复的实例名称。
   :::
   - **环境**：选择 `PROD` 或 `TEST`。
   - **主机**、**用户名**、**密码**：根据 Control Hub 的 **中间件** 部分中的信息填写。

   ![配置MongoDB](/images/how-to/olares/controlhub/middleware/11.jpg#bordered)

3. 点击**创建**保存并连接实例。

现在，你应该能够查看刚刚添加的 MongoDB 实例的详细信息。

![MongoDB 信息](/images/how-to/olares/controlhub/middleware/12.jpg#bordered)

## 编辑数据库

在 Bytebase 中，点击右上角的 **SQL 编辑器**，进入**编辑器**页面以进行进一步操作。

![编辑数据库](/images/how-to/olares/controlhub/middleware/13.jpg#bordered)