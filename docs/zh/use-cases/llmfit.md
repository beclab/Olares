---
outline: [2, 3]
description: 使用 LLMFit 在 Olares 上找到最适合你硬件的 LLM 模型。它会检测你的系统并针对质量、速度、兼容性和上下文长度对模型进行评分。
head:
  - - meta
    - name: keywords
      content: Olares, LLMFit, LLM benchmark, hardware detection, GPU, model recommendation, self-hosted, AI
app_version: "1.0.3"
doc_version: "1.0"
doc_updated: "2026-04-02"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/llmfit.md)为准。
:::

# 使用 LLMFit 找到适合你硬件的最佳 LLM 模型

LLMFit 自动检测你系统的 RAM、CPU 和 GPU，然后推荐在你的硬件上运行良好的 LLM 模型。它在四个维度上对每个模型进行评分：质量、速度、兼容性和上下文长度，因此你可以快速查看哪些模型在你的设备上表现最佳。

:::warning 已知限制

当前版本的 LLMFit 有以下限制：
* **随机节点分配**：在多节点环境中，LLMFit 容器分配到的具体节点是随机的。
* **单 GPU 检测**：LLMFit 只能检测和评估其所在特定节点上分配的 GPU。
* **多 GPU 显示限制**：如果你的节点有多个 GPU，仪表板只显示其中一个（显示哪个不可预测）。API 正确检测多张卡，但它们的 VRAM 在界面中不会聚合或相加。
:::

## 安装 LLMFit

1. 打开 Market 并搜索 "LLMFit"。
   ![安装 LLMFit](/images/manual/use-cases/llmfit.png#bordered)

2. 点击 **获取**，然后点击 **安装**。等待安装完成。

## 使用 LLMFit

从 Launchpad 打开 LLMFit。仪表板显示：

- **系统摘要**：当前节点及其硬件详情，如 CPU、RAM 和 GPU。
- **模型适配探索器**：LLM 模型列表，包含估计的 TPS（每秒 token 数）以及质量、速度、适配和上下文分数。

使用这些分数来决定在你的 Olares 设备上下载和运行哪些模型。

![LLMFit 仪表板](/images/manual/use-cases/llmfit-dashboard.png#bordered)

## 常见问题

### 如何使用 LLMFit TUI？

LLMFit 使用其内置的 Web 仪表板作为主要界面，以简化操作。仪表板提供与 TUI 相同的功能。

如果你更喜欢基于终端的 TUI：
1. 打开 Control Hub，然后导航打开 LLMFit 容器终端。

   ![LLMFit 容器终端](/images/manual/use-cases/llmfit-terminal.png#bordered)

2. 运行以下命令：

   ```bash
   llmfit
   ```

   ![LLMFit TUI](/images/manual/use-cases/llmfit-tui.png#bordered)

## 了解更多

- [通过 Ollama 下载和运行本地 AI 模型](ollama.md)
- [设置 Open WebUI](openwebui.md)
