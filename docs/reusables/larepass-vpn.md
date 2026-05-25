---
search: false
---
<!-- Reusable LarePass VPN content. Include by named region. -->

<!-- #region install-larepass-client -->
To use the secure VPN connection, the LarePass client must be installed on the device you are using to access Olares.

- **Mobile**: Use the LarePass app installed during the Olares ID creation process.
- **Desktop**: Download and install the LarePass desktop client.

1. Visit <AppLinkGlobal />.
2. Download the version compatible with your operating system.
3. Install the application and log in with your Olares ID.
<!-- #endregion install-larepass-client -->

Once installed, enable the VPN directly on the device.

<!-- #region vpn-setup-notes -->
:::tip Always enable VPN for remote access
Keep LarePass VPN enabled. It automatically prioritizes the fastest available route to ensure you always get the best speed possible without manual switching.
:::
:::info iOS and macOS setup
On iOS or macOS, you may be prompted to add a VPN Configuration to your system settings the first time you enable the feature. Allow this to complete the setup.
:::
<!-- #endregion vpn-setup-notes -->

<!-- #region enable-larepass-vpn -->
<tabs>
<template #On-LarePass-mobile-client>

1. Open the LarePass app and go to **Settings**.
2. In the **My Olares** card, toggle on the VPN switch.

   ![Enable LarePass VPN on mobile](/images/manual/get-started/larepass-vpn-mobile.png#bordered)
</template>
<template #On-LarePass-desktop-client>

1. Open the LarePass app and click your avatar in the top-left corner to open the user menu.
2. Toggle on the switch for **VPN connection**.

   ![Enable LarePass VPN on desktop](/images/manual/get-started/larepass-vpn-desktop.png#bordered)
</template>
</tabs>
<!-- #endregion enable-larepass-vpn -->

<!-- #region check-vpn-status -->
Once enabled, check the status indicator in LarePass to verify the connection type:

| Status       | Description                                              |
|:-------------|:---------------------------------------------------------|
| **Intranet** | Direct connection via your local LAN IP. Fastest speeds. |
| **P2P**      | Direct encrypted tunnel between devices. High speed.     |
| **DERP**     | Routed via a secure relay server. Used as a fallback.    |
<!-- #endregion check-vpn-status -->

<!-- #region larepass-vpn-faq -->
### Why doesn't LarePass VPN work on my Mac anymore?

macOS may block the VPN tunnel if the network extension or VPN configuration was not fully set up, or if the extension has become stuck or corrupted.

To resolve this issue, follow [LarePass VPN not working](/manual/help/ts-larepass-vpn-not-working#macos-reset-the-network-extension) to reset the extension and restore the VPN.

### Why can't I enable LarePass VPN on Windows?

Third-party antivirus or security software may mistakenly flag LarePass as suspicious, preventing the VPN service from starting.

To resolve this issue, see [LarePass VPN not working](/manual/help/ts-larepass-vpn-not-working#windows-add-larepass-to-the-allowlist).

### Why does the VPN connection turn off after showing Connecting?

If you turn on the VPN connection, but the status stays on **Connecting** for a while and then turns off automatically, the system time on your LarePass client device may be incorrect.

To resolve this issue, see [LarePass VPN not working](/manual/help/ts-larepass-vpn-not-working.md#sync-device-time).
<!-- #endregion larepass-vpn-faq -->
