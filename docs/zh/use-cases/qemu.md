---
outline: [2, 3]
description: 在 Olares 上使用 QEMU Server 运行 Linux 服务器，通过局域网或 LarePass VPN 连接，调整资源，并通过 Clone 创建不同发行版的独立实例。
head:
  - - meta
    - name: keywords
      content: "Olares, QEMU Server, Linux, 虚拟机, SSH, Overlay 网关, LarePass VPN, 克隆"
app_version: "0.0.27"
doc_version: "1.0"
doc_updated: "2026-10-16"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/qemu.md)为准。
:::

# 使用 QEMU Server 运行 Linux 服务器

QEMU Server 通过 QEMU 和 KVM 硬件加速，在 Olares 上运行 Linux 虚拟机。你可以使用 SSH 和浏览器控制台管理服务器，运行命令行工具、开发环境和后台服务。

每个安装实例都有独立的系统盘和资源配置。例如，你可以保留用于开发的 Ubuntu，再通过 Clone 创建独立的 Debian、Rocky Linux 或 Fedora 服务器。

## 学习目标

在本指南中，你将学习如何：

- 安装 Linux 服务器并查看初始化进度。
- 通过同一局域网或 LarePass VPN 连接。
- 使用克隆功能创建其他发行版。
- 调整资源并检查数据持久化。

## 前提条件

- 一台 amd64 或 ARM64 架构的 Olares 设备，已启用硬件虚拟化，并且具有可用的 `/dev/kvm`。虚拟机架构与宿主机一致。
- 足够的可用 CPU、内存和存储空间。除分配给 Linux 的内存外，每个实例还需要为 QEMU 预留约 1 GiB 内存。

## 安装 QEMU Server

