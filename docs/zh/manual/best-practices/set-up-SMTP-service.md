---
outline: [2,3]
description: 本文介绍如何在 Olares 平台上配置 SMTP 服务，以实现邮件通知、邮箱验证及密码找回等功能。教程涵盖获取邮箱授权码和在 Halo 与 Teable 应用中的配置步骤。
---
# 设置 SMTP 邮件服务
SMTP（Simple Mail Transfer Protocol）是一种应用层协议，用于发送和传输电子邮件。在 Olares 平台上，部分应用需要依赖 SMTP 服务来完成特定功能，例如发送动态通知、验证邮箱地址、找回密码等。

本文将以 Halo 和 Teable 两个应用为例，介绍如何获取邮箱授权码并完成 SMTP 配置，确保通过 Olares 部署的应用可以正常使用邮箱功能。

:::info
未来 Olares 会提供更灵活的方式来管理 SMTP 信息，但目前仍需手动进行部分配置。如果你在配置过程中遇到问题，请联系我们。
:::
:::tip
配置完成后，发送的邮件可能会有延迟，或被误识别为垃圾邮件。如果邮件出现在垃圾箱，请标记为“非垃圾邮件”，以便后续正常接收。
:::
## 目标
通过本教程，你将学习：
- 获取邮箱的 SMTP 授权码。
- 在不同的应用中正确配置 SMTP 服务。

## 第一步：获取邮箱 SMTP 授权码
SMTP 授权码由邮箱服务商（如 QQ 邮箱、Gmail、Outlook 等）提供，用于授权第三方应用通过 SMTP 协议发送邮件。  
它比直接使用邮箱密码更安全，是邮件发送的必需凭据。

以 QQ 邮箱为例：
1. 登录 QQ 邮箱，右上角点击**账号与安全**。
2. 进入**安全设置**页面，找到 **POP3/IMAP/SMTP/Exchange/CardDAV/CalDAV服务**，点击**开启服务**。

   ![开启 SMTP](/images/zh/manual/tutorials/enable-SMTP.png#bordered){width=80%}
3. 按照页面提示完成验证。验证成功后，系统会生成一组授权码。

   ![生成授权码](/images/zh/manual/tutorials/generate-AUTH.png#bordered){width=80%}
4. 复制授权码并备注授权码用途，以便稍后配置 SMTP 服务时使用。

## 第二步：配置 SMTP
根据所使用的应用，SMTP 配置方式主要分为以下两种：
- [通过应用 UI 配置（以 Halo 为例）](#halo)
- [通过 Olares 控制面板配置（以 Teable 为例）](#teable)

### 通过应用 UI 配置（以 Halo 为例）{#halo}
1. 登录 Halo，进入控制台。
2. 选择**系统** > **设置** > **通知设置**，打开**启用邮件通知器**。
3. 根据实际情况填入相应字段：
   - **用户名**：填写你的 QQ 邮箱地址。
   - **发信地址**：留空。
   - **密码**：填写获取的 SMTP 授权码。
   - **显示名称**：自定义显示名称，该名称将包含在邮件标题中。
   - **SMTP 服务器地址**：`smtp.qq.com`
   - **端口号**：`465`
   - **加密方式**：`SSL`
4. 点击**测试邮箱**，验证是否能够成功发送邮件。
5. 测试成功后，点击**保存**。

![Halo 配置 SMTP](/images/zh/manual/tutorials/halo-SMTP.png#bordered)
### 通过 Olares 控制面板配置（以 Teable 为例）{#teable}
Teable 应用需要通过修改环境变量来配置 SMTP 服务：
1. 打开 Olares **控制面板**，在当前用户命名空间下找到 Teable 应用。
2. 点击**配置字典** > **teable-config**，右上角点击<i class="material-symbols-outlined">edit_square</i>，打开 YAML 编辑器。

   ![编辑 teable-config](/images/zh/manual/tutorials/teable-config.png#bordered)
3. 更新 SMTP 相关的环境变量。变量值需包含在英文字符的双引号`""`中：
   ```yaml
   BACKEND_MAIL_AUTH_PASS: "" # 授权码
   BACKEND_MAIL_AUTH_USER: "" # QQ 邮箱地址
   BACKEND_MAIL_HOST: "smtp.qq.com"
   BACKEND_MAIL_PORT: "465"
   BACKEND_MAIL_SECURE: "true"
   BACKEND_MAIL_SENDER: "" # QQ 邮箱地址
   BACKEND_MAIL_SENDER_NAME: "" # 自定义一个发件名称
   ```
   例如：
   ```yaml
   BACKEND_MAIL_AUTH_PASS: "abcdefghijklmnop" # 授权码
   BACKEND_MAIL_AUTH_USER: "123456789@qq.com" # QQ 邮箱地址
   BACKEND_MAIL_HOST: "smtp.qq.com"
   BACKEND_MAIL_PORT: "465"
   BACKEND_MAIL_SECURE: "true"
   BACKEND_MAIL_SENDER: "123456789@qq.com" # QQ 邮箱地址
   BACKEND_MAIL_SENDER_NAME: "Olares's teable" # 自定义一个发件名称
   ```
4. 在**控制面板**中点击**部署** > **teable**，在右上角点击<i class="material-symbols-outlined">more_vert</i> > **<i class="material-symbols-outlined">restart_alt</i>重启**。

   ![重启 teable](/images/zh/manual/tutorials/restart-teable.png#bordered)
5. 在弹窗中输入应用名称，点击**确认**，重启 Teable。重启完成后，邮箱功能即生效。

   ![确认重启 teable](/images/zh/manual/tutorials/confirm-restart-teable.png#bordered){width=60%}

