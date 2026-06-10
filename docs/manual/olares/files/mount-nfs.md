---
outline: [2, 3]
description: Learn how to mount and access NFS shared directories from NAS devices or servers in Olares. Includes detailed steps for connecting NFS shares and managing network files.
---

# Mount NFS shares

Network File System (NFS) is a widely used protocol for accessing files over a network. If you have an NFS server on your local network, you can easily mount NFS shares to Olares and access and manage shared files as if they were local folders.

## Before you begin

- Ensure Olares and the NFS server are on the same local network with stable connectivity.
- You have the following information:
  - **NFS server IP address**: For example, `192.168.1.100`.
  - **NFS share directory path**: For example, `/volume1/share` or `/export/data`.

:::info
Most home or enterprise NFS servers use IP-based access control and do not require additional authentication. If your NFS server is configured with advanced authentication such as Kerberos, you will need to use the hostname instead of the IP address for mounting, and additional configuration is required. This guide only covers the most common unauthenticated NFS mounting method.
:::

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

## Check NFS share status

Due to network fluctuations or temporary unavailability of the NFS server, a mounted share might become unavailable in the **External** directory, and no file operations can be performed.

In this case:
- Right-click the directory and select **Refresh** to recheck the connection status, or wait for the network to recover and the system will automatically mark it as available.
- If the directory remains unavailable for a long time, it is recommended to manually unmount and remount it.

:::info
To prevent silent disconnections and potential data loss, Olares does not automatically unmount unavailable NFS shares. You must perform all unmount operations manually.
:::

## Unmount an NFS share

If you no longer need to access the NFS share, you can safely unmount it.

1. Open the Files app, and then go to **Drive** > **External**.
2. Right-click the NFS share directory you want to unmount, and then select **Unmount**.

   The share is immediately disconnected from Olares and removed from the **External** directory.

## Frequently asked questions

### Why doesn't entering the IP address list any shares?

Possible causes include:
- The Olares node IP is not allowed by the NFS server. Check the `/etc/exports` configuration on the NFS server to ensure it includes the Olares node IP or subnet.
- The NFS service is not started or a firewall is blocking the connection. Check the `rpcbind` and `nfs-server` service status on the NFS server.
- The server address was entered incorrectly, or the network is unreachable.

### Does the NFS share require a username and password?

Usually not. Most NFS deployments use IP/hostname-based access control and do not involve usernames or passwords. If your environment is configured with authentication mechanisms such as Kerberos, you will need to use the hostname (not IP) for mounting and configure Kerberos credentials in advance. For such advanced scenarios, please contact your system administrator.

### Will previously mounted NFS shares reconnect automatically after restarting the Olares node?

Currently, Olares will attempt to automatically remount previously configured NFS shares after a restart. If automatic mounting fails (for example, the server is not powered on), the share will appear as "unavailable" under "External devices". You can manually refresh or remount it later.

### What is the difference between NFS shares and SMB shares in usage?

Basic operations are the same, and both allow you to manage files like a local folder. The main differences are:
- **Connection method**: NFS does not require a username or password, only the server address and share path.
- **Protocol characteristics**: NFS performs better in Unix/Linux environments, with better file permissions and POSIX compatibility.
- **Status management**: Both support unavailable status detection and frontend prompts to avoid operation hangs.
