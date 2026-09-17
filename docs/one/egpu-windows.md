---
outline: [2, 3]
description: Set up and verify an NVIDIA eGPU on Olares One running Windows 11, including recovery steps for built-in GPU driver errors.
---

# Set up an eGPU on Olares One with Windows

Use this guide to connect and verify an NVIDIA eGPU on Olares One running Windows 11. It also explains how to recover the built-in GPU if it reports a driver error after the eGPU is connected.

:::warning Do not connect the eGPU yet
For the fresh-installation path, connect the eGPU only after removing the existing NVIDIA software. Power the enclosure with its dedicated power supply and use a certified Thunderbolt 5 cable.
:::

## Before you start

This guide covers two scenarios:

- If you are installing Windows on Olares One, follow [Set up from a fresh Windows installation](#set-up-from-a-fresh-windows-installation).
- Setup on an existing Windows installation has not been verified. Do not reinstall Windows solely to follow this guide. If the built-in GPU reports an error after an eGPU has been connected, follow [Recover the built-in GPU](#recover-the-built-in-gpu).

This guide has been tested with:

- Olares One
- Windows 11 24H2
- AOOSTAR EG02
- NVIDIA GeForce RTX 4060 Ti
- NVIDIA driver 610.74

Driver 610.74 was used in testing. Confirm the driver recommended for Olares One before installing a different version.

<!-- TODO(tech-review):
Confirm the supported Windows setup paths, including:
- whether users must install the specified Windows image or may use an existing Windows 11 24H2 installation
- the complete setup steps for users who already have Windows and the NVIDIA driver installed
- whether the eGPU should be connected while Windows is running during initial setup
- whether routine Windows hot-plugging is supported
- whether CleanupTool requires a restart, and whether the restart occurs before or after connecting the eGPU
- whether Windows must be restarted after the fresh driver installation
- whether Windows Update may reinstall an NVIDIA driver during this process
-->

<!-- TODO(tech-review):
Confirm the approved public versions, purpose, source, download URL, and integrity-check method for:
- AGBOX4_WIN1124H2EN20260527
- CleanupTool_1.0.21.0
- NVIDIA_APP_11.0.5.420_1146713
- NVIDIA driver 610.74
Also confirm whether Olares recommends an OEM-certified driver for the built-in RTX 5090M.
-->

## Set up from a fresh Windows installation

:::warning Back up your data
Installing Windows can remove the existing operating system, apps, settings, and files from the Windows partition. Back up any data you want to keep before continuing.
:::

1. Install Windows 11 24H2. The tested image is `AGBOX4_WIN1124H2EN20260527`.
2. Run `CleanupTool_1.0.21.0` and remove all NVIDIA apps and drivers.

   <!-- TODO(tech-review): Explain why the tested process removes the NVIDIA software included with the Windows image before reinstalling it. -->

3. Install the GPU in the enclosure and turn on the enclosure.
4. Connect it to the Thunderbolt 5 (USB-C) port on Olares One with a certified Thunderbolt 5 cable.

   Connect the eGPU while Windows is running only at this point in the tested initial setup. Routine hot-plugging has not been verified.

5. Run `NVIDIA_APP_11.0.5.420_1146713` to install NVIDIA App.
6. In NVIDIA App, install the graphics driver. Version 610.74 was used in testing.
7. After the driver installation is complete, continue to [Check the connection](#check-the-connection).

## Recover the built-in GPU

Use these steps if the built-in GPU shows a warning or disappears from NVIDIA App after you connect the eGPU.

1. Open **NVIDIA App** > **Drivers** > **Reinstall**.
2. Select **Custom installation**.
3. Select **Perform a clean installation**, then finish the installation.

   :::warning Do not skip the restart
   In the tested configuration, the eGPU remained at Gen1 until Windows was restarted. Restart Windows to let the link renegotiate at Gen4.
   :::

4. Restart Windows.

<!-- TODO(tech-review): Confirm that the NVIDIA App labels above match the approved public version. -->

## Check the connection

- In **Device Manager** > **Display adapters**, confirm that both the built-in GPU and external GPU appear without warning icons.
- In GPUMon, confirm that the eGPU link operates at Gen4.

<!-- TODO(tech-review): Provide the approved GPUMon source or download location and identify the exact field users should check. -->

## Tested performance

In the tested configuration, the RTX 4060 Ti reached 165 W under FurMark load. GPUMon reported a Gen4 link and no PCIe errors during the test.

Results may vary with the GPU, enclosure, power supply, driver, and workload. Running FurMark is not required to complete the setup.

## Known issues

| Symptom | Status |
|---|---|
| Built-in RTX 5090M reaches about 95 W instead of 175 W | This is a known issue confirmed with NVIDIA, and an upstream fix is pending. In the tested configuration, the external RTX 4060 Ti remained stable at Gen4 and its operation was not affected. |
| Built-in GPU reports an error after connecting the eGPU | Perform the clean installation above and restart. |
| Two or more external GPUs are connected | Multi-eGPU configurations have not been verified. |

<!-- TODO(tech-review): Confirm the status of configurations with two or more external GPUs. Do not count the built-in GPU plus one external GPU as a multi-eGPU configuration. Clarify whether previous testing confirmed only detection or also stability under load. -->
<!-- TODO(tech-review): Confirm that 175 W is the intended full-load power target for the built-in RTX 5090M in this context. -->

## Related pages

- [Olares One eGPU support overview](./egpu.md)
- [Troubleshoot eGPU issues](./ts-egpu.md)
