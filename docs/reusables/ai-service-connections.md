# AI service connections

<!-- #region model-connection-overview -->
:::info How model connections work
A standalone model on Olares runs as a separate service from the client application. To connect them, the client often needs the exact **Model name** and a **Base URL** that matches the API format it supports.

You can get these values from the model console. For a more detailed explanation, see [Connect AI apps](/manual/best-practices/connect-ai-apps.md).
:::
<!-- #endregion model-connection-overview -->

<!-- #region get-model-connection-details -->
1. Open the model application from Launchpad. The Model Console opens automatically.
2. If the model starts downloading, wait until **Model** shows **READY** and **Engine** shows **RUNNING**.
3. Under **Service status**, select **Apps in Olares**. Find the API format supported by the client, then copy the **Model name** and corresponding **Base URL** exactly as displayed.
<!-- #endregion get-model-connection-details -->

<!-- #region app-endpoint-overview -->
:::info How app endpoints work
When a client connects to another Olares app, it uses the app's endpoint as the network address. If the app exposes multiple endpoints, choose the one that matches the feature or protocol required by the client.
:::
<!-- #endregion app-endpoint-overview -->
