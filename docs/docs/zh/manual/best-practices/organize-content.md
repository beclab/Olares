---
outline: [2, 3]
description: 使用 Wise 构建知识库的完整教程，学习如何收集网页内容、管理文档、订阅 RSS 和搜索资料，打造个人信息中心。
---
# 使用 Wise 打造知识中枢

在不同来源和设备之间管理信息往往充满挑战。你可能正在使用多个工具来收藏文章、跟踪 RSS 订阅或管理文档，最终却导致工作流程碎片化。

Wise 是 Olares 的内置应用，旨在集中管理和组织你的知识。它不仅能从网络和设备中收集信息，还能通过本地推荐算法，帮助你以私密方式发现有价值的内容，而不受第三方推荐算法影响。

本教程将指导你如何利用 Wise 和 LarePass 在不同平台间收集、整理和访问内容。

## 目标

通过本教程，你将学会如何：

- 使用 LarePass 浏览器扩展或移动客户端收集网页内容。
- 上传 PDF 和 EPUB 等现有文件到 Wise，同一管理你的内容。
- 订阅 RSS 订阅源，随时了解关注的博客、播客或视频播放列表的最新动态。
- 快速定位并检索你在个人信息中枢中整理的任何内容。

## 开始之前

开始以前，请确保：

- Olares 设备已激活并正常运行。
- Olares ID 已[备份助记词](../larepass/back-up-mnemonics.md)。
- 手机上已安装 LarePass 应用。

## 安装 LarePass 浏览器扩展

LarePass 浏览器扩展是内容发现和内容收集的核心应用。

::: tip 仅支持 Chrome
LarePass 扩展目前仅支持 Chrome 浏览器。
:::

<tabs>
<template #从-Chrome-应用商店安装>

1. 在 Chrome 网上应用店中搜索 LarePass。

2. 打开详情页面，点击**添加至 Chrome**进行安装。

3. 通过导入 Olares ID 登录 LarePass 扩展：

   a. 打开 LarePass 扩展，点击**导入账号**。

   b. 使用相应的助记词导入你的 Olares ID。

   c. 输入 Olares 密码完成登录。
</template>
<template #离线安装>

