---
outline: [2, 3]
description: 了解如何用 olares-cli 管理集群、诊断系统，以及让 AI Agent 代你操作 Olares。涵盖主机、用户和集群内三种模式。
---

# Olares CLI

`olares-cli` 是 Olares 的命令行工具。安装 Olares 时会自带它，你也可以在 macOS、Windows 或 Linux 上单独安装，无需安装完整系统。它可以用来完成以下任务：

- **管理集群**：在新机器上安装 Olares、在版本之间升级、添加或移除节点，以及彻底卸载。
- **诊断和修复系统**：运行安装前检查、收集日志、查询节点状态、排查问题。
- **让 AI Agent 操作 Olares**：让 Agent 通过自然语言代你管理文件、应用和设置。

## Olares CLI 的三种模式
olares-cli 提供三种模式，区别在于运行位置和鉴权方式。

| 模式 | 运行位置 | 鉴权方式 |
|------|---------|---------|
| 主机模式 | 与 Olares OS 同一台机器 | 主机 root 和 kubeconfig，无需登录。|
| 用户模式 <Badge type="tip" text="^1.12.5" /> | 任意安装了 `olares-cli` 的机器，以登录<br>用户的身份运行 | 通过与网页端、LarePass 相同的 HTTP API，使用 profile 和访问令牌 |
| 集群内模式 <Badge type="tip" text="^1.12.6" /> | Olares 应用容器内部 | 凭证以环境变量注入，权限范围由应用的 `OlaresManifest` 决定 |

:::tip
要使用用户模式，请通过 npm 安装 CLI。如果主机上已有的 `olares-cli` 版本早于表中所示版本，相关命令可能不可用。参见[安装 olares-cli](./cli-install.md)。
:::

## 让 AI Agent 操作 Olares

要让 AI Agent 代你运行 Olares CLI：

1. 安装 CLI 和 Agent Skills。推荐用 `npx @olares/cli install`，一步装好两者。参见[安装 olares-cli](./cli-install.md)。
2. 用 Olares ID 登录。参见[登录 Olares](./cli-log-in.md)。
3. 通过 Agent 用自然语言操作 Olares。

## 了解更多

- [集群管理](./install/index.md)：在新机器上安装 Olares、在版本之间升级、管理磁盘和 GPU、收集日志等。
- [安装 olares-cli](./cli-install.md)：在你的机器上安装独立版 CLI。
- [登录 Olares](./cli-log-in.md)：创建 profile，让 CLI 以你的 Olares 用户身份操作。
- [安装与使用 Agent Skills](./cli-agent-skills.md)：为 Agent 添加操作 Olares 所需的技能。
