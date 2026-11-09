---
outline: [2, 3]
description: Mount and access NFS shared directories from NAS devices or servers on Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Files, NFS share, NAS, network storage, mount
---

# Mount NFS shares

Network File System (NFS) is a widely used protocol for accessing files over a network. If you have an NFS server on your local network, you can easily mount NFS shares to Olares and access and manage shared files as if they were local folders.

## Before you begin

:::warning Account requirements
Log in to Olares as the Super Admin or an admin to connect an NFS share. Members cannot connect servers.
:::

- Ensure Olares and the NFS server are on the same local network with stable connectivity.
- You have the following information:
  - **NFS server IP address**: For example, `192.168.1.100`.
  - **NFS share directory path**: For example, `/volume1/share` or `/export/data`.

## Mount an NFS share

1. Log in to Olares as the Super Admin or an admin, open **Files**, and go to **Drive** > **External**.
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

7. Once connected, the NFS share appears under **External**. Open it to check that you can browse the shared files.

## Give members access

In Olares 1.12.8, members have no access to mounted directories by default. To let a member view or edit the files, use **Access settings** on the mounted directory. See [Manage access to Common and External](manage-common-external-access.md) for the steps.

## Unmount an NFS share

Log in to Olares as the Super Admin or an admin with management access to the mounted directory. Members cannot unmount storage, even if they have edit access.

Unmounting disconnects the share from Olares for everyone using that mount. It does not delete files on the NFS server.

1. Open the Files app, and then go to **Drive** > **External**.
2. Right-click the NFS share directory you want to unmount, and then select **Unmount**.

   The share is immediately disconnected from Olares and removed from the **External** directory.
