---
outline: [2, 3]
description: 在 Olares 上使用 n8n，通过可视化节点、凭证、NocoDB 记录、Webhook 和可复用的工作流 JSON 构建自托管自动化工作流。
head:
  - - meta
    - name: keywords
      content: Olares, n8n, 工作流自动化, 自托管, 无代码, 低代码, NocoDB, Webhook, 集成
app_version: "1.0.16"
doc_version: "2.0"
doc_updated: "2026-07-31"
---

:::info AI 翻译说明
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/n8n.md)。
:::

# 使用 n8n 实现工作流自动化

n8n 是一款可自托管的工作流自动化工具，采用可视化节点编辑器。你可以连接应用和 API、转换数据、在需要自定义逻辑时运行 JavaScript，还可以通过定时任务、Webhook、表单或事件触发自动化流程。

在 Olares 上运行 n8n，工作流定义、凭证和执行历史都由你自己掌控。

本教程以版本发布监控为例。相同方法也适用于其他自动化场景，例如调用 API、转换数据、将结果保存到其他应用，以及通过 Webhook 实时接收事件。

## 学习目标

通过本教程，你将学会：

- 从应用市场安装 n8n。
- 创建首个 n8n 所有者账号。
- 构建调用 API 并转换响应数据的工作流。
- 将工作流结果保存到其他应用，本教程以 NocoDB 为例。
- 通过 Webhook 实时接收事件，本教程以 GitHub Release 为例。
- 将工作流下载或导入为 JSON 文件。

## 安装 n8n

