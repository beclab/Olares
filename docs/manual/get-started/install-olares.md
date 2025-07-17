---
description: Get started with Olares on Linux using the one-line script
---
:::warning Note for Mainland China users
The steps in this guide differ for users in Mainland China due to regional differences. For a version tailored to your region, please read the Simplified Chinese documentation.
:::

# Install Olares on Linux

This document introduces how to install and activate Olares on Linux. **Linux** (Ubuntu or Debian) is the recommended platform for running Olares, as it offers the best performance and stability in production environments.

Before installing, make sure to [create an Olares ID](create-olares-id.md) and verify that your operating system and hardware meet the minimum requirements.

:::info Having trouble with installation?  
If you encounter issues during the installation process, feel free to [submit a GitHub Issue](https://github.com/beclab/Olares/issues/new). Please include the following information when submitting: 

- The platform or environment you're using (e.g., Ubuntu, Docker, WSL, etc.).  
- The installation method (script installation or Docker image).  
- Detailed error information (including logs, error messages, or screenshots).  
:::

## System requirements

Make sure your device meets the following requirements.

- CPU: At least 4 cores
- RAM: At least 8GB of available memory
- Storage: At least 64GB of available space (SSD recommended)
- Supported systems:
    - Ubuntu 20.04 LTS or later
    - Debian 11 or later

:::info Version compatibility
While these specific versions are confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::

## Install Olares

In your terminal, run the following command:

<!--@include: ./reusables.md{4,33}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{35,39}-->
