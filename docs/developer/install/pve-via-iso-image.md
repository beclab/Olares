---
Guide to installing Olares on Proxmox VE (PVE) using ISO image with system requirements, VM configuration, installation, and step-by-step activation instructions.
---
# Install Olares on PVE with ISO image
You can install Olares directly on Proxmox Virtual Environment (PVE) using an ISO image. This guide walks you through downloading the Olares ISO, configuring PVE environment, completing the installation, and getting your Olares up and running.

:::warning Not recommended for production use
Currently, Olares on PVE has certain limitations. We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md{45,51}-->


## System requirements
Make sure your device meets the following requirements.

- CPU: At least 4 cores
- RAM: At least 8GB of available memory
- Storage: At least 200GB of available SSD storage
- Supported Systems: PVE 8.2.2
## Download Olares ISO image
Download the official Olares ISO image.

## Configure VM in PVE

To run Olares on PVE, make sure the VM is configured with the following settings. You can either apply them when **creating a new VM** or adjust an **existing VM** to match these requirements.

### Required VM settings

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

Below is a sample configuration for the VM hardware settings in PVE. The screenshot illustrates how the fields should appear after applying the recommended values.

![PVE Hardware](/images/developer/install/pve-hardware.png#bordered)

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

    Press **Enter** and then use **CTRL + ALT + DEL** to reboot the VM.

## Verify installation

After the VM restarts, it will boot into the system. You should see the following prompt:

```yaml
Ubuntu login:
```

Log in using:

- Username: `olares`
- Password: `olares`

To confirm that Olares has been installed successfully, run the following command: `sudo olares-check`.

A successful installation will display results like:

```
...
check Olaresd:  success
check Containerd:  success
```

## Activate Olares

If you installed Olares via ISO on PVE or are using an Olares hardware device with ISO pre-installed:

![ISO Activate](/images/manual/larepass/iso-activate.png#bordered)

1. Open LarePass app.

2. Tap **Discover nearby Olares**. Your Olares device should appear.

3. Tap **Install now** to finish the installation process.

4. Tap **Activate now** to activate Olares and complete initialization.

5. Follow the prompt to set the login password for Olares.

   ![ISO Activate](/images/manual/larepass/iso-activate-2.png#bordered)

Once complete, you can access Olares via the provided URL and your credentials.

:::tip Same network required
Make sure your phone and the PVE host are on the same network.
:::

## Log in to Olares

1. Enter your URL (`https://desktop.{olares-id}.olares.com`) in your browser, and press any key to continue.
2. On the login page, enter your Olares login password.

   ![Log in](/images/manual/get-started/log-in.png#bordered)
3. You will be prompted to complete two-factor verification. You can confirm the login on LarePass, or manually enter the 6-digit verification code.

   ![Confirm login](/images/manual/larepass/confirm-login.png#bordered)
   ::: info
   The verification code is time-sensitive. Ensure you enter it before it expires. If it does, you will need to generate a new code.
   :::

Once you've logged in, you'll be directed to the Olares desktop.🎉

<!--@include: ./reusables.md{39,43}-->