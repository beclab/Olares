---
outline: [2, 3]
description: 了解如何在 Olares 上安装 Ollama，使用 CLI 管理模型，并将其配置为其他应用的中央 AI 服务。
head:
  - - meta
    - name: keywords
      content: Olares, Ollama, local LLM, AI model, CLI, self-hosted, open source models
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/ollama.md)。
:::

# 通过 Ollama 下载和运行本地 AI 模型

Ollama 是一个轻量级平台，允许你直接在机器上运行 `gemma3` 和 `deepseek-r1` 等开源 AI 模型。在 Olares 中，你可以将 Ollama 与 Open WebUI 等图形界面或其他智能体集成，以添加更多功能并简化交互。

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 上使用 Ollama CLI 管理本地 LLM。
- 将 Ollama 配置为同一本地网络上其他应用的 API 服务。

## 前提条件

在开始之前，请确保你具有 Olares 管理员权限。

## 安装 Ollama

1. 打开 **Market**，并搜索 "Ollama"。
2. 点击 **获取**，然后 **安装**，等待安装完成。
   ![安装 Ollama](/images/manual/use-cases/ollama.png#bordered)

## 使用 Ollama CLI 管理模型

Ollama CLI 允许你直接从 Olares 终端管理和与 AI 模型交互。以下是关键命令。

### 下载模型

:::tip 查看 Ollama 库

如果你不确定要下载哪个模型，请查看 [Ollama Library](https://ollama.com/library) 以探索可用模型。
:::

要下载模型，请使用以下命令：
```bash
ollama pull [model]
```

### 运行模型

:::tip
如果指定的模型尚未下载，`ollama run` 命令将自动在运行前下载它。
:::

要运行模型，请使用以下命令：
```bash
ollama run [model]
```

运行命令后，你可以直接在 CLI 中输入查询，模型将生成响应。

当你完成与模型的交互后，输入：
```bash
/bye
```
这将退出会话并返回到标准终端界面。

### 停止模型

要停止当前正在运行的模型，请使用以下命令：
```bash
ollama stop [model]
```

### 列出模型

要查看你已下载的所有模型，请使用：
```bash
ollama list
```

### 删除模型

如果你需要删除模型，可以使用以下命令：
```bash
ollama rm [model]
```

### 显示模型信息

要显示有关模型的详细信息，请使用：
```bash
ollama show [model]
```

### 列出正在运行的模型

要查看所有当前正在运行的模型，请使用：
```bash
ollama ps
```

## 配置 Ollama API 访问

要将 Ollama 用作同一本地网络上其他应用的后端，请确保 API 使用正确的访问设置，并从 Ollama API 页面复制端点。

1. 在 Olares 上，打开 **设置**，然后前往 **应用** > **Ollama**。
2. 在 **入口** 下，点击 **Ollama API**。
3. 确保 **认证级别** 设置为 **内部**。如果你更改了设置，请点击 **提交**。
4. 在 **端点设置** 下，复制 **端点** 中显示的 URL。在你的应用设置中使用此地址作为基础地址或主机。

   ![Ollama 端点](/images/manual/use-cases/lp-endpoint-entrances-ollama-api.png#bordered){width=90%}

:::tip 对于 OpenAI 兼容端点

某些应用期望 OpenAI 兼容的 API。如果标准端点失败，请尝试在你的地址后附加 `/v1`。例如：
```
https://a5be22681.{YOURUSERNAME}.olares.com/v1
```
:::

如果应用强制你输入 API 密钥，你通常可以输入任意字符串（例如，`ollama`）。

### 确保网络连接

因为你将认证设置为 "内部"，你的应用必须与 Olares 在同一本地网络上才能连接。

* **本地网络**：如果你的设备（或运行在 Olares 上的应用）在同一 Wi-Fi 或局域网上，你可以直接使用前端地址连接。
* **远程访问**：如果你正在远程访问 Olares，你必须在客户端设备上启用 [LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass)。这会创建一个安全隧道，有效地将你的设备放置在与 Olares 相同的网络上。

## 了解更多

- [使用 Open WebUI 运行 Ollama 模型](./openwebui.md)
- [将 Ollama 与 DeerFlow 集成](./deerflow.md)
