---
outline: [2, 3]
description: Learn how to access Olares One directly from a computer on the same local network.
head:
  - - meta
    - name: keywords
      content: Olares One, local access, host mappings, .local domain, LAN
---

# Access Olares One on your local network

When your computer is on the same local network as Olares One, you can keep traffic on the LAN instead of routing it through the public reverse proxy. This provides faster access and lets you reach your apps when the internet is unavailable.

## Before you begin

- Make sure Olares One and your computer are on the same local network.
- To use LarePass host mappings, install LarePass Desktop on Windows or macOS and import your Olares ID.

## Find the right method

| If this applies to you | What to do |
| --- | --- |
| You want to keep using standard `olares.com` URLs on this computer | [Configure local access with LarePass](#configure-local-access-with-larepass). |
| You use macOS or iOS and do not want to use LarePass | [Use a `.local` URL](#use-a-local-url-without-larepass). |
| You regularly move between local and remote networks | [Use LarePass VPN](./access-olares-via-vpn.md). |

## Configure local access with LarePass

<!--@include: ../reusables/local-domain.md#larepass-local-domains-summary-->

## Use a .local URL without LarePass

On macOS or iOS, local service discovery can resolve multi-level `.local` hostnames without additional configuration. Open a URL such as:

```text
http://desktop.<username>.olares.local
```

On Windows, use LarePass Desktop to configure the multi-level `.local` hostname.

## Troubleshooting

<!--@include: ../reusables/local-domain.md#local-domain-faq-->

## Learn more

- [Access Olares services locally](../manual/best-practices/local-access.md): Compare all local access methods and find FAQs and troubleshooting steps.
