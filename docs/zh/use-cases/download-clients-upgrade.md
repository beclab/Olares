---
outline: deep
description: 升级到 Olares 1.12.6 时，qBittorrent、NZBGet、Transmission 和 Deluge 下载客户端的升级说明。
head:
  - - meta
    - name: keywords
      content: Olares, 下载客户端, qBittorrent, NZBGet, Transmission, Deluge, 升级, 1.12.6
app_version: "1.0.x"
doc_version: "1.0"
doc_updated: "2026-07-02"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/download-clients-upgrade.md)。
:::

# 下载客户端升级说明

本文介绍在将 Olares 升级到 V1.12.6 后，如何更新下载客户端中的默认下载路径。新的统一目录挂载结构意味着 *Arr 应用和下载客户端可以看到相同的路径，因此你可能需要更新现有的下载位置。

如果你还使用 Sonarr 或 Radarr 等 *Arr 应用，请参阅 [*Arrs 升级说明](./arrs-upgrade.md)，了解应用连接和根文件夹的更改。

## qBittorrent

1. 打开 qBittorrent。
2. 在工具栏中选择 **工具** > **选项**。
3. 在 **选项** 窗口中，进入 **下载** 选项卡。
4. 将 **默认保存路径** 从 `/downloads/home/Downloads/qBittorrent` 更新为 `/olares/userdata/home/Downloads/qBittorrent`。

   ![qBittorrent 默认保存路径](/images/manual/use-cases/update-default-save-path.png#bordered)
   
5. 向下滚动并点击 **保存**。

### 升级后 qBittorrent 要求登录

如果 qBittorrent 之前允许你无需登录即可访问 WebUI，但现在提示输入凭据，可能是迁移脚本没有保留认证设置。

1. 打开 Control Hub，查看 **qBittorrent** 容器日志，找到临时用户名（通常为 `admin`）和密码。
2. 登录 qBittorrent 并确认其正常工作。
3. 根据需要设置新的用户名和密码。
4. 要恢复免密码访问：

   a. 打开 Olares 文件，编辑 `/Data/qbittorrent/qBittorrent.conf`。

   b. 在 `[Preferences]` 部分，添加或更新以下值：

      ```ini
      WebUI\Address=*
      WebUI\ServerDomains=*
      WebUI\Port=8080
      WebUI\CSRFProtection=false
      WebUI\HostHeaderValidation=false
      WebUI\LocalHostAuth=false
      WebUI\AuthSubnetWhitelistEnabled=true
      WebUI\AuthSubnetWhitelist=0.0.0.0/0, ::/0
      ```

   c. 保存文件并重启 qBittorrent。

如果问题仍然存在，请备份现有配置，删除 `qBittorrent.conf`，然后重启 qBittorrent。初始化过程将重新创建一个默认配置，并启用免密码访问。之后，你可以逐行重新应用之前的设置，以找出冲突的选项。

## NZBGet

1. 打开 NZBGet，进入 **设置** > **PATHS**。
2. 更新以下路径：
   - **DestDir**：将 `/downloads/completed` 改为 `/olares/userdata/home/Downloads/nzbget/completed`
   - **InterDir**：将 `/downloads/intermediate` 改为 `/olares/userdata/home/Downloads/nzbget/intermediate`
3. 保存设置并重启 NZBGet。

## Transmission

1. 打开 Transmission，点击右上角的 <i class="material-symbols-outlined">menu</i>，然后选择 **编辑首选项**。
2. 更新以下路径：
   - **下载到**：将 `/downloads/complete` 改为 `/olares/userdata/home/Downloads/transmission/complete`
   - **使用临时文件夹**：将 `/downloads/incomplete` 改为 `/olares/userdata/home/Downloads/transmission/incomplete`

## Deluge

1. 打开 Deluge，进入 **首选项** > **下载**。
2. 将下载路径从 `/downloads` 更新为 `/olares/userdata/home/Downloads/deluge`。
3. 点击 **应用**，然后点击 **确定**。
