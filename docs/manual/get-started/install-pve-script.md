---
description: Guide to installing Olares on Proxmox VE (PVE) with system requirements, installation commands, and step-by-step activation instructions.
---
# Install Olares on PVE via the script
Proxmox Virtual Environment (PVE) is an open-source virtualization platform based on Debian Linux. This guide explains how to install Olares in a PVE environment using the provided installation script.

:::warning Not recommended for production use
Currently, Olares on PVE has certain limitations. We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md{44,51}-->

## System requirements
Make sure your device meets the following requirements.

### Required specifications
- CPU: At least 4 cores
- RAM: At least 8 GB of available memory
- Storage: At least 150 GB of available SSD storage. 
    ::: warning SSD required
    The installation will likely fail if an HDD (mechanical hard drive) is used instead of an SSD.
    :::
- Supported Systems: PVE 8.2.2

:::info Version compatibility
While the specific version is confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::

### Optional hardware

A GPU is not required to install Olares, but is necessary for AI applications.

- GPU (NVIDIA only):
   - Architecture: Turing or newer (e.g., GTX 16 series, RTX 20 series, and later).
   - Verification: Run `lspci | grep -i nvidia` in the PVE host shell to confirm the card is detected.
   - Setup: To utilize the GPU, you must configure PCI passthrough. Please refer to [Configure GPU passthrough in PVE](/manual/best-practices/install-olares-gpu-passthrough.md#configure-gpu-passthrough-in-pve) for detailed instructions.

## Install on PVE

In PVE CLI, run the following command:

<!--@include: ./reusables.md{4,37}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{38,42}-->