1. 打开应用市场，搜索“QEMU Server”。
   <!-- ![应用市场中的 QEMU Server](/images/manual/use-cases/qemu.png#bordered) -->

2. 点击**获取**，然后点击**安装**。按提示设置以下环境变量，然后等待安装完成：
    - **DISTRIBUTION：** 选择 Ubuntu 24.04/26.04 LTS、Debian 12/13、Rocky Linux 9/10 或 Fedora 43/44。
    - **CPU_CORES：** 输入 1–8 个虚拟 CPU 核心，例如 `2`。它们与其他工作负载共享宿主机资源，不代表独占物理核心。
    - **MEMORY_GB：** 输入 1–32 GiB 的 Linux 内存，例如 `2`。
    - **DISK_SIZE_GB：** 输入 16–512 GiB 的虚拟磁盘容量，例如 `40`。
    - **INITIAL_PASSWORD：** 设置 Linux 账户密码，长度为 8–128 个字符。

以上示例仅供参考，不是安装默认值，请在配置表单中自行填写。可选的 CDN 覆盖设置保持不变，即可使用 Olares 配置的 CDN。

## 等待 Linux 初始化

从启动台 打开 QEMU Server。页面会显示系统镜像下载、校验、磁盘准备和初始化进度。请等待服务器就绪后再连接。Olares 中应用显示运行中，并不代表 Linux 已完成下载和初始化。

系统镜像从 Olares CDN 下载，所需时间取决于镜像大小和网络状况。下载中断后会保留部分数据，重试时继续下载。如果页面提示下载失败，重启应用后重新查看进度。

根据发行版使用对应账户，密码为安装时自行设置的密码：

| 发行版 | SSH 用户名 |
| :--- | :--- |
| Ubuntu | `ubuntu` |
| Debian | `debian` |
| Rocky Linux | `rocky` |
| Fedora | `fedora` |

以上账户支持 `sudo`，初始登录账户不是 `root`。

## 使用 SSH 连接

根据电脑与 Olares 的网络位置选择连接方式。两种方式连接的是同一个 Linux 系统，使用相同的账户和密码。

| 方式 | SSH 主机地址 | SSH 端口 | 使用条件 |
| :--- | :--- | :--- | :--- |
| 同一局域网：Overlay 网关 | 分配给当前应用实例的独立 IP | `22` | 已为此实例启用 Overlay 网关，且电脑可以访问该局域网地址。 |
| 不同网络：LarePass VPN | 通过 VPN 可访问的 Olares 设备地址 | 当前实例分配到的 SSH 外部端口 | 电脑已启用 LarePass VPN。 |

以下 IP 和端口均为示例，请替换成设备和应用实际显示的值。浏览器中的应用页面 URL 不能直接作为 SSH 地址。

<tabs>
<template #同一局域网>

### 通过 Overlay 网关连接

Overlay 网关为每个应用实例分配独立的局域网 IP。SSH 直接连接该 IP，使用默认端口 22。

Olares 必须运行在原生 Linux 宿主机上，并使用有线以太网连接，WSL 不支持此功能。电脑可以使用 Wi-Fi，但必须能够访问分配的局域网 IP；访客网络隔离或不同 VLAN 可能阻止连接。

1. 打开 **设置** > **网络** > **Overlay 网关**。如果系统级服务未启用，请由超级管理员开启。

2. 在 **应用** 中，为需要连接的 QEMU Server 实例开启 Overlay 网关，并确认操作。运行中的实例会重启，请等待 Linux 再次就绪。

3. 复制该实例显示的局域网 IP，例如 `192.168.1.51`。

4. 在终端、PowerShell 或其他 SSH 客户端中连接。以 Ubuntu 为例：

   ```bash
   ssh ubuntu@192.168.1.51
   ```

5. 按提示输入 Linux 密码。使用 Xshell 时，选择 SSH 协议，主机填写该实例的 IP，端口填写 `22`。

每个 Clone 实例具有不同的 Overlay IP，因此它们都可以使用端口 22，不会因端口相同而冲突。

</template>
<template #不同网络>

### 通过 LarePass VPN 连接

离开 Olares 所在局域网后，使用 LarePass VPN 建立到 Olares 的网络连接。此时使用 Olares 设备地址和当前实例的 SSH 外部端口，无需先开启 Overlay 网关。

1. 在电脑上打开 LarePass，使用具有设备访问权限的 Olares ID 登录，并启用 VPN。

   ![在电脑上启用 LarePass VPN](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

2. 获取通过 VPN 可以访问的 Olares 设备地址。这里需要的是设备地址，不是应用的网页入口，也不是其他实例的地址。

3. 打开 **设置** > **应用** > **QEMU Server** > **端口**。如果连接 Clone，请选择对应实例的应用名称，找到其 SSH 外部端口。

4. 使用设备地址和外部端口连接。例如，可达设备地址为 `192.168.1.10`，SSH 外部端口为 `43022`：

   ```bash
   ssh -p 43022 ubuntu@192.168.1.10
   ```

5. 输入 Linux 密码。使用 Xshell 时，将相同的设备地址和外部端口分别填入主机与端口字段。

请使用 设置中显示的外部端口，不要直接替换成 22 或应用内部服务端口。同一台 Olares 上的不同实例使用不同的 SSH 外部端口。公网网页入口并不等于公网 SSH 入口；此流程通过 LarePass VPN 访问。实际速度取决于网络状况和 VPN 路由。

</template>
</tabs>

:::tip SSH 端口写法
macOS、Linux 和 Windows OpenSSH 使用 `-p` 指定自定义端口，例如 `ssh -p 43022 ubuntu@192.168.1.10`。不要在主机名后直接拼接 `:43022`，也不要添加 `http://` 或 `https://`。
:::

## 使用浏览器控制台

从 QEMU Server 应用页面打开控制台，使用 Linux 账户登录。控制台适合查看启动输出，或在 SSH 暂时不可用时进行操作。这里提供的是 Linux 服务器控制台，不是图形桌面。

## 使用 Clone 创建其他发行版

Clone 使用相同应用模板创建全新、独立的安装实例，不会复制现有虚拟机的系统盘、已安装软件、文件或 Linux 账户。原服务器继续作为独立实例保留。

以下以保留 Ubuntu、另建 Debian 服务器为例：

1. 打开应用市场，进入 **我的 Olares**。

2. 找到已安装的 QEMU Server，打开 **打开** 旁的下拉菜单，选择 **克隆**。

3. 在 **克隆应用** 中，将 **新应用名称** 设置为容易区分的名称，例如 `Debian Services`，并设置 **桌面快捷方式名称**，然后点击 **确认**。

4. 在 **配置环境变量** 中，将 `DISTRIBUTION` 选择为 Debian 13，填写新实例的 CPU、内存、磁盘容量和密码，然后确认。

5. 打开新实例，等待镜像下载和初始化完成，使用 `debian` 用户及其独立密码登录。

6. 需要局域网访问时，为 Clone 单独开启 Overlay 网关；需要 VPN 访问时，查询该实例自己的 SSH 外部端口。

建议记录每个实例的名称、发行版、用户名和连接信息，例如：

| 实例 | 发行版 | 用户名 | 同一局域网连接 | VPN 连接 |
| :--- | :--- | :--- | :--- | :--- |
| Ubuntu Dev | Ubuntu 24.04 LTS | `ubuntu` | 自己的 Overlay IP，端口 22 | 设备地址，Ubuntu 实例的外部端口 |
| Debian Services | Debian 13 | `debian` | 另一个 Overlay IP，端口 22 | 相同设备地址，Debian 实例的外部端口 |

现有系统盘不能通过修改配置直接切换发行版。需要其他发行版或版本时，请创建新的独立实例。Clone 会额外消耗宿主机资源，所有实例的合计资源需求应在设备可用容量之内。

## 调整 CPU、内存和磁盘容量

1. 备份重要数据，并结束正在进行的工作。应用资源修改会重启虚拟机，断开 SSH 会话。

2. 打开 **设置** > **应用**，选择需要修改的实例。

3. 在 **环境变量** 中点击 **管理环境变量**。

4. 在支持范围内修改 `CPU_CORES`、`MEMORY_GB` 或 `DISK_SIZE_GB`，确认修改后点击 **应用**。

5. 等待 Linux 就绪，重新连接并检查资源：

   ```bash
   nproc
   free -h
   lsblk
   df -h /
   ```

增加磁盘容量后，重启过程会扩展虚拟磁盘及支持自动扩容的根文件系统。由于分区和文件系统元数据也占用空间，根文件系统容量小于整块磁盘容量。Linux 显示的可用内存也可能略低于配置值。

:::warning 不支持缩小磁盘
填写小于现有磁盘的容量时，应用会保留原有的大磁盘，并显示警告，不会回收空间。需要更小的磁盘时，请备份数据后创建容量更小的独立实例。
:::

## 了解数据持久化

虚拟机系统盘保存在当前实例的 Olares App Data 中。写入系统盘的 Linux 软件包、用户文件和服务数据会在应用重启、资源调整后保留。各 Clone 实例使用独立存储。

系统盘不会自动将 Olares Files 中的文件夹挂载到 Linux。可以通过 SFTP、SCP 等基于 SSH 的工具传输文件，或另行配置网络共享。持久化不等于备份，请单独备份重要数据。卸载实例并删除应用数据会移除该虚拟机的系统盘。

## 排查连接问题

| 现象 | 检查内容 |
| :--- | :--- |
| 无法解析主机名 | 移除主机名后的 `:端口`，使用 `-p` 指定自定义 SSH 端口。 |
| 局域网连接超时 | 检查当前实例的 Overlay IP、网关状态、Olares 有线连接，以及局域网路由和隔离规则。 |
| 远程连接超时 | 确认 LarePass VPN 已启用、设备地址可达，且 SSH 外部端口属于目标实例。 |
| 密码被拒绝 | 检查发行版对应的用户名和当前 Linux 密码。在 Linux 内改过密码后，应使用新密码，而非安装时的初始密码。 |
| 应用运行中但 SSH 不可用 | 打开应用页面，等待 Linux 初始化；必要时使用浏览器控制台检查启动状态。 |
| 镜像下载失败 | 重启应用，使用保留的部分数据重试下载。 |

## 了解更多

- [验证 QEMU Server 安装](qemu-validation.md)：可重复执行的验收用例及验证覆盖范围。
- [克隆应用](../manual/olares/market/clone-apps.md)：管理多个应用实例。
- [管理 Overlay 网关](../manual/olares/settings/overlay-gateway.md)：使用条件、权限与网络设置。
- [管理应用环境变量](../manual/olares/settings/manage-app-env.md)：修改实例配置。
