---
outline: [2, 3]
description: Learn how to create a bootable USB installer for Olares OS using balenaEtcher.
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, install Olares, bootable USB, ISO, balenaEtcher
---

# Create a bootable USB drive <Badge type="tip" text="15 min"/>

To reinstall or recover Olares OS on your Olares One, you must create a bootable USB installation drive. This drive allows you to reset the device to its factory state.

:::warning Data loss warning
This process will erase all data on your USB drive. If you have important files on the drive, back them up before proceeding.
:::
:::warning Image compatibility
Olares One requires a specific system image designed for its hardware. If you install the generic self-hosted ISO, Olares One will fail to boot.
:::
## Prerequisites
- USB flash drive: A drive with 8 GB or higher capacity.
  :::info
  Ensure that your USB drive is empty or backed up. The flashing process re-formats the drive, which means existing data on it will be permanently deleted.
  :::
- Computer: A Windows, macOS, or Linux computer to perform the setup.
- Internet connection: Stable network for downloading the image file and related software.

## Step 1. Download Olares image

Get the official installation file for the Olares operating system.

1. Click [Download Olares ISO](https://cdn.olares.com/one/olares.iso).
2. The latest image file `olares.iso` is automatically downloaded to your computer.

## Step 2. Install flashing tool

It is recommended to use balenaEtcher for this task because it is free, easy to use, and works on all major operating systems.

1. Go to the [balenaEtcher website](https://etcher.balena.io/).
2. Download and install the version appropriate for your computer (Windows, macOS, or Linux).

## Step 3. Flash the drive

Use balenaEtcher to turn your USB stick into a bootable installer.

1. Insert your USB flash drive into the computer.
2. Open balenaEtcher.
3. Follow the steps on the screen:

    a. Click **Flash from file** and select the Olares ISO file you downloaded.

    b. Click **Select target** and select your USB drive.

    c. Click **Flash!** to start writing the installer to the USB drive.

    ![Bootable USB](/images/one/balenaEtcher.png#bordered)

4. When the flashing process is completed, the USB drive is ready to use, and you can safely eject it.

## Next steps

You can now insert the bootable USB drive into Olares One or other target hardware to begin the installation.

## Resources

- [Install Olares via ISO](../manual/get-started/install-linux-iso.md)
- [Install Olares](../manual/get-started/install-olares.md)
- [Installation FAQs](../manual/help/installation.md)