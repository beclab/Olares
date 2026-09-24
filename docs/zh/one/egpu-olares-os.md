---
outline: [2, 3]
description: 将 NVIDIA eGPU 连接到 Olares One，在 Olares OS 中检查连接，并在需要时安装 Gen1 临时方案。
---

# 在 Olares OS 上设置 eGPU

按照本文步骤连接 eGPU。仅当开机卡住或系统无法识别显卡时，才安装 Gen1 临时方案。

:::danger 改变连接前必须关机
Olares OS 不支持 eGPU 热插拔。连接或断开 eGPU 前，必须关闭 Olares One。
:::

## 连接 eGPU

1. 打开 **Settings** > **My hardware** > **Shutdown**。

   ![关闭 Olares One](/images/one/shut-down-olares-one.png#bordered)

2. 使用 LarePass 扫描二维码，再点击 **Confirm**。
3. 等待 Olares One 完全关机。
4. 准备 eGPU：

   - 如果设备已经安装 GPU，请连接它的电源适配器。
   - 如果使用 eGPU dock 或 eGPU enclosure，请装入桌面版显卡，并接好显卡所需的全部供电线。

5. 给 eGPU 通电。
6. 断开其他高带宽雷电设备。使用认证的雷电 5 线材，将 eGPU 直接连接到 Olares One 的雷电 5（USB-C）接口。
7. 按下 Olares One 的电源键。

## 检查连接状态

1. 登录 Olares，打开 **Dashboard**。
2. 选择 **GPU** 卡片。
3. 检查内置显卡和 eGPU 是否都已显示。

   ![在 Dashboard 中检查 eGPU](/images/one/egpu-verify.png#bordered)

也可以在终端中检查：

```bash
nvidia-smi
```

如果 Dashboard 和 `nvidia-smi` 中都能看到 eGPU，则无需安装临时方案。

如果开机卡住，或 Dashboard 和 `nvidia-smi` 中都看不到 eGPU，请关闭 Olares One，断开 eGPU，再重新开机。重新连接 eGPU 前，先安装下方的临时方案。

## 必要时安装临时方案

临时方案会在 NVIDIA 驱动加载前，将外置显卡的 PCIe 链路设为 Gen1。它只作用于通过雷电连接的 NVIDIA GPU，不会修改内置显卡。

:::info 性能影响
临时方案会降低系统内存与显存之间的传输带宽，因此模型加载、CPU offload、游戏和实时渲染可能变慢。显卡的计算资源不会改变。
:::

:::warning 临时解决方案
这是 Olares 提供的临时解决方案，不是 NVIDIA 官方修复。请只安装本页提供的文件。后续 Olares 版本会内置该方案，届时无需手动配置。
:::

1. 在不连接 eGPU 的情况下启动 Olares One。
2. 将以下三个文件下载到同一目录：

   - <a href="/downloads/one/egpu/fix-egpu-link.sh" download>`fix-egpu-link.sh`</a>
   - <a href="/downloads/one/egpu/egpu-gen1-fix.service" download>`egpu-gen1-fix.service`</a>
   - <a href="/downloads/one/egpu/99-egpu-gen1-fix.rules" download>`99-egpu-gen1-fix.rules`</a>

3. 在下载目录中打开终端，确认系统可以使用 `setpci`：

   ```bash
   command -v setpci
   ```

   命令应返回 `/usr/sbin/setpci` 等路径。如果没有任何输出，请停止操作并联系 Olares 技术支持。

4. 安装文件并启用服务：

   ```bash
   sudo install -m 0755 fix-egpu-link.sh       /usr/local/sbin/fix-egpu-link.sh
   sudo install -m 0644 egpu-gen1-fix.service  /etc/systemd/system/egpu-gen1-fix.service
   sudo install -m 0644 99-egpu-gen1-fix.rules /etc/udev/rules.d/99-egpu-gen1-fix.rules
   sudo systemctl daemon-reload
   sudo udevadm control --reload-rules
   sudo systemctl enable egpu-gen1-fix.service
   ```

5. 检查安装结果：

   ```bash
   sudo systemctl is-enabled egpu-gen1-fix.service
   ```

   命令应返回 `enabled`。

6. 再次按照[连接 eGPU](#连接-egpu)中的步骤操作。

## 检查临时方案

Olares One 启动后，先在 Dashboard 或 `nvidia-smi` 中检查 eGPU 是否显示，再查看临时方案日志：

```bash
sudo tail -n 20 /var/log/egpu-gen1-fix.log
```

在日志中查找类似内容：

```plain
after rescan: 0000:0a:00.0 speed=2.5 GT/s PCIe driver=nvidia
```

`2.5 GT/s PCIe` 表示外置显卡链路正在以 Gen1 运行。不同系统显示的 PCI 地址可能不同。

如果仍然看不到 eGPU，或运行不稳定，请参考[排查 eGPU 问题](./ts-egpu.md)。

## 断开 eGPU

1. 打开 **Settings** > **My hardware** > **Shutdown**。
2. 在 LarePass 中批准关机，并等待 Olares One 完全关闭。
3. 关闭 eGPU 电源。
4. 从 Olares One 断开雷电线。
5. 重新启动 Olares One。

## 卸载临时方案

:::warning 先断开 eGPU
卸载临时方案前，先关闭 Olares One 并断开 eGPU。
:::

1. 在不连接 eGPU 的情况下启动 Olares One。
2. 卸载临时方案：

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

## 相关资源

- [将 eGPU 连接到 Olares One](./egpu.md)
- [排查 eGPU 问题](./ts-egpu.md)
