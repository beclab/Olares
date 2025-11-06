---
description: Overview of supported Olares installation methods. Recommended for Linux environments via ISO image or installation script. Other platforms like macOS, Windows, PVE, and Raspberry Pi are supported for testing and development.
outline: [2,4]
---

# Install Olares

This page provides an overview of supported installation methods for Olares.

Before installation, make sure you have:
- Created your [Olares ID](create-olares-id.md).
- Verified your operating system and hardware meet the minimum requirements as described in the specific guide.

## Choosing the right method

Olares supports multiple platforms and deployment methods. Choose the installation method that best fits your environment.

### Recommended for production

Linux (Ubuntu or Debian) is the recommended platform for running Olares, as it offers the best performance and stability in production environments.

- [**ISO image**](install-linux-iso.md): Perform a fresh installation on a physical machine, auto<br/> configuring the Linux host environment, container runtime,<br/> drivers, and core dependencies.
- [**One-line script**](install-linux-script.md): Quick install on existing Linux systems.                                   

:::tip Recommendation
The ISO installation method ensures maximum compatibility, performance, and system-level optimization.
:::

### Alternative installation methods

These methods are suitable for **development**, **testing**, or **lightweight environments**. 

#### Windows
- [**One-line script**](install-windows-script.md): Install Olares in Windows Subsystem for Linux 2 (WSL 2).
<!--  [**Docker image**](install-windows-docker.md): Run Olares in Docker with WSL 2 integration. -->

#### macOS
- [**One-line script**](install-mac-script.md): Install Olares in a containerized environment via MiniKube.
- [**Docker image**](install-mac-docker.md): Run Olares in Docker on macOS.

#### PVE
- [**ISO image (recommended)**](install-pve-iso.md): Deploy Olares as a full VM in Proxmox VE using the ISO installer.
- [**One-line script**](install-pve-script.md): Install Olares directly on a PVE node.
- [**LXC container**](install-lxc.md): Deploy Olares in Proxmox VE using Linux containers (LXC).

#### Raspberry Pi (ARM)
- [**One-line script**](install-raspberry-pi.md): Install Olares on ARM-based Raspberry Pi devices.
