---
outline: [2, 3]
description: Mount and access NFS shared directories from NAS devices or servers on Olares.
---

# Mount NFS shares

Network File System (NFS) is a widely used protocol for accessing files over a network. If you have an NFS server on your local network, you can easily mount NFS shares to Olares and access and manage shared files as if they were local folders.

## Before you begin

- Ensure Olares and the NFS server are on the same local network with stable connectivity.
- You have the following information:
  - **NFS server IP address**: For example, `192.168.1.100`.
  - **NFS share directory path**: For example, `/volume1/share` or `/export/data`.

## Mount an NFS share

1. Open the Files app, and then go to **Drive** > **External**.
2. Click **Connect to server** in the upper-right corner.
3. In the popup window, click the **Protocol** list, and then select **NFS**.

   ![Select NFS protocol](/images/manual/olares/files-connect-nfs-protocol.png#bordered)

4. In the **Server address** field, enter the NFS share address. Two input formats are supported:

   - **IP address**: For example, `192.168.1.100`. The system will automatically try to connect to the server and list the available shares for you to select.
   - **IP address with path**: For example, `192.168.1.100:/volume1/share`. This mounts the specified path directly, skipping the directory selection step.

5. (Optional) Click <i class="material-symbols-outlined">add</i> in the lower-left corner to add the current server address to **Favorite servers** for quick access next time.

6. Click **Confirm**.

   - If you entered only the IP address, the system will query all available NFS shares on the server. In the popup window, select the share you want to mount from the list, and then click **Confirm**.
   - If you entered the full path, the system will try to mount the directory directly.

7. After the connection is successful, the NFS share appears in the **External** directory. You can browse and manage files in it just like a local folder.

## Unmount an NFS share

If you no longer need to access the NFS share, you can safely unmount it.

1. Open the Files app, and then go to **Drive** > **External**.
2. Right-click the NFS share directory you want to unmount, and then select **Unmount**.

   The share is immediately disconnected from Olares and removed from the **External** directory.
