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

<!--@include: ../../reusables/larepass-vpn.md{19,24}-->

Enable the LarePass VPN directly on the device you are currently using to access Olares.

<!--@include: ../../reusables/larepass-vpn.md{26,50}-->

## Using the .local domain

Use the `.local` domain when your device and Olares are on the same LAN. 

### URL format

<!--@include: ../../reusables/local-domain.md{7,23}-->

### macOS

No setup is needed. Use the local URL in your browser (for example, `http://desktop.<username>.olares.local`).

### Windows
<!--@include: ../../reusables/local-domain.md{26,40}-->

### FAQs

<!--@include: ../../reusables/larepass-vpn.md{50,75}-->

<!--@include: ../../reusables/local-domain.md{42,75}-->

## Learn more
- [Access Olares locally](../best-practices/local-access.md): Explore detailed instructions for all available local network connection methods.
- [Network](../../developer/concepts/network.md): Learn about the different entry points in Olares.
