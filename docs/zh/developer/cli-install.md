---
outline: [2, 3]
description: 安装并登录 Olares CLI，然后通过 Agent Skills 让 AI Agent 管理 Olares。
head:
  - - meta
    - name: keywords
      content: Olares, Olares CLI, AI Agent, Agent Skills, 安装 Olares CLI, Claude Code, Codex, Cursor
---

# 通过 AI Agent 使用 Olares CLI

本指南介绍如何在电脑上安装 Olares CLI 及配套的 Agent Skills（技能）、登录 Olares，并让 AI Agent 完成第一个 Olares 任务。

:::info 使用运行在 Olares 上的 AI Agent？
运行在 Olares 上的 [Lares](../use-cases/lares.md)、[Hermes Agent](../use-cases/hermes.md) 和 [OpenClaw](../use-cases/openclaw.md) 已经内置 Olares CLI 和技能，无需在应用内重复安装。请直接按照所用应用的指南操作。
:::

:::warning Olares 主机已经自带 Olares CLI
如果直接在运行 Olares OS 的 Linux 主机上操作，请勿覆盖 `/usr/local/bin/olares-cli`。参阅[在 Olares 主机上使用 Olares CLI](#在-olares-主机上使用-olares-cli)。
:::

## 准备工作

开始前，请确认：

- 电脑运行 macOS、Windows 或 Linux，并已安装 Node.js 18 或更高版本。
- Olares 已激活，并且你知道自己的 Olares ID。
- 电脑上装有支持 Agent Skills 的 AI Agent，例如 Codex、Claude Code、Cursor 或 OpenClaw。

## 安装 Olares CLI 和 Agent Skills

打开终端，运行：

```bash
npx @olares/cli@latest install
```

安装向导会安装相同版本的 Olares CLI 和技能。技能会写入 `~/.agents/skills`，并链接到电脑上已有的兼容 AI Agent 技能目录。

输出示例：

```text
┌  Setting up Olares CLI...
│
◇  Installed globally
│
◇  Skills installed
│
└  You are all set!
```

查看已安装的 CLI 版本：

```bash
olares-cli version
```

:::info
安装向导只会在电脑上安装 Olares CLI，不会安装 Olares OS，也不会自动登录 Olares 账号。
:::

## 登录 Olares

登录后，CLI 会创建一个 `profile`（连接配置），记录要连接哪台 Olares，以及以哪个用户身份操作。

1. 使用 Olares ID 开始交互式登录。例如：

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

   请将 `laresprime@olares.com` 替换为你的 Olares ID。

2. 根据提示输入 Olares 密码。终端不会显示密码内容。

3. 如果开启了两步验证，请输入 LarePass 中的六位验证码。

4. 检查当前 profile：

   ```bash
   olares-cli profile whoami
   ```

登录后，Olares CLI 会自动刷新令牌。

## 让 AI Agent 完成第一个任务

使用安装 Olares CLI 时的同一个操作系统账号打开 AI Agent。

在 Agent 对话窗口中，检查它能否找到 Olares 技能：

- 如果 Agent 支持通过斜杠命令列出技能，输入 `/`，查找以 `olares-` 开头的条目。
- 也可以直接询问 Agent：

  ```plain
  列出当前可用的 Olares 技能。
  ```

确认列表中包含 `olares-files`、`olares-market` 等技能。如果找不到，请重新运行 `olares-cli skills install`，检查输出中的安装位置，然后新建一个 Agent 对话。

接着，让 Agent 完成第一个任务：

```plain
列出我 Olares 设备上 Home 文件夹里的文件。
```

Agent 会选择合适的技能并运行 Olares CLI 命令。确认它返回了 Olares Home 文件夹的内容，说明 Agent 已能通过 CLI 访问你的 Olares。

还可以尝试：

```plain
从 Olares 应用市场安装 Firefox，装好后告诉我。

告诉我 Olares 上哪些应用占用了超过 1 GB 内存。

在 Olares 文件中查找名为 quarterly-report 的文档。
```

运行命令或修改 Olares 前，AI Agent 可能会请求批准。批准前请先确认它准备执行的操作。

## 管理 profile

一份 Olares CLI 可以保存多台 Olares 或多个身份的 profile。

| 操作 | 命令 |
|------|------|
| 列出所有 profile | `olares-cli profile list` |
| 查看当前身份 | `olares-cli profile whoami` |
| 切换 profile | `olares-cli profile use <name>` |
| 切回上一个 profile | `olares-cli profile use -` |
| 删除 profile 及保存的令牌 | `olares-cli profile remove <name>` |

Olares CLI 会按照当前操作系统的安全机制保存令牌。需要清除登录信息时，请使用 `profile remove`，不要手动修改凭据文件。

## 更新 Olares CLI 和 Agent Skills

只检查是否有新版本，不做修改：

```bash
olares-cli update --check
```

更新 Olares CLI，并同时安装与新版 CLI 匹配的技能：

```bash
olares-cli update
```

如果需要接收所有已发布版本，包括尚未标记为 `latest` 的版本，请使用 `next` 通道：

```bash
olares-cli update --channel next
```

更新后检查当前副本和版本：

```bash
olares-cli version
```

## 其他安装方式

### 先安装 Olares CLI

如果不使用安装向导，可以先单独安装 CLI：

```bash
npm install -g @olares/cli
```

再安装该版本 CLI 自带的技能：

```bash
olares-cli skills install
```

### 免安装运行单条命令

通过 `npx` 运行单条命令，不把 Olares CLI 加入 `PATH`：

```bash
npx @olares/cli@latest files ls /drive/Home
```

profile 和登录状态仍会保留，之后可以继续使用。

## 在 Olares 主机上使用 Olares CLI

Olares OS 会在 `/usr/local/bin/olares-cli` 中预置由系统管理的 CLI。它用于维护 Olares OS，并随 Olares OS 版本更新。请勿用 npm 软件包覆盖它。

在 Olares 主机上运行安装向导时，向导会识别现有的正式版 CLI 并保留它。如果还需要独立的 npm 版本，请安装到另一个目录：

```bash
npm install -g @olares/cli@latest --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"
olares-cli skills install
```

在当前终端中，`olares-cli` 会优先运行 npm 版本。维护 Olares OS 时，请使用完整路径 `/usr/local/bin/olares-cli` 调用系统自带版本。

:::warning 不要覆盖系统自带的 CLI
请勿在 Olares 主机上运行 `npm install -g @olares/cli --force`。该命令可能覆盖由系统管理的 CLI，破坏它与 `olaresd` 和集群之间的版本关系。
:::

## 了解更多

- [Olares CLI Agent Skills](./cli-agent-skills.md)：查看完整技能清单，了解技能的安装位置、发现方式和版本同步机制。
- [使用 Lares 管理 Olares 并开展研究](../use-cases/lares.md)：直接使用内置 Olares CLI 和技能的官方 Olares AI 助手。
- [Olares、Olares CLI 与 AI Agent 之间的关系](https://www.olares.com/blog/olares-cli-ai-agents-explained/)：了解 Agent 可以在哪里运行，以及 profile 如何把 Agent 连接到 Olares。
