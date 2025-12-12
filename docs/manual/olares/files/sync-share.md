---
outline: [2, 3]
description: Keep files synchronized across devices and share content securely with other Olares members using built-in file sharing capabilities.
---
# Sync and share files

LarePass ensures your files remain consistent and accessible across all devices while facilitating seamless collaboration within your Olares server.

This page covers:
- **Core concepts**: Libraries, roles, and permissions.
- **Library setup**: How to create a new Library to organize your digital content.
- **Synchronization**: How to enable bi-directional sync between the Library and your local computer.
- **Library management**: How to share Libraries with team members and manage user permissions.
- **Sync management**: How to control sync status (pause/unsync) and handle sync conflicts.

## Before you begin

### Understand Sync and Library

#### Sync

Sync in the Files app is similar to cloud storage services like iCloud, where you can keep your most important information up to date, and available across all your devices. Sync also makes it easy to share files with other members within an Olares server.

#### Library

Library is the fundamental unit for organizing, syncing, and sharing your digital content. It is more than just a folder. It's a versatile container designed to meet various data synchronization and sharing needs:

* **Multi-device synchronization**: Libraries ensure your data remains consistent across all your devices.
* **Real-time collaboration**: Share Libraries with other users, enabling simultaneous access and editing of data within the same Library.
* **Flexible management**: Create multiple Libraries to organize different types of data or for various projects, giving you granular control over your synchronization and sharing preferences.

### Roles and permissions

The roles and permissions described here are specific to Library file sharing and Library management within Files. These are distinct from the overall Olares user roles and system-wide permissions.

| Operation                  | Owner | Member |
|----------------------------|-------|--------|
| Create Library             | ✅     | ✅      |
| Manage Library permissions | ✅     | ❌      |
| Invite other members       | ✅     | ❌      |
| Share and rename Library   | ✅     | ❌      |
| Remove members             | ✅     | ❌      |
| Delete Library             | ✅     | ❌      |
| Exit Library               | ❌     | ✅      |

Permission levels:
- **Read-only**: Users can view Library contents but cannot modify them.
- **Read-write**: Users can add, delete, and modify Library contents.

### Prerequisites

Make sure you have installed the LarePass desktop client from the [official website](https://www.olares.com/larepass), and logged in using your Olares ID.

:::info
Currently, local file sync is available for Windows and Mac users. We'll use the Mac version for our examples.
:::

## Create a Library

Each user is automatically provided with their own personal Library as a starting point. To create a new Library:

1. In the left sidebar under **Sync**, click <i class="material-symbols-outlined">add_circle</i>.
2. In the **New library** dialog, enter a name for the Library, and then click **Create**.

## Enable synchronization

### Sync files to local computer

Once a Library is created, you can map it to your local computer for easy access.

1. Open LarePass desktop on your Mac.
2. Hover your mouse over the target Library, click <i class="material-symbols-outlined">more_horiz</i> that appears on the right, and then click **Sync to local**. 
3. Specify your preferred local directory, and then click **Complete**.

   Syncing will begin immediately. Once completed, a green checkmark will appear on the bottom-left corner of the folder icon, indicating active two-way synchronization.

### Sync local files to Library

Once the sync relationship is established, synchronization is automatic.

:::info
If your permission to the Library is read-only, you cannot sync changes from the local folder to the Library. Your newly added and modified files will be read-only, indicated by a gray disabled icon <i class="material-symbols-outlined">remove</i>.
:::

* **Upload**: Drag and drop any file into the local sync folder on your computer. It will be automatically uploaded to the Library.
* **Edit**: Open a file from the local folder, edit it, and then save. The changes will be automatically synchronized to the Library.

## Manage Libraries

You can share Libraries with team members, set permissions, or delete Libraries you no longer need.

:::tip
To add a member in Olares, see [manage team](../settings/manage-team.md).
:::

### Share a Library

You can share a Library with other members within an Olares server:

1. Hover your mouse over the target Library, click <i class="material-symbols-outlined">more_horiz</i> that appears on the right, and then click **Share with**.
2. In the **Invite users** section, click <i class="material-symbols-outlined">add</i>, search for and select the target user or group, and then click **Invite**.
3. In the **Set user permissions** list, click <i class="material-symbols-outlined">chevron_forward</i> to the right of the user avatar to assign specific read or write permissions, and then click **Submit**.
4. Click **Confirm**.

Invited users will see the shared Library in their Sync content list. To revoke sharing permissions, simply remove the user from the sharing window.

### Exit or delete a Library

If you don't want to share a Library, you can exit sharing or delete it.
- **Exit sharing**: Any member can exit a shared Library. When an owner exits, the Library will appear in their personal Library list.
- **Delete**: Only the owner can delete a shared Library.
   :::warning
   Deleting a Library is irreversible. All files in the shared Library will be permanently deleted.
   :::

1. To exit a Library:
   
   a. Select a shared Library, click <i class="material-symbols-outlined">more_horiz</i> that appears on the right, and then click **Exit sharing**.

   b. Click **Confirm**.
2. To delete a Library: 

   a. Select a shared Library, click <i class="material-symbols-outlined">more_horiz</i> that appears on the right, and then click **Delete**.

   b. Click **Confirm**.

## Manage synchronization

You can manage the sync status in the following ways.

### Locate local folder

If you want to quickly locate the sync directory on your local drive, hover your mouse over the target library or folder, click <i class="material-symbols-outlined">more_horiz</i> that appears on the right, and then click **Open local sync folder**. The system will directly open the folder's location on your computer.

### Stop syncing

If you no longer need to sync a folder, hover your mouse over it, click <i class="material-symbols-outlined">more_horiz</i> that appears on the right, and then click **Unsychronize**.

### Pause sync

If you want to temporarily stop data transfer, click <i class="material-symbols-outlined">pause_circle</i> to the right of the **Sync** directory. All sync tasks will be paused.

## Handle sync conflicts

In the rare event that multiple devices edit the same file simultaneously, LarePass automatically handles the conflict to prevent data loss:
* The first completed edit is saved to the Library.
* A backup of the conflicting version is created with a unique filename, including the editor's Olares ID and timestamp: `test.txt(SFConflict name 2024-04-17-12-12-12)`.