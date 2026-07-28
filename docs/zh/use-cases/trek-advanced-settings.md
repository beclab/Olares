---
outline: deep
description: 在 Olares 上配置 TREK 的高级设置，包括 OIDC 单点登录、Google API 密钥、双因素认证和数据备份。
head:
  - - meta
    - name: keywords
      content: Olares, TREK, NOMAD, SSO, OIDC, Google API, 2FA, backup, restore, configuration, administration, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-16"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/trek-advanced-settings.md)。
:::

# 配置 TREK 高级设置

随着 TREK 工作空间的增长，您可能希望集成第三方服务、加强账户安全并保护数据安全。

使用 TREK 的管理设置来配置单点登录（SSO）、改进地图功能、启用双因素认证（2FA）和设置备份。

## 学习目标

在本指南中，您将学习如何：
- 设置第三方单点登录以简化用户登录。
- 连接 Google API 密钥以改进地图搜索结果。
- 使用双因素认证保护用户账户。
- 备份和恢复工作空间数据。

## 设置第三方单点登录

通过允许旅行同伴使用他们现有的 Google、Apple 或其他 OIDC 提供商账户登录，来简化登录体验。

以下工作流程演示如何设置 Google SSO。

### 第 1 步：设置 Google Cloud 项目

在 Google Cloud 中准备一个专用项目来管理您的 TREK 认证凭据。

