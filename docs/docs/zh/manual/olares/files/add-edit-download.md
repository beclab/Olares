---
description: 在 Olares 中执行基础文件操作，包括添加、编辑和下载文件的具体步骤。
---
# 添加、编辑和下载
**文件管理器**使用方式与其他同类软件类似。本文将介绍一些常用操作，帮助你快速上手。

## 上传文件

### 通过 Files 应用上传
1. 打开**文件管理器**。
2. 在左侧边栏中选择要上传文件的目录，例如"文档"。
3. 使用以下任一方法上传多个文件或文件夹：
   - 将文件从本地文件管理器拖放到文件窗口中。
   - 右上角点击 **<i class="material-symbols-outlined">drive_folder_upload</i>上传**。
   - 在空白区域右键单击，在上下文菜单中选择**上传文件**或**上传文件夹**。

:::tip
文件支持断点续传。如果上传被中断，将自动从上次检查点继续上传。
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
   - **文件**：从手机存储中选择文件
   - **图片/视频**：从手机相册中选择图片或视频
  :::tip
  如果需要整理上传的文件，可以先选择**新建文件夹**。
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
* **文件管理器网页版**：下载任务由你的浏览器管理，可使用浏览器下载功能管理下载队列，可暂停、恢复或取消任务。
* **LarePass 桌面端**：支持在 LarePass 里管理下载队列，可暂停、恢复或取消任务，方便查找已下载文件。

:::tip 提示
* 文件夹下载仅在 LarePass 桌面版支持。
* 如需下载大文件或批量下载文件，建议使用 LarePass 桌面端，可获得更强大的下载管理功能和更好的使用体验。详情请访问[官方页面](https://olares.cn/larepass)了解和下载。
:::

1. 打开**文件管理器**。
2. 选中任意文件，右键打开上下文菜单，选择**下载**。
3. 在弹窗中选择保存位置。

## 预览和编辑文件
双击文件即可打开预览。**文件管理器**支持预览以下格式：

* **图片**：JPG、JPEG、PNG、BMP、WEBP、SVG
* **视频**：MP4、MKV、AVI、MOV、MPEG、MTS、TS、WMV、WEBM、RM、3GP
* **音频**：MP3、WMA、WAV、OGG、AAC、M4A、APE、FLAC
* **文本**：PDF、TXT、JS、CSS、XML、YAML、HTML

**文件管理器**还支持编辑以下文本格式：TXT、JS、CSS、XML、YAML、HTML。

![预览](/images/manual/olares/files-preview.png#bordered)
## 搜索文件
通过桌面搜索功能，可以轻松找到**文件管理器**中的文件。
:::tip 提示
**存储盘**（Drive）中的 `/Documents/` 目录支持全文搜索，可搜索文件内容。其他目录则可通过文件名搜索。
:::
1. 点击 Dock 中的<i class="material-symbols-outlined">search</i>图标打开搜索窗口。
2. 在搜索框中输入要查找的文件相关关键词。
3. 使用方向键<i class="material-symbols-outlined">arrow_upward</i><i class="material-symbols-outlined">arrow_downward</i>选择搜索范围：**存储盘**或**同步盘**，按 **Enter** 查看搜索结果。

![搜索](/images/manual/olares/files-search.png#bordered){width="90%""}
## 删除文件
:::warning 警告
删除的文件无法恢复。
:::
1. 打开**文件管理器**。
2. 选中要删除的文件，可通过以下方式操作：
   - 右键点击，从上下文菜单中选择**删除**
   - 右上角点击 **<i class="material-symbols-outlined">more_horiz</i>更多**，选择**删除**
3. 在弹窗中确认删除。

## 切换显示视图

可在列表视图和网格视图之间切换，以不同方式显示文件和文件夹。

![显示视图](/images/manual/olares/files-display-view.png)
## 快捷键
选择多个文件：

* **Windows**：按住 **Control** 选择目标文件
* **Mac**：按住 **Command** 选择点击目标文件