---
outline: [2, 3]
description: Configure the GPU mode on Olares One running Ubuntu.
head:
  - - meta
    - name: keywords
      content: Olares One, Ubuntu, GPU mode, performance mode, quiet mode, NVIDIA, nvidia-powerd, ioport
---

# Configure the GPU mode on Ubuntu

Use this guide to configure the GPU mode on Olares One running Ubuntu.

## Prerequisites

- Your Olares One is powered on and running Ubuntu.
- You can access the Ubuntu terminal on your Olares One, either directly or through SSH, and run commands with `sudo`.

## Step 1: Install the NVIDIA driver

The NVIDIA driver allows Ubuntu to communicate with the discrete GPU. Install it before enabling NVIDIA power management or switching the GPU mode.

:::info Use the NVIDIA .run installer
The `.run` installer installs the NVIDIA driver and includes the `nvidia-powerd` service needed for power management. Drivers installed through `apt` do not include this service.
:::

1. In the Ubuntu terminal on your Olares One, download the NVIDIA driver installer:

   ```bash
   wget https://us.download.nvidia.com/XFree86/Linux-x86_64/595.84/NVIDIA-Linux-x86_64-595.84.run
   ```

   :::info Driver version
   This guide uses driver version `595.84`, which is tested on Olares One. Other versions are untested, and the `nvidia-powerd` service may fail to start on some of them.
   :::

2. Make the installer executable:

   ```bash
   chmod +x ./NVIDIA-Linux-x86_64-595.84.run
   ```

3. Run the installer:

   ```bash
   sudo ./NVIDIA-Linux-x86_64-595.84.run
   ```

4. Follow the prompts shown by the installer until the installation completes.

## Step 2: Enable NVIDIA power management

The `nvidia-powerd` service manages power allocation between the CPU and GPU during demanding workloads. Enabling it allows the NVIDIA driver to apply the system's supported power-management behavior.

1. Enable and start the `nvidia-powerd` service:

   ```bash
   sudo systemctl enable --now nvidia-powerd
   ```

2. Check the service status:

   ```bash
   sudo systemctl status nvidia-powerd
   ```

   The service status should show `active (running)`.

   :::info
   On Olares One, you may see `ERROR! JPAC is not created/already destroyed, ignoring request.` because the device has no battery. This message can be ignored if the service is `active (running)`.
   :::

## Step 3: Set the GPU mode

With NVIDIA power management enabled, you can now set the GPU mode using `ioport`, a command-line tool for accessing hardware I/O ports. Choose performance mode (`3`) or quiet mode (`0`).

1. Install `ioport`:

   ```bash
   sudo apt update && sudo apt install -y ioport
   ```

2. Run one of the following commands for the mode you want to use.

   :::warning Olares One only
   These commands access the Olares One Embedded Controller (EC). Use them only on Olares One, and copy the command for your chosen mode exactly as shown.
   :::

   To enable performance mode (`3`):

   ```bash
   sudo outb 0x302 0x60 && sudo outb 0x303 3
   ```

   To enable quiet mode (`0`):

   ```bash
   sudo outb 0x302 0x60 && sudo outb 0x303 0
   ```

3. Check that the GPU power limit matches the selected mode:

   ```bash
   nvidia-smi
   ```

   In the `Pwr:Usage/Cap` column, check the value for `Cap`. It should be `150W` for performance mode (`3`) or `50W` for quiet mode (`0`).

   :::info Performance mode boosts up to 175W
   In performance mode, `nvidia-smi` shows a power limit of `150W` when the GPU is idle. Under full load, the limit dynamically rises to `175W`.
   :::

   **Example output:**

   ```text
   +-----------------------------------------------------------------------------------------+
   | NVIDIA-SMI 595.84                 Driver Version: 595.84         CUDA Version: 13.2     |
   +-----------------------------------------+------------------------+----------------------+
   | GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
   | Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
   |                                         |                        |               MIG M. |
   |=========================================+========================+======================|
   |   0  NVIDIA GeForce RTX 5090 ...  Off |   00000000:02:00.0 Off |                  N/A |
   | N/A   41C    P8              7W /  150W |    9706MiB /  24463MiB |      0%      Default |
   |                                         |                        |                  N/A |
   +-----------------------------------------+------------------------+----------------------+
   ```

## Resources

- [Install Ubuntu Server on Olares One](install-ubuntu-server.md)
- [Install Ubuntu Desktop on Olares One](install-ubuntu-desktop.md)
- [Manage BIOS and EC](update-firmware.md)
