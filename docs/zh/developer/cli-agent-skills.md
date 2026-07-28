---
outline: [2, 3]
description: 从 Cursor、Claude Code 等 AI 运行时，或 Hermes Agent、OpenClaw 等 Olares 应用中安装和使用 Olares CLI 的 Agent Skills。涵盖各个技能、先装 shared 的顺序，以及完整用法。
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
| `olares-search` | 搜索文件和应用。支持在 `drive` 和 `sync` 中全文搜索，也可以按标题查找已安装应用。 |

:::warning 务必先安装 `olares-shared`
其他所有技能都默认 `olares-shared` 已经加载。它定义了 profile 模型、令牌刷新逻辑，以及其他技能依赖的鉴权错误恢复提示。比如只加载了 `olares-files` 的 Agent，遇到鉴权错误时就无从恢复。
:::

## 手动安装技能

如果你是用 `npx @olares/cli@latest install` 安装的 CLI，这些技能已经一并装好，可以跳过本节。如果是单独安装的 CLI，运行下面的命令把这些技能一次装好：

```bash
npx skills add beclab/Olares -y -g
```

这会把技能装进你正在用的 Agent，比如 Cursor 或 Claude Code。之后只要你提到 Olares 相关的任务，Agent 就会自动加载对应的技能。由于这条命令会把 `olares-shared` 一起装上，“先装 shared”的要求自然就满足了。

:::tip
这些技能也发布在 ClawHub 上。两个渠道读取的是同一份 `SKILL.md`，所以装其中一个即可。如果你的 Agent 接入了 ClawHub，也可以从那里添加。
:::

Olares 上的一些 AI Agent 应用已经内置了这些技能，开箱即可让 Agent 管理 Olares。要在这类应用中使用这些技能，请参考[用 Hermes Agent 管理 Olares](../use-cases/hermes.md#manage-olares-with-your-hermes-agent)或[用 OpenClaw 管理 Olares](../use-cases/openclaw-olares-skills.md)。

## 更新技能

`olares-cli` 和 Agent Skills 会持续更新。有新版本时，请根据你的安装方式选择对应的更新方法。

### 更新 Olares Agent 应用内置的技能

OpenCode、Hermes Agent 等 Olares AI Agent 应用内置了 Olares CLI Agent Skills。更新应用的同时也会更新内置技能，无需额外运行 CLI 命令。

### 更新本机安装的技能

更新技能之前，请先将 `olares-cli` [更新到最新版](./cli-install.md#更新-olares-cli)。然后重新运行安装命令，用最新版本覆盖已安装的技能。

```bash
npx skills add beclab/Olares -y -g
```

该命令会从仓库拉取最新的 `SKILL.md`，并覆盖本机已安装的技能。

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
