---

outline: [2, 3]
description: Install Olares on NVIDIA DGX Spark using the official ISO image, including bootable USB creation, installation steps, and activation process.
---

# Install Olares on DGX Spark via ISO <Badge type="warning" text="RC" />

This guide explains how to install Olares on NVIDIA DGX Spark using the official ISO image.

:::warning RC version
DGX Spark support is currently in Release Candidate (RC). We are actively testing and will release the stable version soon.
:::

<!--@include: ./reusables.md{44,51}-->

## Before you begin

- **DGX Spark**: Ensure your device is connected to a monitor and keyboard.
- **USB flash drive**: A drive with 8 GB or higher capacity.
- **Computer**: A Windows, macOS, or Linux computer to create the bootable USB drive.
- **Network**: An Ethernet cable connecting DGX Spark to your router (recommended for stable connection).

## Create a bootable USB drive

1. Download [the official Olares ISO image for Spark](https://cdn.olares.com/spark/olares.iso).
2. Download and install [Balena Etcher](https://etcher.balena.io/).
3. Insert the USB flash drive into your computer.
4. Launch Etcher and follow these steps:

   a. **Image**: Select the Olares ISO.

   b. **Target disk**: Select your USB drive.

   c. Click **Flash** to write the installer to the USB drive.

   ![Bootable USB](/images/manual/get-started/iso-flash.png#bordered)

## Boot from USB drive

1. Insert the bootable USB drive into DGX Spark.
2. Restart DGX Spark, then immediately press the **Delete** key repeatedly to enter the BIOS setup.
   ![BIOS setup](/images/one/bios-setup.png#bordered)

3. Navigate to the **Boot** tab, set **Boot Option #1** to the USB drive, and then press **Enter**.
   ![Set boot option](/images/one/bios-set-boot-option.png#bordered)

4. Press **F10** to save and restart. The system will automatically boot into the Olares installer interface.

## Install Olares

1. From the Olares installer interface, select **Install Olares to Hard Disk** and press **Enter**.
   ![Olares installer](/images/one/olares-installer.png#bordered)

2. The installer will display a list of available disks (e.g., `sda 200G HARDDISK`).
   Type `/dev/` followed by the disk name (e.g., `/dev/sda`) to select the installation target. The installation typically takes 4–5 minutes.
   :::tip Note
   During installation, you may see prompts related to NVIDIA GPU drivers.
   Simply press **Enter** to confirm.
   :::
3. Once you see the message below, the installation has completed successfully:

   ```shell
   Installation completed successfully!
   ```

4. Remove the USB drive, and manually shut down Spark and then turn it back on.
   :::warning Important
   If you skip this step, the activation process will fail.
   :::

5. To prevent startup delays, turn on DGX Spark and immediately press the **Delete** key repeatedly to enter the BIOS setup. Set the internal hard drive as the **Boot Option #1**.

## Connect to DGX Spark

<tabs>
<template #Set-up-via-wired-LAN>

1. Ensure DGX Spark is connected to your router via Ethernet.
2. In the LarePass app, on your account activation page, tap **Discover nearby Olares**.
3. Select the target Olares instance from the list.

</template>
<template #Set-up-via-wireless-network>

1. In the LarePass app, on your account activation page, tap **Discover nearby Olares**.
2. Tap **Bluetooth network setup** at the bottom.
3. Select your device from the Bluetooth list and tap **Network setup**.
4. Follow the prompts to connect DGX Spark to the Wi-Fi network your phone is currently using.
5. Once connected, return to the main screen and tap **Discover nearby Olares** again to find your device.

</template>
</tabs>

## Activate Olares

1. In the LarePass app, on the device you just found, tap **Install now**.
2. When the installation completes, click **Activate now**.
3. In the **Select a reverse proxy** dialog, select a node that is closer to your geographical location. The installer will then configure HTTPs certificate and DNS for Olares.
   :::tip Note
   You can change this setting later on the [Change reverse proxy](../olares/settings/change-frp.md) page in Olares.
   :::
4. Follow the on-screen instructions to set the login password for Olares, then tap **Complete**.

   ![ISO Activate-2](/images/manual/larepass/iso-activate-2.png#bordered)

Once activation is complete, LarePass will display the desktop address of your Olares device, such as `https://desktop.marvin123.olares.com`.

<!--@include: ./log-in-to-olares.md-->

## Configure GPU memory for AI apps

The system uses **Memory slicing** mode for GPU resource management. When you install an AI application, Olares automatically allocates the minimum required VRAM to ensure the app can start and run properly.

You can manually adjust the VRAM allocation for each AI application based on your specific needs:

1. Open Settings from Olares, and then navigate to **GPU**.
2. In the **Allocate VRAM** section, find the target app.

    ![Memory slicing](/images/manual/get-started/install-spark-memory-slicing.png#bordered){width=70%}

3. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
4. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

<!--@include: ./reusables.md{38,42}-->