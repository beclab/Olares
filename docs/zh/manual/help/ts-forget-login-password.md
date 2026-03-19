---
outline: [2, 3]
description: 排查并重置忘记的 Olares desktop 登录密码。
---

# 忘记登录密码

如果忘记了登录密码，可通过本指南重新获得对 Olares desktop 的访问权限。

## 适用情况

- 无法通过浏览器登录 Olares desktop。
- 反复出现“认证失败，密码错误”的提示。

## 原因

Olares 集群中存储的本地认证凭据与输入内容不匹配。

## 解决方案

要解决此问题，必须访问 Olares 主机终端（通过 SSH 或本地 CLI）手动重置登录密码。

:::tip 硬件准备
- Olares 设备已开机并连接到网络。
- 有一台客户端设备（如电脑）用于访问终端。
:::

### 步骤 1：通过 SSH 访问终端

这是访问 Olares 主机终端最便捷的方式。若无法使用 SSH，直接跳转至[步骤 2](#步骤-2-本地登录)。

1. 获取 Olares 设备的本地 IP 地址。

   a. 打开 LarePass 应用，进入 **设置** > **系统**，打开 **Olares 管理**页面。
   ![访问 Olares 管理](/images/zh/manual/larepass/system.png#bordered)

   b. 点按你的设备卡片：**Selfhosted** 或 **Olares One**。

   c. 向下滚动到 **网络** 部分，记下 **内网 IP**。

2. 获取 SSH 密码。

   - Selfhosted：默认密码是 `olares`（除非之前已修改）。
   - Olares One：在 LarePass 中，进入 **Vault** > **所有 Vault**，找到带有 <span class="material-symbols-outlined">terminal</span> 图标的条目，点按查看密码。
      ![在 Vault 中查看保存的 SSH 密码](/images/zh/manual/olares/ssh-check-password-in-vault1.png#bordered)

3. 连接到主机终端。

   a. 在电脑上打开终端。

   b. 输入以下命令，将 `<local_ip_address>` 替换为内网 IP，然后按回车键：

   ```bash
   ssh olares@<local_ip_address>
   ```

   例如：
   ```bash
   ssh olares@192.168.11.12
   ```

   c. 如果提示，输入 `yes` 确认连接，然后按回车键。

   d. 当提示输入密码时，键入 SSH 密码，然后按回车键。

   e. 看到如下提示符表示连接成功，然后跳转到[步骤 3](#步骤-3-重置密码)：

   ```text
   olares@olares:~$
   ```

### 步骤 2：本地登录

如果无法通过 SSH 连接，可使用显示器和键盘直接登录设备。

1. 将显示器和键盘连接到 Olares 设备。屏幕上会自动显示基于文本的登录提示：

   ```text
   olares login:
   ```

2. 输入用户名 `olares` 并按回车键。
3. 输入在[**步骤 1**](#步骤-1-通过-ssh-访问终端) 中获取的 SSH 密码，然后按回车键。

### 步骤 3：重置密码

成功进入终端后，运行以下命令以启用重置权限并更新密码。

1. 输入以下命令，然后按回车键。此命令允许 CLI 执行重置操作。

   ```bash
   kubectl patch clusterrole backend:auth-provider --type='json' -p='[{"op": "add", "path": "/rules/0/nonResourceURLs/-", "value": "/cli/api/reset/*"}]'
   ```

2. 输入以下命令重置密码，然后按回车键：

   ```bash
   olares-cli user reset-password <username> -p <newpassword>
   ```

   例如，为用户名为 "alice123" 的用户重置密码：

   ```bash
   olares-cli user reset-password alice123 -p NewSecurePassword456!
   ```

3. 验证重置结果。

   当终端返回以下信息时，表示密码重置成功：

   ```text
   Password for user '<username>' reset successfully
   ```

### 步骤 4：验证登录

重置密码后，等待约 10 秒钟让系统服务同步新凭据。然后返回 Olares desktop，使用新密码登录。
