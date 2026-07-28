---
outline: [2, 3]
description: 在 Olares 上的 OpenCode 环境中使用 pkg-install 和原生包管理器管理系统包和语言特定的依赖项。
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, pkg-install, package management, Alpine Linux, self-hosted
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/opencode-packages.md)为准。
:::

# 在 OpenCode 中管理包

你可以在 OpenCode 环境中安装额外的包：

- **系统包**：通过 `pkg-install` 安装的 OS 级工具，它是 Alpine Linux 包管理器的包装器。
- **语言包**：通过原生包管理器（如 `pip` 和 `npm`）安装的依赖项。

## 学习目标

在本教程结束时，你将学习如何：
- 使用 `pkg-install` 列出、搜索、安装和删除系统包。
- 通过 OpenCode 聊天界面使用 `system-admin` 技能安装系统包。
- 使用 `pip`、`npm`、`go install` 和 `cargo install` 安装语言特定的依赖项。
- 了解安装的包如何在容器重启和应用升级期间持久化。

## 系统包

OpenCode 容器以非 root 用户运行，因此 `sudo` 和 `apk` 不能直接可用。所有系统包管理都通过 `pkg-install` 进行。

### 检查预安装的包

OpenCode 默认包含许多常用包。要查看已安装包的完整列表，请打开 OpenCode 终端并输入以下命令：

```bash
pkg-install --list
```

![List installed packages](/images/manual/use-cases/opencode-pkg-install-list.png#bordered)

要按关键词搜索特定包：

```bash
pkg-install --search <keyword>
```

要查看有关包的详细信息：

```bash
pkg-install --info <package>
```

### 安装和删除包

要安装系统包：

```bash
pkg-install <package>
```

要安装特定版本：

```bash
pkg-install <package>=<version>
```

要删除已安装的包：

```bash
pkg-install --remove <package>
```

:::tip
OpenCode 在 Alpine Linux 上运行。包名称可能与 Debian 或 Ubuntu 上的不同。使用 `pkg-install --search` 查找正确的包名称。
:::

### 通过 AI 聊天安装

你也可以通过 OpenCode 聊天界面安装系统包。发送类似 "Install ffmpeg" 的消息，`system-admin` 技能将自动运行相应的 `pkg-install` 命令。

![Install packages through AI chat](/images/manual/use-cases/opencode-ai-pkg-install.png#bordered){width=50%}

如果技能未激活，请在聊天中手动加载它：

```text
/skill load system-admin
```

![Manually load system-admin skill](/images/manual/use-cases/opencode-load-system-admin.png#bordered){width=50%}

## 语言包

对于语言特定的依赖项，请在 OpenCode 终端中使用相应的原生包管理器：

```bash
pip install <package>          # Python
npm install <package>          # Node.js
go install <package>@latest    # Go
cargo install <package>        # Rust
```

## 包持久化

通过 `pkg-install` 安装的包在同一应用版本内的容器重启期间持久化。在应用升级期间，OpenCode 会记录你安装的包，并在新版本初始化后自动重新安装它们。
