---
outline: [2, 3]
description: 获取设备终端登录信息并重置忘记的 Olares 桌面登录密码。
head:
  - - meta
    - name: keywords
      content: Olares, 忘记密码, 重置密码, 桌面登录, olares-cli, SSH
---

# 忘记桌面登录密码

按照本指南，从主机终端重置你的 Olares 桌面登录密码。

## 适用情况

尝试登录 Olares 桌面时，看到“认证失败，密码错误”的提示。

## 原因

忘记了 Olares 桌面的登录密码。

## 解决方案

要重置密码，需要进入运行 Olares 的设备终端，并执行几条命令。

:::warning 需要管理员权限
仅在你所管理的 Olares 设备主机终端中执行这些命令。备用命令会修改密码重置 API 使用的集群级权限。
:::

### 步骤 1：获取终端登录信息

根据使用的 Olares 设备选择对应步骤。

#### Olares One

系统用户名为 `olares`。系统密码会在激活时生成，并保存在 LarePass Vault 中。

1. 打开 LarePass 移动端，点击 **Vault**。
2. 根据提示输入 LarePass 本地密码。如果不知道本地密码，请参考[如果忘记了 LarePass 本地密码或尚未设置](./ts-access-without-mnemonic.md#如果忘记了-larepass-本地密码或尚未设置)。
3. 点击左上角的筛选项，选择 **All vaults**。
4. 打开带有终端图标的条目，查看系统密码。
5. 进入 **Settings** > **System**，打开 Olares One 设备卡片，在 **Network** 下找到 **Intranet IP**。

#### 安装在自有设备上的 Olares

根据 Olares 的安装方式使用相应的登录信息：

- **通过 Olares ISO 安装在专用设备上**：系统用户名和密码均为 `olares`。
- **安装在已有操作系统上**：使用你在该设备上配置的操作系统用户名和密码。

通过 SSH 登录时，还需要该设备的本地 IP 地址。如果不知道 IP 地址或无法使用 SSH，请连接显示器和键盘，在设备上直接登录。其他终端访问方式请参考[访问 Olares 主机终端](../access-olares-terminal.md)。

### 步骤 2：访问设备终端

通过以下任一方式连接到 Olares 设备的主机终端：

- **SSH**：在与设备处于同一局域网的电脑上打开终端，运行 `ssh <用户名>@<设备IP>`。如果使用 Olares One 或通过 Olares ISO 安装，请使用 `olares` 作为用户名。
- **本地登录**：将显示器和键盘直接连接到设备，然后登录。

### 步骤 3：重置密码

1. 执行重置命令。`<用户名>` 只填写 Olares ID 中域名前的部分。例如，`alice123@olares.com` 应填写 `alice123`。

    ```bash
    olares-cli user reset-password <用户名> -p <新密码>
    ```

2. 根据执行结果继续：

    - 如果执行成功，继续[步骤 4](#步骤-4-验证登录)。
    - 如果返回与重置 API 权限相关的错误，继续下一步。
    - 如果是其他错误，不要修改集群权限。请记录完整报错，并参考[收集诊断信息](../collect-diagnostic-information.md)。

3. 仅当出现重置 API 权限错误时，启用重置权限：

    ```bash
    kubectl patch clusterrole backend:auth-provider --type='json' -p='[{"op": "add", "path": "/rules/0/nonResourceURLs/-", "value": "/cli/api/reset/*"}]'
    ```

4. 再次执行重置命令：

    ```bash
    olares-cli user reset-password <用户名> -p <新密码>
    ```

    例如，将用户 "alice123" 的密码重置为 "NewSecurePassword456!"：

    ```bash
    olares-cli user reset-password alice123 -p NewSecurePassword456!
    ```

5. 确认执行结果。你应该看到如下输出：

    ```text
    Password for user '<用户名>' reset successfully
    ```

### 步骤 4：验证登录

等待约 10 秒，待系统同步完成后，使用新密码登录 Olares 桌面。
