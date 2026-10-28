---
outline: [2, 3]
description: 在 Olares 上托管 RustDesk ID 和中继服务，通过 LarePass VPN 或内网 Overlay 配置两台电脑，并建立远程桌面连接。
head:
  - - meta
    - name: keywords
      content: Olares, RustDesk Server, 远程桌面, 自托管, ID 服务器, 中继, LarePass VPN, Overlay 网关
app_version: "0.0.17"
doc_version: "1.0"
doc_updated: "2026-10-28"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/rustdesk.md)为准。
:::

# 使用 Olares 上的 RustDesk Server 连接电脑

RustDesk Server 帮助 RustDesk 客户端发现对方，并在无法直连时中继桌面数据。在 Olares 上运行服务端，在两台电脑上分别安装 RustDesk 客户端，就可以通过 LarePass VPN 从笔记本访问家里的电脑。Olares 应用打开的是连接配置页，远程桌面会话在 RustDesk 客户端中进行。

## 学习目标

本指南将介绍如何：

- 安装 RustDesk Server 并查找客户端连接配置。
- 通过 LarePass VPN 或内网 Overlay 配置两台电脑。
- 发起远程桌面会话，并识别中继地址不匹配的问题。

## 前提条件

- Olares 1.12.7 或更高版本。
- 两台电脑都已安装 [RustDesk 客户端](https://rustdesk.com/download)。
- VPN 示例中的两台电脑都已安装 LarePass，并登录到同一台 Olares 设备。参阅 [LarePass 专用网络访问](../manual/larepass/private-network.md)。

## 安装 RustDesk Server

1. 打开应用市场，搜索“RustDesk Server”。
2. 点击**获取**，然后点击**安装**，等待安装完成。
3. 从启动台打开 RustDesk Server。页面会显示服务器地址和自动生成的公钥。

点击右上角的地球图标可以切换页面语言，页面会记住选择。

## 配置两台电脑

首次连接建议两端使用同一种连接方式。VPN 和内网 Overlay 使用的服务器地址不同，Key 相同。

:::warning 两端都需要能访问协商使用的中继地址
RustDesk 可能将被控端配置的中继地址传给控制端。如果一端使用只能通过 VPN 访问的域名，另一端使用 VPN 电脑无法访问的 Overlay IP，即使两端都能连接 ID 服务器，远程连接仍可能失败。两端使用相同的连接方式可以避免这种地址不匹配。
:::

### 通过 LarePass VPN 连接

1. 在两台电脑上打开 LarePass，从头像菜单开启**专用网络连接**。不同平台的操作请参阅[在 LarePass 中启用专用网络](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。
2. 在 RustDesk Server 页面选择 **VPN 连接**。
3. 在每台电脑的 RustDesk 客户端中打开**设置** > **网络** > **ID/中继服务器**。如果客户端提示解锁设置，先完成解锁。
4. 填写下面四个字段，然后确认修改。

   | 客户端字段 | 填写内容 |
   | --- | --- |
   | **ID 服务器** | 从页面复制完整地址，包括 `:31116` 端口 |
   | **中继服务器** | 从页面复制完整地址，包括 `:31117` 端口 |
   | **API 服务器** | 留空 |
   | **Key** | 从页面复制公钥 |

![RustDesk Server 的英文 VPN 连接配置](/images/manual/use-cases/rustdesk-vpn-settings.png#bordered)

请复制自己实例中的地址和 Key，不要照填截图里的示例值。会话期间保持 LarePass 专用网络连接开启。

### 通过内网 Overlay 连接

两台电脑都与 Olares 位于同一网络时，可以使用这种方式。Overlay 要求 Olares 运行在原生 Linux 主机上，并使用有线以太网连接。可用条件和权限要求请参阅[管理 Overlay 网关](../manual/olares/settings/overlay-gateway.md)。

1. 在 Olares 中打开**设置** > **网络** > **Overlay 网关**。如果系统级网关尚未启用，先启用它。
2. 在**应用**列表中启用 RustDesk Server，并确认重启。等待应用恢复为**运行中**，从此设置页复制分配的内网 IP。
3. 打开 RustDesk Server 页面，选择**内网 Overlay**。`Overlay_IP` 是填写格式中的占位符，不是本页面自动获取的地址。
4. 在两台电脑的 RustDesk 客户端中打开**设置** > **网络** > **ID/中继服务器**，填写以下字段。

   | 客户端字段 | 填写内容 |
   | --- | --- |
   | **ID 服务器** | 设置页分配的 Overlay IP，加上 `:21116` |
   | **中继服务器** | 同一个 Overlay IP，加上 `:21117` |
   | **API 服务器** | 留空 |
   | **Key** | 从 RustDesk Server 页面复制公钥 |

![RustDesk Server 的英文内网 Overlay 配置说明](/images/manual/use-cases/rustdesk-overlay-settings.png#bordered)

在两端确认配置。如果应用的 Overlay 地址变化，请从设置页复制当前地址，并更新两台电脑的客户端。

## 发起远程桌面会话

1. 保持家里的电脑运行 RustDesk，并确认其客户端服务已启用。
2. 在笔记本上输入家里电脑的 RustDesk ID，发起连接。
3. 在家里的电脑上批准请求，或输入其 RustDesk 客户端显示或配置的访问密码。
4. 远程桌面出现后，即可从笔记本操作家里的电脑。使用结束后，在 RustDesk 中断开会话。

API 服务器留空，因为 RustDesk Server OSS 不提供 Pro 管理 API。这不会影响远程桌面连接。

服务端公钥 Key 用于识别服务器，不是被控电脑的访问密码。通过中继连接时，两台电脑无需直接互通，但两端都必须能访问本次会话协商使用的中继地址。

## 处理中继连接错误

如果 RustDesk 提示“无法连接到中继服务器：请稍后再试”，先检查两端选择的连接方式和中继地址。

使用 VPN 示例时，两台电脑都应保持 LarePass 专用网络连接开启，并填写应用页面提供的域名地址。使用内网 Overlay 时，两端填写当前 Overlay IP。此配置下，**使用 WebSocket**、**允许回退到不安全的 TLS 连接**和**禁用 UDP**都保持关闭。

把 Olares 网页入口设为 Public 只会取消连接配置页的登录要求，不会将 RustDesk TCP/UDP 端口发布到公网，也不会让 VPN 端自动访问内网 Overlay IP。公网端口转发是单独的网络配置。

## 了解更多

- [LarePass 专用网络访问](../manual/larepass/private-network.md)：将电脑连接到 Olares 专用网络。
- [管理 Overlay 网关](../manual/olares/settings/overlay-gateway.md)：启用应用的内网地址并查看可用条件。
- [RustDesk Server OSS 文档](https://rustdesk.com/docs/en/self-host/rustdesk-server-oss/)：了解 ID 和中继服务。
