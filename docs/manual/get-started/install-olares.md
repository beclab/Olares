---
description: Install Olares on Linux systems such as Ubuntu or Debian for optimal performance and stability in production environments. Supports deployment via ISO image or one-line script.
---
# Install Olares on Linux

This guide explains how to install Olares on a Linux system.  
We recommend deploying Olares on **Linux systems** (such as Ubuntu or Debian) to achieve the best performance and stability in production environments.

Before starting, please [create your Olares ID](create-olares-id.md) and make sure your operating system and hardware meet the minimum requirements for your chosen installation method.

## Installation methods

Depending on your hardware and usage scenario, Olares provides two installation methods:

| Installation Method                 | Description                                                                                                                                                                   | Recommended for |
|-------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|
| [ISO image](install-iso)            | Installs Olares from the official ISO <br/> boot image, automatically configuring <br/> the host (Linux) environment, container<br/> runtime, drivers, and core dependencies. | Ideal for deploying Olares on physical machines or virtualized environments (e.g., PVE). |
| [One-line command](install-command) | Installs Olares by running a one-line <br/>command on an existing Linux system.                                                                                               | Suitable for users who prefer manual deployment in an existing Linux environment. |

:::info Having installation issues?
If you encounter any issues during installation, [submit a GitHub Issue](https://github.com/beclab/Olares/issues/new). Please include:
- The platform or environment used (e.g., Ubuntu, Docker, WSL).
- Installation method (script or Docker image).
- Detailed error messages, logs, or screenshots.
  :::
