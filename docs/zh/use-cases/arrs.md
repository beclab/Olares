---
outline: deep
description: 了解如何在 Olares 上连接和配置 *Arrs 系列应用（Sonarr、Radarr、Prowlarr、Bazarr 和 qBittorrent）以实现自动化媒体管理。
head:
  - - meta
    - name: keywords
      content: Olares, *Arrs, *Arr, Starr, Starrs, Sonarr, Radarr, Prowlarr, Bazarr, qBittorrent, media server, self-hosted
app_version: "1.0.x"
doc_version: "1.1"
doc_updated: "2026-07-03"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/arrs.md)。
:::

# 使用 *Arrs 生态系统管理你的媒体库

*Arrs 系列是一套开源的、自托管的媒体管理器。Sonarr 管理电视节目，Radarr 管理电影，Lidarr 处理音乐，Readarr 整理书籍。Prowlarr 为这些应用管理索引器，而 Bazarr 处理字幕服务。

通过配置这些工具之间的连接，它们可以相互通信，自动搜索、下载和组织你的媒体。

## 学习目标

在本指南中，你将学习如何：
- 安装 *Arr 应用。
- 在 Olares 中获取 *Arr 应用的内部入口 URL。
- 将 qBittorrent 作为下载客户端连接到 Sonarr。
- 将 Prowlarr 作为索引器管理器连接到 Sonarr。
- 将 Bazarr 作为字幕管理器连接到 Sonarr。

## 前提条件

本指南专门介绍 *Arrs 应用之间的连接配置。它不包括每个应用的完整设置或一般用法。

在连接它们之前，请确保你已正确配置媒体管理器和下载客户端的核心设置。

## 升级说明

从 Olares V1.12.6 开始，*Arr 应用使用内部入口 URL 而不是提供商 URL。如果你是从更早版本升级的，请查看 [*Arrs 升级说明](arrs-upgrade.md) 并更新受影响的设置。

## 安装 *Arrs 应用

安装你的媒体栈所需的 *Arrs 应用。本教程使用 Sonarr、Prowlarr、Bazarr 和 qBittorrent。

