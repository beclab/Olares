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

- An Olares device with GPU support enabled, and network access
- Root or sudo access to the Olares host

## Check current CUDA version
Each NVIDIA driver release bundles a specific CUDA runtime version. 

Run the following command to check the current driver version and CUDA version on your Olares device. Use this information to identify your target nvidia driver version.
```bash
nvidia-smi
```

Example output:
```bash
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 590.44.01              Driver Version: 590.44.01      CUDA Version: 13.1     |
+-----------------------------------------+------------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
|                                         |                        |               MIG M. |
|=========================================+========================+======================|
|   0  NVIDIA GeForce RTX 4060 Ti     Off |   00000000:01:00.0 Off |                  N/A |
|  0%   41C    P8              8W /  165W |   11256MiB /  16380MiB |      0%      Default |
|                                         |                        |                  N/A |
+-----------------------------------------+------------------------+----------------------+

+-----------------------------------------------------------------------------------------+
| Processes:                                                                              |
|  GPU   GI   CI              PID   Type   Process name                        GPU Memory |
|        ID   ID                                                               Usage      |
|=========================================================================================|
|    0   N/A  N/A           60935      C   ./koboldcpp                             242MiB |
+-----------------------------------------------------------------------------------------+
```

In this case, the current driver version is `590.44.01`, and the CUDA version is `13.1`.

:::tip
If you only know the target CUDA version, look up the matching driver version in the [NVIDIA CUDA release notes](https://docs.nvidia.com/cuda/cuda-toolkit-release-notes/index.html#cuda-major-component-versions).
:::

## Install a specific CUDA version
### Step 1: Download the driver runfile

1. Visit the [NVIDIA driver downloads](https://www.nvidia.com/en-us/drivers/) page.
2. Select your GPU product type, series, and model, then choose **Linux 64-bit** as the operating system.
3. Click **Find** and note the driver version number shown in the results (for example, `580.95.05`).
4. On the Olares host, run the following commands to download the runfile. Replace `575.64.05` with the driver version you found:
::: code-group
```bash [curl]
curl -sSOL https://us.download.nvidia.com/XFree86/Linux-x86_64/580.95.05/NVIDIA-Linux-x86_64-580.95.05.run
```
```bash [wget]
wget https://us.download.nvidia.com/XFree86/Linux-x86_64/580.95.05/NVIDIA-Linux-x86_64-580.95.05.run
```
:::
5. 
    ```bash
    chmod +x NVIDIA-Linux-x86_64-580.95.05.run
    ```
### Step 2: Run the installer

Execute the runfile with root privileges, then reboot the host:

```bash
sudo ./NVIDIA-Linux-x86_64-575.64.05.run
sudo reboot now
```

::: warning A reboot is required
You must reboot the host after installing the driver for the changes to take effect.
:::

### Step 3: Update GPU status in Olares

After the host restarts, run the following command to update the node's CUDA and driver version information in Olares:

```bash
olares-cli gpu enable
```

### Step 4: Verify the installation

Check that the new CUDA version is active:

```bash
nvidia-smi
```

If the installation is successful, the output shows the installed driver version and CUDA version. For example:

```bash
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 575.64.05              Driver Version: 575.64.05      CUDA Version: 12.9     |
+-----------------------------------------+------------------------+----------------------+
```
