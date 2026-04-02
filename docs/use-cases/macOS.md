---
outline: deep
description: Run MacOS as a virtual machine on Olares. Learn how to install MacOS from the Market, configure initial settings, and connect via browser-based VNC or VNC client.
head:
  - - meta
    - name: keywords
      content: Olares, MacOS, virtual machine, VNC, MacOS VM
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-03-25"
---

# Run a MacOS VM on your Olares device

Olares allows you to run MacOS as a virtual machine (VM) directly on your device. This enables access to Apple-specific applications and workflows from any computer via a web browser or VNC client.

:::tip System capabilities
- **Hardware dependency**: Performance depends on your CPU. GPU acceleration is not supported.
- **Use Case**: Ideal for macOS applications that do not require high-performance graphics.
:::

## Learning objectives

By the end of this tutorial, you will learn how to:
- Install and set up the MacOS VM on your Olares device.
- Access the MacOS VM directly in your web browser or via a VNC app.

## Install and configure MacOS VM

MacOS is available as an app in the Olares Market.

### Install MacOS

1. Open Market and search for "macOS".

   ![MacOS app in Market](/images/manual/use-cases/macos.png#bordered)

2. Click **Get**, and then click **Install**.
3. When prompted, set the environment variables:
    - **VERSION**: Select your preferred MacOS version from the dropdown list.
    - **DISK_SIZE**: Allocate disk space for MacOS.

   ![Set environment variables](/images/manual/use-cases/macos-set-env-var.png#bordered){width=65%}

4. Click **Confirm** and wait for the installation to finish.

### Format the virtual disk

1. Open the MacOS app from the Launchpad.

   :::tip First launch
   When you launch the app for the first time, Olares automatically downloads and installs the system image of the corresponding MacOS version. This might take several minutes depending on your network speed.
   :::

2. When the **Recovery** screen appears, select **Disk Utility** from the main menu, and then click **Continue**.

   ![MacOS Recovery menu](/images/manual/use-cases/macos-recovery-menu.png#bordered){width=50%}

3. From the left sidebar, select the **Apple Inc. VirtIO Block Media** with the largest capacity, and then click **Erase** on the toolbar.

   ![Select Disk Utility](/images/manual/use-cases/macos-select-disk-utility.png#bordered)

4. Configure the format:

    - **Name**: Enter a name. For example, `Macintosh HD`.
    - **Format**: Select **APFS**.

   ![Configure format](/images/manual/use-cases/macos-configure-format.png#bordered){width=50%}

5. Click **Erase** to format the disk.
6. Once complete, click **Done**.

   ![Disk formatted](/images/manual/use-cases/macos-disk-formatted.png#bordered){width=50%}

7. Close the **Disk Utility** window to return to the main menu.

### Install MacOS system

1. From the main menu, select **Reinstall macOS**, and then click **Continue**.

   ![Reinstall MacOS](/images/manual/use-cases/macos-reinstall.png#bordered){width=50%}

2. Accept the license agreement.
3. Select the disk you just formatted, and then click **Continue**.
4. Wait for the installation to finish. This might take 20-40 minutes depending on your network speed and hardware.

   ![MacOS installation progress](/images/manual/use-cases/macos-installing.png#bordered){width=60%}

### Complete initial setup

After the system installation finishes:

1. Select your country, preferred language, and accessibility features as needed.
2. When you reach **Migration Assistant**, select **Not Now** in the lower left corner to skip migrating data from another Mac.
3. When prompted for **Sign In with Your Apple ID**, select **Set Up Later** in the lower left corner to skip for now.
4. Set up a username and password for the MacOS account. For the remaining setup steps, you can skip or accept the defaults.

   ![MacOS desktop](/images/manual/use-cases/macos-desktop.png#bordered)

## Access the MacOS VM

### Access from browser

Use for initial setup, quick access, or troubleshooting.

Open the MacOS app from the Launchpad to launch the VM directly in your browser. 

### Access using VNC Viewer

Use for daily work with better stability and latency.

#### Step 1: Obtain connection details

:::warning Multiple MacOS instances
Each MacOS instance uses a unique port. If you have cloned the MacOS app, ensure you check the **ACLs** section for the specific instance you want to access.
:::

1. Open Settings, and then go to **Applications** > **MacOS**.
2. Under **Entrances**, click **MacOS**, and then note down the endpoint address. For example, `https://43b9d8ea.alexmiles.olares.com`. 

   ![Locate endpoint](/images/manual/use-cases/macos-endpoint.png#bordered){width=80%}

3. Go back to the previous page, click **ACLs**, and then note down the port number. For example, `49238`.

   ![Locate port number](/images/manual/use-cases/macos-port-number.png#bordered){width=80%}

4. Format the address for VNC connection, that is `43b9d8ea.alexmiles.olares.com:49238`.

#### Step 2: Connect via VNC Viewer

<Tabs>
<template #macOS>

1. Open the LarePass desktop client, and then enable the VPN connection.

   ![Enable VPN on LarePass desktop](/images/manual/use-cases/alex-larepass-vpn-desktop.png#bordered){width=90%}

   When the VPN connection status shows **P2P** or **Intranet**, the secure network is active.

2. (Optional) [Install Homebrew](https://brew.sh) if you did not.
3. Open a terminal on your computer, and then run the following command to install the the VNC Viewer app: 

   ```bash
   brew install --cask vnc-viewer
   ```

   The following message indicates successful installation:

   ```text
   vnc-viewer was successfully installed!
   ```
4. Open VNC Viewer from your computer and sign in with your RealVNC account. If you do not have the account, create one and then sign in.
5. Click **File** > **New connection**.   
6. Enter the address obtained from **Step1**. 

   For example, the browser URL of the VM is `https://43b9d8ea.alexmiles.olares.com/`, and the port number is `49238`, then the address to enter is `43b9d8ea.alexmiles.olares.com:49238`.

   ![New connection in VNC Viewer](/images/manual/use-cases/vnc-new-connection.png#bordered){width=60%}

7. Click **OK**. The connected is saved in the VNC Viewer.

   ![VM connected in VNC Viewer](/images/manual/use-cases/vnc-vm-connected.png#bordered)

8. Double-click the saved connection to connect.
9. If the Unencrypted connection warning appears, click **Continue**.
10. When prompted, enter the username and password you created earlier. 

      You are now connected to your MacOS VM via the VNC Viewer.

11. To disconnect from the MacOS VM, close the VNC Viewer window. 
   
      The MacOS VM continues running on your Olares device and remains ready for you to reconnect.

</template>
<template #Windows>

1. Open the LarePass desktop client, and then enable the VPN connection.

   ![Enable VPN on LarePass desktop](/images/manual/use-cases/alex-larepass-vpn-desktop.png#bordered){width=90%}

   When the VPN connection status shows **P2P** or **Intranet**, the secure network is active.

2. Download and install [RealVNC Viewer](https://www.realvnc.com/en/connect/download/viewer/).

3. Open VNC Viewer and create a new connection:

   a. Click **File** > **New connection**.

   b. Enter the address obtained from **Step1**. 

   c. Save the connection.

4. Double-click the saved connection to connect.

</template>
</Tabs>

## FAQs

### Can I use my Apple ID with this VM?

While you can sign in with an Apple ID during setup, some Apple services might not function correctly in a virtualized environment. For best results, use local accounts or skip Apple ID setup.

### What MacOS versions are supported?

Currently supported versions:
- MacOS 14 Sonoma
- MacOS 13 Ventura
- MacOS 12 Monterey
- MacOS 11 Big Sur
- MacOS 10 Catalina

### The connection closed unexpectedly

When you attempt to connect with the MacOS VM from the VNC Viewer, the error `The connection closed unexpectedly` occurs.

This usually happens when the LarePass VPN is disabled. 

Open your LarePass desktop client and ensure that the VPN connection status is **P2P** or **Intranet**. Then try to connect again.

## Learn more

- [dockur/macos GitHub repository](https://github.com/dockur/macos)
- [Run a Windows VM on your Olares device](./windows.md)
