---
outline: [2, 3]
description: 了解如何从同一局域网内的电脑直连 Olares One。
head:
  - - meta
    - name: keywords
      content: Olares One, 本地访问, 本地服务域名, .local 域名, 局域网
---

# 在局域网内访问 Olares One

当电脑与 Olares One 位于同一局域网时，可以让流量留在局域网内，而不经过公网反向代理。这样不仅访问速度更快，而且在互联网不可用时仍能访问应用。

## 开始前准备

- 确保 Olares One 与电脑位于同一局域网。
- 如需使用 LarePass 本地服务域名，请在 Windows 或 macOS 上安装 LarePass 桌面端，并导入 Olares ID。

## 选择合适的方式

| 如果你的情况是 | 建议操作 |
| --- | --- |
| 希望在当前电脑上继续使用标准 `olares.com` 地址 | [使用 LarePass 配置本地访问](#使用-larepass-配置本地访问)。 |
| 使用 macOS 或 iOS，且不想使用 LarePass | [使用 `.local` 地址](#不使用-larepass-时使用-local-地址)。 |
| 经常在本地和远程网络之间切换 | [使用 LarePass 专用网络](./access-olares-via-vpn.md)。 |

## 使用 LarePass 配置本地访问

<!--@include: ../reusables/local-domain.md#larepass-local-domains-summary-->

## 不使用 LarePass 时使用 .local 地址

在 macOS 或 iOS 上，本地服务发现功能无需额外配置即可解析多级 `.local` 主机名。打开类似下面的地址：

```text
http://desktop.<username>.olares.local
```

在 Windows 上，请使用 LarePass 桌面端配置多级 `.local` 主机名。

## 故障排除

<!--@include: ../reusables/local-domain.md#local-domain-faq-->

## 了解更多

- [在局域网内访问 Olares 服务](../manual/best-practices/local-access.md)：比较所有本地访问方式，并查看常见问题和故障排除方法。
