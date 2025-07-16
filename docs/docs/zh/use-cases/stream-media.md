---
outline: [2, 3]
description: 通过 Olares 搭建私人流媒体服务器，配置 VPN 远程访问、管理媒体文件、使用 Jellyfin 进行视频串流播放。
---
# 远程观看 Olares 中下载的视频
无论是在家还是在外，Olares 都能让你轻松访问个人媒体库。借助内置的专有网络功能和 Jellyfin 等应用，你可以随时随地安全流畅地观看自己收藏的影视作品。本教程将为你介绍如何设置 Olares 专有网络、访问媒体文件，以及配置 Jellyfin。

## 目标
通过本教程，你将学习：
- 配置 Olares 专有网络，实现从任何地方安全访问媒体库，获得局域网般的访问体验。
- 通过 LarePass 客户端或网页浏览器浏览和播放存储在 Olares **文件管理器**中的视频文件。
- 安装和设置 Jellyfin 以实现串流。

## 打开 Olares 专用网络
为了在不同网络中实现流畅的流媒体播放，需要在 LarePass 中[启用专用网络](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。
启用专用网络的设备无论是通过 LarePass 客户端还是浏览器访问 Olares，都会使用专用网络连接。

:::tip
如需下载不同版本的 LarePass，请访问[官方页面](https://olares.cn/larepass)。
:::

<!--@include: ./remote.reusables.md{4,24}-->

## 在 LarePass 中访问媒体文件
启用专用网络后，你可以浏览存储在 Olares 上的媒体文件。

### 通过 LarePass 客户端访问
1. 打开 LarePass，点击**文件**，进入包含电影和电视剧的媒体目录。
2. 点击任意视频文件，即可在电脑或移动设备上开始播放。

![在 LarePass 客户端中播放视频](/images/zh/manual/use-cases/view-video-from-larepass-desktop.png#bordered)

### 通过浏览器访问
1. 以网页模式打开**文件管理器**，或者直接使用其 URL：`https://files.[OlaresDomainName]`。

   ![打开文件管理器](/images/zh/manual/use-cases/view-video-from-files.png#bordered)
2. 找到你的媒体目录，点击视频文件即可开始播放。

   ![从文件管理器播放视频](/images/zh/manual/use-cases/view-video-from-files-2.png#bordered)

## 使用 Jellyfin 访问媒体文件
如果需要高级媒体管理功能，例如字幕支持和多声道音频，可以安装 Jellyfin。

1. 打开应用市场，进入**娱乐**分类。
2. 找到并下载 Jellyfin。
3. 启动 Jellyfin，并完成首次设置：
   - 设置管理员密码。

   ![设置管理员密码](/images/zh/manual/use-cases/jellyfin-set-admin.png#bordered)
   - 配置媒体库目录。

   ![设置管理员密码](/images/zh/manual/use-cases/jellyfin-set-media-library.png#bordered)
4. 等待 Jellyfin 扫描并索引你的媒体库。它会自动获取以下元数据：
   - 电影海报
   - 描述信息
   - 演职员表

   ![Jellyfin](/images/zh/manual/use-cases/jellyfin-details.png#bordered)

## 通过 Jellyfin 客户端访问你的媒体库
要在多种设备上流式播放媒体文件：

1. 配置 Jellyfin 认证。

   a. 打开设置，进入**应用** > **Jellyfin** > **入口**。

   b. 将**认证级别**设置为**内部**。

    ![设置认证级别](/images/zh/manual/use-cases/jellyfin-auth-level.png#bordered)
2. 下载并安装 [Jellyfin 官方客户端](https://jellyfin.org/downloads/)。
3. 将客户端连接到 Olares 中的 Jellyfin 服务器。

   a. 在 Jellyfin 客户端中，自动发现功能会定位到同一网络中的服务器。

   b. 如果自动发现失败，可手动输入 Olares 提供的服务器地址。

4. 使用你的凭据登录 Jellyfin 客户端。

:::tip
远程访问媒体库时，请保持专用网络连接处于激活状态，以获得最佳流媒体播放效果。
:::
