---
outline: [2,3]
description: Reinstall Olares One with Ubuntu Desktop by replacing the existing Olares OS on the primary SSD.
head:
  - - meta
    - name: keywords
      content: Olares One, Ubuntu Desktop, NVMe SSD, OS installation, clean install, graphical setup
---

# Install Ubuntu Desktop on Olares One <Badge type="tip" text="30 min" />

Replace the pre-installed Olares OS with a clean installation of Ubuntu Desktop on the primary NVMe SSD of Olares One.

:::danger Permanent data loss
This installation process completely erases the primary drive, including Olares OS and all locally stored digital assets. Back up all critical files before proceeding.
:::

## Learning objectives

In this guide, you will learn how to:
- Create a bootable Ubuntu Desktop installation USB drive.
- Boot Olares One from the installation media.
- Overwrite the primary drive and configure the graphical Ubuntu Desktop installer.

## Prerequisites

**Hardware**
- The primary NVMe M.2 SSD installed inside Olares One.
- A USB flash drive (8 GB or larger) for the installation media.
- A wired keyboard and mouse.
- A monitor connected to Olares One.

## Step 1: Create a bootable Ubuntu Desktop USB drive

1. Download the Ubuntu Desktop ISO from the [official Ubuntu website](https://ubuntu.com/download/desktop).
2. Download and install [balenaEtcher](https://etcher.balena.io/).
3. Insert the USB flash drive into your computer.
4. Open balenaEtcher and follow these steps:

   a. Click **Flash from file** and select the downloaded ISO.

   b. Click **Select target** and select your USB drive.

   c. Click **Flash!** to write the installer to the USB drive. 

   ![balenaEtcher flashing screen](/images/one/balenaEtcher.png#bordered)

5. Wait for the flashing and validation processes to finish, and then safely eject the USB drive.

## Step 2: Boot from the Ubuntu Desktop USB drive

1. Insert the Ubuntu Desktop USB drive into the USB port on Olares One.
2. Power on Olares One, or restart it if it is currently running.
3. When the Olares logo appears, press the **Delete** key repeatedly to enter the BIOS setup.  

   ![BIOS setup menu](/images/one/bios-setup-interface.png#bordered)

4. Go to the **Save & Exit** tab.
5. Under the **Boot Override** section, select your USB drive from the list, and then press **Enter**.

    ![Select Ubuntu USB on Boot menu in BIOS](/images/one/select-ubuntu-usb-in-bios3.png#bordered)

    The system boots from the USB drive into the Ubuntu Desktop graphical boot menu.

## Step 3: Install and configure Ubuntu Desktop

The graphical installation wizard guides you through replacing Olares OS on the primary drive.

1. In GNU GRUB, select **Try or Install Ubuntu**. Wait for the initial loading sequence to finish and the language selection screen to appear.

   ![Ubuntu install type](/images/one/ubuntu-install-desktop.png#bordered)

2. Follow the on-screen prompts to finish the standard setup.
3. When you see Ubuntu is installed and ready to use, click **Restart now**.
4. Remove the installation USB drive and press **Enter** when prompted. The system restarts and reboots into your fresh Ubuntu Desktop environment.
5. Log in using the account credentials you set during configurations.

## Resources

- [Install Ubuntu Server on Olares One](install-ubuntu-server.md)
- [Ubuntu Server documentation](https://ubuntu.com/server/docs)
