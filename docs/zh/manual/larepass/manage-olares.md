---
description: 使用 LarePass 移动端远程管理 Olares，包括监控状态、网络配置、远程控制与设备信息查看。
---

# 使用 LarePass 管理 Olares

**LarePass** 应用支持你在手机上远程管理 Olares 设备，包括 Olares 升级、系统状态监控、网络配置、远程控制和设备信息查看等功能。

## 前提条件

开始前，请确认：

- 已拥有有效的 **Olares ID**，且 Olares 设备已激活。  
- Olares 设备已通电并连接至网络。
- 当前账户具备管理员权限。  

## 进入 Olares 管理界面

在**Olares 管理**页面，你可以通过 LarePass 统一管理 Olares 设备。
![Olares 管理](/images/zh/manual/larepass/olares-management.png#bordered)

1. 打开 LarePass，进入**设置**。 
2. 在**我的 Olares** 卡片里，点击**系统**，进入 **Olares 管理**页面。

## 升级 Olares

:::warning 仅管理员可以升级
只有 Olares 管理员可以执行系统更新。更新将应用于同一 Olares 集群内的所有用户。
:::

当有新的 Olares 版本可用时，你会在 **Olares 管理** > **系统更新**选项看到`发现新版本`提示。安装更新的方法如下：

![升级设备](/images/zh/manual/larepass/olares-upgrade.png#bordered)

1. 在 **Olares 管理**页面，点击**系统更新**。
2. 在**系统更新**页面，确认**新版本**字段中的可更新版本信息，然后点击**升级**。
3. 在弹出的对话框中，选择升级方式，并点击**确认**：
    - **仅下载**：只下载更新包，下载完成后你需要在系统升级页面点击**升级**才会开始安装。在下载过程中，Olares 可继续使用。
   - **下载并升级**：Olares 将立即下载更新包并安装。在升级过程中系统将暂时不可用。
4. 等待升级完成。升级成功会有提示信息，Olares 会自动恢复正常运行。

## 远程设备控制

在 Olares 管理页右上角点击 <i class="material-symbols-outlined">power_settings_new</i>，可执行：

 ![控制设备](/images/zh/manual/larepass/device-control.png#bordered))
- **重启 Olares** – 设备将重启。如手机和 Olares 在同一网络，重启过程中 Olares 状态将显示为 `正在重启`，约 5–8 分钟后恢复为 `Olares 运行中`。  
- **关闭 Olares** – 设备关机，如手机和 Olares 在同一网络，Olares 状态将显示为 `Olares 已关机`。关机后无法执行远程操作，需手动开机。 

::: tip 注意
如果你在 Olares 之外的网络执行了重启操作，**我的 Olares** 卡片在重启过程中将无法访问，启动完成即可恢复正常。
:::


## 网络配置

你可以在**Wi-Fi 配置**页面查看或变更当前网络设置。

:::tip 注意
此操作需要你的手机和 Olares 处于同一网络。
:::

### 有线切换至无线

若 Olares 通过有线网络激活，可用 LarePass 切换至同一网络的 Wi-Fi：

![Wi-Fi 切换](/images/zh/manual/larepass/switch-wifi.png#bordered)

1. 在 Olares 管理页面，点击**Wi-Fi 配置**选项，进入**选择连接方式**页面。 
2. 点击列表里的 Wi-Fi 网络以连接。 若 Wi-Fi 有密码，在弹出窗口里输入密码并确认。  
3. 连接成功后，网络自动切换至 Wi-Fi，过程大概会持续 5 分钟。Olares 状态首先会显示 `IP 地址变更中`，切换完成后恢复 `Olares 运行中`。  

切换后，你可以用同样的步骤切换回有线网络。

::: tip 建议
为获得最佳稳定性，优先使用有线网络连接 Olares。
:::

### 更新 IP 地址

当 Olares 迁移至新网络：

1. 将 Olares 接入有线网络并开机，并将手机接入同一网络的 Wi-Fi。  
2. 打开 LarePass，进入**设置** > **我的 Olares** > **系统** > **Olares 管理**。
3. LarePass 会自动扫描局域网中的 Olares，找到后状态显示 `IP 地址变更中`。  
4. IP 更新完成后，状态变为 `Olares 运行中`，约需 5–10 分钟。  

### 蓝牙配网

激活 Olares 时，如果设备无法接入有线网络，或 Olares 所连有线网络与你的手机网络不同， LarePass 可能无法完成激活，也无法执行需在同一网络下的设备管理操作（如无线网络配置或恢复出厂设置）。此时，你可以通过蓝牙配网，将 Olares 连接到手机的 Wi-Fi 网络。

![蓝牙配网](/images/zh/manual/larepass/bluetooth-network.png#bordered)

1. 在**未发现 Olares** 提示页面底部，点击**蓝牙配网**选项。LarePass 将使用手机蓝牙扫描附近的 Olares 设备。

2. 设备显示后，点击**配置网络**。

3. 选择手机当前连接的 Wi-Fi 网络。如果该网络有密码保护，请输入密码并点击**确认**。

4. Olares 将开始网络切换过程。完成后您会看到成功消息。此时，如返回到**蓝牙配网**页面，你将看到 Olares 的 IP 地址已更改为你手机 Wi-Fi 一样的网络。

   ::: tip 注意
   如果你的 Olares 之前已激活，此过程将耗时更长，因为网络切换会影响更多服务。
   :::

5. 返回到设备扫描页面，并点击**发现附近的 Olares**，即可顺利发现你的设备。你可以继续[激活设备](activate-olares.md)或执行本文档里的设备管理操作。

## 查看设备信息

在 Olares 管理页面，点击顶部的设备信息区域，可进入设备详情页面并查看以下信息：

- 硬件详情  
- 系统版本
- 当前的网络连接状态，包括内网和外网 IP 地址。  

## 卸载 Olares
:::tip 同一网络操作
此操作需要你的手机和 Olares 处于同一网络。
:::

此操作会将设备恢复到待安装状态，届时可在局域网重新扫描、安装并激活 Olares。
![卸载 Olares](/images/manual/larepass/restore-to-factory.png#bordered)


::: warning 谨慎操作
该操作将永久删除所有账户信息与数据。
:::

1. 在**Olares 管理**页面点击**恢复出厂设置**。  
2. 阅读风险提示，并输入 LarePass 本地锁屏密码；若未设置，将提示先创建。  
3. 等待卸载完成，系统将返回 Olares ID 登录界面。

## 重置 SSH 密码 <Badge type="tip" text="Olares One 专有" />

使用 LarePass 激活 **Olares One** 后，会自动弹出**重置 SSH 密码**对话框。请立即修改默认 SSH 密码，以防止非授权的 SSH 访问。
![Reset SSH Password](/images/zh/manual/larepass/change-ssh-pw.png)

1. 在**重置 SSH 密码**对话框中输入你的新密码（确保满足强度要求）。
2. 点击**确认**。

:::warning 必需操作
在你完成密码重置之前，该对话框会反复弹出。虽然你也可以之后在 Olares 的[**我的硬件**](../olares/settings/my-olares.md#reset-ssh)里操作，我们强烈建议在 LarePass 里看到弹窗时立即完成重置。
:::

