---
outline: [2, 3]
description: 使用 redroid 在 Olares 上部署云端 Android 模拟器，并通过 macOS 和 Windows 上的 adb 和 scrcpy 访问 Android 主机。
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/host-cloud-android.md)为准。
:::

# 使用 redroid 托管你的云端 Android

[redroid](https://github.com/remote-android/redroid-doc)（Remote Android）是一个支持 GPU 加速的 Android-in-Cloud（AIC）解决方案，与 Olares 无缝集成。你可以轻松地在 Olares 上托管高性能 Android 实例，并随时访问它来运行 Android 游戏、应用，甚至自动化测试。

本教程将指导你在 Olares 上安装 redroid，并从 Windows 和 macOS 访问 Android 实例。

## 学习目标

在本教程中，你将学习如何：
- 在 Olares 主机上安装所需的 Linux 内核模块。
- 在 Olares 上安装 redroid 应用并获取服务 URL。
- 使用 `adb` 和 `scrcpy` 从 Windows 和 macOS 连接并操作 Android 实例。
- 在 Android 实例上安装 APK 应用。

## Prerequisites

在开始之前，请确保满足以下条件：
- Olares 已安装并运行在 Linux 机器上。
    ::: tip 配置要求
    - redroid 仅在 Linux 上受支持。确保你的 Olares 实例运行在 Linux 系统上。
    - redroid 资源消耗较高。为获得最佳性能，建议使用至少配备 8 核 CPU 和 16GB 内存的机器运行 Olares。
    :::

- 在你的设备上启用 [LarePass VPN](../manual/larepass/private-network.md)。

## 安装依赖内核模块

redroid 需要特定的内核模块才能在 Linux 上运行。有关详细信息，请参阅 [redroid 官方文档](https://github.com/remote-android/redroid-doc/blob/master/deploy/README.md)。

例如，在 Ubuntu 上，你可以在终端中运行以下命令来安装所需的内核模块：

```bash
sudo apt install linux-modules-extra-`uname -r`
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"
# 此命令在较新的内核上可能会失败，该错误可以安全忽略。
sudo modprobe ashmem_linux
```

## 在 Olares 上安装 redroid

redroid 在 Olares 上作为无界面后端运行。要安装 redroid：

1. 在 Olares Market 中，在"Utilities"分类下找到 redroid，然后点击 **Get**。安装完成后，redroid 将自动启动。

2. 获取访问 redroid 服务的 URL：

    a. 从 Olares 桌面，导航至 **Settings** > **Application** > **redroid**。

    b. 在 **Entrances** > **Set up endpoint** 中，获取 redroid 的基础域名，例如 `beb583c3.<olares_id>.olares.com`。

    c. 将 redroid 的导出端口（`46878`）附加到基础域名后。

    以下是我们访问 redroid 服务的最终 URL 示例：`beb583c3.olares01.olares.com:46878`。

## 连接到 redroid 服务

要访问 Olares 上的 Android 实例，你需要使用 `adb` 连接到 redroid 服务，并使用 `scrcpy` 渲染 UI。

<tabs>
<template #Windows>

 Windows 版本捆绑了 `adb`，因此你无需单独安装它。

1. 从[项目网站](https://github.com/Genymobile/scrcpy/blob/master/doc/windows.md)下载 Windows 版本的 `scrcpy`，并将其解压到特定文件夹。

    ::: tip adb 版本冲突
    如果安装了另一个版本的 `adb`，可能会导致 `adb` 服务器之间的冲突。卸载旧版本，或将其替换为 `scrcpy` 中捆绑的版本。
    :::

2. 打开 PowerShell，然后导航到 `scrcpy` 目录：

    ```powershell
    # 替换为实际路径
    cd .\scrcpy-win64-v3.1
    ```

3. 使用 `adb` 通过之前获取的 URL 连接到 redroid 服务：

    ```powershell
    .\adb.exe connect beb583c3.<olares_id>.olares.cn:46878
    ```

    如果看到示例输出，则连接成功：

    ```powershell
    # 示例输出
    already connected to beb583c3.<olares_id>.olares.cn:46878
    ```

4. 使用 `scrcpy` 渲染 UI 和音频：

    ```powershell
    .\scrcpy.exe -s beb583c3.<olares_id>.olares.cn:46878 --audio-codec=aac --audio-encoder=OMX.google.aac.encoder
    ```

    成功执行后，命令行会输出设备和渲染信息，并且 Android 屏幕会弹出。

     ![渲染视频](/images/manual/tutorials/render-android-windows.png#bordered)
</template>
<template #macOS>

在 macOS 上，`scrcpy` 默认不包含 `adb`，因此你需要单独安装它们。建议通过 Homebrew 安装。

1. 安装 `scrcpy`：

    ```bash
    brew install scrcpy
    ```

2. 安装 `adb`：

    ```bash
    brew install --cask android-platform-tools
    ```

3. 验证安装：

    ```bash
    scrcpy --version
    adb version
    ```
    如果看到版本号，则安装成功。

    :::tip Gatekeeper 警告
    如果被 macOS 安全设置阻止，请前往 **System Settings** > **Privacy & Security** > **Security**，找到对应项，然后点击 **Allow Anyway**。重新运行命令时，系统会提示你输入密码。
    :::

4. 通过 `adb` 连接到之前获取的 redroid 服务 URL：

    ```bash
    adb connect beb583c3.<olares_id>.olares.cn:46878
    ```

    如果看到示例输出，则连接成功。

    ```bash
    # 示例输出
    already connected to beb583c3.<olares_id>.olares.cn:46878
    ```

5. 使用 `scrcpy` 渲染 UI 和音频：

    ```bash
    scrcpy -s beb583c3.<olares_id>.olares.cn:46878 --audio-codec=aac --audio-encoder=OMX.google.aac.encoder
    ```
    成功执行后，命令行会输出设备信息，并且 Android 屏幕会弹出。

     ![渲染视频](/images/manual/tutorials/render-android-mac.png#bordered)
</template>
</tabs>



## 安装 APK

连接后，你可以使用 `adb` 在 Android 实例上安装第三方 APK 应用。

<tabs>
<template #Windows>

1. 获取所有已连接设备的详细信息：

    ```powershell
    .\adb.exe devices -l
    ```

    获取设备的 `transport_id`，在我们的案例中是 `4`：

    ```powershell
    # 示例输出
    List of devices attached
    beb583c3.<olares_id>.olares.com:46878 device
    product:ziyi model:23031PN0DC device:ziyi
    transport_id:4
    ```

2. 将 APK 安装到指定设备。使用 `-t` 指定传输 ID：

    ```powershell
    .\adb.exe -t 4 install C:\Users\YourName\Downloads\your_app.apk
    ```


    如果看到以下消息，则安装成功：

    ```powershell
    # 预期输出
    Performing Streamed Install
    Success
    ```
</template>
<template #macOS>

1. 获取所有已连接设备的详细信息：

    ```bash
    adb devices -l
    ```

    获取设备的 `transport_id`，在我们的案例中是 `4`：

    ```bash
    # 示例输出
    List of devices attached
    beb583c3.<olares_id>.olares.com:46878 device
    product:ziyi model:23031PN0DC device:ziyi
    transport_id:4
    ```

2. 将 APK 安装到指定设备。使用 `-t` 指定传输 ID：

    ```bash
    adb -t 4 install ~/Downloads/your_app.apk
    ```

    如果看到以下消息，则安装成功：

    ```bash
    # 预期输出
    Performing Streamed Install
    Success
    ```

</template>
</tabs>

安装后，再次运行 `scrcpy` 以渲染 Android 屏幕。向上滑动即可查看已安装的 APK。

## 常用 `adb` 命令

:::tip 注意
以下命令适用于 macOS 和 Linux。在 Windows 上，将 `adb` 替换为 `adb.exe`。
:::

```bash
# 启动 adb 服务器
adb start-server

# 连接到设备
adb connect <url>:<port>

# 列出已连接的设备
adb devices

# 断开设备连接
adb disconnect <url>:<port>

# 通过 transport_id 安装 APK
adb -t 3 install your_app.apk

# 查看实时日志
adb logcat

# 将日志导出到文件
adb logcat -v time > log.txt

# 将文件推送到设备
adb push <local_path> <device_path>

# 从设备拉取文件
adb pull <device_path> <local_path>

# 列出设备上的目录内容
adb shell ls <path>

# 查看文件内容
adb shell cat <file_path>

# 重启设备
adb shell reboot

# 关闭设备
adb shell reboot -p
```