---
outline: [2, 3]
description: 通过 npm 或 npx 安装和更新 olares-cli。涵盖首次安装向导、持久安装、免安装运行，以及机器上已运行 Olares OS 的特殊情况。
---

# 安装 olares-cli

Olares 会把 `olares-cli` 装在主机的 `/usr/local/bin/olares-cli`。如果要在其他机器上使用 CLI，或者需要主机自带版本尚未包含的用户模式命令，可以通过 npm 安装独立版 CLI。

:::info
运行在 Olares 内部的应用（如 OpenClaw）的容器镜像里已经预装了 `olares-cli`，集群内使用时无需手动安装。
:::

## 选择安装方式

根据你的使用习惯选择合适的方式。

### 同时安装 CLI 和 Agent Skills

如果希望通过 AI Agent 操作 Olares，推荐使用本方式。

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
向导会运行 `npm install -g @olares/cli`，然后安装 Agent Skills。它不会安装 Olares OS，也不会帮你登录。
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

## 更新 olares-cli

要将通过 npm 安装的 CLI 更新到最新发行版本，可运行以下任一命令。

明确安装最新版：

```bash
npm install -g @olares/cli@latest
```

或者使用 npm 自带的 update 命令：

```bash
npm update -g @olares/cli
```

验证更新结果：

```bash
olares-cli --version
```

输出会显示已安装的版本号，具体数字取决于你当前安装的发行版本。

:::info
更新 CLI 不会同时更新 Agent Skills。CLI 更新完成后如需更新技能，参见[更新技能](./cli-agent-skills.md#更新技能)。
:::

## 特殊情况：在运行 Olares OS 的 Linux 主机上安装 CLI

Olares OS 运行在 Linux 上，并在 `/usr/local/bin/olares-cli` 预置了一份系统自带的 CLI。因此，在这类主机上安装 CLI 时需要注意版本冲突。macOS 和 Windows [直接安装](#选择安装方式)即可。

在运行 Olares OS 的 Linux 主机上执行 `npm install -g @olares/cli` 时，npm 会检测到 `/usr/local/bin/olares-cli` 已存在，报出 `EEXIST` 错误并终止安装。这是一种保护机制，可避免 npm 覆盖非自身安装的文件，系统自带的 `olares-cli` 因此得以保留。若需在系统自带版本之外再安装一份 npm 版本，请将其安装到独立目录：

```bash
npm install -g @olares/cli --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"
```

:::warning 不要覆盖系统自带的 olares-cli
请勿在 Olares 主机上执行 `npm install -g @olares/cli --force`。该命令会覆盖由系统管理的 `/usr/local/bin/olares-cli`，破坏它与 `olaresd` 及集群之间的版本一致性。系统自带版本只能通过 `olares-cli upgrade` 升级。
:::

## 下一步

[登录 Olares](./cli-log-in.md) 创建 profile，然后通过 Agent 操作 Olares。
