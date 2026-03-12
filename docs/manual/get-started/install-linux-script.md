---
outline: [2, 3]
description: Detailed instructions for installing Olares on Linux systems including Ubuntu and Debian. Covers system requirements, installation steps, and activation process.
---
# Install Olares on Linux via the script
This guide explains how to install Olares on Linux using the provided installation script.

<!--@include: ./reusables.md{44,51}-->

## System requirements

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

<!--@include: ./reusables.md{63,65}-->

### Optional hardware

<!--@include: ./gpu-requirements.md{5,}-->

## Install Olares

In your terminal, run the following command:

<!--@include: ./reusables.md{4,36}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{38,42}-->