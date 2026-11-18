---
outline: [2, 3]
description: Learn how to mount and access SMB shared folders from NAS devices or network servers directly in Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Files, SMB share, NAS, network storage, mount
---

# Mount SMB shares

SMB (Server Message Block) is a protocol used to share files, printers, and other resources over a network. If you have a network-attached storage (NAS) device or another SMB server on your local network, you can easily mount SMB shares in Olares to access and manage your shared files.

## Before you begin

:::warning Account requirements
Log in to Olares as the Super Admin or an admin to connect an SMB share. Members cannot connect servers.
:::

- Ensure Olares and the SMB server are on the same local network.
- You have obtained the following details:
  - The SMB share path, which is typically in the format of `//<IP-address>/<Shared-folder-name>`.
  - The username and password required to access the SMB share.

## Save SMB accounts

To keep SMB credentials in one place, save them in **Settings** > **Integration** > **SMB account management**.

1. To view saved accounts, go to **Settings** > **Integration** > **SMB account management**.
2. To add an account, click **Add account**, enter the username and password for the SMB share, and click **Confirm**.
3. To delete an account, click <i class="material-symbols-outlined">delete</i> on the right of the record, and then click **Confirm**.

:::tip
- SMB credentials are stored locally and are not uploaded to the cloud.
- Saved accounts are for reference only. Files does not list them in the mount dialog, so you still need to enter the username and password manually when mounting a share.
:::

## Mount an SMB share

1. Log in to Olares as the Super Admin or an admin, open **Files**, and go to **Drive** > **External**.
2. Click **Connect to server** in the upper-right corner.
3. In the popup window, click the **Protocol** list, and then select **SMB**.

    ![Configure SMB connection](/images/manual/olares/add-SMB-share-path.png#bordered)

4. In the **Server address** field, enter the SMB share path. For example, `//192.168.1.100/Documents`.
5. (Optional) Save frequently used server addresses for quick access next time:

    - To add an address to **Favorite servers**, click <i class="material-symbols-outlined">add</i> after entering the share path.
    - To remove a saved address, click it in **Favorite servers**, and then click <i class="material-symbols-outlined">remove</i>.

6. Click **Confirm**.
7. Enter the username and password, and then click **Confirm**.

    Once connected, the SMB share appears under **External**. Open it to check that you can browse the shared files.

## Give members access

In Olares 1.12.8, members have no access to mounted directories by default. To let a member view or edit the files, use **Access settings** on the mounted directory. See [Manage access to Common and External](manage-common-external-access.md) for the steps.

## Unmount an SMB share

Log in to Olares as the Super Admin or an admin with management access to the mounted directory. Members cannot unmount storage, even if they have edit access.

Unmounting disconnects the share from Olares for everyone using that mount. It does not delete files on the NAS.

1. Open the Files app, and then go to **Drive** > **External**.
2. Right-click the mounted folder, and then select **Unmount**.

    The SMB share is disconnected from Olares immediately, and is removed from the **External** directory.
