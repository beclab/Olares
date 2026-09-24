---
outline: [2, 3]
description: 在 Olares 上使用 Concat 和 Lares 制作短视频。通过 MCP 将图片剪辑成视频，并在文件管理器中找到导出的 MP4。
head:
  - - meta
    - name: keywords
      content: Olares, Concat, Lares, MCP, 视频剪辑, MP4, 自托管
app_version: "0.2.5"
doc_version: "1.0"
doc_updated: "2026-09-24"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/concat.md)为准。
:::

# 使用 Concat 和 Lares 制作视频

Concat 是一款支持多轨道、标题、转场和 MP4 导出的视频编辑器。在 Olares 上，你既可以在浏览器中手动剪辑，也可以通过模型上下文协议（MCP）让 AI 助手完成剪辑。

本示例将使用 Lares，把几张图片制作成一段 Olares 介绍视频。源素材、项目文件和导出的视频都会保存在文件管理器的 Concat 文件夹中。

## 前提条件

- 在 `amd64` 设备上运行 Olares 1.12.7 或更高版本。当前 Concat 应用包使用 CPU 渲染，不需要 GPU。
- 已配置 [Lares](lares.md)，且所选模型支持工具调用。
- 准备三到四张用于制作视频的图片。

## 安装 Concat

1. 打开 Market，搜索“Concat”。
   <!-- ![Market 中的 Concat](/images/manual/use-cases/concat.png#bordered) -->

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 准备素材

1. 打开文件管理器，前往 **Home** > **Documents** > **Concat**。
2. 创建 `assets/my-demo` 文件夹并上传图片。如果需要背景音乐，也把音频文件放入该文件夹。
3. 创建 `outputs/my-demo` 文件夹，用于保存完成的视频。

这些子文件夹只是建议的目录结构，安装 Concat 时不会自动创建。

| 文件管理器中的位置 | Concat MCP 使用的路径 |
|:---|:---|
| `Home/Documents/Concat/assets/my-demo` | `assets/my-demo` |
| `Home/Documents/Concat/outputs/my-demo` | `outputs/my-demo` |

浏览器编辑器会将 `Home/Documents/Concat` 识别为 `/config/Projects`。AI 助手所在电脑中的文件不会自动提供给 Concat，请先把素材上传到上述文件夹。

## 在 Lares 中配置 Concat

### 获取 MCP Endpoint

1. 打开 Settings，前往 **Applications** > **Concat** > **Entrances**。
2. 选择 **Concat API**，复制其中的 **Endpoint** URL。这是 MCP 入口，编辑器使用的是另一个入口。
3. 在复制的 URL 末尾添加 `/mcp`，注意不要重复添加斜杠。

MCP 入口默认仅限 Olares 内部访问。本示例需要使用同一台 Olares 上的 Lares。应用令牌不能替代 Olares 的网络访问要求。

### 获取应用令牌

应用令牌会在安装 Concat 时生成，它不是你的 Olares 密码。请使用有权访问该应用数据的账号，通过以下任一方法获取；你也可以向管理员索取。

<tabs>
<template #文件管理器>

1. 打开文件管理器，前往 **Application** > **Data** > **concat** > **config**，找到 Concat 的应用数据。
2. 以文本方式打开 `api-token` 文件。如果无法预览文本，请下载该文件并使用文本编辑器打开。
3. 复制完整的一行内容。这就是明文令牌，无需解码。

文件位于 `Data/concat/config/api-token`，而不是 `Home/Documents/Concat`。请勿编辑或重命名该文件。

</template>
<template #Control-Hub>

1. 打开 Control Hub，点击 **Browse**。
2. 展开 Concat 安装实例所在的命名空间，名称通常为 `concat-<username>`。
3. 展开 **Secrets**，选择 **concat-api**。
4. 在 **Data** 中找到 **token**。点击可见性按钮解码，然后复制完整的明文令牌。

请使用解码后的值，不要使用 `data.token` 中的 Base64 文本。Control Hub 的操作方法请参阅[查看保密字典](../manual/olares/controlhub/manage-resource.md#查看保密字典)。

</template>
</tabs>

请妥善保管令牌。把它粘贴到 MCP 配置中，不要发到聊天消息或包含在共享截图中。

### 添加 MCP 服务

1. 打开 Lares，进入 MCP 配置。
2. 按以下信息添加一个远程 HTTP MCP 服务：

   | 设置 | 值 |
   |:---|:---|
   | Name | `Concat` |
   | URL | 前面复制的 Endpoint，末尾加上 `/mcp` |
   | Header name | `Authorization` |
   | Header value | `Bearer <your-token>` |

   将 `<your-token>` 替换为刚才复制的令牌。`Bearer` 后保留一个空格。如果 Headers 字段接受 JSON，请输入：

   ```json
   {
     "Authorization": "Bearer <your-token>"
   }
   ```

3. 保存并连接该服务，然后刷新可用工具列表。
4. 新建一个对话，发送：

   ```text
   使用 Concat MCP 检查运行状态并读取帮助信息。
   列出 assets/my-demo 中的文件。暂时不要创建或修改任何项目。
   ```

Lares 应调用 `concat_status`、`concat_help` 和 `concat_list_files`，并返回已上传的图片。Concat 提供七个 MCP 工具。仅看到工具名称不代表调用一定成功。

:::tip 使用其他 AI 助手
兼容远程 HTTP MCP 的客户端也可以使用相同的 Endpoint 和 Authorization Header，包括 OpenCode。Olares 外部的客户端还需要具备网络访问权限，详见[通过本地地址访问 Olares 服务](../manual/best-practices/local-access.md)。该 Endpoint 用于处理 MCP POST 请求，不是旧版 SSE 连接或编辑器页面。
:::

## 制作短视频

1. 让 Lares 开始剪辑前，先保存浏览器编辑器中的所有更改。MCP 剪辑期间编辑器会暂时关闭，并在会话结束后恢复。
2. 发送以下提示词：

   ```text
   使用 Concat MCP，把 assets/my-demo 中的图片制作成一段精致的
   12 秒 Olares 介绍视频。视频使用横向 1080p 分辨率。

   整体风格保持清爽明亮。依次显示以下三个标题：
   “Your digital space”、“Create your way”和“Start with Olares”。
   使用缓慢、柔和的动态效果和转场，并在文字周围留出足够空间。
   仅当素材中有音频文件时，添加音量较低的背景音乐。

   先阅读工具帮助。创建一个名称唯一的新项目，不要修改已有项目。
   导出前检查几个预览帧，确认文字没有被裁切或相互重叠。

   从 concat_begin_edit 开始。将 MP4 导出到 outputs/my-demo，
   使用唯一的文件名。等待渲染完成后，再调用 concat_finish_edit。
   最后告诉我该视频在文件管理器中的完整位置。
   ```

3. 等待 Lares 确认导出成功。返回渲染任务 ID 只表示任务已提交，并不代表 MP4 已经生成。
4. 在文件管理器中打开 **Home** > **Documents** > **Concat** > **outputs** > **my-demo**，播放 Lares 返回的 MP4，检查标题、转场和声音。

如需继续调整，请说明要修改的内容，并要求使用新的输出文件名。Concat 不允许覆盖已有输出文件。最终效果取决于素材和所用模型；导出较长视频前，请先检查预览。

## 在浏览器中剪辑

也可以直接使用浏览器编辑器：

1. 从启动台打开 Concat，在 `/config/Projects` 下新建或打开项目。
2. 点击 **Import**，选择素材并在时间轴上排列片段。
3. 添加标题、转场和音频，然后检查预览。选择 **Full** 预览质量，以便检查小字号文字。
4. 点击 **Export**，选择输出位置和视频设置，等待状态变为 **Exported**。

![包含多轨道和视频预览的 Concat 编辑器](/images/manual/use-cases/concat-editor.png#bordered)

保存项目并不会导出 MP4。手动导出的文件会出现在你选择的位置，因此请选择 `/config/Projects` 下的文件夹，以便在文件管理器中找到它们。

## 常见问题

### 为什么 MCP 连接失败？

如果返回登录页面或 HTML 响应，通常表示请求被 Olares 入口或网络策略拦截。确认复制的是 **Concat API** Endpoint，且客户端可以访问该地址。Concat 返回 `401` 表示应用令牌缺失或无效；请检查明文令牌，并确认 `Bearer` 后有一个空格。

直接在浏览器中打开 `/mcp` 会发送 GET 请求，不能用于测试连接。请使用前面的只读提示词进行检查。

### 为什么 Lares 工作时编辑器会关闭？

Concat 会在浏览器剪辑和 MCP 剪辑之间切换。等待所有渲染任务完成，然后让 Lares 调用 `concat_finish_edit`。MCP 会话期间不要手动编辑。

### 为什么中文标题缺少部分字符？

让 AI 助手使用 `Noto Sans CJK SC` 字体，然后重新生成预览。Olares 中的 Concat 应用包已经包含该字体。

## 了解更多

- [Lares](lares.md)：配置本示例使用的 AI 助手。
- [OpenCode](opencode.md)：在 Olares 上使用编程助手。
