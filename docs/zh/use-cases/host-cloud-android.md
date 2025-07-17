---
outline: [2, 3]
description: 在 Olares 上部署云端 Android 模拟器 redorid，并在 Mac 和 Windows 上通过 adb 和 scacpy 访问云端 Android 主机。
---

# 使用 redroid 搭建云端 Android

[redroid](https://github.com/remote-android/redroid-doc) (Remote Android) 是一款支持 GPU 加速的云端 Android（Android in Cloud）解决方案，完美适配 Olares。redroid 让你轻松在 Olares 上托管高性能 Android 实例，随时随地访问并运行 Android 游戏、应用，甚至进行批量自动化测试。

本教程将带你在 Olares 上完成 redroid 的安装与配置，并从 Windows 和 macOS 上远程连接、操控 Android 实例。

## 目标
通过本教程，你将学习：
- 在 Olares 宿主机上安装并配置 redroid 所需的内核依赖。
- 在 Olares 上安装 redroid 应用并获取对外服务 URL。
- 通过 Windows 和 macOS 上通过 `adb` 和 `scrcpy` 连接并操控 Android 实例。
- 在 Android 实例上安装 APK 应用。

## 开始之前
在开始之前，请确保满足以下条件：
- Olares 已安装并运行。

   ::: tip 配置要求
   - redroid 仅支持在 Linux 上运行，请确保你的 Olares 实例部署在 Linux 系统上。
   - redroid 运行时会消耗较高系统资源。为获得更佳性能，建议使用至少 8 核 CPU 和 16GB 内存的主机运行 Olares。
   :::

- 在连接设备上安装 LarePass 客户端并启用[专用网络](../manual/larepass/private-network.md)。可在 [LarePass 官网](https://olares.cn/larepass)下载正确的版本。
 
 ## 安装内核依赖模块

 在 Linux 系统运行安卓模拟服务需要安装特定内核依赖模块，详见 [redroid 项目文档](https://github.com/remote-android/redroid-doc/blob/master/deploy/README.md)。

以 Ubuntu 系统为例，可在终端执行以下命令安装所需内核模块：

```bash
sudo apt install linux-modules-extra-`uname -r`
sudo modprobe binder_linux devices="binder,hwbinder,vndbinder"
# 以下命令可能会报错，高内核版本可忽略
sudo modprobe ashmem_linux
```

## 在 Olares 上安装 redroid

redroid 在 Olares 上以无界面的服务后端运行。要安装 redroid：

1. 打开 Olares 应用市场，在“系统工具”分类下找到 redroid，点击**获取**。安装成功后，redroid 会自动运行。
2. 获取 redroid 对外服务的地址：
   
   a. 从 Olares 桌面进入**设置** > **应用** > **redroid**：
    
   b. 在**端点设置**里获取 redroid 应用的基础域名: `beb583c3.<olares_id>.olares.cn`。

   c. 将 redroid 对外服务端口 `46878` 附在基础域名后。
   
   这样，我们就得到了 redroid 的对外服务网址，如 `beb583c3.olares01.olares.cn:46878`。

## 连接 redroid 服务

要访问 Olares 托管的 Android 实例，我们需要用安卓调试程序 `adb` 连接 redroid 服务，然后用 `scrcpy` 进行视频和音频渲染。

<tabs>
<template #Windows>

Windows 版本的 `scrcpy` 集成了 `adb` 工具，不用另行安装。

1. 从[项目页面](https://github.com/Genymobile/scrcpy/blob/master/doc/windows.md)下载 `scrcpy`，并解压至指定目录。

   ::: tip adb 版本冲突
   如果你本地已安装了其他版本的 `adb`，可能会出现 `adb server` 版本冲突的问题。此时可以卸载先前安装的版本，或将其替换为 `scrcpy` 使用的版本。
   :::

2. 打开 PowerShell，进入 `scrcpy` 目录：

   ```powershell
   # 替换为实际安装路径
   cd .\scrcpy-win64-v3.1
   ```
3. 使用 `adb` 通过前面获取的 URL 连接至 redroid 服务：

   ```powershell
   # 请将 <olares_id> 替换为你自己的 Olares ID
   .\adb.exe connect beb583c3.<olares_id>.olares.cn:46878
   ```
   
   连接成功会看到示例中的消息提示：
   
   ```powershell
   # 示例输出
   already connected to beb583c3.<olares_id>.olares.cn:46878
   ```

4. 用 `scrcpy` 渲染界面和音频：

   ```powershell
   .\scrcpy.exe -s beb583c3.harvey063.olares.cn:46878 --audio-codec=aac --audio-encoder=OMX.google.aac.encoder
   ```

   执行成功后，命令行会输出连接设备信息，同时在桌面弹出安卓屏幕。

   ![渲染成功](/images/manual/tutorials/render-android-windows.png#bordered)

</template>
<template #macOS>

macOS 版本 `scrcpy` 没有集成 `adb`，需要你单独安装。推荐使用 Homebrew 方式安装。

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
   看到对应的版本信息即表示安装成功。

   :::tip 应用阻止警告
   如果程序被 macOS 的安全设置拦截，可以打开 **系统设置** > **隐私与安全性** > **安全性**页面，找到对应的阻止项并点击**仍要打开**。再次运行时，按提示输入密码即可正常运行。
   :::
   
4. 使用 `adb` 连接至之前获得的 redroid 服务地址:

   ```bash
   # 请将 <olares_id> 替换为你自己的 Olares ID
   adb connect beb583c3.<olares_id>.olares.cn:46878
   ```

   看到示例输出即代表服务连接成功：
   
   ```bash
   ```bash
   # 示例输出
   already connected to beb583c3.<olares_id>.olares.cn:46878
   ```

4. 用 `scrcpy` 渲染界面和音频：
   
   ```bash
   scrcpy -s beb583c3.<olares_id>.olares.cn:46878 --audio-codec=aac --audio-encoder=OMX.google.aac.encoder
   ```
   执行成功后，命令行会输出连接设备信息，同时在桌面弹出安卓屏幕。

   ![渲染成功](/images/manual/tutorials/render-android-mac.png#bordered)

</template>
</tabs>


## 安装 APK 应用
    
连接成功后，你可以尝试用 `adb` 为远程 Android 实例安装一个三方应用。

<tabs>
<template #Windows>
1. 查看当前连接设备详细信息：

   ```powershell
   .\adb.exe devices -l
   ```

   从输出结果中获取目标设备的 `transport_id` 为 `4`：

   ```powershell
   # 示例输出
   List of devices attached
   beb583c3.olares02.olares.cn:46878 device product:ziyi model:23031PN0DC device:ziyi transport_id:4
   ```

2. 在指定设备上安装 apk 应用，需通过 `-t` 参数指定 `transport_id`：
   
   ```powershell
   .\adb.exe -t 4 install C:\Users\YourName\Downloads\your_app.apk
   ```

   看到如下输出则安装成功：
   
   ```powershell
   # 示例输出
   Performing Streamed Install
   Success
   ```
</template>
<template #macOS>
1. 查看当前连接设备详细信息：

   ```bash
   adb devices -l
   ```
   从输出结果中获取目标设备的 `transport_id` 为 `4`：

   ```bash
   # 示例输出
   List of devices attached
   beb583c3.olares02.olares.cn:46878 device product:ziyi model:23031PN0DC device:ziyi transport_id:4
   ```
   

2. 在指定设备上安装 apk 应用，需通过 `-t` 参数指定 `transport_id`:
   
   ```bash
    adb -t 4 install ~/Downloads/your_app.apk
    ```
   
   看到如下输出则安装成功：

   ```bash
   # 示例输出
   Performing Streamed Install
   Success
   ```
</template>
</tabs>

安装成功后，重新执行 `scrcpy` 命令连接 Android。上划屏幕，就能看到刚刚安装的应用了。
   

## 常见 `adb` 命令参考
:::tip 注意
以下提供命令适用于 macOS 及 Linux 系统。Windows 系统用户请在 `adb` 命令后加上 `.exe`。
:::

```bash
# 启动adb
adb start-server
# 连接设备
adb connect url:port
# 查看当前已连接设备列表
adb devices 
# 断开链接
adb disconnect url:port
# 手动安装应用
adb -t 3(transport_id，设备列表可看到) install xx.apk
# 查看日志
adb logcat
# 导出日志
adb logcat -v time > log.txt
# 将文件从本地推送到设备
adb push <本地路径> <设备路径>
# 将文件从设备拉取到本地
adb pull <设备路径> <本地路径>
# 列出设备目录内容
adb shell ls <路径>
# 查看设备文件内容
adb shell cat <文件路径>
# 重启
adb shell
# 关机
adb shell reboot -p
```






