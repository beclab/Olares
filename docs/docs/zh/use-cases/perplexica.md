---
description: Perplexica 应用配置指南，实现高性能 AI 对话服务，支持个性化调整和数据隐私保护。
---
# Perplexica 本地 AI 搜索

Perplexica 是一个开源的 AI 搜索引擎，在提供智能搜索功能的同时注重保护用户隐私。作为 Perplexity AI 的替代方案，它将先进的机器学习技术与全面的网络搜索功能相结合，为你的查询提供准确且附带来源引用的答案。

## 后端引擎：SearXNG
SearXNG 作为 Perplexica 的后端，是一个注重隐私的元搜索引擎。它具有以下特点：
* 聚合多个搜索引擎的结果
* 去除跟踪代码，保护你的隐私
* 为 AI 模型提供干净、无偏见的搜索结果

通过这样的集成，Perplexica 既能作为完整的搜索解决方案，又能确保敏感信息的安全。

## 开始之前
在开始使用前，请确保：
- Olares 环境中已安装并运行 Ollama
- Open WebUI 已安装，并已下载所需的语言模型
  :::tip
  建议使用 `gemma2` 等轻量但功能强大的模型，可在速度和性能间取得良好平衡。
  :::
- 确保访问设备已安装 LarePass 并[启用专用网络](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。

## 配置 Perplexica
1. 管理员从应用市场安装 SearXNG。
   
   :::info
  从 Olares 1.11.6 开始，如果已安装 "SearXNG For Cluster" 或 "SearXNG" 客户端入口，需先卸载这些版本。
  :::

2. 安装 Perplexica。
3. 启动 Perplexica，点击左下角的 <i class="material-symbols-outlined">settings</i> 打开设置界面。
4. 配置搜索环境。以 `gemma2` 为例：
    - **Chat model provider**：`Ollama`
    - **Chat model**：`gemma2:latest`
    - **Embedding model provider**：`Ollama`
    - **Embedding model**：`gemma2:latest`

    ![Perplexica 配置](/images/manual/use-cases/perplexica-configurations.png#bordered){width=50%}
5. 点击底部的 <i class="material-symbols-outlined">cloud_upload</i> 确认按钮保存配置并返回搜索界面。

至此配置完成。你可以搜索感兴趣的主题来测试新的搜索环境。

![Perplexica 示例](/images/manual/use-cases/perplexica-example-question.png#bordered)
