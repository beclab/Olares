---
outline: [2, 3]
description: 了解如何下载和使用 Olares 提供的一体化驱动包来安装所有必要的 Windows 驱动，包括 NVIDIA 显卡驱动。
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, Windows 驱动, NVIDIA 显卡驱动, GPU
---

# 在 Windows 上安装驱动 <Badge type="tip" text="15 min"/>

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/install-nvidia-driver.md)。
:::

为了确保在 Olares One 上运行 Windows 时获得最佳性能和系统稳定性，Olares 提供了一个经过测试的一体化驱动包。该包包含设备的所有必要驱动，如音频、网络和 NVIDIA 显卡驱动。

本指南将引导你完成下载和安装一体化驱动包的过程。

## 学习目标

- 使用驱动包下载并安装所有系统驱动。

## 开始之前

:::info 关于 NVIDIA 显卡驱动更新
这个一体化包包含一个专门为 Olares One 优化的稳定、经过充分测试的 NVIDIA 显卡驱动。为了确保系统稳定性，避免直接从 NVIDIA 官方网站下载和安装独立显卡驱动，因为这可能会引入兼容性问题。
:::

- **管理员权限**：你需要管理员权限来安装系统驱动。
- **互联网连接**：驱动文件很大（约 3.5 GB），因此需要稳定的连接。
- **关闭应用**：在开始安装之前保存你的工作并关闭任何图形密集型程序，如游戏或照片编辑器。更新期间屏幕可能会暂时变黑或闪烁，这可能导致打开的应用崩溃或丢失未保存的数据。
- **Windows 版本**：为了获得最佳的稳定性和驱动兼容性，建议使用 Windows 11 版本 24H2。

## 步骤 1：下载并解压驱动包

1. 下载[驱动包](https://cdn.olares.com/common/OlaresOne_driver_251125.zip)。

    :::tip 浏览器安全警告
    由于下载链接使用标准的 HTTP 连接而非 HTTPS，你的 Web 浏览器可能会将其标记或阻止为不安全下载。如果发生这种情况，请在浏览器的下载管理器中选择 **保留** 或 **允许** 以继续下载。
    :::

    :::info 下载时间
    驱动包很大。根据你的互联网连接速度，下载可能需要一段时间才能完成。
    :::

2. 在计算机上找到下载的 `.zip` 文件。
3. 右键单击该文件并选择 **全部解压**。
4. 打开解压后的文件夹，确认文件存在，包括 `driver_install` 应用。

    ```text
    Extracted_Folder/
    ├── driver/
    ├── uwp/
    ├── driver_install
    └── uwp_install
    ```

## 步骤 2：安装驱动

1. 在解压后的文件夹中，找到文件 `driver_install`，然后双击它。
2. 系统会自动打开命令提示符窗口并开始安装驱动。此过程完全自动化。

    ![安装驱动](/images/one/driver-install.png#bordered)

    :::tip 屏幕闪烁
    安装期间屏幕可能会变黑或闪烁几秒钟。这是系统切换到新驱动时的正常行为。
    :::

3. 驱动成功安装后，设备会自动重启。
4. 系统重启并重新登录 Windows 后，会出现一个命令提示符窗口。按照屏幕上的指示并按任意键直到窗口关闭。这表示安装已完全完成。

    ![系统重启](/images/one/system-restart-windows.png#bordered)
