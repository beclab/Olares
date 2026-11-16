---
outline: [2, 3]
description: 在 Olares 上使用 Plex 整理个人视频，通过浏览器、手机或电脑播放，并启用硬件加速转码。
head:
  - - meta
    - name: keywords
      content: Olares, Plex Media Server, 个人媒体库, 视频播放, NVIDIA, 硬件转码, Plex Pass
app_version: "0.0.14"
doc_version: "1.2"
doc_updated: "2026-11-16"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/plex.md)为准。
:::

# 使用 Plex 整理并播放你的个人视频

Plex Media Server 是一款用于管理和串流播放视频、音乐的应用。在 Olares 上，你可以把媒体收藏存放在自己的设备中，再通过浏览器或 Plex 客户端播放。

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
   绑定失败时，从 claim 页面复制一个新令牌，再返回设置页重试。
   :::

5. 请求完成后，点击 **Open Plex Web**。
6. 在 Plex Web 中登录同一个账号。首次打开时，在欢迎页点击 **Got It!**，进入 **Server Setup**。
7. 设置一个方便识别的服务器名称，例如 `Home media`。取消勾选 **Allow me to access my media outside my home**，然后点击 **Next**。本教程通过 LarePass VPN 访问服务器。
8. 在 **Media Library** 页面点击 **Next**，稍后再添加媒体库，然后点击 **Done** 完成设置。
9. 打开 Plex Web 设置，选中自己的服务器，再进入 **General**，确认 **Server claimed by** 显示的是刚才绑定的账号。

## 添加视频媒体库

Plex 媒体库让你可以浏览和播放所选文件夹中的媒体。下面以一段旅行录像为例，创建个人视频媒体库。你也可以按相同的步骤添加电影和音乐媒体库。

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

Olares 文件管理器中的文件夹，在 Plex 中使用不同的路径。添加媒体库时，请选择下表中对应的 Plex 路径。

| Olares 中的目录 | Plex 中的路径 |
| --- | --- |
| `Home/Movies` | `/media/Movies` |
| `Home/Music` | `/media/Music` |

1. 打开 Plex Web 的设置，选中刚刚绑定的服务器。
2. 在 **Manage** 下打开 **Libraries**，点击 **Add Library**。
3. 对于旅行录像等个人视频，选择 **Other Videos**，将媒体库命名为 `Family videos`。
4. 选择媒体库语言，点击 **Next**。
5. 点击 **Browse For Media Folder**，选择 `/media/Movies/Family`。
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

创建 **Movies** 媒体库时，在文件名中写明片名和上映年份，有助于 Plex 匹配海报和简介。Plex 会按你选择的媒体库语言获取简介。

以后上传新文件后，打开对应媒体库的菜单，选择 **Scan Library Files**，将新文件加入 Plex。

## 播放视频并调整播放方式

### 先播放原始视频

