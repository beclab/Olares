# AI service connections

<!-- #region model-connection-overview -->
:::details How model connections work
A standalone model on Olares runs as a separate service from the client app. To connect them, the client needs the exact **Model name** and a **Base URL** that matches the API format it expects.

You can get both values from the model's console. For more details, see [Connect AI apps](/manual/best-practices/connect-ai-apps.md).
:::
<!-- #endregion model-connection-overview -->

<!-- #region get-model-connection-details -->
1. Open the model app from Launchpad. Its Model Console opens automatically.
2. Wait until **Model** shows **READY** and **Engine** shows **RUNNING**.

   ![Qwen3.6-27B model console](/images/manual/use-cases/qwen3.6-27b-model-console1.png#bordered)

3. Under **Model**, copy the **Model name** exactly as shown.
4. Under **Engine**:

   a. **Connection source**: Select **Apps in Olares**. 
   
   b. **API format**: Select **OpenAI-Compatible**.
   
   c. Copy the provided **Base URL** exactly as shown.

<!-- #endregion get-model-connection-details -->

<!-- #region use-different-model -->
:::details Optional: Use a different model
You can use a different model size or provider instead of the one listed above:

- Install a different model app from Market.
- Create a model instance from [Engine Base apps](llm-base-apps.md) to bring your own model.

If your AI agent app has the [Olares CLI](../developer/cli-install.md) and [Agent Skills](../developer/cli-agent-skills.md) installed, ask it to deploy the model and skip the manual setup. For example:

```plain
Deploy qwen3.5:9b on my Olares using the Ollama Engine Base.
```
:::
<!-- #endregion use-different-model -->

<!-- #region app-endpoint-overview -->
:::details How app endpoints work
When a client connects to another Olares app, it uses that app's endpoint as the network address. If the app exposes multiple endpoints, choose the one that matches the feature or protocol the client needs.
:::
<!-- #endregion app-endpoint-overview -->

<!-- #region get-model-connection-details-anthropic -->
1. Open the model app from Launchpad. Its Model Console opens automatically.
2. Wait until **Model** shows **READY** and **Engine** shows **RUNNING**.

   ![Qwen3.6-27B model console](/images/manual/use-cases/qwen3.6-27b-model-console-anthropic.png#bordered)

3. Under **Model**, copy the **Model name** exactly as shown.
4. Under **Engine**:

   a. **Connection source**: Select **Apps in Olares**. 
   
   b. **API format**: Select **Anthropic-Compatible**.
   
   c. Copy the provided **Base URL** exactly as shown.

<!-- #endregion get-model-connection-details-anthropic -->