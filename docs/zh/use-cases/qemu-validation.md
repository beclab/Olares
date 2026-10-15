---
outline: [2, 3]
description: 验证 QEMU Server 的发行版安装、SSH 连接、资源调整、数据持久化，以及通过 Clone 创建的独立 Linux 实例。
head:
  - - meta
    - name: keywords
      content: "Olares, QEMU Server, Linux, 验收测试, SSH, Clone, 数据持久化"
app_version: "0.0.27"
doc_version: "1.0"
doc_updated: "2026-10-16"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/qemu-validation.md)为准。
:::

# 验证 QEMU Server 安装

安装 QEMU Server 或修改配置后，可以按照以下用例验收。用例覆盖[使用 QEMU Server 运行 Linux 服务器](qemu.md)中的用户流程。提供测试步骤并不代表所有平台和网络路径均已验证通过。

## 准备验证环境

使用没有关键业务的实例，并保留独立备份。资源测试会重启虚拟机并中断连接。Clone 测试需要宿主机有足够资源同时运行第二个实例。

记录应用版本、Olares 版本、宿主机架构、发行版、实例名称和测试日期。连接地址应妥善保存，不要将密码写入测试报告。

资源测试初始配置为 1 vCPU、2 GiB 内存和 20 GiB 磁盘。这些是测试值，不是安装默认值。依次对 Ubuntu 24.04 LTS、Debian 13、Rocky Linux 10 和 Fedora 44 执行安装及资源测试。

## 安装与控制台

| 编号 | 操作步骤 | 预期结果 |
| --- | --- | --- |
| QEMU-01 | 从 Market 安装各发行版，并自行填写密码。打开应用，等待初始化。登录后运行 `cat /etc/os-release`。 | 页面进入就绪状态，Linux 显示所选发行版及版本。 |
| QEMU-02 | 打开浏览器控制台，使用对应发行版账户和安装时设置的密码登录。 | 可以查看启动输出并使用 Shell，账户支持 `sudo`。 |
| QEMU-03 | 镜像正在下载时，重启应用一次，再次打开应用页面。 | 下载服务器支持续传时，会继续使用保留的数据；最终完成校验与初始化。 |

## SSH 连接

| 编号 | 操作步骤 | 预期结果 |
| --- | --- | --- |
| QEMU-04 | 在同一局域网内，为实例启用 Overlay 网关，复制其 IP，通过端口 22 连接 SSH，并运行 `cat /etc/os-release`。 | 密码认证成功，连接到目标 Linux 实例。 |
| QEMU-05 | 将电脑切换到其他网络，开启 LarePass VPN，获取 Olares 设备地址和 Settings 中该实例的 SSH 外部端口。将 `ssh -p PORT USER@DEVICE_ADDRESS` 中的大写占位内容替换为实际值后连接。 | 在不依赖同一局域网的情况下认证成功，并显示目标发行版。 |
| QEMU-06 | 分别使用 macOS 或 Linux OpenSSH，以及 Windows OpenSSH 或 Xshell 重复登录。 | 各客户端使用正确的主机、端口、用户名和密码后，均连接到同一实例。 |

QEMU-05 必须在不同网络中验证。在原局域网内连接成功，不能作为远程连接通过的依据。应记录使用的外部网络，并确认连接经过预期的 VPN 路径。

## 资源调整与数据持久化

调整资源前，在 Linux 内记录基线：

```bash
cat /etc/os-release
nproc
free -h
lsblk -b -o NAME,SIZE,TYPE,MOUNTPOINTS
df -h /
findmnt -n -o FSTYPE /
cat /etc/machine-id
cat /proc/sys/kernel/random/boot_id
printf 'qemu-persistence-check\n' > ~/qemu-persistence-check.txt
```

