---
outline: [2, 3]
description: 排查并重置忘记的 Olares 桌面登录密码。
---

# 忘记桌面登录密码

按照本指南，从主机终端重置你的 Olares 桌面登录密码。

## 适用情况

尝试登录 Olares 桌面时，看到“认证失败，密码错误”的提示。

## 原因

忘记了 Olares 桌面的登录密码。

## 解决方案

要重置密码，需要访问 Olares 设备的主机终端，并执行几条命令。

:::info
你需要准备好 Olares 设备的以下信息：
- 本地 IP 地址
- 设备的用户名和密码
:::

### 步骤 1：访问主机终端

通过以下任一方式连接到 Olares 设备的主机终端：

- **SSH**：在与设备处于同一局域网的电脑上打开终端，运行 `ssh <用户名>@<设备IP>`。
- **本地登录**：将显示器和键盘直接连接到设备，然后登录。

### 步骤 2：重置密码

1. 启用重置权限：

    ```bash
    kubectl patch clusterrole backend:auth-provider --type='json' -p='[{"op": "add", "path": "/rules/0/nonResourceURLs/-", "value": "/cli/api/reset/*"}]'
    ```

2. 执行重置命令：

    ```bash
    olares-cli user reset-password <olares-id> -p <新密码>
    ```

    例如，将用户 "alice123" 的密码重置为 "NewSecurePassword456!"：

    ```bash
    olares-cli user reset-password alice123 -p NewSecurePassword456!
    ```

3. 确认执行结果。你应该看到如下输出：

    ```text
    Password for user '<olares-id>' reset successfully
    ```

### 步骤 3：验证登录

等待约 10 秒，待系统同步完成后，使用新密码登录 Olares 桌面。
