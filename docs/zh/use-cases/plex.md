---
outline: [2, 3]
description: 在 Olares 上使用 Plex 整理个人视频，通过浏览器、手机或电脑播放，并启用硬件加速转码。
head:
  - - meta
    - name: keywords
      content: Olares, Plex Media Server, 个人媒体库, 视频播放, NVIDIA, 硬件转码, Plex Pass
app_version: "0.0.10"
doc_version: "1.1"
doc_updated: "2026-11-03"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/plex.md)为准。
:::

# 使用 Plex 整理并播放你的个人视频

Plex Media Server 是一款媒体服务器，可将你的视频和音乐文件整理为媒体库，并串流到 Plex 播放器。在 Olares 上，你可以把媒体收藏集中存放在自己的设备中，再通过浏览器或 Plex 客户端播放。

## 学习目标

通过本教程，你将学习：

- 安装 Plex、选择加速模式，并绑定自己的 Plex 账号。
- 整理 Olares 中的媒体目录，并为个人视频、电影或音乐选择合适的媒体库。
- 在浏览器中选择播放质量和字幕，并查看是否正在使用硬件转码。
- 通过 LarePass VPN，从手机或电脑上的 Plex 客户端连接自己的媒体服务器。

## 前提条件

- Olares 1.12.7 或更高版本。
- 一个 Plex 账号。免费账号即可绑定服务器并创建媒体库。
- 硬件加速转码需要兼容的 NVIDIA GPU，且服务器所属账号有有效的 Plex Pass。

## 安装 Plex Media Server

