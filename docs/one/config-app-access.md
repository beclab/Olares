---
outline: [2, 3]
description: Configure authentication levels and models to control how users access applications on Olares One.
---

# Configure application access <Badge text="5 min"/>

Each app on Olares has an entrance that controls how users access it. You can configure the access policies for each entrance to match the app's sensitivity and your sharing needs.

## Before you begin

The access policies mainly consist of the following settings:

- **Authentication level**: Defines when authentication is required. The app can be private, accessible over VPN without login, or fully public.
- **Authentication model**: Defines how users authenticate.

| Authentication level | Authentication models | Access behavior |
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
   - **Private** for authenticated access only.
   - **Internal** for VPN access without login.
   - **Public** for access without login.

5. Select the **Authentication model**:
   - **System** to follow the system-wide authentication settings.
   - **One factor** or **Two factor** to apply authentication specifically to this app.
   - **None** for no authentication.

   :::warning
   **None** is available only when the **Authentication level** is set to **Public**. Use it carefully, especially for apps exposed to the public internet.
   :::

6. Click **Confirm** to save your changes.

## Resources

- [Entrance concept](/developer/concepts/network.md#entrance): Learn more about the technical background.
- [Activate custom domain name](/manual/olares/settings/custom-app-domain.md#custom-domain-name): Learn how to bind a custom domain to an app entrance.