---
outline: [2, 3]
description: 在 Olares 上使用 TrustGraph 围绕文档内容提问。通过本地模型检索原文，并在知识图谱中探索实体之间的关系。
head:
  - - meta
    - name: keywords
      content: Olares, TrustGraph, 知识图谱, 文档问答, Document RAG, Graph RAG, 本地模型, 文档检索
app_version: "0.1.23"
doc_version: "1.3"
doc_updated: "2026-10-13"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/trustgraph.md)为准。
:::

# 使用 TrustGraph 将文档转为可检索的知识

TrustGraph 是一个开源平台，可以从文档中构建和查询知识图谱。你可以围绕文件内容提问，查看生成回答时的检索过程，并探索人物、地点和主题之间的关系。

在 Olares 上，TrustGraph 通过 Router 连接本地模型。本指南使用三份虚构的项目文档，查询迁移进展，并探索从正文中提取出的事件、人物及其关系。

## 前提条件

开始前，你需要：

- 一台运行 Olares 1.12.7 或更高版本的 amd64 或 arm64 设备，可为 TrustGraph 提供 10 GiB 内存和 20 GiB 磁盘空间。模型所需资源另行预留。
- 能够访问 [Router](olares-router.md)，用于连接本地模型。
- 从 Market 安装应用的管理员权限。
- 已从 Market 安装 MinIO。
- 从 Market 安装以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 聊天 | Qwen3.8-27B (llama.cpp) | 从 Market 安装 |
  | 嵌入 | EmbeddingGemma | 从 Market 安装 |
  | 重排序 | BGE-Reranker-v2-m3 | 从 Market 安装 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 检查模型是否可用

TrustGraph 需要聊天模型来提取知识和回答问题，嵌入模型来查找相关内容，重排序模型来对检索结果排序。安装前，这三个模型都应能通过 Router 正常调用。

1. 从启动台打开 Router，进入 **Default models** 页面。
2. 将 Qwen3.8-27B (llama.cpp) 设为默认聊天模型，将 EmbeddingGemma 设为默认嵌入模型，将 BGE-Reranker-v2-m3 设为默认重排序模型。如果已选中这三个模型，确认选择无误，并检查三者是否都显示 **Callable**。

## 安装 TrustGraph

1. 打开 Market，搜索“TrustGraph”。

2. 点击 **获取**，再点击 **安装**。按提示设置 `ADMIN_PASSWORD`，密码须为 12–128 个字符且不含换行，然后等待安装完成。

之后可以在 TrustGraph 中修改密码。

## 登录

1. 从启动台打开 TrustGraph。
2. 使用用户名 `admin` 和安装时设置的密码登录。

## 使用 TrustGraph

### 准备项目文档

假设你正在了解产品发布进展。一个共享的登录服务延期了，你想知道计划发生了什么变化，以及延期的原因。相关背景分别记录在发布计划、服务目录和迁移进展中。

你可以使用下面三份示例文件跟着操作，也可以换成自己的项目说明、服务文档或会议记录。建议选择涉及相同项目、服务、团队或人物的相关文档，方便探索它们之间的联系。

如果使用示例，请将以下文件保存到电脑。示例文件为英文，其中的项目、人物和日期均为虚构。

| 文档 | 内容 |
| :--- | :--- |
| [release-plan.txt](/examples/trustgraph/release-plan.txt) | 三个项目的发布日期、负责人和所需服务 |
| [service-directory.txt](/examples/trustgraph/service-directory.txt) | 服务之间的依赖关系、所属团队和技术负责人 |
| [migration-update.txt](/examples/trustgraph/migration-update.txt) | Identity API 调整后的迁移日期及团队协调安排 |

### 上传并处理文档

流程（flow）决定 TrustGraph 如何提取知识，集合（collection）将这些知识组织起来，供跨文档查询。以下步骤使用示例文件。如果使用自己的文档，请为文档填写便于识别的标题，并将相关文件放入同一个集合。

