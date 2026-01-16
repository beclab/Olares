---
outline: [2, 3]
description: Complete guide to accessing the Olares host terminal via SSH or Control Hub.
---
# Access the Olares Terminal

Some development and operational tasks require running commands on the Olares host, such as inspecting disks, verifying host state, or updating host-level configuration. Since Olares hosts are commonly deployed without a monitor or keyboard, terminal access is provided remotely.

You can access the host shell using one of the following methods:
1. **Secure Shell (SSH)** for standard remote management.
2. **Control Hub terminal** for direct root access from the Olares web interface.

## Access via SSH

SSH is the standard protocol for operating the Olares host from a remote development machine. This method establishes a secure session over the network.

### Prerequisites

Before connecting, ensure that you have:

- Network connectivity to the Olares host
  - In most setups, your computer and the Olares host are on the same local network.
  - If you need to connect outside the local network, configure VPN access first. See [Connect over VPN using LarePass](#connect-over-vpn-using-larepass).
- Host IP address, typically `192.168.x.x`.
- Valid login credentials.
    :::info
    On Olares Zero devices, the default username and password are both `olares`.
    :::

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
3. Enter the password when prompted.

### Connect over VPN using LarePass

If your computer is not on the same local network as your Olares device, enable LarePass VPN.

1. In Olares, go to **Settings** > **VPN**, then enable **Allow SSH via VPN**.
2. Open the LarePass desktop client, and click your avatar in the top-left corner to open the user menu.
3. Toggle on the switch for **VPN connection**.
4. Open a terminal on your computer.
5. Run the SSH command using the following format:

    ```bash
    ssh <username>@<host_ip_address>
    ```
    
    Example:
    ```bash
    ssh olares@192.168.31.155
    ```
6. Enter the password when prompted.

## Access via Control Hub

If you can sign in to the Olares web interface, you can open a terminal session directly from Control Hub.

1. Open **Control Hub**.
2. In the left sidebar, click **Olares** in the Terminal section.

:::info Root access
The Control Hub terminal runs commands as `root` by default.

Do not use `sudo` before commands. For example, run `apt update` instead of `sudo apt update`.
:::

## Learn about the root role

:::warning Use with caution
Root privileges allow modification of system-critical files and settings. Incorrect commands may cause system instability or data loss.
:::

The `root` user is the system superuser and has full privileges on the Olares host. When you access the terminal through Control Hub, commands are executed as `root` automatically. This design reduces friction for administrative tasks, since elevated permissions are available by default.

### Confirm the current user

To confirm which user the current terminal session is running as, run:

```bash
whoami
```

You can also infer the privilege level from the shell prompt:

- Prompts ending with `#` indicate a root session.
    
    Example: `root@host:~#`

- Prompts ending with `$` indicate a standard user session.

    Example: `olares@host:~$`
