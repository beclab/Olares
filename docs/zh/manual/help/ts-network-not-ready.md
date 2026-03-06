---
outline: [2, 3]
description: 排查 Olares One 设备已连接网络但无法通过标准访问方式连接的问题。
---

# Network not ready 或 olares connection error

当你的 Olares One 设备已开机并连接至网络，但突然停止响应时，可参考本指南进行排查。

## 适用情况

- LarePass 移动端显示 **Network not ready**，访问 Olares 桌面时出现 **olares connection error**，但路由器显示设备已连接，且设备能响应网络 `ping` 命令。
- 重启设备和路由器后，问题依然存在。

## 原因

Olares One 设备的底层操作系统运行正常，因此能够成功连接至路由器并显示在线。然而，核心 Olares 软件服务（Kubernetes 集群）意外卡死或崩溃。

## 解决方案

按以下步骤收集诊断信息，以便 Olares 团队协助你恢复访问。

### 步骤 1：尝试 SSH 连接

建议优先尝试此方法，因为这是访问设备并收集诊断信息最便捷的方式。

1. 获取 Olares One 的局域网 IP 地址。

   a. 打开 LarePass 移动端，进入 **设置** > **系统**，打开 **Olares 管理**页面。
   ![点击系统标签页](/images/manual/get-started/larepass-system.png#bordered)

   b. 点击 Olares One 设备卡片。

   c. 向下滚动至 **网络** 部分，记录 **内网 IP**。

2. 在 Vault 中查看 SSH 密码。

   a. 在 LarePass 移动端中点击 **Vault**。根据提示输入本地密码解锁。

   b. 点击左上角的 **所有 Vault** 打开侧边导航，然后点击 **所有 Vault** 显示所有已保存条目。
      ![切换 Vault 筛选器](/images/one/ssh-switch-filter.png#bordered)

   c. 找到带有 <span class="material-symbols-outlined">terminal</span> 图标的条目，点击查看密码。
      ![在 Vault 中查看保存的 SSH 密码](/images/one/ssh-check-password-in-vault.png#bordered)

3. 通过 SSH 连接。

   a. 在电脑上打开终端。

   b. 输入以下命令，将 `<local_ip_address>` 替换为内网 IP，然后按 **Enter**：

      ```bash
      ssh olares@<local_ip_address>
      ```

   c. 根据提示输入 SSH 密码，然后按 **Enter**。

   d. 如果连接成功，直接跳转至[步骤 3](#步骤-3检查系统状态)。

### 步骤 2：本地登录设备

当 SSH 无法访问时，使用显示器和键盘本地登录设备。

1. 将显示器和键盘连接至 Olares One。屏幕上会自动显示基于文本的登录提示符。

   ```text
   olares login:
   ```

2. 输入用户名 `olares` 并按 **Enter**。
3. 输入**步骤 1** 中获取的 SSH 密码并按 **Enter**。

### 步骤 3：检查系统状态

1. 登录成功后，输入以下命令并按 **Enter**，查看内部软件组件列表：

   ```bash
   kubectl get pods -A
   ```

2. 检查 **STATUS** 列，查找状态不是 `Running` 或 `Completed` 的组件。
3. 对完整命令输出拍摄清晰照片或截图，或手动记录异常组件。
4. 通过[提交 GitHub Issue](https://github.com/beclab/Olares/issues/new) 将照片或记录连同问题描述发送给 Olares 团队。
