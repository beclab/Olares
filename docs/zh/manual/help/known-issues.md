---
outline: [2, 3]
description: Olares 已知问题，包括影响范围、影响说明、临时绕过方式和修复状态。
---

# 已知问题

本页面记录已确认可能影响 Olares 用户的问题，并提供问题的影响范围、当前状态和临时方案。

:::info 状态更新
Olares 会在问题确认、缓解或修复后同步更新本页面。
:::

## 启用 LarePass 专用网络后，Chrome 实时更新可能异常

### 摘要

| 项目 | 说明 |
| --- | --- |
| 影响平台 | macOS |
| 影响浏览器 | Google Chrome 148 或更高版本 |
| 触发条件 | 在 LarePass 专用网络下通过 `olares.com` 地址访问 Olares |
| 影响 | 依赖 WebSocket 的实时更新可能停止工作 |
| 状态 | 计划在后续 Olares 版本中修复 |

### 问题表现

问题发生时，Chrome 中的 Olares 页面仍能打开，但依赖实时更新的内容可能停止刷新。

你可能会看到以下一种或多种现象：

- 应用状态、通知或任务进度不会自动更新。
- 页面看起来卡住，需要刷新后才恢复。
- 实时消息或依赖 WebSocket 的面板停止更新。

该问题只影响浏览器中的实时通信，不会造成数据丢失、数据损坏、账户异常、应用损坏或设备存储问题。

### 原因

从 Chrome 148 开始，Chrome 对从公共网页来源访问本地或私有网络地址的请求执行了更严格的 Local Network Access 检查。

启用 LarePass 专用网络后通过 `olares.com` 地址打开 Olares 时，Chrome 可能会将部分 WebSocket 连接视为本地网络请求。对于 WebSocket 连接，Chrome 会检查 `101 Switching Protocols` 响应。如果响应中缺少 Chrome 期望的响应头，连接会被拦截。

Olares 目前未在 WebSocket upgrade 响应中携带该响应头，因此会被 Chrome 拦截。该响应头会在后续版本中补充。

### 临时绕过方式

在升级到包含修复的 Olares 新版本前，可使用以下任一方式临时绕过。

#### 方案 1：使用 .local 地址

如果你的 Mac 和 Olares 位于同一局域网，建议使用 `.local` 地址，而不是 `olares.com` 地址。这是推荐的局域网访问方式。

将标准 Olares URL 改为 `.local` URL：

**标准 URL**
```text
https://<entrance_id>.<username>.olares.com
```

**本地 URL**
```text
http://<entrance_id>.<username>.olares.local
```

如果 Chrome 无法打开 `.local` 地址，请确认 macOS 已允许 Chrome 访问本地网络：

1. 打开 Apple 菜单，进入**系统设置**。
2. 进入**隐私与安全性** > **本地网络**。
3. 打开 **Google Chrome** 和 **Google Chrome Helper** 的开关。
4. 重启 Chrome 后再次尝试 `.local` URL。

#### 方案 2：临时关闭 Chrome 本地网络访问检查

如果你需要继续通过 `olares.com` 地址在 LarePass 专用网络下访问 Olares，可临时关闭 Chrome 的本地网络访问检查。

:::warning
这会修改浏览器安全设置。建议仅作为临时绕过方式使用。升级到包含修复的 Olares 新版本后，请将该设置恢复为 **Default** 或 **Enabled**。
:::

1. 在 Chrome 中打开：

   ```text
   chrome://flags/#local-network-access-check
   ```

2. 将 **Local Network Access Checks** 设置为 **Disabled**。
3. 点击 **Relaunch** 重启 Chrome。
4. 重新使用 `olares.com` 地址打开 Olares。

#### 方案 3：刷新页面以恢复

如果暂时无法采用上述方案，可在实时更新停止时刷新页面查看最新进度。

1. 按 `F5`，或点击 Chrome 的刷新按钮。
2. 等待页面重新加载。
3. 检查应用状态或实时内容是否恢复更新。
