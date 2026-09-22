---
search: false
noindex: true
---
## 安装并激活 Olares
:::warning 检查网络连接
为避免激活失败，请确保你的手机和 Olares 设备连接到同一网络。
:::

<!-- #region iso-activation-flow -->
1. 打开 LarePass。如果还没有 Olares ID，点击**创建账号**并按提示完成设置。

   :::warning 中国大陆用户请选择 `.cn` 域名
   如果手机系统语言为英文，请在高级创建选项中选择 `.cn` 域名。
   :::
2. 在激活页面点击**发现附近的 Olares**。LarePass 会列出同一网络中发现的 Olares 实例。
3. 选择你的 Olares 实例，点击**立即安装**。

   ![ISO 激活](/images/manual/larepass/iso-activate1.png#bordered)
4. 安装完成后，点击**立即激活**。

5. 在**选择反向代理**对话框中，选择一个地理位置离你较近的节点并点击**确认**。安装程序会自动为 Olares 配置 HTTPS 证书和 DNS。
   :::tip 提示
    - 你可以稍后在 Olares 中的 [更改反向代理](../olares/settings/change-frp.md) 页面调整此设置。
    - 如果你的 Olares 设备连接的是公网 IP 网络，此步骤会自动跳过。  
      :::

6. 选择 Olares 语言。Olares 支持英语、简体中文、德语、西班牙语、意大利语、法语和日语。

   :::info
   这里选择的是 Olares 语言，不会改变当前 LarePass 的界面语言。后续激活流程仍使用 LarePass 当前语言。激活完成后，Olares 桌面会使用此处选择的语言。
   :::

7. 按照屏幕提示设置 Olares 的登录密码，然后点击**完成**。

   ![ISO Activate-2](/images/manual/larepass/iso-activate-4.png#bordered)

激活完成后，LarePass 将显示 Olares 设备的桌面地址，如 `https://desktop.marvin123.olares.cn`。
<!-- #endregion iso-activation-flow -->
