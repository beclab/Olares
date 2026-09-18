---
description: Compare setup paths and installation requirements for Olares One, self-hosted Linux hardware, and NVIDIA DGX Spark.
outline: [2,4]
head:
  - - meta
    - name: keywords
      content: Olares, install Olares, Olares One, system requirements, Linux, Ubuntu, Debian, DGX Spark
---

# Install Olares

Choose your device first, then follow the setup path for that hardware.

## Olares One

Olares One has dedicated setup and recovery workflows that preserve its hardware-specific features.

:::warning Use the Olares One setup guides
Do not install Olares One using the generic Linux ISO image or one-line script below. The device may be recognized as generic hardware, and some Olares One features may be unavailable.
:::

| Task | Recommended guide |
| --- | --- |
| Set up a new Olares One | [First boot](/one/first-boot) |
| Reinstall or recover Olares OS | [Olares One ISO](/one/create-bootable-usb) |
| Install and activate from the host terminal | [Olares CLI](/manual/best-practices/activate-olares-using-cli) **(Advanced)** |

## Linux

Use the following methods to install Olares on your own compatible hardware. Linux is recommended for production deployments.

### System requirements

:::warning SSD required
Installation on an HDD will fail.
:::

| Item | Requirement |
| --- | --- |
| CPU | 4 cores or more |
| Memory | 8 GB or more |
| Storage | 150 GB or more on an SSD |

### Optional GPU

A GPU is not required to install Olares, but most AI apps need one. Only NVIDIA GPUs are supported.

| Item | Requirement |
| --- | --- |
| Architecture | Turing or newer, including GTX 16xx and RTX 20xx, 30xx, 40xx, and 50xx series |
| VRAM | 8 GB or more recommended |

<!--@include: ./gpu-requirements.md#gpu-compatibility-check-->

### Installation methods

| Method | Best for |
| --- | --- |
| [**ISO image**](install-linux-iso.md) **(Recommended)** | A fresh installation on a physical machine with an Intel or AMD x86-64 processor |
| [**One-line script**](install-linux-script.md) | An existing Ubuntu 22.04–25.04 or Debian 12/13 system |
| [**Docker Compose**](install-linux-docker.md) | A containerized installation on Ubuntu 22.04–25.04 or Debian 12/13 |

## DGX Spark

NVIDIA DGX Spark is a compact AI development platform featuring a high-performance GPU. Olares is optimized to leverage the full capabilities of DGX Spark hardware.

### Installation methods

| Method | Best for |
| --- | --- |
| [**One-line script**](install-spark-script.md) **(Recommended)** | Installing from the existing DGX OS with at least 150 GB of available SSD storage |
| [**ISO image**](install-spark-iso.md) | A fresh installation from a bootable USB drive |
