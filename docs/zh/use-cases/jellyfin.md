---
outline: [2, 3]
description: 在 Olares 上用 Jellyfin 搭建私人媒体服务器，作为自托管的 Plex 替代方案。添加媒体库、启用硬件转码，并串流到你的各种设备和电视。
head:
  - - meta
    - name: keywords
      content: Olares, Jellyfin, jellyfin vs plex, plex alternative, self-hosted media server, jellyfin remote access, DLNA, jellyfin on olares
---
# 使用 Jellyfin 构建你的私人媒体服务器

Jellyfin 是一款强大的开源媒体服务器软件，让你完全掌控自己的媒体资源。在 Olares 上安装 Jellyfin 后，你可以将自己的设备打造成个人专属的流媒体平台，将电影、剧集和音乐整理成精美的媒体库，随时随地安全访问。

## 学习目标

通过本教程，你将学习：
- 在 Olares 设备上安装和设置 Jellyfin。
- 添加和整理你的媒体库。
- 使用硬件加速优化播放体验。
- 安装社区插件。
- 从客户端设备安全访问并串流你的媒体内容。
- 开启 Overlay 网关，实现电视端直连播放与投屏。

## 前提条件

在开始之前，需确保：
- 客户端设备（电脑或手机）已安装 LarePass 应用。
- 已将 Olares ID 导入 LarePass 客户端。

## 将媒体添加到 Olares

