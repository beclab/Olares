---
outline: [2, 3]
description: 排查 Olares 上 Steam Headless 的常见问题，包括应用重启和升级后的软件包持久化问题。
head:
  - - meta
    - name: keywords
      content: Olares, Steam Headless, 常见问题, Flatpak, apt, 应用持久化, 故障排查
app_version: "1.0.43"
doc_version: "1.0"
doc_updated: "2026-09-03"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/steam-common-issues.md)为准。
:::

# Steam Headless 常见问题

查找 Olares 上 Steam Headless 常见问题的解决方法。

## 为什么通过 `apt` 安装的软件包会在 Steam Headless 重启后消失？

通过 `apt` 安装的软件包会写入容器的根文件系统。当 Steam Headless 重启、重新部署或升级时，这个文件系统会重新创建。因此，通过 `apt` 手动安装的软件包不会保留。

如需让额外安装的软件包在 Steam Headless 重启和升级后继续保留，如果有对应的 Flatpak 软件包，请使用 Flatpak，而不是 `apt`。Steam Headless 1.0.43 及更高版本会将 Flatpak 应用、运行时和用户数据保存在持久化应用存储中。

通过 Flatpak 安装软件包：

1. 打开 Control Hub，前往 **Browse** > **steamheadless**。
2. 展开 **Deployments** > **steamheadless**，然后打开正在运行的 Pod。
3. 在 **Containers** 下，点击 **steam-headless** 旁边的 Terminal 图标。
4. 在容器 Shell 中执行 Flatpak 安装命令。

有关 Pod 和容器的更多信息，请参阅[管理容器](../manual/olares/controlhub/manage-container.md)。