| 编号 | 操作步骤 | 预期结果 |
| --- | --- | --- |
| QEMU-07 | 将实例环境变量中的 CPU 改为 2、内存改为 3 GiB，应用配置，初始化完成后重新连接。 | `nproc` 返回 2，内存增加到接近配置容量，启动 ID 发生变化。 |
| QEMU-08 | 将磁盘从 20 GiB 增加到 24 GiB，应用配置，重新连接后检查 `lsblk` 和 `df -h /`。 | 虚拟磁盘为 24 GiB，根文件系统容量增大且小于整盘容量，原有文件仍可读取。 |
| QEMU-09 | 将 CPU 调回 1、内存调回 2 GiB，并请求将磁盘设为 16 GiB，应用后重新连接。 | CPU 和内存减小，磁盘仍为 24 GiB，页面显示不支持缩盘的警告。 |
| QEMU-10 | 重启应用，再次读取标记文件、机器 ID、启动 ID 和磁盘容量。 | 标记内容仍为 `qemu-persistence-check`，机器 ID 不变，启动 ID 改变，磁盘仍为 24 GiB。 |

每次重启后，运行 `cat ~/qemu-persistence-check.txt` 检查标记文件。记录根文件系统类型：已验证镜像中，Ubuntu 和 Debian 使用 ext4，Rocky Linux 使用 XFS，Fedora 使用 Btrfs。应在 Linux 内检查实际扩容结果，不能仅依据应用页面显示的请求容量判断。

## Clone 与实例隔离

1. 保留前述 Ubuntu 实例及标记文件，记录其机器 ID 和连接信息。
2. 在 Market 的 **My Olares** 中，打开实例菜单并选择 **Clone**，创建名为 `Debian Services` 的 Debian 13 实例，配置独立资源及密码。
3. 打开 Clone，等待就绪，使用 `debian` 登录，检查 `/etc/os-release` 和 `/etc/machine-id`。
4. 依次为两个实例开启 Overlay 网关，每次开启后重新查看网关列表。
5. 使用两个实例各自的 Overlay IP 和端口 22 分别连接，再通过 VPN 使用各自的 SSH 外部端口连接。
6. 在 Clone 内创建一个唯一命名的文件，只重启 Clone，然后再次检查两个系统。

| 编号 | 预期结果 |
| --- | --- |
| QEMU-11 | 原实例仍运行 Ubuntu，Clone 运行 Debian，机器 ID 不同。Clone 为全新安装，不包含从原系统盘复制的文件。 |
| QEMU-12 | 两个实例始终保留在 Overlay 网关列表中，IP 不同，均可通过端口 22 连接。开启第二个实例不会导致第一个从列表消失。 |
| QEMU-13 | VPN 连接使用相同的 Olares 设备地址和不同的 SSH 外部端口，各会话连接到正确发行版。 |
| QEMU-14 | 文件保留在各自实例中。重启 Clone 或调整其资源，不会修改原实例的发行版、配置或文件。 |

## 已记录的验证范围

QEMU Server 0.0.27 已记录的回归环境为 Olares 1.12.7、amd64。以下结果仅描述该轮验证，请按上述用例验收自己的环境。

| 发行版 | 文件系统 | 扩容前后根文件系统容量 | 已记录结果 |
| --- | --- | --- | --- |
| Ubuntu 24.04 LTS | ext4 | 18.33 → 22.21 GiB | 安装、内部网络 SSH 密码认证、控制台传输、资源调整及持久化通过。 |
| Debian 13 | ext4 | 19.49 → 23.42 GiB | 上述检查通过，下载中断恢复也通过。 |
| Rocky Linux 10 | XFS | 18.38 → 22.38 GiB | 安装、内部网络 SSH 密码认证、控制台传输、资源调整及持久化通过。 |
| Fedora 44 | Btrfs | 19.90 → 23.90 GiB | 上述检查通过。首次启动出现可恢复的主机名警告，后续已确认主机名正确，初始化无致命错误。 |

自动化控制台检查验证了传输和握手，没有完成完整的浏览器交互登录。SSH 密码认证通过应用内部网络验证。该轮尚未完成外部局域网及 VPN SSH、跨平台 SSH 客户端，以及双实例 Clone/Overlay 行为验证。因此，QEMU-04–06 和 QEMU-11–14 在此版本上应标记为待执行，直到实际完成对应测试。原生 ARM64 和其他列出的发行版版本也需要单独验证。

每个用例应记录“通过”“失败”或“未执行”，并附实际观察结果。失败时保留截图或终端输出，移除凭据和个人地址信息。不能从一个版本安装成功，推断其他架构或发行版版本也已通过。
