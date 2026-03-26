---
outline: [2, 3]
description: Access the Olares host terminal via Control Hub or SSH.
---
# Access the Olares terminal

Some development and operational tasks require running commands on the Olares host, such as inspecting disks, verifying host state, or updating host-level configuration. Since Olares hosts are commonly deployed without a monitor or keyboard, terminal access is provided remotely.

You can connect to the host terminal using one of the following methods:
- **Control Hub Terminal**: A web-based interface for direct `root` access. It is recommended for quick or occasional tasks.
- **Secure Shell (SSH)**: The standard protocol for remote management. It is recommended for advanced or automated operations.

:::tip For Olares One users
If you are using Olares One, refer to [SSH into Olares One](/one/access-terminal-ssh.md).
:::

## Access via Control Hub

For quick access without configuring an SSH client, use the web-based terminal built into Control Hub.

1. Open Control Hub.
2. In the left sidebar, under the **Terminal** section, click **Olares**.
  ![Terminal](/images/developer/develop/controlhub-terminal.png#bordered)

:::tip Run as `root`
The Control Hub terminal runs as `root` by default. You do not need to prefix commands with `sudo`.
:::

## Access via SSH

SSH establishes an encrypted session over the network, allowing you to run command-line operations on the Olares host from your computer.

### Connect over a local network

Before connecting, make sure you have:
- The local IP address of the Olares host.
- The username and password for the Olares host.

1. Open a terminal on your computer.
2. Run the SSH command using the following format:

   ```bash
   ssh <username>@<host_ip_address>
   ```

   For example:
   ```bash
   ssh olares@192.168.31.155
   ```
3. Enter the host password when prompted.

### Connect from a different network

If your computer is not on the same local network as the Olares host, enable LarePass VPN to establish a secure tunnel to your host.

Before connecting, make sure you have:

- LarePass installed and signed in on the computer you will use for the VPN connection.
- The username and password for the Olares host.

1. On Olares, go to **Settings** > **VPN**, and enable **Allow SSH via VPN**.
  ![Allow SSH via VPN](/images/developer/develop/access-terminal-allow-vpn.png#bordered){width=90%}

2. Open the LarePass desktop client, click your avatar in the top-left corner, and turn on **VPN connection**. 
3. On Olares, go to **Settings** > **VPN** > **View VPN connection status**, locate the host entry, and note the IP address that starts with `100.64`.
  ![View tailscale ip](/images/developer/develop/access-terminal-tailscale-ip.png#bordered){width=90%}

4. Open a terminal on your computer. 
5. Run the SSH command using the following format:

    ```bash
    ssh <username>@<tailscale_ip_address>
    ```
    
    For example:
    ```bash
    ssh olares@100.64.0.1
    ```
6. Enter the host password when prompted.