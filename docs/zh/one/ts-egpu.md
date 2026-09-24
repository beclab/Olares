---
outline: [2, 3]
description: 排查 Olares One eGPU 开机与识别问题，并收集 Olares OS 或 Windows 所需的诊断信息。
---

# 排查 Olares One eGPU 问题

先找到与你遇到的现象一致的章节。如果快速检查无法解决问题，请先收集诊断信息，再寻求帮助。

:::danger Olares OS 不支持热插拔
连接或断开 eGPU 前，必须关闭 Olares One。冷启动时，先给 eGPU 通电并连接到 Olares One，再启动 Olares One。
:::

## Olares OS 开机卡在 logo

1. 关闭 Olares One。
2. 断开 eGPU。
3. 重新启动 Olares One。
4. 检查 Gen1 临时方案是否已启用：

   ```bash
   sudo systemctl is-enabled egpu-gen1-fix.service
   ```

   命令必须返回 `enabled`。如果结果不同，请按照[在 Olares OS 上设置 eGPU](./egpu-olares-os.md)重新安装临时方案。

5. 再次尝试连接前，先[收集 Olares OS 诊断信息](#olares-os)。报告可能保留了上次开机卡住时的日志。

## Olares OS 可以启动，但看不到 eGPU

按顺序检查：

1. 检查 eGPU 是否已通电，设备电源是否连接正常。
2. 如果通过 eGPU dock 或 eGPU enclosure 使用桌面版显卡，请检查电源功率和显卡的全部供电线。
3. 将 eGPU 直接连接到 Olares One 的雷电 5（USB-C）接口。不要经过其他 Dock，并断开连接路径中的其他雷电设备。
4. 使用认证的雷电线材，优先使用外置显卡设备附带的线材。
5. 再次冷启动。先给 eGPU 通电，再连接线材，最后启动 Olares One。
6. 检查操作系统和 NVIDIA 驱动是否识别显卡：

   ```bash
   lspci -nn | grep -i nvidia
   nvidia-smi
   ```

两条命令都应显示外置显卡。如果 `lspci` 能看到 eGPU，但 `nvidia-smi` 看不到，请先收集诊断信息，不要立即更换驱动。

## Olares OS 运行任务时掉卡

1. 检查外置显卡设备的电源。如果通过 eGPU dock 或 eGPU enclosure 使用桌面版显卡，还要检查电源功率和显卡的全部供电线。
2. 断开其他高带宽雷电设备，将 eGPU 直接连接到 Olares One 后重试。
3. 掉卡后立即收集诊断信息。

## Windows 显示显卡错误

1. 在 **设备管理器** > **显示适配器** 中打开报错的设备。
2. 记录 **常规** > **设备状态** 中的完整消息和错误代码。
3. 安装所有可用的 Windows 更新，并更新 NVIDIA 驱动。
4. 重启 Windows，检查预期的全部 GPU 是否都已显示，并且没有警告图标。
5. 如果内置显卡仍然报错，请按照[恢复内置显卡](./egpu-windows.md#恢复内置显卡)中的步骤操作。

如果仍然缺少显卡，请收集下方列出的 Windows 信息。

## 在 Olares 论坛求助

如果问题仍未解决，请前往 [Olares 论坛](https://www.olares.cn/forum/)发帖，并填写：

```plain
系统与版本：
eGPU 设备，或 dock、enclosure 与显卡：
连接方式（包括 Dock 或 Hub）：
开机顺序：
问题现象：
已尝试的操作：
附件：
```

发帖前，请根据所用操作系统收集诊断信息，并附上报告或截图。

### Olares OS

1. 下载 <a href="/downloads/one/egpu/collect-egpu-info.sh" download>`collect-egpu-info.sh`</a>。
2. 在下载目录中打开终端并运行：

   ```bash
   chmod +x collect-egpu-info.sh
   sudo ./collect-egpu-info.sh
   ```

3. 打开生成的 `egpu-report-*.txt`，查看 `Summary`。

   - 如果运行脚本时 eGPU 仍然连接，但报告显示 `External GPU detected : NO`，请再次检查供电、线材、接口和连接路径。
   - 如果开机卡住后主动断开了 eGPU，报告显示 `External GPU detected : NO` 属于正常现象。如果系统保留了上一次开机日志，报告仍会包含相关内容。

脚本只读取系统状态，并在当前目录写入一份报告。它不会修改设置、联网或上传报告。

:::warning 分享前检查报告
报告可能包含主机名、内核命令行、硬件拓扑和系统日志。公开发布前，请删除不希望分享的信息。
:::

### Windows

准备以下内容：

1. **设备管理器** > **显示适配器**截图，包含所有警告图标。
2. NVIDIA App 中显示的驱动版本。
3. 每张报错显卡在设备管理器中的完整错误消息和代码。

分享截图前，请先删除个人信息。

## 相关资源

- [将 eGPU 连接到 Olares One](./egpu.md)
- [在 Olares OS 上设置 eGPU](./egpu-olares-os.md)
- [在 Windows 上设置 eGPU](./egpu-windows.md)
