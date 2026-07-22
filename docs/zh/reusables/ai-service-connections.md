# AI 服务连接

<!-- #region model-connection-overview -->
:::info 模型连接的工作原理
Olares 上的独立模型作为与客户端应用分开的服务运行。要连接两者，客户端通常需要准确的 **Model name**，以及与其支持的 API 格式相匹配的 **Base URL**。

你可以从模型控制台获取这些值。有关更详细的说明，请参阅[连接 AI 应用](/zh/manual/best-practices/connect-ai-apps.md)。
:::
<!-- #endregion model-connection-overview -->

<!-- #region get-model-connection-details -->
1. 从启动台打开模型应用。模型控制台会自动打开。
2. 如果模型开始下载，请等待 **Model** 显示 **READY**，且 **Engine** 显示 **RUNNING**。
3. 在 **Service status** 下，选择 **Apps in Olares**。找到客户端支持的 API 格式，然后按显示内容原样复制 **Model name** 和对应的 **Base URL**。
<!-- #endregion get-model-connection-details -->

<!-- #region app-endpoint-overview -->
:::info 应用端点的工作原理
当客户端连接另一个 Olares 应用时，需要使用该应用的端点作为网络地址。如果应用公开了多个端点，请选择与客户端所需功能或协议相匹配的端点。
:::
<!-- #endregion app-endpoint-overview -->
