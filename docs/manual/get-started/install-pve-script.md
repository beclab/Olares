---
description: Guide to installing Olares on Proxmox VE (PVE) with system requirements, installation commands, and step-by-step activation instructions.
---
# Install Olares on PVE via the script
Proxmox Virtual Environment (PVE) is an open-source virtualization platform based on Debian Linux. This guide explains how to install Olares in a PVE environment using the provided installation script.

:::warning Not for production use
Currently, Olares on PVE has certain limitations. We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md#installation-troubleshooting-tip-->

## System requirements

### Required specifications

- **CPU**: At least 4 cores.
- **RAM**: At least 8 GB of available memory.
- **Storage**: At least 200 GB of available SSD storage.
  :::warning SSD required
  The installation will fail if an HDD (mechanical hard drive) is used instead of an SSD.
  :::
- **Supported systems**: PVE 8.2.2

<!--@include: ./reusables.md#version-compatibility-->


### Optional hardware

<!--@include: ./gpu-requirements.md#gpu-requirements-->

:::tip PCI passthrough required
To use the GPU within Olares on PVE, you must configure PCI passthrough first. Refer to [Configure GPU passthrough in PVE](/manual/best-practices/install-olares-gpu-passthrough.md#configure-gpu-passthrough-in-pve) for instructions.
:::

## Install on PVE

In PVE CLI, run the following command:

<!--@include: ./reusables.md#install-script-command-->

<!--@include: ./reusables.md#root-password-tip-->

<!--@include: ./reusables.md#installation-error-tip-->

<!--@include: ./reusables.md#prepare-wizard-heading-->

<!--@include: ./reusables.md#prepare-wizard-details-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md#protect-olares-id-->