---
outline: [2, 3]
description: Set up and verify an NVIDIA eGPU on Olares One running Windows 11, including recovery steps for built-in GPU driver errors.
---

# Set up an eGPU on Olares One with Windows

Use this guide to connect and verify an NVIDIA eGPU on Olares One running Windows 11. It also explains how to recover the built-in GPU if it reports a driver error after the eGPU is connected.

:::warning Do not connect the eGPU yet
Remove the existing NVIDIA software before connecting the eGPU. Power the enclosure with its dedicated power supply and use a certified Thunderbolt 5 cable.
:::

## Before you start

You can use an existing Windows 11 24H2 installation. You do not need to reinstall Windows solely to set up an eGPU. If you are installing Windows, the image `AGBOX4_WIN1124H2EN20260527` was used in testing and includes the system and drivers.

This guide has been tested with:

- Olares One
- Windows 11 24H2
- AOOSTAR EG02
- NVIDIA GeForce RTX 4060 Ti
- NVIDIA driver 610.74

## Prepare Windows

:::warning Back up your data
Installing Windows can remove the existing operating system, apps, settings, and files from the Windows partition. Back up any data you want to keep before continuing.
:::

1. If needed, install Windows 11 24H2. The image used in testing is `AGBOX4_WIN1124H2EN20260527`.
2. Open **Settings** > **Windows Update** and install all available Windows updates. Wait for the update process to finish before continuing.
3. Run `CleanupTool_1.0.21.0` and remove all NVIDIA apps and drivers.

## Connect the eGPU and install the driver

1. Install the GPU in the enclosure and turn on the enclosure.
2. Connect it to the Thunderbolt 5 (USB-C) port on Olares One with a certified Thunderbolt 5 cable.
3. Run `NVIDIA_APP_11.0.5.420_1146713` to install NVIDIA App.
4. In NVIDIA App, download and install the graphics driver. Version 610.74 was used in testing.
5. After the driver installation is complete, continue to [Check the connection](#check-the-connection).

## Recover the built-in GPU

Use these steps if the built-in GPU shows a warning or disappears from NVIDIA App after you connect the eGPU.

1. Open **NVIDIA App** > **Drivers** > **Reinstall**.
2. Select **Custom installation**.
3. Select **Perform a clean installation**, then finish the installation.

   :::warning Do not skip the restart
   In the tested configuration, the eGPU remained at Gen1 until Windows was restarted. Restart Windows to let the link renegotiate at Gen4.
   :::

4. Restart Windows.

## Check the connection

In **Device Manager** > **Display adapters**, confirm that both the built-in GPU and eGPU appear without warning icons.

## Tested performance

In the single-eGPU test, the RTX 4060 Ti reached 165 W under FurMark load at Gen4, with no observed PCIe errors.

Results may vary with the GPU, enclosure, power supply, driver, and workload. Running FurMark is not required to complete the setup.

## Known issues

| Symptom | Status |
|---|---|
| Built-in RTX 5090M reaches about 95 W instead of 175 W | This is a known issue confirmed with NVIDIA. The tested eGPU link remained stable at Gen4, and eGPU use was not affected. |
| Built-in GPU reports an error after connecting the eGPU | Perform the clean installation above and restart. |

## Resources

- [Olares One eGPU support overview](./egpu.md)
- [Troubleshoot eGPU issues](./ts-egpu.md)
