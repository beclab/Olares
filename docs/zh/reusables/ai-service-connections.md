# AI 服务连接

<!-- #region router-prerequisite -->
- 一台运行 Olares 1.12.7 或更高版本、具有足够磁盘空间和内存的 Olares 设备。
- 使用本地模型时，能够访问 Router。
<!-- #endregion router-prerequisite -->

<!-- #region model-connection-overview -->
:::details 模型连接的工作原理
在 Olares 1.12.7 及以上版本中，AI 客户端通过 Olares Router 连接模型。Router 提供 Base URL，并将 `default-chat` 的请求转发给**默认模型（Default models）**页面中设置的聊天模型。

本教程使用 Qwen3.8-27B (llama.cpp) 作为默认聊天模型。连接格式和 API 密钥要求详见[连接 AI 应用](/zh/manual/best-practices/connect-ai-apps.md)。
:::
<!-- #endregion model-connection-overview -->

<!-- #region get-model-connection-details -->
1. 从启动台打开 Router。在 **Default models** 页面将 Qwen3.8-27B (llama.cpp) 设为默认聊天模型。
2. 进入 **LLM** 页面，找到 Qwen3.8-27B (llama.cpp)，点击模型所在行的 **View connection example** 图标。

   ![在 Router 中查看 Qwen3.8-27B 的连接示例](/images/manual/use-cases/router-view-connection-examp.png#bordered)

3. 在 **How to call this model** 窗口中，选择 **Apps in Olares**，复制完整的 **Base URL**，保留末尾的 `/v1`。

   ![复制 Olares 内应用使用的 Router Base URL](/images/manual/use-cases/router-how-to-call-model.png#bordered)

4. 在客户端中将模型名称填写为 `default-chat`。Olares 内的应用无需 Router API 密钥；允许留空时留空，必填时可填写 `olares`。

   `default-chat` 是路由名称，不会出现在模型列表 API 的返回结果中。如果客户端自动获取模型列表，需要手动添加它。如果客户端只允许选择列表中的模型，请改用 Router 中显示的完整模型名称。
<!-- #endregion get-model-connection-details -->

<!-- #region use-different-model -->
:::details 可选：使用其他模型
从应用市场安装其他聊天模型，或通过[引擎基座应用](llm-base-apps.md)创建模型实例，然后在 Router 的 **Default models** 页面将其设为默认聊天模型。使用 `default-chat` 的客户端会调用新的默认模型，无需修改客户端的模型设置。

如果希望某个客户端始终使用指定模型，可改填 Router 的 **How to call this model** 窗口中显示的完整模型名称。
:::
<!-- #endregion use-different-model -->

<!-- #region app-endpoint-overview -->
:::details 应用端点（endpoint）如何工作
当客户端连接另一个 Olares 应用时，会使用该应用的端点作为网络地址。如果应用提供多个端点，请选择与客户端所需功能或协议相匹配的端点。
:::
<!-- #endregion app-endpoint-overview -->

<!-- #region get-model-connection-details-anthropic -->
1. 从启动台打开 Router。在 **Default models** 页面将 Qwen3.8-27B (llama.cpp) 设为默认聊天模型。
2. 进入 **LLM** 页面，找到 Qwen3.8-27B (llama.cpp)，点击模型所在行的 **View connection example** 图标。

   ![在 Router 中查看 Qwen3.8-27B 的连接示例](/images/manual/use-cases/router-view-connection-examp.png#bordered)

3. 在 **How to call this model** 窗口中，选择 **Apps in Olares**，复制 **Base URL**，然后去掉末尾的 `/v1`，供 Anthropic 兼容客户端使用。例如填写 `https://router.<你的 Olares 域名>`；客户端会自动追加 `/v1/messages`。

   ![复制 Olares 内应用使用的 Router Base URL](/images/manual/use-cases/router-how-to-call-model.png#bordered)

4. 在客户端中将模型名称填写为 `default-chat`。Olares 内的应用无需 Router API 密钥；允许留空时留空，必填时可填写 `olares`。

   `default-chat` 是路由名称，不会出现在模型列表 API 的返回结果中。如果客户端自动获取模型列表，需要手动添加它。如果客户端只允许选择列表中的模型，请改用 Router 中显示的完整模型名称。
<!-- #endregion get-model-connection-details-anthropic -->

<!-- #region get-embedding-model-connection-details-openai -->
1. 从启动台打开 Router，进入 **Tools** 页面，找到已安装的嵌入模型，等待其显示 **Callable**。
2. 点击模型所在行的 **View connection example** 图标。
3. 在 **How to call this model** 窗口中，选择 **Apps in Olares**，复制完整的 **Base URL**，保留末尾的 `/v1`。
4. 从此窗口复制完整的 **Model name**，包括 `Olares/` 前缀，填入客户端的嵌入模型设置。Olares 内的应用无需 Router API 密钥；仅在客户端要求必填时填写 `olares`。

这里应填写嵌入模型名称，不能使用 `default-chat`。查询已有知识库时应使用原来的嵌入模型；更换模型可能需要重新索引文档。
<!-- #endregion get-embedding-model-connection-details-openai -->
