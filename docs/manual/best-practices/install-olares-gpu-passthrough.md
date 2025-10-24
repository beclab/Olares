---
outline: [2, 3]
description: Step-by-step tutorial on how to set up GPU passthrough in Proxmox VE (PVE) and install Olares in a virtual machine with GPU acceleration enabled.
---

# Install Olares on PVE via ISO with GPU Passthrough

GPU passthrough in **Proxmox Virtual Environment (PVE)** allows virtual machines (VMs) to directly access the physical GPU, enabling hardware-accelerated computing for workloads like AI model inference and graphics processing.

This tutorial provides a comprehensive, end-to-end process for configuring a PVE host for GPU passthrough and then installing Olares from its official ISO image into a new VM that fully leverage the dedicated GPU.

::: warning Not recommended for production use
Currently, Olares on PVE has certain limitations. We recommend using it only for development or testing purposes.
:::

## Prerequisites

Before proceeding, ensure that your setup meets the following requirements:

- CPU: At least 4 cores, with IOMMU enabled in BIOS
  - Intel: `VT-d`
  - AMD: `AMD-Vi`/`IOMMU`
- GPU: NVIDIA GPU that supports GPU passthrough
- RAM: Recommended 16 GB or more 
- Storage: Minimum 200 GB SSD (installation may fail on HDD)  
- PVE Version: 8.3.2
- Olares ISO Image: Download the [official Olares ISO image](https://dc3p1870nn3cj.cloudfront.net/olares-v1.12.1-amd64.iso) before you start.

## Configure GPU Passthrough in PVE

To use GPU-accelerated workloads in Olares, you must first enable GPU passthrough for the PVE host.

### Enable IOMMU

The **Input-Output Memory Management Unit (IOMMU)** is a hardware feature that allows the operating system to control how devices access memory, which is essential for passthrough.

1. In the PVE command line (Shell), run the following command to open the GRUB configuration file:
        
    ```bash
    nano /etc/default/grub
    ```
    
2. Find the line: `GRUB_CMDLINE_LINUX_DEFAULT="quiet"`
    
    Replace it with the line corresponding to your CPU vendor:
   
    ::: code-group
    ```bash [Intel]
    GRUB_CMDLINE_LINUX_DEFAULT="quiet intel_iommu=on iommu=pt"
    ```
    ```bash [AMD]
    GRUB_CMDLINE_LINUX_DEFAULT="quiet amd_iommu=on iommu=pt"
    ```
    :::

3. Save and close the GRUB configuration file, then apply GRUB changes:
  
    ```bash
    update-grub
    reboot
    ```

4.  After restarting, check whether IOMMU is enabled (still in the PVE host):
    
    ```bash
    dmesg | grep -e DMAR -e IOMMU
    ```
       If successful, you should see output similar to the following:

    ::: code-group
    ```bash [Intel]
    [0.061644] DMAR: IOMMU enabled
    ...
    [0.408103] DMAR: Intel(R) Virtualization Technology for Directed I/O
    ```
    ```bash [AMD]
    [1.219719] AMD-Vi: Found IOMMU at 0000:00:00.2 cap 0x40
    ```
    :::

### Add VFIO modules

**VFIO (Virtual Function I/O)** enables a VM to directly access PCI devices such as GPUs.

1. On the PVE host, run the following command to open the `modules` file:

    ```bash
    nano /etc/modules
    ```

2. Add these lines to the end of the file:
    
    ```
    vfio
    vfio_iommu_type1
    vfio_pci
    vfio_virqfd
    ```

3. Save and close the file.

### Blacklist Host GPU Drivers

To prevent the Proxmox host from using the GPU you plan to pass through, it's best to blacklist its default drivers. This ensures the GPU is available for `vfio-pci`.

1. Run the following command on the PVE host to create the blacklist file:

    ```bash
    nano /etc/modprobe.d/blacklist.conf
    ```

2. Add the following lines to block NVIDIA drivers:

    ```
    blacklist nouveau 
    blacklist nvidia 
    blacklist nvidiafb
    blacklist nvidia_drm
    blacklist nvidia_modeset
    ```

3. Save and close the file.

### Bind GPU to VFIO

1. Run the following command on the PVE host to find your GPU's PCI address:

    ```bash
    lspci | grep NVIDIA
    ```

    **Example output**:

    ```
    01:00.0 VGA compatible controller: NVIDIA Corporation AD106 [GeForce RTX 4060 Ti] (rev a1)
    01:00.1 Audio device: NVIDIA Corporation AD106M High Definition Audio Controller (rev a1)
    ```

    Here, the GPU's PCI address is `01:00`, with two functions listed.

2. Get the IDs of your GPU (`01:00` in this example):
   
    ```bash
    lspci -n -s 01:00
    ```
    **Example output**:

    ```
    01:00.0 0300: 10de:2803 (rev a1)
    01:00.1 0403: 10de:22bd (rev a1)
    ```

    In this case, the IDs of the GPU are `10de:2803` and `10de:22bd`.

3. Bind the IDs to VFIO (replace the IDs with your own):

    ```bash
    echo "options vfio-pci ids=10de:2803,10de:22bd" > /etc/modprobe.d/vfio.conf
    ```

4. Apply all module and driver changes by updating the initramfs, then reboot:

    ```bash
    update-initramfs -u
    reboot
    ```

5. After rebooting, check if the GPU is now using the `vfio-pci` driver:

    ```bash
    lspci -v
    ```

    You should see the output similar to:

    ```
    01:00.0 VGA compatible controller: NVIDIA Corporation AD106 [GeForce RTX 4060 Ti] (rev a1) (prog-if 00 [VGA controller])    
    Subsystem: Gigabyte Technology Co., Ltd AD106 [GeForce RTX 4060 Ti]
    Flags: fast devsel, IRQ 255, IOMMU group 11
    ...
    Kernel driver in use: vfio-pci
    ```


## Set up VM and install Olares

### Create and configure the VM

This section creates and configures a VM using the Olares ISO image:

1. Upload the official Olares ISO you downloaded to your PVE storage (e.g., `local`). You can do this by selecting the storage, clicking **ISO Images**, and then *Upload**. 

2. Click **Create VM**.

3. Configure the settings as follows:

    - OS:
        - `ISO image`: Select the official Olares ISO image you just downloaded.
    - System:
        - `BIOS`: Select OVMF (UEFI).
        - `EFI Storage`: Choose a storage location (for example, a local LVM or directory) to store UEFI firmware variables.
        - `Pre-Enroll keys`: **Uncheck** the option to disable Secure Boot.
    - Disks:
        - `Disk size (GiB)`: At least 200GB.
    - CPU:
        - `Cores`: At least 4 cores
    - Memory:
        - `Memory (MiB)`: At least 8GB

4. Click **Finish**. **Do not** start the VM yet.

    Below is a sample configuration for the VM hardware settings in PVE. 
    ![PVE Hardware](/images/developer/install/pve-hardware.png#bordered)

### Bind GPU to the VM

1. In the PVE interface, select your VM and go to **Hardware** > **Add** > **PCI Device**.
![Add PCI](/images/manual/tutorials/pve-add-pci.png#bordered)

2. Select **Raw Device**, then pick your GPU by the PCI address (for example, `01:00`).

3. In the bottom-right corner, select the **Advanced** options and check **PCI-Express**.

4. Click **Add** to save.
![Add GPU](/images/manual/tutorials/pve-add-pci-gpu.png#bordered)

Now your VM is ready to use GPU passthrough.

### Install Olares

Once the VM is set up, follow these steps to install the ISO on PVE.

1. Select and start the VM you just created.

2. From the boot menu, select **Install Olares to Hard Disk** and press **Enter**.

3. In the Olares System Installer, a list of available disks will display (for example, `sda 200G QEMU HARDDISK`). Select the first disk by typing `/dev/` plus its name (for example, `/dev/sda`). When the on-screen warning appears, just type `yes` to continue.

    ::: tip Note
    During installation, warnings related to the NVIDIA graphics driver may appear. If they do, press **Enter** to ignore them.
    :::

4. Once the installation completes, you'll see the message:

    ```
    Installation completed successfully!
    ```
    
5.  **Reboot the VM** in the Proxmox web interface.

### Verify the installation and GPU passthrough

After the VM restarts, it will boot into the Ubuntu system.

1. Log in to Ubuntu using the default credentials:
  - Username: `olares`
  - Password: `olares`

2. Confirm that Olares has been installed successfully using the following command:

    ```bash
    sudo olares-check
    ```

    The installation is successful if you see results like:

    ```      
    ...
    check Olaresd:  success
    check Containerd:  success
    ```

3. Finally, verify that the GPU is successfully passed through and recognized by Olares using the NVIDIA System Management Interface tool:

    ```bash
    nvidia-smi
    ```

    If successful, this command will display a table with your NVIDIA GPU's details, including its name, driver version, and memory usage.

## Next steps

Olares is now installed and running with full GPU acceleration. 
To start using Olares, activate the device and log in with your account.
For detailed steps, see our official guides:
- [Finish installation and activate Olares](../get-started/install-pve-iso.md#finish-installation-and-activate-olares)
- [Log in to Olares](../get-started/install-pve-iso.md#log-in-to-olares)