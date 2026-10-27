---
outline: [2, 3]
description: 了解 Olares CLI Agent Skills 如何帮助 AI Agent 管理 Olares、包含哪些技能，以及如何安装和更新。
head:
  - - meta
    - name: keywords
      content: Olares, Olares CLI, Agent Skills, olares-shared, 技能清单, AI Agent 集成
---

# Olares CLI Agent Skills

Olares CLI 内置了一组 Agent Skills（技能），用来告诉 AI Agent 如何通过 `olares-cli` 管理 Olares。本页列出可用技能，并介绍如何安装、更新和查看技能内容。

如果是第一次配置 Olares CLI，请按照[通过 AI Agent 使用 Olares CLI](./cli-install.md)完成安装、登录和第一个 AI Agent 任务。

## Agent Skills 清单

| 技能 | 说明 |
|-------|------|
| `olares-shared` | 判断 Olares 请求该交给哪个技能，并说明 `profile`（连接配置）、身份认证和登录恢复。 |
| `olares-chart` | 把项目仓库、Docker Compose 项目或 Helm chart 转成 Olares 应用并部署。 |
| `olares-files` | 管理 Olares 文件，包括上传、下载、压缩、分享，以及挂载 SMB 或 NFS。 |
| `olares-market` | 浏览 Olares 应用市场，安装、升级、卸载、克隆、停止、恢复或重启应用。 |
| `olares-settings` | 管理用户、应用、VPN、网络、备份、集成账号、GPU 和搜索等设置。 |
| `olares-dashboard` | 查看 CPU、内存、磁盘、网络、Pod、GPU、风扇和应用资源占用。 |
| `olares-cluster` | 查看 Kubernetes 运行状态、日志、容器、工作负载、Job、节点和中间件。 |
| `olares-doctor` | 排查安装失败、崩溃、镜像拉取失败、应用无法访问和性能问题。 |
| `olares-router` | 配置云端或本地 AI 模型、默认模型、密钥、配额、用量和模型调用。 |
| `olares-knowledge` | 管理 URL、yt-dlp、aria2、种子、Hugging Face 和 Wise 下载任务。 |
| `olares-search` | 搜索 Drive 文件、Sync 资料库和已安装应用。 |
| `olares-publish` | 把本机运行的应用发布到公开的 Olares 应用市场，包括上架信息和素材。 |

## 安装或导出技能

把 CLI 内置的技能写入标准目录：

```bash
olares-cli skills install
```

Olares CLI 会把一份完整副本写入 `~/.agents/skills`，再链接到电脑上已有的兼容 Agent 技能目录，包括 Cursor、Claude Code、Codex 和 OpenClaw 使用的目录。未安装的 Agent 不会被创建目录。如果系统不支持符号链接，Olares CLI 会改为复制文件。

如果需要把技能套件放进项目目录、容器镜像或其他自定义位置，可以导出：

```bash
olares-cli skills export ./skills
```

## 保持技能与 Olares CLI 版本一致

如果通过 npm 安装了 Olares CLI，可同时更新 CLI 和已安装的技能：

```bash
olares-cli update
```

如果手动更新 CLI 或安装了其他构建版本，请再运行：

```bash
olares-cli skills install
```

如果当前技能来自另一份 CLI 构建，Olares CLI 会在标准错误中显示提示。只有自动化流程不能接受额外输出时，才使用 `OLARES_CLI_NO_SKILL_NOTICE=1` 关闭提示。

请从 Olares CLI 二进制文件安装技能，不要从注册表或 Git 分支获取。注册表和仓库中的副本可能与本机 CLI 支持的命令不一致。ClawHub 上的副本早于内置技能，现已停止更新。

## Agent Skills 的工作方式

每个技能包都包含一份 `SKILL.md`。AI Agent 会根据描述判断何时加载技能，再按照其中的说明选择 Olares CLI 命令、填写参数、完成身份认证和处理常见错误。

源文件位于 Olares 仓库的 [`cli/skills/`](https://github.com/beclab/Olares/tree/main/cli/skills)。复杂命令的详细说明放在技能包的 `references/` 目录中。

发布时，这些技能会编译进 Olares CLI 二进制文件。因此，`skills install`、`skills export` 和 `skills read` 使用的内容始终来自当前运行的 CLI 版本。

## 查看 CLI 内置的技能

列出当前 Olares CLI 内置的技能：

```bash
olares-cli skills list
olares-cli skills list -o json
```

无需安装，直接读取某个技能或其参考文件：

```bash
olares-cli skills read olares-shared
olares-cli skills list olares-router/references
olares-cli skills read olares-router references/olares-router-calling.md
```

这些命令只会读取当前 CLI 内置的技能，不会连接 Olares，也不会检查 Agent 能否读取技能。如需在 Agent 中确认技能是否可用，请参阅[让 AI Agent 完成第一个任务](./cli-install.md#让-ai-agent-完成第一个任务)。
