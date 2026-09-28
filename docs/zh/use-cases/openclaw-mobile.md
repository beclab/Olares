---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/openclaw-mobile
outline: [2, 3]
title: 使用 Openclaw 移动客户端连接
description: 通过 LarePass VPN，使用二维码或手动连接设置，将 OpenClaw Android 或 iOS 应用与 Olares Gateway 配对。
app_version: "1.0.46"
doc_version: "1.0"
doc_updated: "2026-09-28"
---

# 使用 Openclaw 移动客户端连接

<VersionRouteSelect />

使用 OpenClaw 手机应用与 Olares 上运行的智能体聊天。完成一次配对后，在家中或外出时都可保持 LarePass VPN 开启，使用同一个 Gateway 地址连接。

## 前提条件

- 在 Olares 上完成 [OpenClaw 初始化](openclaw.md)，并配置聊天所需的模型。
- 在手机上安装 LarePass，登录拥有此 OpenClaw 实例的 Olares 账户，并开启 VPN。参见[安全访问 Olares 服务](../manual/get-started/local-access.md)。
- 能够打开 Olares 的 **OpenClaw CLI**，用于生成配对码和审批设备。

:::important 在手机上保持 LarePass VPN 开启
**OpenClaw Gateway** 入口使用 **Internal（内部）** 访问策略。通过 LarePass VPN 访问时可跳过 Olares 登录；仅连接同一个 Wi-Fi 并不会获得这一豁免。

无论使用本地 Wi-Fi 还是移动数据，都应保持 VPN 开启。在局域网内，LarePass 可以使用 Intranet 连接。OpenClaw 的令牌或配对码不能替代 Olares 入口认证。
:::

## 确认 Gateway 地址和 Auth Level

1. 在 Olares 中打开**设置** > **应用** > **OpenClaw**。
2. 在**入口**中打开 **OpenClaw Gateway**，复制它的域名，并保持认证级别为 **Internal（内部）**。

Gateway 入口不会显示在启动台中。它的地址与 **OpenClaw CLI**、**Control UI** 和 Olares 桌面地址不同。如果使用克隆应用，请选择对应实例的 Gateway 入口。

下文用 `gateway.example.com` 表示此域名，请替换为实际值。完整连接地址为 `wss://gateway.example.com`，使用 **443** 端口和 **TLS**。通过 Olares 入口连接时，不要填写内部服务端口 `18789`。

## 安装并连接客户端

<Tabs>
<template #iOS>

### 安装 iOS 客户端

