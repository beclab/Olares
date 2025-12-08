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
Double-click a file to open its preview. The Files app supports previewing the following file formats:

* **Images**: JPG, JPEG, PNG, BMP, WEBP, SVG
* **Videos**: MP4, MKV, AVI, MOV, MPEG, MTS, TS, WMV, WEBM, RM, 3GP
* **Audio**: MP3, WMA, WAV, OGG, AAC, M4A, APE, FLAC
* **Text**: PDF, TXT, JS, CSS, XML, YAML, HTML

The Files app also supports editing the following text formats: TXT, JS, CSS, XML, YAML, HTML.

![Preview](/images/manual/olares/files-preview.png#bordered)

## Share files

This feature supports sharing folders with internal members or external users through different methods. You can flexibly set access permissions, passwords, and expiration dates, and manage or cancel these shares at any time.

### Share types and support scope

Currently, file sharing is only supported at the **folder** level. Supported sharing types vary depending on the directory location of the folder. Refer to the following table for details:

| Directory type | Public share | Internal share | SMB share |
| :--- | :---: | :---: | :---: |
| **Home** | Supported | Supported | Supported |
| **External** | - | Supported | Supported |
| **Sync** | - | Supported | - |
| **Data** | Supported | Supported | Supported |
| **Cache** | - | Supported | Supported |

Where,
- **Public share**: Generates an access link, suitable for external sharing (password protection supported).
- **Internal share**: Restricted to registered users or user groups within the system, suitable for internal team collaboration.
- **SMB share**: Shared via the SMB protocol within the Local Area Network (LAN).

### Create share

#### Step 1: Select folder

1. Open the Files app from the Dock or Launchpad on Olares.
2. Enter the share menu based on the directory type:
    * The **Sync** directory: Hover your mouse over the target folder, click the **Operate** icon that appears on the right, and then click **Share with**.
    * Other directories: Right-click the target folder you want to share, and then select a share type.
   
   ![Share file type](/images/manual/olares/share-files.png#bordered)

3. Complete the settings for your selected share type by referring to the procedure in the following step.

#### Step 2: Configure settings

##### Internal share

If you selected **Internal share**, complete the following steps to configure the settings:
1. In the **Invite users** field, click the **+** (Plus) icon, search for and select the target user or group, and then click **Invite**.
2. In the **Set user permissions** list, click the arrow to the right of the user avatar to assign specific read or write permissions for each invited user, and then click **Submit**.
3. Click **Confirm**.

##### Public share

If you selected **Public share**, complete the following steps to configure the settings:
1. In the **Set password** field, enter an access password, or click **Regenerate** to use a system-generated secure password.
2. In the **Set expiration** region, select the specific expiration time for the share link.
3. Select the following restrictions as needed:
    - **Limit file size**: To limit the size of files uploaded by visitors, select this option, enter a value, and then select the unit (eMi/Gi/Ti).
    - **Allow uploads only**: If you selet this check box, visitors can only upload files and cannot view or download existing content.
4. Click **Confirm** to generate the link.
5. Copy or note down the link address, and then click **Confirm**.

##### SMB share

If you selected **SMB share**, complete the following steps to configure the settings:

1. Select whether to make it **Public**.
2. If you selected **Yes**, click **Confirm**, and then it will be accessible to anyone on the LAN.
3. If you selected **No**, authentication is required: 
    
    a. In the **Invite users** field, click the **+** (Plus) icon to select an existing user or group, or click **Add user accounts** to manually add and invite new users.

    b. Click **Confirm**.
    
    c. In the **Set user permissions** list, configure specific permissions for each invited user, and then click **Confirm**.

### View and manage shares

You can view and manage all shared folders in the **Share** list.

#### View share list and attributes

1. Open the Files application from the Dock or Launchpad on Olares.
2. Click **Share** in the left sidebar.
2. You can view the list of all currently shared folders on the right.
3. If you want to view the details of a specific folder, right-click the target folder and click **Attributes**. 

    You will see the sharer, current path, original path, share scope, owner, current permissions, expiration date, and link details.

### Modify password or permissions

For shared folders, you can change their password and permission settings at any time:

1. Open the Files application from the Dock or Launchpad on Olares.
2. Click **Share** in the left sidebar to view the **Share** list on the right.
3. To reset password (**Public share** only):
    
    a. Right-click the folder and select **Reset Password**.
    
    b. Enter a new password or click **Regenerate**.
    
    c. Click **Confirm**.
4. To edit permissions (**Internal share** and **SMB share** only):
    
    a. Right-click the folder and select **Edit permissions**.

    b. Add or remove users, or modify permissions for existing users.

    c. Click **Confirm**.

### Cancel share

If you no longer need to share a folder, you can cancel the share at any time:
1. Open the Files application from the Dock or Launchpad on Olares.
2. Click **Share** in the left sidebar to view the **Share** list on the right.
3. Right-click the target folder and click **Revoke sharing**.
4. Click **Confirm**.

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
