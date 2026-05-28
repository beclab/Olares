---
outline: [2, 3]
description: How to get olares-cli on the right machine. Covers the built-in CLI on the Olares host, the standalone CLI for installation and AI-agent use, and the in-cluster CLI that ships with Olares apps.
---

# Install olares-cli

By default, the Olares installer ships `olares-cli` to `/usr/bin/olares-cli`.

If you need to drive Olares from an AI agent on the same host, follow this guide to download the standalone CLI.

:::info
Apps that run inside Olares such as Hermes Agent ship with `olares-cli` preinstalled in their container image. You don't need to install Olares CLI manually.
:::

## Download Olares CLI
Download the `olares-cli` package for your platform:

::: code-group

```bash [Linux x86_64]
curl -sSOL https://cdn.olares.com/olares-cli-v1.12.5-cli.1_linux_amd64.tar.gz
```

```bash [Linux ARM64]
curl -sSOL https://cdn.olares.com/olares-cli-v1.12.5-cli.1_linux_arm64.tar.gz
```

```bash [Windows x86_64]
curl -sSOL https://cdn.olares.com/olares-cli-v1.12.5-cli.1_windows_amd64.tar.gz
```

```bash [macOS Apple Silicon]
curl -sSOL https://cdn.olares.com/olares-cli-v1.12.5-cli.1_darwin_arm64.tar.gz
```

:::

## Extract and run
:::danger Do not overwrite the built-in CLI
Never move or copy the standalone binary over `/usr/bin/olares-cli`.

The built-in CLI, the `olaresd` daemon, and the cluster components are version-locked to each other. Swapping in a different binary breaks that chain and can leave future upgrades stuck. Always invoke the standalone CLI by its full path, such as `./olares-cli`, and keep `/usr/bin/olares-cli` untouched.
:::


1. Unpack the download.

   ::: code-group

   ```bash [Linux x86_64]
   tar xzf olares-cli-v1.12.5-cli.1_linux_amd64.tar.gz
   ```

   ```bash [Linux ARM64]
   tar xzf olares-cli-v1.12.5-cli.1_linux_arm64.tar.gz
   ```

   ```powershell [Windows]
   tar xzf olares-cli-v1.12.5-cli.1_windows_amd64.tar.gz
   ```

   ```bash [macOS Apple Silicon]
   tar xzf olares-cli-v1.12.5-cli.1_darwin_arm64.tar.gz
   ```

   :::

2. Make the binary executable. Skip this step on Windows.

   ```bash
   chmod +x olares-cli
   ```

3. Confirm the binary runs.

   ::: code-group

   ```bash [Linux and macOS]
   ./olares-cli --help
   ```

   ```powershell [Windows]
   .\olares-cli.exe --help
   ```

   :::
