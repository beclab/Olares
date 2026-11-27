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

:::tip 同步客户端上下文设置
客户端的上下文长度应设为模型实例中 **Engine args** 配置的值或更小。更换 `default-chat` 指向的模型后，同步检查该值。不同引擎使用各自的上下文参数，例如 llama.cpp 使用 `-c`。
:::
<!-- #endregion model-context-window -->
