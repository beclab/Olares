---
outline: [2, 3]
description: Reinstall Olares OS on Olares One using a bootable USB to restore the device to factory state.
head:
  - - meta
    - name: keywords
      content: Olares One, reinstall, factory reset, bootable USB, installation USB
---

# Reset to factory settings using installation USB <Badge type="tip" text="15 min"/>

Resetting to factory settings returns your Olares One to the initial setup state. You can reinstall Olares OS using the bootable USB drive included with Olares One.

:::warning Data loss
This will permanently delete all accounts, settings, and data on the device. This action cannot be undone.
:::

## Prerequisites
**Hardware**<br>
- The bootable USB drive that came with Olares One.
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