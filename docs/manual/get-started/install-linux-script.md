---
outline: [2, 3]
description: Detailed instructions for installing Olares on Linux systems including Ubuntu and Debian. Covers system requirements, installation steps, and activation process.
---
# Install Olares on Linux via the script
This guide explains how to install Olares on Linux using the provided installation script.

<!--@include: ./reusables.md#installation-troubleshooting-tip-->

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

<!--@include: ./reusables.md#version-compatibility-->

### Optional hardware

<!--@include: ./gpu-requirements.md#gpu-requirements-->

## Install Olares

In your terminal, run the following command:

<!--@include: ./reusables.md#install-script-command-->

<!--@include: ./reusables.md#root-password-tip-->

<!--@include: ./reusables.md#installation-error-tip-->

<!--@include: ./reusables.md#prepare-wizard-heading-->

<!--@include: ./reusables.md#prepare-wizard-details-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md#protect-olares-id-->