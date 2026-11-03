---
outline: [2, 3]
description: Share files in Olares through internal, public, or SMB sharing. Configure access permissions and collaborate securely with teams or over a LAN.
head:
  - - meta
    - name: keywords
      content: Olares, Files, share files, public share, internal share, SMB share
---

# Share files

Share folders with Olares members, send password-protected links, or provide access over SMB. Choose a sharing method based on the folder location and who needs access.

## Share types and support scope

Sharing is available for folders only. In Olares 1.12.8, each location supports the following sharing methods:

| Directory type | Public share | Internal share | SMB share |
| :--- | :---: | :---: | :---: |
| **Home** | Supported | Supported | Supported |
| **External** | - | - | Supported |
| **Sync** | - | Supported | - |
| **Data**, **Common**, **Cache**, **Cloud drive** | - | - | - |

- **Public share**: Generates an access link with password protection supported, suitable for external sharing.
- **Internal share**: Restricted to members within the same Olares cluster, suitable for internal team collaboration. For more information, see [Manage your team](../settings/manage-team.md).
- **SMB share**: Shared via the SMB protocol within the Local Area Network (LAN).

![Share file type](/images/manual/olares/share-files.png#bordered)

## Permissions for Internal and SMB shares

Internal shares grant access to Olares members. SMB shares use separate SMB accounts. Permissions assigned to a share apply to access through that share.

To control member access to a Common folder or an External mount in Files, use [Access settings](manage-common-external-access.md). These locations do not support Internal share.

| Permission | Allowed actions |
| :--- | :--- |
| **Admin** | View, download, upload, edit, and delete files. For Internal shares, also manage share members with lower permissions.<br><br>**Note:** This permission does not grant an Olares admin role or allow changes to Common or External access settings. |
| **Edit** | View, download, upload, edit, and delete files. |
| **View** | View and download files. Uploading, editing, and deleting are not allowed. |

## Create share

### Create Internal share

1. Right-click a folder in **Home** or **Sync**, and then click **Share to** > **Internal**.
2. In the **Invite users** section, click <i class="material-symbols-outlined">add</i>, search for and select the target user or group, and then click **Invite**.

![Invite users for Internal Share](/images/manual/olares/internal-share-invite.png#bordered){width=60%}

3. In the **Set user permissions** list, click <i class="material-symbols-outlined">chevron_forward</i> to the right of the user avatar to assign specific permissions, and then click **Submit**.
4. Click **Confirm**.

### Create SMB share

SMB sharing lets computers access folders through Windows File Explorer or macOS Finder over the local network. It uses dedicated SMB accounts that are separate from Olares member accounts.

You can create SMB shares for folders in **Home** or **External**. For External, log in as the Super Admin or an admin with management access to the mounted directory. Select the mounted directory or a folder inside it, not the External root.

To connect a folder from a NAS or server to Olares instead, see [Mount SMB shares](mount-SMB.md).

1. Right-click the target folder in **Home** or **External**, and then click **Share to** > **SMB**.
2. Select whether to make it **Public**.
    - If you selected **Yes**, click **Confirm**, and then it will be accessible to anyone on the LAN.
    - If you selected **No**, authentication is required. You must select or create a dedicated SMB account by using Step 3 or Step 4.
3. Select an existing SMB account to authorize.
    
    a. In the **Invite users** field, click <i class="material-symbols-outlined">add</i>.

    ![Invite existing users for SMB Share](/images/manual/olares/smb-share-invite.png#bordered){width=60%}
    
    b. Search for the SMB account, select it, and then click **Invite**.

    c. In the **Set user permissions** list, click <i class="material-symbols-outlined">chevron_forward</i> to the right of the user avatar to assign specific permissions, and then click **Submit**.

    d. Click **Confirm**.

4. Create a new SMB account to authorize.
    
    a. To the right of **Invite users**, click **Add user account**.

    ![Create new users for SMB Share](/images/manual/olares/smb-share-new-usr.png#bordered){width=60%}
    
    b. Enter a unique user name.
    
    c. In the **Set password** field, a secure system-generated password is provided by default. You can also manually enter a new password or click **Regenerate** to create a new random one.

    d. Click **Confirm**. The system will automatically generate this account for you.
    
    e. In the **Invite users** field, click <i class="material-symbols-outlined">add</i>.
    
    f. Search for the newly created account, select it, and then click **Invite**.

    g. In the **Set user permissions** list, click <i class="material-symbols-outlined">chevron_forward</i> to the right of the user avatar to assign specific permissions, and then click **Submit**.

    h. Click **Confirm**.

5. In the **SMB share** dialog, click **Copy link** to copy the access link.

### Create Public share

1. Right-click a folder in **Home**, and then click **Share to** > **Public**.

![Public share settings](/images/manual/olares/public-share-settings.png#bordered){width=56%}

2. In the **Set password** field, a secure system-generated password is provided by default. You can also manually enter a new password or click **Regenerate** to create a new random one.
3. In the **Set expiration** region, select the specific expiration time for the share link.
4. Select the following restrictions as needed:
     - **Limit file size**: To limit the size of files uploaded by visitors, select this checkbox, enter a value, and then select the unit (Mi/Gi/Ti).
    - **Allow uploads only**: If you select this checkbox, visitors can only upload files and cannot view or download existing content.
5. Click **Confirm** to generate the link.
6. Copy or note down the link address, and then click **Confirm**.

## Manage shares

You can view and manage all shared folders in the **Share** list.

### View share list and attributes

1. Open the Files application from the Dock or Launchpad on Olares.
2. Click **Share** in the left sidebar.
3. You can view the list of all currently shared folders on the right.
4. If you want to view the details of a specific folder, right-click the target folder and click **Properties**. 

    You will see the sharer, current path, original path, share scope, owner, current permissions, expiration date, and link details.

### Modify password or permissions

As the share creator, you can reset a Public share password or change who can access an Internal or SMB share.

1. Open the Files application from the Dock or Launchpad on Olares.
2. Click **Share** in the left sidebar to view the **Share** list on the right.
3. To reset password (**Public** only):
    
    a. Right-click the folder and select **Reset Password**.
    
    b. Enter a new password or click **Regenerate**.
    
    c. Click **Confirm**.
4. To edit permissions (**Internal** and **SMB** only):
    
    a. Right-click the folder and select **Edit permissions**.

    b. Add or remove users, or modify permissions for existing users.

    c. Click **Confirm**.

### Cancel share

If you no longer need to share a folder, you can cancel the share at any time.

1. Open the Files application from the Dock or Launchpad on Olares.
2. Click **Share** in the left sidebar to open the share list.
3. Right-click the target folder and click **Revoke sharing**.
4. Click **Confirm**.

## FAQs

### Why is a share unavailable?

A share can remain in the list even when it no longer provides access. Check these possible causes:

| Cause | What to do |
| --- | --- |
| The creator of an External SMB share lost management access to the mounted directory. | Ask the Super Admin to check the creator's access to that directory. Restoring the required access allows the share to become available again. |
| A share created before upgrading uses a sharing method that is no longer supported, such as an Internal share from External or a share from Common. | Cancel the unsupported share and use a supported method. Granting permissions cannot restore an unsupported share. To give Olares members access to Common or External, use [Access settings](manage-common-external-access.md). |

### What happens to shares when I rename, move, or delete a folder?

If you rename, move, or delete a folder that is shared, or contains shared folders, Files asks you to confirm that the affected shares will be canceled.

- Continue to perform the operation and cancel the affected shares.
- Cancel to keep the folder and its shares unchanged.

After renaming or moving the folder, create new shares if needed. Existing shares do not follow the folder to its new name or location.
