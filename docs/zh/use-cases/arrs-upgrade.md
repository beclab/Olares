---
outline: deep
description: 更新到 Olares 1.12.6 后，调整 *Arr 应用连接和路径的方法。
head:
  - - meta
    - name: keywords
      content: Olares, *Arr 应用, Sonarr, Radarr, Prowlarr, 更新, 1.12.6
app_version: "1.0.x"
doc_version: "1.0"
doc_updated: "2026-07-06"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/arrs-upgrade.md)为准。
:::

# *Arr 应用更新说明

本文介绍更新到 Olares V1.12.6 后，需要对 *Arr 应用进行的配置更改。

## 更新 *Arr 应用连接

从 Olares V1.12.6 开始，*Arr 应用之间通过内部入口 URL 进行通信，而不是提供商 URL。这一更改简化了网络模型，并使媒体栈更加一致。

更新到 V1.12.6 后，现有连接可能会失效。你需要将其改为使用内部入口 URL。

1. 在 Olares 中，将每个 *Arr 应用更新到 V1.12.6 可用的最新版本。
2. 打开**设置**，然后进入**应用** > **[Arr 应用名称]** > **入口**。
3. 确保**认证级别**设置为**内部**。
4. 复制 **Endpoint** URL。这就是内部入口 URL。
5. 在更新以下章节中的连接设置时，使用此 URL。

:::tip
以下章节以 Sonarr、Prowlarr 和 qBittorrent 为例。如果你使用其他 *Arr 应用（如 Radarr、Lidarr、Readarr 或 Bazarr）或其他下载客户端（如 NZBGet、Transmission 或 Deluge），请按照相同步骤更新它们的连接设置。
:::

### 更新 Prowlarr-Sonarr 同步的服务器 URL

1. 在 Prowlarr 中，进入**设置** > **应用**，然后选择 **Sonarr**。
2. 将 **Prowlarr Server** 和 **Sonarr Server** 都从 `http://` 改为 `https://`。例如：
   - **Prowlarr Server**：`https://e5e5b409.alexmiles.olares.com`
   - **Sonarr Server**：`https://9691c178.alexmiles.olares.com`
3. 点击 **Test**。绿色对勾表示连接成功。
4. 点击 **Save**。

### 更新 Sonarr 中的下载客户端连接

1. 在 Sonarr 中，进入**设置** > **下载客户端**，然后选择 **qBittorrent**。
2. 按如下方式填写连接信息：
   - **Host**：输入 qBittorrent 的内部入口 URL。
   - **Port**：输入 `443`。
   - **Use SSL**：启用此选项。
3. 点击 **Test**，然后点击 **Save**。

## 更新统一挂载结构的路径

Olares V1.12.6 使用统一的目录挂载结构。请更新 *Arr 应用和下载客户端使用的路径，使它们都能看到相同的位置。

### 更新 Sonarr 根文件夹

强烈建议将根文件夹更新为新的挂载路径。

1. 在 Sonarr 中，进入**设置** > **媒体管理** > **根文件夹**。
2. 将旧的根文件夹路径替换为新的路径。例如：
   - 旧路径：`/home/Movies`
   - 新路径：`/olares/userdata/home/Movies/`
3. 对于每个现有剧集，打开剧集编辑器并将**路径**更新为新的根文件夹。当提示是否移动文件时，选择**否**。

### 移除远程路径映射

如果你之前在 *Arr 应用和下载客户端之间配置了**远程路径映射**，可以将其移除。新的统一挂载结构意味着 *Arr 应用和下载客户端可以看到相同的路径。

## （可选）更新下载客户端

如果你将 qBittorrent、NZBGet、Transmission 或 Deluge 与 *Arr 应用配合使用，更新到 Olares V1.12.6 后，可能还需要修改它们的默认下载路径。新的统一目录挂载结构让 *Arr 应用和下载客户端使用相同的路径。

详细步骤请参阅[下载客户端更新说明](./download-clients-upgrade.md)。
