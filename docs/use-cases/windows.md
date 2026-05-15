---
outline: [2, 4]
description: A comprehensive guide to installing and running a Windows virtual machine on Olares. Learn how to configure initial credentials, connect via browser-based VNC or Microsoft Remote Desktop (RDP), and transfer files between your computer and the VM.
---

# Run a Windows VM on your Olares device

Olares lets you run a full Windows virtual machine directly on your device, giving you a personal, always-available Windows environment accessible from macOS, Windows, or Linux.

:::info System capabilities
- Olares supports running essential Windows applications. 
- By default, the Windows VM uses CPU-based virtualization and virtual display output.
- Intel integrated GPU support is available only on supported hardware and requires additional host configuration. See [Advanced: Enable Intel integrated GPU for Windows VM](#advanced-enable-intel-integrated-gpu-for-windows-vm).
- Audio output is **only supported** when connected via Remote Desktop (RDP).
:::

This guide walks you through installing the Windows VM, enabling secure networking, and connecting using Remote Desktop for the best experience.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Install and set up the Windows VM on your Olares device.
- Access the Windows VM using the browser-based VNC viewer or Microsoft Remote Desktop (RDP).
- Change your Windows login password from inside the VM.
- Transfer files seamlessly between your computer and the Windows VM.

## Install and configure Windows VM

Windows is available as an app in the Olares Market.

### Install Windows
1. Open the Market, and search for "Windows".
2. Click **Get**, then click **Install**.
   ![Install Windows](/images/manual/use-cases/win-install1.png#bordered)

3. When prompted, set environment variables:
    - **USERNAME:** Create a username for accessing Windows.
    - **PASSWORD:** Set the corresponding password.
    - **VERSION:** Select your preferred Windows version from the dropdown list.
    - **DISK_SIZE:** Allocate disk space for Windows.

    ![Set environment variables](/images/manual/use-cases/win-set-env-var1.png#bordered){width=70%}

4. Wait a few minutes for the installation and initialization to complete.

### Set up Windows

Once the installation is finished, open Windows from Launchpad to start the VM for the first time.

Olares will automatically download and install the system image of the corresponding Windows version. This may take several minutes depending on your network speed.

![Download Windows 11](/images/manual/use-cases/win-downloading-win11.png#bordered)
## Access the Windows VM

You can access your VM in two ways: 
- [**Browser:**](#method-1-access-from-the-browser-vnc) for setup and quick tasks
- [**Remote Desktop:**](#method-2-access-using-a-remote-desktop-client-rdp) for the best daily experience

### Method 1: Access from the browser (VNC)

Open the Windows app from Launchpad to launch the VM directly in your browser using VNC.
::: info
VNC (Virtual Network Computing) provides immediate, clientless access without requiring any additional software. It is ideal for initial setup, troubleshooting, or emergency access when you cannot use RDP. However, it can feel less responsive and lacks advanced features like audio redirection and high-performance graphics.
:::
### Method 2: Access using a Remote Desktop Client (RDP)
RDP (Remote Desktop Protocol) provides a much smoother, native-like experience with better performance, audio support, and seamless file transfer.

#### Locate port number for Windows
:::warning Multiple Windows instances
Each Windows instance uses a unique port. If you have cloned the Windows app, ensure you check the **ACLs** section for the specific instance you want to access.
:::
1. Open Settings, and navigate to **Application** > **Windows**.
2. Under **Permissions**, click **ACLs**.
3. Note the port number listed in the **Port** column. You will need this for the connection step.
   ![Locate port number](/images/manual/use-cases/win-port-number.png#bordered){width=90%}

#### Connect to Windows via RDP
:::info
The following steps show the macOS interface, but the workflow is similar on all platforms.
:::
1. [Enable VPN on LarePass](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device.

    When the VPN connection status shows **P2P**, or **Intranet**, the secure network is active and ready for remote access.

2. Install the Remote Desktop client.
   - **Windows:** No installation needed.
   - **macOS / iOS:** Download [Windows App from the App Store](https://apps.apple.com/us/app/windows-app/id1295203466).
   - **Android:** Download [Windows App from Google Play](https://play.google.com/store/apps/details?id=com.microsoft.rdc.androidx).

3. Open Windows from the Launchpad in your browser. Copy the domain from the address bar (exclude `https://` and any text after the domain).
   ![Domain address](/images/manual/use-cases/win-url.png#bordered)

4. Add your Windows VM as an RDP connection. 

    a. Open the Windows App on your device.

    b. Click the **＋** icon and select **Add PC**.

    c. In **PC name**, enter the domain you get from the previous step, followed by a colon and the port number.

      For example, if your URL is `https://0f4137ed.<username>.olares.com`, and the port is `47374`, enter:
      ```
      0f4137ed.<username>.olares.com:47374
      ```

   ![Add PC](/images/manual/use-cases/win-add-pc1.png#bordered)

    d. Click **Add**.

5. Connect to the Windows VM.

   a. Double-click your saved PC entry, or click **⋯** and choose **Connect**.
   ![Connect to PC](/images/manual/use-cases/win-connect-device1.png#bordered)
        
   b. When prompted, enter the **Username** and **Password** you created earlier.
   ![Log in to PC](/images/manual/use-cases/win-log-in1.png#bordered)

   c. If a security warning appears, click **Continue**.
   ![Continue to log in](/images/manual/use-cases/win-confirm-connect1.png#bordered)

You are now connected to your Windows VM via RDP.
![Windows VM](/images/manual/use-cases/win-vm-interface.png#bordered)

## Optional: Change your Windows login password

You can update your Windows login password directly from inside the VM:
1. Click the search bar in the Windows taskbar and type "password".  
2. Select **Change your password**.  
    ![Change your password](/images/manual/use-cases/win-change-pw.png#bordered)
3. Click **Change** to set your new password.
    ![Set new password](/images/manual/use-cases/win-set-pw.png#bordered)

## Transfer files between your computer and Windows

RDP supports clipboard-based file transfers.

You can: 
- Copy any file on your Mac or PC.
- Paste it directly into the Windows VM.

The file appears immediately in Windows and is ready to use.

## Disconnect from the Windows VM

To end your RDP session, simply close the RDP window.

The Windows VM continues running on your Olares device and is always ready for you to reconnect.

## Advanced: Enable Intel integrated GPU for Windows VM

You can pass an Intel integrated GPU (iGPU) from the Olares host to the Windows VM on supported devices. After configuration, Windows can detect the Intel graphics device through RDP.

### Supported devices

Intel integrated GPU support is available only when the Olares host meets one of the following conditions:

- **Olares One:** You are using the official Olares One device.
- **Self-hosted devices:** You are using a host with an Intel CPU that includes Intel integrated graphics, running Ubuntu 24.04 or later, and Intel VT-d / IGD is supported and enabled in BIOS.

If the host does not have an Intel iGPU, the Windows VM cannot create one from software alone.

### Configure the host environment

Choose the setup path that matches your Olares device.

#### For Olares One

No BIOS or SR-IOV setup is required for Olares One.

The Windows app deployed from the Olares Market already includes the required GPU-related environment variables by default. You can skip directly to [Install the Intel graphics driver in Windows](#install-the-intel-graphics-driver-in-windows).

#### For self-hosted devices

Use this workflow to configure the BIOS and operating system on a self-hosted Olares device.

Make sure you can access the Olares host terminal through SSH before continuing. This setup changes BIOS settings, kernel parameters, DKMS modules, SR-IOV virtual functions, and VFIO device binding. Incorrect configuration may affect host startup or device availability.

1. Configure the BIOS.

   Restart the host and enter the BIOS/UEFI settings. The exact menu names vary by motherboard vendor. Look for these options under menus such as **Advanced**, **Chipset**, **System Agent**, **PCIe**, or **Graphics**:

   - Enable **Intel VT-d**.
   - Enable the integrated graphics device, if it is disabled.
   - Set the primary display or initial display output to **IGD**, **Integrated Graphics**, or a similar option.
   - Save the settings and restart the host.

   If you cannot find an integrated graphics or IGD option, confirm that your Intel CPU includes integrated graphics, then check your motherboard manual for the exact BIOS option name.

2. Verify the iGPU.

   Access the Olares host terminal through SSH and run:

   ```bash
   lspci -nn | grep VGA
   ```

   Confirm that the output includes an Intel integrated GPU and that the device type is `VGA compatible controller`.

   Example output:

   ```plain
   00:02.0 VGA compatible controller [0300]: Intel Corporation AlderLake-S GT1 [8086:4680] (rev 0c)
   ```

3. Enable IOMMU and Intel vGPU.

   This step installs the Intel SR-IOV DKMS module, writes the required IOMMU and iGPU parameters to GRUB, and creates the vGPU virtual functions.

   **Install the DKMS module**

   Run the following commands to update the host, install dependencies, and build the Intel SR-IOV DKMS module:

   ```bash
   sudo apt update && sudo apt upgrade -y

   sudo apt update && sudo apt install git sysfsutils dkms -y

   sudo rm -rf /var/lib/dkms/i915-sriov-dkms*
   sudo rm -rf /usr/src/i915-sriov-dkms*
   rm -rf ~/i915-sriov-dkms

   cd ~
   git clone https://github.com/strongtz/i915-sriov-dkms.git
   cd ~/i915-sriov-dkms

   cp -a dkms.conf{,.bak}
   KERNEL=$(uname -r)
   sed -i 's/"@_PKGBASE@"/"i915-sriov-dkms"/g' dkms.conf
   sed -i "s/PACKAGE_VERSION=\".*\"/PACKAGE_VERSION=\"$KERNEL\"/g" dkms.conf
   sed -i 's/ -j$(nproc)//g' dkms.conf
   cat dkms.conf

   sudo apt install --reinstall dkms -y
   sudo dkms add .
   cd /usr/src/i915-sriov-dkms-$KERNEL
   sudo dkms status
   sudo dkms install -m i915-sriov-dkms -v $KERNEL -k $(uname -r) --force -j 1
   sudo dkms status
   ```

   **Update GRUB**

   Add or modify the following line in `/etc/default/grub`:

   ```text
   GRUB_CMDLINE_LINUX_DEFAULT="quiet intel_iommu=on iommu=pt i915.enable_guc=3 i915.max_vfs=7"
   ```

   If the line already contains other parameters, keep them and add the new parameters inside the quotation marks.

   In this example, `i915.max_vfs=7` configures up to 7 Intel vGPU virtual functions.

   Update the GRUB configuration and restart the host:

   ```bash
   sudo update-grub && sudo reboot
   ```

   **Add the SR-IOV configuration**

   After the host restarts, add the SR-IOV configuration to `/etc/sysfs.conf`.

   The command below uses `00:02.0`, which should match the Intel iGPU PCI address shown in the earlier `lspci` output. If your Intel iGPU uses a different PCI address, update the path accordingly.

   ```bash
   sudo sh -c 'echo "devices/pci0000:00/0000:00:02.0/sriov_numvfs = 7" > /etc/sysfs.conf'

   cat /etc/sysfs.conf

   sudo reboot
   ```

   **Verify the vGPU devices**

   After the host restarts, verify that the vGPU devices were created:

   ```bash
   lspci -nn | grep VGA
   ```

   You should see additional virtual VGA devices.

4. Bind one vGPU to VFIO.

   Bind one vGPU, such as `0000:00:02.1`, to the VFIO driver:

   ```bash
   sudo apt install -y driverctl

   sudo driverctl set-override 0000:00:02.1 vfio-pci
   ```

   In this example, `0000:00:02.1` is the vGPU device that will be assigned to the Windows VM. The override persists across host reboots.

### Configure the Windows app

#### For Olares One

The Windows app deployed from the Olares Market already includes the required GPU environment variables by default. 

Skip directly to [Install the Intel graphics driver in Windows](#install-the-intel-graphics-driver-in-windows).

#### For self-hosted devices

For self-hosted devices, add the GPU-related environment variables to the Windows Deployment. Make sure the `host=` value matches the vGPU device you bound to `vfio-pci` earlier. The example below uses `0000:00:02.1`.

1. Open Control Hub from the Launchpad. 
2. Under **Browse**, select the windows project from the list.
3. Under **Deployments**, select `windows`.
4. In the upper-right corner of the details pane, click <i class="material-symbols-outlined">edit_square</i> to edit the YAML.
5. Under the `env` section, add or update the following variables:

   ```yaml
   env:
     - name: GPU
       value: 'Y'
     - name: VGA
       value: 'vfio-pci,host=0000:00:02.1,multifunction=on,x-vga=on -vga virtio'
   ```
   ![Edit yaml](/images/manual/use-cases/windows-edit-yaml.png#bordered)
6. Click **Confirm** to save the changes, and restart the Windows app.

### Install the Intel graphics driver in Windows

1. Connect to the Windows VM using RDP.

2. Open a browser in Windows.

3. Download the latest Intel graphics driver from Intel:

   ```plain
   https://www.intel.cn/content/www/cn/zh/download/785597/intel-arc-graphics-windows.html
   ```

4. Install the driver.

5. Restart Windows if prompted.

### Verify the result

1. Reconnect to the Windows VM using RDP.
2. Open **Device Manager**.
3. Expand **Display adapters**.
4. Confirm that the Intel graphics device appears and is running without errors.

## FAQ

### The Windows VM shows a blank screen or no desktop

The browser may have suspended the VNC connection due to inactivity to conserve system resources.  
    ![Reconnect VM](/images/manual/use-cases/win-vnc-reconnect.png#bordered)

Click **Connect** to restore the session.

### Windows system image download fails

If the Windows system image fails to download during setup:

- Wait a short while, then restart the application:
    1. Open Control Hub from the Launchpad.
    2. Select the windows project.
    3. Under **Deployment**, click windows.
    4. Click **Restart**.
    ![Restart VM](/images/manual/use-cases/win-restart.png#bordered)

  After the restart, the system image download will automatically retry.
- If repeated failures occur, your IP may have been temporarily blocked by Microsoft due to multiple download attempts in a short period.  
  Wait **24 hours**, then restart or reinstall the application and try again.
- If the issue persists, please contact us for assistance.

### Can I install other Windows versions or languages?

Currently, the following Windows version are supported:
- Windows 11 Pro
- Windows 11 LTSC
- Windows 11 Enterprise
- Windows 10 Pro
- Windows 10 LTSC
- Windows 10 Enterprise
- Windows 8.1 Enterprise
- Windows 7 Ultimate
- Windows Vista Ultimate
- Windows 2000 Professional
- Windows Server 2025
- Windows Server 2022
- Windows Server 2019
- Windows Server 2016
- Windows Server 2012
- Windows Server 2008
- Windows Server 2003

After Windows installation, you can change the display language using the standard Windows language settings.

## Learn more

- [Run a macOS VM on your Olares device](./macos.md)