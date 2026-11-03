---
outline: [2, 3]
description: Control which Olares members can view or edit Common folders and external storage through the Files app.
---

# Manage access to Common and External

Control who can view or edit Common folders and External storage in Files.

This guide covers access settings for Common and External. To share a Home or Sync folder, see [Share files](share-files.md).

:::warning Prerequisites
Access settings require Olares 1.12.8 or later. Log in as the Super Admin or an admin who can manage the target folder. Members cannot change these settings. For External storage, mount the storage before assigning permissions.
:::

## Supported folders

Permissions apply to the selected folder and all files and subfolders inside it. You cannot set separate permissions for individual subfolders. These settings control access through Files and do not change permissions on your NAS.

| Location | Folder you can configure | Example |
| --- | --- | --- |
| <nobr>**Application** > **Common**</nobr> | A top-level folder in Common. | `Common/huggingface` |
| <nobr>**Drive** > **External**</nobr> | A top-level mount in External: an SMB share, NFS share, USB drive, or local disk.<br><br>**Note:** The `ai` system folder is read-only for members and does not support access settings. | `External/ProjectNAS` |

## Role defaults and management scope

The following defaults apply to Common folders and External mounts.

| Olares role | Default access | Whose access they can change |
| --- | --- | --- |
| Super Admin (Owner) | Full management access | Admins and members |
| Admin | Full management access, unless restricted by the Super Admin | Members, while the admin retains management access to the folder |
| Member | No access | Cannot change folder access settings |

For more about Olares roles, see [Manage your team](../settings/manage-team.md).

## Grant member access

1. Log in to Olares as the Super Admin or an admin who can manage the folder, then open **Files**.
2. Go to **Application** > **Common** or **Drive** > **External**.
3. Right-click the top-level folder or mounted directory, and select **Access settings**.

   ![Access settings in an External SMB folder context menu](/images/manual/olares/files-access-settings.png#bordered)

4. Find the member and choose a permission:

   ![Access settings dialog](/images/manual/olares/files-access-settings-dialog.png#bordered)

   | Permission | What the member can do |
   | --- | --- |
   | <nobr>**No access**</nobr> | Cannot see the folder in the list or access its contents through Files. |
   | <nobr>**Can view**</nobr> | Browse, preview, and download files. Cannot add, edit, or delete content. |
   | <nobr>**Can edit**</nobr> | View files, create folders, upload files, and edit or delete content inside the folder.<br><br>**Note:** Members with this permission cannot change access settings, create SMB shares from the mounted directory, or connect or unmount storage. They also cannot rename, move, or delete the top-level Common folder or the mount point itself. |
   | <nobr>**Default (No access)**</nobr> | Restore the member’s default permission: **No access**. |

5. Click **Confirm**.
6. Ask the member to refresh Files and open the folder. Check that they can browse its contents. With **Can view**, they should be unable to upload or delete files.

## Change or revoke access

1. Right-click the same folder and select **Access settings**.
2. Change the user's permission. To revoke a member's access, select **No access**.
3. Click **Confirm**.
4. Ask the member to refresh Files and check that the new permission applies.

Revoking access does not delete the folder or its contents. It does not remove copies the member has already downloaded.

:::info Default depends on the role
For an admin, **Default** restores management access. To restrict an admin's access, the Super Admin must select an explicit permission such as **Can view** or **No access**.
:::

## FAQs

### How do folder access settings, Internal share, and SMB share differ, and when should I use each?

Choose based on where the folder is and how people will access it:

- **Access settings**: Let Olares members view or edit a Common top-level folder or an External mount through Files. Use **Access settings** on the folder. For example, use this to give selected members access to a mounted NAS folder.
- **Internal share**: Share a Home or Sync folder with other Olares members. Use this when you want to invite members to a folder in either location. Common and External do not support Internal share.
- **SMB share**: Let computers access a Home or External folder over the local network using SMB, such as through Windows File Explorer or macOS Finder. When authentication is required, use dedicated SMB accounts, which are separate from Olares member accounts.

Creating an SMB share makes a folder available to other computers. [Mounting an SMB share](mount-SMB.md) connects a folder from a NAS or server to Olares. For sharing steps, see [Share files](share-files.md).

### Why can’t a member access or edit a folder?

Members cannot change these settings themselves. If a member reports a problem, sign in as the Super Admin or an admin and check the following:

| Problem | What to check |
| --- | --- |
| A folder is missing, or the member sees “No folders you can access.” | Check the member's permission on the top-level folder or mount. |
| The member can read but cannot upload or edit. | Check whether the permission is **Can view**. For NAS storage, also check that the NAS share allows writing. |
| **Access settings** is missing or<br> unavailable. | Make sure you selected a Common top-level folder or an External mount, and that your account can manage that folder. |
| Access stopped after an upgrade or a storage change. | Check that the expected storage is mounted. Check the member’s access settings and grant access again if needed. |
