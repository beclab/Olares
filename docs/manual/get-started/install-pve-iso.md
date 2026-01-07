---
description: Guide to installing Olares on Proxmox VE (PVE) using ISO image with system requirements, VM configuration, installation, and step-by-step activation instructions.
---
# Install Olares on PVE with ISO image
You can install Olares directly on Proxmox Virtual Environment (PVE) using an ISO image. This guide walks you through downloading the Olares ISO, configuring PVE environment, completing the installation, and getting your Olares up and running.

:::warning Not recommended for production use
Currently, Olares on PVE has certain limitations. We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md{44,51}-->

## System requirements
Make sure your device meets the following requirements.
### Required specifications
- CPU: Minimum 4 cores with x86-64 architecture (Intel or AMD). ARM-based processors are not currently supported for this method.
- RAM: At least 8 GB of available memory
- Storage: At least 200 GB of available SSD storage. 
    ::: warning SSD required
    The installation will likely fail if an HDD (mechanical hard drive) is used instead of an SSD.
    :::
- Supported Systems: PVE 8.2.2

### Optional hardware

A GPU is not required to install Olares, but is necessary for AI applications.

- GPU (NVIDIA only):
   - Architecture: Turing or newer (e.g., GTX 16 series, RTX 20 series, and later).
   - Verification: Run `lspci | grep -i nvidia` in the PVE host shell to confirm the card is detected.
   - Setup: To utilize the GPU, you must configure PCI passthrough. Please refer to [Configure GPU passthrough in PVE](/manual/best-practices/install-olares-gpu-passthrough.md#configure-gpu-passthrough-in-pve) for detailed instructions.

## Download Olares ISO image
Click [here](https://cdn.olares.com/olares-latest-amd64.iso) to download the official Olares ISO image.

## Configure VM in PVE

To run Olares on PVE, make sure the VM is configured with the following settings. You can either apply them when **creating a new VM** or adjust an **existing VM** to match these requirements.

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

Below is a sample configuration for the VM hardware settings in PVE. 

![PVE Hardware](/images/developer/install/pve-hardware.png#bordered)

:::info Version compatibility
While the specific version is confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::

## Install on PVE

Once the VM is set up, follow these steps to install the ISO on PVE.

1. Start the virtual machine (VM).
2. From the boot menu, select **Install Olares to Hard Disk**.
3. In the Olares System Installer, a list of available disks will appear (for example, `sda 200G QEMU HARDDISK`). Select the first disk by typing `/dev/` plus its name (for example, `/dev/sda`). When the on-screen warning appears, just type `yes` to continue.

   :::tip Note
   During installation, warnings related to the NVIDIA graphics driver may appear. If they do, press **Enter** to ignore them.
   :::

4. Once the installation completes, you’ll see the message:

    ```
    Installation completed successfully!
    ```
    
    Click **Reboot** in the Proxmox web interface to restart the VM.


## Verify installation

After the VM restarts, it will boot into the Ubuntu system. 
1. Log in to the Ubuntu using the default credentials:
     - Username: `olares`
     - Password: `olares`

2. Confirm that Olares has been installed successfully using the following command: 
     ```bash
     sudo olares-check
     ```

    The installation is successful if you see results like:

    ```bash
    ...
    check Olaresd:  success
    check Containerd:  success
    ```

<!--@include: ./install-and-activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{38,43}-->