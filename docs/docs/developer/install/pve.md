---
description: Guide to installing Olares on Proxmox VE (PVE) with system requirements, installation commands, and step-by-step activation instructions.
---
# Install Olares on PVE
Proxmox Virtual Environment (PVE) is an open-source virtualization platform based on Debian Linux. This guide explains how to install Olares in a PVE environment.

:::warning Not recommended for production use
Currently, Olares on PVE has certain limitations. We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md{41,47}-->

## System requirements
Make sure your device meets the following requirements.

- CPU: At least 4 cores
- RAM: At least 8GB of available memory
- Storage: At least 64GB of available space (SSD recommended)
- Supported Systems: PVE 8.2.2

:::info Version compatibility
While the specific version is confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::

## Install on PVE

In PVE CLI, run the following command:

<!--@include: ./reusables.md{4,33}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{35,39}-->