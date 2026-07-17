---
outline: [2, 3]
title: Install Olares on PVE via ISO with GPU passthrough
description: Step-by-step tutorial on configuring GPU passthrough in Proxmox VE and installing Olares from ISO in a VM with GPU acceleration.
---

# Install Olares on PVE via ISO with GPU passthrough

GPU passthrough in Proxmox Virtual Environment (PVE) lets virtual machines (VMs) access the physical GPU directly, enabling hardware-accelerated computing for workloads like AI model inference and graphics processing.

:::warning Not for production use
Currently, Olares on PVE has certain limitations. Use it only for development or testing purposes.
:::

## Learning objectives

By the end of this tutorial, you will be able to:

- Enable GPU passthrough on a PVE host.
- Create a PVE VM with an NVIDIA GPU passed through.
- Install Olares from the official ISO and verify the GPU is recognized.

## Prerequisites

Make sure you have:

- **CPU**: At least 4 cores, with IOMMU enabled in BIOS
  - Intel: `VT-d`
  - AMD: `AMD-Vi`/`IOMMU`
- **GPU**: NVIDIA GPU that supports GPU passthrough
- **RAM**: Recommended 16 GB or more
- **Storage**: Minimum 200 GB SSD (installation may fail on HDD)
- **PVE version**: 8.3.2
- **Olares ISO image**: Download the [official Olares ISO image](https://cdn.olares.com/olares-latest-amd64.iso)

## Configure GPU passthrough in PVE

To use GPU-accelerated workloads in Olares, first enable GPU passthrough on the PVE host.

### Enable IOMMU

**Input-Output Memory Management Unit (IOMMU)** is a hardware feature that lets the operating system control how devices access memory. This control is required for passthrough.

1. In the PVE Shell, open the GRUB configuration file:

   ```bash
   nano /etc/default/grub
   ```

2. Find the line:

   ```plain
   GRUB_CMDLINE_LINUX_DEFAULT="quiet"
   ```

   Replace it with the line for your CPU vendor:

   ::: code-group
   ```bash [Intel]
   GRUB_CMDLINE_LINUX_DEFAULT="quiet intel_iommu=on iommu=pt"
   ```
   ```bash [AMD]
   GRUB_CMDLINE_LINUX_DEFAULT="quiet amd_iommu=on iommu=pt"
   ```
   :::

3. Save and close the file, then update GRUB and reboot the host:

   ```bash
   update-grub
   reboot
   ```

4. After the host restarts, check whether IOMMU is enabled:

   ```bash
   dmesg | grep -e DMAR -e IOMMU
   ```

   If successful, you should see output similar to:

   ::: code-group
   ```plain [Intel]
   [0.061644] DMAR: IOMMU enabled
   ...
   [0.408103] DMAR: Intel(R) Virtualization Technology for Directed I/O
   ```
   ```plain [AMD]
   [1.219719] AMD-Vi: Found IOMMU at 0000:00:00.2 cap 0x40
   ```
   :::

### Add VFIO modules

**Virtual Function I/O (VFIO)** enables a VM to directly access PCI devices such as GPUs.

1. On the PVE host, open the `modules` file:

   ```bash
   nano /etc/modules
   ```

2. Add these lines to the end of the file:

   ```plain
   vfio
   vfio_iommu_type1
   vfio_pci
   vfio_virqfd
   ```

3. Save and close the file.

### Blacklist host GPU drivers

Blacklist the host's default GPU drivers so the GPU is available for `vfio-pci`.

1. Create the blacklist configuration:

   ```bash
   nano /etc/modprobe.d/blacklist.conf
   ```

2. Add the following lines to block NVIDIA drivers:

   ```plain
   blacklist nouveau
   blacklist nvidia
   blacklist nvidiafb
   blacklist nvidia_drm
   blacklist nvidia_modeset
   ```

3. Save and close the file.

### Bind GPU to VFIO

1. On the PVE host, find your GPU's PCI address:

   ```bash
   lspci | grep NVIDIA
   ```

   Example output:

   ```plain
   01:00.0 VGA compatible controller: NVIDIA Corporation AD106 [GeForce RTX 4060 Ti] (rev a1)
   01:00.1 Audio device: NVIDIA Corporation AD106M High Definition Audio Controller (rev a1)
   ```

   In this example, the GPU's PCI address is `01:00`, with two functions listed.

2. Get the PCI identifiers of your GPU:

   ```bash
   lspci -n -s 01:00
   ```

   Example output:

   ```plain
   01:00.0 0300: 10de:2803 (rev a1)
   01:00.1 0403: 10de:22bd (rev a1)
   ```

   In this case, the GPU IDs are `10de:2803` and `10de:22bd`.

3. Bind the IDs to VFIO (replace the IDs with your own):

   ```bash
   echo "options vfio-pci ids=10de:2803,10de:22bd" > /etc/modprobe.d/vfio.conf
   ```

4. Update the `initramfs` and reboot:

   ```bash
   update-initramfs -u
   reboot
   ```

5. After the host restarts, check that the GPU is using the `vfio-pci` driver:

   ```bash
   lspci -v
   ```

   You should see output similar to:

   ```plain
   01:00.0 VGA compatible controller: NVIDIA Corporation AD106 [GeForce RTX 4060 Ti] (rev a1) (prog-if 00 [VGA controller])
   Subsystem: Gigabyte Technology Co., Ltd AD106 [GeForce RTX 4060 Ti]
   Flags: fast devsel, IRQ 255, IOMMU group 11
   ...
   Kernel driver in use: vfio-pci
   ```

## Set up VM and install Olares

With GPU passthrough enabled, you can now install Olares in PVE.

### Create and configure the VM

1. Upload the official Olares ISO to your PVE storage (for example, `local`):

   1. In the PVE web interface, select your target storage.
   2. Click **ISO Images** > **Upload**.
   3. Click **Select File**, choose the Olares ISO file, and click **Upload**.

2. Click **Create VM**.

3. Configure the VM as follows:

   - **OS**:
     - `ISO image`: Select the official Olares ISO image you downloaded.
   - **System**:
     - `BIOS`: Select OVMF (UEFI).
     - `EFI Storage`: Choose a storage location, for example a local LVM or directory, to store UEFI firmware variables.
     - `Pre-Enroll keys`: **Uncheck** to disable Secure Boot.
   - **Disks**:
     - `Disk size (GiB)`: At least 200 GB.
   - **CPU**:
     - `Cores`: At least 4 cores.
   - **Memory**:
     - `Memory (MiB)`: At least 8 GB.

   The following screenshot shows the sample hardware configuration.

   ![PVE VM hardware settings sample](/images/developer/install/pve-hardware.png#bordered)

4. Click **Finish**. **Do not** start the VM yet.

### Bind GPU to the VM

1. In the PVE interface, select your VM and go to **Hardware** > **Add** > **PCI Device**.

   ![Add PCI device to VM](/images/manual/tutorials/pve-add-pci.png#bordered)

2. Select **Raw Device**, then pick your GPU by the PCI address (for example, `01:00`).

3. In the bottom-right corner, select **Advanced** and check **PCI-Express**.

4. Click **Add**.

   ![Add GPU PCI device to VM](/images/manual/tutorials/pve-add-pci-gpu.png#bordered){width=70%}

Now your VM is ready to use GPU passthrough.

### Install Olares

Install Olares from the ISO image:

1. Select and start the VM you just created.

2. From the boot menu, select **Install Olares to Hard Disk** and press **Enter**.

3. In the Olares System Installer interface, select the installation disk:

   1. Review the list of available disks (for example, `sda 200G QEMU HARDDISK`).
   2. Select the first disk by typing `/dev/` plus its name (for example, `/dev/sda`).
   3. When the disk warning appears, type `yes` to continue.

   :::info Ignore NVIDIA driver warnings
   During installation, warnings related to the NVIDIA graphics driver will appear. Press **Enter** to dismiss them.
   :::

4. Once the installation completes, you'll see the message:

   ```plain
   Installation completed successfully!
   ```

5. In the Proxmox web interface, reboot the VM.

### Verify the installation and GPU passthrough

After the VM restarts, it boots into Ubuntu.

1. Log in to Ubuntu using the default credentials:

   - Username: `olares`
   - Password: `olares`

2. Run the following command to confirm that Olares installed successfully:

   ```bash
   sudo olares-check
   ```

   The installation is successful if you see output like:

   ```plain
   ...
   check Olaresd:  success
   check Containerd:  success
   ```

3. Check that the GPU is passed through and recognized by Olares using the NVIDIA System Management Interface tool:

   ```bash
   nvidia-smi
   ```

   If GPU passthrough is set up correctly, this command displays a table with your NVIDIA GPU details, including its name, driver version, and memory usage.

<!--@include: ../get-started/install-and-activate-olares.md-->

<!--@include: ../get-started/log-in-to-olares.md-->
