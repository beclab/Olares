---
outline: [2, 3]
description: Complete guide to expanding storage in Olares. Learn how to connect to SMB servers, use USB auto-mount, use CLI commands, and manually mount HDDs or SSDs to increase local storage capacity and manage large AI model files efficiently.
---
# Expand storage in Olares

This document describes how to expand storage in Olares using different approaches. Choose the method that best matches your scenario:

- **Connect to an SMB server** to access shared files on a NAS or another computer over the network.
- **Use USB auto-mount** for plug-and-play external storage. No Linux commands required.
- **Manually mount an HDD or SSD** under `/olares/share` to keep a disk as independent external storage for large files.
- **Expand system storage via Olares CLI** (`disk extend`) to increase system capacity on LVM-based setups by merging new disk(s) into the system volume.

## Connect to an SMB server

You can easily mount Server Message Block (SMB) shares in Olares to access and manage shared files.

1. On the Olares web interface, navigate to **Files** > **External** > **Connect to server**.

2. Enter the server address (for example: `//192.168.1.10/shared`) and click **Confirm**.

3. Once connected, you can access the shared directory under **Files** > **External**.

For details, please refer to [Mount SMB shares](../olares/files/mount-SMB.md).

## Expand storage via USB devices

**Olares** automatically detects and mounts inserted USB storage devices.

- Once you insert a USB device, it will be mounted automatically — no command-line operations are required.

- You can access it in **Files** > **External** from both Olares and Larepass.

- The system automatically unmounts the device when you unplug it.

- You can manually eject the device via the Olares web interface. Right-click the USB drive in Files and select **Unmount**.

## Manually mount an HDD or SSD

You can manually mount an HDD or SSD to Olares from your **Linux** hosting system.

This approach is recommended for **large data storage** (e.g., AI models) or **long-term storage expansion**.

### Before you begin

Please ensure the following:

- You have **administrator (sudo) privileges** on your Linux system.

- **Olares** is properly installed and running.

- The target drive is **already formatted** (recommended file systems: `ext4` or `XFS`).

:::tip  Mount path restriction

Currently, only mounts under the `/olares/share` directory are supported.

Mounting flexibility will be improved in future versions.
:::

### Identify the drive

1. Insert the hard drive into the host machine.

2. Run the following command to view detected drives:

   ```bash
   fdisk -l
   ```

