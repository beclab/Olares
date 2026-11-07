---
outline: [2, 3]
description: 为 Olares 上的 YuE2 Studio MCP 入口启用 Bearer 令牌认证，并将令牌添加到 Lares 或其他 MCP 客户端。
head:
  - - meta
    - name: keywords
      content: Olares, YuE2 Studio, MCP, Bearer 令牌, 认证, Lares
app_version: "0.0.7"
doc_version: "1.0"
doc_updated: "2026-11-07"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/yue2-studio-mcp-auth.md)为准。
:::

# 使用令牌保护 YuE2 Studio MCP

YuE2 Studio 默认允许客户端连接内部 MCP 入口，无需提供应用专用令牌。你可以启用 Bearer 令牌认证，让包括 Lares 在内的所有客户端在调用 YuE2 Studio 工具前完成认证。

本指南将启用令牌要求，并更新已有的 Lares 连接。其他使用 **YuE2 Studio MCP & API** 入口的客户端也使用同一个令牌。

## 学习目标

通过本指南，你将学会：

- 为 YuE2 Studio MCP 入口启用令牌认证。
- 找到应用生成的令牌。
- 将令牌添加到 Lares 或其他 MCP 客户端。
- 恢复无需令牌的连接方式。

## 前提条件

- 已[在 Olares 上安装 YuE2 Studio](yue2-studio.md#安装-yue2-studio)。
- 已[通过 MCP 将 YuE2 Studio 连接到 Lares](yue2-studio-lares.md#将-yue2-studio-连接到-lares)。

## 启用令牌认证

1. 打开 Olares **Settings** > **Applications** > **YuE2 Studio**。
2. 在 **Environment variables** 下点击 **Manage environment variables**。
3. 将 `MCP_AUTH_ENABLED` 设为 `true`。
4. 点击 **Confirm**，再点击 **Apply**。等待 YuE2 Studio 重启。
5. 打开 Files，前往 **Application** > **Data** > **yue2studio** > **config**。
6. 将 `mcp-token` 作为文本打开，复制完整的一行。如果 Files 无法预览文本，请下载文件并用文本编辑器打开。

不要把令牌放入聊天提示词或共享截图。

## 将令牌添加到 Lares

1. 打开 Lares，进入 **Settings** > **MCP**。
2. 编辑 `yue2studio` 服务。
3. 将 **Headers** 替换为下方的 JSON。把 `<your-token>` 替换为 `mcp-token` 中完整的一行，并在 `Bearer` 后保留一个空格。

   ```json
   {
     "Authorization": "Bearer <your-token>"
   }
   ```

4. 点击 **Save**。
5. 使用 Qwen3.8-27B (llama.cpp) 新建对话，并[检查音乐引擎](yue2-studio-lares.md#检查音乐引擎)。

令牌认证对使用该入口的所有客户端生效。在其他 MCP 客户端中，请将同一个 `Authorization` 请求头添加到 Streamable HTTP 连接。

## 关闭令牌认证

1. 在 Olares Settings 中打开 YuE2 Studio 的应用环境变量。
2. 将 `MCP_AUTH_ENABLED` 设为 `false`，再点击 **Confirm** 和 **Apply**。
3. 等待 YuE2 Studio 重启。
4. 在 Lares 中编辑 `yue2studio` 服务，将 **Headers** 设为 `{}`。

## 排查认证问题

### 为什么客户端返回 HTTP 401？

YuE2 Studio 入口没有收到预期的令牌。请复制 `mcp-token` 中完整的一行，并确认请求头以 `Bearer` 开头，后面保留一个空格。

### 为什么找不到 `mcp-token`？

应用 `MCP_AUTH_ENABLED=true` 后，等待 YuE2 Studio 完成重启，再次打开 `config` 文件夹。

## 了解更多

- [使用 YuE2 Studio 创作并编辑歌曲](yue2-studio.md)：安装 Studio、下载模型套装并在应用中创作音乐。
- [通过 Lares 使用 YuE2 Studio 创作歌曲](yue2-studio-lares.md)：连接 MCP 入口，并在对话中生成歌曲。
- [管理应用环境变量](../manual/olares/settings/manage-app-env.md)：修改应用设置并应用重启。
