---
outline: [2,3]
description: Learn how to configure VPN on Olares using Settings, covering VPN enforcement, SSH access, and subnet routing.
---
# Configure VPN access to Olares

The [LarePass VPN](../../larepass/private-network.md) provides secure remote access to your Olares device, even when you're on a different network or at a remote location. Olares' Settings app offers advanced configurations to tailor VPN access to your specific needs. Here, you can enforce VPN connections, enable SSH access over VPN, or route traffic through custom ports.

## Enforce access using VPN

To ensure that all traffic to your private Olares applications is encrypted and routed securely, you can enforce VPN access. This ensures that connections to your Olares always go through the LarePass VPN, regardless of the network or device used. Enabling this mode will block accesses to Olares via reverse proxy. 

To enable the enforced VPN mode:

1. Enable VPN connections on at least two devices using LarePass (typically a computer and a mobile phone) with LarePass installed. For detailed instructions, see [Enable VPN on LarePass](/manual/larepass/private-network.md).
2. Open Settings app from the Dock or Launchpad.
3. Click on your profile picture in the top-left corner, and scroll down to **Security** settings.
4. Turn on the switch for **Enforce VPN access to private entrance**.

When successful, you'll see a confirmation message at the bottom of the screen.


## Allow SSH connections via VPN
This enables SSH access to your Olares device through the LarePass VPN, even when you are in a different network or working remotely.

1. Open the Settings app, and select **System** > **VPN**.
2. Toggle on **Allow SSH Access via VPN**. Port **22** (SSH) is automatically added to the configuration.

   ![SSH via VPN](/images/manual/olares/ssh-via-vpn.png#bordered)
## Allow subnet routing
This feature allows you to access other devices in the same local network as your Olares through the VPN.

1. Open the Settings app, and select **System** > **VPN**.
2. Toggle on **Enable subnet routes**.

## Configure ACL rules for port access
After enabling subnet routing, you can further configure ACL (Access Control List) rules to allow traffic to specific ports based on the services you want to access.

For example, to access a Windows server via Remote Desktop:
1. Click <i class="material-symbols-outlined">add</i> to open the **Add ACL** dialog.
2. Enter `3389` (default port for Remote Desktop Protocol), and click **Confirm**.
3. Click **Apply** to apply changes.

   ![Add ACL port](/images/manual/olares/add-acl-port.png#bordered)

Now you can use Windows Remote Desktop to access the Windows server in the same LAN as Olares.