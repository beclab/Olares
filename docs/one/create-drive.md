---
outline: [2, 3]
description: Reinstall Olares OS on Olares One using a bootable USB drive to restore the device to a clean initial state.
head:
  - - meta
    - name: keywords
      content: Olares One, reinstall, Olares OS, bootable USB, installation USB
---

# Reinstall Olares OS using bootable USB <Badge type="tip" text="15 min"/>

Reinstalling Olares OS returns your Olares One to a clean initial state. You can do this using the bootable USB drive included with Olares One.

:::warning Data loss
This will permanently delete all accounts, settings, and data on the device. This action cannot be undone.
:::

## Prerequisites
**Hardware**<br>
- The bootable USB drive that came with Olares One.
   :::tip Don't have the USB drive?
   Download the [Olares One ISO](https://cdn.olares.com/one/v1.12.4-amd64.iso), which is device-specific and different from the standard Olares ISO, and flash it to a USB drive (8 GB or larger) using a tool such as [Balena Etcher](https://etcher.balena.io/).
   :::
- A monitor and keyboard connected to Olares One.

## Step 1: Boot from the USB drive

1. Insert the bootable USB drive into Olares One.
2. Power on Olares One or restart it if it is already running.
3. When the Olares logo appears, immediately press the **Delete** key repeatedly to enter **BIOS setup**.
   ![BIOS setup](/images/one/bios-setup.png#bordered)

4. Navigate to the **Boot** tab, set **Boot Option #1** to the USB drive, and then press **Enter**.
   ![Set boot option](/images/one/bios-set-boot-option.png#bordered)

5. Press **F10**, then select **Yes** to save and exit.
   ![Save and exit](/images/one/bios-save-usb-boot.png#bordered)


Olares One will restart and boot into the Olares installer interface.

## Step 2: Install Olares to disk

1. From the installer interface, select **Install Olares to Hard Disk** and press **Enter**.
   ![Olares installer](/images/one/olares-installer.png#bordered)

2. When prompted for the installation target, the installer shows a list of available disks. Type `/dev/` followed by the disk name (e.g. `nvme0n1`) from that list and press **Enter**.
   ![Select disk](/images/one/olares-installer-select-disk.png#bordered)

   For example, to install to `nvme0n1`, enter:
   ```bash
   /dev/nvme0n1
   ```

3. When you see prompts about NVIDIA GPU drivers, press **Enter** to accept the default.
   ![Install NVIDIA drivers](/images/one/olares-installer-install-nvidia-drivers.png#bordered)

4. When you see the message below, the reinstallation is complete:
   ```bash
   Installation completed successfully!
   ```

5. Remove the USB drive, then press **Ctrl + Alt + Delete** to restart.

## Step 3: Verify the installation

After the reboot, the system starts in a clean factory state and shows a text-based Ubuntu login prompt.

1. Log in with the default credentials:
   - **Username**: `olares`
   - **Password**: `olares`
   ![Log in](/images/one/olares-login.png#bordered)

2. (Optional) Run the following command to verify the installation:
   ```bash
   sudo olares-check
   ```
   Example output:
   ![Olares check](/images/one/olares-check.png#bordered)


## Step 4: Complete activation via LarePass

You can then activate Olares One again via LarePass. For detailed instructions, see [First boot](first-boot.md).