---
outline: [2, 3]
description: 了解如何在 Wise 中添加和管理 RSS 订阅、使用 OPML 导入订阅源，以及高效组织内容来源。
---

# 订阅和管理订阅源

Wise 提供灵活的订阅方式，帮助你跟踪喜爱的内容来源，包括文章、博客、播客和视频频道。本文将介绍如何添加和管理这些订阅。

## 了解订阅源

订阅一个源即建立了与内容来源的直接链接。新内容发布后会自动出现在订阅源列表中。

请注意“查看”与“保存”的区别：

- **订阅源用于发现内容**：你可以从订阅源列表中打开并阅读任意内容，但这些内容是临时的。如果该订阅源发生变化或你取消订阅，这些内容可能会无法获取。
- **内容库用于保存内容**：如果你想保留某条内容以便稍后阅读或标注，必须将其**保存到库**。只有保存后的内容才会真正成为你内容库的一部分，并可以使用状态、标签、笔记和过滤视图进行管理。

## 添加订阅

你可以在 Wise 内添加订阅，也可以在浏览器中直接订阅。

::: tip 上传 Cookie
如果链接指向的内容需要登录或其他访问控制，Wise 可能需要 Cookie 才能正确获取内容。若要配置 Cookie，请查看[管理 Cookie](./manage-cookies)。
:::

### 通过链接添加订阅源

如果你已经有网站地址或订阅源地址，可以通过**添加链接**订阅。

1. 点击左下角菜单栏中的 <i class="material-symbols-outlined">add_circle</i>，选择**添加链接**。
2. 粘贴网站或订阅源的 URL。
3. Wise 分析链接后，在操作列表中查看**订阅 RSS 源** 区域：
   - 如果检测到一个或多个订阅源，会在此列出。
   - 选择你想要订阅的订阅源。
![订阅 RSS 源](/images/zh/manual/olares/wise-add-link-subscribe.png#bordered){width=300}

选中的订阅源会被添加到**订阅源** 中，之后 Wise 会自动拉取该源的更新。

### 从 OPML 文件导入订阅源

如果你正在使用其他 RSS 阅读器，或者他人分享了订阅列表，你可以通过 OPML 文件导入这些订阅。

1. 点击左下角的 <i class="material-symbols-outlined">settings</i>，选择**偏好设置**。
2. 在**导入/导出**区域点击**从 OPML 文件导入订阅源**。
3. 选择你的 OPML 文件并打开。

导入完成后，这些订阅源会出现在 <i class="material-symbols-outlined">settings</i> >**订阅源**中。

:::info
Wise 仅导入新的订阅源，历史内容不会同步迁移。
:::

### 使用 LarePass 浏览器扩展订阅

LarePass 浏览器扩展可以让你在浏览网页时直接订阅，无需切回 Wise。

:::tip 安装 LarePass 浏览器扩展
LarePass 扩展目前仅支持 Chrome 浏览器。你可以从[官网](https://olares.cn/larepass)下载。
:::

1. 打开 LarePass 浏览器扩展，选择 **<i class="material-symbols-outlined">inventory_2</i>**。
2. 查看**订阅 RSS 源**区域：
   - 如果在当前页面检测到订阅源，会在此列出。
3. 点击目标订阅源旁边的 <i class="material-symbols-outlined">bookmark_add</i> 即可订阅。

![使用 LarePass 订阅](/images/zh/manual/olares/wise-larepass-collect.png#bordered)

## 管理订阅

你可以通过集中管理页面查看、重命名或整理所有订阅。

1. 点击左下角的 <i class="material-symbols-outlined">settings</i>。
2. 选择**RSS 订阅源**。
3. 按需执行以下操作：
    - 查看订阅源详情和描述。
    - 搜索指定订阅源。
    - 复制订阅源链接。
    - 编辑订阅源的名称和描述。
    - 移除订阅源链接。

此处的更改会同步到 Wise 侧边栏的**订阅源**。

## 取消订阅

如果你不想继续关注某个来源，可以取消订阅。

:::warning
取消订阅将删除该订阅源的所有文章，除非它们已保存到库中（例如设定内容状态为**收件箱**或**稍后阅读**）。此操作不可撤销。
:::

你可以在多个位置取消订阅：

- 阅读时，打开**信息**面板取消订阅：
    ![在信息面板取消订阅Info pane](/images/zh/manual/olares/wise-info-unsub.png#bordered){width=500}
- 阅读时，从阅读工具栏中取消订阅：
    ![Unsubscribe via toolbar](/images/zh/manual/olares/wise-toolbar-unsub.png#bordered){width=500}
- 在 Wise 中通过 <i class="material-symbols-outlined">settings</i> > **RSS 订阅源**取消订阅：
    ![Unsubscribe in RSS feeds](/images/zh/manual/olares/wise-unsub-rss-feeds.png#bordered){width=500}