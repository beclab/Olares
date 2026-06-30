---
outline: [2, 3]
description: 了解如何使用 YT-DLP 将 YouTube 视频保存到 Wise 库中以供离线观看。
head:
- - meta
  - name: keywords
    content: Wise, YouTube, LarePass 浏览器扩展
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/wise-download.md)。
:::

# 在 Wise 中下载 YouTube 视频 <Badge text="15 min"/>

Wise 允许你直接将 YouTube 视频保存到库中以供离线观看。与标准文件下载不同，保存到库可以确保视频在 Wise 应用内显示，让你可以像管理其他内容一样管理、播放和组织它。

## 学习目标

完成本教程后，你将学会如何：

- 通过在 Wise 中粘贴链接将 YouTube 视频保存到你的 Wise 库（推荐）。
- 使用 LarePass 扩展从浏览器保存视频。
- 检查下载进度并离线观看视频。

## 安装 Wise 和 YT-DLP

在下载 YouTube 视频之前，请确保已从 Olares Market 安装以下应用：
- **Wise**：界面应用。
- **YT-DLP**：后台下载服务。

1. 从 Launchpad 打开 Market。
2. 搜索 "Wise"，点击 **Get**，然后点击 **Install**。
3. 搜索 "YT-DLP"，点击 **Get**，然后点击 **Install**。

## 保存视频到库

你可以通过在 Wise 中直接粘贴 URL 或使用 LarePass 扩展在浏览时保存视频。

### 在 Wise 中粘贴 URL <Badge text="推荐"/>

1. 复制 YouTube 视频的 URL。
2. 打开 Wise，点击左下角菜单栏中的 <i class="material-symbols-outlined">add_circle</i>，然后选择 **Add link**。
3. 将 URL 粘贴到输入框中。Wise 会立即分析链接。
4. 在面板中，找到 Save to library 部分：
    - **视频已就绪**：点击视频条目旁边的 <i class="material-symbols-outlined">box_add</i> 进行保存。
    ![通过链接保存视频](/images/one/wise-save-to-lib-add-link.png#bordered){width=60%}

    - **需要登录**：如果你看到登录网站并上传 Cookie 的提示，只需点击 **Upload**。状态更新后，点击 <i class="material-symbols-outlined">box_add</i> 进行保存。

### 使用 LarePass 扩展

如果你经常在浏览时保存视频，[LarePass 扩展](https://www.olares.com/larepass) 可能更方便，但需要先安装扩展。

1. 在浏览器中打开 YouTube 视频。
2. 点击浏览器工具栏中的 LarePass 浏览器扩展图标，然后选择 "Collect" 图标。
3. 在面板中，找到 Save to library 部分：
   - **视频已就绪**：点击视频条目旁边的 <i class="material-symbols-outlined">box_add</i> 进行保存。
    ![通过 LarePass 保存视频](/images/one/wise-save-to-lib-larepass.png#bordered){width=60%}

   - **需要登录**：如果你看到登录网站并上传 Cookie 的提示，只需点击 **Upload**。状态更新后，点击 <i class="material-symbols-outlined">box_add</i> 进行保存。

## 监控和管理视频下载

当你将视频保存到库中时，Wise 会立即创建记录，文件下载在后台运行。

1. 在 Wise 中，点击左下角菜单栏中的 <i class="material-symbols-outlined">settings</i>，然后选择 **Download list**。
2. 使用标签页筛选任务：**All**、**Downloading**、**Completed**、**Failed**。
3. 查看任务列表和状态。
4. 你可以：
   - 点击 <i class="material-symbols-outlined">folder_open</i> 在 Files 中定位已传输的文件。
   - 点击 <i class="material-symbols-outlined">do_not_disturb_on</i> 将其从列表中移除。

下载完成后，即使没有互联网连接，你也可以直接在 Wise 中播放视频。

## 资源

如果你想探索 Wise 除离线视频保存之外的更多功能，另请参阅：

- [Wise 基础](../../manual/olares/wise/basics.md)：应用的一般使用方法。
- [订阅和管理订阅源](../../manual/olares/wise/subscribe.md)：如果你想通过 RSS 关注 YouTube 频道。
- [管理 Wise 的 Cookie](../../manual/olares/wise/manage-cookies.md)：详细的 Cookie 管理配置。
