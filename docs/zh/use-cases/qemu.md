---
outline: [2, 3]
description: 在 Olares 上使用 QEMU Server 运行 Linux 服务器，通过 SSH 连接，调整 CPU、内存和磁盘容量，并为不同任务创建独立服务器。
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

QEMU Server 可以在 Olares 上运行 Linux 虚拟机。你可以用它运行命令行工具、搭建开发环境或部署服务，并通过 SSH 或浏览器控制台管理服务器。

本指南先以 Ubuntu 为例创建服务器，再介绍如何通过克隆功能添加一台独立的 Debian 服务器。每个实例都有自己的系统盘和资源配置。

## 前提条件

- 一台 amd64 或 ARM64 架构、已启用硬件虚拟化的 Olares 设备。
- 足够的可用 CPU、内存和存储空间，用于运行 Linux 服务器。除了分配给 Linux 的内存，还需为 QEMU 额外预留 1 GiB 内存。

## 安装和配置 QEMU Server

### 安装应用

1. 打开应用市场，搜索“QEMU Server”。

2. 点击**获取**，然后点击**安装**。按提示填写以下配置，本例使用 Ubuntu 24.04 LTS。

   | 字段 | 填写内容 |
   | :--- | :--- |
   | **DISTRIBUTION** | 选择 Ubuntu 24.04 LTS。其他选项见[可以选择哪些 Linux 发行版？](#可以选择哪些-linux-发行版) |
   | **CPU_CORES** | 分配 1–8 个虚拟 CPU 核心，例如 `2`。 |
   | **MEMORY_GB** | 为 Linux 分配 1–32 GiB 内存，例如 `2`。 |
   | **DISK_SIZE_GB** | 设置 16–512 GiB 磁盘容量，例如 `40`。 |
   | **INITIAL_PASSWORD** | 设置 Linux 账户密码，长度为 8–128 个字符。 |

   根据设备的可用资源填写。

3. 点击**确认**，等待应用安装完成。

### 初始化 Linux

1. 从启动台打开 QEMU Server，查看系统镜像下载和初始化进度。

   ![Linux 镜像下载和初始化进度](/images/manual/use-cases/qemu-initialization.png#bordered)

   下载时间取决于镜像大小和网络状况。如果下载失败，打开 **设置** > **应用** > **QEMU Server**，点击**暂停**，再点击**恢复**，利用已下载的数据继续重试。

2. 等待下载和初始化完成。页面显示服务器就绪后，即可通过 SSH 或浏览器控制台连接。

   ![Linux 系统初始化完成](/images/manual/use-cases/qemu-initialization-finished.png#bordered)

### 获取用户名和密码

Ubuntu 的用户名为 `ubuntu`，密码是你在安装时设置的密码。不同发行版使用的用户名如下：

| 发行版 | 用户名 |
| :--- | :--- |
| Ubuntu | `ubuntu` |
| Debian | `debian` |
| Rocky Linux | `rocky` |
| Fedora | `fedora` |

SSH、浏览器控制台和文件传输均使用这组账户信息。各账户都可以通过 `sudo` 执行管理命令。

## 使用 SSH 连接

电脑与 Olares 在同一局域网时，使用 [Overlay 网关](../manual/olares/settings/overlay-gateway.md)连接。在其他网络下，使用 [LarePass VPN](../manual/larepass/private-network.md)。以下示例使用 Ubuntu 账户，连接其他发行版时，请将 `ubuntu` 替换为对应的用户名。

### 同一局域网：通过 Overlay 网关连接

Overlay 网关为服务器分配独立的局域网 IP，SSH 使用端口 `22`。使用这种方式，需要将 Olares 直接安装在 Linux 上，并通过网线连接网络。你的电脑可以使用同一局域网中的 Wi-Fi。

1. 打开 **设置** > **网络** > **Overlay 网关**。如果系统级服务尚未开启，请让超级管理员启用 Overlay 网关。

2. 在 **应用** 中，为 QEMU Server 实例开启网关并点击**确认**。实例会重启，请等待 Linux 再次完成初始化。

3. 复制该实例显示的局域网 IP。重启或网络变化后，IP 可能改变，请使用页面上的当前地址。

4. 打开终端或 PowerShell，将以下命令中的 `192.168.1.51` 替换为复制的 IP，然后运行：

   ```bash
   ssh ubuntu@192.168.1.51
   ```

5. 按提示输入 Linux 密码。

使用 Xshell 等 SSH 客户端时，选择 SSH 协议，主机填写实例的局域网 IP，端口填写 `22`，再输入相同的用户名和密码。

### 远程访问：通过 LarePass VPN 连接

在其他网络下，可以通过 LarePass VPN，使用 Olares 域名和实例的 SSH 导出端口连接服务器。

1. 在电脑上打开 LarePass，使用 Olares ID 登录。点击左上角头像，启用 VPN 连接。

   ![在电脑上启用 LarePass VPN](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

2. 打开 **设置** > **应用**，选择 QEMU Server 实例，再打开**导出端口**。记下 SSH 对应的**导出端口**值，例如本例中的 `48113`。

   ![QEMU Server 的 SSH 导出端口](/images/manual/use-cases/qemu-export-ports.png#bordered){width=70%}

3. 使用 Olares ID 对应的域名作为主机地址。例如，Olares ID 为 `laresprime@olares.com` 时，使用 `laresprime.olares.com`。

4. 打开终端或 PowerShell，使用 `-p` 指定导出端口。例如，域名为 `laresprime.olares.com`、导出端口为 `48113` 时，运行：

   ```bash
   ssh -p 48113 ubuntu@laresprime.olares.com
   ```

5. 按提示输入 Linux 密码。

使用 Xshell 等 SSH 客户端时，主机填写 Olares 域名，端口填写 SSH 导出端口，再输入相同的 Linux 用户名和密码。

## 使用浏览器控制台

浏览器控制台适合临时操作，也可以在 SSH 无法连接时用来查看启动信息。

1. 从启动台打开 QEMU Server，然后点击 **Open console**。

   ![打开 QEMU Server 浏览器控制台](/images/manual/use-cases/qemu-browser-console.png#bordered)

2. 在登录提示处输入 `ubuntu`，再输入安装时设置的密码。使用其他发行版时，填写[获取用户名和密码](#获取用户名和密码)中对应的用户名。

登录后，就可以在浏览器中运行 Linux 命令。

![QEMU Server 浏览器控制台中的 Linux 命令行](/images/manual/use-cases/qemu-console-terminal.png#bordered)

## 在电脑和 Linux 之间传输文件

使用 SFTP 客户端，可以向服务器上传文件，也可以将服务器中的文件下载到电脑。以下以 FileZilla Client 为例。

1. 在电脑上安装 [FileZilla Client](https://filezilla-project.org/download.php?type=client)。使用与 SSH 相同的网络连接方式，远程访问时保持 LarePass VPN 开启。
2. 打开 FileZilla，进入 **File（文件）** > **Site Manager（站点管理器）**，点击 **New site（新站点）**，将站点命名为 `Ubuntu on Olares`。
3. 填写连接信息：

   | 字段 | 填写内容 |
   | :--- | :--- |
   | **Protocol（协议）** | 选择 SFTP。 |
   | **Host（主机）** | 局域网连接填写实例的 Overlay IP。VPN 连接填写 Olares 域名，例如 `laresprime.olares.com`。 |
   | **Port（端口）** | 局域网连接填写 `22`。VPN 连接填写 **设置** > **应用** > **QEMU Server** > **导出端口** 中的 SSH 导出端口。 |
   | **Logon Type（登录类型）** | 选择 **Ask for password（询问密码）**。 |
   | **User（用户）** | 填写 `ubuntu`，或所选发行版对应的用户名。 |

4. 点击 **Connect（连接）**，按提示输入 Linux 密码。连接后，左侧显示电脑上的文件，右侧显示服务器上的文件。
5. 上传文件时，在左侧打开文件所在的文件夹，在右侧打开 `/home/ubuntu`，再将文件从左侧拖到右侧。其他发行版请使用对应用户的主目录，例如 `/home/debian`。
6. 下载文件时，先在左侧选择保存位置，再将右侧的文件拖到左侧。等待传输完成后，即可在目标文件夹中找到文件。

上传到 `/home/ubuntu` 的文件，也可以通过 SSH 或浏览器控制台在同一路径下访问。Olares 文件管理器中的文件夹不会自动挂载到虚拟机内。

## 通过克隆创建另一台服务器

克隆会创建全新的 Linux 系统，拥有独立的磁盘、配置和密码，不会复制现有服务器的文件或已安装软件。你可以用它尝试其他发行版，或将开发环境与日常服务分开。

以下以保留 Ubuntu、添加一台 Debian 服务器为例：

1. 打开应用市场，进入**我的 Olares**。
2. 找到 QEMU Server，打开**打开**按钮旁的下拉菜单，选择**克隆**。
3. 填写**新应用名称**，例如 `Debian Services`，以及**桌面快捷方式名称**，然后点击**确认**。
4. 在**配置环境变量**中，将 **DISTRIBUTION** 设为 Debian 13。设置新服务器的 CPU、内存、磁盘容量和密码，然后点击**确认**。
5. 从启动台打开新实例，等待初始化完成。
6. 按照[使用 SSH 连接](#使用-ssh-连接)中的步骤，以 `debian` 用户和新密码登录。局域网连接需要为这个实例单独启用 Overlay 网关，并使用它的 IP。VPN 连接则使用它自己的 SSH 导出端口。

原有 Ubuntu 服务器的文件和设置会保留。请确保设备有足够的可用资源同时运行两台服务器。

:::tip 为每个实例选择发行版
修改 **DISTRIBUTION** 不会替换已有系统盘中的 Linux。要使用其他发行版或版本，请创建新实例。
:::

## 调整 CPU、内存和磁盘容量

你可以根据工作负载调整资源。应用修改会重启服务器并断开 SSH 连接，请先保存正在进行的工作，并备份重要数据。

1. 打开 **设置** > **应用**，选择需要调整的实例。
2. 在**环境变量**中点击**管理环境变量**。
3. 修改 **CPU_CORES**、**MEMORY_GB** 或 **DISK_SIZE_GB**，可用范围见[安装应用](#安装应用)。

   ![修改 QEMU Server 资源配置](/images/manual/use-cases/qemu-resource-settings.png#bordered)

4. 确认修改后，点击**应用**。
5. 从启动台打开实例，等待 Linux 就绪后重新连接。

增大 **DISK_SIZE_GB** 后，重启过程会扩展虚拟磁盘及支持自动扩容的根文件系统。要查看 Linux 中的可用空间，运行：

```bash
df -h /
```

:::warning 磁盘容量不能缩小
将 **DISK_SIZE_GB** 设为更小的值不会改变现有磁盘。需要更小的磁盘时，请先备份文件，再将它们迁移到使用目标容量的新实例中。
:::

## 保存数据

虚拟机的系统盘以 `server.qcow2` 文件的形式保存在 Olares 文件管理器的 **Application** > **Data** > **qemu** > **storage** 中。保存在该系统盘上的软件、文件和服务数据会在应用重启、资源调整后保留。每个克隆实例都有独立的系统盘。

![Olares App Data 中的 QEMU Server 系统盘](/images/manual/use-cases/qemu-app-data.png#bordered)

:::warning 删除实例前先备份
请为重要文件保留独立备份。卸载实例并删除应用数据会移除该实例的系统盘。
:::

## 常见问题

### 可以选择哪些 Linux 发行版？

安装时，可以在 **DISTRIBUTION** 中选择以下发行版：

| 发行版 | 可选版本 |
| :--- | :--- |
| Ubuntu | 24.04 LTS、26.04 LTS |
| Debian | 12、13 |
| Rocky Linux | 9、10 |
| Fedora | 43、44 |

安装后想尝试其他发行版，可以[通过克隆创建另一台服务器](#通过克隆创建另一台服务器)。

### QEMU Server 提示 KVM 不可用怎么办？

QEMU Server 需要运行 Olares 的主机提供硬件虚拟化能力。如果你在自己的 x86 电脑上部署 Olares，请检查 BIOS/UEFI 中是否启用了 Intel VT-x 或 AMD-V/SVM。如果 Olares 本身运行在虚拟机内，还需要由宿主机提供嵌套虚拟化。具体方法可参考 Ubuntu 的 [KVM 支持检查指南](https://ubuntu.com/server/docs/how-to/virtualisation/libvirt/)和[嵌套虚拟化指南](https://ubuntu.com/server/docs/how-to/virtualisation/enable-nested-virtualisation/)。

### 为什么无法连接服务器？

| 问题 | 处理方法 |
| :--- | :--- |
| 无法解析主机名 | 检查主机地址，不要包含 `http://` 或 `https://`。SSH 命令中使用 `-p` 指定端口。 |
| 局域网连接超时 | 检查实例当前的 Overlay IP 和网关状态。Olares 需使用有线以太网，电脑所在网络需允许访问该 IP。访客 Wi-Fi 的隔离设置可能阻止连接。 |
| VPN 连接超时 | 确认 LarePass VPN 已启用，并使用 Olares 域名和目标实例的 SSH 导出端口。 |
| 密码被拒绝 | 使用发行版对应的用户名和当前 Linux 密码。如果在 Linux 内修改过密码，请使用新密码。 |
| 应用运行中，但 SSH 无法连接 | 打开应用页面，等待 Linux 初始化完成。也可以通过浏览器控制台查看启动信息。 |

## 了解更多

- [克隆应用](../manual/olares/market/clone-apps.md)：运行多个应用实例。
- [管理 Overlay 网关](../manual/olares/settings/overlay-gateway.md)：为应用配置局域网访问。
- [管理应用环境变量](../manual/olares/settings/manage-app-env.md)：修改实例配置。
- [在 Olares 设备上运行 Windows 虚拟机](windows.md)：通过浏览器或远程桌面使用 Windows 应用。
- [在 Olares 设备上运行 macOS 虚拟机](macos.md)：安装 macOS，并通过浏览器或 VNC 客户端连接。