1. 打开应用市场，搜索“n8n”。

   ![n8n](/images/manual/use-cases/n8n.png#bordered)

2. 点击**获取**，然后点击**安装**，等待安装完成。

## 设置 n8n

1. 从启动台打开 n8n。
2. 在设置页面输入邮箱、名字、姓氏和密码。首个注册用户会自动成为该 n8n 实例的所有者。

   ![创建 n8n 所有者账号](/images/manual/use-cases/n8n-create-owner.png#bordered)

3. 点击**下一步**。n8n 会打开编辑器页面，你可以在这里创建第一个工作流。

## 构建第一个工作流

本示例检查 n8n 的最新 GitHub Release，并提取版本跟踪所需的字段。如需监控其他项目，请将请求 URL 中的 `n8n-io/n8n` 替换为其他 `owner/repository`，例如 `immich-app/immich`。

### 创建工作流

1. 点击 **Build a workflow** 打开编辑器页面。
2. 点击 **Add first step**，然后在右侧面板中选择 **Trigger manually**。
3. 点击 Manual Trigger 节点后的 <i class="material-symbols-outlined">add</i>，搜索“HTTP Request”并添加该节点。
4. 打开 HTTP Request 节点的 **Parameters** 标签页，设置以下参数：

   - **Method**：`GET`
   - **URL**：
     ```text
     https://api.github.com/repos/n8n-io/n8n/releases/latest
     ```
   - **Authentication**：`None`

5. 点击 **Execute step**。输出面板会显示 API 响应。

   ![执行 HTTP Request 节点](/images/manual/use-cases/n8n-http-request-test.png#bordered)

### 保留所需的版本信息

添加 Edit Fields 节点，将完整的 API 响应整理成简洁的版本摘要。

1. 点击 HTTP Request 节点后的 <i class="material-symbols-outlined">add</i>，搜索“Edit Fields”并添加该节点。

   ![添加 Edit Fields 节点](/images/manual/use-cases/n8n-add-edit-fields-node.png#bordered)

2. 将 **Mode** 设置为 **Manual Mapping**。
3. 点击 **Add Field**，添加以下字段。分别设置字段名称、类型和值：

   | 字段 | 类型 | 值 |
   |:------|:-----|:---|
   | `app` | String | `n8n` |
   | `latest_version` | String | <code v-pre>{{ $json.tag_name }}</code> |
   | `published_at` | String | <code v-pre>{{ $json.published_at }}</code> |
   | `release_notes` | String | <code v-pre>{{ $json.html_url }}</code> |
   | `is_prerelease` | String | <code v-pre>{{ $json.prerelease }}</code> |

4. 点击 **Execute step**。
5. 点击左上角的节点名称，将其重命名，例如 `Monitor n8n releases`。

   ![在 n8n 中编辑字段](/images/manual/use-cases/n8n-edit-fields.png#bordered)

:::tip 发布前先测试
n8n 支持在构建阶段手动执行工作流，并在发布后自动执行工作流。发布前，请先逐个测试节点。
:::

### 运行并检查执行记录

在编辑器中运行一次工作流，然后通过 **Executions** 标签页查看运行历史和节点输出。

1. 打开刚才创建的工作流。
2. 点击编辑器底部的 **Execute workflow**。
3. 等待运行完成。成功运行的节点会显示绿色对勾，节点之间的连线上会显示传递的项目数量。
4. 点击编辑器顶部的 **Executions**。
5. 选择目标执行记录，检查各节点的输入、输出、状态和耗时。

   ![检查 n8n 执行记录](/images/manual/use-cases/n8n-execution-details.png#bordered)

工作流准备好自动运行后，将 Manual Trigger 替换为 **On a schedule**，然后点击 **Publish**。对于版本监控，每天或每周运行一次通常已经足够。

## 将工作流结果保存到 NocoDB

本示例将版本摘要保存到 NocoDB。你也可以使用相同方法，将表单提交、监控结果、Webhook 载荷或 API 响应写入表格。

n8n 会将凭证与工作流逻辑分开保存，因此同一个 NocoDB API token 可以在多个工作流中重复使用。

:::info 先设置 NocoDB
如果尚未创建 NocoDB Base 和表格，请先参阅[使用 NocoDB 构建自托管智能表格数据库](nocodb.md#创建表格和导入数据)，完成后再返回本节。
:::

### 准备 NocoDB

1. 打开 NocoDB 并创建一张表，例如 `Release checks`。
2. 添加以下列。首次测试时可全部使用文本字段：

   | 列 | 类型 |
   |:---|:-----|
   | `app` | Single line text |
   | `latest_version` | Single line text |
   | `published_at` | Single line text |
   | `release_notes` | URL |
   | `is_prerelease` | Single line text |

### 获取 NocoDB 连接信息

<!--@include: ../reusables/ai-service-connections.md#app-endpoint-overview-->

1. 前往 Olares **设置** > **应用** > **NocoDB** > **入口**。
2. 选择 **NocoDB**。
3. 复制 **Endpoint** URL。
4. 在 NocoDB 中前往 **Account Settings** > **API Tokens**，创建并复制 token。

:::warning 妥善保管 API token
API token 可以访问你的 NocoDB 数据。请勿与他人分享、放入截图，或提交到公开代码仓库。
:::

### 添加 NocoDB 节点

1. 返回包含 `Monitor n8n releases` 节点的工作流。
2. 点击 `Monitor n8n releases` 节点后的 <i class="material-symbols-outlined">add</i>，搜索“NocoDB”，选择该节点，然后选择 **Create a row**。
3. 点击 **Set up credential**，输入 NocoDB 连接信息：

   | 字段 | 值 |
   |:-----|:---|
   | **API Token** | 之前创建的 NocoDB API token。 |
   | **Host** | 从 Olares 设置中复制的 NocoDB Endpoint URL。 |

4. 点击 **Save**。n8n 会自动测试连接。

   ![添加 NocoDB 凭证](/images/manual/use-cases/n8n-add-nocodb-credential.png#bordered)

5. 打开 NocoDB 节点的 **Parameters** 标签页，设置以下参数：

   | 设置 | 值 |
   |:-----|:---|
   | **Resource** | `Row` |
   | **Operation** | `Create` |
   | **Base Name or ID** | 选择包含 `Release checks` 的 Base。 |
   | **Table Name or ID** | 选择 `Release checks`。 |
   | **Data to Send** | 根据需要选择数据发送方式。 |

6. 点击 **Execute step**。节点成功运行后，返回 NocoDB，确认 `Release checks` 表中出现了新记录。

   ![检查 NocoDB 结果](/images/manual/use-cases/n8n-check-nocodb-result.png#bordered)

7. 点击 **Execute workflow**，测试完整流程。

如需自动记录版本检查结果，请将 Manual Trigger 替换为 **On a schedule**，测试完成后点击 **Publish**。

## 通过 Webhook 接收事件

Webhook 工作流允许外部服务实时向 n8n 发送事件。本示例从你拥有或管理的 GitHub 仓库接收版本发布事件，然后提取版本信息。

只有拥有仓库管理员权限，才能添加 GitHub Webhook。如需监控你无权管理的公共仓库，例如 `n8n-io/n8n`，请使用本教程前面的定时 HTTP Request 工作流。

:::warning 公开入口存在安全风险
可以将 n8n 入口设为**公开**，无需连接 LarePass VPN 即可使用 Webhook。但这会将 n8n 入口暴露到互联网。为提高安全性，建议尽量保持默认的**内部**身份验证级别并使用 LarePass VPN。
:::

### 创建 Webhook 工作流

:::info 监听前开启 LarePass VPN
当 n8n 入口使用默认的**内部**身份验证级别时，请先在 LarePass 电脑端开启 VPN，再点击 **Listen for test event**。配置和测试 Webhook 期间，请保持 VPN 连接。
:::

1. 在 n8n 左上角点击 <i class="material-symbols-outlined">add</i>，然后选择 **Workflow**。
2. 点击 **Add first step**，然后在右侧面板中选择 **On webhook call**。
3. 打开 Webhook 节点的 **Parameters** 标签页，设置以下参数：

   | 设置 | 值 |
   |:-----|:---|
   | **HTTP Method** | `POST` |
   | **Path** | `github-release-event` |
   | **Authentication** | `None` |
   | **Respond** | `Immediately` |

4. 点击 **Listen for test event**，然后从 Webhook 节点复制 **Test URL**。在 GitHub 中使用 `https://`。如果 n8n 复制出的 URL 以 `http://` 开头，请将其改为 `https://`。

### 在 GitHub 中添加 Webhook

1. 在浏览器中打开 GitHub，进入你拥有或管理的仓库。
2. 在仓库中前往 **Settings** > **Webhooks**，然后点击 **Add webhook**。
3. 设置 Webhook：

   | 设置 | 值 |
   |:-----|:---|
   | **Payload URL** | 粘贴 n8n 提供的 HTTPS 测试 Webhook URL。 |
   | **Content type** | `application/json` |
   | **Secret** | 首次测试时留空。 |
   | **Which events would you like to trigger this webhook?** | 选择 **Let me select individual events**，然后选择 **Releases**。 |

   ![添加 Webhook](/images/manual/use-cases/n8n-add-webhook.png#bordered)

4. 保持 **Active** 开启，然后点击 **Add webhook**。
5. 返回 n8n，等待测试事件。添加 Webhook 后，GitHub 会发送一个 `ping` 事件。
6. 如需测试完整的版本发布载荷，请在仓库中发布一个测试版本，然后返回 n8n 检查 Webhook 节点的输出。

   ![Webhook 测试结果](/images/manual/use-cases/n8n-webhook-test-result.png#bordered)

### 提取版本字段

GitHub 首次发送的 `ping` 事件不包含 release 对象。n8n 收到版本发布事件后，在 Webhook 节点后添加 Edit Fields 节点，使载荷内容更易阅读。

1. 点击 Webhook 节点后的 <i class="material-symbols-outlined">add</i>，搜索“Edit Fields”并添加该节点。
2. 将 **Mode** 设置为 **Manual Mapping**。
3. 点击 **Add Field**，添加以下字段：

   | 字段 | 类型 | 值 |
   |:-----|:-----|:---|
   | `event_action` | String | <code v-pre>{{ $json.body.action }}</code> |
   | `repository` | String | <code v-pre>{{ $json.body.repository.full_name }}</code> |
   | `release_tag` | String | <code v-pre>{{ $json.body.release.tag_name }}</code> |
   | `release_name` | String | <code v-pre>{{ $json.body.release.name }}</code> |
   | `release_url` | String | <code v-pre>{{ $json.body.release.html_url }}</code> |
   | `sender` | String | <code v-pre>{{ $json.body.sender.login }}</code> |

4. 点击 **Execute step**，检查提取后的字段。

   ![提取版本字段](/images/manual/use-cases/n8n-extract-release-fields.png#bordered)

然后，你可以将提取的字段连接到其他节点，例如 NocoDB、Slack 或邮件节点。

### 发布并切换到生产 URL

测试事件正常后，发布工作流，并将 GitHub Webhook 更新为生产 URL。

1. 点击 n8n 右上角的 **Publish**。工作流发布后，生产 Webhook 会开始接收事件。
2. 打开 Webhook 节点并复制 **Production URL**。
3. 如果生产 URL 以 `http://` 开头，请将其改为 `https://`。
4. 返回 GitHub Webhook 设置，将测试 URL 替换为 HTTPS 生产 URL。
5. 保存 GitHub Webhook 设置。

   <!-- ![在 n8n 中设置 Webhook 触发器](/images/manual/use-cases/n8n-webhook-trigger.png#bordered) -->

## 管理工作流

n8n 工作流可以下载为 JSON 文件，也可以导入其他 n8n 实例。这适用于备份、版本审查和与团队成员共享工作流模板。

### 下载工作流

1. 打开工作流。
2. 点击 <i class="material-symbols-outlined">more_horiz</i> > **Download**。

   ![下载工作流](/images/manual/use-cases/n8n-download-workflow.png#bordered)

JSON 文件会下载到你的电脑。

### 导入工作流

1. 创建一个新工作流。
2. 点击 <i class="material-symbols-outlined">more_horiz</i> > **Import from File**。

   ![导入工作流](/images/manual/use-cases/n8n-import-workflow.png#bordered)

3. 选择工作流 JSON 文件。
4. 发布前，检查每个凭证字段并重新连接凭证。

:::warning 检查导入的工作流
导入的工作流可能包含 Code 节点、HTTP 请求和 Webhook 路径。运行来自不可信来源的工作流前，请检查所有节点。
:::

## 常见问题

### 为什么 Webhook 无法接收外部服务发送的事件？

请检查以下项目：

- 确认添加到外部服务的 Webhook URL 以 `https://` 开头。
- 测试时，点击 **Listen for test event** 后，确认 n8n 仍在等待事件。
- 发布后，确认外部服务使用的是 **Production URL**，而不是测试 URL。
- 如果外部服务提供投递日志，请检查其中的响应状态。
- 如果 n8n 入口使用**内部**，请在当前电脑的 LarePass 客户端中开启 VPN，并在测试期间保持连接。
- 如果 Webhook 仍然失败，可以将入口设为**公开**。这会将 n8n 暴露到互联网，请仅在接受相关安全风险时使用。

## 了解更多

- [n8n 工作流文档](https://docs.n8n.io/workflows/)：了解工作流、节点、模板、执行记录和共享功能。
- [n8n 集成文档](https://docs.n8n.io/integrations/)：浏览内置节点、社区节点、仅凭证节点和通用 API 集成选项。
- [HTTP Request 节点](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.httprequest/)：配置 REST API 调用，或将 `curl` 命令导入 n8n。
- [在 Olares 上使用 NocoDB](nocodb.md)：创建表格、导入数据、设置邮件并管理团队访问权限。
