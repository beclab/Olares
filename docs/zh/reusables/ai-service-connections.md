# AI 服务连接

<!-- #region router-openai-base-url -->
复制浏览器中的 Router 地址，只保留协议和域名，再加上 `/v1`。将下方示例中的域名替换为你自己的域名：

   ```text
   https://router.laresprime.olares.com/v1
   ```
<!-- #endregion router-openai-base-url -->

<!-- #region model-context-window -->
1. 在 Router 中打开 **LLM** 页面，找到模型，点击所在行最右侧的信息图标，打开 **Model card**。

   ![从 Router 的 LLM 列表打开模型卡片](/images/manual/use-cases/router-model-card-entry.png#bordered)

2. 查看 **Context window** 和 **Engine args**。下图中的 llama.cpp 模型配置为 `-c 104448`，即上下文大小为 **104448 token**。列表和模型卡片将它简写为 **102K**。客户端要求填写 token 数量时，应读取 **Engine args** 中的精确值。

   ![在 Router 的 Engine args 中查看 llama.cpp 的精确上下文大小](/images/manual/use-cases/router-model-card-context.png#bordered)

请以自己模型实例中的配置为准。`104448` 只是本例的值，并非所有安装的固定值。`-c` 是 llama.cpp 的参数，其他引擎使用不同的上下文参数。

客户端填写的上下文大小不能超过引擎的实际配置。只增大客户端的设置不会扩大引擎的上下文容量。更换 `default-chat` 指向的模型后，也要检查客户端的上下文设置。
<!-- #endregion model-context-window -->
