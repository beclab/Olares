---
outline: [2, 3]
description: 在运行 Olares OS、Ubuntu 或 Windows 的 Olares One 上配置网络唤醒，并通过同一局域网中的手机、Linux、macOS 或 Windows 设备唤醒主机。
head:
  - - meta
    - name: keywords
      content: Olares One, Olares OS, Ubuntu, Windows, 网络唤醒, Wake-on-LAN, WOL, Magic Packet
---

# 为 Olares One 设置网络唤醒

网络唤醒（Wake-on-LAN，WOL）可以通过同一局域网中的其他设备向 Olares One 发送魔术包，唤醒设备。本文适用于运行 Olares OS、Ubuntu 或 Windows 的 Olares One，介绍如何配置主机，以及如何通过手机、Linux、macOS 或 Windows 电脑发送唤醒包。

:::info 仅限局域网
本文仅介绍如何从同一局域网唤醒 Olares One。通过互联网唤醒设备需要额外配置路由器和安全策略，不在本文范围内。
:::

## 开始前准备

- Olares One 正在运行 Olares OS、Ubuntu 或 Windows。
- Olares One 已通过网线连接到路由器。网络唤醒不支持通过设备的 Wi-Fi 连接使用。
- 可以使用管理员账号配置 Olares One。
- 已准备一台与 Olares One 位于同一局域网的手机或电脑。
- EC 固件版本不低于 1.01。检查或更新版本的方法，请参阅[管理 BIOS 和 EC](update-firmware.md)。

## 在 Olares One 上配置网络唤醒

根据 Olares One 上运行的系统，选择以下一种配置方法。

### Olares OS 或 Ubuntu

预装 Olares OS 和自行安装 Ubuntu 的设备使用相同的配置方法。以下命令需要在主机终端中以 `root` 或具有 `sudo` 权限的账号执行。

#### 查找网卡名称和地址

1. 确保 Olares One 已通过网线连接到路由器。
2. 打开 Olares One 主机终端。Olares OS 用户可以使用 [Control Hub 中的 Olares 终端](access-terminal-control-hub.md)；Ubuntu 用户可以在设备上打开终端，或通过 SSH 连接。
3. 运行以下命令：

   ```bash
   ip address
   ```

4. 找到有线网卡。网卡名称通常以 `en` 开头，例如 `enp129s0`。
5. 记录有线网卡的以下信息：

   - **网卡名称**：显示在网卡信息开头的名称。
   - **MAC 地址**：`link/ether` 后面的值。
   - **IPv4 地址**：`inet` 后面的值，不包含子网后缀。例如，`192.168.0.92/23` 应记录为 `192.168.0.92`。
   - **子网广播地址**：`inet` 所在行中 `brd` 后面的值。

   例如，以下输出中的 IPv4 地址为 `192.168.0.92`，子网广播地址为 `192.168.1.255`：

   ```text
   inet 192.168.0.92/23 brd 192.168.1.255 scope global dynamic enp129s0
   ```

   以上地址仅为示例，请使用自己设备上显示的值。

#### 检查网络唤醒状态

1. 安装 `ethtool`：

   ```bash
   sudo apt update
   sudo apt install ethtool -y
   ```

2. 检查当前的网络唤醒状态。将 `<网卡名称>` 替换为之前记录的有线网卡名称。

   ```bash
   sudo ethtool <网卡名称> | grep Wake-on
   ```

3. 确认 **Supports Wake-on** 中包含 `g`，并且显示 `Wake-on: g`：

   ```text
   Supports Wake-on: pumbg
   Wake-on: g
   ```

   `g` 表示网卡收到魔术包后可以唤醒设备。

4. 如果 **Supports Wake-on** 包含 `g`，但 **Wake-on** 显示为其他值，请启用魔术包唤醒：

   ```bash
   sudo ethtool --change <网卡名称> wol g
   ```

5. 再次运行状态检查命令，确认显示 `Wake-on: g`。

:::tip 重启后设置被重置
部分网络配置会在启动时重置网络唤醒设置。如果 Olares One 重启后无法响应魔术包，请重新检查状态，并在睡眠或关机前再次启用网络唤醒。
:::

#### 让 Olares One 睡眠或关机

保持 Olares One 与电源和网线连接，然后选择以下一种方式。

- 让 Olares One 进入睡眠：

  ```bash
  sudo systemctl suspend
  ```

- 关闭 Olares One：

  ```bash
  sudo shutdown -h now
  ```

首次使用网络唤醒时，建议先使用睡眠模式。部分网络环境或电源设置可能不支持从完全关机状态唤醒。

### Windows

以下步骤用于将运行 Windows 的 Olares One 从睡眠状态唤醒。

1. 在 Olares One 上打开**设备管理器**，展开**网络适配器**，右键点击有线网卡，选择**属性**。
2. 在**电源管理**选项卡中，勾选**允许此设备唤醒计算机**，点击**确定**。
3. 按 `Win + R`，输入 `ncpa.cpl` 并按回车。双击正在使用的**以太网**连接，然后点击**详细信息**，记录以下信息：
   - **物理地址**：有线网卡的 MAC 地址。
   - **IPv4 地址**和 **IPv4 子网掩码**：用于确定子网广播地址。例如，IP 地址为 `192.168.1.92`、子网掩码为 `255.255.255.0` 时，广播地址为 `192.168.1.255`。

   :::details 计算子网广播地址
   在 PowerShell 中运行以下命令，将前两行的示例值替换为刚才记录的 IPv4 地址和子网掩码：

   ```powershell
   $ipBytes = ([System.Net.IPAddress]::Parse("192.168.1.92")).GetAddressBytes()
   $maskBytes = ([System.Net.IPAddress]::Parse("255.255.255.0")).GetAddressBytes()
   (0..3 | ForEach-Object {
     $ipBytes[$_] -bor ($maskBytes[$_] -bxor 255)
   }) -join '.'
   ```

   记录输出的广播地址，发送唤醒包时使用。
   :::

