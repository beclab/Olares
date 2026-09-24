# AI service connections

<!-- #region router-prerequisite -->
- An Olares device running Olares 1.12.7 or later, with sufficient disk space and memory.
- Access to Router for local model connections.
<!-- #endregion router-prerequisite -->

<!-- #region model-connection-overview -->
:::details How model connections work
On Olares 1.12.7 and later, AI clients connect through Olares Router. Router provides the Base URL and routes `default-chat` to the model selected on its **Default models** page.

This guide uses Qwen3.8-27B (llama.cpp) as the default chat model. For connection formats and API key requirements, see [Connect AI apps](/manual/best-practices/connect-ai-apps.md).
:::
<!-- #endregion model-connection-overview -->

<!-- #region get-model-connection-details -->
1. Open Router from Launchpad. On **Default models**, set Qwen3.8-27B (llama.cpp) as the default chat model.
2. Go to **LLM**, find Qwen3.8-27B (llama.cpp), and click **View connection example** on its row.

   ![View the Qwen3.8-27B connection example in Router](/images/manual/use-cases/router-view-connection-examp.png#bordered)

3. In **How to call this model**, select **Apps in Olares** and copy the **Base URL**, including `/v1`.

   ![Copy the Router Base URL for apps in Olares](/images/manual/use-cases/router-how-to-call-model.png#bordered)

4. Use `default-chat` as the model name in the client. Apps in Olares do not need a Router API key; leave the key empty where possible, or use `olares` if the client requires a value.

   `default-chat` is a routing name and is not returned by the model-list API. Add it manually if the client fetches a model list. If the client only supports selecting a listed model, use the full model name from Router instead.
<!-- #endregion get-model-connection-details -->

<!-- #region use-different-model -->
:::details Optional: Use a different model
Install another chat model from Market, or create a model instance with [Engine Base apps](llm-base-apps.md). Then select it as the default chat model on Router's **Default models** page. Clients configured with `default-chat` use the new default without changing their model settings.

To keep a client on one specific model, use the full model name from Router's **How to call this model** window instead.
:::
<!-- #endregion use-different-model -->

<!-- #region app-endpoint-overview -->
:::details How app endpoints work
When a client connects to another Olares app, it uses that app's endpoint as the network address. If the app exposes multiple endpoints, choose the one that matches the feature or protocol the client needs.
:::
<!-- #endregion app-endpoint-overview -->

<!-- #region get-model-connection-details-anthropic -->
1. Open Router from Launchpad. On **Default models**, set Qwen3.8-27B (llama.cpp) as the default chat model.
2. Go to **LLM**, find Qwen3.8-27B (llama.cpp), and click **View connection example** on its row.

   ![View the Qwen3.8-27B connection example in Router](/images/manual/use-cases/router-view-connection-examp.png#bordered)

3. In **How to call this model**, select **Apps in Olares** and copy the **Base URL**, then remove the trailing `/v1` for the Anthropic-compatible client. For example, use `https://router.<your-olares-domain>`; the client appends `/v1/messages`.

   ![Copy the Router Base URL for apps in Olares](/images/manual/use-cases/router-how-to-call-model.png#bordered)

4. Use `default-chat` as the model name in the client. Apps in Olares do not need a Router API key; leave the key empty where possible, or use `olares` if the client requires a value.

   `default-chat` is a routing name and is not returned by the model-list API. Add it manually if the client fetches a model list. If the client only supports selecting a listed model, use the full model name from Router instead.
<!-- #endregion get-model-connection-details-anthropic -->

<!-- #region get-embedding-model-connection-details-openai -->
1. Open Router from Launchpad and go to **Tools**. Find the installed embedding model and wait until it shows **Callable**.
2. On its model row, click **View connection example**.
3. In **How to call this model**, select **Apps in Olares** and copy the **Base URL**, including `/v1`.
4. Copy the full **Model name** from this window, including the `Olares/` prefix, and use it in the client's embedding settings. Apps in Olares do not need a Router API key; use `olares` only if the client requires a value.

Use the embedding model's name, not `default-chat`. Keep the same embedding model when querying an existing knowledge base; changing it can require reindexing your documents.
<!-- #endregion get-embedding-model-connection-details-openai -->

<!-- #region model-context-window -->
1. In Router, open **LLM**, find the model, and click the information icon at the right end of its row to open **Model card**.

   ![Open a model card from the Router LLM list](/images/manual/use-cases/router-model-card-entry.png#bordered)

2. Review **Context window** and **Engine args**. For the llama.cpp model shown below, `-c 104448` gives the configured context size: **104448 tokens**. The list and model card display this as **102K**. Use the exact value from **Engine args** when a client asks for a token count.

   ![Read the exact llama.cpp context size in Router Engine args](/images/manual/use-cases/router-model-card-context.png#bordered)

Use the value shown for your own model instance; `104448` is an example, not a fixed value for every installation. The `-c` parameter is specific to llama.cpp; other engines use different context parameters.

The client's context setting must not exceed the engine's configured context size. Increasing the client setting alone does not increase the engine's capacity. If you change the model behind `default-chat`, review the client's context setting as well.
<!-- #endregion model-context-window -->
