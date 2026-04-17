---
outline: [2, 3]
description: Learn how to manage application entrances in Olares, including setting up endpoints and creating access policies.
---

# Manage application entrances

**Entrances** define how users access your applications on Olares. For a deeper understanding, refer to the [Entrance concept](../../../developer/concepts/network.md#entrance) section.

Entrance management in Olares includes two main components:

* **Endpoint settings**: Define the network address and routing configuration for the application.
* **Access policies**: Control the authentication methods required to access the application.


## Access entrance management

To manage an application's entrances:

1. Go to **Settings** > **Application**.
2. Click the target application.
3. Under **Entrances**, click the target entrance.

    ![Manage entrance](/images/manual/olares/app-entrance1.png#bordered){width=90%}

## Set up endpoint

The **Endpoint settings** panel lets you customize how your application is accessed externally via a dedicated URL.

![Endpoint settings panel](/images/manual/olares/app-entrance-endpoint-panel.png#bordered){width=70%}

Options include:

- **Endpoint**: The domain for accessing your app. Click <i class="material-symbols-outlined">content_copy</i> to copy the URL.

- **Default route ID**: The system-assigned identifier for the app route. In this example, the default route ID for Jellyfin is `7e89d2a1`.

- **Set custom route ID**: Click <i class="material-symbols-outlined">add</i> to replace the default route ID. For example, if you set it to "jellyfin", the app will be available at both `https://7e89d2a1.alexmiles.olares.com` and `https://jellyfin.alexmiles.olares.com`.

- **Set custom domain**: Click <i class="material-symbols-outlined">add</i> to add your own domain to this application. For example, `app.yourdomain.com`. You need to configure the required DNS records before the domain can work. For detailed instructions, refer to [Customize application domains](custom-app-domain.md).

## Create access policies

Access policies control who can access your application and their required authentication method. 

![Access policies panel](/images/manual/olares/app-entrance-access-policy-panel.png#bordered){width=70%}

Options include:

* **Authentication Level**: Set the overall authentication requirement for the application:

    * **Public**: Accessible to anyone, with no login required.
    * **Private**: Requires users to log in to access.
    * **Internal**: No login is required if accessing the application via VPN.

* **Authentication mode**: Specify the method used for verifying user identity:

    * **System**: Inherits the system-wide authentication rules defined on the My Olares page.
    * **One Factor**: Requires only the Olares login password.
    * **Two Factor**: Requires the Olares login password plus a second verification code.
    * **None**: No authentication is required for access.

* **Manage sub-policies**: Apply fine-grained access rules to specific paths within the application using **regular expressions**.

  1. Click <i class="material-symbols-outlined">chevron_forward</i> to open the **Manage sub policies** page.
  2. Click **Add sub policy**, then enter the target paths in **Affected URLs** and select an **Authentication model**.
  3. Click **Submit**.