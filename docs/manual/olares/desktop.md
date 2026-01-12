# Get familiar with Desktop

The Desktop application serves as the primary interface for user interaction with the Olares system. It offers an intuitive and efficient way to manage and utilize both built-in system applications and those you install.

## Desktop concepts

![Desktp[]](/images/manual/olares/desktop.png)

### Dock & Launchpad

* **Dock:** An application quick-launch bar located on the side of the screen.
* **Launchpad:** Accessed by clicking the "Launchpad" icon on the Dock, it displays all installed applications.

### Application windows

* Applications open in "window" mode by default.
* Windows support standard operations: dragging, resizing, minimizing, maximizing, and closing.
* **Search:** Enables quick application launching, file searching, and more.

## Use the Launchpad

From the Launchpad, you can:

* View all installed applications.
* Click an application icon to open it.
* Drag icons to reorder them within the Launchpad.
* Drag an icon to the Dock for quick access.

### Uninstall applications

1. Press and hold an application icon to enter the editing mode.
2. If an "X" appears in the top-right corner of the application icon, click it to delete the application.

::: tip Note
Built-in system applications such as Files, Market, and Profile cannot be uninstalled.
:::

### Control application windows

You can access applications via two modes.

By default, applications open in "window" mode, which is an iframe page embedded within the desktop. You can manipulate these windows much like standard computer windows:

* Drag the title bar to move the window.
* Drag the window's edges to resize it.
* Minimize, maximize, or close the window.
* Click the <i class="material-symbols-outlined">open_in_new</i> button to open the application in a new browser tab.

::: tip Note
Some applications only support opening in a tabbed view.
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
| File manager synced drive | Search synced files by filename only. |
| Wise reader content | Search RSS feeds, web pages, and PDFs by name. |

### Configure file search rules

By default, global search searches files by filename only.

To improve search efficiency or enable searching file contents, go to  **Settings** > **Search** > **File Search** to configure full-text search and exclusion rules.

For details, see [Configure file search](/manual/olares/settings/search.md).