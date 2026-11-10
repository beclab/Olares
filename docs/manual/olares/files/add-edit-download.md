---
outline: [2, 3]
description: Learn basic file operations in Olares, such as uploading, previewing, editing, and downloading files.
head:
  - - meta
    - name: keywords
      content: Olares, Olares Files, upload files, download files, preview files, edit files, file management
---

# Basic file operations
Operations in Files are essentially the same as in other file managers. This page will introduce some common tasks in Files to get you started.

For Common folders and External mounts, members need **Can view** to browse and download files, or **Can edit** to upload, edit, and delete content inside the folder. Ask the Super Admin or an admin with management access to grant the required permission. See [Manage access to Common and External](manage-common-external-access.md).

## Upload files

### Upload via the Files app
1. Open the Files app from the Dock or Launchpad on Olares.
2. In the left sidebar, select the directory where you want to upload files. For example, **Documents**.
3. Upload multiple files or folders using one of these methods:

   a. Drag and drop files from your local file manager into the Files window.

   b. Click <i class="material-symbols-outlined">drive_folder_upload</i> in the top right corner.

   c. Right-click in an empty space and select from the context menu.

:::info
Files supports resumable uploads. If an upload is interrupted, it will automatically resume from the last checkpoint.
:::

### Upload via LarePass desktop
:::info Import your Olares ID
To start using LarePass desktop, you must import your Olares ID by pasting your mnemonics. Make sure you have [backed up your mnemonics](/manual/larepass/back-up-mnemonics.md).
:::
LarePass desktop offers the same upload experience as the Files app, with automatic syncing to your Olares ID.

### Upload via LarePass mobile
You can upload files from your phone through the LarePass app.
<Tabs>
<template #Direct-upload>

1. Open LarePass and go to **Files**.
2. To upload to a new folder:

   a. Tap <i class="material-symbols-outlined">add</i> in the bottom-right corner and select **Add folder**.

   b. On the **Select the folder to upload** screen, choose a location, such as **Drive**.

   c. Tap **New folder** and follow the prompts to create the folder. LarePass opens the new folder automatically.
3. Tap <i class="material-symbols-outlined">add</i>, then select **Upload files**.
4. Select the files and follow the prompts to upload them.

   If you start from the Files home page, you will be asked to choose where to save the files. If you start inside a folder, the files upload to that folder.

5. To check upload progress, go to the Files home page and tap <i class="material-symbols-outlined">mobiledata_arrows</i> in the top-right corner to open **Transfer**.
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
5. To check upload progress, go to the Files home page and tap <i class="material-symbols-outlined">mobiledata_arrows</i> in the top-right corner to open **Transfer**.
</template>
</Tabs>

## Download files
When downloading multiple files, the behavior differs between the Files app in Olares and LarePass desktop application.
* **Files in Olares (web interface)**: Download tasks are managed directly in your browser. Manage the download queue, pause, resume, or cancel a download in the download manager of the browser. 
* **LarePass desktop**: Downloads are queued in LarePass, allowing you to pause, resume, or cancel tasks, and easily locate downloaded files.
:::tip Notes
- Folder download is only supported in LarePass desktop. 
- For large files or multiple downloads, it's recommended to use the LarePass desktop application for more powerful download management and a better user experience. Visit the [official page](https://www.olares.com/larepass) for details and download options.
:::

1. Open the Files app from the Dock or Launchpad on Olares.
2. Select any file, right-click to open the context menu, and select **Download**.
3. Select the save location in the popup window.

## Change display view

Switch between list view and grid view to display your files and folders.

1. Open the Files app, and go to your target directory.
2. In the upper-right corner, click <i class="material-symbols-outlined">dock_to_right</i> for list view or <i class="material-symbols-outlined">grid_view</i> for grid view.

![Display view](/images/manual/olares/files-display-view.png)

## Sort files

Sort files and folders in the current directory by different criteria.

