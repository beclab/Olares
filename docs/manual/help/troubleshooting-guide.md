---
outline: [2, 3]
description: Match common Olares symptoms to troubleshooting guides for apps, AI, networking, storage, accounts, and Olares One hardware.
---

# Troubleshoot Olares

Find the category that owns the failing part of Olares, then match what you see to a guide. You do not need to decide whether a problem is simple or complex.

Before making changes, record the exact error message and time, check your Olares and app versions, and review [known issues](./known-issues.md). Avoid reinstalling Olares or deleting data unless a guide specifically requires it.

## Apps and Market

Use this category for app availability, installation, removal, and state problems.

| What you see | Guide |
|---|---|
| An expected app is missing from Market | [Missing apps in Market](./ts-missing-apps.md) |
| A stopped app cannot be removed from App exclusive mode | [Cannot remove a stopped app in App exclusive mode](./ts-cs-app-reappears.md) |
| Market and Control Hub show different app states | [App status differs after using Control Hub](./ts-inconsistent-app-status.md) |
| An app install or update fails, remains in progress, or the app no longer opens | [App fails during or after installation or update](./ts-app-fails-after-update.md) |

## AI and model runtime

Use this category when an AI app or model cannot obtain the memory or GPU resources it needs.

| What you see | Guide |
|---|---|
| Available memory is low or memory remains allocated after stopping apps | [Insufficient memory or memory not freed](./ts-free-memory.md) |
| A GPU app remains stopped or cannot start because VRAM is unavailable | [GPU app remains stopped after installation or resume](./ts-vram-shortage.md) |
| A model does not become ready or its inference engine does not start | [Model or engine is not ready](./ts-model-engine-not-ready.md) |

## Network, access, and domains

Use this category when Olares cannot be reached locally or remotely, a private-network connection fails, or streaming is unstable.

| What you see | Guide |
|---|---|
| LarePass VPN does not connect or cannot reach Olares | [LarePass VPN not working](./ts-larepass-vpn-not-working.md) |
| Olares is not reachable during activation or after restart | [Network not ready or Olares connection error](./ts-network-not-ready.md) |
| App URLs return 404 after a custom route ID was changed | [Custom route ID prevents app access](./ts-custom-route-domain.md) |
| Steam streaming is slow, delayed, or unstable | [Slow or delayed Steam streaming](./ts-steam-stream-lag.md) |

## Storage, backup, and files

Use this category when the system is short on disk space or stored app and model data needs attention.

| What you see | Guide |
|---|---|
| Disk space is full or model and image files use too much storage | [Free up disk space](../free-up-disk-space.md) |

## Accounts and authentication

Use this category for sign-in, password, activation, and authentication failures.

| What you see | Guide |
|---|---|
| You forgot the desktop login password | [Forgotten desktop login password](./ts-forget-login-password.md) |
| LarePass shows **System error** | ["System error" in LarePass](./ts-system-error.md) |
| Sign-in or activation shows a specific error message | [Login and activation error messages](../login-and-activation-errors.md) |

## Olares One hardware

Use this category for Olares One BIOS, startup, and dual-boot problems.

| What you see | Guide |
|---|---|
| Olares One has no status light or a connected display remains blank | [Olares One has no power or display](/one/ts-no-power-or-display.md) |
| The screen stays black when you try to enter the BIOS | [Cannot enter BIOS (Black screen)](/one/ts-bios-black-screen.md) |
| Windows dual boot does not start or behave as expected | [Troubleshoot Windows dual boot](/one/dual-boot-windows-troubleshooting.md) |

## If the issue is not listed

Record the problem context and gather only the relevant logs. See [Collect diagnostic information](../collect-diagnostic-information.md).
