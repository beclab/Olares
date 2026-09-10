---
outline: [2, 3]
description: 从 Cursor、Claude Code 等 AI 运行时，或 Hermes Agent、OpenClaw 等 Olares 应用中安装和使用 Olares CLI 的 Agent Skills。涵盖各个技能、先装 shared 的顺序，以及完整用法。
head:
  - - meta
    - name: keywords
      content: Olares, olares-cli, Agent Skills, AI Agent, Cursor, Claude Code, Hermes Agent, OpenClaw, 自然语言管理
---

# 安装与使用 Agent Skills

`olares-cli` 的用户模式和集群内模式主要是给 AI Agent 用的，而不是让人逐条敲命令。为此，`olares-cli` 附带了一组 Agent Skills，每组命令对应一个。每个技能会告诉 Agent 各命令的作用、哪些参数重要、如何鉴权，以及遇到常见错误怎么处理。

## 了解 Agent Skills

每个技能就是一份 `SKILL.md`，AI 运行时会把它当作工具定义加载进来。当 Agent 收到“列出我 Olares Home 文件夹里的文件”这类请求时，它会对照已加载的技能，找到对应的命令（`olares-cli files ls /drive/Home`），然后代你运行。

这些技能都放在 Olares 仓库的 [`cli/skills/`](https://github.com/beclab/Olares/tree/main/cli/skills) 目录下，每个技能由一份 `SKILL.md` 和一个 `references/` 文件夹组成。`SKILL.md` 负责把请求引导到正确的命令，并讲清楚通用概念和常见错误怎么处理。`references/` 则给每个较复杂的子命令单独准备一个文件，用来补充那些 `--help` 里查不到、又不方便塞进 `SKILL.md` 的细节。

## 可用的 Agent Skills

| Skill | 说明 |
|-------|------|
| `olares-shared` | 登录 Olares、管理账号和 token、处理鉴权失败。使用其他技能前必须先加载 `olares-shared`。 |
| `olares-chart` | 把自己的项目、docker-compose 或 Helm chart 转成 Olares 应用并部署。 |
| `olares-files` | 管理 Olares 文件，支持上传、下载、压缩、解压、分享、挂载 SMB 和 NFS。支持 `drive`、`sync`、`cache`、`external` 等路径。 |
| `olares-market` | 安装、升级、卸载、克隆、停止、恢复、重启 Olares 应用，也可以浏览应用目录、查看状态、上传本地 chart。 |
| `olares-settings` | 修改 Olares 设置，包括用户、应用、VPN、网络、备份恢复、集成账号、GPU、搜索等。 |
| `olares-dashboard` | 查看系统资源使用情况，包括 CPU、内存、磁盘、网络、Pod、GPU、风扇和应用资源排行。 |
| `olares-cluster` | 查看 K8s 运行状态，包括 Pod、容器、工作负载、Job、CronJob、节点和中间件。可以查日志、进入容器、扩缩容、重启、暂停和恢复 CronJob。 |
| `olares-doctor` | 排查应用运行问题，比如安装卡住、崩溃、镜像拉取失败、状态为运行但无法访问、运行缓慢等。会自动调用 `cluster`、`dashboard`、`market` 收集信息。 |
| `olares-router` | 配置和使用 AI 模型：接入云端厂商的模型、安装本地模型应用、设置默认模型、签发 key 和配额、查看用量，也可以直接调用模型做对话、向量、语音转写、语音合成和 OCR。 |
| `olares-knowledge` | 往 Olares 下载内容：新建、查看、暂停、取消 URL、yt-dlp、aria2、种子和 Hugging Face 任务，也包括 Wise 发起的下载。 |
| `olares-search` | 搜索文件和应用。支持在 `drive` 和 `sync` 中全文搜索，也可以按标题查找已安装应用。 |
| `olares-publish` | 把本机已经跑起来的应用发布到公开的 Olares Market：发布目标、上架信息、图标和截图。 |

:::warning 务必先安装 `olares-shared`
其他所有技能都默认 `olares-shared` 已经加载。它定义了 profile 模型、令牌刷新逻辑，以及其他技能依赖的鉴权错误恢复提示。比如只加载了 `olares-files` 的 Agent，遇到鉴权错误时就无从恢复。
:::

## 手动安装技能

如果你是用 `npx @olares/cli@latest install` 安装的 CLI，这些技能已经一并装好，可以跳过本节。如果是单独安装的 CLI，让 CLI 自己把技能写出来：

```bash
olares-cli skills install
```

这些技能是编译进二进制的，所以这条命令不联网、也不会去挑版本：落到磁盘上的，就是你刚运行的这个 `olares-cli` 编译时用的那一份。技能会先在 `~/.agents/skills` 放一份，然后链接到本机上已经存在的各个 Agent 技能目录——Cursor、Claude Code、Codex、OpenClaw 等；没装的 Agent 不会被凭空建出目录来。之后只要你提到 Olares 相关的任务，Agent 就会自动加载对应的技能；`olares-shared` 会和其他技能一起写入，“先装 shared”的要求自然就满足了。

如果这台机器上还没有 `olares-cli`，请[先安装 CLI](./cli-install.md)——技能随它一起来。

:::tip
这些技能在 ClawHub 上也有，但那份副本是技能进二进制之前的，已经不再更新。请改用 CLI 安装：从注册表拿到的副本无法告诉你它和你手上二进制的命令是否已经对不上。
:::

Olares 上的一些 AI Agent 应用已经内置了这些技能，开箱即可让 Agent 管理 Olares。要在这类应用中使用这些技能，请参考[用 Hermes Agent 管理 Olares](../use-cases/hermes.md#manage-olares-with-your-hermes-agent)或[用 OpenClaw 管理 Olares](../use-cases/openclaw-olares-skills.md)。

## 更新技能

`olares-cli` 和 Agent Skills 会持续更新。有新版本时，请根据你的安装方式选择对应的更新方法。

### 更新 Olares Agent 应用内置的技能

OpenCode、Hermes Agent 等 Olares AI Agent 应用内置了 Olares CLI Agent Skills。更新应用的同时也会更新内置技能，无需额外运行 CLI 命令。

### 更新本机安装的技能

先把 `olares-cli` [更新到最新版](./cli-install.md#更新-olares-cli)，然后再运行一次安装命令：

```bash
olares-cli skills install
```

技能来自二进制，所以顺序很重要：更新后的 CLI 会写出与它自己对应的技能，每个技能整份替换。只更新 CLI 不会带上技能，因此在两者对不上期间，除 `skills` 以外的每条命令都会在标准错误上提示一行。确实不能有额外输出的场合，可以用 `OLARES_CLI_NO_SKILL_NOTICE=1` 关掉这行提示。

:::tip
你也可以直接让 AI Agent 代你运行这些更新命令。
:::

## 把 Olares CLI 当作 Agent Skills 使用

加载这些技能后，就能用自然语言操作 Olares，由 Agent 决定运行哪条 CLI 命令。例如：

```plain
# 通过 olares-files 技能列出文件
列出我 Olares 设备上 Home 文件夹里的文件

# 通过 olares-market 技能安装应用
从应用市场安装 Firefox，装好后告诉我

# 通过 olares-dashboard 技能查看资源占用
告诉我哪些应用占用了超过 1 GB 内存
```

:::tip
如果 Agent 没有自动加载 Olares 技能，可以手动用斜杠命令（`/`）调用。
:::
