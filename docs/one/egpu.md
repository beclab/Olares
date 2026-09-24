---
outline: [2, 3]
description: Learn what an eGPU is, check known hardware compatibility, and connect an NVIDIA eGPU to Olares One.
head:
  - - meta
    - name: keywords
      content: Olares One, eGPU, external GPU, Thunderbolt 5, NVIDIA
---

# Connect an eGPU to Olares One

You can add GPU capacity to Olares One by connecting an external GPU (eGPU) over Thunderbolt.

:::danger Shut down before connecting on Olares OS
Do not connect or disconnect an eGPU while Olares OS is running. Shut down Olares One, power on and connect the eGPU, and then start Olares One.
:::

## Check known compatibility

The table lists available compatibility results, not recommended products or every possible combination. Results also depend on the dock or enclosure, power supply, cable, operating system, and NVIDIA driver. An unlisted combination may work, but has not been verified.

| External graphics hardware | Olares OS | Windows 11 |
|---|---|---|
| AOOSTAR EG02 eGPU dock + RTX 4060 Ti | **Works after setup.** Install the [Gen1 workaround](./egpu-olares-os.md#install-the-workaround-if-needed). | **Works after setup.** Reinstall the NVIDIA driver with the eGPU connected. |
| Razer Core X V2 eGPU enclosure + RTX 4090 | **Under verification.** Startup was unstable without the Gen1 workaround. Results with the workaround are not verified. | **Verified only in the tested multi-eGPU setup.** Standalone use has not been verified. |
| Razer Core X eGPU enclosure + RTX 4060 | **Not verified.** | **Verified only in the tested multi-eGPU setup.** Standalone use has not been verified. |
| eGPU dock or enclosure + desktop RTX 5090 | **Not supported.** The NVIDIA driver did not initialize, and no workaround is available. | **Not verified.** |

In this table, **Works** and **Verified** mean that the GPU completed a workload test. Detection in Dashboard, Device Manager, or `nvidia-smi` alone is not treated as a compatibility result.

:::info Multiple eGPUs
If you need multiple eGPUs, use Windows 11. Olares OS currently supports only one eGPU at a time.

We have verified one Windows 11 configuration with three eGPUs connected through a Razer Thunderbolt 5 Dock:

- AOOSTAR EG02 with RTX 4060 Ti
- Razer Core X V2 with RTX 4090
- Razer Core X with RTX 4060

In this configuration, the built-in RTX 5090M was limited to about 95 W instead of 175 W.

Multi-eGPU compatibility depends on the dock and connection topology. Start with one eGPU, then connect and verify additional devices one at a time.
:::

## Choose hardware

Prepare one of the following:

- A Thunderbolt external graphics device with a GPU already installed and its power adapter.
- A Thunderbolt eGPU dock or enclosure, a desktop NVIDIA GPU based on the Turing architecture or newer, and a power supply that meets the GPU requirements.

:::info Thunderbolt 5
Use a certified Thunderbolt 5 cable, preferably the cable supplied with the external graphics device.

Older Thunderbolt external graphics devices may work, but Thunderbolt 5 is recommended.
:::

## Set up your eGPU

Choose the guide for your operating system:

- [Set up an eGPU on Olares OS](./egpu-olares-os.md)
- [Set up an eGPU on Windows 11](./egpu-windows.md)

## Get help

If Olares One does not start, the eGPU is missing, or a GPU reports a driver error, see [Troubleshoot eGPU issues](./ts-egpu.md).
