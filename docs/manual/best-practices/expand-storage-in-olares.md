---
outline: [2, 3]
description: Complete guide to expanding storage in Olares. Learn how to connect to SMB servers, use USB auto-mount, and manually mount HDDs or SSDs to increase local storage capacity and manage large AI model files efficiently.
---
# Expand storage in Olares

This document describes how to expand storage in Olares, including connecting to an SMB server, using automatically mounted USB storage devices, and manually mounting HDDs or SSDs in a Linux environment.

## Connect to an SMB server

Olares supports accessing local or remote file servers through the **SMB protocol**.

This method is ideal for **data sharing** and **team collaboration**.

1. On the Olares web interface, navigate to **Files** > **External** > **Connect to server**.

2. Enter the server address (for example: `//192.168.1.10/shared`) and click **Confirm**.

3. Once connected, you can access the shared directory under **Files** > **External**.

For details, please refer to [Mount SMB shares](../olares/files/mount-SMB.md).

## Expand storage via USB devices

The **Olares daemon** automatically detects and mounts inserted USB storage devices.

This method is ideal for **temporary data exchange** or **portable storage** use cases.

- Once a USB device is inserted, it will be mounted automatically — no command-line operations are required.

- You can access it from the Olares web interface or Larepass, via **Files** > **External**.

- When the USB device is unplugged, the system automatically unmounts it.

## Manually mount an HDD or SSD

You can manually mount internal or other types of non-USB drives to Olares from your **Linux** system.

This approach is recommended for **large data storage** (e.g., AI models) or **long-term storage expansion**.

### Before you begin

Please ensure the following:

- You have **administrator (sudo) privileges** on your Linux system.

- **Olares** is properly installed and running.

- The target drive is **already formatted** (recommended file systems: `ext4` or `XFS`).

:::tip  Mount path restriction

Currently, only mounts under the directory `/olares/share` are supported.

Mounting flexibility will be improved in upcoming versions.
:::


### Identify the drive

1. Insert the hard drive into the host machine.

2. Run the following command to view detected drives:

   ```
   fdisk -l
   ```

3. Identify the target drive type and device name from the output:

    - **NVMe SSD**: typically appear as `/dev/nvme0n1`, `/dev/nvme1n1`, etc.
    - **SATA** or **HDD**: typically appear as `/dev/sda`, `/dev/sdb`, etc.

    :::info
    Each drive lists its partitions below, such as `/dev/nvme1n1p1`, `/dev/nvme1n1p2`, or `/dev/sdb1`.
    :::

4. Confirm the target partition to mount (e.g., `/dev/nvme1n1p1`).

### Temporarily mount a partition

Temporary mounting is suitable for **one-time** or **short-term** use (e.g., file transfer).

The mount configuration will be lost after a system or Olares reboot.

1. Create a mount directory:

    ```
    sudo mkdir -p /olares/share/<directory_name>
    ```

    Replace `<directory_name>` with a custom name.

2. Mount the partition:

    ```
    sudo mount /dev/<partition> /olares/share/<directory_name>    
    ```

    **Example**:

    ```
    sudo mount /dev/nvme1n1p1 /olares/share/hdd0
    ```

3. Verify the mount result:

    After successful mounting, you can access the partition from **Files** > **External**.

    ![Check mount result](/images/manual/tutorials/expand-storage-mount-result-en.png#bordered)

### Permanently mount a partition

If you want the mount configuration to remain after reboot, configure **automatic mounting** in `/etc/fstab`.

1. Identify the target partition (refer to the steps in [Identify the drive](expand-storage-in-olares.md#identify-the-drive)).

2. Retrieve the partition's file system type and UUID:

    ```
    lsblk -f
    ```

    Record the following information:
    - **FSTYPE**: File system type (e.g., `ext4`, `xfs`).
    - **UUID**: Unique identifier of the partition.

    ![Check mount result](/images/manual/tutorials/expand-storage-fstype.png#bordered)

3. Create a mount directory:
    
    ```
    sudo mkdir -p /olares/share/<directory_name>
    ```

    Replace `<directory_name>` with a custom name.

4. Edit the mount configuration file:
    
    ```
    sudo vi /etc/fstab
    ```

5. Add a mount entry using **UUID** (recommended to prevent issues if device names change):

    ```
    UUID=<UUID> /olares/share/<directory_name> <FSTYPE> defaults,nofail 0 0
    ```

    **Example**:

    ```
    UUID=1234-ABCD /olares/share/my_disk ext4 defaults,nofail 0 0
    ```

6. Save and exit the editor.

7. Verify the configuration (recommended):

    ```
    mount -a
    ```
    
    If no errors appear, the setup is successful.

8. After reboot, confirm the drive is automatically mounted via **Files** > **External**.

    :::warning
    An incorrect /etc/fstab configuration may prevent your system from booting.
    It is strongly recommended to run `mount -a` first to validate the configuration before rebooting.
    :::

## Unmount a partition

You can unmount partitions mounted using either temporary or permanent methods.

1. Unmount the partition:

    ```
    sudo umount /olares/share/<directory_name>
    ```

    :::tip NOTE
    Make sure no programs or terminals are accessing the directory before unmounting, or the operation will fail.
    :::

2. Remove the empty directory (optional):

    ```
    rm -rf /olares/share/<directory_name>
    ```

    :::warning
    Ensure the directory is empty and fully unmounted before deleting, or data loss may occur.
    :::

You can also view and remove this directory from **Files** in the Olares web interface.