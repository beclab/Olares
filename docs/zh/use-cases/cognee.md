---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/cognee
outline: [2, 3]
description: 在 Olares 上使用 Cognee 构建和浏览文档知识
head:
  - - meta
    - name: keywords
      content: "Olares, Cognee, 自托管, 知识图谱, 文档, 搜索"
app_version: "0.0.17"
doc_version: "1.0"
doc_updated: "2026-11-06"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/cognee.md)为准。
:::

# 把文档变成 AI 助手的知识库

Cognee 将文档转化为相互关联、可供你和 AI 代理回忆的知识。把文档加入数据集，提问并对照支持性来源，或者通过 MCP 让你的 AI 代理回忆同样的知识。在 Olares 上，你的文档和从中构建的知识都保留在自己的设备上。答案生成使用 Olares Router 中配置的默认对话模型。

## 学习目标

- 创建数据集并添加项目文档。
- 在 Cognee 中提问，并对照支持性来源和知识图谱检查答案。
- 通过 MCP 连接 Lares 或 Hermes Agent，然后在新对话中提问。

## 前提条件

- Olares 1.12.7 或更高版本。
- Olares Router 中为 `default-chat` 分配了状态为 **callable** 的模型。

## 安装 Cognee

1. 打开 Market，搜索 "Cognee"。
2. 点击 **Get**，然后点击 **Install**。提示时设置至少 12 个字符的初始密码。
3. 安装完成后，从 Launchpad 打开 Cognee。
4. 使用内置初始账号 `default_user@example.com` 和安装时设置的密码登录。

## 创建项目数据集

创建一个数据集，把项目文档添加进去。一个数据集可以存放一篇或多篇文档，搜索会覆盖其中的每一篇，所以把相关文档放在同一个数据集中。你可以粘贴文本，也可以上传文件。本指南使用下面的两份示例文档。

1. 打开 **Datasets**，点击 **Create dataset**。如果已有数据集，点击 **New dataset**。
2. 将数据集命名为 `project_rehearsal`。名称使用字母、数字和下划线。
3. 选中新建的数据集。
4. 选择 **Paste text**，把下面每份文档作为独立条目添加。等第一个条目显示 **Ready** 后再添加第二份。

   ::: details 会议纪要

   ```text
   Project Harbor: Migration Rehearsal Meeting Notes

   This document contains fictional test data. All people, dates, and project details are invented for a documentation walkthrough.

   Project: Harbor document migration
   Meeting date: November 4, 2026
   Rehearsal date: November 12, 2026, at 10:00 AM, Asia/Shanghai

   Maya Chen coordinates the migration rehearsal. She schedules the rehearsal and collects the validation results.

   Before handoff, the team must complete the rehearsal, confirm that the source and destination file counts match, verify the SHA-256 checksums, and attach the validation report to the handoff record.

   Omar Lee is the project owner. Handoff requires his approval after he reviews the validation report.

   The migration runbook identifies the escalation contact for checksum mismatches.
   ```

   :::

   ::: details 迁移校验运行手册

   ```text
   Project Harbor: Migration Validation Runbook

   This document contains fictional test data. All people and project details are invented for a documentation walkthrough.

   Scope: Validation for the Harbor document migration rehearsal.

   File validation compares the source and destination file counts and SHA-256 checksums. The validation report records the results and any unresolved differences.

   Noah Zhou is the data integrity lead and the escalation contact when migration checksums do not match. The report sent to Noah must identify the affected files and include the expected and actual checksums.

   A checksum mismatch blocks handoff until it has been investigated, resolved, and checked again. Any remaining mismatch must be recorded as unresolved in the validation report.

   Maya Chen coordinates the rehearsal. Omar Lee, the project owner, approves handoff after reviewing the completed validation report.
   ```

   :::

   :::info 首次处理任务
   第一篇文档的处理时间可能更长，因为 Cognee 需要下载它的嵌入模型。这个小型模型在 Cognee 内部使用 CPU 运行。后续处理任务会复用已保存的模型缓存。等当前导入完成后再重新提交同一文档。
   :::

## 查询项目数据集

对数据集提问，然后对照原始文档检查答案、Evidence 和图谱。

1. 打开 **Search**，在问题输入框上方的数据集选择器中选择 `project_rehearsal`，然后提问：

   ```text
   Who coordinates the rehearsal, what must be complete before handoff, and who should we contact if migration checksums do not match? Distinguish the meeting notes from the runbook in your sources.
   ```

