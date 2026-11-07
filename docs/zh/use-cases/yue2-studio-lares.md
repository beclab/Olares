---
outline: [2, 3]
description: 通过 MCP 将 YuE2 Studio 连接到 Lares，把音乐想法变成歌曲，并在 Lares 或 YuE2 Studio 中播放结果。
head:
  - - meta
    - name: keywords
      content: Olares, YuE2 Studio, Lares, MCP, AI 音乐, 歌曲生成, 本地 AI
app_version: "0.0.7"
doc_version: "1.5"
doc_updated: "2026-11-07"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/yue2-studio-lares.md)为准。
:::

# 通过 Lares 使用 YuE2 Studio 创作歌曲

YuE2 Studio 提供 MCP 服务，让 Lares 可以直接在对话中把音乐想法变成歌曲。生成后，你可以在 Lares 中播放，也可以在 YuE2 Studio 中打开。

本指南通过一次 Lares 对话创作一首简短的原声流行歌曲。你可以将示例描述替换成自己的音乐构思。

## 学习目标

通过本指南，你将学会：

- 将 YuE2 Studio MCP 服务连接到 Lares。
- 在对话中描述音乐想法并生成歌曲。
- 在 Lares 或 YuE2 Studio 中播放结果。

## 前提条件

- 已[安装 YuE2 Studio 并下载一个模型套装](yue2-studio.md#下载模型套装)。保持应用运行，并等待显示 **Engine ready**。
- 已在同一台 Olares 设备上[为 Lares 准备 Qwen3.8-27B (llama.cpp)](lares.md)。
- 有足够的加速器内存同时运行对话模型和 YuE2 Studio。两者使用同一张 GPU 时，一次只运行一个音乐生成请求。

## 将 YuE2 Studio 连接到 Lares

将 YuE2 Studio 的 MCP 端点添加到 Lares，让 Qwen3.8-27B 可以调用音乐创作工具，并在对话中生成歌曲。

1. 打开 Olares **Settings** > **Applications** > **YuE2 Studio** > **Entrances**。
2. 选择 **YuE2 Studio MCP & API**，复制 **Endpoint** 地址。
3. 打开 Lares，进入 **Settings** > **MCP**。
4. 点击 **Add server**，填写以下设置：
    - **Server name:** `yue2studio`
    - **Transport:** `Streamable HTTP`
    - **MCP URL:** 粘贴刚才复制的地址，并在末尾添加 `/mcp`。
    - **Headers:** `{}`
5. 点击 **Save**。Lares 会连接服务并载入工具。

## 生成歌曲

1. 在 Lares 中新建对话，选择 **Qwen3.8-27B (llama.cpp)**。
2. 用自然语言描述你的音乐想法。可以说明风格、情绪、主题、人声或乐器，例如：

   ```text
   创作一首关于海面晨光的温暖原声流行歌曲，使用温暖的女声、原声吉他、钢琴和轻柔的鼓点。
   ```

3. Lares 在调用 YuE2 Studio 工具前，可能会请求权限。查看请求并允许其继续。
4. Lares 生成歌曲时，让当前对话保持在最前面，等待 Lares 提示生成完成。
5. 选择一种方式试听结果：
    - 在当前 Lares 对话中，让 Lares 播放刚生成的歌曲。
    - 打开 YuE2 Studio，进入 **Library**，播放最新生成的歌曲。

## 排查连接问题

### 为什么 Lares 无法载入 YuE2 Studio 工具？

确认 YuE2 Studio 正在运行，传输方式为 **Streamable HTTP**，且地址以 `/mcp` 结尾。

## 了解更多

- [使用 YuE2 Studio 创作并编辑歌曲](yue2-studio.md)：在 Studio 界面中谱曲、编辑和导出音乐。
- [使用令牌保护 YuE2 Studio MCP](yue2-studio-mcp-auth.md)：要求 Lares 和其他 MCP 客户端提供 Bearer 令牌。
- [开始使用 Lares](lares.md)：准备 Qwen3.8-27B (llama.cpp)，并了解应用连接方式。
- [管理应用环境变量](../manual/olares/settings/manage-app-env.md)：修改应用设置并应用重启。
