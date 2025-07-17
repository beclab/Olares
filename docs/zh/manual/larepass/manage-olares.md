---
description: 使用 LarePass 移动端远程管理 Olares，包括监控状态、网络配置、远程控制与设备信息查看。
---

# 使用 LarePass 管理 Olares

**LarePass** 应用让你在手机端即可远程管理 Olares 设备：监控系统状态、配置网络、执行远程控制，并查看关键设备信息。

## 前提条件

开始前，请确认：

- 已拥有有效的 **Olares ID**，且 Olares 设备已激活。  
- Olares 设备已通电并连接网络。  
- 手机与 Olares 设备位于同一局域网。  
- 当前账户具备管理员权限。  

## 进入 Olares 管理界面

1. 打开 LarePass，进入**设置**。  
2. 点击你的 **Olares ID**，进入 Olares 管理页面。  

## 远程设备控制

在 Olares 管理页右上角点击 <i class="material-symbols-outlined">power_settings_new</i>，可执行：

 ![控制设备](/images/zh/manual/larepass/device-control.png)
- **重启 Olares** – 设备将重启，状态显示 `正在重启`，约 5–8 分钟后恢复为 `Olares 运行中`。  
- **关闭 Olares** – 设备关机，状态显示 `Olares 已关机`，需手动开机。  

## 网络配置

你可以在**Wi-Fi 配置**页面查看或变更当前网络设置。

### 有线切换至无线

若 Olares 通过有线网络激活，可用 LarePass 切换至同一网络的 Wi-Fi：

![Wi-Fi 切换](/images/zh/manual/larepass/switch-wifi.png)

1. 在 Olares 管理页面，点击**Wi-Fi 配置**选项，进入**选择连接方式**页面。 
2. 点击列表里的 Wi-Fi 网络以连接。 若 Wi-Fi 有密码，在弹出窗口里输入密码并确认。  
3. 连接成功后，网络自动切换至 Wi-Fi，过程大概会持续 5 分钟。Olares 状态首先会显示 `IP 地址变更中`，切换完成后恢复 `Olares 运行中`。  

切换后，你可以用同样的步骤切换回有线网络。

::: tip 建议
为获得最佳稳定性，优先使用有线网络。
:::

### 更新 IP 地址

当 Olares 迁移至新网络：

1. 将 Olares 接入有线网络并开机，并将手机接入同一网络的 Wi-Fi。  
2. 打开 LarePass，进入 **Olares 管理**。  
3. LarePass 会自动扫描局域网中的 Olares，找到后状态显示 `IP 地址变更中`。  
4. IP 更新完成后，状态变为 `Olares 运行中`，约需 5–10 分钟。  

### 蓝牙配网

如果在激活 Olares 时无法连接有线网络，或者 Olares 连接到的有线网络与你的手机网络不同，LarePass 在局域网里无法发现 Olares，你也无法顺利完成激活或设备管理。在这种情况下，可使用蓝牙配网功能将 Olares 连接到你手机的 Wi-Fi 网络。

![蓝牙配网](/images/zh/manual/larepass/bluetooth-network.png)

1. 在**未发现 Olares**提示页面底部，点击**蓝牙配网**选项。LarePass 将使用手机蓝牙扫描与当前登录账号匹配的 Olares 设备。

2. 找到设备后，点击**配置网络**。

3. 从 Olares 可用的无线网络列表中，点击你手机当前连接的 Wi-Fi 网络。如果该网络有密码保护，请输入密码并点击**确认**。

    ::: tip 注意
    如果你选择的 Wi-Fi 网络与手机不同，连接后 LarePass 仍将无法识别你的 Olares 设备。
    :::
4. Olares 将开始网络切换过程。完成后您会看到成功消息。此时，如返回到**蓝牙配网**页面，你将看到 Olares 的 IP 地址已更改为你手机 Wi-Fi 一样的网络。

   ::: tip 注意
   如果你的 Olares 之前已激活，此过程将耗时更长，因为网络切换会影响更多服务。
   :::

5. 返回到设备扫描页面，现在应该可以找到你的 Olares 设备了。你可以继续[激活设备](activate-olares.md)或执行本文档里的设备管理操作。

## 查看设备信息

点击管理页顶部的设备信息区域，可查看以下设备信息：

- 硬件详情  
- 系统版本  
- 资源使用情况  
- 当前网络连接  

## 卸载 Olares

此操作会将设备恢复到待安装状态，届时可在局域网重新扫描、安装并激活 Olares。
![卸载 Olares](/images/manual/larepass/restore-to-factory.png)

::: warning 谨慎操作
该操作将永久删除所有账户信息与数据。
:::

1. 在**Olares 管理**页面点击**恢复出厂设置**。  
2. 阅读风险提示，并输入 LarePass 本地锁屏密码；若未设置，将提示先创建。  
3. 等待卸载完成，系统将返回 Olares ID 登录界面。  

