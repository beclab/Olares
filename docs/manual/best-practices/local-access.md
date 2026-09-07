---
outline: [2,3]
description: Compare the methods for accessing Olares directly over your local network.
head:
  - - meta
    - name: keywords
      content: Olares, local access, LarePass VPN, local service domain, local DNS, hosts file, .local domain
---
# Access Olares services locally

Olares services normally use standard `olares.com` URLs that work from both local and remote networks. When your computer or another device is on the same LAN as Olares, you can use a local route instead of sending the connection through the public reverse proxy.

Keeping traffic on the LAN can reduce latency, improve transfer speeds, and preserve access when the internet is unavailable. The right method depends on whether you move between networks, want to keep using the standard URL, need the setting to cover multiple devices, or need an app to have its own LAN IP.

## Choose a local access method

| Requirement | Recommended method | URL | Applies to |
|:------------|:-------------------|:----|:-----------|
| Move between local and remote networks | [LarePass VPN](#use-larepass-vpn) | Standard `olares.com` URL | Current device |
| Use direct LAN access on Windows or macOS | [LarePass local service domains](#configure-local-service-domains-with-larepass) | `olares.com` or `olares.local` | Current computer |
| Use local access without LarePass Desktop | [A `.local` URL](#use-a-local-url-without-larepass) | `olares.local` | Current device |
| Configure local resolution for multiple devices | [Local DNS](#configure-local-dns) | Standard `olares.com` URL | Local network |
| Give a supported app a dedicated LAN IP | [Overlay gateway](#access-an-app-through-overlay-gateway) | App-specific IP address | Local network |

:::warning Avoid a public-network detour
On the same LAN, opening a standard `olares.com` URL without VPN, a matching hosts entry, or local DNS may send the connection through the public reverse proxy. The service can still load, but this route is slower and is not recommended for local access.
:::

## Use LarePass VPN

Use this option if you frequently move between networks. LarePass automatically selects a connection type. When you are on the same LAN as Olares, it switches to **Intranet** for a direct local connection.

<!--@include: ../../reusables/larepass-vpn.md#vpn-setup-notes-->

<!--@include: ../../reusables/larepass-vpn.md#enable-larepass-vpn-->

<!--@include: ../../reusables/larepass-vpn.md#check-vpn-status-->

## Configure local service domains with LarePass

Use this option for direct LAN access from a Windows or macOS computer without running the LarePass VPN.

<!--@include: ../../reusables/local-domain.md#local-domain-overview-->

<!--@include: ../../reusables/local-domain.md#larepass-local-domains-->

## Use a .local URL without LarePass

Use this option when the client device and Olares are on the same LAN and you do not want to configure LarePass Desktop.

### Multi-level domain

<!--@include: ../../reusables/local-domain.md#local-domain-url-format-->

On macOS and iOS, local service discovery can resolve multi-level `.local` hostnames without additional configuration. On Windows, use [LarePass local service domains](#configure-local-service-domains-with-larepass).

### Single-level domain

Single-level `.local` hostnames work across operating systems, but support community apps only. Olares system apps such as Desktop and Files do not support this format.

```text
http://<entrance_id>-<username>-olares.local
```

## Configure local DNS

Use local DNS if you want standard `olares.com` URLs to resolve to the Olares LAN IP for multiple devices. This configuration typically applies to the entire network and does not require LarePass on each client.

:::info
Skip this method if you use `.local` URLs. They use local name resolution and do not depend on the `olares.com` DNS records.
:::

### Find the Olares LAN IP

<tabs>
<template #Use-LarePass-mobile>

1. Make sure your phone and Olares are on the same network.
2. Open LarePass and go to **Settings** > **System**.

   ![Open System settings in LarePass](/images/manual/get-started/larepass-system.png#bordered)
3. Tap the Olares device card.

   ![Open the Olares device card](/images/manual/get-started/larepass-device-card.png#bordered)
4. In **Network**, find the **Intranet IP**.

   ![Find the Intranet IP](/images/manual/get-started/larepass-network.png#bordered)

</template>
<template #Use-Control-Hub>

1. In Control Hub, open **Terminal** and select **Olares**.

   ![Open the Olares terminal in Control Hub](/images/manual/get-started/find-internal-ip-from-controlhub.png#bordered)
2. Run `ifconfig`.
3. Find the `inet` address for the active wired or Wi-Fi interface. It is typically in a private range such as `192.168.x.x`.

</template>
</tabs>

### Configure the DNS server

Configure the DNS server on one computer or on your router.

<tabs>
<template #One-computer>

The exact steps depend on your operating system. For example, on macOS:

1. Open the Apple menu and go to **System Settings**.
2. Select **Wi-Fi**, then click **Details** for the connected network.
3. Select **DNS**.
4. Add the Olares LAN IP under **DNS Servers** and move it to the top of the list.
5. Keep the existing DNS server, or add a public resolver such as `1.1.1.1`, below it as a fallback.
6. Click **OK**.

</template>
<template #All-devices>

1. Sign in to your router's administration page.
2. Open its DHCP or DNS settings.
3. Set **Primary DNS** to the Olares LAN IP.
4. Keep the existing primary DNS server, or a public resolver such as `1.1.1.1`, as **Secondary DNS**.
5. Save the settings and reconnect client devices so they receive the updated DNS configuration.

</template>
</tabs>

Once configured, open the standard `olares.com` URLs as usual. They should resolve to the Olares LAN IP while you are on the same network.

:::tip
You can install AdGuard Home from the Olares Market to monitor traffic and manage DNS mappings graphically.
:::

:::info Verify hostname resolution
After configuring LarePass local service domains or local DNS, resolve a standard service hostname from the client computer:

```bash
ping desktop.<username>.olares.com
```

The returned address should match the LAN IP of your Olares. Private LAN addresses commonly begin with `192.168`, `10`, or `172.16`–`172.31`.
:::

## Access an app through overlay gateway

Some apps need to appear as independent devices on your LAN for device discovery, casting, multiplayer connections, or other protocols that do not use an Olares web URL. For supported apps, overlay gateway assigns the app a dedicated LAN IP through a virtual network interface.

This method is separate from LarePass VPN, local service domains, and local DNS. Use the IP address shown for the app rather than an `olares.com` or `olares.local` URL.

Overlay gateway requires Olares to run on a native Linux host with a wired Ethernet connection. For availability, permissions, and setup steps, see [Manage overlay gateway for applications](../olares/settings/overlay-gateway.md).

## Troubleshooting

### A LarePass-managed URL no longer opens locally

The LAN IP of Olares may have changed. Make sure the computer and Olares are on the same LAN, turn off **VPN connection**, then apply the hosts update offered by LarePass.

### LarePass cannot add or update the hosts entries

Make sure LarePass has permission to update the system hosts file. Do not edit LarePass-managed entries manually unless instructed by Olares Support.

## FAQs

<!--@include: ../../reusables/larepass-vpn.md#larepass-vpn-faq-->

<!--@include: ../../reusables/local-domain.md#larepass-local-domain-faq-->

<!--@include: ../../reusables/local-domain.md#local-domain-faq-->
