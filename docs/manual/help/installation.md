---
outline: [2, 3]
description: Find answers to common questions during the installation and activation of Olares.
---
# Olares installation and activation FAQs

This page lists the FAQs about installing, configuring, and activating Olares on your hardware.

## Installation
### What platforms does Olares support?

You can install Olares on Linux (Ubuntu or Debian) for best performance.

For product evaluation, you can also install Olares on the following platforms:
* Proxmox VE
* Raspberry Pi
* macOS
* Windows

### What is the minimum hardware requirements for installing Olares?

The requirements vary by platform. Generally:
* **CPU**: Minimum 4 cores with x86-64 architecture (Intel or AMD).
* **Memory**: At least 8 GB of available RAM.
* **Storage**: Minimum 150 GB SSD.

For detailed requirements, refer to the [installation docs](../get-started/install-olares.md).

### Can I use a mechanical hard drive to install Olares?

No. You must use an SSD. The installation will likely fail if a mechanical hard drive is used due to slower read and write speeds, which can cause timeouts during the system initialization.

### Does the system support NVIDIA GPUs?

Yes. Olares is fully optimized for NVIDIA hardware. It automatically handles driver installation, allowing you to get immediate AI and gaming performance.

It also supports multiple GPUs on a single motherboard (currently NVIDIA only), allowing users with custom hardware to leverage all available compute power for AI workloads.

### How do I manually install NVIDIA drivers if the automatic setup fails?

The Olares installer usually detects and installs drivers automatically. However, if your system previously had NVIDIA drivers installed, the process might be skipped or fail due to conflicts.

In this case, you should:
1. Reboot the machine after the Olares installation to ensure any old driver components are fully cleared.
2. Manually trigger the driver installation using the command `olares-cli gpu install`.

After installation, you can confirm that the drivers are installed and your GPU is recognized by running `nvidia-smi`.

### Why does installation fail with `failed to build Kubernetes objects` or `Ensure CRDs are installed first`?

While these error messages suggest a problem with Custom Resource Definitions (CRDs), they are often a symptom of poor disk performance.

Olares relies on etcd, the backing database for Kubernetes. etcd is very sensitive to storage speed. If you are installing Olares on a slow disk, such as a traditional HDD, etcd cannot respond fast enough. This causes the API server to time out while attempting to apply CRDs.

Installing Olares on SSD storage should fix this issue.

### My Olares installation timed out and didn't show a password, but the system seems to be running. How can I find the password?
This typically occurs when the installation times out due to insufficient system resources, especially in a virtual machine (VM). You can retrieve the password from the installation log file with the following command:
```bash
# Replace v1.12.2 with your specific Olares version number.
grep password $HOME/.olares/versions/v1.12.2/logs/install.log
```
An installation timeout often means some services failed to start correctly. After finding your password, run  `kubectl get pod -A` to check the status of all services.

## Activation
### Is it possible to activate Olares with a non-local network?

Yes. Before activation, users typically access the activation wizard using the local IP address, which generally requires both parties to be on the same network. However, if Olares is assigned a public IP, such as on a public cloud, this local network limitation no longer applies.

Note that IP access is only used during activation. Once activated, devices can be accessed via domain names on both internal and external networks.

### My Olares is powered on and connected to LAN, but I can't find it in LarePass. What should I do?

Ensure both your phone and Olares device are on the same network. If they are not, LarePass cannot discover Olares automatically.

In situations where you cannot connect via Wi-Fi, you can use the Bluetooth network setup in the LarePass app to connect Olares to the same network as your phone.

For details, see [Activate Olares using Bluetooth](../../manual/larepass/activate-olares.md#activate-olares-using-bluetooth).