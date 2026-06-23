---
outline: [2, 3]
description: Olares 上 OpenCode 的常见问题及解决方案，包括终端可用性、CDN 错误、包管理和提供方配置。
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, troubleshooting, common issues, self-hosted
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/opencode-issues.md)。
:::

# OpenCode 常见问题

## ARM64 设备上终端无法接受输入

终端面板加载但不响应键盘输入。

这是因为终端依赖 `bun-pty`，它需要 glibc。OpenCode 容器基于 Alpine Linux（musl libc），因此 `bun-pty` 无法在 ARM64 设备上加载原生 PTY 库。

**解决方法**：改用 Control Hub 中的 OpenCode 容器终端。

## 页面空白并显示 CDN 错误

页面突然变为空白，浏览器控制台显示 `TypeError: ot(...) is not a function`。

OpenCode 在运行时从 CDN (`app.opencode.ai`) 加载前端资源，而不是在本地打包。当 CDN 部署了损坏的前端版本时，所有用户都会受到影响。

要解决此问题：

1. 使用 **Ctrl+Shift+R** 或 **Cmd+Shift+R** 强制刷新浏览器。
2. 如果问题仍然存在，等待 CDN 部署热修复。
3. 检查[上游仓库](https://github.com/sst/opencode)以了解修复进度。

## Plan 模式意外失败

Plan 模式与主 UI 从相同的 CDN 加载前端资源。CDN 端的错误可能导致 Plan 模式失败，即使你的本地安装是最新的。

要解决此问题：

1. 强制刷新浏览器。
2. 检查[上游仓库](https://github.com/sst/opencode/issues)以了解已知问题或更新的版本。

## 某些包安装失败

OpenCode 容器在 Alpine Linux 上运行，它使用 musl libc 而不是 glibc。需要基于 glibc 环境的包，例如某些 Node.js 原生模块或带有 C 扩展的 Python 包，可能无法正确安装或运行。

尽可能使用 `pkg-install` 安装系统级包。有关详细信息，请参阅 [管理包](opencode-packages.md)。

## 无法在 Models 面板中编辑提供方详情

Models 面板仅支持切换已连接模型的开启或关闭。你无法直接从此面板编辑提供方配置。

要更新提供方的设置：

- 断开提供方连接，然后使用更新后的详情重新连接。
- 或者，直接编辑配置文件并重启 OpenCode。有关说明，请参阅 [编辑配置文件](opencode.md#edit-the-config-file)。

## WebFetch 间歇性失败

WebFetch 工具可能偶尔无法从某些 URL 检索内容。这些失败通常是由目标网站的可用性或网络状况引起的，而不是 OpenCode 本身。

如果 WebFetch 反复失败：

1. 稍等片刻后重试请求。
2. 检查目标 URL 是否可以从你的浏览器访问。
3. 如果目标网站暂时不可用，请尝试不同的 URL。

## 当 `opencode.json` 已存在时无法编辑 `opencode.jsonc`

在 `~/.config/opencode/` 下，你可能会看到两个配置文件：

- `opencode.jsonc`：由 OpenCode 在你首次从 UI 添加自定义提供方时自动创建。
- `opencode.json`：由 Olares 构建添加，用于需要频繁手动编辑的插件密集型设置。

OpenCode 在运行时合并两个文件，因此同时拥有两者是正常的。

要从 Olares Files 编辑 `opencode.jsonc`，你通常将其重命名为 `opencode.json`，以便 Files 在 JSON 编辑器中打开它。但当 `opencode.json` 已存在时，重命名会失败，因为该名称的文件已存在。

**解决方法**：

1. 将 `opencode.jsonc` 重命名为临时 `.json` 名称，例如 `opencode1.json`。
2. 在 Files 中打开 `opencode1.json`，进行编辑并保存。
3. 将 `opencode1.json` 重命名回 `opencode.jsonc`。