1. 在 **Welcome** 页面打开 **Document Ingestion**。点击 **+ Add Document**，选择 **File**，然后选择 `release-plan.txt`。
2. 将 **TITLE** 填为 `November release plan`，点击 **Upload**。
3. 选中上传的文档，点击 **Submit for Processing →**。
4. 选择 `default` 流程。点击 **+ Create New Collection**，在 **ID** 中填写 `release-planning`，在 **NAME** 中填写 `Release planning`，然后点击 **Next →**。
5. 在确认面板核对文档、流程和集合，点击 **Submit for Processing**。
6. 用相同方式上传 `service-directory.txt` 和 `migration-update.txt`，标题分别填写 `Service directory` 和 `Identity API migration update`。处理每个文档时，选择 `default` 流程和已有的 `release-planning` 集合。

TrustGraph 会在后台处理文档。每次提交后，请留出处理时间，再开始搜索或提问。你可以按下文步骤，在 Graph Navigator 中查找文档正文里的实体和关系。

:::tip 处理 PDF 和 DOCX 文件
TrustGraph 支持 TXT、PDF 和 DOCX 文件。首次处理 PDF 或 DOCX 时，需要下载解析模型，可能耗时数分钟。
:::

### 查询文档中的信息

先围绕迁移进展提出一个问题。

1. 打开 **Workflows**，选择 **Document RAG Query**。
2. 在右上角，将 **COL** 设为 `release-planning`，将 **FLOW** 设为 `default`。
3. 选择 **Timeline**，输入以下问题，点击 **Query**：

   ```text
   What is the new completion date for the Identity API migration, and why was it delayed?
   ```

