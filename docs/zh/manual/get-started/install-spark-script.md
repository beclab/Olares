---
outline: [2, 3]
description: 在 NVIDIA DGX Spark 上通过命令行安装脚本快速部署 Olares。
---

# 使用命令行安装 Olares

本文介绍如何在 NVIDIA DGX Spark 上通过命令行安装脚本安装 Olares。

<!--@include: ./reusables.md{44,51}-->

## 系统要求

- **DGX Spark**：确保设备已完成[首次启动设置](https://docs.nvidia.com/dgx/dgx-spark/first-boot.html)，创建了用户账户并配置好网络。
- **存储**：DGX Spark 上至少有 150 GB 的可用 SSD 存储空间。
- **访问方式**：你需要访问 DGX Spark 的终端，可通过以下任一方式：
  - 直接访问：将显示器、键盘和鼠标连接到 DGX Spark。
  - 远程访问：通过 SSH 从同一网络的另一台电脑连接。
- **网络**：建议使用网线将 DGX Spark 连接到路由器，以获得稳定的网络连接。

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

<!--@include: ./reusables.md{4,36}-->

## 激活 Olares

使用向导 URL 和初始一次性密码进行激活和 Olares 初始化配置。

1. 在浏览器中输入向导 URL。进入欢迎页面后，按任意键继续。

   ![打开向导](/images/manual/get-started/open-wizard.png#bordered)
2. 输入一次性密码，点击**继续**。

   ![输入密码](/images/manual/get-started/wizard-enter-password1.png#bordered)
3. 选择系统语言。

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

在 DGX Spark 上，Olares 默认使用**显存分片**模式进行 GPU 资源管理。当你安装 AI 应用时，Olares 会自动分配最低所需的内存，以确保应用能够正常启动和运行。

如需调整某个应用的内存分配，可以手动修改：

1. 从 Olares 打开**设置**，然后进入 **GPU** 页面。
2. 在**分配显存**区域，找到目标应用。

   ![显存切片](/images/zh/manual/get-started/install-spark-memory-slicing.png#bordered){width=70%}

3. 点击显存值旁边的 <i class="material-symbols-outlined">edit_square</i>。
4. 在**编辑显存分配**对话框中，输入所需的显存大小（GB），然后点击**确认**。

<!--@include: ./reusables.md{38,42}-->
