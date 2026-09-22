---
outline: [2, 3]
description: 忘记助记词时恢复对 Olares 的访问。
---

# 忘记助记词后无法登录 Olares

如果你忘记了 12 个助记词，且无法登录 Olares，请按本文排查。助记词用于恢复 Olares ID，并不是 Olares 或底层主机操作系统的登录密码。

:::warning
将助记词保存在其他位置之前，不要卸载 LarePass，也不要从中删除 Olares ID。否则可能删除你唯一的助记词副本。
:::

## 适用情况

- 你不记得 Olares ID 的助记词，也没有其他备份。
- 你无法登录 Olares，需要进入主机或重置 Olares 登录密码。

## 解决方案

要恢复访问，请先通过 SSH 登录 Olares 主机，再按需重置 Olares 登录密码。如果无法使用 SSH，也可以通过本地终端登录。

- **在自己的设备上安装 Olares**：使用为该设备设置的主机用户名和密码，然后前往[步骤 3](#步骤-3-进入主机终端)。
- **Olares One**：按步骤 1 和 2 在 LarePass Vault 中查找激活时生成的 SSH 密码。如果你已知道密码，可直接前往[步骤 3](#步骤-3-进入主机终端)。

### 步骤 1：使用 Olares One 时在 LarePass 中打开 Vault

在手机上打开 LarePass。

:::info
连续 6 次输错密码后，LarePass 会锁定账户 15 分钟，期间无法使用生物识别解锁。请停止尝试、关闭 LarePass，等待 15 分钟后再继续。
:::

点击 **Vault**，再根据出现的界面操作：

- **Vault 打开时未要求输入密码**：前往[步骤 2](#步骤-2-查找-olares-one-ssh-密码)。

- **Vault 要求输入密码，而你记得密码**：输入 LarePass 本地密码，然后前往[步骤 2](#步骤-2-查找-olares-one-ssh-密码)。

- **Vault 要求输入密码，而你忘记了密码**：点击面容或指纹图标，尝试使用生物识别解锁。

  如果已启用生物识别解锁，验证通过后，LarePass 会使用保存在手机安全存储区中的本地密码解锁 Vault。

  ![使用生物识别解锁 LarePass](/images/manual/help/olares-one-biometric-verification.png#bordered)

  - 如果 Vault 打开，前往[步骤 2](#步骤-2-查找-olares-one-ssh-密码)。
  - 如果 Vault 无法打开，请保留 LarePass。如果从未启用生物识别解锁，忘记的本地密码无法显示或重置。如果你已知道 Olares One 主机密码，前往[步骤 3](#步骤-3-进入主机终端)；否则，请[联系技术支持](./request-technical-support.md)。

:::info
打开 Vault 后，无需查看本地密码即可继续操作。如果你希望查看密码以备后用，请将 LarePass 更新到最新版本，然后进入 **Settings** > **LarePass Settings** > **Safety** > **Local password**，并完成生物识别验证。
:::

### 步骤 2：查找 Olares One SSH 密码

1. 在 **Vault** 中点击左上角的筛选项，选择 **All vaults**。

   ![在 LarePass 中选择 All vaults](/images/manual/help/olares-one-vault-filter.jpg#bordered)

2. 打开带有终端图标的条目，查看 Olares One SSH 密码。

   ![在 Vault 中查找主机密码条目](/images/manual/help/olares-one-host-password.jpg#bordered)

### 步骤 3：进入主机终端

请使用对应安装环境的主机账户。你可以在主机允许 SSH 连接时通过 SSH 登录，也可以在本地终端登录。

**SSH**

1. 在 LarePass 中点击 **Settings**，在 **My Olares** 下点击 **System**，然后打开设备卡片。根据 LarePass 版本查找并记下 **Intranet IP**：

   - **LarePass 1.11.56 或更高版本**：点击 **Node** > **Network** > **Intranet IP**。
   - **更早版本**：在设备卡片中进入 **Network** > **Intranet IP**。

2. 在连接同一本地网络的电脑上运行：

   ```bash
   ssh <用户名>@<内网-IP>
   ```

3. 根据提示输入主机密码。

:::info Olares One SSH 登录信息
Olares One 的 SSH 用户名为 `olares`，命令为 `ssh olares@<内网-IP>`。激活时生成的 SSH 密码保存在 Vault 中，可以按[步骤 2](#步骤-2-查找-olares-one-ssh-密码)查找。
:::

**本地终端**

将显示器和键盘连接到主机，使用其操作系统账户登录。

进入主机终端后，你就可以访问运行 Olares 的设备。如果也能登录 Olares 桌面，可跳过步骤 4，继续步骤 5。

### 步骤 4：按需重置桌面登录密码

如果你仍能登录 Olares 桌面，可以跳过这一步。如果也忘记了桌面登录密码，请在主机终端运行以下命令：

```bash
kubectl patch clusterrole backend:auth-provider --type='json' \
  -p='[{"op": "add", "path": "/rules/0/nonResourceURLs/-", "value": "/cli/api/reset/*"}]'

olares-cli user reset-password <olares-id> -p '<新密码>'
```

`<olares-id>` 填写 Olares ID 中 `@` 前面的部分。例如，`alice123@olares.com` 应填写 `alice123`。

重置成功后，等待约 10 秒，再使用新密码登录 Olares 桌面。

### 步骤 5：在可查看助记词时备份

如果你仍能解锁 LarePass 并查看助记词，请进入 **Settings** > **LarePass Settings** > **Safety** > **Mnemonic phrase**。按顺序写下全部 12 个单词，并离线保存。登录主机或重置桌面密码都不会显示助记词。

如果某一步与你看到的界面不同，请[联系技术支持](./request-technical-support.md)，说明卡在哪个界面、Vault 是否能打开，以及能否进入主机终端。不要向技术支持提供密码或助记词。
