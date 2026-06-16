---
outline: [2, 3]
description: 了解在 Olares 中执行基础文件操作的步骤，如上传、预览、编辑和下载文件。
---
# 基础文件操作

文件管理器的使用方式与其他同类软件类似。本文将介绍一些常用操作，帮助你快速上手。

## 上传文件

### 通过 Files 应用上传
1. 从应用坞或启动台中打开**文件管理器**。  
2. 在左侧边栏中选择要上传文件的目录，例如**文档**。  
3. 使用以下任一方法上传多个文件或文件夹：
   - 将文件从本地文件管理器拖放到文件窗口中。
   - 点击右上角的 <i class="material-symbols-outlined">drive_folder_upload</i>。
   - 在空白区域右键单击，从上下文菜单中选择**上传文件**或**上传文件夹**。

:::info
文件管理器支持断点续传。如果上传被中断，将自动从上次检查点继续上传。
:::

### 通过 LarePass 桌面端上传
:::info 导入 Olares ID
使用 LarePass 桌面端前，需要先通过助记词导入 Olares ID。请确保已经[备份好助记词](../../larepass/back-up-mnemonics.md)。
:::
LarePass 桌面端的文件上传操作与 Files 应用类似，上传的文件会自动与你的 Olares ID 同步。

### 通过 LarePass 移动端上传
你也可以通过 LarePass 移动应用在手机上上传文件或文件夹。
<Tabs>
<template #直接上传>

1. 打开 LarePass 应用并进入**文件**标签页。
2. 选择要上传文件的目录。
3. 点击右下角的 <i class="material-symbols-outlined">add_circle</i> 图标，选择以下上传方式之一：
   - **文件**：从手机存储中选择文件。
   - **图片/视频**：从手机相册中选择图片或视频。
  :::tip
  如果需要整理上传的文件，可以先新建文件夹。
  :::
4. 按照屏幕提示完成上传。
</template>

<template #分享上传>

:::info
具体步骤可能会因操作系统和浏览器而有所不同。
:::

LarePass 支持通过手机的分享选项快速上传文件或媒体内容。
1. 打开文件的分享菜单。
2. 在分享选项中选择 LarePass 图标，或在操作菜单中选择 **LarePass**，跳转到 LarePass 应用。
3. 在 LarePass 应用中选择上传目标位置：
    - **drive**：将文件上传到存储盘，用于个人存储。
    - **sync**：将文件上传到同步盘，用于同步或共享。
4. 根据所选的上传目标位置，按照屏幕提示完成上传。
</template>
</Tabs>

通过 LarePass 移动应用上传的文件也会自动与你的 Olares ID 同步。

## 下载文件
下载多个文件时，文件管理器网页版和 LarePass 桌面端的行为有所不同：
* **文件管理器网页版**：下载任务由浏览器直接管理，可在浏览器下载管理器中暂停、恢复或取消任务。
* **LarePass 桌面端**：下载队列由 LarePass 管理，可暂停、恢复或取消任务，并方便地查找已下载文件。

