---
description: Keep files synchronized across devices and share content securely with other Olares members using built-in file sharing capabilities.
---
# Sync and share files
LarePass is a powerful tool not only ensures your Files content remains consistent and accessible, but also facilitates seamless collaboration within your Olares server.

This page will mainly cover:
- How to sync files across devices
- How to share files with other team members

## Understand Sync and Library
### Sync
Sync in the Files app is similar to cloud storage services like iCloud, where you can keep your most important information up to date, and available across all your devices. Sync also makes it easy to share files with other members within an Olares server.

### Library
Library is the fundamental unit for organizing, syncing, and sharing your digital content. It is more than just a folder. It's a versatile container designed to meet various data synchronization and sharing needs:

* **Multi-device synchronization**: Libraries ensure your data remains consistent across all your devices.
* **Real-time collaboration**: Share libraries with other users, enabling simultaneous access and editing of data within the same Library.
* **Flexible management**: Create multiple libraries to organize different types of data or for various projects, giving you granular control over your synchronization and sharing preferences.

### Roles and permissions
:::info
The roles and permissions described here are specific to file sharing and Library management within Files. These are distinct from the overall Olares user roles and system-wide permissions.
:::

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

## Before you begin
Make sure you have installed the LarePass desktop client from the [official website](https://olares.com/larepass), and logged in using your Olares ID.

:::info
Currently, local file sync is available for Windows and Mac users. We'll use the Mac version for our examples.
:::

## Create a Library
Each user is automatically provided with their own personal Library as a starting point. To create a new Library:

1. In the left sidebar under **Sync**, click the <i class="material-symbols-outlined">add_circle</i> to open the **New library** dialog.
2. Enter a name for the Library and click **Create**.

## Sync Library files to local

1. Open LarePass on your Mac.
2. Locate your desired Library and click <i class="material-symbols-outlined">more_horiz</i> > **Sync to local**.
3. Select your preferred local directory, and click **Complete**.
4. To initiate the sync, click <i class="material-symbols-outlined">more_horiz</i> > **Sync now**.

Once synchronized, your libraries will display a green icon, indicating active two-way synchronization. Any changes made locally will automatically reflect in your synced Library.

## Sync local files to Library
:::info
If your permission to the Library is read-only, you cannot sync changes from the local folder to the Library. Your newly added and modified files will be read-only, indicated by a gray disabled icon <i class="material-symbols-outlined">remove</i>.
:::

To sync an existing local folder on your Mac, simply create a matching Library in LarePass and move your files into the designated sync directory.

This approach allows you to maintain your current folder structure while benefiting from LarePass's synchronization capabilities.

## Managing sync settings
If you need to pause or stop synchronization for a specific Library:

1. Locate the Library in LarePass.
2. Click <i class="material-symbols-outlined">more_horiz</i> > **Unsynchronized**.

Rest assured, this action won't delete your local files. It simply halts the two-way synchronization process.

## Share a Library
:::tip
To add a member in Olares, see [manage team](../settings/manage-team.md).
:::

You can share a Library with other members within an Olares server:

1. Select a Library, and click <i class="material-symbols-outlined">more_horiz</i> > **Share with**.
2. In the dialog, select users from the dropdown menu, and click **Share to user**.
3. Set file permissions for each user: **Read-write** or **Read-only**.
4. Click **Close**.

   ![Share library](/images/manual/olares/share-library.png#bordered){width="50%"}

Invited users will see the shared Library in their Sync content list. To revoke sharing permissions, simply remove the user from the sharing window.

## Exit or delete a Library
If you don't want to share a Library, you could exit sharing or delete it.
- **Exit sharing**: Any member can exit a shared Library. When an owner exits, the Library will appear in their personal Library list.
- **Delete**: Only the owner can delete a shared Library.
   :::warning
   Deleting a Library is irreversible. All files in the shared Library will be permanently deleted.
   :::

1. To exit a Library:
   
   a. Select a shared Library and click <i class="material-symbols-outlined">more_horiz</i> > **Exit sharing**.

   b. Click **Confirm** in the popup dialog.
2. To delete a Library: 

   a. Select a shared Library and click <i class="material-symbols-outlined">more_horiz</i> > **Delete**.

   b. Click **Confirm** in the popup dialog.

## Handle sync conflicts

In the rare event of a sync conflict, LarePass has you covered. When multiple devices edit the same file simultaneously:

* The first completed edit is saved to the Library.
* A backup of the conflicting version is created with a unique filename, including the editor's Olares ID and timestamp: `test.txt(SFConflict name 2024-04-17-12-12-12)`.

## Learn more
- [Manage team](../settings/manage-team.md)