---
outline: [2, 3]
description: Open WebUI 在 Olares 上的常见问题与解决方法。
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, 常见问题, 故障排除, 模型下载
app_version: "1.0.38"
doc_version: "1.1"
doc_updated: "2026-07-31"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/openwebui-common-issues.md)。
:::

# Open WebUI 常见问题

使用本页面识别并解决 Open WebUI 在 Olares 上的常见问题。

## 模型应用卡在 "Waiting for Ollama" 或 "Needs attention"

如果模型应用在这些状态停留超过几分钟：

1. 前往**设置** > **AI 算力**。
2. 检查你的 GPU 模式：
   - 如果你正在使用**容量切分**，需确保模型应用已关联到 GPU，并分配了足够的 VRAM。
   - 如果你正在使用**独占**，需确保独占应用设置为你的模型应用。
3. 从 Launchpad 重启模型应用，然后再次检查状态。

## 麦克风出现 "Permission denied" 错误

尝试使用听写按钮或 Voice Mode 时，你可能会收到以下错误消息：
- `Permission denied when accessing microphone: NotAllowedError: Permission denied`
- `Permission denied when accessing media devices`

Olares 桌面会在嵌入式框架（iframe）中显示应用。出于严格的安全和隐私考量，现代浏览器会阻止嵌入式框架访问麦克风等敏感硬件，即使你已经在系统设置中授予浏览器相应权限。

要绕过此安全限制并使用麦克风：
1. 在 Olares 桌面的 Open WebUI 窗口右上角，选择 <i class="material-symbols-outlined">open_in_new</i>，在新的浏览器标签页中打开。
2. 在新的浏览器标签页中，点击聊天界面的麦克风图标。
3. 当浏览器提示时，允许麦克风访问。

## Open WebUI 无法搜索网页

如果 AI 回复中没有包含网页搜索结果，请按以下顺序排查：

1. **对话中已启用搜索功能。**  
   在聊天区域，点击 **Integrations** 图标，并确保 **Web Search** 已启用。

2. **嵌入模型已配置且可访问。**  
   前往 **Admin Panel** > **Settings** > **Tools** > **Documents**，验证嵌入模型设置。如果嵌入模型缺失或无法访问，Open WebUI 将无法处理网页搜索结果。

3. **SearXNG 能返回你的查询结果。**  
   从启动台打开 SearXNG，运行相同的搜索。如果 SearXNG 没有返回结果，请检查 **Preferences** > **ENGINES** 中是否已启用可用的搜索引擎。

4. **网页加载器被拦截。**  
   如果你需要完整网页内容，而默认网页加载器因反爬取机制失败，请前往 **Admin Panel** > **Settings** > **Tools** > **Web Search**，然后启用 **Bypass Web Loader**。这将使用搜索结果摘要，而不是获取完整页面。

   :::tip
   如需稳定地检索完整网页，请安装并配置 Firecrawl 作为网页加载器。参阅[将 Firecrawl 用作网页加载器](firecrawl.md#configure-open-webui)。
   :::
