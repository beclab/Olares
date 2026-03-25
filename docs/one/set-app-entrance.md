---
outline: [2, 3]
description: Control who can access your apps on Olares One by configuring authentication levels and models.
---

# Set up app entrances <Badge text="5 min"/>

Each app on Olares has an entrance that controls how users access it. You can configure the authentication requirements for each entrance to match the app's sensitivity and your sharing needs.

## Before you begin

Understand the two settings you'll configure:

- **Authentication level**: Defines when authentication is required. The app can be private, accessible over VPN without login, or fully public.
- **Authentication model**: Defines how users authenticate.

| Authentication level | Available authentication models | Access behavior |
| -- | -- | -- |
| **Private** | **System**, **One factor**, **Two factor** | All users must authenticate before accessing the app. |
| **Internal** | **System**, **One factor**, **Two factor** | Authentication is bypassed only with LarePass VPN enabled. All other access requires authentication. |
| **Public** | **None** | Anyone can access the app without logging in. |

## Set the access policy

1. Go to **Settings** > **Applications**.
2. Select the target application.
3. In the **Entrances** section, click the entrance you want to configure.

   ![Set entrance](/images/one/settings-entrance.png#bordered){width=80%}

4. Under **Access policy**, select the **Authentication level**:
   - **Private**: Require login for all access.
   - **Internal**: Allow access without login over LarePass VPN; require login otherwise.
   - **Public**: Allow access without login.

5. Select the **Authentication model**:
   - **System**: Use the system-wide authentication rules.
   - **One factor**: Require the Olares login password.
   - **Two factor**: Require the login password plus a verification code.
   - **None**: No authentication. Only available when level is set to **Public**.

   :::warning
   Use **None** carefully, especially for apps exposed to the public internet.
   :::

6. Click **Confirm** to save your changes.

## Resources

- [Entrance concept](/developer/concepts/network.md#entrance): Learn more about the technical background.
- [Activate custom domain name](/manual/olares/settings/custom-app-domain.md#custom-domain-name): Learn how to bind a custom domain to an app entrance.