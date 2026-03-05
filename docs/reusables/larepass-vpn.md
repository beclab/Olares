---
search: false
---
<!-- Reusable LarePass VPN content. Include by line range.
     Steps (no headings): Step 1 7-16, Step 2 18-41, Step 3 42-49.
     FAQs: 50-57 -->

To use the secure VPN connection, the LarePass client must be installed on the device you are using to access Olares.

- **Mobile**: Use the LarePass app installed during the Olares ID creation process.
- **Desktop**: Download and install the LarePass desktop client.

1. Visit <AppLinkGlobal />.
2. Download the version compatible with your operating system.
3. Install the application and log in with your Olares ID.

Once installed, enable the VPN directly on the device.

:::tip Always enable VPN for remote access
Keep LarePass VPN enabled. It automatically prioritizes the fastest available route to ensure you always get the best speed possible without manual switching.
:::
:::info iOS and macOS setup
On iOS or macOS, you may be prompted to add a VPN Configuration to your system settings the first time you enable the feature. Allow this to complete the setup.
:::

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

Once enabled, check the status indicator in LarePass to verify the connection type:

| Status       | Description                                              |
|:-------------|:---------------------------------------------------------|
| **Intranet** | Direct connection via your local LAN IP. Fastest speeds. |
| **P2P**      | Direct encrypted tunnel between devices. High speed.     |
| **DERP**     | Routed via a secure relay server. Used as a fallback.    |

### Why doesn't LarePass VPN work on my Mac anymore?

macOS blocks the VPN tunnel if the network extension or VPN configuration was not fully set up, or if the extension has become stuck or corrupted. See [LarePass VPN not working](/manual/help/ts-larepass-vpn-not-working) to reset the extension and restore the VPN.

### Why can't I enable LarePass VPN on Windows?

Third-party antivirus or security software may mistakenly flag LarePass as suspicious, preventing the VPN service from starting. See [LarePass VPN not working](/manual/help/ts-larepass-vpn-not-working) to resolve the issue.
