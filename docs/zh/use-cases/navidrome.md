---
outline: [2, 3]
description: 在 Olares 上设置 Navidrome，构建私有音乐流媒体服务器，整理你的音乐库，并从移动 Subsonic 兼容客户端播放。
head:
  - - meta
    - name: keywords
      content: Olares, Navidrome, music streaming, self-hosted music server, Subsonic, mobile music client
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-05-26"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/navidrome.md)。
:::

# 使用 Navidrome 流媒体播放你的音乐库

Navidrome 是一个开源的、自托管的音乐流媒体服务器，可将你的本地音乐收藏转变为个人云音乐库。它支持几乎所有音频格式，占用资源极少，并且兼容 Subsonic/Airsonic API。这使你可以从任何设备连接各种第三方音乐应用。

## 学习目标

在本指南中，你将学习如何：
- 安装 Navidrome 并创建管理员账户。
- 从 Olares Files 将音乐文件导入 Navidrome 库文件夹。
- 使用 Subsonic 兼容应用将你的库流媒体传输到移动设备。

## 前提条件

- 你的手机上已安装 LarePass 并使用你的 Olares ID 登录。
- 准备好上传到 Olares 的音乐文件。
- 一个移动 Subsonic 兼容音乐客户端。本指南使用 [Stream Music](https://music.aqzscn.cn/en/docs/versions/latest) 作为示例。

## 安装 Navidrome

1. 打开 Market 并搜索 "Navidrome"。

   ![Navidrome](/images/manual/use-cases/navidrome.png#bordered){width=95%}

2. 点击 **获取**，然后 **安装**，等待安装完成。

## 设置管理员账户

首次打开 Navidrome 时，创建用于管理服务器的管理员账户。

1. 从 Launchpad 打开 Navidrome。

2. 按照页面提示创建管理员用户名和密码。

   ![创建 Navidrome 管理员账户](/images/manual/use-cases/navidrome-create-admin.png#bordered){width=95%}

登录后，你应该会看到一个空的 Navidrome 库。上传音乐并且 Navidrome 扫描文件夹后，库将被填充。

## 向 Navidrome 添加音乐

Navidrome 扫描 Olares Files 中的 `Home/Music` 文件夹。放在此文件夹中的文件将在扫描后出现在你的 Navidrome 库中。

1. 从 Launchpad 打开 Files。

2. 前往 **主页** > **音乐**。

3. 将你的音乐文件上传到此文件夹。

   ![上传音乐到 Home/Music](/images/manual/use-cases/navidrome-upload-music.png#bordered){width=95%}

:::tip 按专辑整理
为了更整洁的浏览体验，请在上传前将文件整理到专辑文件夹中。Navidrome 仍然可以识别单个曲目，但专辑文件夹使库更易于扫描和维护。
:::

扫描完成后，Navidrome 将在库中显示你的歌曲、专辑和艺术家。

## 扫描音乐库

Navidrome 自动扫描音乐文件夹。如果新上传的歌曲没有出现，请运行手动完整扫描。

1. 返回 Navidrome。

2. 在右上角，点击 **活动**。

3. 点击 **完整扫描** 图标。

   ![在 Navidrome 中运行完整扫描](/images/manual/use-cases/navidrome-full-scan.png#bordered){width=95%}

4. 等待扫描完成，然后刷新库视图。

## 连接移动音乐客户端

要从手机流媒体播放，请允许客户端访问 Navidrome，启用 LarePass VPN，并从 Subsonic 兼容客户端登录。

1. 更新 Navidrome 的访问策略并复制其端点：

   a. 打开 **设置**，然后前往 **应用** > **Navidrome**。

   b. 在 **入口** 下，点击 **Navidrome**。

   c. 将 **认证级别** 设置为 **内部**，然后点击 **提交**。

   d. 在 **端点设置** 下，复制 **端点** 中显示的 URL。

   ![将 Navidrome 认证级别设置为内部](/images/manual/use-cases/alex-navidrome-endpoint.png#bordered){width=95%}

2. 在手机上启用 LarePass VPN。

   ![在移动设备上启用 LarePass VPN](/images/manual/get-started/larepass-vpn-mobile.png#bordered){width=95%}

3. 在手机上打开 Stream Music，然后选择连接 Navidrome 的选项。

   ![将 Stream Music 连接到 Navidrome](/images/manual/use-cases/navidrome-music-stream-connect.png#bordered){width=95%}

4. 在登录页面，输入你的信息：
   - **主机地址**：你从 Olares 设置复制的 Navidrome 端点。
   - **用户名**：你的 Navidrome 用户名。
   - **密码**：你的 Navidrome 密码。

   ![登录 Navidrome](/images/manual/use-cases/navidrome-log-in.png#bordered){width=95%}

5. 点击 **登录**。

当应用显示登录成功消息时，返回主页。你的 Navidrome 库应该出现在移动客户端中。

## 常见问题

### 我可以添加带有外部 `.lrc` 文件的歌词吗？

Olares 上的 Navidrome 最可靠的方式是使用嵌入式同步歌词。外部 `.lrc` 文件可能不会出现在 Navidrome Web 界面中。某些 Subsonic 兼容客户端可能支持它们，具体取决于 Navidrome 和客户端版本。

要使歌词始终有效，请在将音频文件上传到 `Home/Music` 之前将同步歌词嵌入其中。

## 了解更多

- [使用 Jellyfin 构建你的私有媒体服务器](stream-media.md)：使用 Jellyfin 从 Olares 流媒体播放电影、电视节目和音乐。
- [使用 Komga 构建你的数字图书馆](komga.md)：在 Olares 上管理漫画、manga、杂志和电子书。
