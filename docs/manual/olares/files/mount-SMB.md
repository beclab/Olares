---
outline: [2, 3]
description: Learn how to mount and access SMB shared folders from NAS devices or network servers directly in Olares.
---

# Mount SMB shares

SMB (Server Message Block) is a protocol used to share files, printers, and other resources over a network. If you have a network-attached storage (NAS) device or another SMB server on your local network, you can easily mount SMB shares in Olares to access and manage your shared files.

## Before you begin

- Ensure Olares and the SMB server are on the same local network.
- You have obtained the following details:
  - The SMB share path, which is typically in the format of `//<IP-address>/<Shared-folder-name>`.
  - The username and password required to access the SMB share.

## Mount an SMB share

1. Open the Files app, and then go to **Drive** > **External**.
2. Click **Connect to server** in the upper-right corner.
3. In the popup window, click the **Protocol** list, and then select **SMB**.

    ![Configure SMB connection](/images/manual/olares/add-SMB-share-path.png#bordered)

4. In the **Server address** field, enter the SMB share path. For example, `//192.168.1.100/Documents`.
5. (Optional) Save frequently used server addresses for quick access next time:

    - To add an address to **Favorite servers**, click <i class="material-symbols-outlined">add</i> after entering the share path.
    - To remove a saved address, click it in **Favorite servers**, and then click <i class="material-symbols-outlined">remove</i>.

6. Click **Confirm**.
7. Enter the username and password, and then click **Confirm**.

    Once connected, the SMB share appears in the **External** directory, and you can access your shared files and folders seamlessly.

## Unmount an SMB share

When you no longer need access to the network files, you can safely disconnect the share.

1. Open the Files app, and then go to **Drive** > **External**.
2. Right-click the mounted folder, and then select **Unmount**. 

    The SMB share is disconnected from Olares immediately, and is removed from the **External** directory.
