---
outline: [2, 3]
description: 在 Olares 上使用 Local Deep Research，通过 AI Router 和 SearXNG 研究问题、核查引用来源，并将报告导出为 Markdown 或 PDF。
head:
  - - meta
    - name: keywords
      content: "Olares, Local Deep Research, AI 研究, AI Router, SearXNG, 来源引用, 研究报告, 本地模型"
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-13"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/local-deep-research.md)为准。
:::

# 使用 Local Deep Research 生成带来源引用的研究报告

Local Deep Research 是一个结合网页搜索与语言模型的开源 AI 研究助手。它会生成后续问题、收集资料，并将发现整理成带引用的报告。在 Olares 上，应用通过 AI Router 调用模型，通过 SearXNG 搜索网页，账号数据库和研究历史保存在应用存储中。

本指南以“热泵如何工作”为例，使用快速摘要模式和一轮搜索，帮助你先检查完整使用流程，再开展范围更大的研究。

## 学习目标

通过本指南，你将学会：

- 配置模型连接和网页搜索。
- 创建应用账号并运行研究问题。
- 核查引用、重新打开历史研究，以及导出报告。
- 验证配置，并定位模型或搜索失败的原因。

## 前提条件

- Olares 1.12.7 或更高版本。
- [AI Router](olares-router.md) 已配置能正常回答问题的聊天模型。
- 已从 Market 安装 SearXNG，并确认它可以返回搜索结果。

Local Deep Research 本身不下载或托管语言模型。如果使用本地模型，请先完成模型下载，并等待推理服务就绪。具体步骤参见[在 Olares 上运行本地大语言模型](llm-base-apps.md)。Router 也可以连接已配置的云端模型服务。

## 准备模型和搜索连接

1. 打开 Router，确认聊天模型能回答一个简短问题。应用初始使用 Router 的 `default-chat` 路由。模型出现在列表中，并不代表推理服务已就绪或服务商的凭据有效。

2. 打开**设置** > **应用** > **SearXNG**，在**共享入口**中复制 SearXNG 的地址。保留设置中显示的协议，不要在末尾添加 `/search`。

3. 打开 SearXNG，搜索 `how heat pumps work`。确认页面包含实际搜索结果，而不是只有搜索引擎错误提示。

配置 Local Deep Research 时，使用以下值：

| 字段 | 初始值 | 用途 |
|:---|:---|:---|
| `SEARXNG_URL` | 从设置中复制的 SearXNG 地址 | 必填的搜索服务地址 |
| `LDR_LLM_MODEL` | `default-chat` | Router 的默认聊天路由 |
| `LDR_LLM_OPENAI_ENDPOINT_URL` | 留空 | 自动连接当前 Olares 的 Router |
| `LDR_LLM_OPENAI_ENDPOINT_API_KEY` | 留空 | 应用通过 Olares 身份调用 Router |

这些连接设置与应用账号的用户名和密码相互独立。

## 安装 Local Deep Research

安装时，在 `SEARXNG_URL` 中填入 SearXNG 地址。使用默认 Router 连接时，模型接口地址和 API key 保持为空。

