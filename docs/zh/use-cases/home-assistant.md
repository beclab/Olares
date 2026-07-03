---
outline: [2, 3]
description: 了解如何在 Olares 上安装 Home Assistant，将其连接到本地网络，并自动或手动添加智能家居设备。
head:
  - - meta
    - name: keywords
      content: Olares, Home Assistant, 智能家居, 本地发现, 家庭自动化, 覆盖网关, 大华, IP 摄像头, RTSP, HACS
app_version: "1.0.23"
doc_version: "1.0"
doc_updated: "2026-07-03"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/home-assistant.md)。
:::

# 使用 Home Assistant 搭建智能家居中枢

Home Assistant 是一个开源的智能家居自动化平台，可以将你的智能家居设备集中到一起管理。

本教程将介绍如何在 Olares 上安装 Home Assistant，将其连接到本地网络，并自动或手动添加智能家居设备。

## 学习目标

通过本教程，你将学会如何：

- 从 Olares 应用市场安装 Home Assistant 并设置管理员账号。
- 将 Home Assistant 连接到本地网络，发现并添加附近的智能家居设备。
- 手动添加设备并在仪表板上查看实时画面，以大华 IP 摄像头为例。

## 安装 Home Assistant

1. 打开应用市场，搜索 "Home Assistant"。

   ![安装 Home Assistant](/images/manual/use-cases/home-assistant.png#bordered)

2. 点击 **获取**，然后点击 **安装**。等待安装完成。

## 设置账号

设置本地管理员账号，开始使用 Home Assistant 仪表板。

1. 从启动台打开 Home Assistant。
2. 在欢迎页面选择你偏好的语言，然后选择 **创建我的智能家居**。

   ![打开 Home Assistant](/images/manual/use-cases/home-assistant-welcome.png#bordered)

3. 输入所需的个人信息，包括用户名和密码，然后点击 **创建账号**。
4. 按照屏幕上的提示完成剩余设置。
5. 使用新创建的账号登录 Home Assistant 仪表板。

## 发现并添加本地设备

默认情况下，Home Assistant 与本地网络隔离运行，无法自动发现智能家居设备。Olares 可以通过覆盖网关将 Home Assistant 暴露到本地网络，为其分配一个本地 IP 地址。一旦 Home Assistant 接入本地网络，就可以发现 DLNA 媒体渲染器、UPnP 设备、HomeKit 配件以及其他本地服务。

### 启用覆盖网关

1. 打开 Olares 设置，进入 **网络** > **覆盖网关**。
2. 确保系统级别的 **启用覆盖网关** 选项已开启。如果无法启用，请联系超级管理员先开启该选项。
3. 在 **应用** 下找到 **Home Assistant**，然后为其启用覆盖网关。
4. 在确认对话框中点击 **确认**。Olares 会为 Home Assistant 分配一个本地 IP 地址。

   ![为 Home Assistant 启用覆盖网关](/images/manual/use-cases/ha-enable-overlay-gateway.png#bordered){width=90%}

### 配置 Home Assistant 网络适配器

启用覆盖网关后，需要配置 Home Assistant 监听本地网络适配器。

1. 打开 Home Assistant，进入 **设置** > **系统** > **网络** > **网络适配器**。
2. 取消勾选 **自动配置**。
3. 选择 **适配器：eth0** 和 **适配器：net1**。

   ![Home Assistant 网络适配器设置](/images/manual/use-cases/home-assistant-network-adapter.png#bordered)

4. 点击 **保存**。
5. 返回 **设置** > **系统**，然后点击右上角的 <i class="material-symbols-outlined">power_settings_new</i>。
6. 选择 **重启 Home Assistant**，然后确认重启。

   重启后，Home Assistant 将监听本地网络，并可以发现附近的设备。

### 添加发现的设备

1. 在 Home Assistant 中，进入 **设置** > **设备与服务**。

   **已发现** 面板会显示在本地网络上找到的设备，例如：
   - **DLNA 数字媒体渲染器**：智能电视、机顶盒或音箱，可播放来自 Home Assistant 的媒体流。
   - **UPnP/IGD**：暴露网络信息的路由器或网关。
   - **Synology DSM**：本地网络中的 NAS 设备。
   - **HomeKit 配件**：兼容 HomeKit 的设备，如灯具、插座或传感器。
   - **互联网打印协议（IPP）**：网络打印机。

   ![已发现的设备](/images/manual/use-cases/ha-discovered2.png#bordered)

2. 在想要添加的设备上点击 **添加**，然后按照屏幕提示完成配对。部分设备可能会要求输入 PIN 码。
3. 将设备分配到某个区域，然后点击 **完成**。配对完成后，你可以在 **概览** 仪表板上找到并控制该设备。

## 手动添加设备

如果设备没有出现在 **已发现** 面板中，或者你希望自己设置，也可以手动添加。本节以大华 IP 摄像头为例。

开始前，请确保：

- 大华 IP 摄像头已通电，并与 Olares 设备连接到同一个本地网络。
- 你拥有摄像头的管理员用户名和密码。

### 准备大华摄像头

要让 Home Assistant 发现并连接摄像头，你需要先通过大华摄像头网页界面获取其网络信息。

#### 获取摄像头 IP 地址

1. 根据你的操作系统下载设备发现工具：

   - Windows：访问 [大华技术支持网站](https://support.dahuasecurity.com/en/toolsDownloadDetails?IsDpValue=Q93jdSLr94chjRuQ1y%2FcQQ%3D%3D) 下载 **ConfigTool**。
   - macOS：打开 App Store 安装 **CCTV Super Tool**。本教程以 CCTV Super Tool 为例。

2. 打开 CCTV Super Tool，然后点击 **扫描局域网**。
3. 当系统提示允许应用查找本地网络中的设备时，选择 **允许**。
4. 再次点击 **扫描局域网**。摄像头设备将被发现并列出。

   ![发现设备](/images/manual/use-cases/home-assistant-discover-device.png#bordered)

5. 在已发现设备列表中找到你的摄像头，然后记下其 IP 地址。例如 `192.168.50.43`。

#### 获取摄像头端口

1. 在已发现设备列表中，点击你的设备，然后选择 **打开设备网页**。
2. 输入用户名和密码。默认通常为 `admin`，首次登录时必须修改。
3. 进入 **网络** > **端口**。
4. 记下 HTTP 端口（通常为 `80`）和 RTSP 端口（通常为 `554`）。

### 将摄像头添加到 Home Assistant

你可以通过以下任一方式集成摄像头：
- **通用摄像头（RTSP）集成**：快速获取基本视频流。
- **HACS 集成**：获得更深入的设备控制和高级功能。

<Tabs>
<template #RTSP集成>

通用摄像头集成使用摄像头的实时流媒体协议（RTSP）流地址来显示视频。

#### 步骤 1：添加通用摄像头集成

1. 在 Home Assistant 中，进入 **设置** > **设备与服务**。
2. 点击 **添加集成**。
3. 搜索 "Generic Camera" 并选择它。
4. 在 **流源 URL** 字段中，构造并输入 RTSP 地址。大华摄像头通常使用以下格式：

    ```
    rtsp://{username}:{password}@{camera_ip}:{rtsp_port}/cam/realmonitor?channel=1&subtype=1
    ```

    其中：
    - `username`：摄像头网页界面登录用户名。
    - `password`：摄像头网页界面登录密码。
    - `camera_ip`：摄像头 IP 地址。
    - `rtsp_port`：摄像头 RTSP 端口号（通常为 `554`）。
    - `subtype=1`：视频流质量子类型。`0` 为主码流（高分辨率），`1` 为子码流（低分辨率）。

    例如：

    ```
    rtsp://admin:12345Olares@192.168.50.43:554/cam/realmonitor?channel=1&subtype=1
    ```

   ![通用摄像头设置](/images/manual/use-cases/home-assistant-generic-camera.png#bordered)

5. 保持其余默认设置，然后点击 **提交**。
6. 等待预览加载完成。
7. 确认视频流正常后，点击 **提交**。

#### 步骤 2：将摄像头添加到仪表板

1. 在左侧边栏选择 **概览**。
2. 点击右上角的 <i class="material-symbols-outlined">edit</i>。
3. 从 **常用实体** 列表中选择你的摄像头设备，然后点击 **保存**。

   实时摄像头画面将显示在仪表板的 **收藏夹** 区域。

   ![通用摄像头添加到仪表板](/images/manual/use-cases/home-assistant-dashboard-fav.png#bordered)

4. 点击摄像头画面，可以在单独窗口中打开并查看实时流。

</template>
<template #HACS集成>

Home Assistant 社区商店（HACS）允许你下载社区开发的大华专用集成，以获得更丰富的功能。

#### 步骤 1：下载 HACS 插件

1. 打开浏览器，访问 [HACS 官方 GitHub 仓库](https://github.com/hacs/integration)。
2. 在页面右侧选择 **Releases**。
3. 找到 **Assets** 区域，下载最新的 `hacs.zip` 文件。
4. 在本地计算机上解压下载的 `.zip` 文件，得到 `hacs` 文件夹。

#### 步骤 2：上传 HACS 到 Olares

使用 Olares 文件应用将插件放到正确的系统文件夹中，以便 Home Assistant 读取。

1. 从启动台打开 **文件** 应用。
2. 进入 **应用** > **数据** > **homeassistant**。
3. 点击右上角的 <i class="material-symbols-outlined">create_new_folder</i> 新建文件夹。

   ![在文件应用中为 Home Assistant 新建文件夹](/images/manual/use-cases/home-assistant-new-folder.png#bordered)

4. 输入文件夹名称 `custom_components`，然后点击 **创建**。
5. 双击新建的 `custom_components` 文件夹。
6. 点击右上角的 <i class="material-symbols-outlined">drive_folder_upload</i>，选择 **上传文件夹**，然后上传从本地计算机解压得到的 `hacs` 文件夹。

#### 步骤 3：重启 Home Assistant

重启 Home Assistant，使其检测到新上传的 `custom_components` 文件夹。

1. 在 Home Assistant 中，从左侧边栏选择 **设置**，然后选择 **系统**。
2. 点击右上角的 <i class="material-symbols-outlined">power_settings_new</i>。

   ![重启 Home Assistant](/images/manual/use-cases/home-assistant-restart.png#bordered)

3. 选择 **重启 Home Assistant**，然后点击 **重启**。等待重启完成。

#### 步骤 4：授权并安装 HACS

1. 返回 **设置**，然后选择 **设备与服务**。
2. 点击 **添加集成**。
3. 搜索 **HACS**，然后从列表中选择它。

   ![将 HACS 添加到 Home Assistant](/images/manual/use-cases/home-assistant-add-hacs.png#bordered)

4. 阅读通知内容，勾选所有确认框，然后点击 **提交**。
5. 在 **等待设备激活** 窗口中，复制提供的授权密钥，然后点击显示的链接，例如 https://github.com/login/device。
6. 使用你的 GitHub 账号登录。
7. 粘贴复制的授权密钥，然后点击 **Authorize hacs**。
8. 返回 Home Assistant。HACS 现在会出现在左侧边栏和 **集成** 列表中。

   ![HACS 已添加到 Home Assistant](/images/manual/use-cases/home-assistant-hacs-added.png#bordered)

#### 步骤 5：安装并配置大华集成

1. 从左侧边栏选择 **HACS**，然后搜索 "dahua"。

   ![在 HACS 中搜索产品](/images/manual/use-cases/home-assistant-hacs-search.png#bordered)

2. 从列表中选择目标设备，然后点击 **下载**。
3. 进入 **设置** > **系统**，再次重启 Home Assistant 以应用新集成。
4. 重启后，从左侧边栏选择 **概览**。
5. 点击右上角的 <i class="material-symbols-outlined">add</i>，然后选择 **添加设备**。
6. 搜索品牌名称 "Dahua"，然后从结果中选择它。
7. 在 **添加大华摄像头** 窗口中，使用之前记下的信息配置设备：

    - **用户名**：输入摄像头网页界面登录用户名。
    - **密码**：输入摄像头网页界面登录密码。
    - **地址**：输入摄像头 IP 地址，例如 `192.168.50.43`。
    - **端口**：输入 HTTP 端口号，例如 `80`。
    - **RTSP 端口**：输入 RTSP 端口号，例如 `554`。

8. 保持其余设置默认，然后点击 **提交**。
9. 为设备指定一个名称，然后点击 **提交**。
10. 将其分配到某个区域，例如 **前门**，然后点击 **完成**。

#### 步骤 6：在仪表板上监控设备

1. 从左侧边栏选择 **概览**，然后在 **区域** 部分找到你分配的区域中的摄像头。
2. 点击该区域查看实时画面和设备控制项。

   ![Home Assistant 仪表板区域部分](/images/manual/use-cases/home-assistant-dashboard-areas.png#bordered)

</template>
</Tabs>
