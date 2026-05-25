---
outline: [2,3]
description: Reinstall Olares One with Ubuntu Server by replacing the existing Olares OS on the primary SSD.
head:
  - - meta
    - name: keywords
      content: Olares One, Ubuntu Server, NVMe SSD, OS installation, clean install
---

# Install Ubuntu Server on Olares One <Badge type="tip" text="25 min" />

Replace the pre-installed Olares OS with a clean installation of Ubuntu Server on the primary NVMe SSD of Olares One. 

:::danger Permanent data loss
This installation process completely erases the primary drive, including Olares OS and all locally stored digital assets. Back up all critical files before proceeding.
:::

## Learning objectives

In this guide, learn how to:
- Create a bootable Ubuntu Server installation USB drive.
- Boot Olares One from the installation media.
- Overwrite the primary drive and install Ubuntu Server.

## Prerequisites

**Hardware**
- The primary NVMe M.2 SSD installed inside Olares One.
- A USB flash drive (8 GB or larger) for the installation media.
- A wired keyboard and mouse.
- A monitor connected to Olares One.

## Step 1: Create a bootable Ubuntu Server USB drive

1. Download the Ubuntu Server ISO, version 26.04 LTS or later, from the [official Ubuntu website](https://ubuntu.com/download/server).
2. Download and install [balenaEtcher](https://etcher.balena.io/).
3. Insert the USB flash drive into your computer.
4. Open balenaEtcher and follow these steps:

   a. Click **Flash from file** and select the downloaded ISO.

   b. Click **Select target** and select your USB drive.

   c. Click **Flash!** to write the installer to the USB drive. 

   ![balenaEtcher flashing screen](/images/one/balenaEtcher.png#bordered)

5. Wait for the flashing and validation processes to finish, and then safely eject the USB drive.

## Step 2: Boot from the Ubuntu Server USB drive

1. Insert the Ubuntu Server USB drive to the USB port on Olares One.
2. Power on Olares One, or restart it if it is currently running.
3. When the Olares logo appears, press the **Delete** key repeatedly to enter the BIOS setup.  

   ![BIOS setup menu](/images/one/bios-setup-interface.png#bordered)

4. Go to the **Save & Exit** tab.
5. Under the **Boot Override** section, select your USB drive from the list, and then press **Enter**.

    ![Select Ubuntu USB on Boot menu in BIOS](/images/one/select-ubuntu-usb-in-bios3.png#bordered)

    The system restarts and boots from the USB drive into the Ubuntu Server installation interface.

## Step 3: Install and configure Ubuntu Server

The text-based installation wizard guides you through replacing Olares OS on the primary drive.

1. In GNU GRUB, select **Try or Install Ubuntu Server**. Wait for the initial loading sequence to finish and the language selection screen to appear.

   ![Ubuntu install type](/images/one/ubuntu-install-type.png#bordered)

2. Select your language, and then press **Enter**.

   ![Ubuntu language selection](/images/one/ubuntu-language.png#bordered)

3. Keep the default English US keyboard layout, and then press **Enter**.
4. On the **Choose the type of installation** screen, select **Ubundu Server**, and then press **Enter**.
5. On the **Network configuration** screen, skip network configuration for now by selecting **Continue without network** at the bottom, and then press **Enter**.

    :::tip
    Connecting to the network triggers automatic background downloads for patches and dependencies. This can significantly delay the installation and might cause the installer to hang due to network fluctuations. Skipping this ensures a rapid, completely local installation from the pure ISO image.
    :::

   ![Ubuntu network config](/images/one/ubuntu-network.png#bordered)

6. On the **Proxy configuration** screen, leave it blank unless your environment requires one, and then press **Enter**.
7. On the **Ubuntu archive mirror configuration** screen, keep the default Ubuntu archive mirror URL, ignore the "no network" warning, and then press **Enter**.
8. On the **Guided storage configuration** screen:

   a. Ensure **Use an entire disk** is selected, and the primary disk containing Olares OS is displayed in the dropdown list below.

   b. Navigate down to **Set up this disk as an LVM group** and clear the selection using the **Space** key.

      :::tip
      Disabling LVM forces the installer to automatically create stable, straightforward standard ext4 partitions. This  eliminates the risk of future GRUB bootloader conflicts or errors in a multi-OS environment.
      :::

   c. Go to the bottom of the page, select **Done**, and then press **Enter**.

   ![Ubuntu guided storage configuration](/images/one/ubuntu-guided-storage1.png)

9. Verify the details on the **Storage configuration** summary screen. Under **USED DEVICES**, ensure the primary drive is in the "to be formatted" status, and then press **Enter**.

    ![Ubuntu storage configuration summary](/images/one/ubuntu-storage-summary1.png)

10. In the **Confirm destructive action** window, select **Continue**, and then press **Enter**.
11. On the **Profile configuration** screen, set up your account credentials, select **Done** at the bottom, and then press **Enter**.
12. On the **Upgrade to Ubuntu Pro** screen, select **Skip Ubuntu Pro setup for now**, and then press **Enter**.
13. On the **SSH Configuration** screen, select **Install OpenSSH server** to allow remote terminal management after connecting to the network later, and then press **Enter**.
14. The system begins deployment. Wait for the top banner to display **Installation complete**.

      ![Ubuntu installation complete](/images/one/ubuntu-install-complete.png#bordered)

15. Select **Reboot Now** at the bottom, and then press **Enter**.
16. Remove the installation USB drive and press **Enter** when prompted. The system restarts and reboots automatically into your fresh Ubuntu Server environment.

      ![Ubuntu launch](/images/one/ubuntu-launch.png#bordered)

17. Log in using the account credentials you set previously.

## Resources

- [Install Ubuntu Desktop on Olares One](install-ubuntu-server.md)
- [Ubuntu Server documentation](https://ubuntu.com/server/docs)
