---
outline: [2,3]
title: Access Olares One with LarePass VPN
description: Access Olares One securely with LarePass VPN and verify the connection type on local or remote networks.
head:
  - - meta
    - name: keywords
      content: Olares, LarePass VPN, local access
---

# Access Olares One services securely using LarePass VPN

Typically, you access Olares services through a browser using a URL like `https://desktop.<username>.olares.com`. This way, you can reach your services from any device at any time.

While this address works from anywhere, it's recommended to enable the LarePass VPN to ensure your connection is always secure and efficient. The client automatically detects your network environment and selects the best connection method:
- **At home**: It establishes a direct **Intranet** connection to allow faster file transfers on your local network.
- **From remote**: It switches to a secure encrypted tunnel to ensure you remain connected safely when accessing remotely.

## Prerequisites
**Hardware** <br>
- Your Olares One is set up and accessible.
- A client device (computer or mobile phone) with internet access.

## Step 1: Download LarePass

<!--@include: ../reusables/larepass-vpn.md#install-larepass-client-->

## Step 2: Enable LarePass VPN

Once installed, enable the VPN directly on the device.

<!--@include: ../reusables/larepass-vpn.md#vpn-setup-notes-->

<!--@include: ../reusables/larepass-vpn.md#enable-larepass-vpn-->

## Step 3: Verify the connection type

<!--@include: ../reusables/larepass-vpn.md#check-vpn-status-->

## Troubleshooting

<!--@include: ../reusables/larepass-vpn.md#larepass-vpn-faq-->
