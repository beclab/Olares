---
outline: [2, 3]
description: 通过 npm 或 npx 安装 olares-cli。涵盖首次安装向导、持久安装、免安装运行，以及机器上已运行 Olares OS 的特殊情况。
---

# 安装 olares-cli

Olares 会把 `olares-cli` 装在主机的 `/usr/local/bin/olares-cli`。如果要在其他机器上使用 CLI，或者需要主机自带版本尚未包含的用户模式命令，可以通过 npm 安装独立版 CLI。

:::info
运行在 Olares 内部的应用（如 OpenClaw）的容器镜像里已经预装了 `olares-cli`，集群内使用时无需手动安装。
:::

## 选择安装方式

根据你的使用习惯选择合适的方式。

### 同时安装 CLI 和 Agent Skills

本方式适合用 AI 智能体操作 Olares，也是推荐做法。

打开终端，运行以下命令，通过交互式向导一次装好 CLI 和 Agent Skills：

```bash
npx @olares/cli@latest install
```

输出示例：

```bash
┌  Setting up Olares CLI...
│
◇  Installed globally
│
◇  Skills installed
│
└  You are all set!

Next:
  olares-cli profile login --olares-id <your-olares-id>   # authenticate (browser/password + optional TOTP)
  olares-cli profile current                              # verify

Then tell your AI agent: "Load the olares-shared skill, then use olares-cli to ..."
```

:::info
向导会运行 `npm install -g @olares/cli`，然后安装六个 Agent Skills。它不会安装 Olares OS，也不会帮你登录。
:::

### 仅安装 CLI

本方式适用于只需持久安装 CLI、之后再装 Agent Skills 的情况。

打开终端，运行以下命令：

```bash
npm install -g @olares/cli
```

之后如需用 `npx skills` 安装 Agent Skills，参见[安装与使用 Agent Skills](./cli-agent-skills.md)。

### 免安装运行

本方式适用于只运行单条命令、无需持久安装的场景。

```bash
npx @olares/cli files ls /drive/Home
```

:::info npm 版 CLI 以 Olares 用户身份运行
通过 npm 安装或用 npx 运行的 CLI，以 Olares 用户身份工作。它可以管理运行中的 Olares 上的文件、应用、设置和集群，但不能安装或维护 Olares OS 本身。`upgrade`、`node`、`gpu`、`disk` 等主机命令只能通过 Olares OS 自带的 `/usr/local/bin/olares-cli` 运行。
:::

## 特殊情况：在运行 Olares OS 的 Linux 主机上

只有直接在已运行 Olares OS 的 Linux 机器上安装 CLI 时才会遇到这种情况，因为此时 `/usr/local/bin/olares-cli` 已经存在。macOS 和 Windows 则不会，因为自带的二进制文件位于 Linux 环境中，即便上面运行着 Olares OS 也是如此。这两类机器直接用上面的方式即可。

在 Linux Olares 主机上运行 `npm install -g @olares/cli` 会提示 `EEXIST` 错误并中止。这是正常现象：npm 不会覆盖不归它管理的二进制文件，因此你的系统 `olares-cli` 不受影响。如需在它旁边再装一份 npm 版本，可指定单独的前缀目录：

```bash
npm install -g @olares/cli --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"
```

:::warning
不要在 Olares 主机上运行 `npm install -g @olares/cli --force`。这会覆盖由系统管理的 `/usr/local/bin/olares-cli`，破坏它与 `olaresd` 和集群之间的版本一致性。系统自带版本只能通过 `olares-cli upgrade` 升级。
:::

## 下一步

[登录 Olares](./cli-log-in.md) 创建 profile，然后通过智能体操作 Olares。
