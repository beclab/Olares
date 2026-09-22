---
outline: [2, 3]
description: Set up an NVIDIA eGPU on Olares One running Windows 11 and recover the built-in GPU if its driver reports an error.
---

# Set up an eGPU on Olares One with Windows

Use this guide for the first eGPU setup on Windows 11 or when the built-in GPU reports an error after you connect an eGPU.

:::warning Disconnect the eGPU first
Start Windows without the eGPU connected. If it is already connected, shut down Olares One, disconnect it, and then start Windows again.
:::

## Before you start

You need:

- Windows 11 24H2 with all available Windows updates installed.
- A powered Thunderbolt external graphics device and a certified Thunderbolt cable. The GPU may be preinstalled, or you can install a desktop GPU in an eGPU dock or enclosure.
- Administrator access to Windows.

If Windows is not installed, first follow [Install Windows on the primary drive](./install-windows-primary-drive.md). You do not need to reinstall Windows just to add an eGPU.

## Set up the eGPU for the first time

1. Uninstall all existing NVIDIA apps and graphics drivers from Windows.
2. Restart Windows if the uninstaller asks you to do so.
3. Prepare and power on the eGPU:

   - If the GPU is already installed, connect the device's power adapter.
   - If you use an eGPU dock or enclosure, install the desktop GPU and connect all required GPU power cables.

4. Connect the eGPU directly to a Thunderbolt 5 (USB-C) port on Olares One with a certified Thunderbolt cable.
5. Install NVIDIA App.
6. In NVIDIA App, download and install the graphics driver.
7. Restart Windows.

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
2. Check that the built-in GPU and eGPU both appear without warning icons.
3. Open NVIDIA App and check that it shows both GPUs.

This guide covers one eGPU. It does not cover hot-plugging or multi-eGPU setup.

## If a GPU is missing or reports an error

Follow [Troubleshoot eGPU issues](./ts-egpu.md). Record the Device Manager error code before reinstalling the driver because the code helps identify the failure.

## Related resources

- [Connect an eGPU to Olares One](./egpu.md)
- [Troubleshoot eGPU issues](./ts-egpu.md)
