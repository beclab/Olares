---
outline: [2, 3]
description: 在 OpenClaw 中使用 Olares CLI 技能，让你的智能体能够管理 Olares 设备上的文件和应用。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, Olares CLI 技能, ClawHub, 智能体
app_version: "1.0.17"
doc_version: "1.1"
doc_updated: "2026-07-31"
---

# 使用 OpenClaw 智能体管理 Olares

OpenClaw 内置 Olares CLI [Agent Skills](/zh/developer/cli-agent-skills.md)，因此你的智能体可以开箱即用地管理 Olares 设备上的文件和应用。例如，让它列出文件、读取日志，或从 Olares 应用市场安装应用。

## 学习目标

在本指南中，你将学习如何：
- 使用 Olares ID 对 Olares CLI 进行身份验证。
- 使用自然语言与智能体对话，让它在 Olares 设备上执行任务，例如从应用市场安装应用。

## 前提条件

- 你的 Olares 设备上已安装并运行 OpenClaw。
- 你的 Olares ID 和登录密码。

## 步骤 1：使用 Olares CLI 进行身份验证

在让智能体代你运行 Olares CLI Agent Skills 之前，先使用 Olares ID 对 Olares CLI 进行身份验证。

1. 从启动台打开 OpenClaw CLI。
2. 运行以下命令，确认 `olares-cli` 及其技能已正确安装并启用：

   ```bash
   olares-cli -v
   ```

   示例输出：

   ```
   olares-cli version 1.12.6
   Git commit: d30eca705df2fb614bf2bbea95daa2e6998adeeb
   Build time: 2026-07-06T06:33:00Z
   ```

3. 运行以下命令登录你的 Olares 账号。将 `<your-olares-id>` 替换为你的实际 Olares ID。

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

   例如：

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

4. 按提示输入 Olares 登录密码。输入时密码会被隐藏。
5. 如果你的 Olares 启用了双因素认证，CLI 会提示你输入该 Olares ID 的双因素认证码。输入 LarePass 中的 6 位验证码，然后按 **Enter**。
6. 运行以下命令，验证配置文件已创建并已登录：

   ```bash
   olares-cli profile list
   ```

   示例输出（`*` 表示当前配置文件）：

   ```text
      NAME                   OLARES-ID              STATUS     VERSION
   *  laresprime@olares.com  laresprime@olares.com  logged-in  1.12.6
   ```

## 步骤 2：让智能体执行任务

打开 Control UI，开启新会话，然后使用自然语言向智能体发送任务请求。

例如，让它从 Olares 应用市场安装应用：

```text
Install Firefox from Olares Market and tell me when it is ready
```

## 了解更多

- [管理技能和插件](openclaw-skills.md)：安装和管理其他 OpenClaw 技能。
