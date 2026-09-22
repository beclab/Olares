---
outline: [2, 3]
description: 没有助记词备份时，判断阻止访问 LarePass、Olares 桌面或设备的凭据，并按对应方式恢复访问。
head:
  - - meta
    - name: keywords
      content: Olares, LarePass, 助记词, 本地密码, 登录密码, 系统密码
---

# 恢复对 LarePass、Olares 桌面或 Olares 设备的访问

如果无法使用助记词，并且需要恢复对 LarePass、Olares 桌面或运行 Olares 的设备的访问，请从与你当前情况相符的章节开始。

:::warning 保留现有的 LarePass
不要卸载 LarePass，也不要从仍保留 Olares ID 的设备中删除该 ID。没有助记词备份时，这些操作可能导致该 Olares ID 永久无法再次用于 LarePass。
:::

## 如果忘记了助记词或从未备份

检查所有使用过 LarePass 的手机和电脑。

如果任一 LarePass 客户端中仍有你的 Olares ID，请保留该客户端，不要卸载，并根据客户端类型继续操作。

### 移动端

1. 进入 **Settings** > **LarePass Settings** > **Security** > **Mnemonic phrase**。
2. 如果页面要求输入 LarePass 本地密码，请输入密码。如果不知道本地密码，请参考[如果忘记了 LarePass 本地密码或尚未设置](#如果忘记了-larepass-本地密码或尚未设置)。
3. 查看 12 个单词，按顺序写下并离线保存。
4. 在 LarePass 中完成验证，确认备份内容和顺序正确。

### 桌面端

1. 进入 **Settings** > **Account** > **Manage Account**。
2. 找到你的 Olares ID，按照页面提示查看助记词。
3. 按顺序写下 12 个单词并离线保存。

如果所有 LarePass 客户端中都没有该 Olares ID，并且也没有其他助记词备份，则无法找回助记词。

## 如果忘记了 LarePass 本地密码或尚未设置

本地密码用于解锁当前设备上受保护的 LarePass 功能。每个 LarePass 客户端都有独立的本地密码。

### 移动端

- 如果 LarePass 提示创建本地密码，请按照页面提示完成设置。
- 如果没有出现提示，请进入 **Settings** > **LarePass Settings** 设置本地密码。
- 如果忘记了本地密码，并且已启用生物识别解锁，请进入 **Settings** > **LarePass Settings** > **Security**，点击 **Reveal local password**，并完成生物识别验证。
- 如果未启用生物识别解锁，则无法在该设备上查看本地密码。请改用仍可访问该 Olares ID 的其他 LarePass 客户端。

### 桌面端

- 如果 LarePass 提示创建本地密码，请按照页面提示完成设置。
- 如果没有出现提示，请进入 **Settings** > **Security** 设置本地密码。
- 如果忘记了本地密码，无法在桌面端查看该密码。

## 如果忘记了 Olares 登录密码

请参考[忘记桌面登录密码](./ts-forget-login-password.md)，进入 Olares 设备终端并重置密码。

## 如果找不到双重验证码

如果 LarePass 中没有出现登录确认通知，可以改用双重验证码：

1. 在 Olares 登录页切换到验证码验证方式。
2. 在保留该 Olares ID 的 LarePass 客户端中找到当前的 6 位验证码：

   - 在移动端打开 **Vault**，默认页面会显示双重验证码。也可以进入 **Settings**，找到 **My Olares** 卡片，然后点击身份验证器。
   - 在桌面端打开 **Vault**。身份验证器是列表中的第一项。

3. 在验证码失效前，将其输入 Olares 登录页。

## 如果不知道 Olares 设备终端的用户名、密码或 IP 地址

### Olares One

Olares One 激活后，系统用户名为 `olares`。系统密码会在激活时生成，并保存在 LarePass Vault 中。

如果已经知道系统密码，请直接从步骤 3 开始。

1. 在 LarePass 移动端打开 **Vault**。如果本地密码阻止访问，请参考[如果忘记了 LarePass 本地密码或尚未设置](#如果忘记了-larepass-本地密码或尚未设置)。
2. 点击左上角的筛选项，选择 **All vaults**，再打开带有终端图标的条目查看系统密码。

   ![在 LarePass 中选择 All vaults](/images/manual/help/olares-one-vault-filter.jpg#bordered)

   ![在 Vault 中查看 Olares One 系统密码](/images/manual/help/olares-one-host-password.jpg#bordered)

3. 在 LarePass 中进入 **Settings** > **System**，打开 Olares One 设备卡片，在 **Network** 下找到 **Intranet IP**。
4. 在同一局域网中的电脑上运行：

   ```bash
   ssh olares@<内网-IP>
   ```

5. 输入系统密码。

如果无法使用 SSH，可以将显示器和键盘连接到 Olares One，再使用同一用户名和密码登录。详细操作请参考[通过 SSH 访问 Olares One 终端](/zh/one/access-terminal-ssh.md)或[直接访问 Olares One 终端](/zh/one/access-physical-console.md)。

如果无法打开 Vault，并且也不知道系统密码，请保留 LarePass 并[联系技术支持](./request-technical-support.md)。不要向技术支持发送任何密码或助记词。

### 安装在自有设备上的 Olares

使用安装 Olares 时在该设备上配置的操作系统用户名和密码，通过 SSH 连接或使用显示器和键盘在本地登录。详细操作请参考[访问 Olares 终端](../access-olares-terminal.md)。
