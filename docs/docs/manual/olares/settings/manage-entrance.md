---
outline: [2, 3]
description: Learn how to manage application entrances in Olares, including setting up endpoints and creating access policies.
---

# Manage application entrances

**Entrances** define how users access your applications on Olares. For a deeper understanding, refer to the [Entrance concept](../../concepts/network.md#entrance) section.

Entrance management in Olares includes two main components:

* **Endpoints**: Define the network address and routing configuration for the application.
* **Access policies**: Control the authentication methods required to access the application.


## Access entrance management

To manage an application's entrances:

1. Go to **Settings** > **Application**.
2. Click on the target application from the list to enter the application page.
3. Click on the target entrance from the list to enter the **Entrances** settings page.
![Manage entrance](/images/manual/olares/app-entrance.png#bordered)

## Set up endpoints

The **Set up endpoints** page lets you customize how your application is accessed externally via a dedicated URL.

Options include:

- **Endpoint** – The domain for accessing your app. For example,`1870a8290.marvin112.olares.cn` (WordPress). You can click the copy icon to copy the URL.

**Default route ID** – The system-assigned identifier for the app route. For example, `1870a8290` for WordPress ).

- **Set custom route ID** – Click the **+** icon to replace the default route ID with a custom one. For example, `1870a8290.marvin112.olares.cn` → `wordpress.marvin112.olares.cn`.

- **Set custom domain** – Click the **+** icon to bind your own domain (e.g., app.yourdomain.com) to this application. You will need to follow DNS configuration steps to complete the setup.

For detailed instructions on changing an application's domain, refer to [Customize app domain](custom-app-domain.md).

## Create access policies

Access policies control who can access your application and their required authentication method. Options include:

* **Authentication Level**: Set the overall authentication requirement for the application:

    * **Public**: Accessible to anyone, with no login required.
    * **Private**: Requires users to log in to access.
    * **Internal**: No login is required if accessing the application via VPN.

* **Authentication mode**: Specify the method used for verifying user identity:

    * **System**: Inherits the system-wide authentication rules defined on the My Olares page.
    * **One Factor**: Requires only the Olares login password.
    * **Two Factor**: Requires the Olares login password plus a second verification code.
    * **None**: No authentication is required for access.

* **Sub-policies**: Apply fine-grained access rules to specific paths within the application using **regular expressions**.

  1. Click the **+ icon** in the **Policies** section.
  2. Define specific rules for desired paths.
  3. Choose the appropriate authentication method for each sub-policy.
