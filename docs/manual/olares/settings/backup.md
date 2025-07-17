---
description: Learn how to create and manage backups for your important files and apps on Olares.
---
# Backup your data in Olares

Olares' backup feature lets you create full and incremental backups for specified file directories and the Wise application. You can choose to back up your data on both local and network storages, with options to set up automatic backup schedules.

## Add a backup task

To add a backup task:

1. Go to **Settings › Backup**, and click **Add backup task**.
2. Choose **Backup files** or **Backup applications**.
3. On the **Add backup task** page, configure the following options:

    | Option                                   | Description                                                                                                                                                                                                                                                                                                                   |
    |:-----------------------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
    | **Backup location**                      | - **Local path**: Recommended to select an external <br/>device such as a USB drive, SMB share, or external<br/> hard drive. <br> - **Cloud storage**: Supports Olares Space, AWS S3, and <br/>Tencent COS. Accounts can be added via **Settings › <br/>Integrations** or directly in the dialog by clicking<br/> **Add account**. |
    | **Region** (for network storage)         | Choose the region for the selected storage service.                                                                                                                                                                                                                                                                           |
   | **Backup path** (for file backups)       | Specify the directory to back up.                                                                                                                                                                                                                                                                                             |
   | **Select application** (for app backups) | Choose the application to back up from the dropdown. <br/>Currently, only **Wise** is supported.                                                                                                                                                                                                                                   |
    | **Backup name**                          | Enter a recognizable task name. Recommended to <br/>include the purpose and timestamp.                                                                                                                                                                                                                                             |

    :::warning
    Ensure the selected storage has enough available space to store the backup data.
    :::

3. Set backup schedule & security:
   - **Snapshot frequency** – Choose from daily / weekly / monthly
   - **Snapshot time** – Set the time when the backup task should run
   - **Backup password** – Protect your snapshots with a password

4. Click **Submit** to start the backup task. The first execution will perform a **full backup**. Subsequent runs will perform **incremental backups**.

    :::warning
    Before starting the backup, please make sure:

    - The selected storage has enough available space to store the backup data;
    - You have an active subscription for the selected cloud storage service;
    - You have read and write permissions for the target storage location, for example, an SMB directory.


## Manage backup tasks

Once created, your backup task will appear in the task list. Click the **>** button on the right to open the detail page. Available actions include:

| Action           | Description                                                                                                               |
|:-----------------|:--------------------------------------------------------------------------------------------------------------------------|
| **Manage**       | - **Edit**: Modify the snapshot frequency and backup time  <br/>  - **Pause**:  Pause the backup task <br/> - **Delete**:  Remove the task and all associated snapshots   
| **Snapshot now** | Manually trigger a backup immediately                                                                                     |


## View snapshot records

At the bottom of the backup management page, you’ll see a list of snapshots for the backup task, with snapshot information such as:

- **Creation time**: The execution time of the snapshot.
- **Size**: Size of the snapshot.
- **Status**: The execution status of the snapshot.
- **Backup type**: The snapshot is a full backup or an incremental backup.



