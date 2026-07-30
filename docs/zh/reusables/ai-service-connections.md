# AI 服务连接

<!-- #region model-connection-overview -->
:::details 模型连接的工作原理
Olares 上的独立模型作为与客户端应用分开的服务运行。要连接两者，客户端需要准确的 **Model name**，以及与其所需 API 格式相匹配的 **Base URL**。

你可以从模型控制台获取这两个值。有关更多信息，可参阅[连接 AI 应用](/zh/manual/best-practices/connect-ai-apps.md)。
:::
<!-- #endregion model-connection-overview -->

<!-- #region get-model-connection-details -->
1. 从启动台打开模型应用。其模型控制台会自动打开。
2. 等待**模型**显示**就绪**，且**引擎**显示**运行中**。

   ![Qwen3.6-27B 模型控制台](/images/zh/manual/use-cases/qwen3.6-27b-model-console1.png#bordered)

3. 在**模型**部分，按显示内容原样复制**模型名称**。
4. 在**引擎**部分：

   a. **连接来源**：选择 **Olares 内应用**。

   b. **API 格式**：选择 **OpenAI-Compatible**。

   c.按显示内容原样复制 **Base URL** 地址。

<!-- #endregion get-model-connection-details -->

<!-- #region use-different-model -->
:::details 可选：使用其他模型
你也可以选择不同大小的模型，或改用其他模型提供商：

- 从应用市场安装其他模型应用。
- 通过[引擎基座应用](llm-base-apps.md)创建模型实例，使用自己的模型。

如果 AI Agent 应用已安装 [Olares CLI](../developer/cli-install.md) 和 [Agent Skills](../developer/cli-agent-skills.md)，可以让它部署模型并跳过手动设置。例如：

```plain
使用 Ollama 引擎基座在我的 Olares 上部署 qwen3.5:9b。
```
:::
<!-- #endregion use-different-model -->

<!-- #region app-endpoint-overview -->
:::details 应用端点（endpoint）如何工作
当客户端连接另一个 Olares 应用时，会使用该应用的端点作为网络地址。如果应用提供多个端点，请选择与客户端所需功能或协议相匹配的端点。
:::
<!-- #endregion app-endpoint-overview -->

<!-- #region get-model-connection-details-anthropic -->
1. 从启动台打开模型应用。其模型控制台会自动打开。
2. 等待**模型**显示**就绪**，且**引擎**显示**运行中**。

   ![Qwen3.6-27B 模型控制台](/images/zh/manual/use-cases/qwen3.6-27b-model-console-anthropic.png#bordered)

3. 在**模型**部分，按显示内容原样复制**模型名称**。
4. 在**引擎**部分：

   a. **连接来源**：选择 **Olares 内应用**。

   b. **API 格式**：选择 **Anthropic-Compatible**。

   c.按显示内容原样复制 **Base URL** 地址。

<!-- #endregion get-model-connection-details-anthropic -->
