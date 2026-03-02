---
search: false
---
<!-- Reusable LarePass VPN content. Include by line range.
     Steps (no headings): Step 1 7-16, Step 2 18-41, Step 3 42-49.
     FAQs: 50-75 -->

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

If you successfully enabled the VPN previously, but it has stopped working, you might need to reset the system extension.

:::info
Depending on your macOS version, the UI might look slightly different.
:::

1. Open **System Settings**, search for "Extension", and select **Login Items & Extensions**.
2. Scroll to the **Network Extensions** section and click the info icon (ⓘ) to view loaded extensions.
3. Find LarePass, click the three dots (...), and select **Delete Extension**.
4. Confirm the uninstallation.
5. Restart your Mac and re-enable the VPN in the LarePass desktop client.

### Why can't I enable LarePass VPN on Windows?

Third-party antivirus software might mistakenly flag the LarePass desktop client as suspicious, preventing it from launching the VPN service.

If prompted by your antivirus when opening LarePass for the first time, allow the application to continue.

If the VPN still fails to enable:
1. Open your security software and check if LarePass was blocked.
2. Add the main LarePass executable to the allowlist or exclusions of your antivirus.
3. Restart LarePass and enable the VPN.
