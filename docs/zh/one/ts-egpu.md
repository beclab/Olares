---
outline: [2, 3]
description: 排查 Olares One eGPU 开机、识别、掉卡和 Windows 驱动问题，并收集诊断信息。
---

# 排查 Olares One eGPU 问题

开机卡住、系统无法识别 eGPU、运行中掉卡，或 Windows 报显卡驱动错误时，使用本文排查。

## 先尝试快速修复

| 平台 | 现象 | 检查方法 |
|---|---|---|
| Olares OS | 开机卡在 Olares logo | 关闭 Olares One，断开 eGPU 后重新启动，再确认 [Gen1 临时方案](./egpu-olares-os.md)已安装并启用。 |
| Olares OS | 系统启动但看不到 eGPU | 检查扩展坞供电、连接顺序和认证的雷电 5 线材。确认扩展坞连接到 Olares One 的雷电 5（USB-C）接口。 |
| Olares OS | eGPU 在使用过程中掉卡 | 确认 Gen1 临时方案已生效。检查扩展坞供电是否充足，并确认显卡供电线已牢固连接。 |
| Windows | 内置显卡报错 | [清洁安装驱动](./egpu-windows.md#恢复内置显卡)后重启 Windows。 |

Olares OS 冷启动的正确顺序为：扩展坞通电 → 连接雷电线 → Olares One 开机。首次安装 Windows 驱动时，请按 [Windows 设置指南](./egpu-windows.md)中的顺序连接。

## 在 Olares OS 上确认 eGPU 状态

```bash
lspci -nn | grep -i nvidia
nvidia-smi
```

输出中应同时显示内置显卡和 eGPU。要确认 Gen1 临时方案是否生效，请运行下方诊断脚本并查看 `Summary`。链路速率显示 `2.5 GT/s PCIe` 表示 Gen1 已生效。

## 收集 Olares OS 诊断信息

1. 下载 <a href="/downloads/one/egpu/collect-egpu-info.sh" download>`collect-egpu-info.sh`</a>。
2. 在出现问题的那次开机后运行。如果上次开机卡住，可断开 eGPU 正常开机后再运行。系统保留相关日志时，报告也会包含上一次开机的日志。

   ```bash
   chmod +x collect-egpu-info.sh
   sudo ./collect-egpu-info.sh
   ```

3. 打开当前目录生成的 `egpu-report-*.txt`，查看 `Summary`。

   - 如果运行脚本时 eGPU 仍然连接，但显示 `External GPU detected : NO`，请检查扩展坞供电、雷电 5 线材和雷电 5（USB-C）接口。
   - 如果为了从开机卡住状态中恢复而主动断开了 eGPU，显示 `External GPU detected : NO` 属于正常现象。系统保留相关日志时，报告仍可能包含上一次开机的日志。

脚本不会修改系统设置或上传数据。它会读取系统信息，并在当前目录写入一份报告。报告包含系统与驱动版本、雷电和 PCIe 拓扑、链路速率、BAR 分配、临时方案状态及相关日志。

:::warning 分享前检查报告
报告可能包含设备主机名、内核命令行、硬件拓扑和系统日志。公开分享前，请检查文件并删除不希望公开的信息。
:::

## 收集 Windows 诊断信息

准备以下信息：

1. **设备管理器** > **显示适配器**截图，包含警告图标。
2. 已安装的 NVIDIA 驱动版本。
3. **设备属性** > **常规** > **设备状态** 中的完整消息和错误代码。

检查截图，并删除不希望公开的个人信息或设备信息。

同时记录扩展坞与显卡型号、操作系统、连接方式、开机顺序、现象、发生概率和已尝试的操作。对于偶发的启动问题，请使用相同连接顺序完成数次冷启动，并记录每次结果。冷启动是指完全关闭 Olares One 后，再按推荐顺序给扩展坞通电、连接线材并启动 Olares One。

## 相关文档

- [Olares One eGPU 支持概览](./egpu.md)
- [在 Olares OS 上设置 eGPU](./egpu-olares-os.md)
- [在 Windows 上设置 eGPU](./egpu-windows.md)
- [前往 Olares 论坛讨论 eGPU 配置](https://www.olares.com/forum/)
