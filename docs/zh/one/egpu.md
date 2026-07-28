---
outline: [2,3]
description: 了解如何通过连接外接 GPU (eGPU) 来提升 Olares One 的图形性能。
head:
  - - meta
    - name: keywords
      content: eGPU, 外接显卡, Thunderbolt 5, 硬件扩展
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../one/egpu.md)为准。
:::

# 连接外接显卡 (eGPU) <Badge type="tip" text="5 min" />
Olares One 支持连接外接显卡 (eGPU) 来提升游戏、AI 模型训练等场景的性能。

:::danger 连接前请先关机
请勿热插拔外接 GPU。

Olares One 不支持在系统运行时连接或断开外接 GPU。这样做可能导致系统崩溃、数据丢失或硬件损坏。

在连接或断开设备前，务必完全关闭 Olares One。
:::

## 前提条件
**硬件**<br>
- 外接 GPU 必须是 **NVIDIA Turing 架构或更新版本**（GTX 16xx、RTX 20xx、30xx、40xx、50xx 系列及后续），支持 Thunderbolt 5 协议，且兼容 Ubuntu/Linux。
- 确保外接 GPU 已连接独立的电源。

:::warning 兼容性
外接 GPU 必须是 NVIDIA Turing 或更新架构的 GPU，且兼容 Linux (Ubuntu)。不支持的 GPU 将无法被 Olares 识别，需要 GPU 访问的 AI 应用也将无法运行。
:::

## 步骤 1：连接 eGPU

1. 打开 **Settings**，选择 **My hardware** > **Shutdown**。
   ![关闭 Olares One](/images/one/shut-down-olares-one.png#bordered)

2. 使用 LarePass 应用扫描显示的二维码。应用中出现提示时，点击 **Confirm** 关闭 Olares One。
3. Olares One 完全关机后，将外接 GPU 的线缆插入 Olares One 的 USB-C 端口。
4. 按下电源按钮开启 Olares One。

## 步骤 2：验证连接
要验证外接显卡是否被识别：
1. 登录 Olares，打开 Dashboard。
2. 选择 **GPU** 卡片。你应该能看到外接显卡与内置 GPU 一起列出。
   ![验证 eGPU 连接](/images/one/egpu-verify.png#bordered)

## 断开 eGPU
要安全移除 eGPU：

1. 按照上述步骤完全关闭 Olares One。
2. 关闭外接 GPU 的电源。
3. 从 Olares One 上拔出 USB-C 线缆。
4. 按下电源按钮开启 Olares One。
