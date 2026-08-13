---
outline: [2, 3]
description: 在 Olares 上将 NocoDB 设置为无代码数据库平台。创建表格、导入数据、配置 SMTP、管理团队访问权限，并与 n8n 集成实现工作流自动化。
head:
  - - meta
    - name: keywords
      content: Olares, NocoDB, 无代码数据库, Airtable 替代方案, 自托管, 智能表格, 工作流自动化, n8n
app_version: "1.0.10"
doc_version: "2.0"
doc_updated: "2026-07-31"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/nocodb.md)为准。
:::

# 使用 NocoDB 构建自托管智能表格数据库

NocoDB 是一款开源无代码数据库平台，可以将数据库转换成类似 Airtable 的智能表格界面。它提供功能丰富的 Web UI，方便你直观地管理数据，同时还提供完整的 REST API。在 Olares 上运行 NocoDB，可以获得一个自托管、注重隐私的云端智能表格替代方案。

## 学习目标

通过本教程，你将学会：

- 安装 NocoDB 并创建管理员账号。
- 创建表格并从外部来源导入数据。
- 配置 SMTP 以发送邮件。
- 邀请团队成员并管理权限。
- 继续配置将数据写入 NocoDB 的 n8n 工作流。

## 安装 NocoDB

1. 打开应用市场，搜索“NocoDB”。

   ![NocoDB](/images/manual/use-cases/nocodb.png#bordered)

2. 点击**获取**，然后点击**安装**，等待安装完成。

## 设置 NocoDB

1. 从启动台打开 NocoDB。
2. 输入邮箱和密码，然后点击 **Sign Up**。

   首个注册用户会自动成为超级管理员，可以管理团队成员权限。

   ![NocoDB 注册页面](/images/manual/use-cases/nocodb-register.png#bordered){width=80%}

## 创建表格和导入数据

你可以手动创建表格，也可以导入现有数据。

1. 打开默认的 **Getting Started** Base，或从工作区菜单中选择其他 Base。
2. 使用以下任一方式创建表格：

   - 在 **Overview** 页面点击 **Create New Table**。
   - 在左侧边栏点击 **Create New** > **Table**。

   ![在 NocoDB 中创建表格](/images/manual/use-cases/nocodb-create-table.png#bordered)

3. 如需导入数据，请前往 **Overview**，点击 **Import Data**，然后选择支持的格式：

   - Airtable
   - CSV
   - Excel
   - JSON

   ![将数据导入 NocoDB](/images/manual/use-cases/nocodb-import-data.png#bordered)

## 配置邮件

配置 SMTP 后，NocoDB 可以使用指定的发件人地址发送邮件。

1. 点击左下角的个人资料图标，然后进入 **Account Settings**。
2. 在 **Configure E-mail** 面板中点击 **Configure**。
3. 选择 **SMTP**，然后填写邮件服务提供商给出的 SMTP 设置。

   | 字段 | 值 |
   |:-----|:---|
   | **From address** | 发件人邮箱地址，例如 `name@example.com`。 |
   | **From domain** | `@` 后面的域名，例如 `example.com`。 |
   | **SMTP server** | 邮件服务提供商给出的 SMTP 服务器地址，例如<br>`smtp.example.com`。 |
   | **SMTP port** | 邮件服务提供商指定的 SMTP 端口。TLS 通常使用 `587`，SSL 通常使用 `465`，<br>不安全连接可使用 `25`。 |
   | **Username** | SMTP 用户名，通常是完整的邮箱地址。 |
   | **Password** | SMTP 密码、应用专用密码或授权码。 |

4. 根据邮件服务提供商的要求调整安全开关。
5. 点击 **Test** 检查连接，然后点击 **Save** 保存 SMTP 设置。

## 邀请团队成员

1. 点击左下角的个人资料图标，然后进入 **Account Settings**。
2. 在左侧边栏展开 **Users**，选择 **User Management**。
3. 点击右上角的 **Invite User**。
4. 输入团队成员的邮箱地址，设置适当的访问级别，然后点击 **Invite**。

   ![在 NocoDB 中邀请团队成员](/images/manual/use-cases/nocodb-invite-member.png#bordered)

5. 如果 NocoDB 显示 **Copy Invite URL**，请复制该 URL 并发送给受邀成员。

受邀成员可以通过邀请邮件或邀请 URL 注册。

## 使用 n8n 自动处理 NocoDB 数据

NocoDB 很适合作为 n8n 自动化工作流的数据存储。创建 Base 和表格后，请参阅[将工作流结果保存到 NocoDB](n8n.md#将工作流结果保存到-nocodb)，完成以下操作：

- 从 Olares 设置中复制当前 NocoDB Endpoint。
- 创建 NocoDB API token。
- 在 n8n 工作流中添加 NocoDB 凭证和 NocoDB 节点。
- 测试工作流能否在表格中创建新记录。

## 了解更多

- [使用 n8n 实现工作流自动化](n8n.md)：构建工作流、将结果保存到 NocoDB、接收 Webhook 并复用工作流 JSON。
- [NocoDB 文档](https://docs.nocodb.com/)：查看 NocoDB 功能和 API 参考。
