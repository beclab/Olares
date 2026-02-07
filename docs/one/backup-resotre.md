---
outline: [2, 3] 
description: Learn how to back up and restore files and app data on Olares One.
head:
  - - meta
    - name: keywords
      content: Back up, restore
---
# Back up and restore data <Badge type="tip" text="15 min" />

Olares One provides a built-in backup feature to protect important files and app data. You can create full and incremental backups, store them on local or network storage, and restore data from any available snapshot when needed.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Create backup tasks for folders and supported applications.
- Configure backup locations, schedules, and password protection.
- Manage existing backup tasks.
- Restore files to a specific directory or recover app data from a snapshot.

## Back up your data

A backup task defines what to back up, where to store it, and when it runs.

### Create a backup task

1. Go to **Settings** > **Backup**.
2. Click  **Add backup task**. If prompted, select **Back up files** or **Back up apps** to proceed.
3. On the Add backup task page, configure the following settings:

    | Field                  | Description                                                                                                                                                                                                                                                                                                                                                                                          |
    |------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
    | **Backup location**    | <ul><li>**Local path**: Select an external device such as a USB drive, SMB <br>share, or external hard drive.</li><br><li>**Olares Space**: Scan the QR code in **Settings** > your avatar > **Olares<br> Space** via LarePass.<br></li><li> **AWS S3 or Tencent COS**: Click **Add account** in the dialog, or go to <br>**Settings** > **Integrations** > **Link your accounts & data**.</li></ul> |
    | **Region**             | Cloud only. Select the region of your storage bucket.                                                                                                                                                                                                                                                                                                                                                |
    | **Backup path**        | File backup only. Browse and select the specific directory to back up.                                                                                                                                                                                                                                                                                                                               |
    | **Select application** | App backup only. Select the app to back up. Currently, only Wise is<br> supported.                                                                                                                                                                                                                                                                                                                   |
    | **Backup name**        | Enter a recognizable task name. We recommend including the<br> purpose and a timestamp.                                                                                                                                                                                                                                                                                                              |
    | **Snapshot frequency** | Choose **Every day**, **Every week**, or **Every month**.                                                                                                                                                                                                                                                                                                                                            |
    | **Run backup at**      | Set the specific time of day for the backup to run.                                                                                                                                                                                                                                                                                                                                                  |
    | **Backup password**    | Set a password to encrypt your snapshots.                                                                                                                                                                                                                                                                                                                                                            |
    | **Confirm password**   | Re-enter the password you set.                                                                                                                                                                                                                                                                                                                                                                       |
    
    ![Add backup task](/images/one/settings-add-backup.png#bordered)
4. Click **Submit** to create and start the task.
    - The first run will be a full backup.
    - Subsequent runs will be incremental backups (saving only changes).

### Manage backup tasks

Once created, tasks appear in the list. Click the arrow <i class="material-symbols-outlined">chevron_right</i> next to a task to view details.

| Action           | Description                                                                                                                                                                     |
|------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Manage**       | <ul><li> **Edit**: Modify snapshot frequency and time.<br></li><li> **Pause**: Pause the task.<br></li><li> **Delete**: Remove the task and all associated snapshots.</li></ul> |
| **Snapshot now** | Run a backup immediately.                                                                                                                                                       |

![Manage backup tasks](/images/one/settings-manage-backup-task.png#bordered){width=85% style="display:block;margin-left:0;margin-right:auto;"}

### View snapshot records

At the bottom of the backup detail page, you can review snapshots for this task:

| Field             | Description                                 |
|-------------------|---------------------------------------------|
| **Creation time** | When the snapshot was created.              |
| **Size**          | Total size of the snapshot data.            |
| **Status**        | The execution status of the snapshot.       |
| **Backup type**   | Whether it is a full or incremental backup. |

![View snapshots](/images/one/settings-snapshots.png#bordered){width=70% style="display:block;margin-left:0;margin-right:auto;"}

## Restore data

You can restore files to a specific folder or recover application data using any valid snapshot.

### Create a restore task

1. Go to **Settings** > **Restore**.
2. Click  **Add restore task**.
3. Choose the method below that matches where your backup is stored.

### Restore from a local path

Use this method for backups stored on USB drives or SMB shares connected to Olares.

1. Select the local backup path. The path must point to the specific backup task folder.
   
    If the backup name is `demo`, and the backup location is `/documents`, the path should be: 
    ```text
    /documents/olares-backups/demo-xxxx
    ```
2. Enter your backup password.
3. Click **Query snapshots** to load available snapshots.
4. Click **Restore** next to your target snapshot.
5. Start the restore:
    - Files: Select the restore location then click **Start restore**.
    - Wise: Click **Start restore** directly (no destination path required).

### Restore from Olares Space

Use this method if you backed up to Olares Space. You will need the LarePass mobile app.

1. Open LarePass and scan to sign in to [Olares Space](https://space.olares.com).
2. On the **Backup** page, locate the backup task and click **View Details**.
3. Get a snapshot URL:
   - Click **Restore** (top right) to get the latest snapshot URL, or
   - Select a specific snapshot and click **Restore** next to it.
4. Copy the URL and paste it into the **Backup URL** field in the From Space URL page.
5. Enter your backup password.
6. Start the restore:
    - Files: Select the restore location and destination folder, then click **Start restore**.
    - Wise: Click **Start restore** directly.

### Restore from cloud storage (AWS S3 / Tencent COS)

Use this method for backups stored on AWS S3 or Tencent COS.

1. Open your cloud storage console and locate the `olares-backups` directory.
2. Select the target backup folder and generate a **pre-signed URL**.
    - For AWS S3, see the [AWS S3 documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ShareObjectPreSignedURL.html) on pre-signed URLs.
    - For Tencent COS, follow [Tencent Cloud documentation](https://cloud.tencent.com/document/product/436/68284) for generating temporary access URLs.
3. Copy the generated URL.
4. Paste it into the **Backup URL** field on the From AWS S3 URL or From Tencent COS URL page.
5. Enter your backup password.
6. Click **Query snapshots** to load available snapshots.
7. Click **Restore** next to your target snapshot.
8. Start the restore:
    - Files: Select the restore location and destination folder, then click **Start restore**.
    - Wise: Click **Start restore** directly.

## Monitor restore tasks

Once created, the restore task appears in the task list on the Restore page. Click the arrow <i class="material-symbols-outlined">chevron_right</i> next to a task to view details and manage restore tasks.

- Cancel: Stop the process while it is running.
- View data: Once the status shows `Completed`, click **Open folder** or **Open app** to access the restored data.