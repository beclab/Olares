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

| Method                                               | Description |
|------------------------------------------------------|--------------|
| [**Using ISO image**](install-linux-iso.md)          | Fresh ISO install on a physical machine, automatically configuring the <br/>host (Linux) environment, container runtime, drivers, and core <br/>dependencies. |
| [**Using one-line script**](install-linux-script.md) | Quick install on existing Linux systems. |

:::tip Recommendation
The ISO installation method ensures maximum compatibility, performance, and system-level optimization.
:::

### Alternative installation methods

These methods are suitable for **development**, **testing**, or **lightweight environments**. 

#### Linux

- [**Using Docker Compose**](install-linux-docker.md): Runs Olares in a containerized environment using Docker Compose in Linux.


#### Windows

| Method                                                 | Description                                              |  
|--------------------------------------------------------|----------------------------------------------------------| 
| [**Using one-line script**](install-windows-script.md) | Installs Olares in Windows Subsystem for Linux 2 (WSL 2). |
| [**Using Docker image**](install-windows-docker.md)           | Runs Olares in Docker container with WSL 2 integration.  |

#### macOS

| Method                                         | Description                                                     |  
|------------------------------------------------|-----------------------------------------------------------------| 
| [**Using one-line script**](install-mac-script.md) | Installs Olares within a containerized environment via MiniKube. |  
| [**Using Docker image**](install-mac-docker.md) | Runs Olares in Docker on macOS.                          |  
---

#### PVE

| Method                                          | Description | 
|-------------------------------------------------|--------------| 
| [**Using ISO image**](install-pve-iso.md)             | Deploys Olares as a full VM in Proxmox using the ISO installer. | 
| [**Using one-line script**](install-pve-script.md) | Installs Olares directly on a PVE node. | 

 #### Raspberry Pi (ARM)

- [**Using one-line script**](install-raspberry-pi.md): Installs Olares on ARM-based Raspberry Pi boards. 

