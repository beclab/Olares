---
outline: [2, 3]
description: 在运行 Ubuntu 的 Olares One 上配置网络唤醒，并通过同一局域网中的手机、Linux、macOS 或 Windows 设备唤醒主机。
head:
  - - meta
    - name: keywords
      content: Olares One, Ubuntu, 网络唤醒, Wake-on-LAN, WOL, Magic Packet
---

# 在 Ubuntu 上为 Olares One 设置网络唤醒

网络唤醒（Wake-on-LAN，WOL）可以通过同一局域网中的其他设备向 Olares One 发送魔术包，将设备从休眠或关机状态唤醒。本文介绍如何准备 Olares One，以及如何通过手机、Linux、macOS 或 Windows 发送唤醒包。

:::info 仅限局域网
本文仅介绍如何从同一局域网唤醒 Olares One。通过互联网唤醒设备需要额外配置路由器和安全策略，不在本文范围内。
:::

## 开始前准备

- Olares One 已安装 Ubuntu。如未安装，请参阅[在 Olares One 上安装 Ubuntu Server](install-ubuntu-server.md)或[在 Olares One 上安装 Ubuntu Desktop](install-ubuntu-desktop.md)。
- Olares One 已通过网线连接到路由器。网络唤醒不支持通过设备的 Wi-Fi 连接使用。
- 可以使用具有 `sudo` 权限的账号访问 Ubuntu 终端。
- 已准备一台与 Olares One 位于同一局域网的手机或电脑。

## 在 Olares One 上配置网络唤醒

### 检查 EC 固件

网络唤醒要求 EC 固件版本不低于 1.01。建议升级到最新版本以获得更好的兼容性。检查或更新版本的方法，请参阅[管理 BIOS 和 EC](update-firmware.md)。

### 查找网卡名称和 MAC 地址

1. 确保 Olares One 已通过网线连接到路由器。
2. 打开 Ubuntu 终端。
3. 运行以下命令：

   ```bash
   ip address
   ```

4. 找到有线网卡。网卡名称通常以 `en` 开头，例如 `enp129s0`。
5. 记录有线网卡的以下信息：

   - **网卡名称**：显示在网卡信息开头的名称。
   - **MAC 地址**：`link/ether` 后面的值。
   - **IPv4 地址**：`inet` 后面的值，不包含子网后缀。例如，`192.168.0.92/23` 应记录为 `192.168.0.92`。

### 检查网络唤醒状态

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
部分 Ubuntu 网络配置会在启动时重置网络唤醒设置。如果 Olares One 重启后无法响应魔术包，请重新检查状态，并在休眠或关机前再次启用网络唤醒。
:::

### 让 Olares One 休眠或关机

保持 Olares One 与电源和网线连接，然后选择以下一种方式。

- 让 Olares One 进入休眠：

  ```bash
  sudo systemctl suspend
  ```

- 关闭 Olares One：

  ```bash
  sudo shutdown -h now
  ```

首次使用网络唤醒时，建议先使用休眠模式。部分网络环境或电源设置可能不支持从完全关机状态唤醒。

## 唤醒 Olares One

以下方式均需要使用之前记录的 MAC 地址。发送唤醒包的设备必须与 Olares One 位于同一局域网。

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

4. 保存设备，然后点击该设备发送魔术包。

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

### 通过 Windows 唤醒

1. 打开 PowerShell。
2. 将以下脚本中的 `<MAC 地址>` 替换为 Olares One 有线网卡的 MAC 地址，然后运行脚本：

   ```powershell
   $mac = "<MAC 地址>"
   $macBytes = [byte[]]($mac -split '[:-]' | ForEach-Object {
     [Convert]::ToByte($_, 16)
   })
   $packet = [byte[]](,0xFF * 6 + ($macBytes * 16))
   $udp = [System.Net.Sockets.UdpClient]::new()
   $udp.EnableBroadcast = $true
   [void]$udp.Send($packet, $packet.Length, "255.255.255.255", 9)
   $udp.Close()
   ```

3. 等待 Olares One 启动。

## 故障排查

如果 Olares One 未被唤醒，请检查以下内容：

- Olares One 仍与电源和网线保持连接。
- 发送唤醒包的设备与 Olares One 位于同一局域网。
- 输入的是有线网卡的 MAC 地址，而不是 Wi-Fi 网卡地址。
- EC 固件版本不低于 1.01。
- 在休眠或关机前，`ethtool` 显示 `Wake-on: g`。
- 如果无法从关机状态唤醒，请先让 Olares One 进入休眠状态，然后重试。
- 检查路由器是否将 Wi-Fi 设备与有线设备隔离。启用客户端隔离时，魔术包可能无法到达 Olares One。
- 在 Linux 或 macOS 上，尝试将唤醒包发送到子网广播地址：

  ```bash
  wakeonlan -i <广播地址> <MAC 地址>
  ```
