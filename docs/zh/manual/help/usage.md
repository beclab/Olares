---
outline: [2, 3]
description: 查找关于 Olares 日常使用、应用及系统管理的常见问题解答。
head:
  - - meta
    - name: keywords
      content: Olares, 使用问题, 应用市场, Ollama, ComfyUI, 存储, 多节点集群
---

# 使用常见问题

本文汇总了关于 Olares 日常使用、应用及系统管理的常见问题。

## 应用

### Olares 支持运行哪些应用？

[Olares 应用市场](https://market.olares.com/)提供了 Ollama、ComfyUI 和 Open WebUI 等热门开源应用。

如果你有 Docker 使用经验，也可以将应用市场未收录的应用打包为 [Olares 应用 Chart](../../developer/develop/package/chart.md)，并在自己的 Olares 设备上测试。

### 可以在 Olares 设备上玩游戏吗？

可以。安装 Steam Headless 应用，就能把 Olares 设备变成游戏服务器。

* [**串流**](../../use-cases/steam-stream.md)：你可以在 Olares 上本地运行游戏，并将其画面串流到手机或平板等设备上玩。
* [**直连**](../../use-cases/steam-direct-play.md)：你可以将显示器、键盘和鼠标直接连接到 Olares 设备，无需串流即可玩游戏。

### 如何在 Olares 中使用 Windows 环境？

从 Olares 应用市场安装并运行 [Windows 虚拟机](/zh/use-cases/windows.md)，然后使用任意标准 RDP 客户端连接访问。

### 可以在 Olares 上开发应用吗？

可以。你可以使用自己熟悉的本地开发工具构建应用，将其打包为 [Olares 应用 Chart](../../developer/develop/package/chart.md)，并在提交到应用市场前先在 Olares 设备上测试。

### AI 应用之间是如何连接的？

在 Olares 上使用 AI 时，你通常需要同时操作两个应用：一个 AI 服务应用在后台提供 AI 能力，另一个 AI 客户端应用则提供你直接交互的聊天界面。

- **AI 客户端应用**：提供你直接交互的前端聊天界面或工作流画布，如 LobeHub。它们依赖 AI 服务应用（通常称为 **provider**）来执行生成文本、提取数据等 AI 任务。
- **AI 服务应用**：通过 API 为兼容的客户端提供聊天、搜索、语音识别等 AI 能力。部分应用自带管理 Web 界面，另一些主要作为无界面的后端服务运行。

    在 Olares 上，AI 服务应用分为两类：
    - **LLM 服务应用**：托管大型语言模型（LLM），用于文本生成、代码补全和聊天。包括八个预构建模型应用，以及在[引擎基座应用](/zh/use-cases/llm-base-apps.md)上创建的自定义模型实例。
    - **其他 AI 服务应用**：提供 LLM 之外的功能，例如语音识别（Speaches）和文本提取（PaddleOCR）。

由于不同 provider 使用不同的通信规则，它们依赖特定的 **API 格式**，你可以把这些格式理解为应用之间交流的“语言”。最常见的两种格式是 **OpenAI-Compatible** 和 **Ollama**。连接时，需要在客户端应用中填写服务应用的 Base URL、模型名称和 API key。具体步骤请参考[连接 AI 应用与模型服务](../best-practices/connect-ai-apps.md)。

### 可以手动更新应用版本吗？

:::tip 重要说明
我们建议始终通过应用市场更新应用，以确保稳定性和兼容性。
:::

可以。某些情况下，应用内部可能提示有新版本可用，但该版本尚未正式在应用市场上架。如果急需使用最新功能，可以在控制面板中[手动更新应用镜像](../update-app-image.md)。

请注意，控制面板中的手动编辑都是临时改动，会在下次应用市场更新时被覆盖；手动更新后应用也可能因兼容性问题无法启动。操作前请先阅读[手动更新应用镜像](../update-app-image.md)中的警告说明。

## 存储## 存储

### 如果在运行中的 Olares 机器上添加新硬盘，系统会自动使用吗？

这取决于硬盘类型：
* **USB 驱动器**：会。系统会自动挂载，并立即在文件管理器中显示。
* **内置硬盘**：不会。内置 HDD 或 SSD 需要手动配置才能加入存储池。
* **SMB 共享**：需要手动添加网络存储。可以通过文件管理器中的**外部设备** > **连接服务器**添加。

详细步骤参见[在 Olares 中扩展存储空间](../best-practices/expand-storage-in-olares.md)。

## 多节点集群

### 如何向集群添加更多机器？

默认情况下，Olares 安装为单节点集群。要创建可扩展的多节点集群，需要先将 Olares 安装为 master 节点，然后再添加 worker 节点。

注意，多节点目前为实验性功能，仅支持 Linux 系统。详细步骤参见[安装多节点 Olares 集群](../best-practices/install-olares-multi-node.md)。
