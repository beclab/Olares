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

## Learning objectives

By the end of this guide, you will learn how to:

- Install Windows on a secondary SSD alongside Olares OS.
- Configure BIOS boot settings for dual-boot.
- Set up GRUB to detect and boot both operating systems.
- Switch between Olares OS and Windows at startup.

## Prerequisites
**Hardware**<br>
- A secondary NVMe M.2 SSD physically installed in Olares One.
- A USB flash drive containing a bootable Windows installation media.
- A wired keyboard and mouse.
- A monitor connected to Olares One.

## Step 1: Boot into BIOS
1. Insert the Windows USB boot drive into a USB port on Olares One.
2. Power on Olares One or restart it if it is already running.
3. When the Olares logo appears, immediately press the **Delete** key repeatedly to enter **BIOS setup**.
   ![BIOS setup](/images/one/bios-setup.png#bordered)

## Step 2: Boot from USB

1. Navigate to the **Boot** tab using the arrow keys on your keyboard.
2. Set **Boot Option #1** to your Windows USB flash drive, and press **Enter**.
3. Press **F10**, then select **Yes** to save and exit BIOS.
4. The system restarts and boots from the USB drive into the Windows installation interface.

## Step 3: Install Windows
1. Follow the on-screen prompts to begin the Windows installation.
   :::danger Select the correct drive
   You must carefully identify the secondary SSD.

   Selecting the wrong drive will permanently erase your Olares data.
   :::

2. When the installation finishes and the system restarts, unplug the Windows USB drive.

Once installation is complete, the system will restart into Windows automatically.

## Step 4: Boot back into Olares OS

1. Restart Olares One.
2. When the Olares logo appears, press the **Delete** key repeatedly to enter **BIOS setup**.
3. Go to the **Boot** tab and set **Boot Option #1** to the SSD that contains Olares OS.
4. Press **F10**, then select **Yes** to save and exit BIOS.

Olares One will boot into Olares OS.

## Step 5: Detect Windows and update GRUB

1. Log in using the default credentials:
    * **Username**: `olares`
    * **Password**: `olares`
   
   ![Log in to Olares One](/images/one/one-terminal.png#bordered)
2. Run the following command:

   ```bash
   sudo os-prober
   ```

   If Windows has been installed successfully, you should see an entry similar to:

   ```bash
   /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi:Windows Boot Manager:Windows:efi
   ```

3. Enable GRUB to probe other operating systems and regenerate the boot menu:

   a. Create a symbolic link for GRUB configuration:
      ```bash
      sudo ln -s /boot/efi/grub /boot/grub
      ```

   b. Enable OS prober to detect Windows:
      ```bash
      sudo sed -i 's|GRUB_DISABLE_OS_PROBER=true|GRUB_DISABLE_OS_PROBER=false|' /etc/default/grub
      ```

   c. Regenerate the GRUB boot menu:
      ```bash
      sudo update-grub
      ```

   Example output:

   ```bash
   Sourcing file '/etc/default/grub'
   Generating grub configuration file ...
   Warning: os-prober will be executed to detect other bootable partitions.
   Its output will be used to detect bootable binaries on them and create new boot entries.
   Found Windows Boot Manager on /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi
   Adding boot menu entry for UEFI Firmware Settings ...
   done
   ```

## Step 6: Switch between operating systems

1. Shut down Olares One, wait a few seconds, and then power it on again. You should now see a GRUB menu with both Olares and Windows entries.

   ![Switch systems at startup](/images/one/one-dual-boot.png#bordered)
   :::tip
   The highlighted entry (Olares GNU/Linux) will be executed automatically in 10 seconds.
   :::
2. At the GRUB menu:

- **Boot Olares OS**: Select `Olares GNU/Linux`.
- **Boot Windows**: Select `Windows Boot Manager`.

## Resources

- [Install NVIDIA drivers on Windows](install-nvidia-driver.md)
- [Run a Windows VM on Olares One](windows.md)