---
outline: [2, 3]
description: 学习如何把你的 Olares 设备变成一台游戏主机，直接连接显示器和外设即可畅玩 Steam 游戏。
---

# 在 Olares 本机畅玩 Steam 游戏

Olares 通常作为无头系统使用，但你也可以把显示器、键盘和鼠标直接连接到设备上，在本机畅玩游戏。

不过，连接显示器只是起始一步。由于 Steam Headless 默认针对远程串流做了优化，你需要手动配置音频输出，才能让声音从本地设备播放出来。

## 目标

通过本教程，你将学习：
- 授予应用程序访问主机声卡的权限。
- 修改显示和音频设置以适应本地游戏体验。

## 准备工作
开始前，请确保以下条件已满足：
- Olares 版本为 1.12.3 或更高。
- Olares 运行在配备 NVIDIA GPU 的机器上。
- 拥有一个有效的 Steam 账号以访问你的游戏。
- 已将显示器、键盘和鼠标连接到 Olares 设备。

## 1. 安装并配置 Steam Headless

从 Olares 市场安装 Steam Headless，并在 Steam 客户端内完成初始设置。

### 安装 Steam Headless

请按照以下步骤安装和配置 Steam Headless：

1. 打开应用市场，搜索 “Steam”。
2. 点击**获取**，然后点击**安装**。
   ![安装 Steam Headless](/images/zh/manual/use-cases/steam-install-steam-headless.png#bordered)
3. 系统会弹出提示，要求你配置环境变量。这些变量将作为 Sunshine Web UI 的登录凭据：
    - `SUNSHINE_USER`：设置 Sunshine 访问用户名。
    - `SUNSHINE_PASS`：设置相应的密码。
      :::tip 记住你的登录凭据
      这是你登录 Sunshine 的初始凭据。稍后访问 Sunshine 控制台时会用到它们。
      :::
4. 等待安装完成。

### 完成 Steam 初始设置
登录 Steam 完成设置并访问你的游戏库。

:::tip 安装失败请重试
如果由于网络问题导致 Steam 安装或更新失败，请在 Steam Headless 控制台左上角的菜单，导航至 **Applications** > **Internet** > **Steam** 重新启动安装。
:::

1. 安装完成后，打开 Steam Headless 并点击 **Connect**。
   ![连接到 Steam](/images/zh/manual/use-cases/steam-connect-to-steam.png#bordered)
2. Steam 客户端将自动安装并更新。此过程可能需要几分钟。
   ![安装 Steam](/images/manual/use-cases/steam-install-steam.png#bordered)
   ![更新 Steam](/images/manual/use-cases/steam-update-steam.png#bordered)
3. 安装完成后会出现 Steam 登录界面。使用你的 Steam 账户登录。
   ![登录 Steam](/images/manual/use-cases/steam-sign-in-to-steam.png#bordered)

## 2. 授予声卡权限
如果希望音频通过显示器输出，Steam 容器需要直接访问主机的声卡设备。你需要编辑部署配置，把这些硬件设备“直通”到容器里。
:::warning 谨慎编辑
YAML 语法或缩进错误可能会导致应用程序无法启动。
- <b>验证更改：</b>确认你的缩进与现有行保持一致。
- <b>必要时恢复：</b>如果 YAML 更改后应用无法启动，请检查**修改记录**，删除你新增的行即可恢复之前的配置。
  ![检查修订记录](/images/zh/manual/use-cases/steam-yaml-revision-records.png#bordered)
:::

1. 打开 Olares 控制面板，从侧边栏选择 Steam 项目。
2. 导航至**部署** > **steamheadless**，点击 <span class="material-symbols-outlined">edit_square</span> 打开编辑器。
   ![打开 Yaml 编辑器](/images/zh/manual/use-cases/steam-open-yaml-editor.png#bordered)
   
3. 定位到`volumes`部分，在列表末尾增加以下条目：
   ```yaml
    spec:
      volumes:    
    # 保留现有 volumes 部分，例如 steam-headless-claim0
    # ...
    # 在末尾添加以下条目：
        - name: snd
          hostPath:
            path: /dev/snd
            type: ''
   ```
4. 定位到`volumeMounts`部分，在末尾添加以下挂载点：
   ```yaml
       volumeMounts:
       # 保留现有挂载点，如 steam-headless-claim0
       # ... 
       # 在末尾添加以下条目：
         - name: snd
           mountPath: /dev/snd
   ```
5. 点击 **Confirm** 应用更改。容器将重启，并应用新的权限。

## 3. 配置显示和输入
默认情况下，系统配置为串流模式。你需要修改 X11 配置来使用物理显示器和外设输入。

1. 在 Olares 控制面板，点击正在运行的容器，查看其详情。
   ![点击容器](/images/zh/manual/use-cases/steam-click-pod.png#bordered)

2. 点击 **steam-headless** 旁的终端图标，进入容器 Shell。
   ![Access container's shell](/images/zh/manual/use-cases/steam-access-container-shell.png#bordered)

3. 打开 X 配置文件：
   ```bash
   nano /etc/X11/xorg.conf
   ```
4. 根据以下要求修改配置文件。

   a. 删除所有`Section "InputDevice"`部分，停止使用虚拟输入。

   b. 更新`Display`子部分，使其匹配你的显示器分辨率。例如：
   ```bash
   SubSection     "Display"
   Virtual     1920 1080
   Depth       24
   Modes      "1920x1080R" "1920x1080" "1280x800" "1024x768" "1920x1080" "1600x900" "1440x900"
   ```

   :::details 参考`xorg.conf`配置
   ```bash
   Section "ServerLayout"
   Identifier     "Default Layout"
   Screen      0  "Default Screen"
   EndSection
   
   Section "Monitor"
   Modeline "1920x1080R"  138.50  1920 1968 2000 2080  1080 1083 1088 1111 +hsync -vsync
   Identifier     "Configured Monitor"
   HorizSync       5.0 - 1000.0
   VertRefresh     5.0 - 200.0
   ModeLine       "1920x1080" 148.50 1920 2448 2492 2640 1080 1084 1089 1125 +hsync +vsync
   EndSection
   
   Section "Device"
   Identifier     "Configured Video Device"
   Driver         "nvidia"
   Option         "ModeValidation" "NoMaxPClkCheck, NoEdidMaxPClkCheck, NoMaxSizeCheck, NoHorizSyncCheck, NoVertRefreshCheck, NoVirtualSizeCheck, NoTotalSizeCheck, NoDualLinkDVICheck, NoDisplayPortBandwidthCheck, AllowNon3DVisionModes, AllowNonHDMI3DModes, AllowNonEdidModes, NoEdidHDMI2Check, AllowDpInterlaced"
   Option         "AllowEmptyInitialConfiguration"
   Option         "PrimaryGPU" "yes"
   Option         "AllowExternalGpus" "True"
   BusID          "PCI:1:0:0"
   EndSection
   
   Section "Screen"
   Identifier     "Default Screen"
   Device         "Configured Video Device"
   Monitor        "Configured Monitor"
   DefaultDepth    24
   Option         "ProbeAllGpus" "False"
   Option         "BaseMosaic" "False"
   Option         "AllowEmptyInitialConfiguration" "True"
   Option         "ConnectedMonitor" "DFP"
   SubSection     "Display"
   Virtual     1920 1080
   Depth       24
   Modes      "1920x1080R" "1920x1080" "1280x800" "1024x768" "1920x1080" "1600x900" "1440x900"
   EndSubSection
   EndSection
   
   Section "ServerFlags"
   Option "AutoAddGPU" "false"
   EndSection
   ```
   :::
5. 按`Ctrl+O`保存，然后按`Ctrl+X`退出。
6. 重启显示服务器以应用修改：
   ```bash
   supervisorctl stop xorg
   supervisorctl start xorg
   ```

## 4. 启用音频输出
最后，你需要手动选择正确的音频输出通道。

1. 识别你的声卡。在容器终端运行以下命令：
   ```bash
   aplay -l
   ```
   在输出结果中查找你的显示器名称，例如：
   ```bash
   **** List of PLAYBACK Hardware Devices ****
   card 0: NVidia [HDA NVidia], device 3: HDMI 0 [U2790B]
   Subdevices: 0/1
   Subdevice #0: subdevice #0
   card 0: NVidia [HDA NVidia], device 7: HDMI 1 [HDMI 1]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 0: NVidia [HDA NVidia], device 8: HDMI 2 [HDMI 2]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 1: PCH [HDA Intel PCH], device 3: HDMI 0 [HDMI 0]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 1: PCH [HDA Intel PCH], device 7: HDMI 1 [HDMI 1]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 1: PCH [HDA Intel PCH], device 8: HDMI 2 [HDMI 2]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   card 1: PCH [HDA Intel PCH], device 9: HDMI 3 [HDMI 3]
   Subdevices: 1/1
   Subdevice #0: subdevice #0
   ```

   在此示例中，输出对应的是`card 0`、`device 3`。

2. 加载音频模块。把`0,3`替换为你的具体声卡和设备编号：
   ```bash
   pactl load-module module-alsa-sink device=plughw:0,3 sink_name=nvhdmi
   ```

3. 检查音频通道状态（使用上一步用到的声卡编号）：
   ```bash
   amixer -c 0 scontents
   ```
   示例输出：
   ```bash
   Simple mixer control 'IEC958',0
   Capabilities: pswitch pswitch-joined
   Playback channels: Mono
   Mono: Playback [off]
   Simple mixer control 'IEC958',1
   Capabilities: pswitch pswitch-joined
   Playback channels: Mono
   Mono: Playback [off]
   Simple mixer control 'IEC958',2
   Capabilities: pswitch pswitch-joined
   Playback channels: Mono
   Mono: Playback [off]
   ```
4. 如果输出结果包含`Mono: Playback [off]`，则表示该通道已静音。运行以下命令取消静音：
   ```bash
   amixer -c 0 sset 'IEC958' on
   ```
现在，音频会直接从你的显示器或扬声器播放。你可以开始畅玩游戏了。

## 可选：启用远程串流音频
:::info
有关串流和配对说明的详情，请参见[使用 Steam Headless 串流喜爱的游戏](stream-game.md).
:::
本地游玩与远程串流的音频模式互斥。
- <b>本地游玩</b>需要加载`module-alsa-sink`模块。

- <b>远程串流</b>需要卸载该模块，Sunshine 才能抓取音频。

如果你之后需要把游戏串流到其他设备上，需要先卸载本地音频模块，才能恢复串流声音。

要为串流启用音频：
1. 通过 Olares 控制面板打开容器终端。
2. 卸载本地音频模块：
   ```bash
   pactl unload-module module-alsa-sink
   ```

## 常见问题
### 为什么我没在玩游戏时，显示器上仍然显示 Steam 界面？
Olares 设备通常会显示终端提示符。但当 Steam 运行时，会激活图形界面并接管显示输出。

要让显示器回到标准终端界面，你需要在**应用市场**或**设置**中停止 Steam 应用。