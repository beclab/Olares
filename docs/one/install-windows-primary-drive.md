---
outline: [2, 3]
description: Install Windows on the primary drive of Olares One by replacing Olares OS.
head:
  - - meta
    - name: keywords
      content: Olares One, Windows, primary drive, Olares OS, Windows installation, BIOS
---

# Install Windows on Olares One <Badge type="tip" text="30 min" />

Install Windows on the primary drive of Olares One when you want to use the device as a dedicated Windows machine.

This process replaces Olares OS. If you want to keep Olares OS and Windows on the same device, use a dual-boot setup instead.

:::danger This erases Olares OS
Installing Windows on the primary drive permanently deletes Olares OS, local accounts, installed apps, settings, and data stored on that drive. Back up anything you need before continuing.
:::

## Learning objectives

By the end of this guide, you will learn how to:

- Create a bootable Windows installation USB drive.
- Boot Olares One from the Windows USB drive.
- Install Windows on the primary drive by removing the existing Olares OS partitions.
- Install the Windows drivers provided for Olares One.

## Prerequisites

**Hardware**<br>
- A computer to create the installation USB drive.
- A USB flash drive, 8 GB or larger, for Windows installation media.
- A wired keyboard and mouse connected to Olares One.
- A monitor connected to Olares One.
- Connect an Ethernet cable to Olares One, as a wired internet connection is required during Windows setup.
- A Microsoft account for completing the Windows setup process.

## Step 1: Create a bootable Windows USB drive

1. Open the [Microsoft download page](https://www.microsoft.com/en-us/software-download/windows11), go to **Download Windows 11 Disk Image (ISO) for x64 devices**, and download the ISO. Choose x64 even if you create the USB drive on an Apple Silicon Mac. Windows will be installed on Olares One.

   :::warning
   Creating the installer erases the selected USB drive. Back up its files and check the drive name and capacity before writing.
   :::

2. Download the tool for your computer's operating system, then follow its instructions to create a bootable USB drive from the Windows 11 ISO.

   - **macOS**: Download [Wingman](https://github.com/duggan/Wingman/releases). Requires macOS 13 or later.
   - **Windows**: Download [Rufus](https://rufus.ie/).

   :::details Example using Wingman on macOS

   1. Open the downloaded `Wingman.dmg` and drag Wingman to **Applications**.
   2. Open Wingman. If you see **Wingman needs permission to format USB drives**, click **Open System Settings**. Go to **General > Login Items & Extensions** and turn on **Wingman** under **Allow in the Background**. If macOS asks for authentication, use Touch ID or your Mac login password. Return to Wingman after approving the change. The app detects the permission automatically.
   3. Insert your USB drive. Click **Choose ISO…** and select the Windows 11 x64 ISO you downloaded.
   4. Under **USB drive**, select the target USB drive. Wingman displays the device name and capacity, which may differ from the volume name shown in Finder. Check these details to identify the correct drive.
   5. Click **Create Windows USB**. In the confirmation dialog, check the device name and capacity again, then click **Erase and Create**. This erases all existing data on that USB drive.
   6. Keep the USB drive connected while Wingman formats it, copies the installation files, and splits `install.wim`. Wait until Wingman displays **Windows 11 USB created. You can unplug it now.** The installation USB drive has been ejected and is ready to unplug.

   :::

## Step 2: Boot from the Windows USB drive

1. Insert the Windows USB drive into a USB port on Olares One.
2. Power on Olares One, or restart it if it is already running.
3. When the Olares logo appears, press the **Delete** key repeatedly to enter BIOS setup.

   ![BIOS setup](/images/one/bios-setup-interface.png#bordered)

4. Go to the **Save & Exit** tab.
5. Under **Boot Override**, select the Windows USB drive, and then press **Enter**. Olares One restarts and boots into the Windows installer.

   :::tip
   If a Ventoy screen appears, select the Windows ISO file, and then select **Boot in normal mode**.
   :::

## Step 3: Install Windows on the primary drive

The Windows setup wizard guides you through the installation.

1. Follow the on-screen prompts until you reach the **Select location to install Windows 11** screen.

   ![Windows 11 installation location screen showing the internal SSD and USB installer](/images/one/windows-select-install-location.png)

2. Identify the internal SSD that contains Olares OS.

   :::tip Disk and partition labels
   Each **Disk** number identifies a disk, and the **Partition** entries beneath that number belong to the same disk. The **Primary** label in the **Type** column describes a partition type. It does not identify the SSD you should install Windows on.
   :::

   In the example above:

   | Entry | What it represents | What to do |
   | --- | --- | --- |
   | Disk 0 Partition 1, 512.0 MB, System | The boot partition on the internal SSD | Delete this partition when replacing Olares OS. |
   | Disk 0 Partition 2, 1.9 TB, Unknown | The existing Olares OS partition on the same internal SSD | Delete this partition when replacing Olares OS. |
   | Disk 1 Partition 1: WINDOWS, 57.7 GB, Primary | The Windows installation USB drive | Keep this partition. Do not delete or format it. |
   | Disk 2 Unallocated Space, 0 B | An entry with no usable capacity | Ignore this 0 B entry. |

   :::warning Confirm your disk before deleting anything
   This image is an example, not a fixed disk map. Disk numbers, volume names, capacities, and partition layouts can differ on your device.

   You are responsible for confirming the target disk and completing your backup. Delete only partitions on the SSD you intend to erase. If you cannot confidently identify that SSD, stop and contact support before continuing.
   :::

3. Select each partition on the confirmed target SSD and click **Delete Partition**. In this example, delete **Disk 0 Partition 1** and **Disk 0 Partition 2**. Leave Disk 1 and Disk 2 unchanged.
4. After deleting all partitions on the target SSD, select the unallocated space on that same disk. In this example, select **Disk 0 Unallocated Space** (**1.9 TB**).

   ![Select Disk 0 Unallocated Space after deleting the target SSD partitions](/images/one/windows-select-unallocated-space.png)

5. Click **Next** to continue.
6. On the **Ready to install** screen, review the installation summary, then click **Install** to start installing Windows.
7. Wait while Windows copies files and restarts several times.
8. When the final setup process starts, follow the on-screen prompts to configure Windows.
9. After the Windows desktop appears, unplug the Windows USB drive.

## Step 4: Install Olares One Windows drivers

Install the tested Olares One driver package after Windows starts. The package includes audio, network, chipset, and NVIDIA graphics drivers.

1. Follow [Install drivers on Windows](install-nvidia-driver.md) to download and install the all-in-one driver package.
2. Restart Windows if the driver installer prompts you to do so.
3. After Windows restarts, check that network, audio, display, and GPU devices are working normally.
