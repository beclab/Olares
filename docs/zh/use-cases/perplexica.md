---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/perplexica
outline: [2, 3]
description: 在 Olares 上自托管 Vane，通过 Olares Router 连接聊天模型，搭配内置嵌入模型和 SearXNG 网络搜索，获得带有来源引用的回答。
head:
  - - meta
    - name: keywords
      content: Olares, Vane, Perplexica, Perplexity 替代方案, 自托管, AI 搜索, SearXNG, 本地模型
app_version: "1.1.35"
doc_version: "1.3"
doc_updated: "2026-10-30"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/perplexica.md)为准。
:::

# 自托管私人 AI 搜索引擎 Vane

<VersionRouteSelect />

Vane（前身为 Perplexica）是一款开源 AI 问答引擎。它结合网络搜索与本地或云端大语言模型，在保护查询隐私的同时，提供带有来源引用的对话式回答。

本指南通过 Olares Router 使用 Qwen3.8-27B (llama.cpp) 生成回答，并让 Vane 直接连接 SearXNG 获取网络搜索结果。

## 学习目标

通过本指南，你将学会：

- 在 Vane 设置中或通过 Olares 设置中的应用环境变量连接 SearXNG。
- 在 Vane 中选择 Router 聊天模型和内置嵌入模型。
- 提出问题并查看回答引用的来源。

## 前提条件

开始前，你需要：

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- 已从应用市场安装并运行 [SearXNG](searxng.md)。
- 以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 聊天 | Qwen3.8-27B (llama.cpp) | 从应用市场安装 |
  | 嵌入 | all-MiniLM-L6-v2 | Vane 内置 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 安装 Vane

1. 打开应用市场，搜索“Vane”。
   ![Vane](/images/manual/use-cases/vane.png#bordered)

2. 点击**获取**，然后点击**安装**，等待安装完成。

## 配置模型

在 Vane 中连接 Olares Router 来使用聊天模型，并选择一个内置嵌入模型。

### 获取 Router 连接信息

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

继续之前，确认聊天模型在 Router 的 **LLM** 页面显示为 **Callable**。

### 在 Vane 中选择模型

1. 从启动台打开 Vane。首次打开时使用设置向导，已完成设置的用户可打开左下角的 **Settings** > **Models**。
2. 点击 **Add Connection**，选择 **OpenAI**，填写以下信息：

   | 字段 | 值 |
   | :--- | :--- |
   | **Connection Name** | `Olares Router` |
   | **API Key** | `olares` |
   | **Base URL** | 从 Router 复制的地址，保留末尾的 `/v1` |

3. 点击 **Add Connection**。在该连接的 **CHAT MODELS** 下，点击 **Add**。
4. 将 **Model name** 设为 `Router chat`，将 **Model key** 设为 `default-chat`，然后点击 **Add Model**。
5. 在设置向导中，点击 **Next** 进入 **Select models**。如果使用 **Settings** > **Models**，在同一页面找到 **Select models** 区域。
6. 在 **Select Chat Model** 下选择 **Olares Router - Router chat**，在 **Select Embedding Model** 下选择内置的 **Transformers - all-MiniLM-L6-v2**。
7. 如果使用设置向导，点击 **Finish**。

如果新添加的模型没有出现在 **Select Chat Model** 中，刷新页面，然后重新打开 **Settings** > **Models** 选择模型。

之后可在 Vane 的 **Settings** > **Models** 中修改模型或连接。

## 连接 SearXNG

先获取 SearXNG 地址，再选择以下一种方式配置。已完成初始设置时，可直接在 Vane 设置中填写地址。首次打开 Vane 之前，也可以通过环境变量初始化地址。

1. 打开 Olares **设置**，进入**应用** > **SearXNG** > **入口** > **SearXNG**。
2. 复制 **Endpoint** 地址，作为 SearXNG 实例地址，无需添加 `/search` 或查询参数。

### 方式一：在 Vane 中配置

1. 打开 Vane，进入左下角的 **Settings** > **Search**。
2. 将入口地址粘贴到 **SearXNG URL**。
3. 点击输入框外的区域保存。

### 方式二：通过环境变量配置

1. 打开 Olares **设置**，进入**应用** > **Vane** > **管理环境变量**。
2. 点击 `SEARXNG_API_URL` 右侧的编辑图标，粘贴入口地址，然后点击**确认**。
3. 点击**应用**。Vane 会自动重启，等待应用恢复运行后再打开。
4. 打开 Vane 的 **Settings** > **Search**，确认 **SearXNG URL** 显示刚才填写的地址。

:::info 已保存的搜索设置优先
只有在已保存的搜索地址为空时，Vane 才会使用 `SEARXNG_API_URL`。如果已有地址，请使用方式一修改。两种方式选择一种即可。
:::

:::info 可以使用 Router 的 default-search 吗？
Vane 当前直接连接 SearXNG 实例。**SearXNG URL** 应填写 SearXNG 入口地址，而不是 `default-search` 或 Router 的 Base URL。在 Router 中配置搜索工具不会让 Vane 自动连接该工具。通过 Router 使用搜索工具的应用示例，请参阅[使用 Lares 执行深度研究任务](lares.md#运行深度研究任务)。
:::

## 开始提问

1. 在 Vane 首页，打开输入框旁的来源选择器，启用 **Web**。
2. 输入问题，例如：

   ```text
   Olares 是什么？请找到它的官方网站和文档，说明它的用途，并附上来源链接。
   ```

3. 提交问题并等待回答。Vane 会在 **Sources** 下列出检索到的网页，并在回答中添加引用链接。
4. 打开引用链接，对照来源页面查看回答。可以在追问框中继续询问你感兴趣的细节。

   ![Vane 回答中的来源列表和引用链接](/images/manual/use-cases/vane-example-question.png#bordered)

## 了解更多

- [Olares Router](olares-router.md)：管理 Vane 使用的模型。
- [管理应用环境变量](../manual/olares/settings/manage-app-env.md)：修改应用配置并使其生效。
- [Vane GitHub 仓库](https://github.com/ItzCrazyKns/Vane)：查看上游项目介绍、架构说明及社区 Discord。
