---
outline: [2, 3]
description: Install Olares on a physical machine using the official ISO image, including system requirements, installation steps, and activation process.
---

# Install Olares via ISO

This guide explains how to install Olares on a physical machine using the official ISO image.

<!--@include: ./reusables.md{44,51}-->

## Prerequisites

### Required
- **Host requirements**:
  - **CPU**: Minimum 4 cores with **x86-64 architecture** (Intel or AMD). ARM-based processors are not currently supported for this method.
  - **Memory**: At least 8 GB of available RAM.
  - **Storage**: Minimum 150 GB SSD.
  - **Network**: The device must be connected to a wired LAN.
  ::: warning SSD required
  The installation will likely fail if an HDD (mechanical hard drive) is used instead of an SSD.
  :::

- **USB flash drive**: **8 GB** capacity or larger.

### Optional

A GPU is not required to install Olares, but is necessary for AI applications.

- **GPU (NVIDIA only)**:
  - **Architecture**: Turing or newer (e.g., GTX 16 series, RTX 20 series).
  - **Verification**: Run `lspci | grep -i nvidia` and check the [compatible GPU table](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus).

## Create a bootable USB drive

1. Click [here](https://cdn.olares.com/olares-latest-amd64.iso) to download the official Olares ISO image.
2. Download and install [**Balena Etcher**](https://etcher.balena.io/).
3. Insert the USB flash drive into your computer.
4. Launch Etcher and follow these steps:
   ![Bootable USB](/images/manual/get-started/iso-flash.png#bordered)
   - **Image**: Select the Olares ISO.
   - **Target disk**: Select your USB drive.
   - Click **Flash** to write the installer to the USB drive.

## Boot from USB drive

1. Insert the bootable USB drive into the target machine.
2. Enter the **BIOS setup** and set **USB Boot** as the first boot option.
3. Save and restart. The system will automatically boot into the Olares installer interface.

## Install Olares

1. From the Olares installer interface, select **Install Olares to Hard Disk** and press **Enter**.
2. The installer will display a list of available disks (e.g., `sda 200G HARDDISK`).  
   Type `/dev/` followed by the disk name (e.g., `/dev/sda`) to select the installation target. The installation typically takes **4–5 minutes**.
   :::tip Note
   During installation, you may see prompts related to NVIDIA GPU drivers.  
   Simply press **Enter** to confirm.
   :::
3. Once you see the message below, the installation has completed successfully:

   ```shell
   Installation completed successfully!
   ```

You can now remove the USB drive and press **Ctrl + Alt + Del** to restart.

## Verify installation

After rebooting, the system will enter Ubuntu.

1. Log in using the default credentials:

    * **Username**: `olares`
    * **Password**: `olares`

2. Run the following command to verify the installation:

   ```bash
   sudo olares-check
   ```

   If the output shows the following, Olares has been installed successfully:

   ```bash
   check list ---------
   check Olaresd: success
   check Containerd: success
   ```

<!--@include: ./install-and-activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{38,42}-->