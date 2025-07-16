---
outline: [2, 3]
description: 了解如何在 Olares 环境中使用 Ollama CLI 在本地下载和管理 AI 模型。
---

# 通过 Ollama 在本地运行 AI 模型
Ollama 是一个轻量级平台，可以让你在本地机器上直接运行 `deepseek-r1` 和 `gemma3` 等开源 AI 模型。如果你更偏向于使用图形化界面，也可以在 Open WebUI 中管理 Ollama 模型，以添加更多功能并简化交互。

本文档将介绍如何在 Olares 上设置和使用 Ollama CLI。

## 开始之前
请确保满足以下条件：
- 当前登录账号为 Olares 管理员。
- 确保设备已安装 LarePass 并[启用专用网络](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。

## 安装 Ollama

直接从应用市场安装 Ollama。

安装完成后，可以从启动台访问 Ollama 终端。

![Ollama](/images/manual/use-cases/ollama.png#bordered)
## Ollama CLI
Ollama CLI 让你可以直接管理和使用 AI 模型。以下是主要命令及其用法：

### 下载模型
:::tip 查看 Ollama 模型库
如果不确定下载哪个模型，可以在 [Ollama Library](https://ollama.com/library) 浏览可用模型。
:::
使用以下命令下载模型：
```bash
ollama pull [model]
```

### 运行模型
:::tip
如果尚未下载指定的模型，ollama run 命令会在运行前自动下载该模型。
:::

使用以下命令运行模型：
```bash
ollama run [model]
```

运行命令后，直接在 CLI 中输入问题，模型会生成回答。

完成交互后，输入：
```bash
/bye
```
这将退出会话并返回标准终端界面。

### 停止模型
要停止当前运行的模型，使用以下命令：
```bash
ollama stop [model]
```

### 列出模型
要查看系统中已安装的所有模型，使用：
```bash
ollama list
```

### 删除模型
如果需要删除模型，可以使用以下命令：
```bash
ollama rm [model]
```
### 显示模型信息
要显示模型的详细信息，使用：
```bash
ollama show [model]
```

### 列出运行中的模型
要查看所有当前运行的模型，使用：
```bash
ollama ps
```

## 了解更多
- [了解如何通过 Open WebUI 运行 Ollama 模型](openwebui.md)