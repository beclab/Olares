---
outline: [2, 3]
description: 了解 Olares 系统应用 Wise 的基本操作。学习收集、阅读和组织内容，使用标签和笔记功能，追踪阅读进度，实现个性化阅读体验。
---
# Wise 基本操作

Wise 助你搭建专注、高效的阅读环境。本文将带你了解日常核心操作，帮你轻松收集内容，安排阅读计划，并记录自己的想法。

## 开始之前

为充分体验 Wise 的能力，建议在 Olares 应用市场搜索并安装以下配套应用：

- **Rss Subscribe**：用于在浏览网页时直接订阅 RSS 源。
- **YT-DLP**：用于将网页上的音频和视频下载到 Wise 中。

即使不安装这些应用，你也可以正常使用 Wise，但无法使用浏览器内订阅和媒体下载功能。

## 打造你的内容库

Wise 通过以下两种方式将内容收集到库中：

- **保存内容**：包括你主动保存的页面、文件、音频或视频。这些文件会出现在主界面的**收件箱**以及文件类型对应文件夹中（**文章**、**视频**、**音频**、**图书**、**PDF**），其状态为**收件箱**或**稍后阅读**。
- **订阅源**：包括你订阅的网站、博客、播客或频道。新的内容会出现在**订阅源**里，你可以从中挑选感兴趣的内容保存到自己的库里。

本文重点介绍如何保存和阅读单条内容。如需了解订阅源和订阅管理，请查看[订阅和管理订阅源](./subscribe)。

### 保存内容

你可以通过以下三种方式将内容保存到 Wise：上传文件、在 Wise 内部添加链接，或使用 LarePass 浏览器扩展直接保存。

#### 上传文件

直接从你的电脑导入文件，包括 PDF、EPUB、音频、视频和其他文件类型。Wise 会根据文件格式，自动将其放入库中对应的内容文件夹。例如，PDF 放入 **PDF**，EPUB 放入**图书**，媒体文件放入**视频**或**音频**。

1. 点击左下角菜单栏的 <i class="material-symbols-outlined">add_circle</i>。  
2. 选择**上传**。
3. 从本地电脑选择一个或多个文件。

#### 通过链接添加

粘贴 URL 来保存文章、视频或订阅源。

::: tip 上传 Cookie
如果链接指向的内容需要登录或其他访问控制，Wise 可能需要 Cookie 才能正确获取内容。若要配置 Cookie，请查看[管理 Cookie](./manage-cookies)。
:::

