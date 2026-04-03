---
outline: [2, 3]
description: Manage system packages and language-specific dependencies in the OpenCode environment on Olares using pkg-install and native package managers.
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, pkg-install, package management, Alpine Linux, self-hosted
---

# Manage packages

You can install additional packages in the OpenCode environment:

- **System packages**: OS-level tools installed through `pkg-install`, a wrapper around the Alpine Linux package manager.
- **Language packages**: Dependencies installed through native package managers such as `pip` and `npm`.

## System packages

The OpenCode container runs as a non-root user, so `sudo` and `apk` are not available directly. All system package management goes through `pkg-install`.

### Check pre-installed packages

OpenCode includes many common packages by default. To view the full list of installed packages, open the OpenCode terminal and enter the following command:

```bash
pkg-install --list
```

![List installed packages](/images/manual/use-cases/opencode-pkg-install-list.png#bordered)

To search for a specific package by keyword:

```bash
pkg-install --search <keyword>
```

To view detailed information about a package:

```bash
pkg-install --info <package>
```

### Install and remove packages

To install a system package:

```bash
pkg-install <package>
```

For example, to install `ffmpeg`:

```bash
pkg-install ffmpeg
```

To install a specific version:

```bash
pkg-install <package>=<version>
```

To remove an installed package:

```bash
pkg-install --remove <package>
```

:::tip
OpenCode runs on Alpine Linux. Package names may differ from those on Debian or Ubuntu. Use `pkg-install --search` to find the correct package name.
:::

### Install via AI chat

You can also install system packages through the OpenCode chat interface. Send a message such as "Install ffmpeg", and the `system-admin` skill runs the appropriate `pkg-install` command automatically.

![Install packages through AI chat](/images/manual/use-cases/opencode-ai-pkg-install.png#bordered)

If the skill does not activate, load it manually by entering the following command in the chat:

```text
/skill load system-admin
```

![Manually load system-admin skill](/images/manual/use-cases/opencode-load-system-admin.png#bordered)

## Language packages

For language-specific dependencies, use the appropriate native package manager in the OpenCode terminal:

```bash
pip install <package>          # Python
npm install <package>          # Node.js
go install <package>@latest    # Go
cargo install <package>        # Rust
```

## Package persistence

Packages installed through `pkg-install` persist across container restarts within the same app version. During an app upgrade, OpenCode records your installed packages and reinstalls them automatically after the new version initializes.
