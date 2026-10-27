---
outline: [2, 3]
description: Troubleshoot an Olares app that fails to install or update, remains stuck during an operation, or no longer opens afterward.
head:
  - - meta
    - name: keywords
      content: Olares, app install failed, app update failed, Market logs, app stuck, troubleshooting
---

# App fails during or after installation or update

Use this guide when an app installation or update fails, remains in progress, or finishes but the app does not start.

## Check the operation that failed

1. Open **Market** and go to **My Olares**.
2. Click **Logs** in the upper-right corner.
3. Find the failed installation or update. Record its time, operation type, app version, and complete error message.
4. Download the operation log before retrying. A later attempt can replace the most useful context shown in the interface.

## Match the error to the next action

| Error or symptom | What to do |
|---|---|
| Insufficient CPU or memory | Stop apps you do not need, then follow [Memory is insufficient or not freed](./ts-free-memory.md). |
| Insufficient disk space | Follow [Free up disk space](../free-up-disk-space.md), then retry the operation. |
| GPU or VRAM allocation failed | Follow [GPU app remains stopped after installation or resume](./ts-vram-shortage.md). |
| A dependency is missing or stopped | Check **Dependencies** on the app details page. Install or resume the required app first. Cluster members may need the administrator to install a shared dependency. |
| The app requires a newer Olares version | Update Olares from **Settings** > **System** > **Update** before retrying the app update. |
| The operation completed, but the app does not open | Check the app in **Settings** > **Applications**. If it is stopped, resume it once and note any message shown in the launch dialog. |

## Check configuration changed outside Settings

An app update can replace configuration that was edited directly in Control Hub or in the underlying workload. Those edits are not the supported persistent configuration path.

If the app needs environment variables or other app settings, review them in **Settings** > **Applications** > **[app]** and in the app's own settings. Re-enter only values you recognize and can verify. Do not copy secrets into screenshots or public reports.

For persistent environment-variable options, see [Manage application environment variables](../olares/settings/manage-app-env.md).

## Retry safely

After resolving the error shown in the operation log:

1. Retry the installation or update once from Market.
2. Wait for the operation to finish before starting another lifecycle action.
3. Confirm that the app shows **Running** in **Settings** > **Applications**, and then open it from the Launchpad.

Do not delete the app namespace, Application resource, deployments, or shared services to force a reinstall. These resources can be shared or managed by Olares, and removing them can affect other users or leave the app record inconsistent.

## If the app is still stuck

Provide support with:

- Olares version and the app version before and after the attempted operation.
- Whether this was a new installation, an update, or a resume.
- The exact status shown in Market and Settings.
- The downloaded Market operation log.
- The time of the failure and any recent configuration change.

If more system information is needed, follow [Collect diagnostic information](../collect-diagnostic-information.md) and share the archive through a private support channel.