1. 点击左下角菜单栏的 <i class="material-symbols-outlined">add_circle</i>。
2. 选择**添加链接**。
3. 粘贴或输入 URL。

   Wise 会解析链接，并列出所有可行操作：
    - **保存到库**：将内容保存为库中的一个文件，并添加至**收件箱**。
    - **订阅 RSS 源**：如果 Wise 检测到站点有一或多个 RSS 源，会在此列出。选择你想要订阅的源，之后该订阅源下的更新会自动出现在[**订阅源**](./subscribe)中。
    ![订阅 RSS 源](/images/zh/manual/olares/wise-add-link-subscribe.png#bordered){width=300}
    - **下载文件**：如果 Wise 检测到页面上有视频或其他可下载文件，会出现此选项。选择你要下载的文件，将其保存以便离线访问。使用此功能需安装[YT-DLP](https://market.olares.com/app/market.olares/ytdlp)。

    ![Download files](/images/zh/manual/olares/wise-add-link-download.png#bordered){width=300}

新保存的内容会出现在对应类型文件夹下。之后你可以将其状态设定为**收件箱**或**稍后阅读**。

#### 使用 LarePass 扩展

你也可以在浏览器中通过 LarePass 扩展直接将内容保存到 Wise，无需打开 Wise 应用。

:::tip 安装 LarePass 浏览器扩展
LarePass 扩展目前仅支持 Chrome 浏览器。你可以从[官网](https://www.olares.cn/larepass)下载。
:::

将网页内容保存到 Wise：

1. 打开 LarePass 浏览器扩展并点击“收集”。
2. 在**保存到库**区域查看当前页面检测到的内容。
3. 点击目标内容右侧的<i class="material-symbols-outlined">box_add</i>。
![Save content via LarePass extension](/images/zh/manual/olares/wise-larepass-add-to-lib.png#bordered)

   通过 LarePass 扩展保存的内容会添加到内容库中，同时出现在主界面**收件箱**以及对应的内容类型文件夹中。

### 监控和管理媒体下载

当你添加新的音频或视频内容时，Wise 会自动创建下载任务并将媒体文件保存到 Olares。这可以：

- 确保你能离线访问媒体。
- 在库中保留副本，即使原始来源被删除也不受影响。
- 加快打开和播放内容的速度。

要管理所有下载任务：

1. 进入 **<i class="material-symbols-outlined">settings</i> 设置** > **传输** > **下载**.
2. 查看媒体列表及其状态。
3. 你可以：

    - 点击<i class="material-symbols-outlined">folder_open</i>在 Olares 文件管理器中定位文件。
    - 点击<i class="material-symbols-outlined">do_not_disturb_on</i>从列表中移除文件。

## 管理内容状态

你可以将内容标记为<b>`收件箱`</b>或<b>`稍后阅读`</b>来组织内容。在列表视图或阅读视图中，可以通过相应图标更改其状态。

![Wise content status](/images/zh/manual/olares/wise-item-status.png#bordered)
![Wise content status in list](/images/zh/manual/olares/wise-list-status.png#bordered)

- **保留在待办列表**  
  点击<i class="material-symbols-outlined">inbox</i>：该内容将被视为活跃内容，出现在主界面**收件箱**以及文件类型对应文件夹的<b>`收件箱`</b>页签（例如**文章** > **`收件箱`**）。

- **归档到稍后阅读**  
  点击<i class="material-symbols-outlined">schedule</i>：该内容将从主界面**收件箱**和文件类型对应文件夹的<b>`收件箱`</b>页签移出，出现在各自的<b>`稍后阅读`</b>页签（例如**文章** > **`稍后阅读`**）。

::: tip 
内容状态更改后，其所在位置会立即更新。你可以随时在**收件箱**和**稍后阅读**之间切换。 
:::

## 使用阅读工具

Wise 提供多种阅读辅助工具，帮助你专注阅读并记录重要内容。

![Wise reading toolbar](/images/zh/manual/olares/wise-reading-toolbar.png#bordered)

### 追踪阅读进度

Wise 会在文章封面上用绿色圆点标记未读内容。当你打开文章，Wise 会自动将其标记为已读。

![Wise unseen content](/images/zh/manual/olares/wise-unseen-content.png#bordered){width=600}

在阅读器工具栏中，你也可以手动切换 <i class="material-symbols-outlined">playlist_add_check</i>（已读）或 <i class="material-symbols-outlined">playlist_remove</i>（未读），维护你的阅读进度。

### 记录笔记

你可以为库中的所有内容添加私人笔记：

1. 浏览时，点击 <i class="material-symbols-outlined" style="font-variation-settings: 'wght' 200;">right_panel_open</i> 打开**信息**面板。
2. 在**笔记**区域输入你的想法。
3. 点击**保存**。

你可以随时打开该面板，编辑或删除笔记。

### 使用标签

你可以使用标签，灵活标记内容，以便日后轻松检索。

1. 在列表页，点击内容卡片上的<i class="material-symbols-outlined" style="font-variation-settings: 'wght' 200;">sell</i> 为其添加标签。
2. 选择现有标签，或输入新名称创建标签

![Tags](/images/zh/manual/olares/wise-tags.png#bordered){width=600}

你可以在<b><i class="material-symbols-outlined">settings</i>设置</b> > **标签**下统一管理所有标签。

::: tip
标签与过滤视图结合使用，效果更佳。详情请见[使用过滤视图管理内容](./filter)。
:::

## 搜索内容库

当你在 Wise 中收集了大量内容后，可以借助 Olares 的聚合搜索功能，在 Wise 中按主题或具体内容进行搜索。

1. 点击应用坞的<i class="material-symbols-outlined">search</i>打开搜索窗口。
2. 将搜索范围限定为 Wise，然后输入关键词开始搜索。
![Search in Wise](/images/manual/tutorials/wise-search.png#bordered)

## 自定义外观

默认情况下，Wise 会跟随系统的明暗主题设置。你也可以单独为 Wise 指定显示风格：

1. 点击左下角的<i class="material-symbols-outlined">settings</i>，选择**偏好设置**。
2. 在**主题**下，选择你偏好的主题模式：
   - 浅色
   - 暗黑

你的选择会应用到 Wise 界面和阅读器中。  