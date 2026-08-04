---
outline: [2, 3]
description: 了解如何在同一网络中使用 `.local` 域名访问 Olares 服务。
head:
  - - meta
    - name: keywords
      content: Olares, .local 域名, 本地访问
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../one/access-olares-via-local-domain.md)为准。
:::

# 通过 .local 域名访问 Olares

当你的电脑或手机与 Olares One 处于同一本地网络时，可以使用 `.local` 域名访问 Olares 服务，这样流量将保持在局域网内。

## 前提条件

**硬件**
- Olares One 已完成设置并连接到网络。
- 一台与 Olares One 处于同一网络的客户端设备（电脑或手机）。

**LarePass**（Windows 必需）
- Windows 设备上已安装 LarePass 桌面客户端。
- 已在 LarePass 桌面客户端中导入你的 Olares ID。

## URL 格式

<!--@include: ../reusables/local-domain.md#local-domain-url-format-->

## macOS

无需额外设置。在浏览器中使用本地 URL 即可（例如 `http://desktop.<username>.olares.local`）。

## Windows

<!--@include: ../reusables/local-domain.md#windows-local-domain-->

## 故障排除

<!--@include: ../reusables/local-domain.md#local-domain-faq-->

## 了解更多
- [在本地访问 Olares 服务](../manual/best-practices/local-access.md)：DNS 配置、hosts 文件及其他本地访问方式。