4. 阅读回答，了解调整后的完成日期和延期原因。时间线中的 **Exploration** 显示检索到的文本片段数量和 ID。要将回答与原文对照，请打开电脑上的 `migration-update.txt`。

   ![Document RAG 回答 Identity API 迁移问题](/images/manual/use-cases/trustgraph-query-migration.png#bordered)

要尝试其他视图，请选择 **Simple** 查看回答，或选择 **Full DAG** 用图的形式查看检索和生成回答的步骤，然后重新提交查询。

### 探索实体及其关系

通过 Graph Navigator，了解迁移事件与延期原因、发布更新的人物之间的关系。

1. 打开 **Workflows**，选择 **Graph Navigator**，保持 **COL** 为 `release-planning`。
2. 选择 **Full Explorer**，点击左上方工具栏中的 **Search**，在搜索面板中搜索 `Identity API`。
3. 选择迁移事件，例如 `Identity API database migration`。标题为 `Identity API migration update` 的结果代表来源文档。
4. 点击相连节点，查看关系标签。例如，`is-delayed-by` 将迁移事件与数据同步问题相连，`announced-revised-finish-date-for` 将 Maya Patel 与迁移事件相连。点击节点可打开详情面板，继续探索其中的链接。

   ![迁移事件与延期原因、Maya Patel 之间的关系](/images/manual/use-cases/trustgraph-migration-relationships.png#bordered)

模型提取出的实体名称和关系标签可能有所不同。请根据图谱中实际出现的关系展开探索，并与迁移进展原文对照。

### 使用 Graph RAG 查询关系

Document RAG 检索文本片段，Graph RAG 检索提取出的实体与关系。可以对照两种模式中的检索事件，了解它们如何为回答查找信息。

先围绕在 Graph Navigator 中找到的一条关系提问。

1. 返回 **Workflows**，选择 **Graph RAG Query**。
2. 保持 **COL** 为 `release-planning`，**FLOW** 为 `default`。选择 **Split + Graph** 视图。
3. 输入以下问题，点击 **Query**：

   ```text
   What caused the Identity API database migration to be delayed?
   ```

4. 阅读回答，探索回答中用于解释延期原因的关系。查看 **Exploration** 和 **Focus** 事件，再打开电脑上的 `migration-update.txt`，将回答与原文对照。

   ![Graph RAG 检索迁移事件关系及来源上下文](/images/manual/use-cases/trustgraph-query-migration-relationship.png#bordered)

要用其他视图查看同一个问题，请先切换视图，再重新提交查询。

如果刚提交文档，请等后台处理完成后，再次查询。如果回答仍然没有相关关系作为依据，请参阅[为什么 Graph RAG 提示没有相关信息？](#为什么-graph-rag-提示没有相关信息)。

## 修改模型连接

要切换聊天或重排序模型，打开 Router 的 **Default models** 页面，修改 `default-chat` 或 `default-rerank`。其他使用该默认配置的应用也会随之切换模型。对于已有集合，请保持嵌入模型不变。

要连接其他模型 API，打开 **设置** > **应用** > **TrustGraph** > **管理环境变量**：

1. 将 `LLM_BASE_URL` 更新为新的 API URL。
2. 如果该 API 需要密钥，更新 `LLM_API_KEY`。
3. 点击 **确认**，再点击 **应用**。

## 常见问题

### 为什么上传文档后还无法获得回答？

如果只上传了文件，请在 **Document Ingestion** 中选中文档，点击 **Submit for Processing**。等待后台处理后，在查询页面选择处理时使用的集合和流程。如果仍然没有可检索的内容，请检查聊天和嵌入模型是否正在运行，并可通过 Router 调用。

### 为什么图谱中只有文档标题？

文档标题节点以及指向 `Document` 或 `Entity` 的 `type` 关系描述的是文件本身，还没有展示正文中的项目、服务和依赖关系。

在 **Document Ingestion** 中，检查三份文件是否都已提交到 `default` 流程和 `release-planning` 集合，并在 Router 中检查处理期间是否有失败的模型调用。如果处理时模型不可用导致失败，请先恢复模型连接，再重新提交受影响的文档进行处理。重试 Graph RAG 前，通过 Graph Navigator 的 **Search** 按钮查找正文中的实体。

### 为什么 Graph RAG 提示没有相关信息？

Graph RAG 需要使用从文档中提取的实体和关系。如果 **Exploration** 显示 `0 edges`，说明本次查询没有找到可用于回答的关系。

1. 提交文档后，先等待后台处理，再次查询。Graph Navigator 中已经出现的关系，可能还需要一段时间才能被 Graph RAG 检索到。
2. 确认查询使用的集合和流程与提交文档时一致。
3. 如果处理失败，请检查聊天和嵌入模型是否可通过 Router 调用，并解决模型或应用配置问题。然后在 **Document Ingestion** 中选中已有文档，点击 **Submit for Processing**，提交到相同的流程和集合。等待处理后，再次提问，无需重新上传文件。

如果处理仍然失败，请检查是否有应用更新，或在 Control Hub 中查看文档处理日志。

使用 Qwen 等推理模型时，较高的默认思考强度可能压缩用于输出抽取结果的容量，应选择适合结构化知识抽取的模型配置。当前应用包在知识抽取和查询生成阶段都会明确请求低思考强度，并将输出上限设为 8192 tokens。这个数值只是上限，不能保证任何文档都能完整输出；模型服务也可能有自己的限制。

如果 Graph Navigator 中已有正确关系，但 Graph RAG 仍无法回答，可能是检索没有命中目标实体。名称相似、仅编号不同的实体尤其容易出现语义检索偏差。可改用 Document RAG 查找原始段落，或在 Graph Navigator 中直接查看实体。应核对回答引用的内容是否属于目标实体；根据其他实体的命名规律猜出的合理答案，不属于已验证结果。

### 需要调整 thinking 或 temperature 吗？

本指南的操作可以先使用应用提供的设置。更换模型时，请参考模型提供方对推理强度和温度的建议。适用于某个模型的设置，不一定适用于其他模型。

如果文档处理持续失败，请检查模型支持的参数和输出上限。解决问题后，重新提交受影响的文档。

### 为什么自定义模型 API 提示认证错误？

检查 `LLM_BASE_URL` 对应 API 的凭据要求。如果使用 Router 签发的 API Key，请在 Router 中检查其权限和有效期。然后在 **设置** > **应用** > **TrustGraph** > **管理环境变量** 中更新 `LLM_API_KEY`，点击 **确认** 和 **应用**。

### 上传失败时应该怎么处理？

刷新页面并重新登录。再次上传前，先在 **Document Ingestion** 中查找文件，避免重复上传。

### 为什么 Agent 查询不到其他集合中的文档？

默认的 `knowledge-query` 工具指向 `default` 集合。当前版本不会自动让它跟随 Agent 页面或请求中选择的集合，因此切换顶部 **COL** 不一定会改变这个工具的查询范围。查询其他集合时，优先使用 Document RAG 或 Graph RAG，并核对所选集合。不要因为 Agent 返回“没有信息”就认为文档没有入库。

## 了解更多

- [Olares Router](olares-router.md)：管理模型、默认配置和 API 凭据。
- [TrustGraph 文档](https://docs.trustgraph.ai/)：了解知识图谱及高级流程。
