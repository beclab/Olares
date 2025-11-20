---
outline: [2, 3]
description: A comprehensive guide to installing and running a Windows virtual machine on Olares. Learn how to configure initial credentials, connect via browser-based VNC or Microsoft Remote Desktop (RDP), and transfer files between your computer and the VM.
---

# Run a Windows VM on your Olares device

Olares lets you run a full Windows virtual machine directly on your device, giving you a personal, always-available Windows environment accessible from macOS, Windows, or Linux.

:::info System capabilities
- Olares supports running essential Windows applications. 
- Workflows are limited to **CPU or integrated graphics performance**. GPU passthrough is not yet supported, meaning heavy GPU-accelerated applications may not perform optimally.
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
1. Open the **Market** app in your Olares web interface.
2. Use the search bar and type "Windows".
3. Click **Get**, then click **Install**.
    ![Install Windows](../public/images/manual/use-cases/win-install.png#bordered)
4. When prompted, set your Windows login credentials:
    - `WINDOWS_USERNAME`: Create a username for Windows access.
    - `WINDOWS_PASSWORD`: Set the corresponding password.
    ![Set environment variables](../public/images/manual/use-cases/win-set-env-var.png#bordered)

  These credentials are required for your first login.  
  You can change them later inside Windows.

5. Wait a few minutes for the installation and initialization to complete.

### Set up Windows

Once installation finishes, open Windows from Launchpad to start the VM for the first time.

Olares will automatically download and install the Windows 11 system image (about 5.5 GB). This may take several minutes depending on your network speed.
    ![Download Windows 11](../public/images/manual/use-cases/win-downloading-win11.png#bordered)

## Access the Windows VM

You can access your VM in two ways: 
- **Browser (VNC)**: for setup and quick tasks
- **Remote Desktop (RDP)**: for the best daily experience

### Access from the browser (VNC)

Open the Windows app from Launchpad to launch the VM directly in your browser using VNC.
::: info
VNC (Virtual Network Computing) provides immediate, clientless access without requiring any additional software. It is ideal for initial setup, troubleshooting, or emergency access when you cannot use RDP. However, it can feel less responsive and lacks advanced features like audio redirection and high-performance graphics.
:::
### Access using a Remote Desktop Client (RDP)
RDP (Remote Desktop Protocol) provides a much smoother, native-like experience with better performance, audio support, and seamless file transfer.

To connect via RDP:
1. [Enable VPN on LarePass](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device.

    When the VPN connection status shows `DERP`, `P2P`, or `Intranet`, the secure network is active and ready for remote access.

2. Install the Remote Desktop client.
   - **Windows:** No installation needed.
   - **macOS / iOS:** Download [Windows App from the App Store](https://apps.apple.com/us/app/windows-app/id1295203466).
   - **Android:** Download [Windows App from Google Play](https://play.google.com/store/apps/details?id=com.microsoft.rdc.androidx).

3. Find your Windows VM address.

    a. Open Windows from Launchpad in the brower view.

    b. In the browser URL bar, copy the main domain part and add `:46879` to the end.
    
    **Example**
    - Browser URL:
    ```
    https://7e89d2a1.laresprime.olares.com
    ``` 
    - RDP address: 
    ```
    7e89d2a1.laresprime.olares.com:46879
    ```

4. Add your Windows VM as an RDP connection. The following steps show the macOS interface, but the workflow is nearly identical on Windows and other platforms.

    a. Open the **Windows App** / **Microsoft Remote Desktop** on your device.

    b. Click the **＋** icon and select **Add PC**.

    c. In **PC name**, enter your Windows VM address you get from the previous step.
        ![Add PC](../public/images/manual/use-cases/win-add-pc.png#bordered)

    d. Click **Add**.

    e. Double click your saved PC entry, or click **⋯** and choose **Connect**.
        ![Connect to PC](../public/images/manual/use-cases/win-connect-device.png#bordered)
        
    f. When prompted, enter the `WINDOWS_USERNAME` and `WINDOWS_PASSWORD` you created earlier.
        ![Log in to PC](../public/images/manual/use-cases/win-log-in.png#bordered)

    g. If a security warning appears, click **Continue**.
        ![Continue to log in](../public/images/manual/use-cases/win-confirm-connect.png#bordered)

You are now connected to your Windows VM via RDP.
![Windows VM](../public/images/manual/use-cases/win-vm-interface.png#bordered)

## Optional: Change your Windows login password

You can update your Windows login password directly from inside the VM:
1. Click the search bar in the Windows taskbar and type "password".  
2. Select **Change your password**.  
    ![Change your password](../public/images/manual/use-cases/win-change-pw.png#bordered)
3. Click **Change** to set your new password.
    ![Set new password](../public/images/manual/use-cases/win-set-pw.png#bordered)

## Transfer files between your computer and Windows

RDP supports clipboard-based file transfers.

You can: 
- Copy any file on your Mac or PC.
- Paste it directly into the Windows VM.

The file appears immediately in Windows and is ready to use.

## Disconnect from the Windows VM

To end your RDP session, simply close the RDP window.

The Windows VM continues running on your Olares device and is always ready for you to reconnect.

## FAQ

### The Windows VM shows a blank screen or no desktop

The browser may have suspended the VNC connection due to inactivity to conserve system resources.  
    ![Reconnect VM](../public/images/manual/use-cases/win-vnc-reconnect.png#bordered)

Click **Connect** to restore the session.

### Windows system image download fails

If the Windows system image fails to download during setup:

- Wait a short while, then restart the application:
    1. Open Control Hub from the Launchpad.
    2. Select the windows project.
    3. Under **Deployment**, click windows.
    4. Click **Restart**.
    ![Restart VM](../public/images/manual/use-cases//win-restart.png#bordered)

  After the restart, the system image download will automatically retry.
- If repeated failures occur, your IP may have been temporarily blocked by Microsoft due to multiple download attempts in a short period.  
  Wait **24 hours**, then restart or reinstall the application and try again.
- If the issue persists, please contact us for assistance.

### Can I install other Windows versions or languages?

- Currently, only the default Windows 11 (English) installation image is supported. You cannot choose other Windows versions or language editions during installation.
- However, after Windows 11 is installed, you may **change the display language** within Windows itself using the standard Windows language settings.