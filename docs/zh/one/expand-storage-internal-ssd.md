---
outline: [2, 3]
description: Learn how to expand storage on Olares One using an NVMe M.2 SSD.
head:
  - - meta
    - name: keywords
      content: Olares One, NVMe SSD, expand storage, LVM, olares-cli
---

# Expand storage with an NVMe M.2 SSD <Badge type="tip" text="15 min" />

If you have installed a secondary internal NVMe SSD in your Olares One, you can merge it into the main system storage using the `olares-cli`.

Unlike external drives which are mounted as separate folders, this method uses Logical Volume Management (LVM) to seamlessly extend your root file system.


## Before you begin
:::warning One-time irreversible operation
- This is a one-time operation. The SSD becomes an integral part of the system volume and cannot be easily separated later.
- If you physically uninstall this SSD after extension, the Olares file system will be incomplete, causing the OS to crash or fail to boot. You will need to manually reinstall Olares OS.
  :::

The Olares One motherboard includes two PCIe SSD slots:
* Slot 1 (PCIe 4.0): Occupied by the preinstalled 2TB system SSD.
* Slot 2 (PCIe 5.0): Available for expansion.

You can use this second slot to expand your current system storage.

## Prerequisites
**Hardware**
- Your Olares One is set up and running.
- A secondary NVMe M.2 SSD is physically installed in Olares One.

**SSH access**
- [SSH access to Olares One](access-terminal-ssh.md).

**Experience**
- Basic familiarity with terminal commands and the command-line interface (CLI).

## Step 1: Identify the unmounted disk

1. Connect to your Olares One terminal via SSH or from the Control Hub.

2. List block devices on the host:

    ```bash
    lsblk | grep -v loop
    ```

3. Check the size and mount points to identify the new disk.

   Example output:
   
   ```text
   NAME        MAJ:MIN RM   SIZE RO TYPE MOUNTPOINTS
   sda           8:0    0 931.5G  0 disk
   ├─sda1        8:1    0   512M  0 part /boot
   └─sda2        8:2    0   931G  0 part /
   nvme1n1     259:3    0 931.5G  0 disk
   ```
In this example, `sda` is the system drive which is mounted at `/` and `/boot`, while `nvme1n1` is the new, unmounted SSD.

## Step 2: Extend system storage

1. Verify that Olares recognizes the unmounted disk:

    ```bash
    olares-cli disk list-unmounted
    ```

2. Run the extension command:

    ```bash
    olares-cli disk extend
    ```

3. When the command prompts for confirmation, type `YES` to proceed.
    ```bash
    WARNING: This will DESTROY all data on /dev/<device>
    Type 'YES' to continue, CTRL+C to abort:
    ```

   Example output:
    ```bash
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
## Step 3: Verify the extension

You can verify the storage increase from Control Hub (UI), or via the command line.

<tabs>
<template #Via-Dashboard-(UI)>

1. Open Dashboard.
2. Check the Disk section to confirm the total system storage capacity has increased.

   ![Check disk volume in Dashboard](/public/images/manual/tutorials/expand-dashboard-disk.png#bordered)
</template>

<template #Via-command-line>

1. Check the total size of the `/olares` directory:

    ```bash
    df -h /olares
    ```

    Example output:
    ```text
    Filesystem                  Size   Used  Avail Use% Mounted on
    /dev/mapper/olares--vg-root 1.8T   285G   1.4T  17% /olares
    ```
   The **Size** column should now reflect the combined capacity.
2. Confirm the disk structure:
    ```bash
    lsblk | grep -v loop
    ```
   Example output:

    ```bash
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
You should see your new disk (e.g., `nvme0n1`) listed under the `olares--vg-data` group, sharing the same mount point as your primary data partition.

</template>
</tabs>


## Resources
- [Expand storage via USB devices](expand-storage-usb-drive.md)
- [Expand storage using external SSD](expand-storage-external-ssd.md)
- [`olares-cli disk`](../developer/install/cli/disk.md).