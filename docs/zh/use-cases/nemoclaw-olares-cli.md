---
outline: [2, 3]
description: 在 NemoClaw 中使用 Olares CLI 技能，通过自然语言管理 Olares 设备上的文件和应用。
head:
  - - meta
    - name: keywords
      content: Olares, NemoClaw, Olares CLI, ClawHub, skills, 自然语言, 文件管理, 应用安装
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-05-11"
---

# 使用 Olares CLI 管理 Olares

Olares CLI 技能可以让 NemoClaw Agent 通过自然语言管理 Olares 设备上的文件和应用。如需了解其他技能，请参阅[管理技能和插件](openclaw-skills.md)。

## 准备工作

- NemoClaw 已在 Olares 上安装并运行。
- 你的 Olares ID 和登录密码。

## 步骤 1：登录 Olares CLI

Olares CLI 需要你的账号密码。Agent 使用 Olares CLI 技能前，需要先在 NemoClaw CLI 中登录。

1. 从启动台打开 NemoClaw CLI 应用。
2. 连接到运行时沙盒：

   ```bash
   nemoclaw my-assistant connect
   ```

   等待终端显示沙盒提示符，例如 `sandbox@my-assistant:~$`。

3. 登录你的 Olares 账号。将 `<your-olares-id>` 替换为你的 Olares ID：

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

   例如：

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

   按提示输入你的 Olares 登录密码。

4. 检查 profile 是否已登录。

   ```bash
   olares-cli profile list
   ```

   示例输出：

   ```text
   NAME                   OLARES-ID              STATUS
   *  laresprime@olares.com  laresprime@olares.com   logged-in (23h59m)
   ```

## 步骤 2：从 ClawHub 安装 Olares 技能

1. 打开 OpenClaw Web UI，并进入 **Skills**。
2. 在 ClawHub 搜索框中输入 `olares`，找到 Olares 技能。

   ![ClawHub 中的 Olares 技能](/images/manual/use-cases/nemoclaw-install-olares-skills.png#bordered)

3. 先安装 **Olares Shared** 技能，这是其他 Olares 技能的基础。
4. 安装其余 Olares 技能，例如 **Olares Files** 和 **Olares Market**。
5. 打开 OpenClaw Web UI 的聊天页面，运行 `/reset` 开启新会话，让 Agent 识别新安装的技能。如果你已配置 Discord 等频道，也需要在每个频道会话中运行 `/reset`。

:::info 遇到 429 错误时重试
如果下载技能时看到 429 错误，稍等片刻后重试。
:::

## 步骤 3：用自然语言与 Agent 对话

打开 OpenClaw Web UI 或 OpenClaw TUI，并用自然语言向 Agent 提出请求。例如：

- 列出 `/drive/Home/` 下的所有文件和文件夹：

  ```text
  List drive/Home/
  ```

  ![列出文件的结果](/images/manual/use-cases/nemoclaw-openclaw-olares-cli-list-files.png#bordered)

- 读取文件：

  ```text
  Read the last 10 lines of the nemoclaw.log file in the Home directory.
  ```

  ![读取文件的结果](/images/manual/use-cases/nemoclaw-openclaw-olares-cli-read-file.png#bordered)

- 从 Olares 应用市场安装应用：

  ```text
  Install Firefox
  ```

- 从 Olares 应用市场卸载应用：

  ```text
  Uninstall Firefox
  ```

## 了解更多

- [使用本地 LLM 运行 NemoClaw](nemoclaw.md)：使用本地模型设置 NemoClaw。
- [管理技能和插件](openclaw-skills.md)：安装和管理其他 OpenClaw 技能。
