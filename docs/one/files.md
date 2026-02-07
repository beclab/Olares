---
outline: [2, 3]
description: Get started with the Files app on your Olares One. Learn the interface layout, how to upload your first files, preview and edit content, search your drive, and connect external cloud storage like Google Drive.
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, upload files, manage files, access cloud files, Olares Files app
---

# Manage files <Badge type="tip" text="15 min" />

Olares's built-in Files app provides a secure and efficient solution for file storage and synchronization. It allows you to access your personal files anywhere, keep data synchronized across devices, and manage third-party cloud storage in one place.

Operations in Files are essentially the same as in other file managers. This guide introduces some common tasks in Files to get you started.

## Learning objectives

- Know the Files interface.
- Upload files from your local computer. This guide uploads a `.pdf` document and a `.txt` file as the example.
- Preview and edit supported file types.
- Locate files using the system search.

## Before you begin

### Interface overview

Familiarize yourself with the workspace. The Files interface mainly consists of the following areas:

* Drive: Stores your personal files such as documents and pictures.
* Sync: A library-based storage area for files requiring frequent modifications or real-time synchronization.
* Application: Reserved for application-specific data. Primarily for development and debugging, not general file storage.
* Cloud storage: Access files from integrated services such as Google Drive and AWS S3.
* Share: A centralized hub for managing shared folders.

![Files](/images/manual/olares/files-index.png#bordered)

### File formats support

The Files app supports the following file formats for previewing and editing:

- Preview:
   * Images: JPG, JPEG, PNG, BMP, WEBP, SVG
   * Videos: MP4, MKV, MOV, MPEG, MTS, WMV, WEBM, RM, 3GP
   * Audio: MP3, WMA, WAV, OGG, AAC, M4A, APE, FLAC
   * Text: PDF, TXT, JS, CSS, XML, YAML, HTML
- Edit: 
   * TXT, JS, CSS, XML, YAML, HTML

## Upload files

Populate your server with content by uploading documents via the Files app.

1. Open the Files app from the Dock or Launchpad.

    ![Find and open the Files app](/images/one/find-files-app.png#bordered)

2. In the left sidebar, select the directory where you want to upload files. For example, **Documents**.

    ![Select upload destination](/images/one/upload-file.png#bordered)

3. Upload the PDF document and text file using one of these methods:
   - Drag and drop files from your local file manager into the Files window. 
   - Click <i class="material-symbols-outlined">drive_folder_upload</i> in the top right corner. 
   - Right-click in an empty space and select from the context menu.

    :::info Upload progress indicator
    During upload, a notification in the lower right corner displays the upload status.
    :::

4. When the upload is completed, the files are listed in the directory you selected.

     ![File upload completed](/images/one/upload-file-completed.png#bordered)

:::info Resumable uploads
If your internet connection drops, the upload stops automatically. When the connection is back, Olares resumes the upload from exactly where it left off, so you don't have to start over.
:::

## Preview & edit files

You can view and modify many common file types directly in the Files app without downloading them.

- To preview the PDf document, double-click it.

    ![Find Files app](/images/one/preview-file.png#bordered)

- To edit the text file:
   
   1. Double-click it.

    ![Preview a text file](/images/one/open-txt.png#bordered)

   2. Click <i class="material-symbols-outlined">edit_square</i> in the top-right corner.
   3. Modify the content, and then save your changes by clicking <i class="material-symbols-outlined">save</i>.
    
    ![Edit and save text file](/images/one/edit-save-file.png#bordered)

## Search files

Locate files quickly using the global desktop search.

1. Click <i class="material-symbols-outlined">search</i>in the Dock to open the search window.

    ![Search files](/images/one/search-dock.png#bordered)

2. Select **Drive** as the search scope.

    ![Select drive](/images/one/select-drive.png#bordered)

3. In the **Search files** field, enter the keywords related to the file you're looking for. For example, `txt`. The results are listed.

    ![Search a txt file](/images/one/search-txt.png#bordered)

4. Double-click to open it in the Files app directly.

## Download files

Download files to your local computer.

1. Open the Files app from the Dock or Launchpad.
2. Right-click the target file, and then select **Download**. 

    The file is downloaded via your web browser's standard download manager.

:::info
Web browsers cannot download entire folders or extremely large files reliably. For these tasks, install the LarePass Desktop application to unlock folder downloads and resume capabilities. 
:::

## Delete files

:::warning
Deleted files cannot be recovered.
:::

1. Open the Files app from the Dock or Launchpad.
2. Right-click the target file, and then click  **Delete**.
3. In the **Delete** window, click **Confirm**.
<!--## Next steps

Integrate external storage services like Google Drive, Dropbox, or AWS S3 to view and manage your cloud-based files directly in Olares without switching applications.

1. Authorize the connection and add the drive to your Files app.

    a. Connect your service using the supported method accordingly:

    - OAuth-based storage services such as Google Drive and Dropbox: Connect via the [LarePass app](../manual/larepass/integrations.md#add-a-cloud-drive-via-oauth).
    - API credential-based services such as AWS S3 and Tencent Cloud Object Storage (COS): Connect via the [LarePass app](../manual/larepass/integrations.md#add-a-cloud-storage-using-api-keys) or [Olares Settings](../manual/olares/settings/integrations.md#add-cloud-object-storage-via-api-keys).

    b. Open the Files app and confirm that the service appears as a drive under **Cloud Drive**.

2. After the cloud storage is connected, you can interact with the cloud files from the Files app as if they were local.

    a. Open the Files app and select the cloud storage.

    b. Manage the cloud files the same way as you manage your local files. Changes made in the Files app sync instantly to the cloud provider.
    
    - Upload and download files
    - Preview supported file types
    - Rename, move, or delete files and folders

3. To remove the drive, delete the integration in the LarePass app or Olares Settings.

    a. Remove the integration to disconnect the drive:

    - [Remove integration in LarePass](../manual/larepass/integrations.md#disconnect-integrations)

    - [Remove integration in Olares Settings](../manual/olares/settings/integrations.md#view-and-manage-connected-services)

    b. Verify that the drive disappears from the Files app.

    The drive disappears from your Files app, but your data remains safe on the cloud service.-->

## Resources

- [Basic file operations](../manual/olares/files/add-edit-download.md)
- [Share files](../manual/olares/files/share-files.md)
- [Sync files to local](../manual/olares/files/sync-files.md)
- [Mount SMB](../manual/olares/files/mount-SMB.md)
- [Mount cloud storage](../manual/olares/files/mount-cloud-storage.md)
- [LarePass official page](https://www.olares.com/larepass)