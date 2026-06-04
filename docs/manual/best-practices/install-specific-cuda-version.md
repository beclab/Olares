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

## Learning objectives

By the end of this tutorial, you will learn how to:

- Check the current CUDA and driver version on your Olares host.
- Download and install a specific NVIDIA driver version from a runfile.
- Update the GPU status in Olares after installing a new driver.

## Prerequisites

Before you start, ensure that your setup meets the following requirements:

- An Olares device with GPU support enabled, and network access
- Root or sudo access to the Olares host

## Check the current CUDA version

Run the following command on the Olares host to check the current driver and CUDA version:

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

In this example, the current driver version is `590.44.01` and the CUDA version is `13.1`.

::: tip
If you only know the target CUDA version, look up the matching driver version in the [NVIDIA CUDA release notes](https://docs.nvidia.com/cuda/cuda-toolkit-release-notes/index.html#cuda-major-component-versions).
:::

## Download and install the driver

### Step 1: Download the driver runfile

1. Visit the [NVIDIA driver downloads](https://www.nvidia.com/en-us/drivers/) page.
2. Select your GPU product type, series, and model, then choose **Linux 64-bit** as the operating system.
3. Click **Find** and note the driver version number shown in the results. For example, `580.95.05`, which corresponds to CUDA 13.0.
4. On the Olares host, run the following commands to download the runfile. Replace `580.95.05` with the driver version you found:

    ::: code-group
    ```bash [curl]
    VERSION=580.95.05
    curl -sSOL https://us.download.nvidia.com/XFree86/Linux-x86_64/${VERSION}/NVIDIA-Linux-x86_64-${VERSION}.run
    ```
    ```bash [wget]
    VERSION=580.95.05
    wget https://us.download.nvidia.com/XFree86/Linux-x86_64/${VERSION}/NVIDIA-Linux-x86_64-${VERSION}.run
    ```
    :::

5. Make the runfile executable:

    ```bash
    chmod +x NVIDIA-Linux-x86_64-580.95.05.run
    ```

### Step 2: Run the installer

1. Execute the runfile with root privileges:

    ```bash
    sudo ./NVIDIA-Linux-x86_64-580.95.05.run
    ```

2. When the installer prompts you to choose a kernel module type, select **NVIDIA Proprietary**.
3. Follow the on-screen prompts to continue the installation until you are asked to reboot the system.
4. Reboot the host:

    ```bash
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
| NVIDIA-SMI 580.95.05              Driver Version: 580.95.05      CUDA Version: 13.0     |
+-----------------------------------------+------------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
|                                         |                        |               MIG M. |
|=========================================+========================+======================|
|   0  NVIDIA GeForce RTX 4060 Ti     Off |   00000000:01:00.0 Off |                  N/A |
|  0%   41C    P0             28W /  165W |       0MiB /  16380MiB |      0%      Default |
|                                         |                        |                  N/A |
+-----------------------------------------+------------------------+----------------------+

+-----------------------------------------------------------------------------------------+
| Processes:                                                                              |
|  GPU   GI   CI              PID   Type   Process name                        GPU Memory |
|        ID   ID                                                               Usage      |
|=========================================================================================|
|  No running processes found                                                             |
+-----------------------------------------------------------------------------------------+
```

In this example, the CUDA version is `13.0` and the driver version is `580.95.05`.
