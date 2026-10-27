---
search: false
head:
  - - meta
    - name: keywords
      content: Olares One, 重置 SSH 密码, SSH 访问, Vault, LarePass, 激活时重置
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../one/reusables-reset-ssh.md)为准。
:::

## 重置 SSH 密码

### 激活时重置
<!-- #region reset-ssh-upon-activation -->
激活 Olares 后，LarePass 应用会提示你重置 SSH 密码。密码会自动生成并保存到你的 Vault 中。

<!-- #region view-saved-ssh-password -->
要在 Vault 中查看已保存的密码：

1. 打开 LarePass 手机应用，点击 **Vault** 选项卡。
2. 按提示输入本地密码解锁。
3. 点击左上角的 **Authenticator** 打开侧边导航，然后点击 **All vaults** 显示所有已保存的项目。
    ![切换 Vault 筛选器](/images/one/ssh-switch-filter.png#bordered)

4. 找到带有 <span class="material-symbols-outlined">terminal</span> 图标的项目并点击它以显示密码。
    ![在 Vault 中查看已保存的登录密码](/images/one/ssh-check-password-in-vault.png#bordered)
<!-- #endregion view-saved-ssh-password -->
<!-- #endregion reset-ssh-upon-activation -->
