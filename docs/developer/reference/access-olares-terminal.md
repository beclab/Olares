---
outline: [2, 3]
description: Complete guide to accessing the Olares host terminal via SSH or Control Hub.
---
# Access the Olares Terminal

Some development and operational tasks require running commands on the Olares host, such as inspecting disks, verifying host state, or updating host-level configuration. Since Olares hosts are commonly deployed without a monitor or keyboard, terminal access is provided remotely.

You can access the host shell using one of the following methods:
- **Secure Shell (SSH)** for standard remote management.
- **Control Hub terminal** for direct root access from the Olares web interface.

## Method 1: Access via SSH

SSH is the standard protocol for operating the Olares host from a remote development machine. This method establishes a secure session over the network.

### Prerequisites

Before connecting, ensure that you have:

- Network connectivity to the Olares host:
  - In most setups, your computer and the Olares host are on the same local network.
  - If you need to connect outside the local network, configure VPN access first. See [Connect over VPN using LarePass](#optionl-connect-over-vpn-using-larepass).
- Host IP address, typically `192.168.x.x`.
- Valid login credentials.

### Connect over local network

1. Open a terminal on your computer.
2. Run the SSH command using the following format:

   ```bash
   ssh <username>@<host_ip_address>
   ```

   Example:
   ```bash
   ssh olares@192.168.31.155
   ```
3. Enter the host password when prompted.

### Optionl: Connect over VPN using LarePass

If your computer is not on the same local network as your Olares device, enable LarePass VPN.

1. In Olares, open Settings, then navigate to **VPN**.
2. Enable **Allow SSH via VPN**.
3. Open the LarePass desktop client, and click your avatar in the top-left corner to open the user menu. 
4. Toggle on the switch for **VPN connection**. 
5. Open a terminal on your computer. 
6. Run the SSH command using the following format:

    ```bash
    ssh <username>@<host_ip_address>
    ```
    
    Example:
    ```bash
    ssh olares@192.168.31.155
    ```
7. Enter the host password when prompted.

## Method 2: Access via Control Hub

You can open a terminal session directly from Control Hub.

1. Open Control Hub.
2. In the left sidebar, click **Olares** in the **Terminal** section.

:::info Root access
The Control Hub terminal runs commands as `root` by default.

Do not use `sudo` before commands. For example, run `apt update` instead of `sudo apt update`.
:::