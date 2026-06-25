---
outline: [2,3]
description: 了解如何使用 LarePass VPN 安全地访问 Olares 服务。
head:
  - - meta
    - name: keywords
      content: Olares, LarePass VPN, 本地访问
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/access-olares-via-vpn.md)。
:::

# 使用 LarePass VPN 安全访问 Olares 服务

通常情况下，你可以通过浏览器使用 `https://desktop.<username>.olares.com` 这样的 URL 访问 Olares 服务。这样你可以随时随地从任何设备访问你的服务。

虽然这个地址在任何地方都能使用，但建议启用 LarePass VPN 以确保连接始终安全高效。客户端会自动检测你的网络环境并选择最佳连接方式：
- **在家中**：建立直接的**内网**连接，以便在本地网络上实现更快的文件传输。
- **远程访问**：切换到安全加密隧道，确保远程访问时保持安全连接。

## 前提条件
**硬件** <br>
- Olares One 已完成设置并可访问。
- 一台可联网的客户端设备（电脑或手机）。

## 步骤 1：下载 LarePass

<!--@include: ../reusables/larepass-vpn.md#install-larepass-client-->

## 步骤 2：启用 LarePass VPN

安装完成后，直接在设备上启用 VPN。

<!--@include: ../reusables/larepass-vpn.md#vpn-setup-notes-->

<!--@include: ../reusables/larepass-vpn.md#enable-larepass-vpn-->

## 步骤 3：验证连接类型

<!--@include: ../reusables/larepass-vpn.md#check-vpn-status-->

## 故障排除

<!--@include: ../reusables/larepass-vpn.md#larepass-vpn-faq-->
