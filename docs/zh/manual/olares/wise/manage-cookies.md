---
outline: [2, 3]
description: 为 Wise 管理 Cookies，让它可以访问受保护的网站和订阅源。了解 Cookies 的作用、支持的格式、获取方式以及如何上传与维护。
---

# 管理 Cookie

部分网站需要登录后才能访问其内容。为了允许 Wise 从这些受保护的来源获取内容，你需要上传 Cookie，从而完成身份验证。

通过这种方法，Wise 得到授权，能够通过登录检查，确保**保存到库**或**订阅 RSS 源**在受限网站上顺利运行。

本文将介绍如何在 Olares 中为 Wise 管理 Cookie。

## 上传 Cookie

Olares 提供多种方式，让你可以灵活上传 Cookie，你可选择在浏览时即时捕获，或批量手动上传。

### 使用 LarePass 浏览器扩展上传

这是处理 Cookie 最简单的方式。当你浏览一个需要登录的网站时，[LarePass 扩展](https://www.olares.cn/larepass)可以一次性捕获并上传当前页面所需的 Cookie。

1. 打开目标网站，并确认你已经登录。
2. 打开 LarePass 浏览器扩展，点击“收集”按钮。
3. 点击**上传**来上传当前页面的 Cookie。

![使用 LarePass 扩展上传 Cookie](/images/zh/manual/olares/wise-upload-cookie-larepass.png#bordered){width=40%}

上传完成后，当你在该页面执行**保存到库**或**订阅 RSS 源** 时，Wise 会自动使用这些 Cookie。

### 手动批量上传

你可以手动粘贴 Netscape、JSON 或 Header String 格式的 Cookie 字符串。使用诸如 [Cookie-Editor](https://chromewebstore.google.com/detail/cookie-editor/iphcomljdfghbkdcfndaijbokpgddeno?utm_source=ext_app_menu)
等浏览器扩展从特定网站导出凭据时，可以使用此方法。

你可以通过 Wise 应用界面上传 Cookie 字符串，或直接在系统**设置**中上传。

<Tabs>
<template #上传至-Wise>

使用此方法，你无需离开 Wise 应用即可直接上传 Cookie。

:::info
Wise 仅支持上传 Cookie。如需查看、管理或删除 Cookie，需导航至 **设置** > **集成** > **管理您的 Cookie**。
:::

1. 在 Wise 中，点击左下角的 <i class="material-symbols-outlined">settings</i>**设置**。
2. 选择**偏好设置**。
3. 勾选**开启批量上传 Cookie**。
4. 点击左下角菜单栏中的<i class="material-symbols-outlined">add_circle</i>。
5. 选择**上传 Cookie**。
6. 选择 Cookie 文件的格式，并将对应的原始文本粘贴到输入框中。
7. 点击**确认**。

![在 Wise 批量上传 Cookie](/images/zh/manual/olares/wise-upload-cookie-in-pre.png#bordered)

以这种方式上传的 Cookie 仅存储在 Wise 中，在 Wise 访问匹配域名时自动使用。

</template>

<template #上传至集成>

使用此方法，你可以在同一界面上传、查看并管理你的 Cookie。

1. 从启动台打开**设置**。
2. 前往 **集成** > **管理您的 Cookie**。
3. 点击 **导入 Cookie**或右上角的 <i class="material-symbols-outlined">add_circle</i> 图标打开**上传 Cookie**对话框。
4. 选择 Cookie 文件的格式，并将对应的原始文本粘贴到输入框中，点击**确认**。

Wise 会在访问匹配的域名时自动检测并使用这些 Cookie。
</template>
</Tabs>

## 管理 Cookie

通常你不需要频繁维护 Cookie，但定期查看或清理 Cookie 也是有帮助的。

### 查看 Cookie 状态

要查看已有 Cookie 及其详情：

1. 从启动台打开**设置**。
2. 前往 **集成** > **管理您的 Cookie**。

如有需要，你可以按照上述步骤重新上传 Cookie 字符串来完成更新。

### 移除 Cookie

如果你不再需要访问特定站点，或想清理旧的凭据：

1. 从启动台打开**设置**。
2. 前往 **集成** > **管理您的 Cookie**。
3. 找到目标 Cookie 并点击 <i class="material-symbols-outlined">delete</i> 或**删除全部**。

删除后，Wise 将不再使用这些 Cookie，访问该站点的受保护内容可能会失败，除非你重新上传 Cookie。