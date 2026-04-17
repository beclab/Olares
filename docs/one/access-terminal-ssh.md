---
outline: [2, 3]
description: Learn how to access the Olares One host terminal for command-line usage via SSH.
head:
  - - meta
    - name: keywords
      content: Olares One, terminal, SSH
---

# Access Olares One terminal via network

Secure Shell (SSH) establishes an encrypted session over the network, allowing you to run command-line operations on Olares One from your own computer. 

You can connect over your local network, or use a VPN to connect securely from a different location.

## Prerequisites

**Hardware**
- Your Olares One is set up and connected to a network.
- A client device, such as a computer, to access the terminal.
- A mobile device with the LarePass app installed.

**Experience**
- Basic familiarity with terminal commands and the command-line interface (CLI).

## Connect over your local network

Follow these steps if your device and your Olares One are on the same local network.

### Step 1: Get the local IP address of Olares One

1. Open the LarePass app on your mobile device, and then go to **Settings** > **System**.

   ![Tap the System card](/images/manual/get-started/larepass-system.png#bordered)

2. Tap the Olares One device card.
3. Scroll down to the **Network** section, and then note down the **Intranet IP**.

   :::tip Check via Control Hub
   You can check the IP using the `ifconfig` command in the Control Hub terminal.

   Look for your active interface, typically `enp3s0` (wired) or `wlo1` (wireless). The IP address appears after `inet`.
   :::

### Step 2: Check SSH password in Vault

<!--@include: ./reusables-reset-ssh.md{7,16}-->

### Step 3: Connect via SSH

The default username for Olares One is `olares`.

1. Open a terminal on your computer.
2. Run the following command, replacing `<local_ip_address>` with the Intranet IP you noted down:
   ```bash
   ssh olares@<local_ip_address>
   ```
3. When prompted, enter the SSH password obtained in Step 2.

## Advanced: Connect remotely from a different network

If your device is not on the same local network as your Olares One, use LarePass VPN to establish a secure connection without exposing your device to the internet.

### Step 1: Find the VPN IP address

1. On your Olares desktop, go to **Settings** > **VPN** > **View VPN connection status**.
2. Find **olares**, and then click it to expand the connection details.
3. Locate the IP address that starts with `100.64`, and then note it down.
    
    ![Enable VPN on LarePass desktop client](/images/one/ssh-remote-ip.png#bordered){width=80%}

### Step 2: Check SSH password in Vault

<!--@include: ./reusables-reset-ssh.md{7,16}-->

### Step 3: Enable SSH via VPN

1. Open your Olares desktop, and then go to **Settings** > **VPN**.
2. Toggle on **Allow SSH via VPN**.
3. On your computer, open the LarePass desktop client.
4. Click your avatar in the top-left corner and toggle on **VPN connection**.
    ![Enable LarePass VPN on desktop](/images/one/ssh-enable-vpn.png#bordered)

### Step 4: Connect via SSH

The default username for Olares One is `olares`.

1. Open a terminal on your computer.
2. Run the following command, replacing `<vpn_ip_address>` with the VPN IP address you noted down:
   ```bash
   ssh olares@<vpn_ip_address>
   ```
   :::info
   After you enable SSH via VPN, the first SSH access is slower because VPN routes are being applied. Wait a short time for the connection to complete.
   :::
3. When prompted, enter the SSH password obtained in Step 2.

:::tip Connect using the local IP address instead
If **Subnet routes** is enabled in **Settings** > **VPN**, all devices on Olares One's local network become reachable through the VPN. You can then SSH using the local IP address (`192.168.x.x`) instead of the VPN IP (`100.64.x.x`), even when accessing from a different network.
:::

## Reset SSH password
<!--@include: ./reusables-reset-ssh.md{19,}-->