1. 从 [App Store](https://apps.apple.com/app/openclaw-ai-that-does-things/id6780396132) 安装官方 OpenClaw 客户端。
2. 打开 LarePass 并确认 VPN 已连接，再打开 OpenClaw。选择以下任意一种配对方式。

### 路径一：扫码自动配对

1. 从 Olares 启动台打开 **OpenClaw CLI**。
2. 生成手机配对码：

   ```bash
   openclaw qr
   ```

   全新安装的 Olares 应用版本 **1.0.46 及以上** 会将 Gateway 入口设为配对地址。检查输出中的 **Gateway** 一行，确认它使用刚才复制的域名，并以 `wss://` 开头。

3. 如果从旧版升级、修改过入口域名，或显示的地址不正确，请显式指定地址：

   ```bash
   openclaw qr --url 'wss://gateway.example.com'
   ```

   此命令不会修改已保存的配置。升级会保留原有 `openclaw.json`，不会用新的默认值覆盖已有安装。

4. 在 iPhone 上打开 OpenClaw，进入首次启动的 Gateway 设置，或打开 **Settings** > **Gateway** 添加 Gateway。
5. 选择二维码扫描选项，按提示允许相机访问，然后扫描终端中的二维码。
6. 核对客户端显示的 Gateway 地址，确认与 Olares 的 Gateway 入口一致，然后连接。配对码元数据匹配时，官方客户端可自动完成配对；如果仍提示等待审批，按[审批待配对设备](#审批待配对设备)操作。
7. 进入聊天界面，发送一条消息并确认收到回复。

二维码包含短期有效的配对凭据，请妥善保管，过期后重新生成。默认的 `wss://` 配对码授予手机节点访问权限和完整 Gateway 操作员权限；需要较低权限时，可在生成命令中添加 `--limited`。参见[上游 QR 命令说明](https://docs.openclaw.ai/cli/qr)。

### 路径二：手动配对

1. 在 **OpenClaw CLI** 中显示 Gateway 令牌：

   ```bash
   openclaw gateway auth-token --show
   ```

   将令牌私下复制到手机。它是 OpenClaw 的访问凭据，不是 Olares 账户密码。


2. 在 iPhone 上打开 **Settings** > **Gateway**，启用 **Use Manual Gateway**，部分版本称为 **Manual Host**。
3. 填写连接参数：

   | 参数 | 填写内容 |
   | --- | --- |
   | 主机 | Gateway 入口域名，例如 `gateway.example.com`。主机和端口分开填写时，不要添加协议前缀或路径。 |
   | 端口 | `443` |
   | 令牌 | 第 1 步获取的 Gateway 令牌 |
   | 密码 | Olares 默认使用令牌认证，此处留空 |
   | 连接安全性 / TLS | 选择**安全（TLS）**或启用 TLS |

   如果客户端只有一个完整 URL 输入框，填写 `wss://gateway.example.com`。如果已将 OpenClaw 改为密码认证，则填写配置的 Gateway 密码。

4. 点击**测试连接**、**连接**或**保存并连接**。出现 **pairing required** 表示已到达 Gateway，需要审批设备。
5. 按[审批待配对设备](#审批待配对设备)操作，然后返回客户端重新连接。
6. 进入聊天界面，发送一条消息并确认收到回复。

</template>

<template #Android>

### 安装 Android 客户端

1. 从 [Google Play](https://play.google.com/store/apps/details?id=ai.openclaw.app) 安装官方 OpenClaw 客户端。也可按照[官方安装指南](https://docs.openclaw.ai/platforms/android#install-outside-google-play)，下载并校验发布页面提供的签名 APK；并非每个 Gateway 版本都提供 APK。
2. 打开 LarePass 并确认 VPN 已连接，再打开 OpenClaw。选择以下任意一种配对方式。

### 路径一：扫码自动配对

1. 从 Olares 启动台打开 **OpenClaw CLI**。
2. 生成手机配对码：

   ```bash
   openclaw qr
   ```

   全新安装的 Olares 应用版本 **1.0.46 及以上** 会将 Gateway 入口设为配对地址。检查输出中的 **Gateway** 一行，确认它使用刚才复制的域名，并以 `wss://` 开头。

3. 如果从旧版升级、修改过入口域名，或显示的地址不正确，请显式指定地址：

   ```bash
   openclaw qr --url 'wss://gateway.example.com'
   ```

   此命令不会修改已保存的配置。升级会保留原有 `openclaw.json`，不会用新的默认值覆盖已有安装。

4. 在 Android 手机上打开 OpenClaw，首次设置时选择**扫描二维码或配对码**；如果已经完成设置，进入**设置** > **Gateway** > **添加 Gateway**。
5. 打开二维码扫描，按提示允许相机访问，然后扫描终端中的二维码。
6. 核对客户端显示的 Gateway 地址，确认与 Olares 的 Gateway 入口一致，然后连接。配对码元数据匹配时，官方客户端可自动完成配对；如果仍提示等待审批，按[审批待配对设备](#审批待配对设备)操作。
7. 进入聊天界面，发送一条消息并确认收到回复。

二维码包含短期有效的配对凭据，请妥善保管，过期后重新生成。默认的 `wss://` 配对码授予手机节点访问权限和完整 Gateway 操作员权限；需要较低权限时，可在生成命令中添加 `--limited`。参见[上游 QR 命令说明](https://docs.openclaw.ai/cli/qr)。

### 路径二：手动配对

1. 在 **OpenClaw CLI** 中显示 Gateway 令牌：

   ```bash
   openclaw gateway auth-token --show
   ```

   将令牌私下复制到手机。它是 OpenClaw 的访问凭据，不是 Olares 账户密码。


2. 在 Android 客户端选择**手动设置**，或进入**设置** > **Gateway** > **手动 Gateway**。
3. 填写连接参数：

   | 参数 | 填写内容 |
   | --- | --- |
   | 主机 | Gateway 入口域名，例如 `gateway.example.com`。主机和端口分开填写时，不要添加协议前缀或路径。 |
   | 端口 | `443` |
   | 令牌 | 第 1 步获取的 Gateway 令牌 |
   | 密码 | Olares 默认使用令牌认证，此处留空 |
   | 连接安全性 / TLS | 选择**安全（TLS）**或启用 TLS |

   如果客户端只有一个完整 URL 输入框，填写 `wss://gateway.example.com`。如果已将 OpenClaw 改为密码认证，则填写配置的 Gateway 密码。

4. 点击**测试连接**、**连接**或**保存并连接**。出现 **pairing required** 表示已到达 Gateway，需要审批设备。
5. 按[审批待配对设备](#审批待配对设备)操作，然后返回客户端重新连接。
6. 进入聊天界面，发送一条消息并确认收到回复。

</template>
</Tabs>

## 审批待配对设备

1. 保持手机停留在连接界面，并开启 LarePass VPN。在 **OpenClaw CLI** 中运行：

   ```bash
   openclaw devices list
   ```

2. 在 **Pending** 表中，通过设备名称和设备 ID 确认自己的手机，核对申请的角色与权限。使用此列表中的当前 **Request** ID：

   ```bash
   openclaw devices approve <requestId>
   ```

3. 返回手机，重试连接或确认已完成审批。
4. 再次运行 `openclaw devices list`。手机应用同时使用 `node` 和 `operator` 连接；如果自己的手机出现另一条请求，请核对后审批该当前请求。

:::tip Request ID 可能变化
如果重试时申请的角色、权限范围或公钥发生变化，Gateway 可能用新的 Request ID 替换待审批请求。手机仍可能保留之前的错误信息和 ID。

遇到 `unknown requestId` 或手机与 CLI 的 ID 不一致时，请刷新 `openclaw devices list`，使用已核实设备的当前请求。不要反复审批旧 ID，也不要审批无关设备。如果手机已经在 **Paired** 列表中，且没有待审批请求，重新连接即可，无需再次审批。
:::

## 验证本地和外网访问

1. 将 OpenClaw 保持在手机前台，并检查连接：

   ```bash
   openclaw devices list
   openclaw nodes status
   ```

   手机应显示为已配对，应用处于前台时节点应处于已连接状态。

2. 发送一条简短聊天消息并确认收到回复。配对成功本身并不能证明模型配置正确。
3. 保持 LarePass VPN 开启，关闭 Wi-Fi 并切换到移动数据。使用**同一个已保存的 Gateway 地址**重新连接，再发送一条消息。

切换网络时无需修改域名、端口或 TLS 设置。手机系统可能挂起后台应用，测试节点能力时请将应用切回前台。

## 故障排查

| 现象 | 检查方法 |
| --- | --- |
| `Expected HTTP 101 response but was '400 Bad Request'` | WebSocket 握手在配对前失败。先检查本机 LarePass VPN 是否开启、主机是否为 **Gateway** 入口，以及是否使用 `443` 和 TLS。仅凭 `400` 无法确定是哪个代理返回的错误。 |
| 二维码包含回环地址、容器 IP 或错误域名 | 使用 `openclaw qr --url 'wss://gateway.example.com'` 重新生成。 |
| 配对码过期或被拒绝 | 重新生成二维码并扫码配对。不要将二维码中的配对凭据填入手动设置的令牌栏。 |
| 令牌认证失败 | 从当前 OpenClaw 实例重新获取令牌。不要使用 Olares 密码或其他克隆实例的令牌。 |
| `pairing required` 或 `unknown requestId` | 按[审批待配对设备](#审批待配对设备)操作，使用自己手机的最新请求。 |
| CLI 已显示配对成功，手机仍显示旧错误 | 保持 VPN 开启并重新连接，必要时关闭并重新打开手机应用。 |
| iOS 显示 `Gateway setup incomplete` | 生成新的二维码或配对码，重新配对，以获得节点和操作员两类凭据。 |

## 了解更多

- [管理应用入口](../manual/olares/settings/manage-entrance.md)
- [OpenClaw Android 文档](https://docs.openclaw.ai/platforms/android)
- [OpenClaw iOS 文档](https://docs.openclaw.ai/platforms/ios)
