---
outline: [2, 3]
description: Learn how to access the Olares One host terminal for command-line usage via Control Hub or SSH.
---

# Connect to Olares One via SSH

Certain development and operational tasks, such as connecting Olares One with other devices to form a cluster, require direct command-line access. 

Olares One is primarily used as a headless device, so you can access the terminal remotely without a monitor or keyboard connected to it.

You can connect to the host shell using one of the following methods:
- **Control Hub Terminal** is a web-based interface for direct root access. It is recommended for quick tasks.
- **Secure Shell (SSH)** is the standard protocol for remote management and complex operations.

## Before you begin

- The default username and password for Olares One are both `olares`.
   :::warning Reset default SSH password
   Even if you primarily use the Control Hub terminal, you must reset this password immediately in **Settings > My hardware** to secure your device against unauthorized access.
   :::
- SSH access grants powerful control over the system. Ensure you keep your credentials secure.

## Prerequisites
**Hardware**
- Your Olares One is set up and connected to a network.
- A client device, such as a computer, is required to access the terminal.

**Experience**
- Basic familiarity with terminal commands and the command-line interface (CLI).

**LarePass** (Required for remote access)
- The LarePass app is installed on your device. This is required only if you plan to connect via SSH from a remote location outside your local network.

## Method 1: Access via Control Hub
For quick access without configuring SSH clients, you can use the web-based terminal built directly into Control Hub.

1. Open the Control Hub app. 
2. In the left sidebar, under the **Terminal** section, click **Olares**.
   ![Terminal](/images/manual/olares/controlhub-terminal.png#bordered)

You can now execute system commands directly in the embedded terminal.

:::tip Run as `root`
The Control Hub terminal runs as `root` by default. You do not need to use `sudo` before commands.
:::

## Method 2: Access via SSH

SSH establishes a secure session over the network, allowing you to use command-line operations for Olares One on your current device.

### Get IP address of Olares One
To connect via SSH, you first need to find the internal IP address of your Olares One.

1. Open the LarePass app, and go to **Settings** > **System** to navigate to the **Olares management** page.
   ![Tap the System card](/images/manual/get-started/larepass-system.png#bordered)

2. Tap on the device card.
3. Scroll down to the **Network** section. You can find the **Intranet IP** there.

:::tip Check via Control Hub
You can check the IP using the `ifconfig` command in the Control Hub terminal.

Look for your active connection, typically named `enp3s0` (wired) or `wlo1` (wireless). The IP address follows `inet`.
:::
### Connect via SSH (local network)
If your computer and the Olares One are on the same Wi-Fi or LAN:

1. Open a terminal on your computer. 
2. Run the `ssh` command using Olares One's local IP address:
    ```bash
    ssh <username>@<host_ip_address>
    ```
    For example:

    ```bash
    ssh olares@192.168.31.155
    ```
3. Enter the host password when prompted.

### Connect via SSH (remote access)
If you are away from home, you can use LarePass VPN to bridge the connection securely.

#### Allow SSH access via VPN
For security, SSH access via VPN is disabled by default. You must enable it once.
1. On Olares, open the Settings app. 
2. Navigate to **VPN**. 
3. Toggle on **Allow SSH via VPN**.

#### Enable LarePass VPN
1. Open the LarePass desktop client on your computer. 
2. Click your avatar in the top-left corner and toggle on **VPN connection**.
    ![Enable LarePass VPN on desktop](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

3. Open a terminal on your computer. 
4. Run the `ssh` command using Olares One's local IP address. LarePass handles the routing automatically.
    ```bash
    ssh <username>@<host_ip_address>
    ```
    For example:

    ```bash
    ssh olares@192.168.31.155
    ```
5. Enter the host password when prompted.