:::tip 提示
* 文件夹下载仅在 LarePass 桌面端支持。
* 如需下载大文件或批量下载文件，建议使用 LarePass 桌面端，可获得更强大的下载管理功能和更好的使用体验。详情请访问[官方页面](https://olares.cn/larepass)了解和下载。
:::

1. 打开**文件管理器**。  
2. 选中任意文件，右键打开上下文菜单，选择**下载**。  
3. 在弹窗中选择保存位置。  

## 切换显示视图

切换列表视图和网格视图，以不同方式显示文件和文件夹。

1. 打开**文件管理器**，进入目标目录。
2. 点击右上角的 <i class="material-symbols-outlined">dock_to_right</i> 切换为列表视图，或点击 <i class="material-symbols-outlined">grid_view</i> 切换为网格视图。

   ![显示视图](/images/manual/olares/files-display-view.png)

## 排序文件

按不同条件对当前目录中的文件和文件夹进行排序。

1. 打开**文件管理器**，进入目标目录。
2. 排序方式：

   - 列表视图下，点击列标题：

     - **名称**：按文件或文件夹名称字母顺序排序。
     - **大小**：按文件大小排序。
     - **类型**：按文件类型排序。
     - **修改时间**：按最近修改日期排序。

   - 网格视图下，点击左上角的**修改时间**，选择排序选项。

3. 若要反转排序顺序，再次点击当前活动的列标题或排序选项。

## 预览和编辑文件

文件管理器支持多种文件格式的预览和编辑。下表按类别列出了一些典型示例。

### 支持预览的格式

| 类别 | 格式示例 |
|----------|---------|
| 文档 | Word（`.docx`）、Excel（`.xlsx`）、PowerPoint（`.pptx`）、<br>PDF（`.pdf`）、RTF（`.rtf`）、EPUB（`.epub`） |
| 图片 | JPEG（`.jpg`）、PNG（`.png`）、GIF（`.gif`）、BMP（`.bmp`）、<br>SVG（`.svg`）、TIFF（`.tiff`） |
| 媒体 | MP3（`.mp3`）、M4A（`.m4a`）、WAV（`.wav`）、FLAC（`.flac`）、<br>OGG（`.ogg`）、MP4（`.mp4`）、MKV（`.mkv`）、MOV（`.mov`）、<br>AVI（`.avi`）、WebM（`.webm`）、3GP（`.3gp`） |
| 字幕 | SRT（`.srt`）、VTT（`.vtt`）、ASS（`.ass`） |
| 文本与标记 | 文本（`.txt`）、Markdown（`.md`）、HTML（`.html`）、XML（`.xml`） |
| 代码与配置 | Python（`.py`）、JavaScript（`.js`）、Java（`.java`）、<br>C/C++（`.c`/`.cpp`）、Go（`.go`）、Rust（`.rs`）、Shell（`.sh`）、<br>JSON（`.json`）、YAML（`.yaml`）、TOML（`.toml`）、INI（`.ini`） |

### 支持编辑的格式

| 类别 | 格式示例 |
|----------|---------|
| 文本与标记 | 文本（`.txt`）、Markdown（`.md`）、HTML（`.html`）、XML（`.xml`） |
| 代码与配置 | Python（`.py`）、JavaScript（`.js`）、Java（`.java`）、<br>C/C++（`.c`/`.cpp`）、Go（`.go`）、Rust（`.rs`）、Shell（`.sh`）、<br>CSS（`.css`）、JSON（`.json`）、YAML（`.yaml`）、<br>TOML（`.toml`）、INI（`.ini`） |

### 操作步骤

- 若要预览文件，双击目标文件。
- 若要编辑文件：
   1. 双击支持的文本文件。
   2. 点击右上角的 <i class="material-symbols-outlined">edit_square</i>。
   3. 修改内容，然后点击 <i class="material-symbols-outlined">save</i> 保存更改。

      ![编辑文本文件](/images/zh/manual/olares/files-preview-1.png#bordered)

## 搜索文件
通过桌面搜索功能，可以轻松找到**文件管理器**中的文件。
:::tip 提示
**存储盘**（Drive）中的 `/Documents/` 目录支持全文搜索，可搜索文件内容。其他目录则可通过文件名搜索。
:::
1. 点击应用坞中的 <i class="material-symbols-outlined">search</i> 图标打开搜索窗口。
2. 在搜索框中输入要查找的文件相关关键词。
3. 使用方向键 <i class="material-symbols-outlined">arrow_upward</i><i class="material-symbols-outlined">arrow_downward</i> 选择搜索范围：**存储盘**或**同步盘**，按 **Enter** 查看搜索结果。

![搜索](/images/manual/olares/files-search.png#bordered){width="90%"}

## 删除文件
:::warning 警告
删除的文件无法恢复。
:::
1. 从应用坞或启动台中打开**文件管理器**。  
2. 选中要删除的文件，可通过以下方式操作：
   - 右键点击，从上下文菜单中选择**删除**
   - 点击右上角的 <i class="material-symbols-outlined">more_horiz</i>，选择**删除**
3. 在弹窗中确认删除。

## 快捷键
选择多个文件：

- **Windows**：按住 **Control** 并点击目标文件
- **Mac**：按住 **Command** 并点击目标文件
