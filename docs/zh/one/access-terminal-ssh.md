---
outline: [2, 3]
description: Learn how to access the Olares One host terminal for command-line usage via Control Hub or SSH.
head:
  - - meta
    - name: keywords
      content: SSH, Olares Terminal, Control Hub
---

# Connect to Olares One via SSH  <Badge type="tip" text="10 min" />

Because Olares One is a headless device, you access its terminal remotely rather than through a directly connected monitor or keyboard. This is required for tasks such as cluster setup, system configuration, and maintenance.

You can connect to the host shell using one of the following methods:
- **Control Hub Terminal** is a web-based interface for direct `root` access. It is recommended for quick or occasional tasks.
- **Secure Shell (SSH)** is the standard protocol for remote management and more advanced or automated operations.

## Prerequisites

**Hardware**
- Your Olares One is set up and connected to a network.
- A client device, such as a computer, is required to access the terminal.

**Experience**
- Basic familiarity with terminal commands and the command-line interface (CLI).

## Method 1: Access via Control Hub

For quick access without configuring an SSH client, use the web-based terminal built into Control Hub.

1. Open the Control Hub app. 
2. In the left sidebar, under the **Terminal** section, click **Olares**.
   ![Terminal](/images/manual/olares/controlhub-terminal.png#bordered)

You can now execute system commands directly in the embedded terminal.

:::tip Run as `root`
The Control Hub terminal runs as `root` by default. You do not need to prefix commands with `sudo`.
:::

## Method 2: Access via SSH

SSH establishes an encrypted session over the network, allowing you to run command-line operations on Olares One from your current device.

### Step 1: Get the local IP address of Olares One

1. Open the LarePass app, and go to **Settings** > **System** to navigate to the **Olares management** page.
   ![Tap the System card](/images/manual/get-started/larepass-system.png#bordered)

2. Tap the Olares One device card.
3. Scroll down to the **Network** section and note the **Intranet IP**.

:::tip Check via Control Hub
You can check the IP using the `ifconfig` command in the Control Hub terminal.

Look for your active interface, typically `enp3s0` (wired) or `wlo1` (wireless). The IP address appears after `inet`.
:::

### Step 2: Check SSH password in Vault

<!--@include: ./reusables-reset-ssh.md{7,16}-->

### Step 3: Connect via SSH
The default username for Olares One is `olares`.

1. Open a terminal on your computer.
2. Run the following command, replacing `<local_ip_address>` with the Intranet IP:
   ```bash
   ssh olares@<local_ip_address>
   ```
3. When prompted, enter the SSH password.

## Advanced: SSH into Olares One from a different network

If you are not on the same local network as Olares One, use LarePass VPN to establish a secure connection without exposing your device to the internet.

### Step 1: Find the Tailscale IP of Olares One

1. On Olares, go to **Settings** > **VPN** > **View VPN connection status**.
2. Find **olares**, and click it to expand the connection details.
3. Locate the IP address that starts with `100.64`, and note it down.
    ![Enable LarePass VPN on desktop](/images/one/ssh-remote-ip.png#bordered){width=80%}

### Step 2: Check SSH password in Vault

<!--@include: ./reusables-reset-ssh.md{7,16}-->

### Step 3: Allow SSH access via VPN

1. On Olares, go to **Settings** > **VPN**.
2. Toggle on **Allow SSH via VPN**.
3. On your computer, open the LarePass desktop client.
4. Click your avatar in the top-left corner and toggle on **VPN connection**.
    ![Enable LarePass VPN on desktop](/images/one/ssh-enable-vpn.png#bordered)

### Step 4: Connect via SSH
The default username for Olares One is `olares`.
1. Open a terminal on your computer.
2. Run the following command, replacing `<tailscale_ip_address>` with the Tailscale IP address:
   ```bash
   ssh olares@<tailscale_ip_address>
   ```
   :::info
   After you enable SSH over VPN, the first SSH access is slower because VPN routes are being applied. Wait a short time for the connection to complete.
   :::
3. When prompted, enter the SSH password.

:::tip Connect using the local IP address instead
If **Subnet routes** is enabled in **Settings** > **VPN**, all devices on Olares One's local network become reachable through the VPN. You can then SSH using the local IP address (`192.168.x.x`) instead of the Tailscale IP (`100.64.x.x`), even when accessing from a different network.
:::

## Reset SSH password
<!--@include: ./reusables-reset-ssh.md{19,}-->