3. Identify the target drive type and device name from the output:

    - **NVMe SSD**: typically appear as `/dev/nvme0n1`, `/dev/nvme1n1`, etc.
    - **SATA** or **HDD**: typically appear as `/dev/sda`, `/dev/sdb`, etc.

    Each drive lists its partitions under the Device column, such as `/dev/nvme1n1p1`, `/dev/nvme1n1p2`, or `/dev/sdb1`.
    ![Partition list](/images/manual/tutorials/expand-storage-partition.png#bordered)
   

4. Confirm the target partition to mount (e.g., `/dev/nvme1n1p1`).

### Temporarily mount a partition

Temporary mounting is suitable for **one-time** or **short-term** use (e.g., file transfer).

The mount configuration will be lost after a Linux or Olares reboot.

1. Create a mount directory:

    ```bash
    sudo mkdir -p /olares/share/<directory_name>
    ```

    Replace `<directory_name>` with a custom name.

2. Mount the partition:

    ```bash
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

1. Run the following command to list all drives and find the target partition:

    ```bash
    lsblk -f
    ```

    Record the following information:
    - **FSTYPE**: File system type (e.g., `ext4`, `xfs`).
    - **UUID**: Unique identifier of the partition.

    ![Check mount result](/images/manual/tutorials/expand-storage-fstype.png#bordered)

2. Create a mount directory:
    
    ```bash
    sudo mkdir -p /olares/share/<directory_name>
    ```

    Replace `<directory_name>` with a custom name.

3. Edit the mount configuration file:
    
    ```bash
    sudo vi /etc/fstab
    ```

4. Add a mount entry using **UUID** (recommended to prevent issues if device names change):

    ```
    UUID=<UUID> /olares/share/<directory_name> <FSTYPE> defaults,nofail 0 0
    ```

    **Example**:

    ```
    UUID=1234-ABCD /olares/share/my_disk ext4 defaults,nofail 0 0
    ```

5. Save and exit the editor.

6. Verify the configuration (recommended):

    ```bash
    mount -a
    ```
    
    If no errors appear, the setup is successful.

7. After reboot, confirm the drive is automatically mounted via **Files** > **External**.

    :::warning
    An incorrect /etc/fstab configuration may prevent your system from booting.
    It is strongly recommended to run `mount -a` first to validate the configuration before rebooting.
    :::

### Unmount a partition

You can unmount partitions mounted using either temporary or permanent methods.

1. Unmount the partition:

    ```bash
    sudo umount /olares/share/<directory_name>
    ```

    :::tip NOTE
    Make sure no programs or terminals are accessing the directory before unmounting.
    :::

2. Remove the empty directory (optional):

    ```bash
    rm -rf /olares/share/<directory_name>
    ```

    :::warning
    Ensure the directory is empty and fully unmounted before deleting.
    :::

    You can also view and remove this directory from **Files** in Olares.

## Expand system storage via Olares CLI

If your Olares system uses LVM-based storage, you can expand its system storage capacity using the `disk` command.

Manual mounting adds an external drive under `/olares/share`. In contrast, `disk extend` expands Olares system storage. After extension, the added drive is no longer shown as an independent mount point.

:::warning Data loss
`disk extend` will destroy all data on the selected disk.  
Make sure the disk does not contain important data, or back up the data before continuing.
:::

### Before you begin

- Connect the external drive to the Olares host machine.
- [SSH](/developer/install/cli/access-olares-terminal.md) into the Olares terminal.

### Identify the unmounted disk

List block devices on the host:

```bash
lsblk | grep -v loop
```

Identify the newly added disk by checking its size and confirming it has no mount points. Do not select the disk that contains `/` or `/boot`.

**Example output**:

```text
NAME        MAJ:MIN RM   SIZE RO TYPE MOUNTPOINTS
sda           8:0    0 931.5G  0 disk
├─sda1        8:1    0   512M  0 part /boot
└─sda2        8:2    0   931G  0 part /
nvme1n1     259:3    0 931.5G  0 disk
```
In this example, `sda` is the system drive which is mounted at `/` and `/boot`, while `nvme1n1` is the newly connected disk.

### Extend system storage

1. Verify that Olares recognizes the unmounted disk:

    ```bash
    olares-cli disk list-unmounted
    ```

2. Add the disk to the system volume:

    ```bash
    olares-cli disk extend
    ```

3. Type `YES` to proceed when the command prompts for confirmation.
    ```text
    WARNING: This will DESTROY all data on /dev/<device>
    Type 'YES' to continue, CTRL+C to abort:
    ```

    **Example output**:
    ```text
    Selected volume group to extend: olares-vg
    Selected logical volume to extend: data
    Selected unmounted device to use: /dev/nvme0n1
    Extending logical volume data in volume group olares-vg using device /dev/nvme0n1
    WARNING: This will DESTROY all data on /dev/nvme0n1
    Type 'YES' to continue, CTRL+C to abort: YES
    Selected device /dev/nvme0n1 has existing partitions. Cleaning up...
    Deleting existing partitions on device /dev/nvme0n1...
    Creating partition on device /dev/nvme0n1...
    Creating physical volume on device /dev/nvme0n1...
    Extending volume group olares-vg with logic volume data on device /dev/nvme0n1...
    Disk extension completed successfully.

    id  LV    VG         LSize    Mountpoints
    1   data  olares-vg  <3.63t   /var,/olares
    2   root  olares-vg  100.00g  /
    3   swap  olares-vg  1.00g
    ...
    ```
### Verify the extension

You can verify the storage increase in both terminal and UI.

#### In terminal

- Check the size of the `/olares` directory where data is stored to confirm expansion:

    ```bash
    df -h /olares
    ```

    **Example output**:
    ```text
    Filesystem                  Size   Used  Avail Use% Mounted on
    /dev/mapper/olares--vg-root 1.8T   285G   1.4T  17% /olares
    ```

- Confirm if the new disk is now part of the `olares--vg-data` volume:
    ```bash
    lsblk | grep -v loop
    ```
    **Example output**:
    ```text
    NAME                MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
    nvme0n1             259:0    0  1.9T  0 disk
    └─nvme0n1p1         259:2    0  1.9T  0 part
      └─olares--vg-data 252:2    0  3.6T  0 lvm  /olares /var
    nvme1n1             259:3    0  1.9T  0 disk
    ├─nvme1n1p1         259:4    0  512M  0 part /boot/efi
    └─nvme1n1p2         259:5    0  1.9T  0 part
      ├─olares--vg-root 252:1    0  100G  0 lvm  /
      └─olares--vg-swap 252:0    0    1G  0 lvm  [SWAP]
    ```

#### In UI
Open Dashboard from Launchpad and confirm that total system storage capacity has increased.

![Check disk volume in Dashboard](/public/images/manual/tutorials/expand-dashboard-disk.png#bordered)

For full command usage and options, please refer to [`disk`](/developer/install/cli/disk.md).