---
outline: [2, 3]
description: 在 Olares 上使用 TrustGraph，将文档转为可检索的知识。配置 Router 模型、处理文件、结合来源进行问答，并探索实体之间的关系。
head:
  - - meta
    - name: keywords
      content: Olares, TrustGraph, 知识图谱, 文档问答, Document RAG, Graph RAG, 本地模型, 文档检索
app_version: "0.1.20"
doc_version: "1.0"
doc_updated: "2026-10-13"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/trustgraph.md)为准。
:::

# 使用 TrustGraph 将文档转为可检索的知识

TrustGraph 是开源的语义智能平台，可将源数据转为能够追溯来源的知识。它结合文档检索和知识图谱，让你围绕文件内容提问，并查看从文档中提取的实体与关系。

在 Olares 上，TrustGraph 使用 Olares 管理的存储保存文档，通过 Router 调用模型。本案例使用本地模型处理一段实验室介绍，再进行问答和图谱探索。

## 前提条件

- 运行 Olares 1.12.7 或更高版本的 amd64 或 arm64 设备。应用需要 10 GiB 内存和 20 GiB 磁盘空间；模型所需资源另行计算。
- 可以访问 [Olares Router](olares-router.md)，且聊天模型、EmbeddingGemma 和 BGE Reranker v2 M3 已安装并可调用。

## 检查模型可用性

TrustGraph 需要三种模型能力：聊天模型提取知识并生成回答，向量模型查找相关内容，排序模型对检索结果重新排序。

1. 从启动台打开已安装的模型应用，等待模型下载并启动完成。
2. 从启动台打开 Router。在 **Default models** 页面，将聊天模型设为 `default-chat`，将 BGE Reranker v2 M3 设为 `default-rerank`。
3. 确认以下模型在 Router 中可调用：

   | 能力 | TrustGraph 默认流程使用的模型 |
   | :--- | :--- |
   | 聊天 | `default-chat`，路由到你选择的聊天模型 |
   | 向量 | `Olares/embeddinggemma-300m`，生成 768 维向量 |
   | 排序 | `default-rerank`，路由到 BGE Reranker v2 M3 |

   TrustGraph 的流程明确指定了向量模型。修改 Router 的默认向量模型不会改变这个流程。

## 安装 TrustGraph

1. 打开 Market，搜索“TrustGraph”。
2. 点击 **获取**，再点击 **安装**。按提示设置 `ADMIN_PASSWORD`，密码须为 12–128 个字符且不含换行，然后等待安装完成。

## 登录

1. 从启动台打开 TrustGraph。
2. 使用用户名 `admin` 和安装时设置的密码登录。

管理员密码、Olares 密码和 Router API Key 是不同的凭据。通过这个登录表单进入应用时，无需填写 TrustGraph API Key。重启和升级会保留已有账号密码，安装字段不会重置密码。

## 使用 TrustGraph

### 上传并处理文档

上传文件后，还需提交处理。流程（flow）决定处理方式，集合（collection）用于组织处理后可检索的知识。

1. 在电脑上将以下内容保存为 `orion-laboratory.txt`：

   ```text
   Orion laboratory is located in Suzhou. It is directed by Dr. Chen and studies hydrogen fuel cells.
   ```

2. 在 **Welcome** 页面打开 **Document Ingestion**。点击 **+ Add Document**，选择 **File**，然后选择文本文件。
3. 将 **TITLE** 填为 `Orion laboratory`，点击 **Upload**。
4. 选中上传的文档，点击 **Submit for Processing →**。
5. 选择 `default` 流程。点击 **+ Create New Collection**，在 **ID** 中填写 `orion-notes`，在 **NAME** 中填写 `Orion notes`，然后点击 **Next →**。
6. 在确认面板核对文档、流程和集合，点击 **Submit for Processing**。
7. 等待处理完成，再进行问答。后续步骤中，将顶部 **COL** 设为 `orion-notes`，将 **FLOW** 设为 `default`。

导入页面显示文档、流程和存储之间的连接。仅看到已上传文件或提交次数，并不代表知识已经可以查询。

:::tip 先用简短文本验证
可以上传 TXT、PDF 和 DOCX 文档。首次处理 PDF 或 DOCX 时，文档解析服务可能需要下载解析模型，耗时数分钟。上传和处理是两个独立操作，请先等待处理完成，再决定是否重新提交文件。
:::

### 围绕文档内容提问

1. 打开 **Workflows**，选择 **Document RAG Query**。
2. 确认 **COL** 为 `orion-notes`，**FLOW** 为 `default`。
3. 保持 **Timeline** 模式，输入以下问题，点击 **Query**：

   ```text
   Where is Orion laboratory located, and who directs it?
   ```