1. 使用您的 Google 账户登录 [Google Cloud 控制台](https://console.cloud.google.com/auth/clients)。
2. 创建一个新项目，或从顶部导航栏选择一个现有项目。

### 第 2 步：配置 OAuth 同意屏幕

定义用户选择使用 Google 登录时将看到的授权屏幕。

1. 在左侧边栏中，将鼠标悬停在 **APIs & Services** 上，然后选择 **OAuth consent screen**。

   ![Select OAuth consent screen](/images/manual/use-cases/trek-oauth-consent-menu.png#bordered)

2. 在 **OAuth Overview** 页面上，点击 **Get started**。
3. 完成以下项目配置，然后点击 **Create**：
   - **App Information**：输入应用名称，例如 `TREK`，选择用户支持邮箱，然后点击 **Next**。
   - **Audience**：选择 **External** 作为目标受众类型，然后点击 **Next**。
   - **Contact Information**：输入用于接收项目更新的邮箱，然后点击 **Next**。
   - **Finish**：选择 **Agree to the Google API Services: User Data Policy**，然后点击 **Continue**。

### 第 3 步：创建 OAuth 客户端 ID

生成 TREK 与 Google 通信所需的凭据。

1. 在 **Metrics** 部分，点击 **Create OAuth client**。

   ![Create OAuth client](/images/manual/use-cases/trek-create-oauth-client.png#bordered)

2. 指定以下设置，然后点击 **Create**：
   - **Application type**：选择 **Web application**。
   - **Name**：输入 OAuth 2.0 客户端的名称以便识别。
   - **Authorized JavaScript origins**：点击 **ADD URI**，然后输入您的 TREK 域名。
   
     例如，`https://8eb06391.alexmiles.olares.com`。
   - **Authorized redirect URIs**：点击 **ADD URI**，然后以 `https://<your-trek-domain>/api/auth/oidc/callback` 格式输入您的回调 URL。
   
     例如，`https://8eb06391.alexmiles.olares.com/api/auth/oidc/callback`。

     ![Google OAuth client](/images/manual/use-cases/trek-google-oauth1.png#bordered)

3. 在 **OAuth client created** 窗口中，复制 **Client ID** 和 **Client secret**，然后点击 **OK**。新的客户端 ID 将显示在 **OAuth 2.0 Client IDs** 页面上。

   ![OAuth 2.0 Client IDs](/images/manual/use-cases/trek-google-oauth-id.png#bordered)

### 第 4 步：将 Google SSO 连接到 TREK

将您在 Google Cloud 中生成的凭据与 TREK 管理设置集成。

1. 返回 TREK，点击您的用户头像，然后选择 **Admin**。
2. 点击 **Settings** 选项卡，然后找到 **Single Sign-On (OIDC)** 面板。
3. 指定以下设置，然后点击 **Save**：
   - **Display Name**：输入 `Google`。
   - **Issuer URL**：输入 `https://accounts.google.com`。
   - **Client ID**：输入您复制的客户端 ID。
   - **Client Secret**：输入您复制的客户端密钥。

   ![Paste OIDC credentials](/images/manual/use-cases/trek-oidc-config.png#bordered)

4. 退出 TREK。
5. 在 **Sign In** 页面上，选择 **Sign in with Google**。

   ![TREK sign in with Google account](/images/manual/use-cases/trek-google-login.png#bordered)

6. 选择您的 Google 账户登录。
7. 当提示 **Sign in to olares.com** 时，点击 **Continue**。您现在已登录 TREK。

## 使用 Google API 密钥改进地图搜索

默认情况下，TREK 使用基础地图。要显示丰富的地点详情，例如照片、评分和营业时间，请将 Google Places API 密钥连接到您的工作空间。

1. 确保您已在 Google Cloud 控制台中创建了 API 密钥。更多信息，请参阅[创建 API 密钥](https://docs.cloud.google.com/docs/authentication/api-keys#create)。

   :::tip 必需的 Google API
   为获得最佳搜索体验，请确保已启用以下 API：Directions API、Geocoding API、Geolocation API、Maps Elevation API、Maps Embed API、Maps JavaScript API、Maps SDK for Android、Places API、Places API (New)、Roads API 和 Time Zone API。
   :::

2. 登录 TREK，点击您的用户头像，然后选择 **Admin**。
3. 点击 **Settings** 选项卡，然后找到 **API Keys** 面板。

   ![API key settings](/images/manual/use-cases/trek-api-key-settings.png#bordered)

4. 在 **Google Maps API Key** 下，输入您的 API 密钥，然后点击 **Test**。将显示 **Connected** 状态，表示密钥有效。
5. 点击 **Save**。
6. 打开您的行程并添加新地点。结果现在会显示额外的上下文信息，例如照片、评分和营业时间。

   ![Place details](/images/manual/use-cases/trek-place-details.png#bordered)
<!--
1. Sign in to the [Google Cloud Console](https://console.cloud.google.com/auth/clients) using your Google account.
2. Create a project or select an existing one.

### Step 2: Enable search-specific APIs

A standard map key only shows the map. To improve search, you must Enable the following APIs:
- Directions API
- Geocoding API
- Geolocation API
- Maps Elevation API
- Maps Embed API
- Maps JavaScript API
- Maps SDK for Android
- Places API
- Places API (New)
- Roads API
- Time Zone API

1. In the left sidebar, hover over **APIs & Services**, and then select **Library**.
2. Search for the API, click it, and then click **Enable**.

   ![Enable Places API key](/images/manual/use-cases/trek-enable-mapapi.png#bordered)

3. Repeat the same steps to enable all the above APIs.

### Step 3: Create API keys

1. In the left sidebar, hover over **APIs & Services**, and then select **Credentials**.
2. Click **Create credentials** at the top of the page, and then select **API key**.

   ![API key menu](/images/manual/use-cases/trek-create-api-key.png#bordered)

3. In the **Create API key** panel, configure the following settings:

   - **Name**: Enter a name to identify the API key. 
   - **Select API restrictions**: Select the APIs you just enabled, and then click **OK**.
   - **Authenticate API calls through a service account**: Do not select.
   - **Application restrictions**: Select **None**.

   ![Create API key settings in google cloud console](/images/manual/use-cases/trek-create-api-key-settings.png#bordered)

4. Click **Create**. 
5. In the **API key created** panel, note down your API key.
-->

## 使用双因素认证保护账户

通过为登录过程添加第二层安全保护来保护您的旅行数据。TREK 支持基于 TOTP 的双因素认证（2FA），使用 Google Authenticator 或 Authy 等应用。

### 启用 2FA

您可以选择仅保护您的个人账户，或为整个工作空间强制实施 2FA。在为所有工作空间成员要求 2FA 之前，您必须在自己的管理员账户上启用 2FA。

#### 为管理员账户启用 2FA

设置身份验证器应用，以在密码之外要求 6 位数字代码。

1. 登录 TREK，点击您的用户头像，然后选择 **Settings**。
2. 点击 **Account** 选项卡，然后找到 **Two-factor authentication (2FA)** 部分。
3. 点击 **Set up authenticator**。

   ![Enable 2FA for admin account](/images/manual/use-cases/trek-2fa-admin.png#bordered)

4. 使用您的身份验证器应用扫描 QR 码，在 TREK 中输入生成的 6 位数字代码，然后点击 **Enable 2FA**。
5. 保存屏幕上显示的备份代码，将其存放在安全的地方，然后点击 **OK**。

   ![Note down backup codes](/images/manual/use-cases/trek-2fa-backup-codes.png#bordered)

6. 退出 TREK，然后重新登录。
7. 输入您的凭据，然后点击 **Sign in**。
8. 输入来自身份验证器应用的验证码，然后点击 **Verify**。

   ![2FA authentication upon login](/images/manual/use-cases/trek-2fa-login.png#bordered)

#### 为成员启用 2FA

在您的管理员账户上启用 2FA 后，您可以为所有访问 TREK 工作空间的成员强制实施强制的 2FA 策略。

1. 在 TREK 中，点击用户头像，然后选择 **Admin**。
2. 点击 **Settings** 选项卡，找到 **Require two-factor authentication (2FA)** 面板，然后将其打开。

   ![Enable 2FA](/images/manual/use-cases/trek-2fa-enable.png#bordered)

   当工作空间成员登录时，TREK 会将他们引导到 **Settings** 页面，并显示消息：`Your administrator requires two-factor authentication. Set up an authenticator app below before continuing`。他们必须点击 **Set up authenticator**，扫描 QR 码，并配置他们的身份验证器应用，然后才能查看或编辑任何行程。

   ![Member enable 2FA notification](/images/manual/use-cases/trek-2fa-member-enable.png#bordered)

### 禁用 2FA

如果 2FA 是全局强制实施的，则移除 2FA 是一个两步过程。关闭工作空间要求不会自动删除成员的 2FA 配置。每个用户仍必须在自己的账户上手动禁用它。

#### 禁用工作空间范围的 2FA 要求

作为管理员，如果您全局强制实施了 2FA，则必须先关闭此要求，然后任何人（包括您自己）才能禁用个人 2FA。

1. 登录 TREK，选择您的用户头像，然后选择 **Admin**。
2. 选择 **Settings** 选项卡，找到 **Require two-factor authentication (2FA)** 部分，然后将其关闭。

      ![Disable 2FA for all users](/images/manual/use-cases/trek-2fs-disable-all.png#bordered)

#### 禁用个人账户的 2FA

从您的登录过程中移除 2FA 步骤。

:::info 工作空间限制
如果您是工作空间成员，在您的[管理员关闭全局要求](#禁用工作空间范围的-2fa-要求)之前，您无法禁用个人 2FA。
:::

1. 使用您的密码和当前 2FA 代码登录 TREK。
2. 选择您的用户头像，然后选择 **Settings**。
3. 选择 **Account** 选项卡，然后找到 **Two-factor authentication (2FA)** 部分。
4. 输入您的登录密码。
5. 输入来自身份验证器应用的验证码。
6. 选择 **Disable 2FA**。

      ![Disable 2FA](/images/manual/use-cases/trek-2fa-disable.png#bordered)

7. 退出 TREK，然后重新登录。验证步骤不再要求。

## 备份和恢复数据

定期保存您的行程、笔记和预算以防止数据丢失。TREK 允许您创建手动备份或设置自动备份。

### 备份工作空间数据

选择立即触发备份或设置定期计划。

1. 创建备份：

   - **手动备份**：点击用户头像，选择 **Admin**，进入 **Backup** 选项卡，然后点击 **Create Backup**。

      ![Manual backup](/images/manual/use-cases/trek-backup-manual.png#bordered)

   - **自动备份**：在 **Admin** > **Backups** > **Auto Backup** 设置下配置定期备份。

      ![Automatic backup](/images/manual/use-cases/trek-backup-auto.png#bordered)
2. 点击 **Download** 将备份包保存在本地。

### 从备份恢复数据

如果您需要恢复丢失的信息，可以上传之前保存的备份文件。

1. 在 **Backup** 页面上，点击 **Upload Backup**，然后选择您的本地包。
2. 当提示时，查看 **Restore Backup** 警告，然后确认恢复以用备份数据覆盖当前工作空间。
