---
outline: [2, 3]
description: Access the Olares host terminal via SSH.
---
# Access the Olares terminal

Some development and operational tasks require running commands on the Olares host, such as inspecting disks, verifying host state, or updating host-level configuration. Since Olares hosts are commonly deployed without a monitor or keyboard, terminal access is provided remotely.

Connect to the Olares host from your computer using SSH. Use an account with `sudo` permission for commands that require administrator privileges.

:::tip For Olares One users
If you are using Olares One, refer to [SSH into Olares One](/one/access-terminal-ssh.md).
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
  ![View Tailscale IP](/images/developer/develop/access-terminal-tailscale-ip.png#bordered){width=90%}

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