本地视频播放[免费](https://support.plex.tv/articles/202526943-plex-free-vs-paid/)。[远程播放需要订阅](#从手机和电脑连接-plex)，通过浏览器播放也一样。

1. 打开 `Family videos`，选择刚才添加的旅行录像。
2. 以原始质量播放视频。

Plex 会根据文件格式、播放器能力和质量设置决定是否转码。先以原始质量观看，可以避免不必要的画质损失。

### 选择合适的播放质量

Plex 可能采用以下几种播放方式：

| 播放方式 | 含义 |
| --- | --- |
| Direct Play | 播放器直接播放原文件，服务器不转换视频 |
| Direct Stream | 服务器调整文件封装，不重新编码视频 |
| Transcode | 服务器转换视频或音频，以适应播放器或质量设置 |

播放频繁缓冲时，可以降低分辨率或码率：

1. 在播放器中打开 **Settings** > **Quality**。
2. 点击 **Show all**，展开可用的转换选项。
3. 选择低于源视频的质量。例如，播放 1080p 视频时，选择 **Convert to 720p HD (High)**。

Plex 会将视频转码为所选质量。要恢复原始质量，打开质量菜单，选择 **Play Original Quality**。

### 添加并选择字幕

将字幕文件放在视频所在的文件夹中。以英文字幕为例，文件可以这样命名：

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

## 通过本地 HTTPS 地址连接

通过 Olares 使用浏览器时，可以打开 Olares 中的私有入口。使用 Plex 原生客户端时，先启用 Overlay，为 Plex 分配独立的局域网 IP，再将安全连接地址添加到 Plex。播放设备需要位于同一局域网，或通过 VPN 路由访问该网络。

### 获取本地 HTTPS 地址

1. 打开 Olares 的**设置 → 网络 → Overlay**，启用 Plex，等待页面显示 IP 地址。
2. 从启动台打开 **Plex Account Setup**。绑定账号后，这个入口仍会保留。
3. 在 **Local connection** 区域点击 **Refresh local address**。
4. 页面显示 **HTTPS certificate verified** 后，点击 **Copy address**。页面会根据当前 Overlay IP 自动生成 HTTPS `plex.direct` 地址，无需手动拼接。

如果页面提示启用 Overlay，请先完成第 1 步。如果安全地址尚未就绪，请检查 Plex 账号是否已绑定，等待 Plex 启动完成后刷新地址。

### 将地址添加到 Plex

1. 打开 Plex Web，进入**设置**并选择自己的服务器。
2. 打开**网络**，点击**显示高级设置**。
3. 将复制的地址粘贴到**自定义服务器访问 URL**。保留已有地址，多个地址用英文逗号分隔。
4. 点击**保存修改**。

这里填写复制的基础地址，不要添加 `/web/` 或 `/identity`。Plex Account Setup 只显示和复制地址，不会自动将其保存到 Plex。这个 HTTPS 地址仍指向局域网 IP，不会让服务器自动具备公网访问能力。

:::warning 重启 Plex 后检查连接地址
在当前 Olares 版本中，停止并重新启动应用后，Overlay IP 可能发生变化。HTTPS 地址包含这个 IP，因此之前保存的地址可能无法继续连接服务器。

重启 Plex 后，打开 **Plex Account Setup**，点击 **Refresh local address**。如果 IP 已变化，请复制新的 HTTPS 地址，替换 Plex **自定义服务器访问 URL** 中原来的本地地址。同时更新浏览器书签或客户端中手动填写的地址，再重新打开 Plex 客户端。应用升级后也请检查地址。
:::

### 在另一台电脑的浏览器中打开 Plex

在同一局域网的另一台电脑上，将复制的 HTTPS 地址粘贴到浏览器，并在末尾添加 `/web/`。使用服务器所属的 Plex 账号，或有权访问媒体库的账号登录。

Plex 可能打开 `app.plex.tv` 的授权页面，并显示本地 `plex.direct` 地址。核对它与刚复制的地址一致后，授权登录并返回 Plex Web。选择自己的服务器，再打开 `Family videos` 媒体库。

播放个人媒体时，请选择自己服务器中的媒体库。Plex 的**电影和电视节目**来源提供在线内容，使用的媒体和字幕来源与本地媒体库不同。

### 在电视上播放

1. 将电视连接到 Plex Overlay IP 所在的局域网。
2. 打开电视上的 Plex 应用，使用服务器所属账号或有权访问媒体库的账号登录。
3. 保存自定义服务器地址后，重新打开电视应用，选择自己的服务器，例如 `Home media`。
4. 打开 `Family videos`，播放旅行录像。先使用原始质量。如果电视无法直接播放源格式，Plex 可以为播放器转码。

如果服务器没有出现，请检查自定义服务器地址和账号权限。电视登录账号后，仍需要能访问服务器地址。异地电视需要具备通向 Overlay 网络的路由，不能认为在另一台手机或电脑上开启 VPN 后，电视也会自动接入。

## 从手机和电脑连接 Plex

在手机或电脑上开启 LarePass VPN 后，可以连接 Olares 上的 Plex 服务器。播放期间保持专用网络连接开启。

:::info 远程视频播放需要订阅
远程视频播放需要你使用的账号有有效的 Plex Pass 或 Remote Watch Pass。Plex 也可能将 VPN 连接判定为远程连接。详见 [Plex 远程播放要求](https://support.plex.tv/articles/requirements-for-remote-playback-of-personal-media/)。
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

在电脑上使用浏览器观看时，保持 LarePass VPN 开启，然后从 Launchpad 打开 **Plex Media Server**。

## 使用硬件加速转码

硬件加速转码使用 GPU 解码或编码视频。播放器不支持源文件格式，或你选择了较低的分辨率或码率时，Plex 会进行转码。

以下步骤要求安装时已选择 NVIDIA 模式，且服务器所属账号有有效的 Plex Pass。

### 开启硬件加速

1. 打开 Plex Web 设置，选中自己的服务器，再进入 **Transcoder**。
2. 显示高级选项，开启 **Use hardware acceleration when available** 和 **Use hardware-accelerated video encoding**。
3. 在 **Hardware transcoding device** 中选择为该实例分配的 NVIDIA GPU。
4. 点击 **Save Changes**。

设置保存后即可生效，无需重启 Plex。

### 查看当前播放是否使用 GPU

1. 播放视频，并[选择低于源视频的质量](#选择合适的播放质量)，例如将 1080p 视频转换为 720p。
2. 在 Plex Web 的 **Activity** 中打开 **Dashboard**，展开当前播放的详细信息。
3. 查看视频格式旁是否有 `(hw)` 标记。该标记表示正在使用硬件加速解码或编码。
4. 查看完毕后，可以将播放质量恢复为原始质量。

### HEVC 播放异常时改用 H.264 输出

如果 Dolby Vision/HDR10 视频在 Chrome 中进行 HEVC 转码时出现画面冻结或黑屏，打开服务器的**转码器**设置，将**启用 HEVC 视频编码**设为**从不（Never）**。保持 **Use hardware acceleration when available** 和 **Use hardware-accelerated video encoding** 开启，保存后重新开始播放。

需要转换视频时，Plex 可以使用 H.264 硬件转码。关闭 HEVC 输出不会禁止播放 HEVC 源文件。如果仍无法播放，可换一个视频，检查问题是否只出现在特定文件上。

## 常见问题

### 在 Plex 中看不到上传的视频

先确认文件已经上传完成，再检查媒体库目录。例如，`Home/Movies/Family` 在 Plex 中对应 `/media/Movies/Family`。确认路径后，运行 **Scan Library Files**。

### 为什么看不到硬件转码选项？

确认服务器绑定的账号有有效的 Plex Pass。在服务器的 **Transcoder** 设置中展开高级选项，即可找到硬件加速选项。

### 为什么播放时没有 `(hw)` 标记？

Direct Play 和 Direct Stream 不会重新编码视频，因此不会使用硬件转码。要查看 GPU 是否参与转码，可以[降低播放质量，再打开 Dashboard](#查看当前播放是否使用-gpu)。

如果视频正在转码，但没有 `(hw)` 标记，请检查安装时是否选择了 NVIDIA 模式，以及 **Transcoder** 下的两个硬件加速选项是否已开启。

### 字幕没有出现或显示乱码

将字幕和视频放在同一文件夹中，检查文件名是否对应，例如 `Family-trip.mp4` 和 `Family-trip.en.srt`，然后重新扫描媒体库并刷新视频元数据。出现乱码时，将字幕保存为 UTF-8，再重新上传。支持的格式和命名规则见 [Plex 本地字幕指南](https://support.plex.tv/articles/200471133-adding-local-subtitles-to-your-media/)。

### 需要重新启动 Plex 时怎么操作

1. 在 Olares 中进入**设置** > **应用** > **Plex Media Server**，点击**暂停**。
2. 等待应用停止后，点击**恢复**，重新启动应用。

如果资源占用导致应用无法恢复，参见[恢复应用](../manual/olares/market/market.md#如何恢复应用)。

### 开启 VPN 后，Plex 客户端仍找不到服务器

确认 Plex 正在运行，且 LarePass 已连接到运行 Plex 的 Olares 设备。在 Plex Web 中打开服务器的 **General** 设置，查看服务器所属账号。使用该账号登录客户端，再重新打开服务器列表。

## 了解更多

- [Jellyfin](jellyfin.md)：在 Olares 上搭建开源媒体服务器，作为 Plex 的替代选择。
- [LarePass 专用网络](../manual/larepass/private-network.md)：了解手机和电脑上的 VPN 开关及连接状态。
- [Plex 媒体库创建指南](https://support.plex.tv/articles/200288926-creating-libraries/)：了解媒体类型和目录选择。
- [Plex 本地字幕指南](https://support.plex.tv/articles/200471133-adding-local-subtitles-to-your-media/)：了解字幕命名、语言代码和支持格式。
- [Plex 硬件加速指南](https://support.plex.tv/articles/115002178853-using-hardware-accelerated-streaming/)：了解 Plex Pass 要求与硬件转码状态。
