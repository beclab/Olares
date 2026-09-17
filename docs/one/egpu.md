---
outline: [2, 3]
description: Check Olares One eGPU compatibility, tested hardware, current limitations, and setup options for Olares OS and Windows.
head:
  - - meta
    - name: keywords
      content: Olares One, eGPU, external GPU, Thunderbolt 5, NVIDIA
---

# Olares One eGPU support overview

Use this page to check whether your eGPU setup has been tested and choose the right setup guide. Results depend on the enclosure, GPU, power supply, cable, operating system, and NVIDIA driver.

The tables below show verified, unverified, and unsupported configurations, along with configurations that remain under investigation. An unlisted combination may still work, but it has not been evaluated.

:::warning Connect before startup
For Olares OS, do not connect or disconnect an eGPU while the system is running. Shut down Olares One completely, power the enclosure, connect the cable, and then start Olares One.
:::

## Requirements

Before using an eGPU, make sure you have:

- A Thunderbolt eGPU enclosure with its own power supply.
- A certified Thunderbolt 5 cable, preferably the cable supplied with the enclosure.
- An NVIDIA GPU supported by your operating system and NVIDIA driver.

Meeting the requirements does not mean every GPU and enclosure combination has been verified. Check the tables below for tested configurations.

For the first test, connect only one eGPU and disconnect other high-bandwidth Thunderbolt devices.

<!-- TODO(tech-review): Confirm the minimum supported NVIDIA GPU architecture before adding it to the requirements. -->

## Support status

Each result is labeled **Verified**, **Requires setup**, **Not verified**, **Under investigation**, or **Not supported**.

| Setup | Windows | Olares OS |
|---|---|---|
| AOOSTAR EG02 + RTX 4060 Ti, connected before startup | **Verified.** | **Requires setup.** Apply the [Gen1 workaround](./egpu-olares-os.md). |
| Razer Core + RTX 4090 | **Not verified.** | **Not verified.** Startup was unstable without the workaround. This combination has not been tested with it. |
| Built-in RTX 5090M + AOOSTAR EG02 with RTX 4060 Ti | **Requires setup.** Perform a clean driver installation. The built-in GPU also has a known power issue confirmed with NVIDIA. | **Requires setup.** Apply the workaround. Both GPUs worked together in the tested configuration. |
| Desktop RTX 5090 | **Not verified.** | **Under investigation.** Driver initialization failed in the tested configuration. PCIe resource allocation is being investigated. |
| Multiple eGPUs | **Not verified.** Multi-eGPU support is still being evaluated. | **Not supported.** Connect only one eGPU at a time. |
| Hot-plugging | **Not verified.** | **Not supported.** Connect and power the eGPU before startup. |

<!-- TODO(tech-review): Confirm the Windows hot-plug configuration before documenting it. -->
<!-- TODO(tech-review): Confirm whether multi-eGPU testing means multiple external GPUs or the built-in GPU plus one external GPU. -->

## Set up your eGPU

- [Set up an eGPU on Olares OS](./egpu-olares-os.md)
- [Set up an eGPU on Olares One with Windows](./egpu-windows.md)
- [Troubleshoot eGPU issues on Olares One](./ts-egpu.md)

## Tested combinations

### Olares OS

| Enclosure | GPU | Result | Recommendation |
|---|---|---|---|
| AOOSTAR EG02 | RTX 4060 Ti | Stable after applying the workaround. No further disconnects were observed. | Follow the [Olares OS setup](./egpu-olares-os.md). |
| AOOSTAR EG02 | RTX 4060 Ti + built-in RTX 5090M | Both GPUs worked and could be assigned separately to AI apps. | Apply the workaround. |
| Razer Core | RTX 4090 | Startup is unstable without the workaround. The configuration has not been tested with it. | No recommendation is available until testing is complete. |
| Not specified | Desktop RTX 5090 | Driver initialization failed in the tested configuration. PCIe resource allocation is under investigation. | Not currently recommended. |
| Chained Thunderbolt enclosures | RTX 3090 + 2× RTX 2080 Ti | The tested multi-eGPU configuration did not work. | Connect only one eGPU. |

### Windows

| Enclosure | GPU | Result | Recommendation |
|---|---|---|---|
| AOOSTAR EG02 | RTX 4060 Ti | Reached 165 W at stable Gen4 speeds with no observed PCIe errors. | Follow the [Windows setup](./egpu-windows.md). |
| Chained Thunderbolt enclosures | RTX 3090 + 2× RTX 2080 Ti | The tested multi-eGPU configuration did not work. | Multi-eGPU configurations are not verified. |

<!-- TODO(tech-review): Confirm which enclosure was used for the desktop RTX 5090 test. -->

## Current limitations

- **Gen4 on Olares OS**
  - **Symptom**: Startup may stall, the GPU may not be detected, or the eGPU may disconnect while in use.
  - **Possible cause**: NVIDIA driver initialization involves timing-sensitive communication. The tested Thunderbolt path is less stable at Gen4. This explanation has not been confirmed as the root cause.
  - **What to do**: Use the [Gen1 workaround](./egpu-olares-os.md).

- **Hot-plugging on Olares OS**
  - **Symptom**: The eGPU may not be detected, or the system may stop responding when the eGPU is connected after startup.
  - **Possible cause**: Creating the Thunderbolt tunnel while the system is running may interfere with driver initialization.
  - **What to do**: Shut down Olares One, power and connect the eGPU, and then start Olares One.

- **Multiple eGPUs**
  - **Symptom**: Additional eGPUs may not be detected.
  - **Possible cause**: Each additional enclosure and GPU increases PCIe address-space and power requirements.
  - **What to do**: On Olares OS, connect only one eGPU at a time. Multi-eGPU configurations on Windows have not been verified.

- **Desktop GPUs with large amounts of VRAM**
  - **Symptom**: The GPU may appear on the PCI bus, but the driver may fail to initialize.
  - **Possible cause**: PCIe resource allocation is under investigation. The system may not reserve enough address space for the GPU.
  - **What to do**: No verified workaround is currently available.

- **Built-in RTX 5090M power on Windows**
  - **Symptom**: The built-in GPU reached about 95 W instead of 175 W under full load.
  - **Cause**: This is a known issue confirmed with NVIDIA. The eGPU link itself remains stable at Gen4.
  - **What to do**: Wait for an upstream fix. This issue does not affect eGPU use.

## Terms

| Term | Meaning |
|---|---|
| Gen1 / Gen4 | PCIe link-speed generations. At the same link width, Gen4 provides about eight times the raw transfer rate of Gen1. |
| Thunderbolt tunnel | A path that carries PCIe traffic over Thunderbolt between Olares One and the eGPU enclosure. |
| BAR / video-memory mapping | A PCIe address window that lets the system access GPU resources, including portions of video memory. |
| Cold start / hot-plug | A cold start means powering and connecting the eGPU before starting Olares One. Hot-plugging means connecting or disconnecting it while the system is running. |
| GPU disconnect | The eGPU stops responding or is no longer visible to the operating system while in use. |
