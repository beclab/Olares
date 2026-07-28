---
search: false
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../one/reusables-reset-ssh.md)为准。
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

<!-- #region reset-ssh-in-settings -->
### 在 Olares Settings 中重置

如果你希望使用自定义的 SSH 密码而不是自动生成的密码，可以在 Settings 中手动重置密码。

1. 打开 Settings。在 **My Olares** 页面，选择 **My hardware**。
2. 在底部选择 **Reset SSH login password**。
   ![重置 SSH 登录密码](/images/one/ssh-reset-password-in-settings.png#bordered){width=70%}

3. 在对话框中，输入符合所有强度要求的新 SSH 密码，然后点击 **OK**。
4. 打开 LarePass 应用并扫描屏幕上显示的二维码。
5. 在 LarePass 上点击 **Confirm** 完成。重置后的 SSH 密码将保存到 Vault 中。
<!-- #endregion reset-ssh-in-settings -->
