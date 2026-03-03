---
outline: [2,3]
description: 了解如何使用 LarePass VPN 或 .local 域名安全访问 Olares 服务。
---
# 安全访问 Olares 服务

通常，你会通过浏览器使用类似 `https://desktop.<username>.olares.com` 的网址访问 Olares，从而在任何设备上随时访问服务。你可以从家庭网络或外网安全访问 Olares。

- [使用 LarePass VPN](#使用-larepass-vpn)：无论在内网还是外网均可使用。
- [使用 .local 域名](#使用-local-域名)：仅当客户端设备与 Olares 在同一局域网时使用。

## 使用 LarePass VPN

建议启用 LarePass VPN，以获得更安全、高效的连接。客户端会自动检测网络环境并选择最佳连接方式：

- **居家**：建立直连内网，加快局域网文件传输。
- **远程**：切换至加密隧道，保障远程访问安全。

<!--@include: ../../reusables/larepass-vpn.md{19,24}-->

在用来访问 Olares 的当前设备上直接启用 LarePass VPN。

<!--@include: ../../reusables/larepass-vpn.md{26,50}-->

## 使用 .local 域名

当设备与 Olares 在同一局域网时，可使用 .local 域名。

### URL 格式

<!--@include: ../../reusables/local-domain.md{7,23}-->

### macOS

无需配置。在浏览器中直接使用本地 URL（例如 `http://desktop.<username>.olares.local`）。

### Windows
<!--@include: ../../reusables/local-domain.md{26,40}-->

### 常见问题

<!--@include: ../../reusables/larepass-vpn.md{50,75}-->

<!--@include: ../../reusables/local-domain.md{42,75}-->

## 了解更多
- [本地访问 Olares 服务](../best-practices/local-access.md)：查看所有本地网络连接方式的详细说明。
- [网络](../../developer/concepts/network.md)：了解 Olares 中的各类应用入口。
