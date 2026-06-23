---
outline: [2, 3]
description: 在 Olares 上设置 Jellyfin 的逐步指南，用于个人媒体流媒体播放。了解如何使用 LarePass 管理媒体文件、添加媒体库、通过插件增强元数据、启用硬件加速，以及通过 Olares VPN 从任何设备安全地串流您的媒体。
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../use-cases/stream-media.md)为准。
:::

# 使用 Jellyfin 构建你的私人媒体服务器

Jellyfin 是一款强大的开源媒体服务器软件，让你完全掌控自己的媒体。通过在 Olares 上安装它，你可以将自己的设备变成一个个人流媒体平台，将电影、剧集和音乐整理成一个精美的媒体库，随时随地安全访问。

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 设备上安装和设置 Jellyfin。
- 添加和整理你的媒体库。
- 使用硬件加速优化播放体验。
- 安装社区插件。
- 从客户端设备安全地访问和串流你的媒体。

## 前提条件

在开始之前，请确保：
- 在客户端设备（桌面或移动设备）上安装了 LarePass 应用。
- 已将 Olares ID 导入 LarePass 客户端。

## 将媒体添加到 Olares

在设置 Jellyfin 之前，你需要确保媒体已经可以在 Olares 上访问。你可以通过以下几种方式添加：
- **直接上传文件**<br>
  将你的媒体上传到 Files 中的 `/home/Movies/` 文件夹。为了获得更好的速度和进度可见性，[使用 LarePass 桌面客户端上传](../manual/olares/files/add-edit-download.md#upload-via-larepass-desktop)。
- **挂载外部硬盘**<br>
  将 USB 硬盘插入 Olares 设备后，它会自动挂载并可供访问。其中的文件位于 `/external/` 目录下。
- **挂载网络共享**<br>
  如果你的媒体存储在 NAS 或其他网络服务器上，可以将其连接到 Olares。有关详细说明，请参阅[挂载 SMB 共享](../manual/olares/files/mount-SMB.md)。

:::tip 命名规范
正确的文件命名是 Jellyfin 获取准确元数据和精美海报的关键。
请遵循 Jellyfin 的官方命名指南：
- [电影命名规范](https://jellyfin.org/docs/general/server/media/movies/#naming)
- [剧集命名规范](https://jellyfin.org/docs/general/server/media/shows/#naming)
:::
:::tip 文件夹组织
将电影和剧集放在**不同的文件夹**中，以便于管理和正确获取元数据。
:::

## 安装和配置 Jellyfin

媒体准备好后，安装 Jellyfin 并完成其设置向导。

### 安装 Jellyfin

1. 在 Olares 网页界面中打开 **Market** 应用。
2. 在 **Fun** 类别中找到 **Jellyfin**，或使用搜索栏。
3. 点击 **Get**，然后点击 **Install**。
   ![安装 Jellyfin](/images/manual/use-cases/jellyfin-install.png#bordered)
4. 安装完成后，点击 **Open** 启动设置向导。

### 完成初始设置

按照向导提示完成 Jellyfin 的配置。
1. 选择你喜欢的显示语言，然后点击 **Next**。
2. 为你的 Jellyfin 管理员账户创建用户名和密码，然后点击 **Next**。
3. 当提示设置媒体库时，你可以暂时跳过此步骤。
4. 对于元数据，选择你喜欢的语言和国家，然后点击 **Next**。
5. 对于远程访问，保持默认设置（未勾选），然后点击 **Next**。Olares VPN 将负责安全的远程访问。
6. 点击 **Finish** 完成设置向导。
7. 你将被带到登录页面。使用你刚刚创建的管理员凭据登录。

   ![登录 Jellyfin](/images/manual/use-cases/jellyfin-sign-in.png#bordered){width=90%}

## 添加媒体库

Jellyfin 安装并运行后，下一步是告诉它你的媒体存储在哪里。
1. 在 Jellyfin 的侧边栏中，进入 **Dashboard** > **Libraries** > **Libraries**。
2. 点击 **Add Media Library**。

   ![添加媒体库](/images/manual/use-cases/jellyfin-add-media-lib.png#bordered){width=90%}

3. 配置媒体库设置：
   - **Content type**：选择媒体类型（例如 Movies、Shows、Music）。对于同时包含电影和剧集的文件夹，选择 **Mixed Movies and Shows**。
   - **Display name**：输入要在库中显示的名称。<br>
   - **Folders**：点击 + 添加你的媒体路径。<br>
      - **Olares Files**：`/home/movies/<YourMediaFolder>`
      - **External storage**：`/external/<YourMediaFolder>`
4. 点击 **Ok** 保存，并为其他媒体类型重复操作（例如，分别创建一个 "Movies" 和一个 "TV Shows"）。

保存后，Jellyfin 将自动扫描你的文件夹并开始构建媒体库。根据你的收藏规模，此过程可能需要几分钟。

## 启用转码

为了确保高分辨率视频的流畅播放，请启用硬件加速。这允许 Jellyfin 使用你设备的硬件进行更快、更高效的转码。
1. 在 Jellyfin **Dashboard** 中（点击 ≡ 图标 > Dashboard），进入 **Playback** > **Transcoding**。
2. 在 **Hardware acceleration** 下，根据你的 Olares 设备硬件选择你喜欢的方法。

   ![启用转码](/images/manual/use-cases/jellyfin-transcoding.png#bordered){width=90%}

## 通过社区插件增强体验

你可以安装插件来改进元数据、获取更好的 artwork 和添加新功能。


所有插件的安装过程都相同。以下以 **Skin Manager** 为例：
1. 在 Dashboard 中，进入 **Plugins** > **Catalog**。

   ![目录](/images/manual/use-cases/jellyfin-catalog.png#bordered){width=90%}

2. 点击 <span style="font-size: 1.1em;">&#9881;</span> 图标进入 **Repositories** 页面，然后点击 **+** 添加新仓库。
3. 输入插件的 **Repository Name** 和 **Repository URL**，然后点击 **Save**。

   ![添加插件仓库](/images/manual/use-cases/jellyfin-plugin-repo.png#bordered){width=90%}

4. 点击 **Ok** 确认安装。

   ![确认插件安装](/images/manual/use-cases/jellyfin-confirm-plug.png#bordered){width=90%}

5. 返回 **Catalog** 标签页，找到你想要的插件（可能需要刷新），然后点击 **Install**。

   ![目录插件](/images/manual/use-cases/jellyfin-catalog-plug.png#bordered){width=90%}
   ![安装插件](/images/manual/use-cases/jellyfin-plug-install.png#bordered){width=90%}

6. 安装完成后，会出现重启 Jellyfin 的提示。进入 **Dashboard** 页面并点击 **Restart**。

   ![重启 Jellyfin](/images/manual/use-cases/jellyfin-restart.png#bordered){width=90%}

7. 重启后，返回 **Dashboard** > **Plugins** > **My Plugins**，确认你安装的插件已列出且状态为 **Active**。

   ![插件状态](/images/manual/use-cases/jellyfin-plug-status.png#bordered){width=90%}

安装插件后，你可能需要启用或配置它们才能生效。
由于每个插件的行为不同，请查看插件的 **GitHub repository** 或 **README** 了解设置详情。

## 通过 Jellyfin 客户端访问你的媒体库
### 获取 Jellyfin 的端点

Jellyfin 设置完成且媒体库准备就绪后，你可以从客户端设备连接并开始串流你的媒体。

:::info 启用 LarePass VPN
在开始之前，请确保 LarePass VPN 已启用。

如果未启用，请参阅[在 LarePass 上启用 VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass)。
:::

1. 在 Olares 上，打开 Settings，然后进入 **Application** > **Jellyfin**。
2. 在 **Entrances** 下，点击 **Jellyfin**。
3. 确保 **Authentication level** 设置为 **Internal**。如果你更改了设置，点击 **Submit**。
4. 在 **Endpoint settings** 下，复制 **Endpoint** 中显示的 URL。在你的 Jellyfin 客户端中使用此地址作为服务器 URL。

   ![Jellyfin 端点](/images/manual/use-cases/lp-endpoint-jellyfin.png#bordered){width=90%}

### 连接你的 Jellyfin 客户端

假设你已经在设备上安装了官方的 [Jellyfin client app](https://jellyfin.org/downloads/)。

1. 在你的设备上打开 Jellyfin 客户端应用。
2. 点击 **Add Server**。

   ![添加服务器](/images/manual/use-cases/jellyfin-add-server.png#bordered){width=90%}

3. 将你刚刚复制的 Jellyfin URL 粘贴到客户端中，然后点击 **Connect**。

   ![连接到服务器](/images/manual/use-cases/jellyfin-connect.png#bordered){width=90%}

4. 使用你的 Jellyfin 管理员账户登录。

你现在应该在应用中看到你的媒体库了。

:::tip
为了获得最佳体验，在远程访问 Jellyfin 时保持 LarePass VPN 连接处于活动状态。这可以确保你始终能够安全地连接到你的 Jellyfin 服务器。
:::
