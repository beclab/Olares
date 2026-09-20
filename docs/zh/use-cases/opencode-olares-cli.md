---
outline: [2, 3]
description: 在 OpenCode 中使用 Olares CLI 技能，让你的编码 Agent 通过自然语言管理 Olares 设备上的文件和应用。
head:
  - - meta
    - name: keywords
    - content: Olares, OpenCode, Olares CLI skills, AI Agent, 自然语言, 文件管理, 应用安装
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/opencode-olares-cli.md)为准。
:::

# 使用 Olares CLI 管理 Olares

OpenCode 内置 Olares CLI [Agent Skills](/developer/cli-agent-skills.md)，你的 Agent 可以通过自然语言管理 Olares 设备上的文件和应用。比如让它列出文件、读取日志，或从 Olares Market 安装应用。

## 准备工作

- OpenCode 已在 Olares 上安装并运行，且已连接模型。
- 你的 Olares ID 和登录密码。

## 步骤 1：使用 Olares ID 认证 Olares CLI

OpenCode 要代替你运行 Olares CLI Agent Skills，需要先用你的 Olares ID 认证 Olares CLI。

1. 从启动台打开 OpenCode Terminal，它会直接进入命令行界面。
2. 运行以下命令确认 `olares-cli` 已正确安装：

   ```bash
   olares-cli -v
   ```

   示例输出：

   ```text
   olares-cli version 1.12.6
   Git commit: d30eca705df2fb614bf2bbea95daa2e6998adeeb
   Build time: 2026-07-06T06:33:00Z
   ```

3. 运行以下命令登录你的 Olares 账号。将 `<your-olares-id>` 替换为你的 Olares ID。

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

   示例：

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

4. 根据提示输入 Olares 登录密码。输入时密码会被隐藏。
5. 如果你的 Olares 账号开启了两步认证，CLI 会提示你输入该 Olares ID 的两步验证码。输入 LarePass 中的 6 位验证码，然后按 **Enter**。
6. 运行以下命令验证 Profile 已创建且处于登录状态：

   ```bash
   olares-cli profile list
   ```

   示例输出（`*` 标记当前 Profile）：

   ```text
      NAME                   OLARES-ID              STATUS     VERSION
   *  laresprime@olares.com  laresprime@olares.com  logged-in  1.12.6
   ```

   :::info
   一次登录可以让 OpenCode 保持认证状态最多 30 天。到期后需要重新登录一次。
   :::

## 步骤 2：让 Agent 执行任务

1. 从启动台打开 OpenCode。
2. 开始一个新会话，并选择你之前连接的模型。
3. 用自然语言向 Agent 发送请求。比如让它从 Olares Market 安装应用：

   ```text
   Install Firefox from Olares Market and tell me when it is ready
   ```

   你也可以让它管理文件、读取日志、查看系统状态，and more。

## 了解更多

- [设置 OpenCode 作为 AI 编码 Agent](opencode.md)：安装 OpenCode 并连接模型。
- [Olares CLI 与 AI Agent，原理解析](https://www.olares.com/blog/olares-cli-ai-agents-explained)：Agent 如何通过 Olares CLI 管理你的系统。
