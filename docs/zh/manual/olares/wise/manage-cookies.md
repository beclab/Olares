---
outline: [2, 3]
description: 为 Wise 上传、更新或删除 Cookie，帮助其访问需要登录的网站和订阅源。
head:
  - - meta
    - name: keywords
      content: Olares, Wise, Cookie 管理, 上传 Cookie, 受限网站, LarePass 浏览器扩展
---

# 为 Wise 管理 Cookie <Badge type="tip" text="Olares 1.12.8+" />

有些网站需要登录后才能访问。如果 Wise 因此无法收藏网页、订阅来源或下载资源，可以上传该网站的 Cookie。Cookie 可能过期，上传后也不一定能解决所有访问问题。

本文介绍如何上传、更新和删除 Wise 使用的 Cookie。

## 上传 Cookie

你可以在浏览网页时通过 LarePass 上传，也可以手动粘贴导出的 Cookie。

### 通过 LarePass 浏览器扩展上传

登录目标网站后，可以用 [LarePass 浏览器扩展](https://www.olares.cn/larepass)上传该网站的 Cookie。

1. 打开目标网站，确认已登录。
2. 打开 LarePass 浏览器扩展，点击“收集”图标。
3. 点击**上传 Cookie**，上传当前网页的 Cookie。

上传后，再试一次收藏或订阅。如果还是无法访问，请检查 Cookie 是否过期、是否来自同一个网站。

### 手动批量上传

如果已经用 [Cookie-Editor](https://chromewebstore.google.com/detail/cookie-editor/iphcomljdfghbkdcfndaijbokpgddeno?utm_source=ext_app_menu) 等浏览器扩展导出 Cookie，也可以手动粘贴。Wise 支持 Netscape、JSON 和 Cookie header 格式。

你可以在 Wise 中上传，也可以直接通过 Olares 设置上传。

<Tabs>
<template #在-Wise-中上传>

这种方法无需离开 Wise。

:::info
Wise 仅提供上传入口。如需查看、管理或删除这些 Cookie，请打开 **Olares 设置** > **集成** > **管理 Cookie**。
:::

1. 在 Wise 中点击左下角的设置图标，选择**偏好设置**。
2. 在**上传设置**下勾选**开启批量上传 Cookie**。
3. 点击左下角的加号图标，选择**上传 Cookie**。
4. 在**上传 Cookie**窗口选择 **Netscape**、**JSON** 或 **Cookie header**，粘贴导出的 Cookie。
5. 点击**确认**。

    ![在 Wise 中批量上传 Cookie](/images/zh/manual/olares/wise-upload-cookie-in-pre1.png#bordered)

访问对应网站时，Wise 可以使用这些 Cookie。

</template>

<template #在-Olares-设置中上传>

这种方法可以在同一个界面上传、查看和管理 Cookie。

1. 打开 **Olares 设置**。
2. 前往**集成** > **管理 Cookie**，打开**Cookie 管理**。
3. 点击**导入 Cookie**，打开**上传 Cookie**窗口。
4. 选择 **Netscape**、**JSON** 或 **Cookie header**，粘贴导出的 Cookie，然后点击**确认**。

访问对应网站时，Wise 可以使用这些 Cookie。

</template>
</Tabs>

## 更新或删除 Cookie

1. 打开 **Olares 设置** > **集成** > **管理 Cookie**，查看已上传的 Cookie。
2. 如果 Cookie 已过期，使用上述任一种方法上传新的 Cookie。
3. 如果不再需要某个 Cookie，找到对应项目，点击 <i class="material-symbols-outlined">delete</i>。

删除后，在重新上传 Cookie 前，Wise 可能无法访问需要这些 Cookie 的内容。
