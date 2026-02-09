---
outline: [2,3]
description: Learn how to install Windows on a secondary SSD to create a dual-boot system on Olares One.
head:
  - - meta
    - name: keywords
      content: Dual-boot, Windows, NVMe SSD, BIOS, Windows installation
---

# Dual-boot Windows on a secondary SSD <Badge type="tip" text="30 min" />
For competitive gaming or Windows-exclusive software, you can add a secondary NVMe SSD to create a dual-boot system.

This dual-drive configuration physically isolates the systems. This ensures Olares OS remains stable and secure while providing full native performance for your Windows applications.

## Prerequisites
**Hardware**<br>
- A secondary NVMe M.2 SSD physically installed in Olares One.
- A USB flash drive containing a bootable Windows installation media.
- A wired keyboard and mouse.
- A monitor connected to Olares One.

## Step 1: Boot into BIOS
1. Insert the Windows USB boot drive into a USB port on Olares One.
2. Power on Olares One or restart if it is already running.
3. Immediately press and hold the **Delete** key repeatedly until the BIOS setup screen appears.

## Step 2: Boot from USB

1. Navigate to the **Boot** tab using the arrow keys on your keyboard.
2. Locate **Boot Option #1** or **Boot Override** and select your USB flash drive.
3. Press **Enter** to boot from the USB drive immediately.
4. The system will restart and load the Windows installation interface.

## Step 3: Install Windows
1. Follow the on-screen prompts to begin the Windows installation.
   :::danger Select the correct drive
   You must carefully identify the secondary SSD.

   Selecting the wrong drive will permanently erase your Olares data.
   :::

2. When the installation finishes and the system restarts, unplug the Windows USB drive.

Once installation is complete, the system will restart into Windows automatically.

## Switch between operating systems

Because the operating systems are on separate physical drives, you switch between them using the BIOS boot priority.

### Switch to Olares OS
1. Restart Olares One.
2. Press **Delete** repeatedly to enter BIOS.
3. Go to the **Boot** tab.
4. Set **Boot Option #1** to the SSD containing Olares OS.
5. Press **F4** to confirm.

### Switch to Windows
1. Restart Olares One.
2. Press **Delete** repeatedly to enter BIOS.
3. Set **Boot Option #1** to the secondary SSD containing Windows.
4. Press **F4** to confirm.

## Resources
- [Install NVIDIA drivers on Windows](install-nvidia-driver.md)
- [Run a Windows VM on Olares One](windows.md)