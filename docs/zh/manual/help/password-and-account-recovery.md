---
outline: [2, 3]
description: 辨认 Olares 使用的凭据，了解它们保护的内容，并找到相应的恢复方式。
---
# 识别需要的 Olares 凭据

Olares 使用不同的凭据登录账户、解锁 LarePass、恢复身份和访问主机。它们保护的内容不同，不能相互替代。

:::warning
如果无法取得助记词的任何副本，就无法恢复 Olares ID。
:::

## 常见提示

| 提示出现的位置 | 凭据 | 用途 |
| --- | --- | --- |
| 激活向导 | Wizard 一次性密码 | 进入[激活流程](../get-started/join-olares.md#步骤-2激活账号) |
| Olares 登录页 | [Olares 登录密码](#olares-登录密码) | 登录 Olares 及其登录系统保护的应用 |
| 输入登录密码后 | Olares 登录验证 | 通过第二重验证[确认登录](../get-started/join-olares.md#步骤-3登录-olares) |
| LarePass | [LarePass 本地密码](#larepass-本地密码) | 解锁当前 LarePass 安装中的受保护功能 |
| 导入或恢复账户时 | [Olares ID 助记词](#olares-id-助记词) | 导入已有 Olares ID |
| 通过 SSH 或主机本地登录时 | [主机或 SSH 密码](#主机或-ssh-密码) | 访问操作系统和 Olares 主机终端 |

## 凭据详情

### Olares 登录密码

用于登录 Olares 及其登录系统保护的应用。它不能解锁 LarePass 或登录主机操作系统。

- **保存位置**：以受保护的形式保存在 Olares 设备上，由身份验证服务管理。
- **忘记后**：
  - **团队成员**：请团队管理员在[设置中重置密码](../olares/settings/manage-team.md#重置密码)。
  - **有权访问主机终端的 Olares 管理员**：[从终端重置密码](./ts-forget-login-password.md)。
- **修改密码**：在[设置中修改自己的密码](../password-and-devices.md#更改密码)。

### LarePass 本地密码

用于解锁 LarePass 的受保护功能，包括查看助记词。当前安装中的所有 Olares ID 共用这个密码。

- **保存位置**：当前安装的 LarePass。每个安装都有独立的本地密码。

- **忘记后**：
  - **已开启生物识别解锁**：在移动端打开**设置** > **安全**，查看密码。
  - **其他情况**：卸载并重新安装 LarePass，使用助记词导入 Olares ID，并设置新的本地密码。

### Olares ID 助记词

这组由 12 个单词组成的助记词用于在另一处 LarePass 安装中导入 Olares ID。Olares 无法重置或重新提供助记词。

- **保存位置**：LarePass 会在本地保存加密副本。请在自己控制的安全位置另存一份完整副本。

参见[备份助记词](../larepass/back-up-mnemonics.md)和[在 LarePass 中管理账户](../larepass/manage-accounts.md#导入账户)。

### 主机或 SSH 密码

用于登录运行 Olares 的设备上的操作系统，包括访问主机终端。

- **保存位置**：主机操作系统管理账户密码。在 Olares One 上，激活后自动生成的 SSH 密码也会保存到 Vault。
- **忘记 Olares One 的密码**：在 LarePass 的 Vault 中查看已保存的密码。

参见[通过 SSH 访问 Olares One 终端](../../one/access-terminal-ssh.md)。

主机密码与 Olares 登录密码相互独立。重置其中一个不会改变另一个。
