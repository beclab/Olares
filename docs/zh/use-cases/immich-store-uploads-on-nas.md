---
outline: [2, 3]
description: 将通过 Immich 上传的照片和视频直接存储到 NAS，同时将数据库保留在 Olares 中。
head:
  - - meta
    - name: keywords
      content: Olares, Immich, NAS 上传存储, UPLOAD_LOCATION, NFS, SMB, 照片备份
app_version: "1.0.37"
doc_version: "1.0"
doc_updated: "2026-10-27"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/immich-store-uploads-on-nas.md)为准。
:::

# 将 Immich 上传的照片和视频存储到 NAS

Immich 可以将通过网页或手机应用上传的照片和视频直接保存到 NAS，数据库仍保留在 Olares 中。缩略图、转码视频等生成的媒体文件也会保存到所选目录。

如果你希望浏览 NAS 上已有的照片，而不改变 Immich 的上传位置，请参阅[从 NAS 导入照片到 Immich](immich-import-from-nas.md)。该教程介绍的是外部图库，与本文的上传存储设置不同。

## 准备工作

- 使用 Olares 应用包版本 1.0.37（Immich 3.3.5），或后续提供 `UPLOAD_LOCATION` 配置的版本。如果找不到该配置，说明当前安装包尚不支持此功能。
- 将 Olares 与 NAS 连接到同一网络。
- 准备具有读写权限的 NFS 或 SMB 共享。NFS 共享需允许 UID/GID `1000:1000` 读写；SMB 共享需使用具有读写权限的 NAS 账户挂载。

本文使用实际演示环境中的 NFS 共享 `nfs_share2`，在其中创建专用媒体目录 `immich-upgrade-20261027-Kk3SeD`。这是演示时创建的测试目录，不是 Immich 要求的固定目录；名称中的日期用于标识测试，末尾的随机字符用于避免重名。文中的地址和目录名称仅作为示例，请替换为你自己的配置，不要直接照抄。

建议在首次上传照片前完成配置。如果已经使用 Immich，请先按照“迁移已有媒体”一节处理原有文件。

## 挂载 NAS 共享目录

1. 打开文件管理器，进入 **Drive > 外部存储**，点击 **连接服务器**。
2. 选择 **NFS**，填写共享地址。本次实际演示环境使用：

   ```text
   192.168.50.124:/volume1/nfs_share2
   ```

   请将该示例地址替换为你自己的 NAS 地址和 NFS 共享路径。使用 SMB 时，选择 **SMB**，填写 `//<NAS-IP>/<共享目录>`，并按提示输入具有读写权限的 NAS 账户和密码。完整 NFS 操作步骤请参阅[挂载 NFS 共享](../manual/olares/files/mount-nfs.md)。
3. 确认连接后，检查外部存储中显示的共享目录名称。本次演示为 `nfs_share2`。
4. 在共享目录中先创建一个空的专用媒体文件夹，并确认可以在其中创建文件。目录名称可以自行决定，例如 `immich`；本次演示创建的测试目录名为 `immich-upgrade-20261027-Kk3SeD`。

请为 Immich 使用专用目录，将其他 NAS 文件保存在该目录之外。

## 设置上传位置

1. 打开 **设置 > 应用 > Immich**。
2. 在环境变量区域，点击 **管理环境变量**。
3. 编辑 `UPLOAD_LOCATION`，填写完整媒体目录路径。本次实际演示环境使用：

   ```text
   /external_storage/nfs_share2/immich-upgrade-20261027-Kk3SeD
   ```

   其中，`nfs_share2` 是演示环境中的挂载目录名称，`immich-upgrade-20261027-Kk3SeD` 是提前创建的专用媒体目录。请根据你自己的挂载目录和媒体目录填写完整路径，不要直接照抄，也不要在此处填写 NAS 的 IP 地址。

   例如，如果你在 `nfs_share2` 共享中创建的目录名为 `immich`，应填写：

   ```text
   /external_storage/nfs_share2/immich
   ```
4. 点击 **确认**，再点击 **应用**，等待 Immich 重新启动。

这里需要填写以 `/external_storage/` 开头的应用内路径，不是 SMB 地址或 NFS 共享地址。

| `UPLOAD_LOCATION` 的值 | 上传文件保存位置 |
| --- | --- |
| 留空 | Olares 中默认的本地媒体目录 |
| 已存在的 `/external_storage/` 下的目录 | 指定的 NFS 或 SMB 目录 |

只需配置这一个变量。NAS 共享本身的挂载和卸载仍在文件管理器中操作。

## 验证上传结果

1. 打开 Immich，上传一张新照片。
2. 确认缩略图正常显示，并且可以打开和下载原图。
3. 在文件管理器中打开 NAS 上的专用目录，检查 `upload` 或 `library` 等子目录中是否出现新文件。

Immich 会管理目录结构和存储文件名，因此 NAS 中的文件名可能与照片详情中显示的原文件名不同。数据库仍使用 Olares 存储。

## 迁移已有媒体

修改 `UPLOAD_LOCATION` 只会改变应用使用的媒体目录，不会复制或同步文件。

1. 在应用设置中停止 Immich。
2. 备份 PostgreSQL 数据库和完整媒体目录。备份方法可参考 Immich 的[备份与恢复指南](https://docs.immich.app/administration/backup-and-restore/)。
3. 将完整媒体目录的内容复制到目标位置，包括隐藏的 `.immich` 文件和所有子目录。当前应用包的默认本地媒体目录位于 **文件管理器 > Drive > Data > immich > data > upload**。
4. 检查目标目录中的文件是否完整，比较文件数量、大小和校验值。
5. 将 `UPLOAD_LOCATION` 设置为目标目录并应用。确认原有照片可以打开、原图可以下载，并测试一次新上传，再考虑删除旧副本。

目标目录应保持原有媒体结构。配置的路径下应直接包含 `upload`、`library`、`thumbs`、`encoded-video`、`profile`、`backups` 等目录，不要额外嵌套一层。

如果要切回默认本地存储，请先将完整媒体目录恢复到本地，再清空 `UPLOAD_LOCATION`。以后重新使用 NAS 时，也需要将本地存储期间新增的媒体复制过去。

## NAS 不可用时如何处理

如果配置了 NAS 路径，但启动时目标目录未挂载，Immich 会阻止启动，不会自动将新文件写入本地。

- **继续使用 NAS**：在文件管理器中恢复挂载，确认目标目录可访问，再应用上传位置配置。
- **改用本地存储**：将备份中的媒体恢复到默认本地目录，清空 `UPLOAD_LOCATION` 并应用。本地模式启动不依赖 NAS 挂载；尚未恢复到本地的 NAS 照片仍无法访问。

:::info 市场状态可能未同步恢复
缺失挂载导致启动失败后，即使 Immich 已经可以访问，Olares 市场仍可能保留失败状态。切换到本地存储后也观察到这一情况。如果状态持续异常，请联系支持，不要通过删除图库或数据库来清除错误标记。
:::

## 了解更多

- [使用 Immich 管理照片](immich.md)：了解 Immich 的初始化和照片管理功能。
- [从 NAS 导入照片到 Immich](immich-import-from-nas.md)：将已有 NAS 文件夹添加为外部图库。

- [管理应用环境变量](../manual/olares/settings/manage-app-env.md)：了解如何在 Olares 设置中修改应用配置。
