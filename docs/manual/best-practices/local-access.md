---
outline: [2,3]
description: Learn the different methods to access Olares services locally for improved speed and offline capability.
---
# Access Olares services locally

Olares is designed to provide seamless access to your self-hosted services anytime, anywhere.

However, accessing your devices locally provides several advantages:
- **Maximum performance**: Transfer files at full speed without the latency and potential bottlenecks of the internet.
- **Enhanced privacy**: Keep your traffic contained within your home network for added security.
- **Offline independence**: Access your data and apps even when your internet service is unavailable.

## Objectives
By the end of this tutorial, you will learn how to:

- Establish a secure, high-speed local connection using the LarePass VPN.
- Access Olares services using `.local` domains.
- Configure local DNS to allow standard URLs to resolve locally across your entire network.
- Manually map hosts files to ensure access on specific machines without internet.

## Choose a connection method
There are four ways to establish a local connection:

* **[Method 1: Enable LarePass VPN](#method-1-enable-larepass-vpn)**<br/>
  Uses LarePass VPN to automatically detect your local network and optimize the connection speed without changing settings.
* **[Method 2: Use `.local` domain](#method-2-use-local-domain)**<br/>
  Access the device via a specific local URL format. No installation required.
* **[Method 3: Configure local DNS](#method-3-configure-local-dns)**<br/>
  Updates your router or computer's DNS settings to map the standard Olares URL to the local IP address.
* **[Method 4: Modify hosts files](#method-4-modify-hosts-files)**<br/>
  Manually maps the standard Olares URL to the local IP on a single computer.

## Method 1: Enable LarePass VPN
The LarePass VPN is designed to secure your connection while optimizing performance. When enabled, LarePass detects if you are on the same network as your device and switches to **Intranet** mode.

<!--@include: ../../reusables/larepass-vpn.md{19,24}-->

Enable the LarePass VPN directly on the device you are currently using to access Olares.

<!--@include: ../../reusables/larepass-vpn.md{26,50}-->

## Method 2: Use `.local` domain

If you prefer not to install additional apps, you can access services using the `.local` domain. There are two domain formats available depending on your operating system.

### Single-level domain (All operating systems)
:::warning Supported for community apps only
Olares system apps such as Desktop and Files do not support this URL format and will not load correctly.
:::
This format uses a single-level domain by connecting the entrance ID and the username with hyphens (`-`).

**Standard URL**
```plain
https://<entrance_id>.<username>.olares.com
```
**Local URL**
```plain
http://<entrance_id>-<username>-olares.local
```

### Multi-level domain

The multi-level format below matches the structure of your standard Olares URL. Use it as shown.

<!--@include: ../../reusables/local-domain.md{7,23}-->

![Multi-level local domain](/images/manual/get-started/multilevel-local-domain-mac.png#bordered)

#### macOS and iOS
Apple devices support local service discovery via [Bonjour](https://developer.apple.com/bonjour/) (zero‑configuration networking), which can resolve multi‑label domains under `.local` on macOS and iOS.

Therefore, no extra setup is needed. You can directly use local URL in your browser.

#### Windows

<!--@include: ../../reusables/local-domain.md{26,40}-->

## Method 3: Configure local DNS
For a seamless experience where standard URLs resolve to your local IP address automatically, you can configure your network DNS. This configuration ensures consistent access across all devices on the network without requiring individual client setup.

### Find the internal IP for Olares device
To configure DNS, first you need to find the internal IP for your Olares device.
<tabs>
<template #Check-via-the-LarePass-mobile-client>

If your phone and Olares device are on the same network:
1. Open the LarePass app, and go to **Settings** > **System** to navigate to the **Olares management** page
   ![Tap the System card](/images/manual/get-started/larepass-system.png#bordered)

2. Tap on the device card.
   ![Tap the device card](/images/manual/get-started/larepass-device-card.png#bordered)

3. Scroll down to the **Network** section. You can find the **Intranet IP** there.
   ![Find Network section](/images/manual/get-started/larepass-network.png#bordered)

</template>
<template #Check-via-Olares-Terminal>

Control Hub provides a built-in terminal that allows you to run system commands directly from the browser, without needing an external SSH client.
1. Open the Control Hub app, and under **Terminal**, select **Olares** in the left navigation bar.
   ![Find internal IP from Control Hub](/images/manual/get-started/find-internal-ip-from-controlhub.png#bordered)

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
   In this example, `192.168.50.116` is the internal IP.
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

   b. Ensure the Olares IP is listed at the top. Add your original DNS (or `1.1.1.1`) below it as a fallback. <br/>This ensures that if your Olares device shuts down, the router will automatically switch to the secondary DNS, keeping your internet connection alive.

4. Click **OK** to save changes.

</template>

<template #Configure-for-all-devices>

Update the DNS on your router to apply changes to every device in your network.

1.  Log in to your router's admin panel.
2.  Navigate to **DHCP / DNS Settings**.
3.  Set **Primary DNS** to your Olares device's internal IP (e.g., `192.168.x.x`).
4.  Set **Secondary DNS** to your current Primary DNS (or a public provider like `1.1.1.1`). <br/>This ensures that if your Olares device shuts down, the router will automatically switch to the secondary DNS, keeping your internet connection alive.
5.  Save and reconnect your devices to refresh the DNS cache.
</template>
</tabs>

Once configured, you can access Olares using both your standard public address and your local address.
:::tip
You can install AdGuard Home from the Olares Market to monitor traffic and manage DNS mappings graphically.
:::
## Method 4: Modify hosts files
If you cannot change router settings and need immediate offline access on a specific computer, you can manually map the domains in your hosts file.

1. Locate your hosts file:
   - **Windows**: `C:\Windows\System32\drivers\etc\hosts`
   - **macOS/Linux**: `/etc/hosts`
2. Open the file with a text editor, which requires Administrator privileges.
3. Add the mapping lines:
    ```plain
    # Replace with the actual internal IP and the username
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
4. Save the file to apply changes and ensure local access without an internet connection.

Verify the changes by checking the URL for quick loading or using the terminal:
```bash
ping desktop.<username>.olares.com
```
If the IP address starts with `192.168`, it indicates successful configuration.


## FAQs

<!--@include: ../../reusables/larepass-vpn.md{50,75}-->

<!--@include: ../../reusables/local-domain.md{42,75}-->