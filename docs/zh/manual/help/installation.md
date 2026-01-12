---
outline: [2, 3]
description: 查找有关 Olares 安装与激活过程中的常见问题解答。
---
# Olares 安装与激活常见问题

本文汇总了关于在硬件上安装、配置及激活 Olares 的常见问题。

## 安装
### Olares 支持哪些平台？

为了获得最佳性能，建议在 Linux (Ubuntu 或 Debian) 上安装 Olares。

若仅用于产品体验，也可以在以下平台安装：
* Proxmox VE
* Raspberry Pi
* macOS
* Windows

### 安装 Olares 的最低硬件要求是什么？

具体要求因平台而异。通常建议配置如下：
* **CPU**：至少 4 核，x86-64 架构 (Intel 或 AMD)。
* **内存**：至少 8 GB 可用内存。
* **存储**：至少 150 GB SSD。

详细要求参见[安装文档](../get-started/install-olares.md)。

### 可以使用机械硬盘安装 Olares 吗？

不可以。必须使用 SSD。由于机械硬盘读写速度较慢，极易导致系统初始化超时，从而造成安装失败。

### 系统支持 NVIDIA 显卡吗？

支持。Olares 针对 NVIDIA 硬件进行了深度优化。系统会自动处理驱动安装，让你即刻获得 AI 和游戏性能加速。

系统还支持单主板多 GPU 配置 (目前仅限 NVIDIA)，允许自定义硬件用户利用所有可用算力处理 AI 负载。

### 如果自动设置失败，如何手动安装 NVIDIA 驱动？

Olares 安装程序通常会自动检测并安装驱动。但如果系统中此前已安装过 NVIDIA 驱动，可能会因冲突导致安装过程跳过或失败。

此时，按以下步骤操作：
1. Olares 安装完成后重启机器，确保旧的驱动组件被完全清理。
2. 使用命令 `olares-cli gpu install` 手动触发驱动安装。

安装完成后，运行 `nvidia-smi` 确认驱动已安装且 GPU 被正确识别。

### 为什么安装时报错 `failed to build Kubernetes objects` 或 `Ensure CRDs are installed first`？

虽然报错信息指向 Custom Resource Definitions (CRD) 问题，但这通常是磁盘性能不足的表现。

Olares 依赖 Kubernetes 的后端数据库 etcd。etcd 对存储速度非常敏感。如果在传统 HDD 等低速硬盘上安装，etcd 无法及时响应，从而导致 API Server 在尝试应用 CRD 时超时。

更换为 SSD 存储通常能解决此问题。

### Olares 安装超时未显示密码，但系统似乎在运行。如何找回密码？

这通常是由于系统资源不足导致安装超时，特别是在虚拟机环境中。你可以通过以下命令从安装日志中获取密码：
```bash
# 将 v1.12.2 替换为你的具体 Olares 版本号
grep password $HOME/.olares/versions/v1.12.2/logs/install.log
```
安装超时通常意味着部分服务启动失败。找回密码后，运行 `kubectl get pod -A` 检查所有服务的状态。

## 激活
### 可以在非局域网环境下激活 Olares 吗？

可以。通常情况下，用户通过局域网 IP 访问激活向导，这要求双方在同一网络下。但如果 Olares 被分配了公网 IP (如在公有云上)，则不受此局域网限制。

注意，IP 访问仅用于激活过程。激活完成后，无论在内网还是外网，均通过域名访问设备。

### Olares 已开机并连接局域网，但 LarePass 搜不到设备。怎么办？

确保手机和 Olares 设备处于同一网络。如果不在同一网络，LarePass 无法自动发现 Olares。

如果无法通过 Wi-Fi 连接，你可以利用 LarePass 应用中的蓝牙配网功能，将 Olares 连接到与手机相同的网络中。

详见[使用蓝牙激活 Olares](../larepass/activate-olares.md#通过蓝牙激活)。