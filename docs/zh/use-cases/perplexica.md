---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/perplexica
outline: [2, 3]
description: 在 Olares 上自托管 Vane（前身为 Perplexica），作为私有的 Perplexity 替代方案。通过 Olares Router 连接模型，结合 SearXNG 搜索后端，获得带引用的 AI 回答。
head:
  - - meta
    - name: keywords
      content: Olares, Vane, Perplexica, perplexity alternative, self-hosted perplexity, AI search, SearXNG, vane on olares
app_version: "1.12.0"
doc_version: "1.2"
doc_updated: "2026-09-23"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/perplexica.md)为准。
:::

# 自托管私人 AI 搜索引擎 Vane

<VersionRouteSelect />

Vane（前身为 Perplexica）是一款开源的 AI 驱动问答引擎。它将网络搜索与本地或云端大语言模型（LLM）相结合，在保护你查询隐私的同时，提供带有引用来源的对话式回答。

本指南通过 Olares Router 使用 Qwen3.8-27B (llama.cpp)，并以 SearXNG 为搜索后端。

## 前提条件

开始前，你需要：

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- 从应用市场安装 Qwen3.8-27B (llama.cpp)。Vane 可使用内置嵌入模型。

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

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

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

1. 打开 Vane，在首次设置向导或 **Settings** 的模型设置中添加 **OpenAI** 提供商。
2. **Base URL** 填写从 Router 复制的地址，保留 `/v1`。**API Key** 填写 `olares`。

   <!--
   TODO: 素材清单 17，待补 Router 截图：vane-provider-router.png；替换下方旧图后再取消注释。
   ![Manage connections](/images/manual/use-cases/vane-manage-connections.png#bordered)
   -->

3. 在该提供商下手动添加聊天模型，名称可填 `Router chat`，模型 ID 填写 `default-chat`。
4. 选择该聊天模型和一个 Vane 内置嵌入模型，完成设置。

   <!--
   TODO: 素材清单 18，待补 Router 截图：vane-models-router.png；替换下方旧图后再取消注释。
   ![Configure models](/images/manual/use-cases/vane-configure-models.png#bordered)
   -->


之后可通过左下角的设置图标修改模型和连接。

## 开始提问

尝试搜索一下，测试你的全新私有搜索环境。
![Vane example](/images/manual/use-cases/vane-example-question.png#bordered)

## 了解更多

- [Olares Router](olares-router.md)：管理 Vane 使用的模型。
- [Vane on GitHub](https://github.com/ItzCrazyKns/Vane)：上游项目 README、架构说明及社区 Discord。
