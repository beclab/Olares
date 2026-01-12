---
outline: [2, 3]
description: 查找关于 Olares 使用及社区应用的常见问题解答。
---
# 使用常见问题

本文汇总了关于 Olares 日常使用、应用及系统管理的常见问题。

## 应用

### Olares 支持运行哪些应用？

[Olares 应用市场](https://market.olares.com/)提供了 Ollama、ComfyUI 和 Open WebUI 等热门开源应用。

如果你有 Docker 使用经验，也可以在测试环境中[手动部署](../../developer/develop/tutorial/index.md)应用市场未收录的应用。

### 我可以在 Olares 设备上玩游戏吗？

可以。通过 Steam Headless 应用，你可以将 Olares 设备转变为游戏服务器。

* [**串流**](../../use-cases/stream-game.md)：你可以在 Olares 上本地运行游戏，并将其画面串流到手机或平板等设备上。
* [**直接游玩**](../../use-cases/play-games-directly.md)：你可以将显示器、键盘和鼠标直接连接到 Olares 设备，无需串流即可游玩。

### 如何在 Olares 中使用 Windows 环境？

你可以从应用市场运行 Windows 虚拟机，并使用任意标准 RDP 客户端进行连接。

### 可以在 Olares 上开发应用吗？

可以。安装 [Studio](../../developer/develop/tutorial/index.md) 后，你既可以直接在浏览器中编写代码，也可以将本地 VS Code 连接到设备。这能提供与本地机器相似的开发体验，同时利用服务器硬件更强大的算力。

## 存储
### 如果在运行中的 Olares 机器上添加新硬盘，系统会自动使用吗？

这取决于硬盘类型：
* **USB 驱动器**：是的，系统会自动挂载，并立即显示在文件管理器中。
* **内置硬盘**：不会。内置 HDD 或 SSD 不会自动加入存储池，你需要手动配置。
* **SMB 共享**：网络存储可以通过文件管理器中的**外部设备** > **连接服务器**进行添加。

详细步骤参见[扩展 Olares 存储](../best-practices/expand-storage-in-olares.md)。

## 多节点集群
### 如何向集群添加更多机器？

默认情况下，Olares 安装为单节点集群。不过，你可以将 Olares 安装为 Master 节点，然后添加 Worker 节点，从而构建可扩展的多节点集群。

注意，多节点目前为实验性功能，仅支持 Linux 系统。详细步骤参见[安装 Olares 多节点集群](../best-practices/install-olares-multi-node.md)。