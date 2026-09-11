---
outline: [2, 3]
description: Learn how to manage application entrances in Olares, including setting up endpoints and creating access policies.
head:
  - - meta
    - name: keywords
      content: Olares, application access, authentication, access policy, entrance, custom domain, reverse proxy
---

# Manage application entrances

Entrances define how users access your applications on Olares. For more details, see the [Entrance](../../../developer/concepts/network.md#entrance) concept.

Each entrance consists of two parts:

* The **endpoint**: the URL used to access the app. To customize the endpoint, such as setting a custom route ID or domain, see [Customize application URLs](custom-app-domain.md).
* The **access policy**: controls who can access the app and which authentication method is required. This page covers access policies.

## Understand authentication levels and modes

Access policies control who can access your application and their required authentication method.

Use the following table to choose the right authentication level for each entrance:

| Authentication level | Available authentication modes | Access behavior |
| --- | --- | --- |
| **Public** | None | Anyone can open the app without logging in. |
| **Private** | System, One Factor, Two Factor | Everyone must authenticate before access. |
| **Internal** | System, One Factor, Two Factor | Users on LarePass VPN skip authentication; all other access requires it. |

* **Authentication level**: The overall authentication requirement for the entrance.

    * **Public**: Accessible to anyone, with no login required.
    * **Private**: Requires users to log in to access.
    * **Internal**: No login is required if accessing the application via VPN.

* **Authentication mode**: The method used for verifying user identity.

    * **System**: Inherits the system-wide authentication rules defined on the My Olares page.
    * **One Factor**: Requires only the Olares login password.
    * **Two Factor**: Requires the Olares login password plus a second verification code.
    * **None**: No authentication is required for access.

## Set the authentication level and mode

1. Go to **Settings** > **Application**.
2. Click the target application.
3. Under **Entrances**, click the target entrance.

    ![Manage entrance](/images/manual/olares/app-entrance1.png#bordered){width=90%}

4. Under **Access policies**, set the **Authentication level** and **Authentication mode**.

    ![Access policies panel](/images/manual/olares/app-entrance-access-policy-panel.png#bordered){width=70%}

5. Click **Submit**.

## Manage sub-policies

Use sub-policies to apply fine-grained access rules to specific paths within the application using **regular expressions**.

1. On the **Access policies** panel, click <i class="material-symbols-outlined">chevron_forward</i> to open the **Manage sub policies** page.
2. Click **Add sub policy**, then enter the target paths in **Affected URLs** and select an **Authentication mode**.
3. Click **Submit**.

## Resources

- [Customize app domain](custom-app-domain.md): Use your own domain to access an app.
- [Entrance concept](../../../developer/concepts/network.md#entrance): Learn the technical background.
