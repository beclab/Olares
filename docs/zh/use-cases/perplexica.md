---
outline: [2, 3]
description: 在 Olares 上部署 Vane（前身为 Perplexica），实现私有的 AI 驱动搜索与问答。使用 Ollama 本地大模型和 SearXNG 作为搜索后端。
head:
  - - meta
    - name: keywords
      content: Olares, Vane, Perplexica, AI search, privacy, Ollama, SearXNG, self-hosted
app_version: "1.12.0"
doc_version: "1.2"
doc_updated: "2026-04-17"
---
# 配置注重隐私的 AI 搜索引擎：Vane

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/perplexica.md)。
:::

Vane（前身为 Perplexica）是一款开源的 AI 驱动问答引擎。它将网络搜索与本地或云端大语言模型（LLM）相结合，在保护你查询隐私的同时，提供带有引用来源的对话式回答。

本指南使用 Ollama 作为模型提供商，SearXNG 作为搜索后端。

## 前提条件

开始前，请确保：
- [Ollama 已安装](ollama.md) 并在你的 Olares 环境中运行。
- Ollama 中已安装至少一个聊天模型。嵌入模型为可选项，因为 Vane 内置了嵌入模型。

## 安装 SearXNG

SearXNG 是一款注重隐私的元搜索引擎，它聚合多个搜索引擎的结果，且不会追踪用户。Vane 通过它获取干净、无偏见的搜索结果，供 AI 模型处理。

1. 打开 Market，搜索 "SearXNG"。
   ![SearXNG](/images/manual/use-cases/perplexica-searxng.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 安装 Vane

1. 打开 Market，搜索 "Vane"。
   ![Vane](/images/manual/use-cases/vane.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 配置 Vane

1. 启动 Vane。首次启动时会打开设置向导，Ollama 及其已安装的模型将被自动检测到。
   ![Manage connections](/images/manual/use-cases/vane-manage-connections.png#bordered)

2. 点击 **下一步**。
3. 选择一个聊天模型和一个嵌入模型，然后点击 **完成**。
   ![Configure models](/images/manual/use-cases/vane-configure-models.png#bordered)

   :::tip 嵌入模型选项
   如果你在 Ollama 中没有嵌入模型，可以选择 Vane 内置的嵌入模型之一。
   :::

   你将进入主聊天页面。如需稍后更改模型或连接设置，点击左下角的 <i class="material-symbols-outlined">settings</i> 打开 **设置** 页面。

## 开始提问

尝试搜索一下，测试你的全新私有搜索环境。
![Vane example](/images/manual/use-cases/vane-example-question.png#bordered)

## 了解更多

- [Ollama](ollama.md)：在 Olares 上运行本地 LLM，作为 Vane 的模型后端。
- [Vane on GitHub](https://github.com/ItzCrazyKns/Vane)：上游项目 README、架构说明及社区 Discord。
