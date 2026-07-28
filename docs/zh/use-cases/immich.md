---
outline: deep
description: 在 Olares 上自托管 Immich，作为开源的 Google Photos 替代方案。用 AI 备份、整理和搜索照片与视频，数据始终留在你自己的设备上。
head:
  - - meta
    - name: keywords
      content: Olares, Immich, google photos alternative, immich vs google photos, self-hosted photos, photo backup, immich synology, immich on olares
app_version: "1.0.15"
doc_version: "1.0"
doc_updated: "2026-03-26"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/immich.md)为准。
:::

# 用 Immich 自托管你的照片

Immich 是一个开源、自托管的照片和视频备份解决方案。它支持自动备份、原始质量存储和时间线浏览。借助内置的机器学习模型，Immich 可以自动识别照片中的人物、地点和物体，让照片管理更智能、更高效。

在 Olares 上运行 Immich 可以获得 Google Photos 般的体验，同时完全掌控你的数据。结合其移动应用，它是个人或家庭构建私人媒体库的理想选择。

## 学习目标

在本指南中，你将学习如何：
- 安装 Immich 并设置管理员账户。
- 通过网页上传、移动备份和外部导入来填充你的库。
- 浏览、编辑和管理你的照片时间线。
- 使用 AI 驱动的智能搜索和人脸识别。
- 在本地和公开地与他人分享照片和相册。

## 安装 Immich

