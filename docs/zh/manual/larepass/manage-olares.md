---
outline: [2, 3]
description: 了解如何使用 LarePass 应用管理你的 Olares 设备。
---
# 使用 LarePass 管理 Olares

LarePass 应用可以让你轻松管理你的 Olares 设备。你可以通过手机升级 Olares、监控系统状态、管理网络连接、执行远程控制以及查看设备信息。

## 前提条件

开始前，请确认：

- 已拥有有效的 Olares ID，且 Olares 设备已激活。  
- Olares 设备已通电并连接至网络。
- 当前账户具备管理员权限。  

## 访问 Olares 管理页面

**Olares 管理**页面是使用 LarePass 管理 Olares 设备的核心入口。要访问 **Olares 管理**页面：

1. 打开 LarePass，进入**设置**。 
2. 在**我的 Olares** 卡片里，点击**系统**，进入 **Olares 管理**页面。

![访问 Olares 管理](/images/zh/manual/larepass/system.png#bordered)

在此页面上，你可以：
- 查看设备信息
- 升级 Olares
- 远程重启或关闭 Olares
- 修改网络设置
- 将 Olares 恢复出厂设置
- 重置 SSH 密码 <Badge type="tip" text="Olares One 专有" />

![Olares 管理](/images/zh/manual/larepass/olares-management1.png#bordered)

## 查看设备信息

点击顶部的设备信息区域，可以查看设备详情，包括：
- 硬件详情
- 系统版本
- 当前的网络连接状态，包括内网和外网 IP 地址

![查看设备信息](/images/zh/manual/larepass/view-device-information.png#bordered)

## 升级 Olares
:::warning 仅管理员可以升级
只有 Olares 管理员可以执行系统更新。更新将应用于同一 Olares 集群内的所有用户。
:::

要安装系统更新：

1. 在 **Olares 管理**页面，点击**系统更新**。

2. 在**系统更新**页面，确认**新版本**字段中的可更新版本信息，然后点击**升级**。
   ![检查可用更新](/images/zh/manual/larepass/check-version.png#bordered)

3. 在弹出的对话框中，选择升级方式，并点击**确认**：
    - **仅下载**：只下载更新包，下载完成后你需要在系统升级页面点击**升级**才会开始安装。在下载过程中，Olares 可继续使用。
    - **下载并升级**：Olares 将立即下载更新包并安装。在升级过程中系统将暂时不可用。
   ![升级方法](/images/zh/manual/larepass/olares-upgrade1.png#bordered)

4. 等待升级完成。升级成功会有提示信息，Olares 会自动恢复正常运行。

## 远程重启或关闭 Olares

在 **Olares 管理**页右上角点击 <i class="material-symbols-outlined">power_settings_new</i>，可执行：
- **关闭 Olares**：设备将关机。如手机和 Olares 在同一网络，Olares 状态将显示为 `Olares 已关机`。关机后无法执行远程操作，需手动开机。
- **重启 Olares**：设备将重启。如手机和 Olares 在同一网络，重启过程中 Olares 状态将显示为 `正在重启`，约 5–8 分钟后恢复为 `Olares 运行中`。  
  ::: tip 注意
  如果你在 Olares 之外的网络执行了重启操作，**我的 Olares** 卡片在重启过程中将无法访问，启动完成即可恢复正常。
  :::

![远程重启或关闭 Olares](/images/zh/manual/larepass/device-control1.png#bordered)

## 修改网络设置
:::tip 需要同一网络
此操作需要你的手机和 Olares 处于同一网络。
:::
:::tip 建议
为获得最佳稳定性，优先使用有线网络连接 Olares。
:::

你可以在 Olares 设备的有线和 Wi-Fi 网络选项之间切换连接。

1. 在 Olares 管理页面，点击**Wi-Fi 配置**选项。 

2. 在可用网络列表中，点击你想要切换到的网络：  
   - 要切换到 Wi-Fi，请点击所需的 Wi-Fi 网络。
   - 要切换到有线，请点击有线网络选项。
   ![切换网络](/images/zh/manual/larepass/switch-wifi1.png#bordered)

3. 如果你选择的 Wi-Fi 网络受密码保护，请输入密码并点击**确认**。

4. 网络将开始切换。此过程大约需要 5 分钟。Olares 状态首先会显示 `IP 地址变更中`，切换完成后恢复 `Olares 运行中`。  

## 将 Olares 恢复出厂设置
:::warning 谨慎操作
该操作将永久删除所有账户信息与数据。
:::
:::tip 需要同一网络
此操作需要你的手机和 Olares 处于同一网络。
:::

将 Olares 恢复出厂设置将把你的设备重置到准备安装阶段，之后你可以在局域网重新扫描、安装并激活 Olares。

1. 在**Olares 管理**页面点击**恢复出厂设置**。

2. 阅读风险提示，并输入 LarePass 本地锁屏密码；若未设置，将提示先创建。
   ![阅读风险提示](/images/zh/manual/larepass/review-risk-prompt.png#bordered)

3. 输入你的 LarePass 本地解锁密码，然后点击**确认**。如果你尚未设置本地密码，系统会提示你先设置一个。
   ![输入本地解锁密码](/images/zh/manual/larepass/enter-password-to-uninstall.png#bordered)

4. 等待重置完成，系统将返回账号激活页面。

## 重置 SSH 密码 <Badge type="tip" text="Olares One 专有" />
:::warning 必须操作
在你完成密码重置之前，该对话框会反复弹出。虽然你也可以之后在 Olares 的[**我的硬件**](../olares/settings/my-olares.md#reset-ssh)里操作，但我们强烈建议在 LarePass 里看到弹窗时立即完成重置。
:::
使用 LarePass 激活 **Olares One** 后，会自动弹出**重置 SSH 密码**对话框。请立即修改默认 SSH 密码，以防止非授权的 SSH 访问。

![重置 SSH 密码](/images/zh/manual/larepass/change-ssh-pw.png)

1. 在**重置 SSH 密码**对话框中输入一个新密码。确保它满足所有强度要求。
2. 点击**确认**。

## 常见问题
### 将 Olares 设备移动到新网络后如何重新连接？
如果你的 Olares 设备移动到不同的网络，它将暂时不可用。要重新连接它：

1. 将 Olares 设备连接到你的新有线网络并开机。
2. 确保你的手机连接到同一网络的 Wi-Fi。 
3. 在手机上打开 LarePass，进入**设置** > **我的 Olares** > **系统** > **Olares 管理**页面。 
4. LarePass 将自动扫描网络内的 Olares 设备。找到后，Olares 在 LarePass 中将显示为 `IP 地址变更中`。 
5. 一旦 IP 更新完成，状态将恢复为`Olares 运行中`。此过程可能需要 5-10 分钟。

### 我把 Olares 设备搬到了一个没有有线网络的新地方，要怎么连接 Wi-Fi？
如果你把已经激活的 Olares 设备搬到一个没有有线网络的地方，它在 LarePass 应用中会显示为离线。

此时可通过**蓝牙配网**选项重新链接。你可以使用手机的蓝牙找到 Olares 设备，并将其连接到新的 Wi-Fi 网络。

这个过程可能会比初次激活时更长（大约 5-10 分钟），因为网络切换需要更新设备上的所有现有服务。