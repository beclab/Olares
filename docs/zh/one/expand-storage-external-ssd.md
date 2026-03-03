---
outline: [2, 3]
description: Learn how to manually mount an external SSD to Olares One for temporary or permanent storage expansion.
head:
  - - meta
    - name: keywords
      content: Olares One, external SSD, expand storage
---

# Expand storage via external SSD <Badge type="tip" text="30 min" />

You can manually mount high-capacity external SSDs to specific system paths on Olares One.

This approach is recommended for long-term storage expansion, such as downloading more or larger local AI models.

:::warning HDD support
This guide is intended for SSDs. Mechanical Hard Disk Drives (HDDs) have not been tested with Olares One.
:::
:::info Mount path
Currently, only mounts under the `/olares/share` directory are supported.

Mounting flexibility will be improved in future versions.
:::

## Prerequisites
**Hardware**
- Your Olares One is set up and running.
- The external SSD is connected to Olares One.

**SSH access**
- [SSH access to Olares One](access-terminal-ssh.md).

**Experience**
- Basic familiarity with terminal commands and the command-line interface (CLI).

## Step 1: Identify the drive

1. Connect to your Olares One terminal via SSH or from the Control Hub.

2. Run the following command to view detected drives:

   ```bash
   sudo fdisk -l
   ```

3. Identify your target drive from the output. Each drive lists its partitions under the **Device** column, such as `/dev/nvme1n1p1`, `/dev/nvme1n1p2`, or `/dev/sdb1`.

   ![Partition list](/images/manual/tutorials/expand-storage-partition.png#bordered)

4. Note the target partition you intend to mount. For example: `/dev/nvme1n1p1`.

## Step 2: Mount the partition
### Option A: Temporarily mount a partition

Temporary mounting is suitable for one-time tasks such as data migration. The configuration will be lost if the device reboots.

1. Create a directory for the mount point:

    ```bash
    sudo mkdir -p /olares/share/<directory_name>
    ```

   Replace `<directory_name>` with a custom name.

2. Mount the partition to this directory:

    ```bash
    sudo mount /dev/<partition> /olares/share/<directory_name>    
    ```

   For example:

    ```bash
    sudo mount /dev/nvme1n1p1 /olares/share/hdd0
    ```

3. Navigate to the **External** directory in Files to verify the mount. You should see the new folder content.

   ![Check mount result](/images/manual/tutorials/expand-storage-mount-result-en.png#bordered)

### Option B: Permanently mount a partition
For long-term usage, you must configure the system to mount the drive automatically at boot using the `/etc/fstab` file.

1. Get the UUID. 
   :::tip Use UUID to identify device
   Using the UUID is safer than using device names (like `/dev/sdb1`), which can change if you plug drives into different ports.
   :::
   a. Run the following command:
    ```bash
    lsblk -f
    ```
   b. Note down following information:
    - **FSTYPE**: File system type (e.g., `ext4`, `xfs`).
    - **UUID**: Unique identifier of the partition.

   ![Check mount result](/images/manual/tutorials/expand-storage-fstype.png#bordered)

2. Create the mount directory:

    ```bash
    sudo mkdir -p /olares/share/<directory_name>
    ```

   Replace `<directory_name>` with a custom name.

3. Open the configuration file.

    ```bash
    sudo vi /etc/fstab
    ```

4. Add the mount entry. Add a new line at the end of the file using the following format:

    ```bash
    UUID=<UUID> /olares/share/<directory_name> <FSTYPE> defaults,nofail 0 0
    ```

    For example:

    ```bash
    UUID=1234-ABCD /olares/share/my_disk ext4 defaults,nofail 0 0
    ```

5. Press `Esc`, type `:wq`, and hit **Enter** to save your changes and exit the editor.

6. Run the following command to verify the configuration.

    ```bash
    mount -a
    ```
   :::tip Prevent boot failure
   An incorrect `/etc/fstab` configuration might prevent your system from booting.

   It is strongly recommended to run `mount -a` first to validate the configuration before rebooting.
   :::
   If no errors appear, the setup is successful.

7. After reboot, confirm the drive is automatically mounted in the **External** directory.

## Step 3: Unmount a partition
:::warning Irreversible operation
Ensure no programs or terminals are accessing the directory before unmounting.
:::

To safely remove the drive or delete the mount point configuration:
1. Unmount the partition:

    ```bash
    sudo umount /olares/share/<directory_name>
    ```

2. (Optional) Remove the directory If you no longer need the folder:

    ```bash
    rm -rf /olares/share/<directory_name>
    ```

## Resources
- [Manage files in Olares](../manual/olares/files/index.md)
- [Expand storage via USB devices](expand-storage-usb-drive.md)
- [Expand storage using internal SSD](expand-storage-internal-ssd.md)