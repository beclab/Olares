---
outline: [2, 3]
description: 在 Olares 上使用 Plex 整理个人视频，获取 Overlay HTTPS 地址，通过浏览器、手机、电脑或电视播放，并按需启用 NVIDIA 硬件转码。
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

Plex Media Server 将你的视频和音乐文件整理为媒体库，并通过 Plex 播放器串流播放。在 Olares 上，你可以把文件集中存放在自己的设备中，再通过浏览器浏览和播放。

本文以一段旅行录像为例，先完成账号绑定、文件入库和浏览器播放，再按需添加字幕、扩展电影或音乐收藏，并启用 NVIDIA 硬件转码。

## 学习目标

通过本教程，你将学习：

- 将 Plex 服务器绑定到自己的 Plex 账号。
- 整理 Olares 中的媒体目录，并为个人视频、电影或音乐选择合适的媒体库。
- 在浏览器中选择播放质量和字幕，并查看是否正在使用硬件转码。
- 从 Plex Account Setup 获取 Overlay HTTPS 地址，连接浏览器、手机、电脑和电视客户端。
- 通过 LarePass VPN 在外访问，并在 HEVC 播放异常时改用 H.264 输出。

## 前提条件

- Olares 1.12.7 或更高版本。
- 一个可以登录的 Plex 账号。
- 一段自己的视频文件，例如旅行录像。

NVIDIA 硬件转码的额外条件在后文说明。你可以先完成媒体库设置，再决定是否使用该功能。

## 绑定 Plex 账号

Plex 有两个入口，分别用于账号连接和日常播放：

| 入口 | 用途 |
| --- | --- |
| Plex Account Setup | 绑定 Plex 账号并复制服务器的本地 HTTPS 地址 |
| Plex Media Server | 打开 Plex Web，管理媒体库和播放内容 |

Plex 网页中的账号头像表示当前登录的账号，不能单独说明服务器归属。先完成服务器绑定，再使用同一个账号进入 Plex Web。

