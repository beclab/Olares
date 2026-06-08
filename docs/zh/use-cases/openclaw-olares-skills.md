---
outline: [2, 3]
description: 在 OpenClaw 中安装 Olares CLI 技能，让你的智能体能够管理 Olares 设备上的文件和应用。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw 智能体, Olares CLI 技能
app_version: "1.0.3"
doc_version: "1.0"
doc_updated: "2026-06-09"
---

# 使用 OpenClaw 智能体管理 Olares

在 OpenClaw 中安装 Olares CLI 技能，让你的智能体能够管理 Olares 设备上的文件和应用。例如，你可以让它列出文件、读取日志，或从 Olares 应用市场安装应用。

## 学习目标

通过本指南，你将学会：
- 在 OpenClaw CLI 中使用 Olares CLI 进行身份验证。
- 安装 Olares 技能。
- 使用自然语言与智能体对话，让它在 Olares 设备上执行任务。

## 前提条件

- 你的 Olares 设备上已安装并运行 OpenClaw。
- 你的 Olares ID 和登录密码。

## 步骤 1：使用 Olares CLI 进行身份验证

要让智能体获得执行系统操作的权限，你需要先使用账号凭据登录 Olares CLI。

1. 从桌面打开 OpenClaw CLI。
2. 运行以下命令，登录你的 Olares 账号。将 `<your-olares-id>` 替换为你的 Olares ID：

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

   例如：

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

3. 按照提示输入 Olares 登录密码。注意：出于安全考虑，你输入的密码不会显示。
4. 运行以下命令，验证登录状态。

   ```bash
   olares-cli profile list
   ```

   示例输出：

   ```text
   NAME                   OLARES-ID              STATUS
   *  laresprime@olares.com  laresprime@olares.com   logged-in
   ```

## 步骤 2：安装 Olares 技能

从 ClawHub 安装 Olares 技能，让智能体获得管理设备的能力。

1. 打开 Control UI，从左侧边栏选择 **Skills**。
2. 在 **ClawHub** 下的搜索框中输入 `olares`，查找 Olares 技能。

   ![ClawHub 中的 Olares 技能](/images/manual/use-cases/openclaw-install-olares-skills1.png#bordered)

3. 先安装 **Olares Shared** 技能，因为它是其他 Olares 技能的基础。
4. 安装其余的 Olares 技能，例如 **Olares Files** 和 **Olares Market** 等。
5. 进入聊天页面，运行 `/reset` 命令开启新会话，使智能体加载新安装的技能。如果你配置了 Discord 等其他聊天渠道，也需要在每个渠道的对话中运行 `/reset` 命令。

:::info 遇到 429 错误时重试
如果下载技能时出现 `429` 错误，请稍等片刻后重试。
:::

## 步骤 3：让智能体执行任务

打开 Control UI，使用自然语言向智能体发送任务指令。

例如，让它从 Olares 应用市场安装应用：

```text
Install Firefox
```

![安装应用](/images/manual/use-cases/openclaw-olares-cli-install-app.png#bordered)

## 了解更多

- [管理技能和插件](openclaw-skills.md)：安装和管理其他 OpenClaw 技能。