在设置 Jellyfin 之前，需要确保已可以在 Olares 上访问媒体文件。你可以通过以下几种方式添加文件：
- **直接上传文件**<br>
  将你的媒体上传到文件管理器中的 `/home/Movies/` 文件夹。为了获得更好的速度和进度可见性，[使用 LarePass 桌面客户端上传](../manual/olares/files/add-edit-download.md#upload-via-larepass-desktop)。
- **挂载外部硬盘**<br>
  将 USB 硬盘插入 Olares 设备后，它会自动挂载并可供访问。其中的文件位于 `/external/` 目录下。
- **挂载网络共享**<br>
  如果你的媒体存储在 NAS 或其他网络服务器上，可以将其连接到 Olares。有关详细说明，可参阅[挂载 SMB 共享](../manual/olares/files/mount-SMB.md)。

:::tip 命名规范
正确的文件命名是 Jellyfin 能够准确获取元数据和精美海报的关键。
需遵循 Jellyfin 的官方命名指南来整理文件：
- [电影命名规范](https://jellyfin.org/docs/general/server/media/movies/#naming)
- [剧集命名规范](https://jellyfin.org/docs/general/server/media/shows/#naming)
:::
:::tip 文件夹组织
将电影和剧集放在**不同的文件夹**中，以便于管理和正确获取元数据。
:::

## 安装和配置 Jellyfin

媒体准备好后，安装 Jellyfin 并完成其设置向导。

### 安装 Jellyfin

1. 在 Olares 网页界面中打开应用市场。
2. 在**休闲娱乐**类别中找到 **Jellyfin**，或使用搜索栏。
3. 点击**获取**，然后点击**安装**。
   ![安装 Jellyfin](/images/manual/use-cases/jellyfin-install.png#bordered)
4. 安装完成后，点击**打开**启动设置向导。

### 完成初始设置

按照向导提示完成 Jellyfin 的配置。
1. 选择你喜欢的显示语言，然后点击 **Next**。
2. 为你的 Jellyfin 管理员账户创建用户名和密码，然后点击 **Next**。
3. 当提示设置媒体库时，你可以暂时跳过此步骤。
4. 对于元数据，选择你喜欢的语言和国家，然后点击 **Next**。
5. 对于远程访问，保持默认设置（未勾选），然后点击 **Next**。Olares VPN 将负责安全的远程访问。
6. 点击 **Finish** 完成设置向导。
7. 系统将跳转至登录页面。使用你刚刚创建的管理员账号登录。

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

为了确保高分辨率视频能流畅播放，需启用硬件加速。启用后，Jellyfin 会调用你设备的底层硬件，更快速高效地转码。
1. 在 Jellyfin 的 **Dashboard** 中（点击 ≡ 图标 > Dashboard），进入 **Playback** > **Transcoding**。
2. 在 **Hardware acceleration** 下，根据你的 Olares 设备硬件配置选择对应的加速方案。

   ![启用转码](/images/manual/use-cases/jellyfin-transcoding.png#bordered){width=90%}

## 安装社区插件提升体验

你可以安装社区插件来优化元数据抓取，获取更精美的海报，以及添加全新功能。

各类插件的安装流程基本一致。本教程以 **Skin Manager** 为例，其插件仓库地址为：`https://github.com/danieladov/jellyfin-plugin-skin-manager`。

1. 点击 **Dashboard**，滚动至 **Plugins**。

   ![目录](/images/manual/use-cases/jellyfin-plugin.png#bordered){width=90%}

2. 点击右上角的 **Manage Repositories**，进入 **Repositories** 页面，然后点击 **+ New Repository** 添加新的插件仓库。
3. 输入插件的 **Repository Name** 和 **Repository URL**，然后点击 **Add**。

   ![添加插件仓库](/images/manual/use-cases/jellyfin-plugin-repo1.png#bordered){width=90%}

4. 返回 **Plugins** 页面，搜索“Skin Manager”并打开插件详情，然后点击 **Install**。如果未显示该插件，刷新页面。

   ![查找目标插件](/images/manual/use-cases/jellyfin-plugin-skin-manager.png#bordered){width=90%}
   ![安装插件](/images/manual/use-cases/jellyfin-plug-install1.png#bordered){width=90%}

5. 出现确认对话框后，查看警告信息，然后点击 **Install**。
6. 安装完成后，Jellyfin 会提示你重启服务器。前往 **Dashboard** 页面，然后点击 **Restart**。

   ![重启 Jellyfin](/images/manual/use-cases/jellyfin-restart1.png#bordered){width=90%}

7. Jellyfin 重启后，前往 **Dashboard** > **Plugins**，选择 **Installed**，确认 Skin Manager 的状态显示为 **Active**。

   ![插件状态](/images/manual/use-cases/jellyfin-plug-status1.png#bordered){width=90%}

部分插件在安装后需要进一步配置或手动开启才能生效。由于不同插件的功能差异较大，需查阅插件的 **GitHub 仓库**或 **README** 文件获取详细的设置指南。

## 开启 Overlay 网关连接电视并投屏

Overlay 网关会为 Jellyfin 分配一个局域网 IP 地址，使处于同一网络下的电视能够发现 Jellyfin，从而实现本地播放或 DLNA 投屏。

:::warning 仅在可信网络环境中开启
Overlay 网关会将 Jellyfin 直接暴露到当前局域网中。仅当需要电视自动发现或 DLNA 投屏时，在可信网络环境中开启。
:::

:::info 网络要求
需通过有线以太网连接 Olares 设备。设置和播放媒体时，确保电视与 Olares 设备在同一网络中。
:::

### 为 Jellyfin 开启 Overlay 网关

1. 打开 Olares 的**设置**，然后进入**网络** > **Overlay 网关**。
2. 确保系统级**启用 Overlay 网关**选项已开启。如果你无法自行开启，需先联系超级管理员开启。
3. 在**应用**下找到 Jellyfin，并为其开启 Overlay 网关。
4. 在确认弹窗中，点击**确认**。

   ![为 Jellyfin 开启 Overlay 网关](/images/manual/use-cases/jellyfin-enable-overlay-gateway.png#bordered){width=90%}

如果 Jellyfin 正在运行，Olares 会重启 Jellyfin 以应用网络变更。等待 Jellyfin 恢复到**运行中**状态后，再尝试从电视端连接。

### 使用电视上的 Jellyfin 应用直接播放

使用 Jellyfin 应用直接在电视上播放媒体。

1. 在电视上安装最新版 Jellyfin 应用。
2. 打开电视上的 Jellyfin 应用，然后选择添加或选择服务器。应用通常会通过 Overlay 网关的局域网 IP 自动发现你的 Jellyfin 服务器。
3. 在列表中，选择与 Olares **设置**中为 Jellyfin 分配的 Overlay 网关 IP 相同的服务器。
4. 使用你在设置过程中创建的 Jellyfin 用户名和密码登录。
5. 打开一个媒体项目，确认视频能够在电视上正常播放。

### 从 Olares 投屏到电视

在 Olares 上的 Jellyfin 中打开视频，直接推送到支持 DLNA 的电视上播放。此操作无需在电视上打开 Jellyfin 应用。

1. 在 Olares 上打开 Jellyfin。
2. 点击左上角的 <i class="material-symbols-outlined">menu</i> 图标，然后在侧边栏中点击 **Dashboard**。

   ![Jellyfin 控制台](/images/manual/use-cases/jellyfin-dashboard.png#bordered){width=90%}

3. 在 Dashboard 侧边栏中，向下滚动到 **Plugins**，然后点击 **Catalog**。
4. 搜索 "DLNA"，打开 DLNA 插件结果，然后点击 **Install**。
5. 插件安装完成后，在 Olares 中重启 Jellyfin：

   a. 打开**设置**，然后进入**应用** > **Jellyfin**。

   b. 点击**暂停**。

   c. Jellyfin 停止后，点击**恢复**。

6. 在 Olares 上重新打开 Jellyfin。
7. 打开要投屏的视频。
8. 点击右上角的 <i class="material-symbols-outlined">cast</i> 图标（**Play on**）。Jellyfin 应该会自动发现你的电视。

   ![在 Jellyfin 的 Play on 中选择电视](/images/manual/use-cases/jellyfin-play-on-tv.png#bordered){width=90%}

9. 选择电视，确认视频开始在电视上播放。

## 通过 Jellyfin 客户端访问你的媒体库
### 获取 Jellyfin 的端点

Jellyfin 设置完成且媒体库准备就绪后，你可以从客户端设备连接并开始串流你的媒体。

:::info 启用 LarePass VPN
在开始之前，需确保 LarePass VPN 已启用。

如果未启用，可参阅[在 LarePass 上启用 VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass)。
:::

1. 在 Olares 上，打开设置，然后进入**应用** > **Jellyfin**。
2. 在**入口**下，点击 **Jellyfin**。
3. 确保**认证级别**设置为**内部**。如果你更改了设置，点击**提交**。
4. 在**端点配置**下，复制**端点**中显示的 URL。在你的 Jellyfin 客户端中使用此地址作为服务器 URL。

   ![Jellyfin 端点](/images/manual/use-cases/lp-endpoint-jellyfin.png#bordered){width=90%}

### 连接你的 Jellyfin 客户端

假设你已经在设备上安装了官方的 [Jellyfin 客户端](https://jellyfin.org/downloads/)。

1. 在你的设备上打开 Jellyfin 客户端应用。
2. 点击 **Add Server**。

   ![添加服务器](/images/manual/use-cases/jellyfin-add-server.png#bordered){width=90%}

3. 将你刚刚复制的 Jellyfin URL 粘贴到客户端中，然后点击 **Connect**。

   ![连接到服务器](/images/manual/use-cases/jellyfin-connect.png#bordered){width=90%}

4. 使用你的 Jellyfin 管理员账户登录。

登录成功后，你就可以在应用中看到你的媒体库内容了。

:::tip
为了获得最佳体验，在远程访问 Jellyfin 时保持 LarePass VPN 处于连接状态。这能确保你始终能够安全地连接到你的 Jellyfin 服务器。
:::