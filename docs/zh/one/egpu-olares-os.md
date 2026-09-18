---
outline: [2, 3]
description: 在 Olares One 上连接 NVIDIA eGPU、应用临时 Gen1 方案、确认显卡状态并安全断开设备。
---

# 在 Olares OS 上设置 eGPU

本文介绍如何在 Olares One 上连接 NVIDIA eGPU，并应用临时 Gen1 方案。该方案在实测的 Olares OS 环境中提高了驱动初始化的可靠性。

:::danger 连接前先关机
不要在 Olares OS 运行期间连接或断开 eGPU。改变连接前，必须完全关闭 Olares One。
:::

在实测的 Olares OS 环境中，eGPU 的 PCIe 链路运行在 Gen4 时，有时无法完成驱动初始化。本文的临时方案会在 NVIDIA 驱动初始化前，将扩展坞内部 PCIe 链路限制为 Gen1。

## 开始前

本文方案已在以下环境中测试：

- 基于 Ubuntu 24.04 的 Olares OS 1.12.6
- NVIDIA 驱动 595.84
- AOOSTAR EG02
- NVIDIA GeForce RTX 4060 Ti

在该配置中，应用临时方案后也能同时使用内置 RTX 5090M 与外置 RTX 4060 Ti。

临时方案文件没有写死特定的扩展坞或显卡型号。已测试的硬件组合请查看 [eGPU 支持概览](./egpu.md)。

还需要准备：

- Olares OS 的 `sudo` 权限
- 终端访问权限
- 带独立电源的雷电 eGPU 扩展坞
- 认证的雷电 5 线材，优先使用扩展坞附带的线材
- 基于 Turing 或更新架构的 NVIDIA 显卡

雷电协议可以向下兼容。为获得更可靠的使用体验，建议使用雷电 5 硬件，不建议使用低于雷电 5 的扩展坞。

## 性能影响

此临时方案会将 eGPU 链路限制为 Gen1。它不会减少显卡的计算资源，但在系统内存与显存之间传输数据时，性能可能下降。

- 如果模型和工作数据能够长期保留在显存中，影响可能主要出现在模型加载阶段。
- 使用 CPU offload、显存不足、频繁通过 PCIe 传输数据、游戏和实时渲染时，影响可能更明显。

实际影响取决于具体工作负载。

## 安装临时方案

:::warning 系统级临时方案
此方案会修改 eGPU 的 PCIe 链路设置，并在启动期间暂时移除设备后重新扫描。请只使用本页提供的文件。如果你的配置与上述实测配置不同，请先查看 [eGPU 支持概览](./egpu.md)。
:::

1. 将以下文件下载到同一目录：

   - <a href="/downloads/one/egpu/fix-egpu-link.sh" download>`fix-egpu-link.sh`</a>
   - <a href="/downloads/one/egpu/egpu-gen1-fix.service" download>`egpu-gen1-fix.service`</a>
   - <a href="/downloads/one/egpu/99-egpu-gen1-fix.rules" download>`99-egpu-gen1-fix.rules`</a>

2. 在下载目录中打开终端，确认系统可以使用 `setpci`：

   ```bash
   command -v setpci
   ```

   命令应返回类似 `/usr/sbin/setpci` 的路径。临时方案依赖 `pciutils` 软件包提供的 `setpci`。如果没有任何输出，请先停止操作，并向技术支持确认当前 Olares OS 版本支持的安装方式。

3. 安装并启用配置：

   ```bash
   sudo install -m 0755 fix-egpu-link.sh       /usr/local/sbin/fix-egpu-link.sh
   sudo install -m 0644 egpu-gen1-fix.service  /etc/systemd/system/egpu-gen1-fix.service
   sudo install -m 0644 99-egpu-gen1-fix.rules /etc/udev/rules.d/99-egpu-gen1-fix.rules
   sudo systemctl enable egpu-gen1-fix.service
   ```

4. 确认服务已启用：

   ```bash
   sudo systemctl is-enabled egpu-gen1-fix.service
   ```

   命令应返回 `enabled`。

临时方案先根据 PCI vendor 和显示设备类型识别 NVIDIA 显示控制器，然后只处理标记为可移除的设备，从而排除内置显卡。

## 关闭设备并连接 eGPU

1. 打开 **Settings**，选择 **My hardware** > **Shutdown**。

   ![关闭 Olares One](/images/one/shut-down-olares-one.png#bordered)

2. 使用 LarePass 扫描二维码。出现提示时，点击 **Confirm** 关闭 Olares One。
3. 等待 Olares One 完全关机。
4. 将显卡装入扩展坞，并连接扩展坞的独立电源。
5. 给扩展坞通电，然后使用认证的雷电 5 线材，将扩展坞连接到 Olares One 的雷电 5（USB-C）接口。
6. 按下电源键启动 Olares One。

系统会在启动时自动运行临时方案。

## 确认连接状态

请完成以下两项检查。Dashboard 用于确认 Olares 已识别 eGPU，链路速率和日志用于确认临时方案已经生效。

### 在 Dashboard 中确认 eGPU 已识别

1. 登录 Olares，打开 **Dashboard**。
2. 选择 **GPU** 卡片，确认内置显卡和外置显卡均已显示。

   ![在 Dashboard 中确认 eGPU](/images/one/egpu-verify.png#bordered)

### 确认临时方案已生效

1. 确认两张显卡都已出现：

   ```bash
   nvidia-smi
   ```

2. 查看临时方案日志：

   ```bash
   sudo tail -n 20 /var/log/egpu-gen1-fix.log
   ```

`2.5 GT/s PCIe` 表示 Gen1 已生效。成功日志类似：

```plain
after rescan: 0000:0a:00.0 speed=2.5 GT/s PCIe driver=nvidia
```

如果开机卡在 logo，关机并断开 eGPU 后重新开机，再参考[故障排查](./ts-egpu.md)。

## 安全断开 eGPU

1. 打开 **Settings** > **My hardware** > **Shutdown**。
2. 在 LarePass 中确认关机，并等待 Olares One 完全关闭。
3. 关闭扩展坞电源。
4. 从 Olares One 拔下雷电线。
5. 按下电源键重新启动 Olares One。

## 卸载临时方案

:::warning 先断开 eGPU
卸载临时方案前，请关闭 Olares One 并断开 eGPU。卸载后如果在 eGPU 仍连接的情况下启动 Olares One，初始化问题可能再次出现。
:::

1. 按照[安全断开 eGPU](#安全断开-egpu)中的步骤操作，并在不连接 eGPU 的情况下启动 Olares One。

2. 打开终端并卸载临时方案：

   ```bash
   sudo systemctl disable egpu-gen1-fix.service
   sudo rm /usr/local/sbin/fix-egpu-link.sh
   sudo rm /etc/systemd/system/egpu-gen1-fix.service
   sudo rm /etc/udev/rules.d/99-egpu-gen1-fix.rules
   sudo systemctl daemon-reload
   sudo udevadm control --reload-rules
   ```

3. 重启 Olares One：

   ```bash
   sudo reboot
   ```

下次连接 eGPU 时，系统将使用默认 PCIe 链路行为，初始化问题可能再次出现。

## 相关资源

- [Olares One eGPU 支持概览](./egpu.md)
- [排查 eGPU 问题](./ts-egpu.md)