1. 打开 Market 并搜索 "Sonarr"。

   ![Market 中的 Sonarr 应用](/images/manual/use-cases/arrs-sonarr.png#bordered)

2. 点击 **Get**，然后点击 **Install**。等待安装完成。
3. 搜索 "Prowlarr" 并安装它。

   ![Market 中的 Prowlarr 应用](/images/manual/use-cases/arrs-prowlarr.png#bordered)

4. 搜索 "Bazarr" 并安装它。

   ![Market 中的 Bazarr 应用](/images/manual/use-cases/arrs-bazarr.png#bordered)

5. 搜索 "qBittorrent" 并安装它。

   ![Market 中的 qBittorrent 应用](/images/manual/use-cases/arrs-qbittorrent.png#bordered)

## 完成初始应用设置

为了保护你的媒体服务器并防止未经授权的访问，如果提示，你必须在首次启动时为应用配置管理员凭据。

以下步骤演示 Sonarr 的初始设置。

1. 从 Launchpad 打开 Sonarr。出现 **Authentication Required** 页面。

   ![首次启动时的 Sonarr 初始设置](/images/manual/use-cases/arrs-sonarr-ini-settings.png#bordered)

2. 选择认证方法：
   - **Basic (Browser Popup)**：要使用浏览器的原生登录提示，请选择此选项。此方法通常与自动密码管理器更兼容。
   - **Forms (Login Page)**：要使用 Sonarr 内置的自定义登录界面以获得更视觉化的集成体验，请选择此选项。

3. 在 **Authentication Required** 列表中，保持默认选择 **Enabled** 以确保最大安全性。此选项无论从哪里访问应用都需要用户名和密码。
4. 在 **Username** 字段中，输入管理员用户名。
5. 在 **Password** 字段中，输入安全密码。
6. 在 **Password Confirmation** 字段中，再次输入相同的密码。
7. 点击 **Save**。你将登录到 Sonarr。

   ![Sonarr 着陆页](/images/manual/use-cases/arrs-sonarr-landing.png#bordered)

## 获取内部入口 URL

要连接 *Arr 应用，你需要每个应用的内部入口 URL。以下步骤演示如何获取 Sonarr 的内部入口 URL。对 Prowlarr、Bazarr 和下载客户端使用相同的方法。

1. 打开 Olares Settings，然后进入 **Applications** > **Sonarr** > **Entrances** > **Sonarr**。

   ![Settings 中的 Sonarr 入口](/images/manual/use-cases/arrs-sonarr-entrance.png#bordered){width=75%}

2. 确保 **Authentication level** 设置为 **Internal**。
3. 复制 **Endpoint** URL。例如 `https://9691c178.alexmiles.olares.com`。

   ![Settings 中的 Sonarr 内部端点](/images/manual/use-cases/arrs-sonarr-endpoint1.png#bordered){width=75%}

## 将下载客户端连接到媒体管理器

要下载媒体，你必须将下载客户端（如 qBittorrent 或 Transmission）连接到你的媒体管理器（Sonarr、Radarr、Lidarr 和 Readarr）。

以下步骤演示如何将 qBittorrent 连接到 Sonarr。你可以对其他媒体管理器应用相同的过程。

1. 打开 Sonarr，点击左侧边栏中的 **Settings**，然后选择 **Download Clients**。

   ![Sonarr Download Clients 页面](/images/manual/use-cases/arrs-sonarr-download-clients.png#bordered)

2. 点击 <span class="material-symbols-outlined">add_2</span>，然后向下滚动选择 **qBittorrent** 以添加新的客户端连接。
3. 按如下方式指定连接详细信息：

   ![Sonarr Download Clients 设置](/images/manual/use-cases/arrs-sonarr-download-clients-settings1.png#bordered)

   - **Host**：输入 qBittorrent 的内部入口 URL。例如，`https://44e535c5.alexmiles.olares.com`。
   - **Port**：输入 `443`。
   - **Use SSL**：启用此选项。
   - **Username** 和 **Password**：
      - 如果你的 qBittorrent 客户端需要认证，请输入你的用户名和密码。
      - 如果你使用默认的 qBittorrent 设置且没有密码，则将两个字段留空。
4. 点击 **Test**。出现绿色对勾，表示连接成功。
5. 选择 **Save**。qBittorrent 在 **Download Clients** 部分中显示为已启用。

   ![Sonarr Download Clients 已启用](/images/manual/use-cases/arrs-sonarr-download-clients-enabled.png#bordered)

## 将索引器管理器连接到媒体管理器

要自动在多个索引器（搜索站点）中搜索媒体文件，你必须将索引器管理器（如 Prowlarr）连接到你的媒体管理器。

以下步骤演示如何将 Prowlarr 连接到 Sonarr。你可以对其他媒体管理器应用相同的过程。

### 步骤 1：获取 Sonarr API 密钥

1. 打开 Sonarr，点击左侧边栏中的 **Settings**，然后选择 **General**。
2. 在 **Security** 部分中，记下 API 密钥。在本例中，它是 `e4ee9f376d754fd3b7146629d737644f`。

   ![Sonarr API 密钥](/images/manual/use-cases/arrs-sonarr-api.png#bordered)

### 步骤 2：在 Prowlarr 中添加索引器

1. 打开 Prowlarr 并登录。

   ![Prowlarr 着陆页](/images/manual/use-cases/arrs-prowlarr-landing.png#bordered)

2. 点击 **Add New Indexer**。
3. 添加你喜欢的索引器并确保它们连接成功。例如，要添加 Uindex：

   a. 搜索 Uindex，然后从结果列表中点击它。

   ![Prowlarr 添加索引器](/images/manual/use-cases/arrs-prowlarr-add-indexer.png#bordered)

   b. 点击 **Test**。出现绿色对勾，表示连接成功。

   :::tip
   如果连接测试失败，索引器可能需要 Cloudflare 绕过。有关解决此问题的说明，请参阅[使用 FlareSolverr 在 Prowlarr 中访问 Cloudflare 保护的站点](flaresolverr.md)。
   :::

   c. 点击 **Save**。

4. 关闭 **Add Indexer** 窗口。Prowlarr 显示已启用的索引器。

   ![Prowlarr 索引器已添加](/images/manual/use-cases/arrs-prowlarr-indexer-added.png#bordered)

### 步骤 3：将 Prowlarr 与 Sonarr 同步

1. 在 Prowlarr 中，点击左侧边栏中的 **Settings**，然后选择 **Apps**。
2. 点击 **Apps**。

   ![Prowlarr 添加应用](/images/manual/use-cases/arrs-prowlarr-add-apps.png#bordered)

3. 点击 <span class="material-symbols-outlined">add_2</span>，然后选择要连接的应用，即 Sonarr。
4. 在 **Add Application - Sonarr** 窗口中，指定以下设置：

   - **Prowlarr Server**：输入 Prowlarr 的内部入口 URL，例如 `https://e5e5b409.alexmiles.olares.com`。
   - **Sonarr Server**：输入 Sonarr 的内部入口 URL，例如 `https://9691c178.alexmiles.olares.com`。
   - **API Key**：输入你之前记下的 API 密钥。

   ![Prowlarr 添加应用配置](/images/manual/use-cases/arrs-prowlarr-add-apps-confi1.png#bordered)

5. 点击 **Test**。出现绿色对勾，表示连接成功。
6. 点击 **Save**。Sonarr 出现在 **Applications** 部分中，Prowlarr 自动将索引器推送到 Sonarr。

   ![Prowlarr 应用已添加](/images/manual/use-cases/arrs-prowlarr-apps-added.png#bordered)

7. 要验证同步，打开 Sonarr，然后前往 **Settings** > **Indexers**。你可以看到从 Prowlarr 导入的数据源。

   ![Prowlarr Sonarr 同步验证](/images/manual/use-cases/arrs-prowlarr-sync-verify.png#bordered)

   现在当你在 Sonarr 中添加新电视节目时，它会搜索这些索引器以获取可用文件，并触发 qBittorrent 下载它们。

## 将字幕管理器连接到媒体管理器

要自动为媒体库下载缺失的字幕，你必须将字幕管理器（如 Bazarr）连接到你的媒体管理器（Sonarr 和 Radarr）。

以下步骤演示如何将 Bazarr 连接到 Sonarr。

1. 从 Launchpad 打开 Bazarr，然后点击左侧边栏中的 **Sonarr**。
2. 打开 **Enabled** 开关，然后指定以下设置：

   - **Address**：输入 Sonarr 的内部入口 URL，不包括 `https://` 前缀。例如，`9691c178.alexmiles.olares.com`。
   - **Port**：输入 `443`。
   - **API Key**：输入你之前记下的 Sonarr API 密钥。
   - **SSL**：启用此选项。

3. 点击 **Test**。如果连接成功，Bazarr 会显示已连接的 Sonarr 版本号。

   ![Bazarr 连接成功](/images/manual/use-cases/arrs-bazarr-test-connection1.png#bordered)

4. 选择左上角的 **Save**。

   Bazarr 现在监控 Sonarr。每当 Sonarr 下载电视节目时，Bazarr 会自动检测它并根据你的语言设置下载相应的字幕。

## 常见问题

### 为什么我无法将某些应用（如 JDownloader）连接到我的媒体管理器？

某些应用不受 *Arr 媒体管理器支持，因为它们没有实现 Sonarr、Radarr 等应用所期望的 API。例如，JDownloader 无法连接到 Sonarr 或 Radarr，因为 *Arr 应用不包含兼容 JDownloader 的下载客户端集成。

有关支持的下载客户端和集成的完整列表，请参阅官方 [Servarr Wiki](https://wiki.servarr.com/en/sonarr/supported)。

### 为什么我的连接测试失败？

- **检查 URL 格式**：查看你配置的特定应用的要求。
  - **Prowlarr 连接到 Sonarr**：使用 `https://` 前缀。
  - **Sonarr 连接到 qBittorrent** 或 **Bazarr 连接到 Sonarr**：输入不带 `https://` 前缀的内部入口 URL。
- **验证端口**：确保将端口设置为 `443`。
- **启用 SSL**：对于使用端口 `443` 的连接，请确保启用 SSL 或 Use SSL 选项。

## 了解更多

- [Servarr Wiki](https://wiki.servarr.com/)