2. 核对答案，并把 **Evidence** 中的支持性文本与两份原始文档对照。如果节选被截断，不要把它当作完整的来源核对。

   在这个例子中，答案会指出 Maya Chen 是协调人，说明交接前必须完成的校验和审批，并给出校验和不匹配时的联系人 Noah Zhou。

   :::tip 对话模型连接
   如果答案生成失败，请确认 **Default models** 页面已分配默认对话模型，且状态为 **callable**。参见 [Olares Router](olares-router.md)。
   :::

3. 打开 **Graph**，在左上角的 **DATASET** 选择器中选择 `project_rehearsal`。图谱布局和生成的关系可能因模型而异。

## 在 Lares 或 Hermes Agent 中查询

把 Lares 或 Hermes Agent 连接到 Cognee 的 MCP 服务器，这样你就可以在对话中搜索数据集，不用再粘贴文档。

### 获取 Cognee 的 MCP 地址

从 Olares Settings 复制 MCP endpoint URL，然后在末尾加上 `/mcp`。

1. 打开 Olares **Settings**，进入 **Applications** > **Cognee** > **Cognee MCP**，复制它的 endpoint URL。例如 `https://ba2717462.laresprime.olares.com`。
2. 在复制的 URL 末尾加上 `/mcp`。这就是下面助手连接要用的 MCP URL。

:::info 默认无需密钥访问
默认情况下，客户端不需要 API key。MCP 服务器使用内置初始账号 `default_user@example.com`，因此助手可以访问你通过这个账号添加的知识。已连接的助手共享该账号的权限。官方工具除了检索知识，也可以添加和删除知识。批准工具请求前请先检查。
:::

### 连接 Lares

在 Lares 中把 Cognee 注册为 MCP 服务器，然后提出一个会使用 `recall` 工具的问题。

1. 打开 Lares，打开它的设置，选择 **MCP**，然后点击 **Add server**。填写：

   | 字段 | 值 |
   | --- | --- |
   | **Server name** | `cognee` |
   | **Transport** | **Streamable HTTP** |
   | **MCP URL** | 上面获取的 MCP URL，包含 `/mcp` |
   | **Headers** | 保持默认的空 JSON 对象 `{}`。 |

2. 点击 **Save**。Lares 会连接并加载官方工具，包括 `recall` 和 `remember`。
3. 新建对话，发送下面的提示词。提示授权时批准 `recall` 工具。

      ```text
      Use Cognee recall to search dataset project_rehearsal for who coordinates the rehearsal and who to contact when migration checksums do not match. Set include_references to true.
      ```

   这条提示词指定了三件事：

   | 部分 | 是否必需 | 作用 |
   | --- | --- | --- |
   | `recall` 工具 | 是 | 在 Cognee 中搜索你已存储的知识。 |
   | 数据集名称 | 是 | `recall` 一次只搜索一个数据集。使用 Cognee 中显示的准确名称。 |
   | `include_references` | 否 | 同时返回答案背后的原文。这里设为 `true`，方便你在下一步和原始文档对照。只需要答案时可以省略。 |

4. 将答案和返回的原文与你添加的原始文档对照。

### 连接 Hermes Agent

1. 打开 Files，进入 **Data** > **hermesagent** > **home**。
2. 双击 `config.yaml`，点击右上角的编辑图标。
3. 在 `mcp_servers` 下添加以下连接，替换已有的 `cognee` 条目，其他条目保持不变。把 `<mcp-url>` 替换为上面获取的 MCP URL。这个示例只启用 `recall` 工具。

   ```yaml
   mcp_servers:
     cognee:
       url: <mcp-url>
       tools:
         include:
           - recall
   ```

4. 保存文件。
5. 从 Launchpad 打开 **Hermes CLI**。
6. 如果有已打开的 Hermes Agent 会话，输入 `/exit` 返回命令行。
7. 运行 `hermes` 启动新的 agent 会话并加载配置。使用和 Lares 相同的 recall 提示词。
8. 将答案和返回的原文与你添加的原始文档对照。

## 了解更多

- [Olares Router](olares-router.md)：管理模型路由，包括 `default-chat` 端点。
- [Cognee 文档](https://docs.cognee.ai/)：数据集、图谱和 Cognee API 的官方参考。
