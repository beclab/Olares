---
outline: [2, 3]
description: 排查 LarePass VPN 在 macOS 或 Windows 上无法使用的问题。
---

# LarePass VPN 无法使用

当 LarePass VPN 开关没有反应、连接停滞在“连接中”，或之前可以正常使用的 VPN 连接在 macOS 或 Windows 上突然失效时，可参考本指南。

## 适用情况

- 在 macOS 上，点击 LarePass 桌面客户端的 VPN 开关没有反应，或 VPN 停滞在“连接中”。
- 在 macOS 上，LarePass VPN 之前在此设备上可以正常使用，但现在无法连接或连接后立即断开。
- 在 Windows 上，点击 LarePass 桌面客户端的 VPN 开关没有反应，或 VPN 无法启用。

## 原因

在 macOS 上，LarePass VPN 需要完整设置系统级网络扩展和 VPN 配置。如果在初次设置时跳过或未完成任一步骤，或者网络扩展出现卡死或损坏，macOS 将阻止 LarePass 建立 VPN 隧道。

在 Windows 上，第三方杀毒或安全软件可能误将 LarePass 桌面客户端标记为可疑程序，导致 VPN 服务无法启动。

## 解决方案

### macOS

重置网络扩展并完整走完系统设置流程，以恢复 VPN 连接。

:::info
不同 macOS 版本下界面可能略有差异。
:::

1. 打开**系统设置**，搜索"扩展"，选择**扩展**。
2. 滚动到**网络扩展**部分，点击 <span class="material-symbols-outlined">info</span> 查看已加载的扩展。
   ![系统设置中的网络扩展部分](/images/manual/help/ts-vpn-network-extensions.png#bordered){width=70%}

3. 找到 **LarePass**，点击三个点（**...**），选择**删除扩展**。
4. 确认卸载。
5. 重启 Mac。
6. 重新打开 LarePass 桌面客户端并启用 VPN。
7. 按照系统提示完成扩展和 VPN 配置的恢复：

   a. 当 macOS 提示添加 LarePass 网络扩展时，点击**打开系统设置**。
   ![添加 LarePass 网络扩展的提示](/images/manual/help/ts-vpn-add-network-extension.png#bordered){width=40%}

   b. 打开 **LarePass** 开关。
   ![打开 LarePass 网络扩展开关](/images/manual/help/ts-vpn-toggle-on-network-extension.png#bordered){width=70%}

   c. 当提示添加 VPN 配置时，点击**允许**。
   ![添加 VPN 配置的提示](/images/manual/help/ts-vpn-add-vpn-configuration.png#bordered){width=40%}

### Windows

:::info LarePass 在首次启动时被拦截
如果杀毒软件在安装后首次打开 LarePass 时将其拦截，先在安全软件中放行该应用，再执行以下步骤。
:::

1. 在杀毒或安全软件中，打开**白名单**、**排除项**或**例外**设置。
2. 将 LarePass 主程序或安装目录加入白名单。常见路径包括：
   - `C:\Users\<用户名>\AppData\Local\LarePass\`
   - `C:\Program Files\LarePass\`
3. 应用更改，如有需要重启杀毒软件。
4. 退出并重新打开 LarePass 桌面客户端。
5. 在 LarePass 中再次尝试启用**专用网络连接**。
