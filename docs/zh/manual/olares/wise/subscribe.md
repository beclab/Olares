---
outline: [2, 3]
description: 在 Wise 中订阅 RSS 来源、挑选要保存的内容，并管理或取消订阅。
head:
  - - meta
    - name: keywords
      content: Olares, Wise, RSS 订阅, 订阅源管理, OPML, LarePass 浏览器扩展
---

# 订阅和管理订阅源 <Badge type="tip" text="Olares 1.12.8+" />

用 Wise 订阅网站、博客或播客后，可以在**订阅源**中集中查看新内容。想留着以后再看，就将它加入**收件箱**或**稍后阅读**。本文还介绍如何管理和取消订阅。

## 前提条件

根据要使用的功能，先准备：

- **获取订阅更新**：在应用市场搜索“RSS Subscribe”并安装插件。
- **从浏览器订阅**：[安装 LarePass 浏览器扩展](../../install-larepass-browser-extension)，并将其连接到 Olares。

## 添加订阅

你可以粘贴链接、导入 OPML 文件，或在浏览网页时使用 LarePass 订阅。

:::tip 需要登录的内容
如果链接需要登录或受到其他访问限制，Wise 可能需要 Cookie 才能获取内容。参阅[为 Wise 管理 Cookie](./manage-cookies)。
:::

### 通过链接添加订阅源

如果知道网站或订阅源的 URL，可以通过**添加链接**订阅。

1. 点击左下角的 <i class="material-symbols-outlined">add_circle</i>，选择**添加链接**。
2. 粘贴网站或订阅源的 URL。如果检测到订阅源，它们会显示在**订阅 RSS 源**区域。
3. 点击目标订阅源旁的 <i class="material-symbols-outlined">bookmark_add</i>。

    ![通过链接订阅 RSS 源](/images/zh/manual/olares/wise-add-link-subscribe1.png#bordered){width=400}

### 从 OPML 导入订阅源

:::info
OPML 文件只包含订阅列表，不包含文章和阅读历史。导入后，Wise 会从这些订阅源获取目前可用的最新内容。
:::

1. 点击左下角的 <i class="material-symbols-outlined">settings</i>，选择**偏好设置**。
2. 在**导入/导出**下点击**从 OPML 导入订阅源**。
3. 选择 OPML 文件并打开。

### 使用 LarePass 扩展订阅

通过 [LarePass 浏览器扩展](https://www.olares.cn/larepass)，你可以在浏览网页时订阅，无需切回 Wise。

1. 打开 LarePass 浏览器扩展，点击“收集”图标。
2. 如果检测到订阅源，它们会显示在**订阅 RSS 源**区域。
3. 点击目标订阅源旁的 <i class="material-symbols-outlined">bookmark_add</i>。

    ![通过 LarePass 订阅 RSS 源](/images/zh/manual/olares/wise-larepass-collect2.png#bordered)

## 浏览并保存内容

订阅后，新内容会出现在**订阅源**，不会自动加入**收件箱**或**稍后阅读**。如果订阅源发生变化，或你取消订阅，没有保存的内容可能会消失。要留着以后再看，请按以下步骤保存：

1. 打开**订阅源**查看新内容。你也可以直接点开阅读。
2. 找到想保留的内容，在以下任一位置选择**收件箱**或**稍后阅读**：
   - **内容卡片上**：使用卡片旁的操作按钮。下图以**收件箱**为例。

     ![从内容卡片添加到收件箱](/images/zh/manual/olares/wise-add-to-inbox.png#bordered)

   - **阅读器中**：阅读内容时使用工具栏。下图以**稍后阅读**为例。

     ![从阅读器添加到稍后阅读](/images/zh/manual/olares/wise-add-to-read-later.png#bordered)

3. 保存后，可在相应的视图或标签页中找到它。

## 管理订阅源设置

在 Wise 中打开 <i class="material-symbols-outlined">settings</i> **设置** > **RSS 订阅源**。找到要管理的订阅源后，可以通过**操作**列：

- **刷新**订阅源，检查是否有新内容。
- **复制**订阅源链接。
- **编辑**名称、描述及更新选项。
- **移除**订阅源。参阅[取消订阅](#取消订阅)。

![在 Wise 中管理订阅源](/images/zh/manual/olares/wise-rss-feeds.png#bordered)

要编辑订阅源：

1. 点击编辑图标，打开**编辑 RSS 源**。
2. 修改标题或描述、调整内容和下载选项，或设置更新间隔，然后点击**确认**。

要将订阅列表导出为 OPML 文件：

1. 打开**设置** > **偏好设置**。
2. 在**导入/导出**下点击**将订阅源导出为 OPML**。

导出的文件只包含订阅列表，不包含已保存的文章或阅读历史。

## 取消订阅

如果不想继续关注某个来源，可以取消订阅。

:::warning 取消订阅前

- **保存到库中的内容会保留**：取消订阅后，该来源中未保存的内容会被移除。
- **检查关联文件**：确认窗口提供**同时删除关联文件**选项，确认前请留意是否勾选。

:::

你可以从以下位置取消订阅：

- 该订阅源内容的**信息**面板。
    ![通过信息面板取消订阅](/images/zh/manual/olares/wise-info-unsub.png#bordered)
- 阅读该订阅源的内容时，使用阅读器工具栏。
    ![通过阅读器工具栏取消订阅](/images/zh/manual/olares/wise-toolbar-unsub.png#bordered)
- Wise 的 <i class="material-symbols-outlined">settings</i> > **RSS 订阅源**。
    ![在 RSS 订阅源列表中取消订阅](/images/zh/manual/olares/wise-unsub-rss-feeds.png#bordered)
