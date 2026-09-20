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

## Mount an SMB share## Unmount an SMB share

When you no longer need access to the network files, you can safely disconnect the share.

1. Open the Files app, and then go to **Drive** > **External**.
2. Right-click the mounted folder, and then select **Unmount**. 

    The SMB share is disconnected from Olares immediately, and is removed from the **External** directory.
