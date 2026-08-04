---
outline: deep
description: 了解如何在 Olares 上安装、配置和使用 Komga 来管理你的数字媒体库。
head:
  - - meta
    - name: keywords
      content: Olares, Komga, media server, digital library, comics, manga
app_version: "1.0.7"
doc_version: "1.0"
doc_updated: "2026-03-20"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/komga.md)。
:::

# 使用 Komga 构建你的数字图书馆

Komga 是一个专门的开源媒体服务器，旨在管理你的漫画、 manga、Bandes Dessinées (BD)、杂志和电子书的数字收藏。

本指南向你展示如何在 Olares 上安装 Komga、组织媒体文件以进行自动扫描，以及使用内置阅读器和元数据编辑器来增强你的数字阅读体验。

## 学习目标

在本指南中，你将学习如何：
- 安装 Komga 并设置管理员账户。
- 在 Olares Files 应用中准备和组织媒体文件以进行自动检测。
- 创建和配置图书馆以分类内容。
- 阅读书籍并完善其元数据。

## 安装 Komga

1. 打开 Olares Market 并搜索 "Komga"。

   ![从 Market 搜索 Komga](/images/manual/use-cases/install-komga.png#bordered){width=90%}

2. 点击 **获取**，然后点击 **安装**。等待安装完成。

## 设置管理员账户

首次打开 Komga 时，你必须注册一个账户。这是管理员账户，可让你完全控制服务器设置、图书馆和用户管理。

1. 打开 Komga。

   ![打开 Komga](/images/manual/use-cases/open-komga.png#bordered){width=50%}

2. 输入电子邮件和密码以创建你的主管理员账户。
3. （可选）从 **翻译** 列表中选择你的首选语言。
4. （可选）从 **主题** 列表中选择你的首选背景颜色模式。
5. 点击 **创建用户账户**。**欢迎使用 Komga** 页面出现。

   ![Komga 欢迎页面](/images/manual/use-cases/komga-welcome.png#bordered)

## 准备你的媒体文件

Komga 扫描专用的 Olares 目录来填充你的图书馆。你必须将媒体文件放在指定文件夹中，以便应用检测它们。

1. 从 Launchpad 打开 Files 应用。
2. 导航到 `Application/Data/komga/data/`。
3. 将媒体文件上传到此目录，或在其下创建子文件夹来分类文件。

   ![Komga 数据目录](/images/manual/use-cases/komga-data-path.png#bordered)

## 创建图书馆

上传文件后，通过创建图书馆将它们连接到 Komga 界面。图书馆是一组书籍。你可以创建多个图书馆来分离不同类型的内容。

1. 在 **Komga** 中，点击主屏幕上的 **添加图书馆**，或点击左侧边栏 **图书馆** 旁边的 <i class="material-symbols-outlined">add_2</i>。**添加图书馆** 窗口出现。

   ![添加图书馆菜单](/images/manual/use-cases/add-lib-menu.png#bordered)

2. 在 **常规** 标签页上，配置以下设置，然后点击 **下一步**：
   - **名称**：指定图书馆的名称。
   - **根文件夹**：点击 **浏览** 选择文件位置。
       - 要将所有媒体文件和子文件夹包含在一个大型图书馆中，选择 `/data` 文件夹。
       - 要将图书馆内容限制为特定类别，选择子文件夹，如 `/data/{子文件夹名称}`。

   ![Komga 常规设置](/images/manual/use-cases/komga-general.png#bordered){width=60%}

3. 在 **扫描器** 标签页上，配置 Komga 如何识别媒体文件，如扫描间隔和文件类型。点击 **下一步**。
4. 在 **选项** 标签页上，设置文件分析和封面图像生成的首选项。点击 **下一步**。
5. 在 **元数据** 标签页上，选择要导入的元数据类型，然后点击 **添加**。

   Komga 自动扫描关联目录并显示你的书籍及其封面图像和标题。

   ![添加图书馆完成](/images/manual/use-cases/komga-library-added.png#bordered){width=90%}

## 编辑元数据

Komga 自动从媒体文件中提取嵌入的元数据，如摘要和发布日期。但你可以手动完善这些数据，以美化图书馆的外观并使其更易于浏览。

1. 将鼠标悬停在书籍封面上，然后点击 <i class="material-symbols-outlined">edit</i>。
2. 在 **编辑元数据** 窗口中，使用以下标签页自定义内容：

   | 标签页 | 功能 |
   |:---|:----------------|
   | 常规 | 定义书籍在图书馆列表中的显示方式：<ul><li>输入 **标题** 和 **摘要** 为你的收藏提供上下文。</li><li>使用 **编号** 设置系列中书籍的卷号。<br> 例如，第一本书用 "1"，第二本用 "2"。</li><li>使用 **排序编号** 指定书籍在系列列表中的显示顺序，<br>不受标题或日期影响。</li><li>输入 **发布日期** 和 **ISBN** 以跟踪收藏历史。</li></ul>|
   | 作者 | 按角色添加贡献者，如 **作者**、**铅笔稿** 和 **墨水稿**。<br>这允许你按特定艺术家筛选图书馆。|
   | 标签 | 创建自定义标签以按主题分组书籍，如 "90 年代美学"。<br>这允许你基于标签搜索。 |
   | 链接 | 点击加号图标添加外部 URL，如官方系列网站。<br>这允许快速访问有关书籍的更多信息。 |
   | 海报 | 拖放新图像或选择文件以替换默认书籍封面。<br>这允许你个性化数字书架的外观。 |

## 阅读书籍

图书馆组织好后，你可以开始使用内置 Webreader 阅读。

1. 找到目标书籍。
2. 使用以下方法之一打开书籍：
   - 将鼠标悬停在书籍封面上并点击 <i class="material-symbols-outlined">auto_stories</i>。
   - 点击书籍标题，然后在书籍详情页面上点击 **阅读**。

   ![阅读书籍按钮](/images/manual/use-cases/komga-read-btn.png#bordered){width=90%}

   书籍立即在 Webreader 中打开。

   ![在 Webreader 中打开的书籍](/images/manual/use-cases/komga-book-opened.png#bordered){width=90%}

## 扫描图书馆文件

默认情况下，Komga 根据你在创建图书馆时设置的间隔定期扫描图书馆。但是，如果你添加、重命名或移动媒体文件并希望它们立即出现，可以触发手动扫描。

1. 从左侧边栏找到目标图书馆。
2. 点击其旁边的 <i class="material-symbols-outlined">more_vert</i>，然后选择 **扫描图书馆文件**。图书馆将刷新并拉入任何新内容或更改的内容。

   ![扫描图书馆文件](/images/manual/use-cases/komga-scan-lib-files.png#bordered){width=50%}

## 常见问题

### 如何执行彻底卸载？

如果你卸载 Komga 并希望删除所有剩余数据：
1. 打开 Files 应用，前往 **应用** > **数据**
2. 删除 `komga` 文件夹。这将删除所有数据库配置和缓存的元数据。

### 为什么我无法使用 Komga 中的 "导入" 功能？

导入功能旨在将文件从图书馆外部位置（如 `Downloads` 文件夹）移动或复制到现有系列中。在 Olares 环境中，出于安全和隐私考虑，文件选择器限制在 `/data` 目录。由于你的图书馆已经位于此目录中，系统没有外部文件可检测，因此导入工具无法使用。

要添加新文件，请使用 Files 应用将媒体文件上传到：`Application/Data/komga/data/`。

上传后，Komga 会在下一次计划扫描期间检测它们，或者你可以通过选择 **扫描图书馆文件** 触发手动扫描。
