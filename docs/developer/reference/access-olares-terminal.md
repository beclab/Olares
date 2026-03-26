---
outline: [2, 3]
description: Access the Olares host terminal via SSH or Control Hub.
---
# Access the Olares Terminal

Some development and operational tasks require running commands on the Olares host, such as inspecting disks, verifying host state, or updating host-level configuration. Terminal access is typically provided remotely through Control Hub or SSH.

You can connect to the host shell using one of the following methods:
- **Control Hub Terminal** is a web-based interface for direct `root` access. It is recommended for quick or occasional tasks.
- **Secure Shell (SSH)** is the standard protocol for remote management and more advanced or automated operations.

:::tip For Olares One users
If you are using Olares One, refer to [Connect to Olares One via SSH](/one/access-terminal-ssh.md).
:::

## Method 1: Access via Control Hub

For quick access without configuring an SSH client, use the web-based terminal built into Control Hub.

1. Open Control Hub.
2. In the left sidebar, under the **Terminal** section, click **Olares**.
  ![Terminal](/images/developer/develop/controlhub-terminal.png#bordered)

:::tip Run as `root`
The Control Hub terminal runs as `root` by default.

You do not need to prefix commands with `sudo`.
:::

## Method 2: Access via SSH

SSH establishes an encrypted session over the network, allowing you to run command-line operations on the Olares host from your current device.

### Prerequisites

Before connecting, ensure that you have the following:

- Access to the Olares host over the same local network, or through LarePass VPN if connecting from a different network. For remote access, see [Connect from a different network](#connect-from-a-different-network).
- The IP address of the Olares host.
- The username and password for the Olares host.

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

### Connect from a different network

If your computer is not on the same local network as the Olares host, enable LarePass VPN.

1. On Olares, go to **Settings** > **VPN**, and enable **Allow SSH via VPN**.
2. Open the LarePass desktop client, and click your avatar in the top-left corner to open the user menu. 
3. Toggle on the switch for **VPN connection**. 
4. On Olares, go to **Settings** > **VPN** > **View VPN connection status**, locate the host entry, and note the IP address that starts with `100.64`.
5. Open a terminal on your computer. 
6. Run the SSH command using the following format:

    ```bash
    ssh <username>@<tailscale_ip_address>
    ```
    
    Example:
    ```bash
    ssh olares@100.64.0.1
    ```
7. Enter the host password when prompted.