1. Open the Files app, and go to your target directory.
2. Sort the items:

   - In list view, click a column header:

     - **Name**: Sort alphabetically by file or folder name.
     - **Size**: Sort by file size.
     - **Type**: Sort by file type.
     - **Last modified**: Sort by the most recent modification date.

   - In grid view, click **Last modified** in the upper-left corner, and select a sorting option.

3. To reverse the sort order, click the active column header or sorting option again.

## Preview and edit files
<!-- #region files-format-support -->
### Supported formats

The Files app supports many file formats for previewing and editing. The following tables list some typical examples by category.

#### Preview-supported formats

| Category | Format examples |
|----------------|------------------------|
| Documents | Word (`.docx`), Excel (`.xlsx`), PowerPoint (`.pptx`),<br> PDF (`.pdf`), RTF (`.rtf`), EPUB (`.epub`) |
| Images | JPEG (`.jpg`), PNG (`.png`), GIF (`.gif`), BMP (`.bmp`),<br> SVG (`.svg`), TIFF (`.tiff`) |
| Media | MP3 (`.mp3`), M4A (`.m4a`), WAV (`.wav`), FLAC (`.flac`),<br> OGG (`.ogg`), MP4 (`.mp4`), MKV (`.mkv`), MOV (`.mov`),<br>AVI (`.avi`), WebM (`.webm`), 3GP (`.3gp`) |
| Subtitles | SRT (`.srt`), VTT (`.vtt`), ASS (`.ass`) |
| Text & markup | Text (`.txt`), Markdown (`.md`), HTML (`.html`), XML (`.xml`) |
| Code & config | Python (`.py`), JavaScript (`.js`), Java (`.java`),<br> C/C++ (`.c`/`.cpp`), Go (`.go`), Rust (`.rs`), Shell (`.sh`),<br> JSON (`.json`), YAML (`.yaml`), TOML (`.toml`), INI (`.ini`) |

#### Edit-supported formats

| Category | Format examples |
|----------------|------------------------|
| Text & markup | Text (`.txt`), Markdown (`.md`), HTML (`.html`), XML (`.xml`) |
| Code & config | Python (`.py`), JavaScript (`.js`), Java (`.java`),<br> C/C++ (`.c`/`.cpp`), Go (`.go`), Rust (`.rs`), Shell (`.sh`),<br> CSS (`.css`), JSON (`.json`), YAML (`.yaml`), TOML (`.toml`), INI (`.ini`) |
<!-- #endregion files-format-support -->

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
1. Click <i class="material-symbols-outlined">search</i> in the Dock to open the search window.
2. In the search field, enter keywords related to the file you're looking for.
3. Use arrow keys <i class="material-symbols-outlined">arrow_upward</i><i class="material-symbols-outlined">arrow_downward</i> to select the search scope: **Drive** or **Sync**, and press **Enter** to see search results.

![Search](/images/manual/olares/files-search.png#bordered){width="90%"}

## Delete files
:::warning
Deleted files cannot be recovered.
:::
1. Open the Files application from the Dock or Launchpad on Olares.
2. Select the file(s) you want to delete and choose one of these methods:
   - Right-click and select **Delete** from the context menu.
   - Click <i class="material-symbols-outlined">more_horiz</i> in the top right corner and select **Delete**.
3. Confirm the deletion in the popup window.

   If the folder is shared or contains shared folders, Files also asks you to confirm that the affected shares will be canceled. Continue only if you want to cancel those shares and delete the folder. See [What happens to shares when I rename, move, or delete a folder?](share-files.md#what-happens-to-shares-when-i-rename-move-or-delete-a-folder).

## Shortcuts

To select multiple files:

* **On Windows**: Control + click
* **On Mac**: Command + click

## Resources

- [Share files](share-files.md): Share files and folders with others.
- [Sync files to local](sync-files.md): Keep files in sync across your devices.
- [Mount SMB shares](mount-SMB.md): Access network shares directly in Files.
- [Mount cloud storage](mount-cloud-storage.md): Manage Google Drive, Dropbox, and S3 in Files.
