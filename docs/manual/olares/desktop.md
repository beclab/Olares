---
description: Get familiar with Olares Desktop, including the Dock, Launchpad, application windows, widgets, layout reset, and global search.
---

# Get familiar with Desktop

Desktop is the primary interface for interacting with Olares. From here, you can open and manage built-in system apps as well as the apps you install.

## Desktop basics

![Desktop](/images/manual/olares/desktop.png#bordered)

### Dock

The Dock is an application quick-launch bar on the left side of the screen. Use it to open frequently used apps and access key Desktop features.

### Launchpad

Launchpad shows all installed applications. Click the Launchpad icon in the Dock to open it.

### Application windows

By default, applications open in window mode as an embedded page within Desktop. You can manage windows like you would on a standard computer:

- Drag the title bar to move the window.
- Drag the window edges to resize it.
- Minimize, maximize, or close the window.
- Click <i class="material-symbols-outlined">open_in_new</i> to open the app in a new browser tab.

:::info
Some applications only support opening in a browser tab.
:::

### Search and notifications

- **Search**: Quickly launch applications and find supported content across Olares.
- **Notifications**: Click the notification icon to view system and application notifications.

### Widgets

Desktop can display optional widgets in the lower-right corner:

- **Date & time**: Shows the current time, weekday, and date.
- **Dashboard**: Shows CPU, disk, and memory usage.

The **Widgets** switch is the master switch. Turning it off hides all widgets. When it is on, you can control each widget separately.

1. Open **Settings** from the Dock or Launchpad.
2. Select **Appearance** in the sidebar.
3. Use the **Widgets** switch as the master control:
   - Off: Hide all widgets.
   - On: Enable individual widget settings.
4. Under **Date & time**, turn the clock and date widget on or off. If it is on, set **24-hour format** and **Date format** as needed.
5. Turn **Show dashboard** on or off to show or hide CPU, disk, and memory usage.

If both individual widgets are off, Desktop shows no widgets even when **Widgets** is on.

## Reset Desktop layout

If you want to restore the default Desktop organization, reset the layout from Settings.

1. Open **Settings** from the Dock or Launchpad.
2. Select **Appearance** in the sidebar.
3. Scroll to **Reset desktop layout**, then click **Reset**.

4. In the confirmation dialog, choose what to do:
   - Click **Cancel** to keep your current layout.
   - Click **Reset** to restore the default Launchpad and Dock layout.

:::warning
Resetting the desktop layout restores the default Launchpad and Dock layout. Custom icon positions and Dock items are reset. App data is not deleted, but the layout reset cannot be undone.
:::

## Use Launchpad

From Launchpad, you can:

- View all installed applications.
- Click an application icon to open it.
- Drag icons to reorder them within Launchpad.
- Drag an icon to the Dock for quick access.

### Uninstall applications

1. Press and hold an application icon to enter editing mode.
2. If a <i class="material-symbols-outlined">close_small</i> icon appears in the top-left corner of the app icon, click it to uninstall the application.

:::info
Built-in system applications such as Files, Market, and Settings cannot be uninstalled.
:::

## Search within Olares

You can quickly open global search using one of the following methods:

- Press the keyboard shortcut: `Shift + Space`
- Click the <i class="material-symbols-outlined">search</i> icon in the Dock.

Global search can find applications, files, and other supported search targets.

| Supported search target | Supported search capability |
|:--|:--|
| Applications (built-in and installed) | Search by application name and open the app directly.| 
| Directories enabled for full-text search | Search files by filename and by text content inside <br>supported documents. |
| Other directories in File manager storage  | Search files by filename only.  |
| Team shared files | Search shared files by filename only. |
| Syned files | Search synced files by filename only. |
| Wise reader content | Search RSS feeds, web pages, and PDFs by name. |

### Configure file search rules

By default, global search searches files by filename only.

To improve search efficiency or enable searching file contents, go to  **Settings** > **Search** > **File Search** to configure full-text search and exclusion rules.

For details, see [Configure file search](/manual/olares/settings/search.md).
