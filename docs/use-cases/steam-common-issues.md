---
outline: [2, 3]
description: Troubleshoot common Steam Headless issues on Olares, including package persistence after app restarts and upgrades.
head:
  - - meta
    - name: keywords
      content: Olares, Steam Headless, common issues, Flatpak, apt, app persistence, troubleshooting
app_version: "1.0.43"
doc_version: "1.0"
doc_updated: "2026-09-03"
---

# Steam Headless common issues

Find solutions to common Steam Headless problems on Olares.

## Why do packages installed with `apt` disappear after Steam Headless restarts?

Packages installed with `apt` are written to the container's root filesystem. Steam Headless recreates this filesystem when the app restarts, is redeployed, or is upgraded. As a result, packages installed manually with `apt` are not retained.

To keep additional packages across Steam Headless restarts and updates, use Flatpak instead of `apt` when a Flatpak package is available. In Steam Headless 1.0.43 and later, Flatpak applications, runtimes, and user data are stored in persistent app storage.

To install a package with Flatpak:

1. Open Control Hub and go to **Browse** > **steamheadless**.
2. Expand **Deployments** > **steamheadless**, then open the running Pod.
3. Under **Containers**, click the Terminal icon next to **steam-headless**.
4. Run the Flatpak installation command in the container shell.

For more information about Pods and containers, see [Manage containers](../manual/olares/controlhub/manage-container.md).
