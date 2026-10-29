---
outline: [2, 3]
description: 了解如何在 NVIDIA DGX Spark 上通过脚本安装 Olares，以及如何停止、重新启动和安全卸载 Olares。
head:
  - - meta
    - name: keywords
      content: Olares, DGX Spark, 命令行安装, NVIDIA, GPU 显存
---

# 在 Spark 上使用命令行安装 Olares

本文介绍如何在 NVIDIA DGX Spark 上通过脚本安装 Olares。如果你已通过脚本完成安装，也可以直接查看[停止和启动 Olares](#停止和启动-olares) 或[卸载 Olares](#卸载-olares)。

## 前提条件

- **Olares 设备**：一台至少有 150 GB 可用 SSD 存储空间的 NVIDIA DGX Spark。
<!--@include: ./reusables.md#larepass-prerequisite-->
- **初始设置**：DGX Spark 已完成[首次启动设置](https://docs.nvidia.com/dgx/dgx-spark/first-boot.html)，包括创建用户账号和配置网络。
- **终端访问**：使用以下任一方式：
  - 直接访问：将显示器、键盘和鼠标连接到 DGX Spark。
  - 远程访问：通过 SSH 从同一网络的另一台电脑连接。
- **网络**：建议使用网线连接 DGX Spark 和路由器。

## 准备 DGX Spark

在安装 Olares 之前，你需要移除 DGX Spark 上预装的容器运行时。

Olares 会安装并管理自己的 containerd 运行时。如果系统已安装 Docker 或 containerd，将会导致冲突，使 Olares 无法正常运行。

:::warning 移除容器运行的影响
以下操作将：
- 卸载 Docker 软件包
- 停止并禁用现有的 containerd 服务
- 清除现有的网络规则
:::

在 DGX Spark 上打开终端并运行：

```bash
sudo apt remove docker*
sudo systemctl disable --now containerd
sudo rm -f /usr/bin/containerd
sudo nft flush ruleset
```

## 安装 Olares

1. 在同一个终端中，运行以下命令：

<!--@include: ./reusables.md#install-script-command-->

<!--@include: ./reusables.md#root-password-tip-->

<!--@include: ./reusables.md#installation-error-tip-->

<!--@include: ./reusables.md#prepare-wizard-heading-->

<!--@include: ./reusables.md#prepare-wizard-details-->

## 激活 Olares

使用向导 URL 和初始一次性密码进行激活和 Olares 初始化配置。

1. 在浏览器中输入向导 URL。进入欢迎页面后，按任意键继续。

   ![打开向导](/images/manual/get-started/open-wizard.png#bordered)
2. 输入一次性密码，点击**继续**。

   ![输入密码](/images/manual/get-started/wizard-enter-password1.png#bordered)
3. 选择系统语言。Olares 支持英语、简体中文、德语、西班牙语、意大利语、法语和日语。选择后，向导会立即切换语言。激活后，Olares 桌面也会使用该语言。

   ![选择语言](/images/manual/get-started/select-language.png#bordered)
4. 选择一个距你所在位置最近的反向代理节点。你也可以之后在 Olares 的[更改反向代理](../olares/settings/change-frp.md)页面进行调整。

   ![选择 FRP](/images/zh/manual/get-started/wizard-frp.png#bordered)

5. 使用 LarePass 应用激活 Olares。

   :::warning 网络要求
   为避免激活失败，管理员用户需确保手机和 Olares 设备连接到同一网络。
   :::
   a. 打开 LarePass 应用，点击**扫描二维码**，扫描向导页面上的二维码完成激活。

   ![激活 Olares](/images/manual/get-started/activate-olares1.png#bordered)

   b. 按照 LarePass 上的提示重置 Olares 的登录密码。

设置成功后，LarePass 应用会自动返回主界面，向导页面则会跳转到登录界面。

<!--@include: ./log-in-to-olares.md-->

## 配置 AI 应用的 GPU 显存

DGX Spark 采用统一内存架构，CPU 和 GPU 共享 128 GB 的 LPDDR5x 内存。与传统 GPU 拥有独立显存不同，DGX Spark 不区分系统内存和 GPU 显存。

在 DGX Spark 上，Olares 默认使用**容量分片**模式进行 GPU 资源管理。当你安装 AI 应用时，Olares 会自动分配最低所需的内存，以确保应用能够正常启动和运行。

如需调整某个应用的内存分配，可以手动修改：

1. 从 Olares 打开**设置**，然后进入 **GPU** 页面。
2. 在**分配显存**区域，找到目标应用。

   ![显存切片](/images/zh/manual/get-started/install-spark-memory-slicing.png#bordered){width=70%}

3. 点击显存值旁边的 <i class="material-symbols-outlined">edit_square</i>。
4. 在**编辑显存分配**对话框中，输入所需的显存大小（GB），然后点击**确认**。

## 停止和启动 Olares

如果想暂时停止 Olares 的服务和应用，同时保留数据，可以使用 `stop` 命令。停止后，DGX Spark 仍保持开机，但 Olares 桌面和应用暂时无法访问，需要重新启动 Olares 后才能使用。

在 DGX Spark 上打开终端，或通过 SSH 连接后执行以下命令。

1. 保存当前工作，等待模型训练、文件传输等任务完成。
2. 停止 Olares：

   ```bash
   sudo olares-cli stop
   ```

   等待命令执行完毕，确认输出包含 `StopOlares execute successfully` 且没有报错。停止不会删除已安装的应用、配置或用户数据，也不会禁用 DGX Spark 重启后的自动启动。

3. 需要继续使用时，启动 Olares：

   ```bash
   sudo olares-cli start
   ```

   等待命令执行完毕，然后打开 Olares 桌面，确认可以登录并访问应用。

## 卸载 Olares

如果不再需要在 DGX Spark 上运行 Olares，可以将其卸载。加上 `--all` 选项还会移除安装准备阶段安装的依赖，包括 Olares 使用的容器运行时。

:::warning 卸载前先备份
卸载会删除 Olares 的配置、已安装的应用和用户数据。即使不加 `--all`，`uninstall` 也会删除这些数据。

请在 Olares 仍在运行时，将[文件和支持备份的应用数据](../olares/settings/backup.md)备份到外部存储或云存储。其他应用的数据请使用各自的备份工具导出，并确认备份已成功完成。如果 Olares 已停止，请先启动。若要保留安装和数据，请改用 `stop`。
:::

1. 在 DGX Spark 上打开终端，或通过 SSH 连接后运行：

   ```bash
   sudo olares-cli uninstall --all
   ```

2. 等待命令执行完毕，确认输出包含 `Uninstall Olares execute successfully`，并检查是否有报错。

卸载不会恢复[准备 DGX Spark](#准备-dgx-spark) 时移除的 Docker 或原有 containerd 环境。如需继续使用，请单独重新安装和配置。

<!--@include: ./reusables.md#protect-olares-id-->

<!--@include: ./reusables.md#installation-troubleshooting-tip-->