4. 保持 Olares One 连接电源和网线，选择**开始** > **电源** > **睡眠**。

:::info Windows 唤醒范围
请使用睡眠模式完成此流程。能否从关机状态唤醒取决于 Windows 电源设置和硬件支持，详情请参阅 [Microsoft 的网络唤醒说明](https://learn.microsoft.com/en-us/troubleshoot/windows-client/setup-upgrade-and-drivers/wake-on-lan-feature)。
:::

## 发送魔术包唤醒 Olares One

在同一局域网中的另一台设备上，选择以下一种方式发送唤醒包。使用之前记录的 Olares One 有线网卡地址。

### 通过手机唤醒

以下步骤以 Easy WOL 为例。也可以使用其他支持发送网络唤醒魔术包的应用。

1. 将手机连接到 Olares One 所在的局域网。
2. 安装并打开网络唤醒应用。
3. 添加 Olares One，并输入以下信息：

   - **Device Name**：输入便于识别 Olares One 的名称。
   - **Address**：输入之前记录的 IPv4 地址。
   - **MAC**：输入有线网卡的 MAC 地址。
   - **Port**：输入 `9`。

   ![在手机上配置网络唤醒设备](/images/one/wol-add-device.png#bordered){width=50%}

4. 保存设备。需要唤醒 Olares One 时，点击已保存的设备。应用会立即发送魔术包。

   ![网络唤醒包发送成功提示](/images/one/wol-packet-sent.png#bordered){width=35%}

5. 等待 Olares One 启动。

### 通过 Linux 唤醒

1. 安装 `wakeonlan`：

   ```bash
   sudo apt update
   sudo apt install wakeonlan -y
   ```

2. 发送魔术包。将 `<MAC 地址>` 替换为 Olares One 有线网卡的 MAC 地址。

   ```bash
   wakeonlan <MAC 地址>
   ```

   运行命令后会立即发送魔术包。等待 Olares One 启动。

### 通过 macOS 唤醒

1. 如果尚未安装 Homebrew，请根据 [Homebrew 官网](https://brew.sh/)的说明完成安装。
2. 安装 `wakeonlan`：

   ```bash
   brew install wakeonlan
   ```

3. 发送魔术包。将 `<MAC 地址>` 替换为 Olares One 有线网卡的 MAC 地址。

   ```bash
   wakeonlan <MAC 地址>
   ```

   运行命令后会立即发送魔术包。等待 Olares One 启动。

### 通过 Windows 唤醒

以下以 Magic Packet Utility 为例，也可以使用其他支持 Wake-on-LAN 的工具。

1. 在发送端 Windows 电脑上下载 [Magic Packet Utility](https://cdn.olares.com/common/magic_packet_utility.zip)，解压 `magic_packet_utility.zip`，打开 `MAGPAC.EXE`。
2. 选择 **Magic Packets** > **Power On One Host**。
3. 填写以下两个字段：
   - **IP Broadcast Address**：Olares One 所在子网的广播地址，例如 `192.168.1.255`。将默认值替换为你的实际广播地址。
   - **Destination Ethernet Address**：Olares One 有线网卡的 MAC 地址。

   ![在 Magic Packet Utility 中填写广播地址和 MAC 地址](/images/one/wol-windows-send.jpg#bordered){width=80%}

4. 点击 **Send**，等待 Olares One 唤醒。

:::details 使用 PowerShell（无需下载工具）
在发送端电脑上打开 PowerShell，将 `<MAC 地址>` 和 `<广播地址>` 替换为之前记录的值。MAC 地址使用冒号或连字符分隔，例如 `84:F7:58:3F:72:29`。运行以下脚本发送唤醒包：

```powershell
$mac = "<MAC 地址>"
$broadcast = "<广播地址>"
$macBytes = [byte[]]($mac -split '[:-]' | ForEach-Object {
  [Convert]::ToByte($_, 16)
})
$packet = [byte[]](,0xFF * 6 + ($macBytes * 16))
$udp = [System.Net.Sockets.UdpClient]::new()
$udp.EnableBroadcast = $true
[void]$udp.Send($packet, $packet.Length, $broadcast, 9)
$udp.Close()
```
:::

## 故障排查

如果 Olares One 未被唤醒，请检查以下内容：

- Olares One 仍与电源和网线保持连接。
- 发送唤醒包的设备与 Olares One 位于同一局域网。
- 输入的是有线网卡的 MAC 地址，而不是 Wi-Fi 网卡地址。
- 确认发送工具中的广播地址与 Olares One 所在子网一致。
- EC 固件版本不低于 1.01。
- Olares OS 或 Ubuntu：在睡眠或关机前，确认 `ethtool` 显示 `Wake-on: g`。
- Windows：确认有线网卡已勾选**允许此设备唤醒计算机**，并且设备处于睡眠状态。如果仍无法唤醒，检查网卡**高级**选项卡中的 **Wake on Magic Packet** 是否已启用（如有此选项）。
- 如果无法从关机状态唤醒，请先让 Olares One 进入睡眠状态，然后重试。
- 检查路由器是否将 Wi-Fi 设备与有线设备隔离。启用客户端隔离时，魔术包可能无法到达 Olares One。
- 在 Linux 或 macOS 上，尝试将唤醒包发送到子网广播地址：

  ```bash
  wakeonlan -i <广播地址> <MAC 地址>
  ```
