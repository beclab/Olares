---
outline: [2, 3]
description: Run Linux servers on Olares with QEMU Server. Connect over your LAN or LarePass VPN, resize resources, and create independent distributions with Clone.
head:
  - - meta
    - name: keywords
      content: "Olares, QEMU Server, Linux, virtual machine, SSH, Overlay gateway, LarePass VPN, Clone"
app_version: "0.0.27"
doc_version: "1.0"
doc_updated: "2026-10-16"
---

# Run Linux servers with QEMU Server

QEMU Server runs a Linux virtual machine on Olares with KVM hardware acceleration. It provides a server environment for command-line tools, development, and services, with SSH access and a browser console.

Each installation has its own system disk and resource settings. You can keep Ubuntu for development and create a separate Debian, Rocky Linux, or Fedora server with Clone.

## Prerequisites

- An amd64 or ARM64 Olares device with hardware virtualization enabled and a usable `/dev/kvm` device. The guest architecture follows the host architecture.
- Available CPU, memory, and storage for your virtual machines. Allow an additional 1 GiB of memory per instance for QEMU, beyond the guest memory allocation.

## Install QEMU Server

1. Open Market and search for "QEMU Server".
   <!-- ![QEMU Server in Market](/images/manual/use-cases/qemu.png#bordered) -->

2. Click **Get**, complete the required configuration when prompted, then click **Install**, and wait for installation to complete.

Use these settings for a new server. The example values are suggestions; enter your own values in the installation form.

| Setting | Accepted values | Example | Purpose |
| --- | --- | --- | --- |
| `DISTRIBUTION` | Ubuntu 24.04/26.04 LTS, Debian 12/13, Rocky Linux 9/10, Fedora 43/44 | Ubuntu 24.04 LTS | Select the operating system for this installation. |
| `CPU_CORES` | 1–8 | 2 | Number of virtual CPU cores. These are shared host resources, not dedicated physical cores. |
| `MEMORY_GB` | 1–32 | 2 | Guest memory in GiB. |
| `DISK_SIZE_GB` | 16–512 | 40 | Virtual system disk capacity in GiB. |
| `INITIAL_PASSWORD` | 8–128 characters | Your own strong password | Initial Linux account password. |

The application also supports an optional CDN override. Leave it unchanged to use the configured Olares CDN.

## Wait for Linux initialization

Open QEMU Server from Launchpad. The page shows image download progress, verification, disk preparation, and initialization. Wait until the server is ready before connecting. An application marked as running in Olares might still be downloading or initializing Linux.

Images download from the Olares CDN. Download time depends on the image size and network connection. Interrupted downloads retain partial data and resume on retry. If the page reports a download failure, restart the application and check the progress again.

Use the account for your selected distribution and the password you entered during installation:

| Distribution | SSH username |
| --- | --- |
| Ubuntu | `ubuntu` |
| Debian | `debian` |
| Rocky Linux | `rocky` |
| Fedora | `fedora` |

These accounts support `sudo`. The initial account is not `root`.

## Connect with SSH

Choose the method that matches your network. Both methods connect to the same Linux system with the same account and password.

| Method | SSH host | SSH port | Requirements |
| --- | --- | --- | --- |
| Same LAN: Overlay gateway | The IP assigned to this application instance | `22` | Overlay gateway enabled for this instance; your computer can reach that LAN address. |
| Different network: LarePass VPN | The Olares device address reachable through VPN | The external SSH port assigned to this instance | LarePass VPN enabled on your computer. |

The IP addresses and port numbers below are examples. Replace them with the values shown for your device and application. The browser page URL is not an SSH address.

<tabs>
<template #Same-LAN>

### Connect through Overlay gateway

Overlay gateway gives each instance a dedicated local IP. Your SSH connection uses that IP and the default port 22.

Olares must run on a native Linux host with wired Ethernet for Overlay gateway. WSL is not supported. Your computer can use Wi-Fi, provided its network can reach the assigned local IP; guest Wi-Fi isolation or separate VLANs can block access.

1. Open **Settings** > **Network** > **Overlay gateway**. Ask the Super admin to enable the system-level service if it is off.
2. Under **Applications**, enable the QEMU Server instance you want to access and confirm the change. A running instance restarts. Wait for Linux to become ready again.
3. Copy the local IP shown for that instance. For example, `192.168.1.51`.
4. Connect from Terminal, PowerShell, or another SSH client. For Ubuntu:

   ```bash
   ssh ubuntu@192.168.1.51
   ```

5. Enter the Linux password when prompted. In Xshell, select SSH, use the instance IP as the host, and set the port to `22`.

Each clone has a separate Overlay IP. All instances can use port 22 because their IP addresses differ.

</template>
<template #Different-network>

### Connect through LarePass VPN

LarePass VPN connects your computer to Olares when you are away. This method uses the Olares device address and an instance-specific external SSH port. Enabling Overlay gateway is not a prerequisite for this method.

1. Open LarePass on your computer, sign in with the Olares ID that has access to the device, and enable VPN.

   ![Enable LarePass VPN on a computer](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

2. Obtain the Olares device address reachable through your VPN connection. Use the device address, rather than an application's web entrance or an address copied from an unrelated instance.
3. Open **Settings** > **Applications** > **QEMU Server** > **Ports**. For a clone, select its own application title. Find the external port for SSH.
4. Connect using both values. For example, if the reachable device address is `192.168.1.10` and the external SSH port is `43022`:

   ```bash
   ssh -p 43022 ubuntu@192.168.1.10
   ```

5. Enter the Linux password. In Xshell, enter the same device address and external port in their separate fields.

Do not substitute port 22 or the internal service port for the external port shown in Settings. Clones on the same Olares device use different external SSH ports. A public web entrance does not expose SSH; this workflow uses LarePass VPN. Connection speed depends on your network and the VPN route.

</template>
</tabs>

:::tip SSH command syntax
On macOS, Linux, and Windows OpenSSH, specify a custom port with `-p`. Use `ssh -p 43022 ubuntu@192.168.1.10`, rather than appending `:43022` to the host. Do not include `http://` or `https://`.
:::

## Use the browser console

Open the console from the QEMU Server application page and sign in with the Linux account. The console is useful for checking boot output or working when SSH is unavailable. This is a Linux server console, not a graphical desktop.

## Create another distribution with Clone

Clone creates a fresh, independent installation using the same application template. It does not copy the existing VM disk, installed software, files, or Linux accounts. The original server remains a separate instance.

For example, create a Debian server alongside an existing Ubuntu server:

1. Open Market and select **My Olares**.
2. Find the installed QEMU Server application. Open the drop-down menu beside **Open** and select **Clone**.
3. In **Clone app**, set **New app title** to a recognizable name, such as `Debian Services`, and set **Desktop shortcut name**. Click **Confirm**.
4. In **Configure Environment Variables**, select Debian 13 for `DISTRIBUTION`. Enter CPU, memory, disk capacity, and a password for the new instance, then confirm.
5. Open the new instance and wait for its image download and initialization. Sign in as `debian` using its own password.
6. Enable Overlay gateway separately for the clone to use local access, or find its own external SSH port for VPN access.

Keep a record of each title, distribution, username, and connection address. For example:

| Instance | Distribution | Username | Same-LAN connection | VPN connection |
| --- | --- | --- | --- | --- |
| Ubuntu Dev | Ubuntu 24.04 LTS | `ubuntu` | Its Overlay IP, port 22 | Device address, Ubuntu instance's external port |
| Debian Services | Debian 13 | `debian` | A different Overlay IP, port 22 | Same device address, Debian instance's external port |

The distribution cannot be changed on an existing system disk. Use a new independent installation for another distribution or release. Clones consume additional host resources; their combined allocations must fit your device's available capacity.

## Adjust CPU, memory, and disk capacity

1. Back up important data and finish active work. Applying resource changes restarts the VM and disconnects SSH sessions.
2. Open **Settings** > **Applications** and select the instance.
3. Under **Environment variables**, click **Manage environment variables**.
4. Edit `CPU_CORES`, `MEMORY_GB`, or `DISK_SIZE_GB` within the supported ranges. Confirm your changes and click **Apply**.
5. Wait until Linux is ready. Reconnect and check the resources:

   ```bash
   nproc
   free -h
   lsblk
   df -h /
   ```

Increasing the disk capacity expands the virtual disk and supported root filesystems during restart. The root filesystem is smaller than the whole disk because partitions and filesystem metadata also use space. Guest-visible memory can also be slightly below the configured amount.

:::warning Disk capacity cannot be reduced
Entering a smaller disk size retains the existing larger disk and displays a warning. It does not reclaim storage. Back up your data and create a smaller independent installation to reduce capacity.
:::

## Understand persistent storage

The VM system disk is saved in this instance's Olares App Data. Linux packages, user files, and services stored on that disk survive application restarts and resource changes. Clones have separate storage.

The VM disk does not automatically expose your Olares Files folders inside Linux. Transfer files with SSH-based tools such as SFTP or SCP, or configure a network share separately. Persistence is not a backup: keep a separate copy of important data. Uninstalling an instance and deleting its application data removes its VM disk.

## Troubleshoot connections

| Symptom | Check |
| --- | --- |
| Cannot resolve hostname | Remove `:port` from the hostname; use `-p` for a custom SSH port. |
| Connection times out on the LAN | Confirm the instance's current Overlay IP, gateway state, Ethernet connection on Olares, and LAN routing or isolation rules. |
| Connection times out remotely | Confirm LarePass VPN is enabled, the device address is reachable, and the external SSH port belongs to this instance. |
| Password is rejected | Check the distribution's username and the current Linux password. If you changed it inside Linux, the initial password is no longer the login password. |
| Application is running but SSH is unavailable | Open the application page and wait for Linux initialization. Use the browser console to inspect startup. |
| Image download failed | Restart the application to retry the download using retained partial data. |

## Learn more

- [Verify your QEMU Server installation](qemu-validation.md): Repeatable acceptance checks and validation coverage.
- [Clone applications](../manual/olares/market/clone-apps.md): Manage multiple application instances.
- [Manage Overlay gateway](../manual/olares/settings/overlay-gateway.md): Requirements, permissions, and network configuration.
- [Manage application environment variables](../manual/olares/settings/manage-app-env.md): Change instance settings.
