---
description: Protect your Olares instances with cloud backup and restore features. Handle snapshots, perform restore operations, and manage storage quotas in Olares Space.
---
# Back up and restore

Olares Space is the official solution to back up snapshots for your Olares instances. You can restore an Olares to its most recent state whenever needed. This section provides instructions for managing backups and restores in Olares Space. 

:::tip
Each Olares is provided with 10 GB of free backup space. Any usage beyond this will be charged according to the cloud provider's pricing.
:::

## View backup list

The backup task list shows information for each backup task, including:

- Initial creation time
- Most recent snapshot time
- Overall storage usage 

![alt text](/images/how-to/space/backup_list.jpg#bordered)

Click **View Details** on a task to see its detail page. The detail page shows the storage usage since the task was created and a list of all successful snapshots.

:::info NOTE
Currently, only restoring from the most recent snapshot is supported.
:::

## Restore backup to the Olares Space

![alt text](/images/how-to/space/restore_backup_to_the_olares_space.jpg#bordered)

Restoring a snapshot to the cloud is similar to setting up a new cloud-based Olares.

1. Set up relevant details.

   a. Select the cloud service provider and their data center location. 

   b. Choose the hardware configuration for the instance. 

   c. Confirm the snapshot details and enter the backup password.

2. Understand charges for storage and bandwidth. <br>Each instance includes a certain amount of free storage and traffic. Any usage exceeding these quotas will incur charges.

3. Confirm the order and complete the payment. After that, the Olares begins to install.

:::info NOTE
During the installation process, Olares will verify the backup password. If it is incorrect, you'll be asked to re-enter the correct one. If you forget the backup password, the restoration process won't be able to continue. In this case, please return your instance and try restoring again.
:::

:::info NOTE
To avoid conflicts or other unforeseeable problem, you must return the existing Olares that uses the same name before restoring to a cloud-based Olares.
:::