1. 打开 Olares Market，搜索 "Immich"。

   ![从 Market 搜索 Immich](/images/manual/use-cases/immich.png#bordered)

2. 点击 **Get**，然后点击 **Install**。等待安装完成。

## 设置管理员账户

第一个注册的用户将成为管理员，负责管理实例和添加其他用户。

1. 从 Launchpad 打开 Immich，然后点击 **Getting Started**。

   ![Immich 欢迎页面](/images/manual/use-cases/immich-welcome.png#bordered)

2. 在 **Admin Registration** 页面上，设置管理员邮箱、密码和用户名。

   ![注册管理员账户](/images/manual/use-cases/immich-admin-registration.png#bordered){width=50%}

3. 使用新凭据登录。你将进入 **Photos** 页面，该页面以时间线视图显示所有照片。

## 填充你的库

通过多种来源将照片导入 Immich，构建你的库。

### 从电脑上传

添加存储在你电脑上的照片。

1. 在 Immich 网页 UI 的 **Photos** 页面上，点击中央的上传区域，或点击右上角的 **Upload**。

   ![上传照片到 Immich](/images/manual/use-cases/immich-upload-photos.png#bordered)
2. 选择要上传的照片。

   上传后，它们将按日期自动排序显示在照片时间线上。

   ![照片已上传到 Immich](/images/manual/use-cases/immich-photos-uploaded.png#bordered)

### 从移动设备上传

使用 Immich 移动应用将移动设备上的照片上传到 Immich 服务器，创建私人备份。

1. 安装 Immich 移动应用。
   - iOS：从 App Store 下载 "Immich"。
   - Android：从 [Immich GitHub Releases](https://github.com/immich-app/immich/releases) 页面下载 APK，或从 [Google Play](https://play.google.com/store/apps/details?id=app.alextran.immich) 安装。

2. 在你的移动设备上打开 LarePass 并启用 VPN，以确保与 Olares 的安全连接。

      ![启用 LarePass VPN](/images/manual/use-cases/alex-larepass-vpn-mobile.png#bordered)

3. 打开 Immich 移动应用，然后输入你的 Immich 服务器 URL、管理员邮箱和密码以登录。你从网页 UI 上传的照片将显示出来。

   :::tip 获取 Immich 服务器 URL
   打开 Settings，前往 **Applications** > **Immich** > **Entrances** > **Immich** > **Endpoint settings**，然后复制 Endpoint 地址。这就是你的 Immich 服务器地址。
      ![获取 Immich 服务器 URL](/images/manual/use-cases/immich-endpoint.png#bordered){width=70%}
   :::

4. 点击右上角的 <i class="material-symbols-outlined">backup</i>。

   ![开始移动备份](/images/manual/use-cases/immich-mobile-backup.png#bordered){width=90%}

5. 在 **Backup Albums** 中，点击 **Select**，然后选择要上传的文件夹。

   ![选择移动备份的文件](/images/manual/use-cases/immich-mobile-backup-select-file.png#bordered){width=90%}

6. 打开 **Enable Backup**。

   ![在移动设备上启用备份](/images/manual/use-cases/immich-enable-backup.png#bordered){width=90%}

   完成后，上传的照片将显示在 Immich 网页 UI 的照片时间线中。移动设备上所选文件夹中的任何新照片都将自动同步到网页 UI。

### 从 Olares Files 上传

如果你有存储在 Olares Files 应用中的照片，你可以将 Immich 配置为扫描该文件夹作为外部库。然后 Immich 将索引并显示它们，而无需移动实际文件。

1. 在 Immich 网页 UI 中，点击右上角的用户头像，然后选择 **Administration**。
2. 选择 **External Libraries** > **Create Library**。
3. 将 **Owner** 设为你的管理员账户，然后点击 **Create**。
4. 在 **Folders** 区域中，点击 **Add**。

   ![创建外部库](/images/manual/use-cases/immich-external-libraries.png#bordered)

5. 输入导入路径，该路径区分大小写，然后点击 **Add**。例如，

   ```text
   /home/Pictures
   ```
6. 点击 **Scan**。Immich 将在照片时间线中显示此目录下的照片。

   ![扫描外部库](/images/manual/use-cases/immich-scan-library.png#bordered)

### 从 NAS 导入

如果你有大量照片存储在本地 NAS 设备上，你可以直接在 Olares 中挂载你的 NAS，并将这些目录映射到 Immich 中，而无需复制文件。

有关设置 SMB 连接和映射路径的详细步骤，请参阅[从 NAS 导入照片](./immich-import-from-nas.md)。

## 浏览和管理照片

库填充完成后，你现在可以与照片进行交互。

### 查看照片

1. 点击时间线上的目标照片以打开预览。
2. 要查看照片的详细信息，点击右上角的 <i class="material-symbols-outlined">info</i>。

   拍摄日期、相机型号和文件格式等元数据将显示在右侧。

   ![查看照片详细信息和元数据](/images/manual/use-cases/immich-photo-details.png#bordered)

### 编辑照片

1. 打开照片预览，然后点击右上角的 <i class="material-symbols-outlined">tune</i>。
2. 使用 **Editor** 面板裁剪、旋转或镜像照片。

   :::info
   Immich 支持照片的非破坏性编辑。这意味着你对素材所做的任何编辑都不会修改原始文件，而是创建一个应用了编辑的新版本素材。
   :::
3. 点击 **Save**。
4. 要随时恢复为原始照片，点击同一面板中的 **Reset changes**。

### 收藏照片

1. 将鼠标悬停在时间线上的目标照片上，然后点击左上角的 <i class="material-symbols-outlined">check_circle</i> 以选择它。
2. 点击 <i class="material-symbols-outlined">favorite</i>。该照片将被添加到侧边栏中的 **Favorites**。

   ![收藏照片](/images/manual/use-cases/immich-favorite-photo.png#bordered)

### 删除照片

1. 将鼠标悬停在时间线上的目标照片上，然后点击左上角的 <i class="material-symbols-outlined">check_circle</i> 以选择它。
2. 点击右上角的 <i class="material-symbols-outlined">more_vert</i>，然后选择 **Delete**。删除的照片将被移动到侧边栏中的 **Trash**，并将在 30 天后永久删除。
3. 要立即永久删除它，点击左侧边栏中的 **Trash**，选择照片，然后点击 <i class="material-symbols-outlined">delete_forever</i>。

   ![永久删除照片](/images/manual/use-cases/immich-trash-delete.png#bordered)

### 恢复照片

1. 点击左侧边栏中的 **Trash**。
2. 将鼠标悬停在照片上，然后点击左上角的 <i class="material-symbols-outlined">check_circle</i> 以选择它。
3. 点击 **Restore**。

   ![从 Trash 恢复照片](/images/manual/use-cases/immich-trash-restore.png#bordered)

### 下载照片

1. 将鼠标悬停在时间线上的目标照片上，然后点击左上角的 <i class="material-symbols-outlined">check_circle</i> 以选择它。
2. 点击右上角的 <i class="material-symbols-outlined">more_vert</i>，然后选择 **Download**。

   照片将保存到你的电脑。选择多张照片时，它们将被打包成 `.zip` 文件以进行批量下载。

   ![下载照片](/images/manual/use-cases/immich-download-photo.png#bordered)

## 整理照片

将你的媒体分组到精心策划的集合中，或使用传统的文件资源管理器结构浏览你的库。

### 创建相册

将照片分组到主题集合中，以便更轻松地访问和分享。

1. 点击左侧边栏中的 **Albums**，然后点击页面中央的相册创建区域，或点击页面顶部的 **Create album**。

   ![创建新相册](/images/manual/use-cases/immich-create-album.png#bordered)

2. 设置相册名称和描述，然后点击 **Select photos**。
3. 要添加现有照片，选择照片，然后点击 **Add assets**。
4. 要从本地设备添加照片，点击 **Select from computer**。

   新相册将出现在左侧边栏中的 **Albums** 列表中。

### 按文件结构整理

启用文件夹视图，使用 Olares 文件系统的原始目录层次结构来导航你的媒体文件。

1. 点击你的头像，然后点击 **Account Settings**。
2. 展开 **Features** > **Folders**，然后启用它。

   ![启用文件夹视图](/images/manual/use-cases/immich-enable-folder-view.png#bordered){width=65%}
3. 点击 **Save**。左侧边栏中将显示一个 **Folders** 节点。
4. 点击 **Folders**。你可以看到文件以类似文件资源管理器的视图进行组织。

   ![文件夹视图](/images/manual/use-cases/immich-folders-view.png#bordered)

## 搜索照片

Immich 使用内置的 AI 模型分析你的图像内容，提供灵活的搜索体验。

### 按上下文搜索

使用自然语言搜索人物、地点和物体，无需依赖照片文件元数据中的关键词。

1. 在页面顶部的搜索栏中输入搜索词。例如，`halloween`。

   ![照片智能搜索](/images/manual/use-cases/immich-smart-search.png#bordered)

2. Immich 根据视觉内容识别相关图像，即使它们没有手动标签。

   ![智能搜索结果](/images/manual/use-cases/immich-smart-search-result.png#bordered)

### 按人脸搜索

Immich 识别你照片和视频中的人脸，并将它们分组到 **Explore** 页面的 **People** 中。你可以为这些人分配名称并搜索他们。

1. 点击左侧边栏中的 **Explore**，查看按人自动分组的人脸。

   ![探索人脸识别结果](/images/manual/use-cases/immich-face-recognition.png#bordered)

2. 点击一个人脸组，然后在 **Add a name** 字段中输入一个名称以标记此人脸。

   ![为人脸组命名](/images/manual/use-cases/immich-face-group-name.png#bordered)

3. 命名后，你可以直接在搜索栏中搜索该人。

### 按位置搜索

Immich 根据 GPS 元数据将你的媒体聚类，因此你可以通过导航到全球地图上的特定区域来搜索。

1. 选择左侧边栏中的 **Map**，以全局视图查看你的照片。
2. 使用缩放控件或滚动到特定国家或城市。
3. 选择一个蓝色位置聚类，以查看与该区域关联的照片。

   ![地图视图](/images/manual/use-cases/immich-map-view.png#bordered)

## 分享照片

Immich 支持两种分享类型：面向外部收件人的公开链接，以及面向同一 Olares 集群上用户的本地协作相册。

### 与外部用户分享

创建安全的公开链接，与 Olares 网络之外的人分享照片。

1. 打开 Immich 网页界面，从 **Photos** 页面中选择你要分享的照片。
2. 点击右上角的 <i class="material-symbols-outlined">share</i>。
3. 根据需要指定分享链接的设置，例如 URL、描述、访问密码和过期日期。
4. 点击 **Create link**。Immich 将生成一个分享链接和二维码，其他人可以使用它们查看分享的照片。

   ![创建照片分享链接](/images/manual/use-cases/immich-share-link.png#bordered){width=40%}

### 与本地成员分享

与你的 Olares 实例上的其他用户协作共享相册。

1. 添加用户。

   a. 点击你的用户头像，然后点击 **Administration**。

   b. 点击左侧边栏中的 **Users**，然后点击 **Create user** 为家人或朋友创建账户。

      ![创建分享用户](/images/manual/use-cases/immich-share-create-user.png#bordered){width=40%}

   c. 根据需要输入新用户账户的相关信息，然后点击 **Create**。

   d. 通过点击左上角的 immich 返回主页。

2. 邀请用户加入相册。

   a. 从左侧边栏中选择相册，然后点击右上角的 <i class="material-symbols-outlined">share</i>。

   b. 在 **Options** 窗口中，点击 **Invite People**，选择你要分享相册的用户，然后点击 **Add**。

   ![设置共享相册](/images/manual/use-cases/immich-shared-album.png#bordered){width=40%}

   c. 在 **People** 下，为用户分配 Editor 或 Viewer 访问权限。

   受邀用户在登录 Immich 时可以查看共享相册中的照片，如果被授予 Editor 访问权限，则可以上传或下载照片。

## 了解更多

- [Immich 文档](https://immich.app/docs)
