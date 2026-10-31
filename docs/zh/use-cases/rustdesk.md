---
outline: [2, 3]
description: 在 Olares 上运行 RustDesk Server，通过 LarePass 专用网络从笔记本远程操作家里的电脑，或在同一局域网内通过 Overlay 连接。
head:
  - - meta
    - name: keywords
      content: Olares, RustDesk Server, 远程桌面, 自托管, ID 服务器, 中继, LarePass VPN, Overlay 网关
app_version: "0.0.17"
doc_version: "1.0"
doc_updated: "2026-10-31"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/rustdesk.md)为准。
:::

# 使用 RustDesk 远程操作家里的电脑

借助 RustDesk 和 LarePass 专用网络，你可以在笔记本上远程操作家里的电脑。RustDesk Server 运行在 Olares 上，帮助两台电脑建立连接，并在无法直连时转发远程桌面数据。

## 学习目标

本指南将介绍如何：

- 安装 RustDesk Server，查找服务器地址和公钥。
- 配置两台电脑，通过 LarePass 专用网络或内网 Overlay 连接服务端。
- 连接并远程操作家里的电脑。

## 前提条件

- Olares 1.12.7 或更高版本。
- 两台电脑都已安装 [RustDesk 客户端](https://rustdesk.com/download)。
- 两台电脑都已安装 LarePass，并登录到同一台 Olares 设备。参阅 [LarePass 专用网络访问](../manual/larepass/private-network.md)。

## 安装 RustDesk Server

1. 打开应用市场，搜索“RustDesk Server”。
2. 点击**获取**，然后点击**安装**，等待安装完成。

## 获取服务器配置

1. 从启动台打开 RustDesk Server。应用会显示连接配置页，列出服务器地址和公钥。接下来需要将这些值填入两台电脑的 RustDesk 客户端。
2. 选择 **VPN 连接**。

   ![RustDesk Server 页面中的 VPN 服务器地址和公钥](/images/manual/use-cases/rustdesk-vpn-settings.png#bordered)

配置客户端时，从自己的应用页面复制地址和 Key。

点击右上角的地球图标可以切换页面语言，页面会记住选择。

## 为两台电脑配置 LarePass 专用网络连接

以下步骤使用 LarePass 专用网络连接两台电脑。如果两台电脑都与 Olares 位于同一网络，也可以[通过内网 Overlay 连接](#通过内网-overlay-连接)。

两台电脑应使用同一种连接方式。如果一台电脑的 RustDesk 客户端填写 VPN 域名地址，另一台填写 Overlay IP，而其中一台无法访问本次连接使用的中继地址，连接就可能失败。

1. 在两台电脑上打开 LarePass，从头像菜单开启**专用网络连接**，并在远程控制期间保持开启。不同平台的设置请参阅[在 LarePass 中启用专用网络](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。
2. 在每台电脑的 RustDesk 客户端中打开**设置** > **网络** > **ID/中继服务器**。如果提示解锁设置，先完成解锁。
3. 填写以下字段，然后在两台电脑上保存设置。

   | 客户端字段 | 填写内容 |
   | --- | --- |
   | **ID 服务器** | 从应用页面复制完整地址，包括 `:31116` 端口 |
   | **中继服务器** | 从应用页面复制完整地址，包括 `:31117` 端口 |
   | **API 服务器** | 留空。RustDesk Server OSS 不提供 Pro 管理 API |
   | **Key** | 从应用页面复制公钥 |

   Key 用于识别 RustDesk 服务端。家里电脑的 RustDesk 客户端中显示的访问密码，用于授权远程控制这台电脑。

## 开始远程控制

:::warning 远程控制 Mac
如果要控制的电脑是 Mac，请先为 RustDesk 授予屏幕录制和辅助功能权限，以便共享画面并控制键盘和鼠标。操作步骤参阅 [RustDesk macOS 设置指南](https://rustdesk.com/docs/en/client/mac/)。
:::

1. 在家里的电脑上保持 RustDesk 运行，并确认客户端服务已启用。
2. 在笔记本上输入家里电脑的 RustDesk ID，发起连接。
3. 在家里的电脑上接受连接请求，或输入其 RustDesk 客户端显示或配置的访问密码。
4. 家里电脑的桌面出现后，用笔记本的键盘和鼠标操作它。使用结束后，在 RustDesk 中断开连接。

## 通过内网 Overlay 连接

如果两台电脑都与 Olares 位于同一网络，可以使用 Overlay 连接，无需开启 LarePass 专用网络。Overlay 会为 RustDesk Server 分配独立的内网 IP。

:::warning Overlay 使用条件
Olares 必须运行在原生 Linux 主机上，并使用有线以太网连接。超级管理员需要先启用系统级 Overlay 网关，之后才能为 RustDesk Server 启用。可用条件和权限说明参阅[管理 Overlay 网关](../manual/olares/settings/overlay-gateway.md)。
:::

1. 在 Olares 中打开**设置** > **网络** > **Overlay 网关**。如果系统级网关尚未启用，请以超级管理员身份启用，或联系超级管理员操作。
2. 在**应用**列表中启用 RustDesk Server，并确认重启。等待应用恢复为**运行中**，从此设置页复制分配的内网 IP。
3. 从启动台打开 RustDesk Server，选择**内网 Overlay**。填写客户端配置时，用从设置页复制的 IP 替换 `Overlay_IP` 占位符。

   ![RustDesk Server 页面中的内网 Overlay 地址格式](/images/manual/use-cases/rustdesk-overlay-settings.png#bordered)

4. 在两台电脑的 RustDesk 客户端中打开**设置** > **网络** > **ID/中继服务器**，填写以下字段。

   | 客户端字段 | 填写内容 |
   | --- | --- |
   | **ID 服务器** | 分配的 Overlay IP，加上 `:21116` |
   | **中继服务器** | 同一个 Overlay IP，加上 `:21117` |
   | **API 服务器** | 留空 |
   | **Key** | 从 RustDesk Server 页面复制公钥 |

5. 在两台电脑上保存设置，然后[开始远程控制](#开始远程控制)。

如果应用的 Overlay 地址变化，请从设置页复制当前地址，并更新两台电脑的客户端。

## 常见问题

### 为什么无法连接到中继服务器？

建立连接时，RustDesk 可能会将家里电脑配置的中继地址发送给笔记本，两台电脑都需要能访问这个地址。如果笔记本无法访问，即使两台电脑都能连接 ID 服务器，仍可能提示“无法连接到中继服务器：请稍后再试”。

将两台电脑设为同一种连接方式，然后按以下步骤检查：

1. 检查两台电脑的**中继服务器**设置。使用 VPN 时，填写应用页面提供的完整域名地址，并保持 LarePass 专用网络连接开启。使用 Overlay 时，填写设置页显示的当前内网 IP，端口为 `21117`。
2. 在 RustDesk 网络设置中，将**使用 WebSocket**、**允许回退到不安全的 TLS 连接**和**禁用 UDP**保持关闭。
3. 保存修改，重新连接。

### 将网页入口设为 Public 有什么作用？

设为 Public 后，无需登录就能打开连接配置页。要从公网访问 RustDesk 的 TCP/UDP 端口，还需要单独配置网络。按本指南连接时，通过 LarePass 专用网络使用应用域名，或在同一局域网内使用 Overlay IP。

## 了解更多

- [LarePass 专用网络访问](../manual/larepass/private-network.md)：将电脑连接到 Olares 专用网络。
- [管理 Overlay 网关](../manual/olares/settings/overlay-gateway.md)：启用应用的内网地址并查看可用条件。
- [RustDesk Server OSS 文档](https://rustdesk.com/docs/en/self-host/rustdesk-server-oss/)：了解 ID 和中继服务。
