---
outline: [2,3]
description: Learn how to configure VPN on Olares using Settings, covering VPN enforcement, SSH access, and ACL ports.
---
# Configure VPN access to Olares

<!--
The [LarePass VPN](../../larepass/private-network.md) provides secure remote access to your Olares device, even when you're on a different network or at a remote location. Olares' Settings app offers advanced configurations to tailor VPN access to your specific needs. Here, you can enforce VPN connections, enable SSH access over VPN, or route traffic through custom ports.
-->

The [LarePass VPN](../../larepass/private-network.md) provides secure remote access to your Olares device, even when you're on a different network or at a remote location. Olares' Settings app lets you enforce VPN connections, enable SSH access over VPN, and manage ACL ports for services that need VPN access.

![VPN](/images/manual/olares/vpn-12.6.png#bordered)

## Allow SSH connections via VPN
This enables SSH access to your Olares device through the LarePass VPN, even when you are in a different network or working remotely.

1. Open the Settings app, and select **System** > **VPN**.
2. Toggle on **Allow SSH Access via VPN**. Port **22** (SSH) is automatically added to the configuration.

<!-- ## Allow subnet routing
This feature allows you to access other devices in the same local network as your Olares through the VPN.

1. Open the Settings app, and select **System** > **VPN**.
2. Toggle on **Enable subnet routes**.
-->

## Enforce access using VPN

To ensure that all traffic to your private Olares applications is encrypted and routed securely, you can enforce VPN access. This ensures that connections to your Olares always go through the LarePass VPN, regardless of the network or device used. Enabling this mode will block accesses to Olares via reverse proxy.

To enable the enforced VPN mode:

1. Enable VPN connections on at least two devices using LarePass (typically a computer and a mobile phone) with LarePass installed. For detailed instructions, see [Enable VPN on LarePass](/manual/larepass/private-network.md).
2. In Olares, navigate to **Settings** > **VPN**.
3. Toggle on **Enforce VPN access**.

When successful, you'll see a confirmation message at the bottom of the screen.

## Configure ACL ports
ACL ports let you allow VPN traffic to specific destination ports based on the services you want to access. Add only the ports required by the service.

For example, to access a Windows server via Remote Desktop:

1. In **ACL ports**, click <i class="material-symbols-outlined">add</i> to open the **Add port** dialog.
2. Enter `3389`, the default RDP port, and click **Add**.
3. Click **Apply** to apply changes.

Now you can use Windows Remote Desktop to access the Windows server in the same LAN as Olares.