---
description: 使用 LarePass 在多设备间收集网页内容、视频和 RSS 源。
---

# 使用 LarePass 收藏内容

LarePass 让你轻松收集并管理来自网络的各种信息——无论是文章、播客、视频还是 RSS 源。通过浏览器扩展和移动端应用，只需几次点击即可完成内容收集，并无缝同步至 Olares 的知识管理应用 Wise。

本文将介绍如何使用LarePass 扩展和移动客户端收集内容，以及如何订阅 RSS 源以持续获取更新。  
关于内容整理与使用，请参阅 [Wise 文档](../olares/wise/)。

## 准备工作

开始收集网页内容前，请确保：

- Olares 设备已启动  
- 手机已安装 LarePass  
- 已在 Chrome 中 [安装 LarePass 浏览器扩展](index.md#安装-larepass-浏览器扩展)

## 通过 LarePass 扩展收集内容

浏览器扩展是内容发现与收集的核心工具。  

::: tip 上传 Cookies 获得更好体验
某些网站对匿名用户有限制，可上传浏览器 Cookie 至 Olares 以提升抓取成功率。

1. 登录目标网站并打开LarePass 扩展。  
2. 进入**收集** > **Cookie**，点击**上传**。  
3. 悬停查看每个 Cookie 详情，如需排除点击 **X** 即可。  
:::

收集网页内容流程：

1. 访问目标页面（例如 CNN 文章）。  
2. 打开 LarePass 扩展，系统会自动检测可收集内容。  
3. 在**收集** > **页面**中，点击标题旁的 <i class="material-symbols-outlined">add_box</i> 将页面添加至 Wise。  

   ![Collect web content](/images/manual/tutorials/wise-collect-web-content.png#bordered)

收集完成的内容将出现在 Wise 的**库** > **文章**中，相关媒体文件（图片、视频、音频）保存至 `/download/Wise/Article`。

## 通过 LarePass 应用收集内容

你也可在移动浏览器中直接分享内容到 LarePass。

:::info
步骤因设备与浏览器而异，以下示例基于 iOS Safari。
:::

1. 在浏览器中点击 <i class="material-symbols-outlined">ios_share</i> **分享** 按钮：  
   - 从菜单中选择 **LarePass**，或  
   - 点击**更多操作** 并选择 **LarePass**。  

   ![Share to Wise](/images/manual/tutorials/wise-add-articles-via-share.png#bordered)

2. 跳转至 LarePass 应用后，系统会自动识别可收集内容并提示添加至 Wise。  
3. 点击**确认**完成收集。  

::: tip 复制链接快速分享
也可复制网址后直接打开 LarePass，它会自动检测剪贴板中的 URL 并提示收集。
:::

添加成功的内容同样会出现在 **Wise** 的**库** > **文章**中。

## 订阅 RSS 源

LarePass 支持 RSS 订阅，轻松跟踪播客、博客等内容源。

通过 LarePass 浏览器扩展订阅：

1. 访问 RSS 源或播客页面（如 `https://www.spreaker.com/podcast/paranormal-mysteries--2321086`）。  
2. 打开 **LarePass 扩展**，系统会自动检测并显示 **RSS** 选项卡。  
3. 点击想要订阅的源旁的 <i class="material-symbols-outlined">bookmark_add</i> 即可订阅。  

   ![Subscribe to podcast](/images/manual/tutorials/wise-sub-podcast.png#bordered)

已订阅的源将同步至 Wise，所有新内容可在同一位置集中查看。
```