1. 打开 Market，搜索“Local Deep Research”。
   <!-- ![Market 中的 Local Deep Research](/images/manual/use-cases/local-deep-research-market.png#bordered) -->

2. 点击**获取**，然后点击**安装**，等待安装完成。

## 创建应用账号

1. 从 Launchpad 打开 Local Deep Research。
2. 在登录页点击 **Register here**。
3. 输入用户名和密码，再次输入密码进行确认。
4. 阅读加密数据的提示，勾选确认框，然后点击 **Create Encrypted Account**。

这个账号在 Local Deep Research 网页中创建，与 Olares 账号相互独立，也不通过安装时的环境变量设置。之后访问应用时，使用这个应用账号登录。

:::warning 妥善保管密码
密码用于保护加密的账号数据库，该数据库不支持密码找回。保存重要研究或文档之前，请先安全保存密码。
:::

## 使用 Local Deep Research

### 运行第一个研究问题

1. 打开 **New Research**，输入：

   ```text
   How do heat pumps work? Give a short explanation with source citations.
   ```

2. 选择 **Quick Summary**。
3. 展开 **Advanced Options**，配置以下参数：

   | 设置项 | 值 |
   |:---|:---|
   | **Model Provider** | OpenAI-Compatible Endpoint |
   | **Language Model** | `default-chat`，或你选定的具体 Router 模型 |
   | **Search Engine** | SearXNG |
   | **Search Iterations** | `1` |
   | **Questions Per Iteration** | `1` |
   | **Search Strategy** | Source-Based |

   通过应用环境变量提供的模型和接口字段可能显示为不可编辑。这是正常行为，需要在 Olares 设置中修改这些值。

4. 点击 **Start Research**。在进度页查看模型生成问题、搜索资料和整理报告的过程。
5. 处理结束后，点击 **View Results**。

![Local Deep Research 的问题输入与研究模式](/images/manual/use-cases/local-deep-research-query.png#bordered)

一轮迭代可能同时搜索原问题和模型生成的后续问题。增加迭代次数可以扩展调查范围，也会增加模型调用和搜索请求。

### 核查并导出报告

1. 阅读报告，确认它包含对问题的回答，而不是错误信息或故障排查报告。
2. 打开一条引用，将报告中的说法与原始资料对照。检查发布时间、发布机构，以及原文是否支持这条结论。
3. 点击 **Back to History**，再重新打开这次研究，确认结果已保存。
4. 如需文本副本，点击 **Export** > **Markdown**。
5. 如需便于阅读或分享的文档，点击 **Download PDF**。打开下载的文件，检查报告正文和引用列表。

引用让结论可以追溯，但不能保证结论正确。将报告用于决策之前，请核查原始来源。

### 选择研究模式

| 模式 | 适用场景 |
|:---|:---|
| **Quick Summary** | 获得初步回答，或快速检查配置 |
| **Detailed Report** | 生成按多个章节组织的较长研究报告 |
| **Chat Mode** | 在实验性的对话流程中提出后续问题 |

建议先使用 Quick Summary，确认模型和搜索引擎正常后再扩大研究范围。下面的检查项覆盖快速摘要流程，不代表所有研究模式和文档索引选项都已验证。

## 验证配置

### 检查核心流程

使用上面的热泵问题完成以下检查。判断实际报告内容，不要只看进度百分比。

| 测试项 | 操作 | 预期结果 |
|:---|:---|:---|
| 账号访问 | 退出登录，再使用应用账号登录 | 能正常进入应用 |
| 模型与搜索连接 | 运行一轮迭代的 Quick Summary | 生成有实际内容和引用的回答，且没有身份验证错误 |
| 来源核查 | 打开几条支持关键结论的引用 | 来源可以访问，且原文支持报告中的说法 |
| 历史保存 | 刷新报告页，再从 **History** 中打开 | 问题、报告和引用仍然保留 |
| Markdown 导出 | 下载并打开 Markdown 文件 | 文件包含报告正文和来源引用 |
| PDF 导出 | 下载并打开 PDF | 文档能正常打开，正文和引用可读 |

在一次通过 Router 调用 Qwen3.6-27B MTP 的参考运行中，热泵示例约 130 秒完成，保存了 8 个来源。历史保存、Markdown 导出和两页 PDF 均已检查通过。耗时、来源数量和文档长度取决于模型、搜索结果和设置；这些数字是参考观察值，不是测试通过的标准。来源内容仍需单独核查。

### 尝试其他研究问题

以下问题可用于进一步评估自己的配置，不是基准测试结果，也不保证得到特定答案。

| 类型 | 示例问题 | 检查重点 |
|:---|:---|:---|
| 技术比较 | 比较 PostgreSQL 和 MySQL 在事务型网页应用中的适用性。讨论并发控制、索引和复制，尽可能引用官方文档。 | 比较是否均衡，版本是否明确，结论是否有当前文档支持 |
| 历史解释 | 哪些因素促成了西罗马帝国的灭亡？区分学术解释与存在争议的说法，并引用来源。 | 是否解释多个原因、说明不确定性，并有资料支持其解释 |
| 科学解释 | 为什么锂离子电池的容量会随时间衰减？区分日历老化与循环老化，并引用权威资料。 | 是否区分两类机制，使用合适的技术来源，避免无依据的数值结论 |

如果希望中文输出，可在问题末尾添加 `Answer in Chinese and retain source citations.`。比较回答时，除了语言是否流畅，也要检查引用来源。

## 修改模型连接

如需指定某个 Router 模型，请在 Router 中打开该模型，从 **How to call this model** 窗口复制模型标识。使用包含服务商前缀的完整 Router 标识。模型应用在 Launchpad 页面显示的原始模型名可能不包含这个前缀。

打开**设置** > **应用** > **Local Deep Research**，找到应用环境变量，将 `LDR_LLM_MODEL` 设置为复制的标识。例如，提供参考运行所用模型的 Router 实例会显示：

```text
Olares/unsloth/Qwen3.6-27B-MTP-GGUF:UD-Q4_K_XL
```

仅在自己的 Router 中已有该模型时使用这个示例值。接口地址和 API key 保持为空，继续通过 Router 调用。等待应用完成设置更新；如果出现登录页，再使用应用账号登录。

如需连接其他兼容 OpenAI 的服务，请填写它的模型名、包含 `/v1` 的 API 基础地址，以及服务要求的 API key。服务商的 API key 用于验证模型请求，不能替代 Local Deep Research 的账号密码。

## 排查常见问题

| 现象 | 检查方法 |
|:---|:---|
| 出现 `401`、无效 API key 或模型身份验证错误 | 检查 Router 中的服务商凭据。默认 Olares Router 连接的应用 key 留空是正确设置；必要时选择另一个可用的 Router 模型。 |
| 提示找不到模型 | 从 Router 复制完整的可调用标识，不要假设本地模型的原始名称或其他实例的标识可以直接使用。 |
| 搜索没有返回来源 | 检查 `SEARXNG_URL`，并在 SearXNG 中直接搜索同一问题。查看搜索引擎错误；如果上游限流，稍后再试。 |
| 进度显示已完成，但报告包含错误 | 阅读报告和研究日志，解决模型或搜索问题后再运行。进度完成不等于回答成功。 |
| 研究表单中的模型字段不可编辑 | 在 Olares 设置中修改应用环境变量，并等待设置生效。 |
| 重启后要求重新登录 | 使用应用账号登录。出现新的登录提示，本身不代表已保存的研究被删除。 |

网页研究会将问题发送给所选搜索服务；使用云端模型时，用于推理的内容也会发送给该服务。使用本地模型可以将模型推理留在 Olares 上，但公开网页搜索仍需要联网。

## 了解更多

- [使用 Olares Router 作为 AI 网关](olares-router.md)：了解模型标识、访问方式和服务商配置。
- [在 Olares 上运行本地大语言模型](llm-base-apps.md)：配置可供 Router 使用的本地模型。
- [共享应用](../manual/olares/market/shared-apps.md)：了解 Olares 的共享应用访问方式。
