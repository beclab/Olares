# Open WebUI 嵌入配置

<!-- #region configure-embeddings -->
Open WebUI 使用嵌入模型将文档文本转换成向量，以便检索。`default-embedding` 会调用 Router 中设置的默认嵌入模型。

以后更换 Router 的默认嵌入模型时，请先为已有文档重新建立索引，再进行检索。

1. 从启动台打开 Router。在 **Default models** 页面将 EmbeddingGemma 设为默认嵌入模型，等待同一页面中的 `default-embedding` 行显示为 **Callable**。
2. <!--@include: ./ai-service-connections.md#router-openai-base-url-->

3. 在 Open WebUI 中点击头像图标，进入 **Admin Panel** > **Settings**。
4. 在左侧边栏找到 **Tools**，选择 **Documents**。
5. 在 **Embedding** 区域填写以下设置：

   | 字段 | 值 |
   | :--- | :--- |
   | **Embedding Model Engine** | **OpenAI** |
   | **API Base URL** | 步骤 2 中的 Router Base URL，包含 `/v1` |
   | **Embedding Model** | `default-embedding` |

   如果 API 密钥字段要求必填，填写 `olares`。

6. 滚动到页面底部，点击右下角的 **Reindex** 以应用更改。
7. 点击 **Save**。
<!-- #endregion configure-embeddings -->
