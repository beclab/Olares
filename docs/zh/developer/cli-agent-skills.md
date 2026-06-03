---
outline: [2, 3]
description: 从 Cursor、Claude Code 等 AI 运行时，或 Hermes Agent、OpenClaw 等 Olares 应用中安装和使用 Olares CLI 的 Agent Skills。涵盖各个技能、先装 shared 的顺序，以及完整用法。
---

# 安装与使用 Agent Skills

`olares-cli` 的用户模式和集群内模式主要是给 AI Agent 用的，而不是让人逐条敲命令。为此，`olares-cli` 附带了一组 Agent Skills，每组命令对应一个。每个技能会告诉 Agent 各命令的作用、哪些参数重要、如何鉴权，以及遇到常见错误怎么处理。

## 了解 Agent Skills

每个 Agent Skill 就是一份 `SKILL.md`，AI 运行时会把它当作工具定义加载进来。当 Agent 收到“列出我 Olares Home 文件夹里的文件”这类请求时，它会对照已加载的技能，找到对应的命令（`olares-cli files ls /drive/Home`），然后代你运行。

这些技能都放在 Olares 仓库的 [`cli/skills/`](https://github.com/beclab/Olares/tree/main/cli/skills) 目录下，每个技能由一份 `SKILL.md` 和一个 `references/` 文件夹组成。`SKILL.md` 负责把请求引导到正确的命令，并讲清楚通用概念和常见错误怎么处理。`references/` 则给每个较复杂的子命令单独准备一个文件，用来补充那些 `--help` 里查不到、又不方便塞进 `SKILL.md` 的细节。

## 可用的 Agent Skills

| Skill | 说明 |
|-------|------|
| `olares-shared` | profile 模型、登录流程、令牌存储、自动刷新和鉴权错误恢复。<br>其他所有技能的基础。 |
| `olares-files` | 列出、上传、下载、编辑、分享、挂载 SMB，以及管理 Sync 仓库。 |
| `olares-market` | 浏览、安装、升级、卸载，以及上传本地 chart。 |
| `olares-settings` | 读取和修改网页端开放的设置。 |
| `olares-dashboard` | 总览和应用指标，JSON 结构稳定。 |
| `olares-cluster` | 读取和修改 Pod、工作负载、节点、Job、CronJob，以及中间件密码。 |

:::warning 务必先安装 `olares-shared`
其他所有技能都默认 `olares-shared` 已经加载。它定义了 profile 模型、令牌刷新逻辑，以及其他技能依赖的鉴权错误恢复提示。比如只加载了 `olares-files` 的 Agent，遇到鉴权错误时就无从恢复。
:::

## 手动安装技能

如果你是用 `npx @olares/cli@latest install` 安装的 CLI，这些技能已经一并装好，可以跳过本节。如果是单独安装的 CLI，运行下面的命令把六个技能一次装好：

```bash
npx skills add beclab/Olares -y -g
```

这会把技能装进你正在用的 Agent，比如 Cursor 或 Claude Code。之后只要你提到 Olares 相关的任务，Agent 就会自动加载对应的技能。由于这条命令会把 `olares-shared` 一起装上，“先装 shared”的要求自然就满足了。

:::tip
这些技能也发布在 ClawHub 上。两个渠道读取的是同一份 `SKILL.md`，所以装其中一个即可。如果你的 Agent 接入了 ClawHub，也可以从那里添加。
:::

Olares 上的一些 AI Agent 应用已经内置了这些技能，开箱即可让 Agent 管理 Olares。要在这类应用中使用这些技能，请参考[用 Hermes Agent 管理 Olares](../use-cases/hermes.md#manage-olares-with-your-hermes-agent)或[用 OpenClaw 管理 Olares](../use-cases/openclaw-olares-skills.md)。

## 把 Olares CLI 当作 Agent Skills 使用

加载这些包后，就能用自然语言操作 Olares，由 Agent 决定运行哪条 CLI 命令。例如：

```plain
# 通过 olares-files 技能列出文件
列出我 Olares 设备上 Home 文件夹里的文件

# 通过 olares-market 技能安装应用
从 Market 安装 Firefox，装好后告诉我

# 通过 olares-dashboard 技能查看资源占用
告诉我哪些应用占用了超过 1 GB 内存
```

:::tip
如果 Agent 没有自动加载 Olares 技能，用斜杠命令显式调用它们。
:::