1. 访问 [https://olares.cn/larepass](https://olares.cn/larepass)，手动下载 LarePass 扩展的安装文件并解压。

2. 在 URL 地址栏中输入 `chrome://extensions/`，进入扩展管理页面。

3. 在右上角启用**开发者模式**。

4. 点击**加载已解压的扩展程序**，选择解压后的 LarePass 扩展文件夹完成安装。

5. 通过导入 Olares ID 登录 LarePass 扩展：

   a. 打开 LarePass 扩展，点击**导入账号**。

   b. 使用相应的助记词导入你的 Olares ID。

   c. 输入 Olares 密码完成登录。
</template>
</tabs>

:::tip 快捷访问
安装完成后，将 LarePass 扩展固定到 Chrome 扩展菜单中，方便后续快速访问。
:::

登录后，LarePass 浏览器扩展将与你的 Olares 设备同步。这意味着你通过 LarePass 扩展收集的所有内容都会自动添加到 Wise 库中。

## 收集在线内容

你可以使用 LarePass 浏览器扩展或移动客户端收集在线内容，包括网页文章、视频和播客等。

### 通过 LarePass 扩展收集

::: tip 上传 Cookie 以优化体验
一些网站会限制匿名用户访问。这种情况下，你可以将 Cookie 上传到 Olares 以优化体验。

1. 登录该网站，打开 LarePass 扩展。
2. 进入**收集**> **Cookie** 页面，并点击**上传**。鼠标悬停可查看 Cookie 详情。如果不想上传某个 Cookie 项，可以点击 **X** 按钮取消选择。
:::

通过 LarePass 扩展收集网页内容：

1. 打开内容页面，如 B 站视频。
2. 打开 LarePass 扩展。扩展会自动检测并获取当前页面的可收集内容。
3. 在**收集** > **网页**页签下，点击内容标题旁的 <i class="material-symbols-outlined">add_box</i>，将该页面添加到 Wise 库中。

   ![收集在线内容](/images/zh/manual/tutorials/wise-collect-web-content.png#bordered)

收集成功后，你可以在 Wise 的**库** > **文章**中找到对应内容。页面上音频、视频和图片等媒体文件也会被下载到 Olares 本地的 `/download/Wise/Article` 文件夹。

![查看文章](/images/zh/manual/tutorials/wise-view-article.png#bordered)

### 通过 LarePass 移动端

你可以将网页链接分享到 LarePass 移动客户端来收集内容。
:::info
具体步骤可能会因操作系统和浏览器而有所不同。以下以 Safari 为例。
:::

1. 在浏览器中点击 <i class="material-symbols-outlined">ios_share</i>，选择 LarePass 为分享对象。你可以：
   - 在分享选项中选择 LarePass 的图标，或者
   - 在**其他操作**栏中选择 **LarePass**

   ![收藏到 Wise](/images/zh/manual/tutorials/wise-add-articles-via-share.png#bordered)

   LarePass 应用会自动打开并检测待分享页面的内容，并提示是否要添加至 Wise。
2. 点击**确认**添加。

::: tip 复制 URL 分享
你也可以直接复制网页 URL 并打开 LarePass。LarePass 会自动检测剪贴板中的 URL 和可收藏的内容。
:::

收集完成后，可以在 Wise 的**库** > **文章**中阅读收集的文章。

## 上传 PDF 及电子书内容

你可以将本地的 PDF 或 EPUB 电子书内容上传到 Wise 进行集中管理。这样可以将阅读材料、笔记和相关内容保存在一处，方便整理、检索和随时查阅。

1. 打开 Wise, 点击菜单栏下方的 <i class="material-symbols-outlined">add_circle</i> 按钮，选择**上传**。
2. 进入包含你想要上传文件的目录，选择文件，并点击**确认**。

你可以在**库** > **PDF** 下查看上传的 PDF，在**库** > **图书**下查看 EPUB 电子书。

![查看 PDF](/images/zh/manual/tutorials/wise-pdf.png#bordered)

::: tip 
你可以用标签高效地分类和关联相关内容，或者通过笔记直接在内容旁记录见解和想法。详细用法请参考[组织你的阅读](../olares/wise/basics.md#组织你的阅读)。
:::

## 订阅 RSS 内容
你可以在 Wise 中订阅播客、博客和喜爱的视频播放列表。系统会自动下载更新的剧集和内容，让你轻松追更，同时也无需担心原内容已被删除或无法访问。另外，对于那些往往不提供 RSS 订阅源的视频网站，Wise 也能自动下载你收藏的节目。

### 通过浏览器扩展订阅

通过 LarePass 扩展订阅 RSS 步骤如下：

1. 在浏览器中打开要订阅的 RSS 页面，例如 “加州101”的播客：`https://www.xiaoyuzhoufm.com/podcast/5e280faf418a84a0461fbd0d`。
2. 打开 LarePass 扩展。扩展会自动识别页面的 RSS 订阅源，并显示 **RSS** 页签。
3. 在 **RSS** 页签下，找到正确的订阅源，点击 <i class="material-symbols-outlined">bookmark_add</i> 按钮以完成订阅。

   ![订阅播客](/images/manual/tutorials/wise-sub-podcast.png#bordered)

### 手动添加 RSS 源

通过以下步骤手动添加 RSS 源至 Wise：

1. 获取目标 RSS 订阅链接，如 HackerNews 头条订阅源 `https://hnrss.org/frontpage`。
2. 打开 Wise，在菜单栏点击 <i class="material-symbols-outlined">add_circle</i> 按钮，并选择 **RSS 源**。
3. 输入网址后，Wise 将自动识别出可用的 RSS 订阅源。

   ![手动添加 RSS](/images/zh/manual/tutorials/wise-add-rss.png#bordered){width=50%}
4. 点击**添加**完成订阅。

### 自动下载收藏视频 <Badge type="tip" text="^1.11.3" />

除了常规的 RSS 订阅，你可以通过 LarePass 扩展和 Wise 自动下载收藏的视频。以 B 站为例：

1. 在浏览器里打开并登陆 B 站。
2. 打开 LarePass 扩展，在 Cookie 页签下点击**上传**按钮以完成 Cookie 上传。Olares 需要使用视频网站的 Cookie 来访问你的收藏夹并下载视频。 
   
   ::: tip 打开自动同步 Cookie 功能
   Cookie 信息可能会过期。建议你在 Cookie 页面启用**自动同步**功能，以确保每次访问网站时，Cookie 会自动更新。
   :::
3. 进入你的 B 站收藏夹管理页面，新建一个收藏夹并打开。此处，我们创建了一个`收藏到 Olares`的收藏夹。
4. 打开 LarePass 扩展。扩展会自动获取当前页面的订阅源并显示 **RSS** 页签。

   ![订阅 B 站收藏夹](/images/zh/manual/tutorials/wise-bilibili-rss.png#bordered)
5. 选择正确的订阅源。此处我们选择 **RSS** 页签下第二个源：“UP主非默认收藏夹”。点击 <i class="material-symbols-outlined">bookmark_add</i> 完成订阅。

订阅完成后，任何添加到此收藏夹的视频都会被 Olares 自动下载至本地。

::: tip 自动下载点赞投币视频
你也可以通过 RSS 订阅方式自动下载 B 站点赞或投币的视频：
1. 在 B 站的**个人空间** > **个人资料**里获取你的 UID，通常是一串 8 位的数字。
2. 手动添加以下 RSS 订阅源至 Wise：
   - 点赞视频：`rsshub://bilibili/user/like/你的uid/`
   - 投币视频：`rsshub://bilibili/user/coin/你的uid/`
:::

### 访问 RSS 内容

要访问所有通过 RSS 方式订阅的内容：
1. 在左侧菜单栏中点击**订阅** > **订阅源**。
2. 选择一个未读的 RSS 项目，进入即可观看视频、收听播客或阅读文章。

::: tip 智能内容推荐
除了常规的 RSS 订阅，Wise 还提供了完全本地运行、保护隐私的智能内容推荐系统，为你提供个性化内容推送。详情请参考[本地推荐算法](../olares/wise/recommend.md)。
:::

## 搜索知识库内容
<!--@include: ./wise.reusables.md{4,13}-->

## 了解更多

- [Wise 基本操作](../olares/wise/basics.md)
- [本地推荐算法](../olares/wise/recommend.md)
- [订阅和管理订阅源](../olares/wise/subscribe.md)
