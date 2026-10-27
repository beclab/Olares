---
outline: [2, 3]
description: 使用 olares-cli 维护 Olares OS、管理运行中的 Olares，并通过内置 Agent Skills 让 AI Agent 执行任务。
head:
  - - meta
    - name: keywords
      content: Olares, olares-cli, 命令行工具, 集群管理, Agent Skills, AI Agent
---

# Olares CLI

`olares-cli` 是 Olares 官方命令行工具，可用于安装、升级和维护 Olares OS，也可以管理运行中的 Olares 上的文件、应用、设置和系统状态。CLI 内置 Agent Skills（技能），让 AI Agent 可以把自然语言请求转换为具体命令。

你可以在 macOS、Windows 或 Linux 上单独安装 Olares CLI，让另一台电脑上的 AI Agent 管理正在运行的 Olares。Olares OS 主机也会自带一份用于系统维护的 CLI。

## 开始使用

### 通过 AI Agent 管理 Olares

在运行 AI Agent 的电脑上安装 Olares CLI 和技能，再用 Olares ID 登录。之后用自然语言描述任务，Agent 就能通过命令管理 Olares 上的文件、应用、设置、模型和系统状态。按照[通过 AI Agent 使用 Olares CLI](./cli-install.md)完成配置和第一个任务。

运行在 Olares 上的 [Lares](../use-cases/lares.md)、[Hermes Agent](../use-cases/hermes.md) 和 [OpenClaw](../use-cases/openclaw.md) 已经内置 Olares CLI 和技能。请按照所用应用的指南开始使用。

### 维护 Olares OS

在运行 Olares OS 的 Linux 主机上，使用系统自带的 `/usr/local/bin/olares-cli` 安装或升级 Olares、管理节点和硬件、收集日志、排查系统问题。这些操作使用主机的 root 权限和 kubeconfig，不需要登录 Olares ID。

检查状态、收集日志和排查问题，请参阅[Olares CLI 常用命令](./install/cli/olares-cli.md)。有关安装架构、环境变量和版本规则，请参阅[集群管理](./install/index.md)。

## 了解更多

- [Olares CLI Agent Skills](./cli-agent-skills.md)：查看技能清单，了解技能的分发和版本同步机制。
- [使用任意 AI Agent，通过 Olares CLI 管理 Olares](https://www.olares.com/blog/olares-cli-ai-agents-explained/)：了解 Olares、CLI、Agent Skills 与 AI Agent 如何协同工作。
