---
outline: [2, 4]
description: A comprehensive guide to installing and running a Windows virtual machine on Olares. Learn how to configure initial credentials, connect via browser-based VNC or Microsoft Remote Desktop (RDP), and transfer files between your computer and the VM.
head:
  - - meta
    - name: keywords
      content: Olares One, Windows VM
---

# Run a Windows VM on Olares One <Badge type="tip" text="20 min" />

Olares lets you run a full Windows virtual machine directly on your device, giving you a personal, always-available Windows environment accessible from macOS, Windows, or Linux.

:::info System capabilities
- Olares supports running essential Windows applications. 
- Workflows are limited to **CPU or integrated graphics performance**. GPU passthrough is not yet supported, meaning heavy GPU-accelerated applications may not perform optimally.
- Audio output is **only supported** when connected via Remote Desktop (RDP).
:::

This guide walks you through installing the Windows VM, enabling secure networking, and connecting using either a browser or Remote Desktop.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Install and set up the Windows VM on Olares One.
- Access the Windows VM using the browser-based VNC viewer or Microsoft Remote Desktop (RDP).
- Change your Windows login password from inside the VM.
- Transfer files seamlessly between your computer and the Windows VM.

## Install and configure Windows VM

Windows is available as an app in the Olares Market.

### Install Windows
1. Open Market, and search for "Windows".
   ![Install Windows](/images/one/win-market.png#bordered)

2. Click **Get**, then **Install**.
3. When prompted, set environment variables:
    - `USERNAME`: Create a username for accessing Windows.
    - `PASSWORD`: Set the corresponding password.
    - `VERSION`: Select your preferred Windows version from the dropdown list.
    - `DISK_SIZE`: Allocate disk space for Windows using units of `G`, for example, `128G`.

    ![Set environment variables](/images/manual/use-cases/win-set-env-var1.png#bordered){width=70%}

4. Wait a few minutes for the installation and initialization to complete.

### Set up Windows

Once the installation is finished, open Windows from Launchpad to start the VM for the first time.

Olares will automatically download and install the system image of the corresponding Windows version. This may take several minutes depending on your network speed.

![Download Windows 11](/images/one/win-download.png#bordered){width=90%}
## Access the Windows VM

You can access your VM in two ways: 
- [Browser](#method-1-access-from-the-browser-vnc): For initial setup or quick tasks.
- [Remote Desktop](#method-2-access-using-a-remote-desktop-client-rdp): Recommended for daily use with audio support, file transfer capabilities, and smoother interaction. 

### Method 1: Access from the browser (VNC)

Open the Windows app from Launchpad to launch the VM directly in your browser using VNC.
:::info
VNC (Virtual Network Computing) provides immediate, clientless access without requiring any additional software. It is ideal for initial setup, troubleshooting, or emergency access when you cannot use RDP. However, it can feel less responsive and lacks advanced features like audio redirection and high-performance graphics.
:::
### Method 2: Access using a Remote Desktop Client (RDP)

RDP (Remote Desktop Protocol) provides a much smoother, native-like experience with better performance, audio support, and file transfer.

#### Prerequisites

Before connecting, ensure you have the following ready:
- Secure tunnel: [Enable VPN on LarePass](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device.
- RDP client: Download Microsoft's remote desktop client on the device you are connecting from.
    - **Windows:** No installation needed.
    - **macOS / iOS:** Download [Windows App from the App Store](https://apps.apple.com/us/app/windows-app/id1295203466).
    - **Android:** Download [Windows App from Google Play](https://play.google.com/store/apps/details?id=com.microsoft.rdc.androidx).

#### Locate port number for Windows

:::warning Multiple Windows instances
If you have more than one Windows instance, each instance uses a different port.  
Make sure you check the **ACLs** for the specific instance you want to connect to.
:::

1. Open Settings, and navigate to **Application** > **Windows**.
2. Under **Permissions**, click **ACLs**.
3. Note the port number listed in the **Dst** column. You will need this for the connection step.
   ![Locate port number](/images/one/win-port-no.png#bordered)

#### Connect to Windows

:::info
The following steps show the macOS interface, but the workflow is similar on all platforms.
::: 

1. Open Windows from the Launchpad. Copy the domain from the address bar (exclude `https://` and any text after the domain).
2. Open the Windows App (or Remote Desktop) on your device.
3. Add your Windows VM as an RDP connection. 

    a. Click <i class="material-symbols-outlined">add</i> in the top-right corner and select **Add PC**.

    b. In **PC name**, enter the domain you get from the previous step, followed by a colon and the port number.

      For example, if your URL is `https://0f4137ed.<username>.olares.com`, and the port is `47374`, enter:
      ```text
      0f4137ed.<username>.olares.com:47374
      ```

   ![Add PC](/images/one/win-add-pc.png#bordered)

    c. Click **Add**.

4. Connect to the Windows VM.

   a. Double-click your saved PC entry, or click **⋯**, then choose **Connect**.
   ![Connect to PC](/images/one/win-connect.png#bordered){width=90%}
        
   b. When prompted, enter the **Username** and **Password** you created earlier.
   ![Log in to PC](/images/one/win-login.png#bordered){width=90%}

   c. If a security warning appears, click **Continue**.
   ![Continue to log in](/images/one/win-warning.png#bordered){width=90%}

You are now connected to your Windows VM via RDP.
![Windows VM](/images/one/win-desktop.png#bordered)

## Common tasks

### Change your Windows login password

You can update your Windows login password directly from inside the VM:
1. Click the search bar in the Windows taskbar and type "password".  
2. Select **Change your password**.  
3. Click **Change** to set your new password.

### Transfer files between your computer and Windows

RDP supports clipboard-based file transfers.

To transfer a file via RDP clipboard:
1. Copy a file on your Mac or PC.
2. Paste it into the Windows VM.

The file appears immediately in Windows and is ready to use.

### Disconnect from the Windows VM

To end your RDP session, simply close the RDP window.

The Windows VM continues running on your Olares device and is always ready for you to reconnect.

## FAQ

### The Windows VM shows a blank screen or no desktop

The browser may have suspended the VNC connection due to inactivity to conserve system resources.  
    ![Reconnect VM](/images/manual/use-cases/win-vnc-reconnect.png#bordered){width=90%}

Click **Connect** to restore the session.

### Windows system image download fails

If the Windows system image fails to download during setup:

- Wait a short while, then restart the application:
    1. Open Control Hub from the Launchpad.
    2. Select the Windows project.
    3. Under **Deployment**, click **windows**.
    4. Click **Restart**.
    ![Restart VM](/images/one/win-restart.png#bordered)

  After the restart, the system image download will automatically retry.

- If repeated failures occur, your IP may have been temporarily blocked by Microsoft due to multiple download attempts in a short period.  
  Wait **24 hours**, then restart or reinstall the application and try again.
- If the issue persists, please contact us for assistance.

### Can I install other Windows versions or languages?

Currently, the following Windows versions are supported:
- **Windows 11**: Pro, LTSC, Enterprise
- **Windows 10**: Pro, LTSC, Enterprise
- **Legacy Desktop**: Windows 8.1, 7 Ultimate, Vista Ultimate, 2000 Professional
- **Windows Server**: 2025, 2022, 2019, 2016, 2012, 2008, 2003

After Windows installation, you can change the display language using the standard Windows language settings.

## Resources

- [Clone applications](/manual/olares/market/clone-apps.md): Learn how to run multiple Windows instances on Olares.