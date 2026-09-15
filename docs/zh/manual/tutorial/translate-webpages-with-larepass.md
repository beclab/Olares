---
outline: [2, 3]
description: 使用 LarePass Chrome 扩展和运行在 Olares 上的 Hy-MT2 模型，更私密地翻译网页。
head:
  - - meta
    - name: keywords
      content: Olares, LarePass, 网页翻译, 沉浸式翻译, Hy-MT2, 本地翻译, 隐私翻译
---

# 使用 LarePass 私密翻译网页 <Badge type="tip" text="^ 1.12.7" />

使用云端翻译扩展时，网页文本需要发送到服务商的服务器处理。LarePass 采用自托管方式，将文本通过 Router 交给运行在自己 Olares 上的 **Hy-MT2**，而不是第三方翻译服务，从而更自主地控制内容的处理方式。

同时，还能获得和常见浏览器翻译工具一样简单的使用体验：

- **隐私优先**：Router 将 LarePass 获取的网页文本转发给运行在自己 Olares 上的模型，而不是交给第三方翻译服务处理。
- **轻量且支持多语言**：Hy-MT2 支持 33 种语言互译，其 18 亿参数规模兼顾了翻译质量和本地运行所需的资源。
- **自动集成**：Router 使用 Olares 身份验证请求，并自动识别和配置已安装的 Hy-MT2 模型，无需配置模型地址或 API 密钥。

## 前提条件

- [安装 LarePass 浏览器扩展](../install-larepass-browser-extension.md)。
- [导入 Olares 账户](../larepass/manage-accounts.md#chrome-扩展)。
- 确保 Olares 上已安装 **Router**。

## 开始本地翻译

1. 在 [Olares 应用市场](../olares/market/market.md#安装模型)搜索并安装 **Hy-MT2-1.8B**。等待安装完成，以便 Router 自动识别并配置该模型。
2. 打开一个网页，点击 Chrome 工具栏中的 LarePass 图标。点击 <i class="material-symbols-outlined">translate</i>，选择源语言和目标语言，并将**翻译服务**设置为 **Olares Hy-MT2**，然后点击**翻译当前页面**。

如需恢复原文，点击**显示原文**。如果安装扩展或模型时网页已经打开，请先刷新再翻译。

## 调整翻译体验

在**翻译**页面，可以调整：

- **自动翻译**：跟随全局设置、总是翻译当前网站，或从不翻译当前网站。
- **仅显示翻译**：翻译后隐藏原文。
- **悬浮显示原文**：开启**仅显示翻译**后才会出现。将鼠标悬停在译文段落上时，临时显示原文。

如需修改全局默认设置和译文样式，请点击 <i class="material-symbols-outlined">settings</i>。

## 故障排查

- **无法使用翻译服务**：Router 可能仍在识别并配置 Hy-MT2。等待配置完成，然后刷新网页。
- **页面无法翻译**：Chrome 扩展无法翻译 `chrome://extensions/` 等内部页面，也无法翻译未提供常规网页文本的页面。可以换一篇普通文章或博客进行尝试。
