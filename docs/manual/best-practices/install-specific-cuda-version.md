---
outline: [2, 3]
description: Learn how to install a specific NVIDIA CUDA driver version on the Olares host when the latest official release does not meet your needs.
---

# Install a specific CUDA version

To run NVIDIA GPU-based applications on Olares, the host and application containers both need CUDA drivers installed. While the two versions generally need to match, applications can usually run even when the container's CUDA version is higher than the host's.

Olares officially maintains only the latest CUDA version to support cutting-edge AI applications. However, you may need a different version in the following cases:

- A specific application or AI model requires a particular CUDA or driver version.
- You prefer to lock the version for stability and avoid automatic upgrades.
- The latest driver has compatibility issues with your workload.

In these situations, you can manually install a specific CUDA driver version on the host.

## Prerequisites

Before you start, ensure that your setup meets the following requirements:

- An Olares installation with GPU support enabled
- A compatible NVIDIA GPU
- Root or sudo access to the Olares host

## Step 1: Download the driver runfile

1. Visit the [NVIDIA driver downloads](https://www.nvidia.com/en-us/drivers/) page.
2. Select your GPU product type, series, and model, then choose **Linux 64-bit** as the operating system.
3. Click **Search** and note the driver version number shown in the results (for example, `575.64.05`).
4. On the Olares host, replace the version number in the following command with the one you found, then run it to download the runfile:

```bash
curl -sSOL https://us.download.nvidia.com/XFree86/Linux-x86_64/575.64.05/NVIDIA-Linux-x86_64-575.64.05.run
chmod +x NVIDIA-Linux-x86_64-575.64.05.run
```

## Step 2: Run the installer

Execute the runfile with root privileges, then reboot the host:

```bash
sudo ./NVIDIA-Linux-x86_64-575.64.05.run
sudo reboot now
```

::: warning A reboot is required
You must reboot the host after installing the driver for the changes to take effect.
:::

## Step 3: Update GPU status in Olares

After the host restarts, run the following command to update the node's CUDA and driver version information in Olares:

```bash
olares-cli gpu enable
```

## Step 4: Verify the installation

Check that the new CUDA version is active:

```bash
nvidia-smi
```

If the installation is successful, the output shows the installed driver version and CUDA version. For example:

```
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 575.64.05              Driver Version: 575.64.05      CUDA Version: 12.9     |
+-----------------------------------------+------------------------+----------------------+
```
