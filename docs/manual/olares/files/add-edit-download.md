---
outline: [2, 3]
description: Learn essential file operations in Olares including adding new files, editing existing content, and downloading files across different devices.
---
# Basic file operations
Operations in Files are essentially the same as in other file managers. This page will introduce some common tasks in Files to get you started.

## Upload files

### Upload via the Files app
1. Open the Files app from the Dock or Launchpad on Olares.
2. In the left sidebar, select the directory where you want to upload files. For example, **Documents**.
3. Upload multiple files or folders using one of these methods:
   - Drag and drop files from your local file manager into the Files window. 
   - Click <i class="material-symbols-outlined">drive_folder_upload</i> in the top right corner. 
   - Right-click in an empty space and select from the context menu.

:::info
Files supports resumable uploads. If an upload is interrupted, it will automatically resume from the last checkpoint.
:::

### Upload via LarePass desktop
:::info Import your Olares ID
To start using LarePass desktop, you must import your Olares ID by pasting your mnemonics. Make sure you have [backed up your mnemonics](/manual/larepass/back-up-mnemonics.md).
:::
LarePass desktop offers the same upload experience as the Files app, with automatic syncing to your Olares ID.

### Upload via LarePass mobile
You can also upload files or folders on your phone via the LarePass app.
<Tabs>
<template #Direct-upload>

1. Open LarePass app and navigate to the **Files** tab.
2. Select the directory where you want to upload files.
3. Tap <i class="material-symbols-outlined">add_circle</i> in the bottom-right corner, and select one of the following upload options:
   - **File**: Select from your phone's storage.
   - **Image/Video**: Select from your phone's gallery.
   :::tip
   If you want to organize your uploads, you can create a **Create folder** first.
   :::
4. Follow the on-screen instructions to complete the upload.
</template>

<template #Share-to-upload>

:::info
The exact steps may vary depending on your operating system and browser.
:::

This method allows you to quickly upload files or media via your phone's sharing options.
1. Open the share menu for the file.
2. Select the LarePass icon in the sharing options, or select **LarePass** in the **Other actions** menu. You will be directed to the LarePass app.
3. In the LarePass app, select the destination for your upload:
   - **drive**: Upload files to your Drive storage for personal use.
   - **sync**: Upload files to your Sync storage for sharing or synchronization.
4. Follow the on-screen instructions to complete the process based on your selected target location.
</template>
</Tabs>

Files uploaded via the LarePass mobile app will also sync automatically with your Olares ID.

## Download files
When downloading multiple files, the behavior differs between the Files in Olares and LarePass desktop application.
* **Files in Olares (web interface)**: Download tasks are managed directly in your browser. Manage the download queue, pause, resume, or cancel a download in the download manager of the browser. 
* **LarePass desktop**: Downloads are queued in LarePass, allowing you to pause, resume, or cancel tasks, and easily locate downloaded files.
:::tip Notes
- Folder download is only supported in LarePass desktop. 
- For large files or multiple downloads, it's recommended to use the LarePass desktop application for more powerful download management and a better user experience. Visit the [official page](https://www.olares.com/larepass) for details and download options.
:::

1. Open the Files app from the Dock or Launchpad on Olares.
2. Select any file, right-click to open the context menu, and select **Download**.
3. Select the save location in the popup window.

## Preview and edit files

### Supported formats

The Files app supports the following file formats for previewing and editing:
- Preview:
   * Images: JPG, JPEG, PNG, BMP, WEBP, SVG
   * Videos: MP4, MKV, MOV, MPEG, MTS, WMV, WEBM, RM, 3GP
   * Audio: MP3, WMA, WAV, OGG, AAC, M4A, APE, FLAC
   * Text: PDF, TXT, JS, CSS, XML, YAML, HTML
- Edit: 
   * TXT, JS, CSS, XML, YAML, HTML

### Procedure

- To preview a file, double-click the target file.
- To edit a file:
   1. Double-click a supported text file.
   2. Click <i class="material-symbols-outlined">edit_square</i> in the top-right corner.
   3. Modify the content, and then save your changes by clicking <i class="material-symbols-outlined">save</i>.

      ![Preview](/images/manual/olares/files-preview.png#bordered)

## Search files
You can easily find files in the Files app using desktop search.
:::tip
Full-text search is available for the `/Documents/` directory in **Drive**, allowing you to search within file contents. For other directories, you can search files using their file names.
:::
1. Click <i class="material-symbols-outlined">search</i>in the Dock to open the search window.
2. In the search field, enter keywords related to the file you're looking for.
3. Use arrow keys <i class="material-symbols-outlined">arrow_upward</i><i class="material-symbols-outlined">arrow_downward</i> to select the search scope: **Drive** or **Sync**, and press **Enter** to see search results.

![Search](/images/manual/olares/files-search.png#bordered){width="90%""}
## Delete files
:::warning
Deleted files cannot be recovered.
:::
1. Open the Files application from the Dock or Launchpad on Olares.
2. Select the file(s) you want to delete and choose one of these methods:
   - Right-click and select **Delete** from the context menu.
   - Click <i class="material-symbols-outlined">more_horiz</i> in the top right corner and select **Delete**.
3. Confirm the deletion in the popup window.

## Change display view

Switch between list view and grid view to display your files and folders differently.

![Display view](/images/manual/olares/files-display-view.png)
## Shortcuts
To select multiple files:

* **On Windows**: Control + click
* **On Mac**: Command + click
