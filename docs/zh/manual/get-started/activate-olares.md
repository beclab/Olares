---
search: false
---
## 激活 Olares

使用向导 URL 和初始一次性密码进行激活和 Olares 初始化配置。

1. 在浏览器中输入向导 URL。进入欢迎页面后，按任意键继续。

   ![打开向导](/images/manual/get-started/open-wizard.png#bordered)
2. 输入一次性密码，点击**继续**。

   ![输入密码](/images/manual/get-started/wizard-enter-password.png#bordered)
3. 选择系统语言。

   ![选择语言](/images/manual/get-started/select-language.png#bordered)
4. 选择一个距你所在位置最近的反向代理节点。你也可以之后在 Olares 的[更改反向代理](../olares/settings/change-frp.md)页面进行调整。
   ![选择 FRP](/images/zh/manual/get-started/wizard-frp.png#bordered)
   :::tip 提示
   如果你的 Olares 设备连接的是公网 IP 网络，此步骤会自动跳过。
   :::
5. 使用 LarePass 应用激活 Olares。

   a. 打开 LarePass 应用，点击**扫描二维码**，扫描向导页面上的二维码完成激活。
   :::warning 网络要求
   为避免激活失败，管理员用户需确保手机和 Olares 设备连接到同一网络。
   :::

   ![激活 Olares](/images/manual/get-started/activate-olares.png#bordered)

   b. 按照 LarePass 上的提示重置 Olares 的登录密码。

   ::: tip 提示
   如果你重新安装了 Olares，原有实例将不可用。你可以使用同一个 ID [重新激活 Olares](../larepass/activate-olares.md#使用同一-olares-id-重新激活)：
   :::

设置成功后，LarePass 应用会自动返回主界面，向导页面则会跳转到登录界面。