---
Guide to installing Olares on Proxmox VE (PVE) using ISO image with system requirements, installation commands, and step-by-step activation instructions.
---
# Install Olares on PVE with ISO image
You can use an ISO image to install Olares directly on Proxmox Virtual Environment (PVE). This guide walks you through downloading the Olares ISO, configuring the necessary parameters in PVE, completing the installation, and getting your system up and running.

:::warning Not recommended for production use
Currently, Olares on PVE has certain limitations. We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md{45,51}-->

## Download Olares ISO image
Download the official Olares ISO image at our website.

## System requirements
Make sure your virtual machine meets the following requirements.

- CPU: At least 4 cores
- RAM: At least 8GB of available memory
- Storage: At least 200GB of available SSD storage
- Supported Systems: PVE 8.2.2
- BIOS: OVMF (UEFI)
- EFI Disk: Choose a storage location (for example, a local LVM or directory) to store UEFI firmware variables.
- Pre-Enroll keys: **Uncheck** the option to disable Secure Boot.

::: warning SSD required
The installation will likely fail if an HDD (mechanical hard drive) is used instead of an SSD.
:::

:::info Version compatibility
While the specific version is confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::

## Install on PVE

1. Start the virtual machine (VM).
2. From the boot menu, select **Install Olares to Hard Disk**.
3. In the Olares System Installer, available disks will be listed. Enter the drive letter of the first disk as your target disk.

    - A warning will appear:

    ```text
    WARNING: This will DESTROY all data on <your target disk>
    ```

    Type `yes` when prompted with `Continue? (yes/no):` to proceed.

4. The installation will begin. You will see log messages and the graphics driver installation process. Some warnings may appear as the following example: 

    ```text
    WARNING:
    nvidia-installer was forced to guess the X Iibrary path 'usr/lib'and X module path
    /usr/lib/xorg/modules'; these paths were not queryable from the system. If X fails to
    find the NVIDIA X driver module, please install the `pkg-config` utility and the X.Org
    SDK/development package for your distribution and reinstall the driver.
    ```

    These can be safely ignored. Press **Enter** to confirm OK when prompted.

5. Once the installation completes, you’ll see the message:

    ```
    Installation completed successfully!
    ```

    Press **Enter** or use **CTRL + ALT + DEL** to reboot the VM.

## Verify installation

After the VM restarts, it will boot into the system. You should see the following prompt:

```yaml
Ubuntu login:
```

Log in using:

- Username `olares`
- Password: `olares`

To confirm that Olares has been installed successfully, run the following command: `sudo olares-check`.

A successful installation will display results like:

```
...
check Olaresd:  success
check Containerd:  success
```

This indicates that all essential Olares services are running properly.


<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{39,43}-->