---
outline: [2, 3]
description: 了解如何在 Olares One 上使用 Files 应用。学习界面布局、如何上传文件、预览和编辑内容、搜索文件，以及如何连接 Google Drive 等外部云存储。
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, 上传文件, 管理文件, 访问云文件, Olares Files 应用
---

# 管理文件 <Badge type="tip" text="15 min" />

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../one/files.md)为准。
:::

Olares 内置的 Files 应用为文件存储和同步提供了安全高效的解决方案。它允许你随时随地访问个人文件，在设备间保持数据同步，并在一个地方管理第三方云存储。

Files 中的操作与其他文件管理器基本相同。本指南介绍 Files 中的一些常见任务，帮助你快速上手。

## 学习目标

- 了解 Files 界面。
- 从本地计算机上传文件。本指南以上传 `.pdf` 文档和 `.txt` 文件为例。
- 预览和编辑支持的文件类型。
- 使用系统搜索定位文件。

## 开始之前

### 界面概览

熟悉工作区。Files 界面主要由以下区域组成：

* Drive：存储你的个人文件，如文档和图片。
* Sync：基于库的存储区域，用于需要频繁修改或实时同步的文件。
* Application：为应用特定数据保留。主要用于开发和调试，非一般文件存储用途。
* Cloud storage：访问来自 Google Drive 和 AWS S3 等集成服务的文件。
* Share：管理共享文件夹的集中中心。

![Files](/images/zh/manual/olares/files-index-3.png#bordered)

### 支持的文件格式

文件管理器支持多种文件格式的预览和编辑。下表按类别列出了一些典型示例。

#### 支持预览的格式

| 类别 | 格式示例 |
|----------|---------|
| 文档 | Word（`.docx`）、Excel（`.xlsx`）、PowerPoint（`.pptx`）、<br>PDF（`.pdf`）、RTF（`.rtf`）、EPUB（`.epub`） |
| 图片 | JPEG（`.jpg`）、PNG（`.png`）、GIF（`.gif`）、BMP（`.bmp`）、<br>SVG（`.svg`）、TIFF（`.tiff`） |
| 媒体 | MP3（`.mp3`）、M4A（`.m4a`）、WAV（`.wav`）、FLAC（`.flac`）、<br>OGG（`.ogg`）、MP4（`.mp4`）、MKV（`.mkv`）、MOV（`.mov`）、<br>AVI（`.avi`）、WebM（`.webm`）、3GP（`.3gp`） |
| 字幕 | SRT（`.srt`）、VTT（`.vtt`）、ASS（`.ass`） |
| 文本与标记 | 文本（`.txt`）、Markdown（`.md`）、HTML（`.html`）、XML（`.xml`） |
| 代码与配置 | Python（`.py`）、JavaScript（`.js`）、Java（`.java`）、<br>C/C++（`.c`/`.cpp`）、Go（`.go`）、Rust（`.rs`）、Shell（`.sh`）、<br>JSON（`.json`）、YAML（`.yaml`）、TOML（`.toml`）、INI（`.ini`） |

#### 支持编辑的格式

| 类别 | 格式示例 |
|----------|---------|
| 文本与标记 | 文本（`.txt`）、Markdown（`.md`）、HTML（`.html`）、XML（`.xml`） |
| 代码与配置 | Python（`.py`）、JavaScript（`.js`）、Java（`.java`）、<br>C/C++（`.c`/`.cpp`）、Go（`.go`）、Rust（`.rs`）、Shell（`.sh`）、<br>CSS（`.css`）、JSON（`.json`）、YAML（`.yaml`）、<br>TOML（`.toml`）、INI（`.ini`） |

## 上传文件

通过 Files 应用上传文档，为你的服务器填充内容。

1. 从 Dock 或 Launchpad 打开 Files 应用。

    ![查找并打开 Files 应用](/images/one/find-files-app.png#bordered)

2. 在左侧边栏中，选择你要上传文件的目录。例如，**Documents**。

    ![选择上传目标](/images/one/upload-file.png#bordered)

3. 使用以下方法之一上传 PDF 文档和文本文件：
   - 从本地文件管理器将文件拖放到 Files 窗口中。
   - 点击右上角 <i class="material-symbols-outlined">drive_folder_upload</i>。
   - 在空白处右键单击并从上下文菜单中选择。

    :::info 上传进度指示器
    上传期间，右下角的通知会显示上传状态。
    :::

4. 上传完成后，文件会显示在你选择的目录中。

     ![文件上传完成](/images/one/upload-file-completed.png#bordered)

:::info 可恢复上传
如果你的互联网连接断开，上传会自动停止。当连接恢复后，Olares 会从断点处继续上传，无需从头开始。
:::

## 预览和编辑文件

你可以直接在 Files 应用中查看和修改许多常见文件类型，无需下载。

- 要预览 PDF 文档，双击它。

    ![查找 Files 应用](/images/one/preview-file.png#bordered)

- 要编辑文本文件：
   
   1. 双击它。

    ![预览文本文件](/images/one/open-txt.png#bordered)

   2. 点击右上角 <i class="material-symbols-outlined">edit_square</i>。
   3. 修改内容，然后点击 <i class="material-symbols-outlined">save</i> 保存更改。
    
    ![编辑并保存文本文件](/images/one/edit-save-file.png#bordered)

## 搜索文件

使用全局桌面搜索快速定位文件。

1. 点击 Dock 中的 <i class="material-symbols-outlined">search</i> 打开搜索窗口。

    ![搜索文件](/images/one/search-dock.png#bordered)

2. 选择 **Drive** 作为搜索范围。

    ![选择 drive](/images/one/select-drive.png#bordered)

3. 在 **搜索文件** 字段中，输入与你要查找的文件相关的关键词。例如，`txt`。结果会列出。

    ![搜索 txt 文件](/images/one/search-txt.png#bordered)

4. 双击即可直接在 Files 应用中打开。

## 下载文件

将文件下载到本地计算机。

1. 从 Dock 或 Launchpad 打开 Files 应用。
2. 右键单击目标文件，然后选择 **下载**。

    文件通过你的 Web 浏览器标准下载管理器下载。

:::info
Web 浏览器无法可靠地下载整个文件夹或非常大的文件。对于这些任务，请安装 LarePass Desktop 应用以解锁文件夹下载和断点续传功能。
:::

## 删除文件

:::warning
删除的文件无法恢复。
:::

1. 从 Dock 或 Launchpad 打开 Files 应用。
2. 右键单击目标文件，然后点击 **删除**。
3. 在 **删除** 窗口中，点击 **确认**。

## 资源

- [基本文件操作](../manual/olares/files/add-edit-download.md)
- [共享文件](../manual/olares/files/share-files.md)
- [同步文件到本地](../manual/olares/files/sync-files.md)
- [挂载 SMB](../manual/olares/files/mount-SMB.md)
- [挂载云存储](../manual/olares/files/mount-cloud-storage.md)
- [LarePass 官方页面](https://www.olares.com/larepass)
