---
outline: [2, 3]
description: Set up an NVIDIA eGPU on Olares One running Windows 11 and recover the built-in GPU if its driver reports an error.
---

# Set up an eGPU on Olares One with Windows

Use this guide for the first eGPU setup on Windows 11 or when the built-in GPU reports an error after you connect an eGPU.

:::warning Complete the cleanup before connecting
For the initial setup, leave the eGPU disconnected until CleanupTool finishes and Windows restarts. If the eGPU is already connected, shut down Olares One, disconnect it, and then start Windows again.
:::

## Before you start

You need:

- Windows 11 updated to the latest version available through Windows Update. The tested configurations used Windows 11 24H2.
- A powered Thunderbolt external graphics device and a certified Thunderbolt cable. The GPU may be preinstalled, or you can install a desktop GPU in an eGPU dock or enclosure.
- Administrator access to Windows.

If Windows is not installed, first follow [Install Windows on the primary drive](./install-windows-primary-drive.md). You do not need to reinstall Windows or use a specific Windows image just to add an eGPU.

## Set up the eGPU for the first time

1. Open **Settings** > **Windows Update**. Install all available updates, then check again until Windows reports that it is up to date.
2. Download and run [`CleanupTool_1.0.21.0`](https://cdn.olares.com/common/CleanupTool_1.0.21.0.exe) to remove all existing NVIDIA apps, graphics drivers, and related utilities.

3. When CleanupTool prompts you to restart, select **Yes** and wait for Windows to start again.
4. Shut down Olares One.
5. Prepare and power on the eGPU:

   - If the GPU is already installed, connect the device's power adapter.
   - If you use an eGPU dock or enclosure, install the desktop GPU and connect all required GPU power cables.

6. Connect the eGPU directly to a Thunderbolt 5 (USB-C) port on Olares One with a certified Thunderbolt cable.
7. Start Olares One.
8. Install NVIDIA App.
9. In NVIDIA App, download and install the latest graphics driver available for your GPU.
10. Restart Windows.

Before connecting other Thunderbolt devices, [check the setup](#check-the-setup).

## Recover the built-in GPU

Use these steps if the built-in GPU has a warning icon in Device Manager or disappears from NVIDIA App after you connect the eGPU.

Keep the eGPU connected during this procedure.

1. Open **NVIDIA App** > **Drivers** > **Reinstall**.
2. Select **Custom installation**.
3. Select **Perform a clean installation** and finish the installation.
4. Restart Windows.

## Check the setup

1. Open **Device Manager** > **Display adapters**.
2. Check that the built-in GPU and every connected eGPU appear without warning icons.
3. Open NVIDIA App and check that it shows all expected GPUs.

This guide covers the recommended setup with one eGPU. A tested three-eGPU topology is summarized in [Check known compatibility](./egpu.md#check-known-compatibility). This guide does not cover hot-plugging.

## Known issues

**Reduced power for the built-in RTX 5090M**

In the tested multi-eGPU configuration, the built-in RTX 5090M was limited to about 95 W instead of 175 W. This limitation affects the built-in GPU, not the connected eGPUs.

## If a GPU is missing or reports an error

Follow [Troubleshoot eGPU issues](./ts-egpu.md). Record the Device Manager error code before reinstalling the driver because the code helps identify the failure.

## Related resources

- [Connect an eGPU to Olares One](./egpu.md)
- [Troubleshoot eGPU issues](./ts-egpu.md)
