---
description: Restore files to a specific directory or recover application data using backup snapshots. Learn how to restore data from local paths, Olares Space, or AWS S3.
---

# Restore backup data

You can use existing backup snapshots to restore files to a specified directory or recover application data. This guide covers how to restore data from local storage, Olares Space, and AWS S3.

## Add a restore task

To add a restore task:

1. Go to **Settings** › **Restore**, then click **Add restore**.
2. Choose a restore method based on your backup location:

<tabs>
<template #Restore-from-local>

1. Select the local backup path. The path must point to the backup task folder. For example, if the task name is `demo` and the location is `/documents`, the correct path would be: `/documents/olares-backups/demo-xxxx`.
2. Enter your backup password.
3. Click **Query snapshots** to get available snapshots.
4. Click **Restore** next to the desired snapshot to load it.
5. If restoring **files**, specify the restore location and destination folder, then click **Start Restore**.  
If restoring the **Wise** application, simply click **Start Restore** without specifying a path.

</template>
<template #Restore-from-Olares-Space>

1. Use the LarePass app to scan and log in to [Olares Space](https://space.olares.com).
2. On the **Backup** page, locate the desired backup and click **View Details** on the right.
3. Click the **Restore** button at the top right to get the latest snapshot URL, or select a specific snapshot and click **Restore** next to it.
4. Copy the URL and paste it into the **Backup URL** field in the Olares restore page.
5. Enter your backup password.
6. If restoring **files**, specify the restore location and destination folder, then click **Start Restore**.  
   If restoring the **Wise** application, simply click **Start Restore** without specifying a path.

</template>
<template #Restore-from-AWS-S3>

1. Go to the [AWS S3 Console](https://console.aws.amazon.com/s3), navigate to your bucket, and locate the `olares-backups` directory.
2. Select the target backup folder, then generate a **pre-signed URL** for that folder.  
   See [AWS S3 documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ShareObjectPreSignedURL.html) for help.
3. Copy the URL and paste it into the **Backup URL** field in the Olares restore page.
4. Enter your backup password.
5. Click **Query snapshots** to load available snapshots.
6. Click **Restore** next to the desired snapshot.
7. If restoring **files**, specify the restore location and destination folder, then click **Start Restore**.  
   If restoring the **Wise** application, simply click **Start Restore** without specifying a path.
</template>
</tabs>

::: tip 
If you are restoring from Tencent COS, follow the steps for AWS S3. Refer to [Tencent Cloud documentation](https://cloud.tencent.com/document/product/436/68284) for instructions on how to get the pre-signed URL for your backup files.
:::

## View restore tasks

Once created, your restore task will appear in the task list on the Restore page. Click the **›** button on the right to view the task details. Available actions are:


- **Cancel restore task** – Click **Cancel** to interrupt and stop the restore process.
- **View files or app** – Once completed, click **Open App** or **Open Folder** to access the restored data.