---
outline: [2, 3]
description: 通过选择连接来源和 API 格式、复制 Base URL、填写模型名称和 API key，将 AI 客户端应用连接到模型服务。
head:
  - - meta
    - name: keywords
      content: Olares, AI 应用连接, 模型控制台, LLM, API 格式, Base URL, Ollama, OpenAI-Compatible
---

# 连接 AI 应用与模型服务 <Badge type="tip" text="^ 1.12.6" />

在 Olares 上，AI 服务应用通过 API 提供 AI 能力，客户端应用则提供实际使用的界面或工作流。不同应用的连接方式遵循同一套逻辑：选择客户端访问服务的位置、匹配 API 格式，并复制服务地址和模型名称。

本文只介绍这套通用方法。不同客户端中的具体字段和按钮，请参考文末对应的应用教程。

## 开始之前

- 安装 AI 服务应用和 AI 客户端应用。
- 对于 LLM 服务应用，从启动台打开应用，确认 **Model** 显示 **Ready**、**Engine** 显示 **Running**。

## 选择连接来源和 API 格式

从启动台打开 LLM 服务应用，启动其**模型控制台**，然后选择与客户端应用匹配的选项：

- **连接来源（Connection source）**：选择与客户端应用运行位置匹配的选项。例如，客户端安装在同一个 Olares 集群中时，选择 **Apps in Olares**。
- **API 格式**：选择客户端应用支持的格式，例如 **OpenAI-Compatible** 或 **Ollama**。模型控制台会显示与所选格式对应的 Base URL。

:::info
PaddleOCR 等非 LLM 服务不使用这些通用格式。它们使用自己工具专属的协议进行通信，因此无需为它们配置 provider 格式。
:::

## 复制 Base URL

Base URL 是服务应用接收并处理请求的地址。

- **对于 LLM 服务应用**：复制模型控制台中显示的 **Base URL**。请原样复制，包括 `/v1` 等路径后缀。
- **对于其他 AI 服务应用**：打开 Olares **设置**，进入 **应用** > **[应用名称]** > **入口**，然后复制 **Endpoint URL**。请确保入口的**认证级别**设置为 **Internal**，以便其他应用无需登录即可访问。

    :::tip 多个入口
    部分应用会暴露多个入口。请根据客户端的协议或使用场景选择对应的入口。例如，网页访问使用主入口，程序化集成使用专用 API 入口。
    :::

## 填写模型名称和 API key

- **模型名称**：按照模型控制台中的显示原样复制**模型名称**。不要缩写，也不要删除仓库前缀（如 `unsloth/`）或量化标签（如 `UD-Q4_K_XL`），否则客户端可能会返回类似 “Model not found” 的错误。
- **API key**：部署在 Olares 本地的 AI 服务应用信任来自同一集群中其他应用的请求，因此通常不需要真实的 API key。如果客户端应用仍要求该字段有值，可以输入任意占位文本，例如 `olares` 或 `local`。

## 在客户端中添加服务

打开客户端应用的模型、Provider 或集成设置，然后填写从服务应用取得的信息：

| 客户端设置 | 填写内容 |
|---|---|
| Provider 或 API 格式 | 模型控制台中选择的格式，例如 **OpenAI-Compatible** 或 **Ollama** |
| Base URL 或 Endpoint | 从模型控制台或应用入口复制的完整 URL |
| 模型名称或 Model ID | 模型控制台中显示的完整模型名称 |
| API key | 服务要求的真实密钥；如果本地服务不需要密钥，但客户端不允许留空，则填写占位值 |

不同客户端使用的字段名称可能不同。如果客户端要求其他参数，或允许修改请求的发起位置，请按照该客户端的教程配置，不要直接猜测。

## 验证连接

保存 Provider 设置，然后运行客户端的连接检查或刷新模型列表。如果客户端没有这两个功能，可以新建会话，选择刚配置的模型并发送一条简短请求。收到回复说明客户端可以访问服务并调用该模型。

## 常见连接错误快速判断

| 现象 | 可能原因与解决方法 |
|---|---|
| 客户端提示 “Model not found” | 模型名称被缩写或缺少前缀。请从模型控制台复制完整的模型名称。 |
| 连接检查失败，或 Base URL 无法访问 | 连接来源与客户端运行位置不匹配。重新打开模型控制台，选择匹配的**连接来源**后再次复制 Base URL。 |
| 浏览器提示跨域（CORS）错误，或出现 Olares 身份认证页面 | 客户端可能从浏览器前端发送了请求。请查看对应客户端教程，确认请求模式和服务入口。 |

## 应用教程

- [使用 LobeHub 构建本地 AI Agent](/zh/use-cases/lobechat.md)
- [使用 Open WebUI 搭建本地 AI 对话界面](/zh/use-cases/openwebui.md)
- [使用 Dify 定制本地 AI 助手](/zh/use-cases/dify.md)

## 了解更多

- [AI 应用之间是如何连接的？](../help/usage.md#ai-应用之间是如何连接的)
- [使用 Ollama、vLLM、llama.cpp 和 SGLang 运行本地大模型](/zh/use-cases/llm-base-apps.md)
- [管理应用入口](../olares/settings/manage-entrance.md)