1. 从 Launchpad 打开 **Plex Account Setup**。
2. 在另一个浏览器标签页打开 [Plex claim 页面](https://www.plex.tv/claim/)，登录准备绑定的 Plex 账号。
3. 复制以 `claim-` 开头的临时令牌。
4. 返回 Plex Account Setup，将令牌粘贴到 **Claim token**，点击 **Connect account**。
5. 页面确认服务器已绑定后，点击 **Open Plex Web**。
6. 在 Plex Web 中登录同一个账号。首次打开时，按页面提示设置一个方便识别的服务器名称，例如 `Home media`。
7. 打开 Plex 的服务器设置，在 **General** 页面查看服务器所属账号，确认是刚才绑定的账号。

:::tip 已经连接过账号
如果 Plex Account Setup 已显示服务器已连接，可以直接打开 Plex Web。需要使用硬件转码时，确认服务器所属账号就是拥有 Plex Pass 的账号。
:::

令牌有效期较短。绑定失败时，重新打开 claim 页面获取新令牌，再返回设置页连接。

## 添加视频媒体库

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

把不同用途的视频放入不同子文件夹，创建媒体库时就能只选择需要的目录。例如，个人录像放在 `Family`，电影收藏放在另一个 `Films` 文件夹。不要把它们都指向同一个上级目录，否则不同媒体库可能反复扫描同一批文件。

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
4. 根据内容选择媒体库的语言。该语言用于从网络获取的简介等信息，不会改变视频本身的音轨。
5. 进入添加文件夹的步骤，点击 **Browse for Media Folders**，选择 `/media/Movies/Family`。
6. 添加目录并完成媒体库创建，等待 Plex 扫描文件。
7. 打开 `Family videos`，找到刚刚上传的视频。

Plex 对 `Home/Movies` 和 `Home/Music` 的访问为只读。需要整理、重命名或删除源文件时，在 Olares 文件管理器中操作。

### 扩展到电影和音乐收藏

同样的方法也可以用于电影或音乐。新增媒体库时，按内容类型和实际文件位置选择设置：

| 内容 | 媒体库类型 | 目录示例 |
| --- | --- | --- |
| 旅行录像、家庭视频 | **Other Videos** | `/media/Movies/Family` |
| 电影收藏 | **Movies** | `/media/Movies/Films` |
| 音乐收藏 | **Music** | `/media/Music` |

选择 **Movies** 时，清楚的片名和年份有助于 Plex 匹配海报及简介。个人录像使用 **Other Videos**，避免把它当作商业电影去匹配。为音乐新建独立的 **Music** 媒体库，不要把音乐目录加进视频媒体库。

以后向这些目录上传新文件后，可以在对应媒体库的菜单中选择 **Scan Library Files**，让 Plex 重新扫描。

## 播放视频并调整播放方式

### 先播放原始视频

1. 打开 `Family videos`，选择刚才添加的旅行录像。
2. 点击播放，先保持原始质量。
3. 从播放器中尝试暂停、继续播放和拖动进度，确认可以观看需要的片段。

播放器中的质量选择会影响服务器是否需要转换视频。先以原始质量观看，可以避免不必要的画质损失。

Plex 会根据文件格式、播放器能力和质量设置决定是否转码。可以直接播放时，无需为了使用 GPU 而降低画质。

### 选择合适的播放质量

Plex 可能采用以下几种播放方式：

| 播放方式 | 含义 |
| --- | --- |
| Direct Play | 播放器直接播放原文件，服务器不转换视频 |
| Direct Stream | 服务器调整封装，视频通常不重新编码 |
| Transcode | 服务器转换视频或音频，以适应播放器或质量设置 |

网络和播放器允许时，优先保持原始质量。播放频繁缓冲时，可以在播放器的质量菜单中降低分辨率或码率，例如将 1080p 内容改为 720p。降低视频质量通常会增加服务器的转码工作。

### 添加并选择字幕

已有字幕可以与视频一起放入媒体目录。以旅行录像的中文字幕为例：

```text
Home/Movies/Family/
  Family-trip.mp4
  Family-trip.zh.srt
```

`Family-trip` 与视频文件名对应，`zh` 是中文字幕的语言代码。Plex 也支持 ASS 外部字幕。文本字幕建议保存为 UTF-8，减少中文乱码。

1. 通过 Olares 文件管理器，将字幕上传到视频所在的文件夹。
2. 在 Plex 中扫描该媒体库。未识别新字幕时，刷新对应视频的元数据。
3. 打开视频的字幕选项，选择刚刚添加的字幕，再开始播放。

部分字幕需要由服务器烧录进画面，因此开启字幕后，播放方式可能从直接播放变成转码。字幕的语言、时间轴和视频版本应当对应，Plex 不会自动修复字幕时间不匹配的问题。

## 从手机、电脑和电视连接 Plex

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

### 通过 LarePass VPN 连接

手机或电脑位于家庭局域网之外时，打开 LarePass，连接到运行 Plex 的 Olares 设备。

| 设备 | 操作 |
| --- | --- |
| 手机 | 打开 LarePass 的**设置**，在**我的 Olares** 卡片中开启 VPN |
| 电脑 | 点击 LarePass 左上角的头像，开启**专用网络连接** |

首次使用时，按操作系统提示允许添加 VPN 配置。等待连接完成后再打开 Plex，并在播放期间保持连接。详细步骤见[在 LarePass 中启用专用网络](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。

打开 Plex 客户端，登录并选择自己的服务器和媒体库。复制的本地地址需要能通过 VPN 访问。电脑也可以在开启 LarePass VPN 后，通过浏览器打开 Olares 中的 Plex Media Server 入口。

## 使用 NVIDIA 硬件转码

硬件转码把部分视频转换工作交给 GPU，适用于播放器不支持原始格式，或需要降低分辨率和码率的场景。直接播放的视频不需要为了使用 GPU 而强制转换。

本节适用于已选择 NVIDIA 加速模式的 Plex 实例，并要求服务器所属的 Plex 账号有有效的 Plex Pass。CPU 模式不会仅靠 Plex 内部的设置开关获得 NVIDIA GPU。

### 开启硬件加速

1. 打开 Plex Web 设置，选中自己的服务器，再进入 **Transcoder**。
2. 显示高级选项，开启 **Use hardware acceleration when available** 和 **Use hardware-accelerated video encoding**。
3. 在 **Hardware transcoding device** 中选择为该实例分配的 NVIDIA GPU。
4. 点击 **Save Changes**。

设置保存后即可生效，无需重启 Plex。服务器所属账号需要有有效的 Plex Pass，仅在播放器中登录另一个订阅账号不会改变服务器的订阅状态。

### 查看当前播放是否使用 GPU

1. 播放视频，打开播放器的质量设置，选择低于源视频的分辨率或码率。例如，对 1080p 视频选择 720p 转换选项。
2. 在 Plex Web 的 **Activity** 中打开 **Dashboard**，展开当前播放的详细信息。
3. 查看视频状态。显示转换时，视频格式旁的 `(hw)` 表示对应环节正在使用硬件加速。
4. 查看完毕后，可以将播放质量恢复为原始质量。

如果仍显示 Direct Play 或 Direct Stream，说明这次播放没有触发视频转码。对于码率较低的视频，Plex 可能继续使用原始视频，无需反复调整 GPU 设置。

### HEVC 播放异常时改用 H.264 输出

如果 Dolby Vision/HDR10 视频在 Chrome 中进行 HEVC 转码时出现画面冻结或黑屏，打开服务器的**转码器**设置，将**启用 HEVC 视频编码**设为**从不（Never）**。保持 **Use hardware acceleration when available** 和 **Use hardware-accelerated video encoding** 开启，保存后重新开始播放。

需要转换视频时，Plex 可以使用 H.264 硬件转码。关闭 HEVC 输出不会禁止播放 HEVC 源文件。如果仍无法播放，可换一个视频，检查问题是否只出现在特定文件上。

## 常见问题

### 在 Plex 中看不到上传的视频

先确认文件已经上传完成，再检查媒体库目录。例如，`Home/Movies/Family` 在 Plex 中对应 `/media/Movies/Family`，不是电脑本地的下载目录。确认路径后，运行 **Scan Library Files**。音乐文件应加入 **Music** 类型的媒体库。

### 看不到硬件转码选项，或播放时没有 `(hw)`

先确认服务器所属账号有有效的 Plex Pass，再检查这台 Plex 实例使用的是 NVIDIA 模式。最后查看当前播放是否真的在转换视频。如果显示 Direct Play 或 Direct Stream，没有 `(hw)` 通常不需要处理。

### 字幕没有出现或显示乱码

确认字幕与视频位于同一目录，文件名主体一致，扩展名是支持的字幕格式。重新扫描媒体库，并刷新视频元数据。中文乱码时，先将字幕保存为 UTF-8，再重新上传。字幕识别的更多设置见 [Plex 本地字幕指南](https://support.plex.tv/articles/200471133-adding-local-subtitles-to-your-media/)。

### 需要重新启动 Plex 时怎么操作

在 Olares 中先停止 Plex，等待停止完成后，再启动或恢复该应用。保留原来的 GPU 选择。

### 开启 VPN 后，Plex 客户端仍找不到服务器

先确认 LarePass 已连接到运行 Plex 的 Olares，且 Plex 应用处于运行状态。检查复制的 Overlay HTTPS 地址能否从播放设备访问，以及 Plex 的“自定义服务器访问 URL”是否包含当前地址。然后确认客户端登录的是服务器所属账号，或有媒体库访问权限的账号，再重新打开客户端。

## 了解更多

- [LarePass 专用网络](../manual/larepass/private-network.md)：了解手机和电脑上的 VPN 开关及连接状态。
- [Plex 媒体库创建指南](https://support.plex.tv/articles/200288926-creating-libraries/)：了解媒体类型和目录选择。
- [Plex 本地字幕指南](https://support.plex.tv/articles/200471133-adding-local-subtitles-to-your-media/)：了解字幕命名、语言代码和支持格式。
- [Plex 硬件加速指南](https://support.plex.tv/articles/115002178853-using-hardware-accelerated-streaming/)：了解 Plex Pass 要求与硬件转码状态。