4. 检查回答是否指出所在地为 Suzhou、负责人为 Dr. Chen。查看时间线和来源链接，将检索内容与原始文档进行核对。

   ![TrustGraph 文档问答结果及处理时间线](/images/manual/use-cases/trustgraph-query-document.png#bordered)

   你的集合名称和回答措辞可能与截图不同。

**Simple** 只显示回答，**Timeline** 显示检索和生成回答的步骤，**Full DAG** 用可查看的图展示这些步骤。来源信息便于核验回答，但不能保证所有回答都正确。

### 探索知识图谱中的关系

1. 打开 **Workflows**，选择 **Graph Navigator**，保持相同集合。
2. 选择 **Full Explorer**，搜索 `Orion`，选择匹配的实体。
3. 选择节点，查看属性并浏览相连实体，寻找实验室的所在地和负责人关系。

   ![TrustGraph 图谱中的实体与关系](/images/manual/use-cases/trustgraph-explore-graph.png#bordered)

   图中显示的实体取决于集合中的文档。

4. 要根据图谱上下文提问，返回 **Workflows**，选择 **Graph RAG Query**，使用相同集合和流程提交同一个问题。

Document RAG 检索相关的原始文档文本，Graph RAG 检索从文档中提取并构建的图谱知识。具体关系名称和提取结果取决于聊天模型及源文档内容。

## 修改模型连接

要为所有使用 Router 默认配置的应用切换聊天或排序模型，可在 Router 的 **Default models** 页面修改 `default-chat` 或 `default-rerank`，并确保替换后的模型可调用。

需要修改模型 API URL 或 API Key 时，打开 **设置** > **应用** > **TrustGraph** > **管理环境变量**。修改 `LLM_BASE_URL` 或 `LLM_API_KEY`，然后点击 **确认** 和 **应用**。

同一个集合应保持向量模型一致。默认流程使用 768 维 EmbeddingGemma 向量。如果要尝试支持的 1024 维 Jina CLIP v2 模型，请创建独立流程和新集合，再重新处理文档，不要把不同维度的向量混入同一个集合。

## 常见问题

### 为什么文档上传成功后还无法回答问题？

确认已提交处理，聊天和向量模型可调用，查询时的集合和流程与处理时一致。首次处理 PDF 或 DOCX 时，应等待解析模型下载完成。可先用本案例的短 TXT 文件区分解析等待和模型连接问题。

### 为什么 Graph RAG 提示没有可用信息？

Graph RAG 需要所选集合中已有成功抽取的知识。文件上传成功、已提交处理，并不代表知识抽取成功。

1. 确认查询使用的集合和流程与处理时一致。
2. 确认聊天和向量模型可调用。在 Control Hub 中查看 `text-completion` 和 `triples` 的日志，检查抽取失败记录。输出多次达到 token 上限、JSONL 解析失败或字符串未结束，都可能意味着模型响应不完整。
3. 修复模型或应用配置后，将受影响文档重新提交处理。配置变更不会自动补回缺失的知识。

使用 Qwen 等推理模型时，较高的默认思考强度可能压缩用于输出抽取结果的容量，应选择适合结构化知识抽取的模型配置。当前应用包会明确请求低思考强度，并将抽取输出上限设为 8192 tokens。这个数值只是上限，不能保证任何文档都能完整输出；模型服务也可能有自己的限制。

:::details 管理员说明：知识抽取配置
当前版本的 TrustGraph OpenAI 处理器在启动配置中读取 `max_output`，不是 `max_output_tokens`。在 `openai` 模式下，`thinking: off` 表示不发送推理参数，并不显式关闭模型思考；`thinking: low` 会发送 `reasoning_effort: low`，模型服务需要支持这个参数。

这些设置由应用 Chart 提供，普通用户无需手动修改 ConfigMap。排查旧安装时，应确认已包含修正后的 Chart 配置，再重新提交文档处理。
:::

### 为什么模型调用提示认证错误？

在 Router 中检查 API Key 是否过期，以及是否允许调用相关模型。需要有效 Key 时，打开 **设置** > **应用** > **TrustGraph** > **管理环境变量**，更新 `LLM_API_KEY` 并应用变更。

### 上传失败时应该怎么处理？

刷新页面并重新登录。再次上传前，检查文档是否已经创建。持续的 WebSocket 连接失败可能中断上传，应检查应用入口和运行日志。

## 了解更多

- [Olares Router](olares-router.md)：管理可调用模型、默认模型和 API 凭据。
- [TrustGraph 文档](https://docs.trustgraph.ai/)：了解知识图谱、来源追踪及高级流程。
