---
outline: [2,3]
description: Learn how to access Olares apps and services directly via your local network (LAN) for maximum speed, privacy, and offline reliability.
---
# Access Olares services locally
While remote access is convenient, accessing your device directly over your Local Area Network (LAN) offers significant advantages:
- **Maximum performance:** Transfer files at full gigabit speeds without routing through the internet.
- **Enhanced privacy:** Keep your traffic strictly within your home network.
- **Offline independence:** Access your data and dashboard even when your internet service provider is down.

This guide covers several methods to establish a local connection:
- [Enable LarePass VPN (Recommended)](#method-1-enable-larepass-vpn-recommended)<br/>This method is the easiest solution, as it automatically establishes the fastest connection without manual configuration.
- [Use `.local` domain](#method-2-use-local-domain)<br/>This method requires no installation, though you must use specific URL formats based on your operating system.
- [Configure local DNS (Advanced)](#method-3-configure-local-dns)<br/>This method allows standard URLs to work locally by updating DNS settings on your router or individual computer.
- [Modify host files (Fallback)](#method-4-modify-host-files)<br/>This method manually maps standard URLs to your local IP on a single computer, ensuring access even without an internet connection.

## Method 1: Enable LarePass VPN (Recommended)
The most robust way to connect, whether you are sitting next to the device or traveling, is using the LarePass VPN. It intelligently detects when you are on the same network and switches to a direct **Intranet** mode for maximum speed.

:::tip Always enable VPN for remote access
Keep **LarePass VPN** enabled. It automatically prioritizes the fastest available route to ensure you always get the best speed possible without manual switching.
:::
:::info iOS and macOS setup
On iOS or macOS, you may be prompted to add a **VPN Configuration** to your system settings the first time you enable the feature. Please allow this to complete the setup.
:::

<tabs>
<template #On-LarePass-mobile-client>

1. Open the LarePass app, and go to **Settings**.
2. In the **My Olares** card, toggle on the VPN switch.
</template>
<template #On-LarePass-desktop-client>

1. Open the LarePass app, and click your avatar in the top-left corner to open the user menu.
2. Toggle on the switch for **VPN connection**.
</template>
</tabs>

Once enabled, you can check the **Status** indicator in LarePass to confirm you are using a local connection:

| Status       | Description                                              |
|:-------------|:---------------------------------------------------------|
| **Intranet** | Direct connection via your local LAN IP. Fastest speeds. |
| **P2P**      | Direct encrypted tunnel between devices. High speed.     |
| **DERP**     | Routed via a secure relay server. Used as a fallback.    |

## Method 2: Use `.local` domain
If you prefer not to use a VPN, you can access services using the `.local` domain. There are two domain formats available depending on your compatibility needs.

### Single-level hostname (All operating systems)
:::warning Supported for community apps only
Olares system apps such as Desktop and Files do not support this URL format and will not load correctly.
:::
This format uses a single-level hostname by connecting the entrance ID and the username with hyphens.
- **Default URL**:
   ```plain
   https://<entrance_id>.<username>.olares.com
   ```
- **Local-access URL**:
   ```plain
   http://<entrance_id>-<username>-olares.local
   ```

For example, if the default URL is `https://a45f345b.laresprime.olares.com`, then the corresponding local URL is `http://a45f345b-laresprime-olares.local`.

### Multi-level hostname (macOS and iOS only)
Apple devices support local service discovery via [Bonjour](https://developer.apple.com/bonjour/) (zero‑configuration networking), which can resolve multi‑label hostnames under `.local` on macOS and iOS. This allows a local URL format that mirrors the remote address.

- **Default URL**:
   ```plain
   https://<entrance_id>.<username>.olares.com
   ```
- **Local-access URL**:
   ```plain
   http://<entrance_id>.<username>.olares.local
   ```
For example, if the default URL is `https://a45f345b.laresprime.olares.com`, then the corresponding local URL is `http://a45f345b.laresprime.olares.local`.

## Method 3: Configure local DNS
For a seamless experience where standard URLs (`olares.com`) resolve to your local IP address automatically, you can configure your network DNS. This configuration ensures consistent access across all devices on the network without requiring individual client setup.

### Find the internal IP for Olares device

<tabs>
<template #Check-via-the-LarePass-mobile-client>

If your phone and Olares device are on the same network:
1. Open the LarePass app, and go to **Settings**.
2. Tap the **System** area to navigate to the **Olares management** page.
3. Tap on the device card and scroll down to the **Network** section. You can find the **Intranet IP** there.

</template>
<template #Check-via-Olares-Terminal>

Control Hub provides a built-in terminal that allows you to run system commands directly from the browser, without needing an external SSH client.
1. Open the Control Hub app, and select **Terminal** in the left navigation bar.
2. Type `ifconfig` in the terminal and press **Enter**.
3. Look for your active connection, typically named `enp3s0` (wired) or `wlo1` (wireless). The IP address follows `inet`.

   Example output:
   ```bash
    enp3s0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
    inet 192.168.50.116  netmask 255.255.255.0  broadcast 192.168.50.255
    inet6 fe80::4194:4045:c35e:7b32  prefixlen 64  scopeid 0x20<link>
    ether d8:43:ae:54:ce:fc  txqueuelen 1000  (Ethernet)
    RX packets 80655321  bytes 71481515308 (71.4 GB)
    RX errors 0  dropped 136  overruns 0  frame 0
    TX packets 51867817  bytes 15924740708 (15.9 GB)
    TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
   ```
   In this example, `192.168.50.116` is the internal IP for your Olares device.
</template>
</tabs>

### Configure DNS
With the internal IP address identified, you must now configure your DNS settings to route traffic correctly. You can apply this configuration to a single computer for individual access, or update your router to enable seamless local resolution for all devices on your network.
<tabs>
<template #Configure-for-local-device>

Update the DNS settings on your specific computer. For example, on macOS:
1. Open Apple menu and go to **System Settings**. 
2. Select **Wi-Fi**, then click **Details** on your connected network.
3. Select **DNS** and update the server list:

   a. Click the **+** button under **DNS Servers** to add your Olares device's internal IP (e.g., `192.168.x.x`).

   b. Ensure the Olares IP is listed at the top. Add your original DNS (or 1.1.1.1) below it as a fallback. <br/>This ensures that if your Olares device shuts down, the router will automatically switch to the secondary DNS, keeping your internet connection alive.

4. Click **OK** to save changes.

</template>

<template #Configure-for-all-devices>

Update the DNS on your router to apply changes to every device in your network.

1.  Log in to your router's admin panel.
2.  Navigate to **DHCP / DNS Settings**.
3.  Set **Primary DNS** to your Olares device's internal IP (e.g., `192.168.x.x`).
4.  Set **Secondary DNS** to your **current Primary DNS** (or a public provider like `1.1.1.1`). <br/>This ensures that if your Olares device shuts down, the router will automatically switch to the secondary DNS, keeping your internet connection alive.
5.  Save and reconnect your devices to refresh the DNS cache.
</template>
</tabs>

Once configured, you can access Olares using both your standard public address (`<entrance_id>.<username>.olares.com`) and your local address (`<entrance_id>.<username>.olares.local`).
:::tip
You can install AdGuard Home from the Olares Market to monitor traffic and manage DNS rewrites graphically.
:::
## Method 4: Modify host files
If you cannot change router settings and need immediate offline access on a specific computer, you can manually map the domains in your hosts file.

1.  Locate your hosts file:
    - **Windows:** `C:\Windows\System32\drivers\etc\hosts`
    - **macOS/Linux:** `/etc/hosts`
2.  Open the file with a text editor, which requires Administrator privileges.
3.  Add the mapping lines:
    ```plain
    # Replace the internal IP and the username
    # Olares apps
    192.168.31.208  desktop.<username>.olares.com
    192.168.31.208  auth.<username>.olares.com
    192.168.31.208  files.<username>.olares.com
    192.168.31.208  market.<username>.olares.com
    192.168.31.208  settings.<username>.olares.com
    192.168.31.208  dashboard.<username>.olares.com
    192.168.31.208  control-hub.<username>.olares.com
    192.168.31.208  profile.<username>.olares.com
    192.168.31.208  vault.<username>.olares.com
    # Add other community apps as needed
    192.168.31.208  <entrance_id>.<username>.olares.com
    ```
This allows you to access Olares locally without requiring internet connectivity.
## FAQs
### Why enabling LarePass VPN on Mac does not work anymore?
If you successfully enabled the VPN previously, but it has stopped working, you might need to reset the system extension.
:::info
Depending on your macOS version, the UI could be slightly different.
:::
1. Open **System Settings**, search for "Extension", and select **Login Items & Extensions**.
2. Scroll to the **Network Extensions** section and click the info icon (ⓘ) to view loaded extensions.
3. Find LarePass, click the three dots (...), and select **Delete Extension**.
4. Confirm the uninstallation.
5. Restart your Mac and re-enable the VPN in the LarePass desktop client.

### Why I cannot enable LarePass VPN on Windows?

### Why the `.local` domain does not work in Chrome (macOS)?
Chrome may fail to access local URLs if macOS blocks local network permissions.
To enable access:
1. Open Apple menu and go to **System Settings**.
2. Go to **Privacy & Security** > **Local Network**.
3. Find Google Chrome and Google Chrome Helper in the list and enable the toggles.
   ![Enable local network](/public/images/manual/larepass/mac-chrome-local-access.png#bordered){width=400}

4. Restart Chrome and try accessing the local URL again.