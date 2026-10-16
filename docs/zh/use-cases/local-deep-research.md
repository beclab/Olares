---
outline: [2, 3]
description: 在 Olares 上使用 Local Deep Research，通过 AI Router 和 SearXNG 研究问题、核查引用来源，并将报告导出为 Markdown 或 PDF。
head:
  - - meta
    - name: keywords
      content: "Olares, Local Deep Research, AI 研究, AI Router, SearXNG, 来源引用, 研究报告, 本地模型"
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-16"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/local-deep-research.md)为准。
:::

# 使用 Local Deep Research 生成带来源引用的研究报告

Local Deep Research 是一个结合网页搜索与语言模型的开源 AI 研究助手。给它一个问题，它会生成后续问题、收集来源，并将发现整理成带引用的报告。

在 Olares 上，它通过 Router 调用模型，通过 SearXNG 搜索网页。本指南使用 Quick Summary 模式和一轮搜索研究热泵的工作原理，让你先看到完整流程，再开展范围更大的研究。

## 学习目标

通过本指南，你将学会：

- 准备模型和搜索连接。
- 创建用户账号并运行研究问题。
- 核查引用并导出报告。

## 前提条件

开始前，你需要：

- 一台运行 Olares 1.12.7 或更高版本、具有足够内存和磁盘空间的 Olares 设备。
- 能够访问 [Router](olares-router.md)，用于连接本地模型。
- 已从 Market 安装 [SearXNG](searxng.md)，且应用正在运行。
- 安装以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 聊天 | Qwen3.8-27B (llama.cpp) | 从 Market 安装 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 检查模型是否可用

Local Deep Research 通过 Router 的 `default-chat` 路由调用聊天模型。

1. 从启动台打开 Router，进入 **Default models** 页面。
2. 确认 Qwen3.8-27B (llama.cpp) 已设为默认聊天模型，且显示 **Callable**。

## 准备 SearXNG 连接

Local Deep Research 通过 SearXNG 搜索网页。安装 Local Deep Research 时需要填写 SearXNG 地址。

打开 Olares **设置**，进入 **应用** > **SearXNG** > **入口** > **SearXNG**，复制**端点**地址。

## 安装 Local Deep Research

1. 打开 Market，搜索“Local Deep Research”。
2. 点击 **获取**，再点击 **安装**。
3. 出现提示时，填入复制的 SearXNG 地址，点击 **确认**。
4. 等待安装完成。

## 创建用户账号

1. 从启动台打开 Local Deep Research。
2. 在登录页点击 **Register here**。
3. 输入用户名和密码，再次输入密码进行确认。
4. 阅读加密数据的提示，勾选确认框，点击 **Create Encrypted Account**。之后会自动登录。

:::warning 妥善保管密码
密码用于保护加密的账号数据库，该数据库不支持密码找回。保存重要研究或文档之前，请先安全保存密码。
:::

## 运行第一个研究问题

1. 在 **New Research** 页面，于 **Research Query** 中输入以下问题：

   ```text
   How do heat pumps work? Give a short explanation with source citations.
   ```

2. 选择 **Quick Summary**。

   :::tip 研究模式
   **Quick Summary** 返回初步回答，可作为快速配置检查。**Detailed Report** 运行按章节组织的较长研究。**Chat Mode** 是用于后续提问的实验性对话流程。先用 Quick Summary 确认所选模型和搜索引擎可用，再扩大研究范围。
   :::

3. 找到 **Advanced Options** 部分，配置以下设置：

   | 设置项 | 值 |
   | :--- | :--- |
   | <nobr>**Model Provider**</nobr> | OpenAI-Compatible Endpoint |
   | <nobr>**Language Model**</nobr> | `default-chat` |
   | <nobr>**Search Engine**</nobr> | SearXNG |
   | <nobr>**Search Iterations**</nobr> | `1` |
   | <nobr>**Questions Per Iteration**</nobr> | `1` |
   | <nobr>**Search Strategy**</nobr> | Source-Based |

   :::info
   你无法在此编辑模型或接口地址字段，因为它们由应用的环境变量决定。如需修改，请前往 **设置** > **应用** > **Local Deep Research** > **管理环境变量**。
   :::

4. 点击 **Start Research**。你可以跟踪研究进度，包括模型生成问题、搜索来源和准备报告的过程。
5. 处理完成后，点击 **View Results**。

## 核查并导出报告

1. 阅读报告，确认它回答了你的问题。
2. 打开引用，确认来源页面支持报告中的说法。
3. 如需保存副本，点击 **Export**，选择格式，例如 **Markdown**。
4. 如需便于阅读或分享的文档，点击 **Download PDF**。

   ![带引用和来源列表的研究报告](/images/manual/use-cases/local-deep-research-report.png#bordered)

## （可选）使用特定模型

默认情况下，Local Deep Research 通过 Router 的 `default-chat` 路由调用模型。如需使用特定的 Router 模型：

1. 在 Router 中进入 **LLM** 页面，找到目标模型，确认显示 **Callable**，然后点击 **View connection example** 图标。
2. 在 **How to call this model** 中复制完整标识符，包括服务商前缀。例如 `Olares/unsloth/Qwen3.6-27B-MTP-GGUF:UD-Q4_K_XL`。
3. 打开 Olares **设置** > **应用** > **Local Deep Research** > **管理环境变量**，将 `LDR_LLM_MODEL` 设为复制的标识符。
4. 点击 **应用**。应用会自动重启。
5. 重新打开 Local Deep Research 并登录。

## 常见问题

### 为什么出现 `401`、API key 无效或其他模型身份验证错误？

检查 Router 中的服务商凭据。使用默认 Olares Router 连接时，应用 key 留空是正确设置。必要时选择另一个可用的 Router 模型。

### 为什么提示找不到模型？

从 Router 复制完整的可调用标识符。不要假设本地模型的原始名称或其他实例的标识符可以直接使用。

### 为什么搜索没有返回来源？

检查 `SEARXNG_URL`，并在 SearXNG 中直接搜索同一问题。查看搜索引擎错误，如果上游引擎限流，稍后再试。

### 为什么进度显示完成，报告却包含错误？

阅读报告和研究日志，解决模型或搜索问题后重新运行。进度完成并不代表回答成功。

### 为什么研究表单中的模型字段是锁定状态？

在 Olares 设置中修改应用环境变量，然后等待设置生效。

### 为什么重启后要求重新登录？

使用你的用户账号登录。出现新的登录提示，本身不代表已保存的研究被删除。

## 了解更多

- [使用 Olares Router 作为 AI 网关](olares-router.md)：了解模型标识符、访问方式和服务商配置。
- [在 Olares 上运行本地大语言模型](llm-base-apps.md)：为 Router 配置本地模型。
