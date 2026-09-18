---
outline: [2, 3]
description: 排查 Olares 上 Steam Headless 的常见问题，包括软件包持久化和《黑神话：悟空》帧生成问题。
head:
  - - meta
    - name: keywords
      content: Olares, Steam Headless, 常见问题, Flatpak, apt, 黑神话悟空, DLSS 帧生成, 故障排查
app_version: "1.0.46"
doc_version: "1.1"
doc_updated: "2026-09-18"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/steam-common-issues.md)为准。
:::

# Steam Headless 常见问题

查找 Olares 上 Steam Headless 常见问题的解决方法。

## 通过 `apt` 安装的软件包在 Steam Headless 重启后消失

通过 `apt` 安装的软件包会写入容器的根文件系统。当 Steam Headless 重启、重新部署或升级时，这个文件系统会重新创建。因此，通过 `apt` 手动安装的软件包不会保留。

如需让额外安装的软件包在 Steam Headless 重启和升级后继续保留，如果有对应的 Flatpak 软件包，请使用 Flatpak，而不是 `apt`。Steam Headless 1.0.43 及更高版本会将 Flatpak 应用、运行时和用户数据保存在持久化应用存储中。

通过 Flatpak 安装软件包：

1. 打开 Control Hub，前往 **Browse** > **steamheadless**。
2. 展开 **Deployments** > **steamheadless**，然后打开正在运行的 Pod。
3. 在 **Containers** 下，点击 **steam-headless** 旁边的 Terminal 图标。
4. 在容器 Shell 中执行 Flatpak 安装命令。

有关 Pod 和容器的更多信息，请参阅[管理容器](../manual/olares/controlhub/manage-container.md)。

## 《黑神话：悟空》无法开启帧生成

通过 Proton 运行《黑神话：悟空》时，游戏的图形设置中可能无法开启**帧生成**。Steam Headless 1.0.46 及更高版本内置了修复脚本，用于在游戏的 Proton 环境中配置 DirectX 12、DLSS 和硬件加速 GPU 调度。

1. 在 Steam 中安装《黑神话：悟空》。
2. 启动一次游戏，让 Steam 创建 Proton 环境，然后完全退出游戏。
3. 打开 Control Hub，前往 **Browse** > **steamheadless**。
4. 展开 **Deployments** > **steamheadless**，然后打开正在运行的 Pod。
5. 在 **Containers** 下，点击 **steam-headless** 旁边的 Terminal 图标。
6. 执行以下命令：

   ```bash
   /usr/bin/fix-wukong-frame-gen.sh
   ```

7. 确认命令返回以下信息：

   ```plain
   [OK] Done. Launch Black Myth: Wukong and enable Frame Generation in the graphics menu.
   ```

8. 启动游戏，在图形设置中开启**帧生成**。

重新安装或更新游戏后，需要再次执行该脚本。如果脚本提示游戏正在运行，请完全退出游戏后重试。
