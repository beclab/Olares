# Open WebUI embedding setup

<!-- #region configure-embeddings -->
Open WebUI uses an embedding model to turn document text into vectors for retrieval. `default-embedding` uses the default embedding model selected in Router.

If you change Router's default embedding model later, reindex your existing documents before searching them.

1. Open Router from Launchpad. On **Default models**, set EmbeddingGemma as the default embedding model. Wait until the `default-embedding` row on the same page shows **Callable**.
2. <!--@include: ./ai-service-connections.md#router-openai-base-url-->

3. In Open WebUI, select your profile icon, then go to **Admin Panel** > **Settings**.
4. On the left sidebar, find **Tools** and select **Documents**.
5. Under **Embedding**, enter the following settings:

   | Field | Value |
   | :--- | :--- |
   | **Embedding Model Engine** | **OpenAI** |
   | **API Base URL** | The Router Base URL from step 2, including `/v1` |
   | **Embedding Model** | `default-embedding` |

   If the API key field requires a value, enter `olares`.

6. Scroll to the bottom of the page and click **Reindex** in the lower-right corner to apply the changes.
7. Click **Save**.
<!-- #endregion configure-embeddings -->
