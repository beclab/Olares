---
outline: [2, 3]
description: 通过选择连接来源和 API 格式、复制 Base URL、填写模型名称和 API key，将 AI 客户端应用连接到模型服务。
head:
  - - meta
    - name: keywords
      content: Olares, AI 应用连接, 模型控制台, LLM, API 格式, Base URL, Ollama, OpenAI-Compatible
---

# 连接 AI 应用与模型服务 <Badge type="tip" text="^ 1.12.6" />

在 Olares 上，AI 服务应用通过 API 提供 AI 能力，AI 客户端应用（如 LobeHub）则提供你直接使用的聊天界面。要让它们协同工作，需要从服务应用收集连接信息，并在客户端应用中填写。

## 开始之前

- 安装 AI 服务应用和 AI 客户端应用。
- 对于 LLM 服务应用，从启动台打开应用，确认 **Model** 显示 **READY**、**Engine** 显示 **RUNNING**。

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

## 验证连接

在客户端应用中保存 provider 设置，并运行连接检查。例如，在 LobeHub 中，点击 **Model List** 旁边的 **Fetch models** 加载模型，启用该模型，然后在 **Connectivity Check** 右侧选择该模型并点击 **Check**。检查通过时，连接即建立成功。

:::warning 在 LobeHub 中禁用 Client Request Mode
不要在 LobeHub 中启用 **Use Client Request Mode**。启用后，应用会改为通过浏览器发起前端调用，可能触发跨域（CORS）限制或 Olares 安全认证提示。保持关闭可确保安全的后端到后端通信。
:::

## 常见连接错误快速判断

| 现象 | 可能原因与解决方法 |
|---|---|
| 客户端提示 “Model not found” | 模型名称被缩写或缺少前缀。请从模型控制台复制完整的模型名称。 |
| 连接检查失败，或 Base URL 无法访问 | 连接来源与客户端运行位置不匹配。重新打开模型控制台，选择匹配的**连接来源**后再次复制 Base URL。 |
| 客户端出现跨域（CORS）错误或认证提示 | 对于 LobeHub，请关闭 **Use Client Request Mode**，让请求直接在应用之间传递。 |

## 常见问题

### 如何连接非 AI 应用？

相同的内部入口模式也适用于非 AI 应用之间的连接。例如：

- *Arrs 媒体栈使用内部入口 URL 连接 Sonarr、Radarr、Prowlarr、Bazarr 和 qBittorrent。详见[使用 *Arrs 生态管理媒体库](/zh/use-cases/arrs.md)。
- SearXNG 本身不是 AI 模型，但它可以连接到 Vane 等 AI 客户端，用于私有增强搜索。详见[将 SearXNG 连接到 Vane](/zh/use-cases/perplexica.md)。

## 了解更多

- [AI 应用之间是如何连接的？](../help/usage.md#ai-应用之间是如何连接的)
- [使用引擎基座应用托管本地大语言模型](/zh/use-cases/llm-base-apps.md)
- [管理应用入口](../olares/settings/manage-entrance.md)
