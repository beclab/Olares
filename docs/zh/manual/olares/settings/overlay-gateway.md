---
outline: [2, 3]
description: 在 Olares 设置中启用 Overlay Gateway，让支持的应用使用局域网访问能力，同时保留 Olares 平台网络。
---

# 管理 Overlay Gateway

Overlay Gateway 可以让支持的应用获得与 Olares 设备位于同一局域网的网络地址。对于需要局域网发现或直连访问的应用很有用，例如投屏、DLNA、UDP 发现和智能家居集成。

启用 Overlay Gateway 后，应用可以使用局域网能力，同时继续保留 Olares 默认网络，用于平台服务、DNS 和其他 Olares 组件通信。

## 开始之前

Overlay Gateway 只有在满足以下条件时可用：

- Olares 运行在原生 Linux 宿主机上。Windows WSL 环境不支持。
- Olares 设备使用有线网络连接。
- 应用支持 Overlay Gateway。
- 如果需要启用或关闭该功能，你必须是 Owner。

如果启用 Overlay Gateway 后，Olares 设备从有线网络切换到 Wi-Fi，Olares 的正常功能不受影响，但 Overlay Gateway 暂时不会生效。恢复有线网络后，该功能才会继续生效。

## 角色权限

所有用户都可以打开**设置** > **网络**，并查看 **Overlay Gateway**。不同角色可执行的操作不同：

- **Owner**：可以在系统级别启用或关闭 Overlay Gateway，也可以为每个支持的应用单独启用或关闭。
- **Admin**：只能查看 Overlay Gateway 状态。
- **成员**：只能查看 Overlay Gateway 状态。反向代理、Hosts 等其他网络设置入口会被隐藏。

## 启用 Overlay Gateway

1. 打开**设置**，进入**网络** > **Overlay Gateway**。
2. 查看页面顶部的状态。

   如果 Overlay Gateway 当前不可用，系统级开关会置灰，应用列表会被隐藏。页面会显示不可用原因，例如宿主环境不支持，或当前不是有线网络连接。

3. 如果你是 Owner，打开系统级 Overlay Gateway 开关。
4. 在应用列表中，为需要局域网访问能力的应用打开 Overlay Gateway。
5. 在确认弹窗中点击**确认**。

如果当前没有已安装的应用支持 Overlay Gateway，启用系统级开关后，页面会显示空状态。

## 管理应用状态

Overlay Gateway 需要为每个支持的应用单独启用。

当你为正在运行的应用打开或关闭 Overlay Gateway 时，Olares 会重启该应用，让配置生效。应用会进入加载状态，直到恢复为**运行中**。

如果应用重启失败，应用会变为**已停止**，但 Overlay Gateway 设置仍会保存。你可以稍后再启动该应用。

如果应用原本处于停止状态，修改 Overlay Gateway 开关后，Olares 只会保存设置，不会重启应用。应用会继续保持**已停止**状态。

## Overlay Gateway 不可用时

如果 Olares 检测到 Overlay Gateway 不再可用，所有应用级 Overlay Gateway 设置会自动关闭。例如，设备不再使用有线网络，或当前宿主环境不支持该功能时，都可能发生这种情况。

当 Overlay Gateway 重新可用后，Owner 需要再次启用系统级开关，并为需要的应用重新打开 Overlay Gateway。
