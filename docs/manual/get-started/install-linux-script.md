---
outline: [2, 3]
description: Detailed instructions for installing Olares on Linux systems including Ubuntu and Debian. Covers system requirements, installation steps, and activation process.
---
# Install Olares on Linux via the script
This guide explains how to install Olares on Linux using the provided installation script.

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
- **Supported systems**:
  - Ubuntu 22.04-25.04 LTS
  - Debian 12 or 13

:::info Version compatibility
While these specific versions are confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::

### Optional hardware

A GPU is not required to install Olares, but is necessary for AI applications.

- **GPU (NVIDIA only)**:
  - **Architecture**: Turing or newer (e.g., GTX 16 series, RTX 20 series).
  - **Verification**: Run `lspci | grep -i nvidia` and check the [compatible GPU table](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus).

## Install Olares

In your terminal, run the following command:

<!--@include: ./reusables.md{4,36}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{38,42}-->