1. 打开 Market，搜索 "Plex Media Server"。
2. 点击 **Get**，然后点击 **Install**。
3. 在 **Choose a hardware accelerator** 弹窗中选择模式：
   - **CPU**：使用软件转码。
   - **NVIDIA GPU**：使用 NVIDIA GPU 进行[硬件加速转码](#使用硬件加速转码)。

   :::info 加速模式
   安装后无法直接切换加速模式。
   :::

4. 点击 **Confirm**。如果提示填写 **PLEX_CLAIM**，保持为空并确认继续，下一节再绑定账号。
5. 等待安装完成，按钮变为 **Open**。

## 绑定 Plex 账号

你会在 Launchpad 中看到 Plex Media Server 的两个入口。先通过 **Plex Account Setup** 将服务器绑定到自己的 Plex 账号，再通过 **Plex Media Server** 管理媒体库和播放内容。

1. 从 Launchpad 打开 **Plex Account Setup**。
2. 在另一个浏览器标签页打开 [Plex claim 页面](https://www.plex.tv/claim/)，登录准备绑定的 Plex 账号。
3. 复制以 `claim-` 开头的临时令牌。
4. 返回 Plex Account Setup，将令牌粘贴到 **Claim token**，点击 **Connect account**。

   :::tip 绑定失败
   绑定失败时，重新打开 claim 页面获取新令牌，再返回设置页连接。
   :::

5. 请求完成后，点击 **Open Plex Web**。
6. 在 Plex Web 中登录同一个账号。首次打开时，按页面提示设置一个方便识别的服务器名称，例如 `Home media`。
7. 打开 Plex Web 设置，选中自己的服务器，再进入 **General**，确认服务器登录的是刚才绑定的账号。

## 添加视频媒体库

Plex 媒体库会汇集所选文件夹中的媒体，方便你浏览和播放。下面先用一段旅行录像创建个人视频媒体库，再介绍如何添加电影或音乐收藏。

### 上传自己的视频

1. 打开 Olares 文件管理器，进入 `Home/Movies`。
2. 新建一个子文件夹，例如 `Family`。
3. 将自己的视频上传到该目录。以下以 `Home/Movies/Family/Family-trip.mp4` 为例，文件名可以自行替换。

整理后的目录可以是：

```text
Home/
  Movies/
    Family/
      Family-trip.mp4
  Music/
```

为每个媒体库使用独立的文件夹。例如，个人录像放在 `Family`，电影收藏放在 `Films`。

上传大文件时，可以使用 [LarePass 桌面客户端上传](../manual/olares/files/add-edit-download.md#通过-larepass-桌面端上传)。

### 在 Plex 中选择媒体目录

Plex 使用应用内路径读取媒体文件。选择文件夹时，使用下表右侧的路径。

| Olares 中的目录 | Plex 中的路径 |
| --- | --- |
| `Home/Movies` | `/media/Movies` |
| `Home/Music` | `/media/Music` |

1. 打开 Plex Web 的设置，选中刚刚绑定的服务器。
2. 在 **Manage** 下打开 **Libraries**，点击 **Add Library**。
3. 对于旅行录像等个人视频，选择 **Other Videos**，将媒体库命名为 `Family videos`。
4. 选择媒体库语言，点击 **Next**。
5. 进入添加文件夹的步骤，点击 **Browse for Media Folders**，选择 `/media/Movies/Family`。
6. 点击 **Add**，再点击 **Add Library**，等待 Plex 扫描文件。
7. 打开 `Family videos`，找到刚刚上传的视频。

Plex 对 `Home/Movies` 和 `Home/Music` 的访问为只读。需要整理、重命名或删除源文件时，在 Olares 文件管理器中操作。

### 扩展到电影和音乐收藏

同样的方法也可以用于电影或音乐。新增媒体库时，按内容类型和实际文件位置选择设置：

| 内容 | 媒体库类型 | 目录示例 |
| --- | --- | --- |
| 旅行录像、家庭视频 | **Other Videos** | `/media/Movies/Family` |
| 电影收藏 | **Movies** | `/media/Movies/Films` |
| 音乐收藏 | **Music** | `/media/Music` |

选择 **Movies** 时，清楚的片名和年份有助于 Plex 匹配海报及简介，媒体库语言决定这些信息的语言。

以后向这些目录上传新文件后，可以在对应媒体库的菜单中选择 **Scan Library Files**，让 Plex 重新扫描。

## 播放视频并调整播放方式

### 先播放原始视频

本地视频播放[免费](https://support.plex.tv/articles/202526943-plex-free-vs-paid/)。如果 Plex 将浏览器连接判定为远程连接，则适用[远程播放的订阅要求](#从手机和电脑连接-plex)。

1. 打开 `Family videos`，选择刚才添加的旅行录像。
2. 点击播放，先保持原始质量。

Plex 会根据文件格式、播放器能力和质量设置决定是否转码。先以原始质量观看，可以避免不必要的画质损失。

### 选择合适的播放质量

Plex 可能采用以下几种播放方式：

| 播放方式 | 含义 |
| --- | --- |
| Direct Play | 播放器直接播放原文件，服务器不转换视频 |
| Direct Stream | 服务器调整封装，视频通常不重新编码 |
| Transcode | 服务器转换视频或音频，以适应播放器或质量设置 |

网络和播放器允许时，优先保持原始质量。播放频繁缓冲时，可以在播放器的质量菜单中降低分辨率或码率，例如将 1080p 内容改为 720p。降低视频质量通常会增加服务器的转码工作。

### 添加并选择字幕

已有字幕可以与视频一起放入媒体目录。以旅行录像的英文字幕为例：

```text
Home/Movies/Family/
  Family-trip.mp4
  Family-trip.en.srt
```

`Family-trip` 与视频文件名对应，`en` 是英文字幕的语言代码。Plex 也支持 ASS 外部字幕。文本字幕建议保存为 UTF-8，减少乱码。

1. 通过 Olares 文件管理器，将字幕上传到视频所在的文件夹。
2. 在 Plex 中扫描该媒体库。未识别新字幕时，刷新对应视频的元数据。
3. 打开视频的字幕选项，选择刚刚添加的字幕，再开始播放。

部分字幕需要烧录进画面，因此会触发视频转码。

## 从手机和电脑连接 Plex

在手机或电脑上开启 LarePass VPN 后，可以连接 Olares 上的 Plex 服务器。播放期间保持专用网络连接开启。

:::info 远程视频播放需要订阅
Plex 判定为远程的视频播放，需要本教程使用的账号有有效的 Plex Pass 或 Remote Watch Pass。这也适用于被判定为远程的 VPN 连接。详见 [Plex 远程播放要求](https://support.plex.tv/articles/requirements-for-remote-playback-of-personal-media/)。
:::

### 开启 LarePass VPN

在要播放媒体的手机或电脑上[安装 LarePass](../manual/larepass/index.md)，使用 Olares ID 登录，并确认当前连接的是运行 Plex 的 Olares 设备。

<!--@include: ../reusables/larepass-vpn.md#enable-larepass-vpn-->

在 iOS 或 macOS 上，若系统提示添加 VPN 配置，请允许添加。等待连接完成后再打开 Plex，详见 [LarePass 连接状态说明](../manual/larepass/private-network.md#了解连接状态)。

### 在 Plex 客户端中播放

1. 在手机或电脑上安装并打开 [Plex 客户端](https://www.plex.tv/media-server-downloads/#plex-app)。
2. 使用前面绑定服务器时的 Plex 账号登录。
3. 在媒体来源或服务器列表中，选择你在 Plex 中设置的服务器名称，例如 `Home media`。
4. 打开 `Family videos` 媒体库，选择旅行录像并播放。

电脑也可以在开启 LarePass VPN 后，通过浏览器打开 Olares 中的 Plex Media Server 入口，继续使用 Plex Web。

## 使用硬件加速转码

硬件转码把部分视频转换工作交给 GPU，适用于播放器不支持原始格式，或需要降低分辨率和码率的场景。

以下步骤使用安装时选择的 NVIDIA 模式，以及服务器所属的 Plex Pass 账号。

### 开启硬件加速

1. 打开 Plex Web 设置，选中自己的服务器，再进入 **Transcoder**。
2. 显示高级选项，开启 **Use hardware acceleration when available** 和 **Use hardware-accelerated video encoding**。
3. 在 **Hardware transcoding device** 中选择为该实例分配的 NVIDIA GPU。
4. 点击 **Save Changes**。

设置保存后即可生效，无需重启 Plex。

### 查看当前播放是否使用 GPU

1. 播放视频，打开播放器的质量设置，选择低于源视频的分辨率或码率。例如，对 1080p 视频选择 720p 转换选项。
2. 在 Plex Web 的 **Activity** 中打开 **Dashboard**，展开当前播放的详细信息。
3. 查看视频状态。显示转换时，视频格式旁的 `(hw)` 表示对应环节正在使用硬件加速。
4. 查看完毕后，可以将播放质量恢复为原始质量。

Direct Play 或 Direct Stream 表示视频正在播放，且未重新编码。

## 常见问题

### 在 Plex 中看不到上传的视频

先确认文件已经上传完成，再检查媒体库目录。例如，`Home/Movies/Family` 在 Plex 中对应 `/media/Movies/Family`。确认路径后，运行 **Scan Library Files**。

### 看不到硬件转码选项，或播放时没有 `(hw)`

先确认服务器所属账号有有效的 Plex Pass，再检查这台 Plex 实例使用的是 NVIDIA 模式。最后查看当前播放是否真的在转换视频。如果显示 Direct Play 或 Direct Stream，没有 `(hw)` 通常不需要处理。

### 字幕没有出现或显示乱码

确认字幕与视频位于同一目录，文件名主体一致，扩展名是支持的字幕格式。重新扫描媒体库，并刷新视频元数据。出现乱码时，先将字幕保存为 UTF-8，再重新上传。字幕识别的更多设置见 [Plex 本地字幕指南](https://support.plex.tv/articles/200471133-adding-local-subtitles-to-your-media/)。

### 需要重新启动 Plex 时怎么操作

1. 在 Olares 中进入**设置** > **应用** > **Plex Media Server**，点击**暂停**。
2. 等待应用停止后，点击**恢复**，重新启动应用。

如果资源占用导致应用无法恢复，参见[恢复应用](../manual/olares/market/market.md#如何恢复应用)。

### 开启 VPN 后，Plex 客户端仍找不到服务器

先确认 LarePass 已连接到运行 Plex 的 Olares，且 Plex 应用处于运行状态。然后核对客户端登录的 Plex 账号是否与服务器所属账号一致。你可以回到 Plex Web 的服务器 **General** 设置中查看归属，再重新打开客户端。

## 了解更多

- [Jellyfin](jellyfin.md)：在 Olares 上搭建开源媒体服务器，作为 Plex 的替代选择。
- [LarePass 专用网络](../manual/larepass/private-network.md)：了解手机和电脑上的 VPN 开关及连接状态。
- [Plex 媒体库创建指南](https://support.plex.tv/articles/200288926-creating-libraries/)：了解媒体类型和目录选择。
- [Plex 本地字幕指南](https://support.plex.tv/articles/200471133-adding-local-subtitles-to-your-media/)：了解字幕命名、语言代码和支持格式。
- [Plex 硬件加速指南](https://support.plex.tv/articles/115002178853-using-hardware-accelerated-streaming/)：了解 Plex Pass 要求与硬件转码状态。
