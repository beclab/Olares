---
outline: [2, 3]
description: Common issues and solutions for OpenCode on Olares, including terminal availability, CDN errors, package management, and provider configuration.
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, troubleshooting, common issues, self-hosted
---

# Common issues

## Terminal does not accept input on ARM64 devices

The terminal panel loads but does not respond to keyboard input.

This happens because the terminal depends on `bun-pty`, which requires glibc. The OpenCode container is based on Alpine Linux (musl libc), so `bun-pty` cannot load the native PTY library on ARM64 devices.

**Workaround**: Use the OpenCode container's terminal in Control Hub instead.

## Page goes blank with a CDN error

The page suddenly goes blank, and the browser console shows `TypeError: ot(...) is not a function`.

OpenCode loads frontend resources from a CDN (`app.opencode.ai`) at runtime instead of bundling them locally. When the CDN deploys a broken frontend version, all users are affected.

To resolve this:

1. Hard-refresh your browser with **Ctrl+Shift+R** or **Cmd+Shift+R**.
2. If the issue persists, wait for the CDN to deploy a hotfix.
3. Check the [upstream repository](https://github.com/sst/opencode) for fix progress.

## Plan mode fails unexpectedly

Plan mode loads frontend resources from the same CDN as the main UI. A CDN-side bug can cause Plan mode to fail even if your local installation is up to date.

To resolve this:

1. Hard-refresh your browser.
2. Check the [upstream repository](https://github.com/sst/opencode/issues) for known issues or newer releases.

## Some packages fail to install

The OpenCode container runs on Alpine Linux, which uses musl libc instead of glibc. Packages that require a glibc-based environment, such as certain Node.js native modules or Python packages with C extensions, might not install or run correctly.

Use `pkg-install` for system-level packages where possible. See [Manage packages](opencode-packages.md) for details.

## Cannot edit provider details in the Models panel

The Models panel only supports toggling connected models on or off. You cannot edit provider configurations directly from this panel.

To update a provider's settings:

- Disconnect the provider, then reconnect it with the updated details.
- Alternatively, edit the config file directly and restart OpenCode. See [Edit the config file](opencode.md#edit-the-config-file) for instructions.

## WebFetch fails intermittently

The WebFetch tool might occasionally fail to retrieve content from certain URLs. These failures are typically caused by the target website's availability or network conditions rather than OpenCode itself.

If WebFetch fails repeatedly:

1. Retry the request after a short wait.
2. Check whether the target URL is accessible from your browser.
3. Try a different URL if the target site is temporarily unavailable.

## Cannot edit `opencode.jsonc` when `opencode.json` already exists

Under `~/.config/opencode/`, you might see two configuration files:

- `opencode.jsonc`: Automatically created by OpenCode the first time you add a custom provider from the UI.
- `opencode.json`: Added by the Olares build for plugin-heavy setups that require frequent manual edits.

OpenCode merges both files at runtime, so having both is expected.

To edit `opencode.jsonc` from Olares Files, you normally rename it to `opencode.json` so Files opens it in the JSON editor. But when `opencode.json` already exists, the rename fails because a file with that name already exists.

**Workaround**:

1. Rename `opencode.jsonc` to a temporary `.json` name, for example `opencode1.json`.
2. Open `opencode1.json` in Files, make your edits, and save.
3. Rename `opencode1.json` back to `opencode.jsonc`.
