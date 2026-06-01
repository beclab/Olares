---
outline: [2,3]
description: Learn how to access your Olares services securely using LarePass VPN or the .local domain.
---
# Access Olares services securely

Typically, you access Olares services through a browser using a URL like `https://desktop.<username>.olares.com`. This way, you can reach your services from any device at any time. You can access Olares securely from your home network or from elsewhere. 

- [Using LarePass VPN](#using-larepass-vpn): Use this whether you are on your home network or away.
- [Using the .local domain](#using-the-local-domain): Use this only when your client device and Olares are on the same LAN.

## Using LarePass VPN

It is recommended to enable the LarePass VPN to ensure your connection is always secure and efficient. The client automatically detects your network environment and selects the best connection method:

- **At home**: It establishes a direct Intranet connection to allow faster file transfers on your local network.
- **From remote**: It switches to a secure encrypted tunnel so you remain connected safely when accessing remotely.

<!--@include: ../../reusables/larepass-vpn.md#vpn-setup-notes-->

Enable the LarePass VPN directly on the device you are currently using to access Olares.

<!--@include: ../../reusables/larepass-vpn.md#enable-larepass-vpn-->

<!--@include: ../../reusables/larepass-vpn.md#check-vpn-status-->

## Using the .local domain

Use the `.local` domain when your device and Olares are on the same LAN. 

### URL format

<!--@include: ../../reusables/local-domain.md#local-domain-overview-->

### macOS

No setup is needed. Use the local URL in your browser (for example, `http://desktop.<username>.olares.local`).

### Windows
<!--@include: ../../reusables/local-domain.md#windows-local-domain-->

## FAQs

<!--@include: ../../reusables/larepass-vpn.md#larepass-vpn-faq-->

<!--@include: ../../reusables/local-domain.md#local-domain-faq-->

## Learn more
- [Access Olares locally](../best-practices/local-access.md): Explore detailed instructions for all available local network connection methods.
- [Network](../../developer/concepts/network.md): Learn about the different entry points in Olares.
