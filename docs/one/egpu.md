---
outline: [2, 3]
description: Learn what an eGPU is, check known hardware compatibility, and connect an NVIDIA eGPU to Olares One.
head:
  - - meta
    - name: keywords
      content: Olares One, eGPU, external GPU, Thunderbolt 5, NVIDIA
---

# Connect an eGPU to Olares One

You can add GPU capacity to Olares One by connecting an external GPU (eGPU) over Thunderbolt. This page covers the hardware options, known compatibility, and setup paths for Olares OS and Windows 11.

:::danger Shut down before connecting on Olares OS
Do not connect or disconnect an eGPU while Olares OS is running. Shut down Olares One, power on and connect the eGPU, and then start Olares One.
:::

## Check known compatibility

The table lists the single-eGPU hardware combinations for which compatibility results are available. It is not a list of recommended products or a complete compatibility list. An unlisted combination may work, but has not been verified.

| External graphics hardware | Olares OS | Windows 11 |
|---|---|---|
| ROG XG Mobile (2025) with NVIDIA GeForce RTX 5070 Ti Laptop GPU | **Works after connection.** No additional setup was needed after a cold start. | **Not verified.** |
| AOOSTAR EG02 eGPU dock + RTX 4060 Ti | **Works after setup.** Install the [Gen1 workaround](./egpu-olares-os.md#install-the-workaround-if-needed). | **Works after setup.** Reinstall the NVIDIA driver with the eGPU connected. |
| Razer Core X V2 eGPU enclosure + RTX 4090 | **Under verification.** Startup was unstable without the Gen1 workaround. Results with the workaround are not verified. | **Under verification.** |
| eGPU dock or enclosure + desktop RTX 5090 | **Not supported.** The NVIDIA driver did not initialize, and no workaround is available. | **Not verified.** |

Compatibility also depends on the eGPU dock or enclosure, power supply, cable, operating system, and NVIDIA driver.

:::warning Start with one eGPU
Olares OS does not support multiple eGPUs. Windows has detected up to two eGPUs, but multi-eGPU support is still under verification. A chained setup with three eGPUs did not work on either operating system.
:::

## Before you start

- For a desktop GPU, use an NVIDIA GPU based on the Turing architecture or newer. Make sure the eGPU dock or enclosure and its power supply meet the GPU requirements.
- Use a certified Thunderbolt 5 cable. The cable supplied with the external graphics device is recommended.
- For the first setup, connect one eGPU directly to Olares One. Disconnect other high-bandwidth Thunderbolt devices and do not route the eGPU through another dock.

:::info Older Thunderbolt hardware
Older Thunderbolt external graphics devices may work, but Thunderbolt 5 is recommended.
:::

## Set up your eGPU

### Olares OS

Start with a cold connection. Shut down Olares One, power on and connect the eGPU, and then start Olares One. If the eGPU is detected and remains available, no other setup is needed.

Follow [Set up an eGPU on Olares OS](./egpu-olares-os.md). The guide includes a Gen1 workaround for startup, detection, or disconnection problems.

Olares OS does not support hot-plugging.

### Windows 11

You can use an existing Windows 11 installation. The first setup requires reinstalling the NVIDIA driver with the eGPU connected.

Follow [Set up an eGPU on Windows](./egpu-windows.md). The guide covers one eGPU and does not cover hot-plugging.

## Get help

If Olares One does not start, the eGPU is missing, or a GPU reports a driver error, see [Troubleshoot eGPU issues](./ts-egpu.md).
