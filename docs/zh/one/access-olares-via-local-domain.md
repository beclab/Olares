---
outline: [2, 3]
description: Learn how to access your Olares services on the same network using the `.local` domain.
head:
  - - meta
    - name: keywords
      content: Olares, .local domain, local access
---

# Access Olares via .local domain

When your computer or phone is on the same local network as Olares One, you can use a `.local` domain to reach your Olares services so traffic stays on your LAN.

## Prerequisites

**Hardware**
- Olares One is set up and connected to your network.
- A client device (computer or phone) on the same network as Olares One.

**LarePass** (Required for Windows)
- The LarePass desktop client is installed on your Windows device.
- You have imported your Olares ID on the LarePass desktop client.

## URL format

<!--@include: ../../reusables/local-domain.md{10,23}-->

## macOS

No setup is needed. Use the local URL in your browser (for example, `http://desktop.<username>.olares.local`).

## Windows

<!--@include: ../../reusables/local-domain.md{26,40}-->

## Troubleshooting

<!--@include: ../../reusables/local-domain.md{42,75}-->

## Learn more
- [Access Olares services locally](../manual/best-practices/local-access.md): DNS configuration, hosts file, and other local access methods.
