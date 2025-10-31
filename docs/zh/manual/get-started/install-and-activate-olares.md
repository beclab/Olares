---
search: false
---
## 安装并激活 Olares
:::warning 检查网络连接
为避免激活失败，请确保你的手机和 Olares 设备连接到同一网络。
:::

1. 打开 LarePass，在账号激活页面点击**发现附近的 Olares**。LarePass 将列出同一网络中检测到的 Olares 实例。
2. 从列表中选择目标 Olares 实例，并点击**立即安装**。
3. 安装完成后，点击**立即激活**。
4. 在**选择反向代理**对话框中，选择一个地理位置离你较近的节点并点击**确认**。安装程序会自动为 Olares 配置 HTTPS 证书和 DNS。
   ![ISO 激活](/images/manual/larepass/iso-activate.png#bordered)
   :::tip 提示
    - 你可以稍后在 Olares 中的 [更改反向代理](../olares/settings/change-frp.md) 页面调整此设置。
    - 如果你的 Olares 设备连接的是公网 IP 网络，此步骤会自动跳过。  
      :::
   
5. 按照屏幕提示设置 Olares 的登录密码，然后点击**完成**。

   ![ISO Activate-2](/images/manual/larepass/iso-activate-2.png#bordered)

激活完成后，LarePass 将显示 Olares 设备的桌面地址，如 `https://desktop.marvin123.olares.cn`。
