---
description: Guide to installing Olares on Proxmox VE (PVE) with system requirements, installation commands, and step-by-step activation instructions.
---
# Install Olares on PVE via the script
Proxmox Virtual Environment (PVE) is an open-source virtualization platform based on Debian Linux. This guide explains how to install Olares in a PVE environment using the provided installation script.

:::warning Not for production use
Currently, Olares on PVE has certain limitations. We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md{44,51}-->

## System requirements

Make sure your device meets the following requirements.

### Required specifications

- **CPU**: At least 4 cores.
- **RAM**: At least 8 GB of available memory.
- **Storage**: At least 150 GB of available SSD storage.
  :::warning SSD required
  The installation will fail if an HDD (mechanical hard drive) is used instead of an SSD.
  :::
- **Supported systems**: PVE 8.2.2

:::info Version compatibility
While the specific version is confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::

### Optional hardware

<!--@include: ./gpu-requirements.md{5,}-->

:::tip PCI passthrough required
To use the GPU within Olares on PVE, you must configure PCI passthrough first. Refer to [Configure GPU passthrough in PVE](/manual/best-practices/install-olares-gpu-passthrough.md#configure-gpu-passthrough-in-pve) for instructions.
:::

## Install on PVE

In PVE CLI, run the following command:

<!--@include: ./reusables.md{4,37}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{38,42}-->