---
description: Restore files from your Olares Space backups, including viewing backup usage and snapshots.
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, restore, cloud backup, snapshots
---
# Restore data from Olares Space backups

Use Olares Space to view your cloud backups and restore files to your Olares.

## View backup storage usage

Check how much storage your backups are using against your quota.

:::info
For self-hosted Olares users, it's important to monitor the storage usage for backup services. These services may incur charges based on usage. For more information, see [Billing](billing.md).
:::

1. Click **Backup** in the left navigation pane.

  ![Backup list in Olares Space](/images/how-to/space/backup_list.png#bordered)

2. Locate the **Backup Usage** section. It shows the total storage used by all backups and the total quota.

## View backup details

View the snapshots of a backup task to see what has been backed up.

1. Click **Backup** in the left navigation pane.
2. Locate the **Backup List** section. You can find all backup tasks in the table.
3. To view the snapshots of a backup task:

    a. Click **View Details** on the backup task.

    b. Review the backup information and snapshots.

    ![Snapshots list in Olares Space](/images/how-to/space/snapshots_list.png#bordered)

## Restore from a backup

Restore files from a backup to your Olares.

:::info
Understand charges for storage and bandwidth before using backup and restore services. Each instance includes a certain amount of free traffic, and any usage exceeding the quota will incur charges. For more information, see [Billing](billing.md).
:::

1. On the backup details page, select the snapshot you want to restore:

   - To restore the most recent snapshot, click **Restore** in the upper-right corner.
   - To restore a specific snapshot, find it in the **Snapshots** table, and then click **Restore** on that row.

2. In the **Restore** window:

    a. Copy the **Backup Url**.

      ![Restore window in Olares Space](/images/how-to/space/backup_restore_dialog.png#bordered){width=70%}

    b. Click the **Settings > Restore** link. This opens the **Restore** page in a new browser tab. You can also go to **Settings** > **Restore** in your Olares directly.

3. On the **Restore** page, add a restore task:

   - If the page is empty, click **Add restore task**, and then select **From Olares Space**.
   - If the page already has restore tasks, click <i class="material-symbols-outlined">add</i> in the upper-right corner, and then select **From Olares Space**. 
 
  ![Add restore task options](/images/how-to/space/restore_add_task.png#bordered){width=70%}

4. Fill in the restore information:

    a. **Backup URL**: Paste the copied **Backup Url**.

    b. **Restore password**: Enter the password you set when creating the backup task.

    c. **Restore location**: Select the directory where you want to restore the files.

    d. **New folder name**: Enter a name for the new folder that will contain the restored files.

    ![Restore from Olares Space form](/images/how-to/space/restore_from_olares_space.png#bordered){width=70%}

5. Click **Start restore**.
6. After the restoration finishes, view and access the restored files:

    a. Click the restore task card to open the restore details.

      ![Restore task completed](/images/how-to/space/restore_complete.png#bordered){width=70%}

    b. On the **Restore details** page, click **Open in Files**. The Files app opens the folder containing the restored files.

      ![Restore details page](/images/how-to/space/restore_details.png#bordered){width=70%}

## Resources

- [Back up your data in Olares](../olares/settings/backup.md): Learn how to create backup tasks in Olares.
