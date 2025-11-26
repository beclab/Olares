---
description: Keep files synchronized across devices and share content securely with other Olares members using built-in file sharing capabilities.
---
# Sync and share files

Sync in Olares is similar to cloud storage services like iCloud, where you can keep your most important information up to date, and available across all your devices. Sync also makes it easy to share files with other members within an Olares server. 

**LarePass** plays a central role in sync and share. It not only keeps your Files content synchronized across devices but also facilitates collaboration through file sharing and library management.

This page will mainly cover:
- How to sync files across devices
- How to share files with other team members

:::info
- Currently, file sync is available on the **LarePass desktop client** for Windows and macOS.  
- Library management is supported on both the **Files app** on Olares and the **LarePass desktop client**. 
- **LarePass mobile app** supports shared library access, but not creation.
:::

## Before you begin

Make sure you have installed the LarePass desktop client from the [official website](https://www.olares.com/larepass), and logged in using your Olares ID.


## Manage your libraries

Library is the fundamental unit for organizing, syncing, and sharing your digital content.

### Understand roles and permissions

:::info
The roles and permissions described here are specific to file sharing and library management within Files. These are distinct from the overall Olares user roles and system-wide permissions.
:::

| Operation                  | Owner | Member |
|----------------------------|-------|--------|
| Create library             | ✅     | ✅      |
| Manage library permissions | ✅     | ❌      |
| Invite other members       | ✅     | ❌      |
| Share and rename library   | ✅     | ❌      |
| Remove members             | ✅     | ❌      |
| Delete library             | ✅     | ❌      |
| Exit library               | ❌     | ✅      |


### Create a library

Each user is automatically provided with their own personal library as a starting point. To create a new one:

1. Open LarePass desktop client.
2. From the left sidebar, navigate to **Files** > **Sync**, click the <i class="material-symbols-outlined">add_circle</i> to open the **New library** dialog.
3. Enter a name for the Library and click **Create**.

Alternatively, you can create a library in the **Files** app in Olares.

### Share a library

:::tip
To add a member in Olares, see [manage team](/manual/olares/settings/manage-team.md).
:::

To share a library with other members in the same Olares cluster:

1. Select a library, and click <i class="material-symbols-outlined">more_horiz</i> > **Share with**.
2. In the dialog, select users from the dropdown menu, and click **Share to user**.
3. Set file permissions for each user:
   - Read-only: Users can view Library contents but cannot modify them.
   - Read-write: Users can add, delete, and modify Library contents.
4. Click **Close**.

   ![Share library](/images/manual/olares/share-library.png#bordered){width="50%"}

Invited users will see the shared library in their Sync content list. To revoke sharing permissions, simply remove the user from the sharing window.

### Exit or delete a library

If you don't want to share a library, you could exit sharing or delete it.

- **Exit sharing**: Any member can exit a shared Library. When an owner exits, the library will appear in their personal library list.
- **Delete**: Only the owner can delete a shared Library.
   :::warning
   Deleting a library is irreversible. All files in the shared library will be permanently deleted.
   :::

1. To exit a library:
   
   a. Select a shared Library and click <i class="material-symbols-outlined">more_horiz</i> > **Exit sharing**.

   b. Click **Confirm** in the popup dialog.
2. To delete a library: 

   a. Select a shared library and click <i class="material-symbols-outlined">more_horiz</i> > **Delete**.

   b. Click **Confirm** in the popup dialog.

## Sync your library

Once you've created a library, you can set up synchronization to keep its contents up to date across your devices.

### Sync library files to local

You can download and keep a copy of your library files on your local device with two-way synchronization.

1. Open LarePass desktop client, and navigate to **Files** > **Sync** > **My Libraries**. 
2. Locate the library you just created, then click <i class="material-symbols-outlined">more_horiz</i> > **Sync to local**.
3. Select your local folder, and click **Confirm**.

The library now display a sync symbol icon, indicating active two-way synchronization. Any changes made in the library will automatically sync to your local folder.

### Sync local files to library

Because sync is two-way, any files placed in the synced local folder will also be uploaded to your library.

:::tip Note
If your permission to the library is read-only, you cannot sync changes from the local folder to the Library. Your newly added and modified files will be read-only, indicated by a gray disabled icon <i class="material-symbols-outlined">remove</i>.
:::

To sync an existing local folder to your library: 

1. Create or select a Library in LarePass.
2. Specify the local folder for sync. 
3. Move your files into the local folder you selected.

This lets you use your familiar local directory structure while keeping it in sync with Olares.

### Manage sync settings

If you need to pause or stop synchronization for a specific Library:

1. Locate the sync library in LarePass.
2. Click <i class="material-symbols-outlined">more_horiz</i> > **Unsynchronize**.

This action won't delete your local files. It simply halts the two-way synchronization process.

## Handle sync conflicts

In the rare event of a sync conflict, LarePass has you covered. When multiple devices edit the same file simultaneously:

* The first completed edit is saved to the Library.
* A backup of the conflicting version is created with a unique filename, including the editor's Olares ID and timestamp: `test.txt(SFConflict name 2024-04-17-12-12-12)`.

## Learn more
- [Manage team](../olares/settings/manage-team.md)
- [Manage files](../olares/files/)