---
outline: [2, 3]
description: Configure how apps on Olares One are accessed by setting up endpoints and defining access policies.
---

# Set up app entrances <Badge text="10 min"/>

App **entrances** define how users access applications installed on your Olares One. You can customize the app’s external URL and control who can access it by configuring endpoints and access policies.

An app entrance consists of:

- **Endpoint**: Defines the domain and routing configuration used to access the app.
- **Access policy**: Defines authentication requirements and access control rules.

## Learning objectives 
By the end of this tutorial, you will learn how to:

- Use a custom route ID for better readability.
- Bind a custom domain to access your app with your own URL.
- Enable two-factor authentication for admin or sensitive apps.
- Use sub-policies to separate public and restricted content within the same app.

## Access entrance settings

To configure an app entrance:

1. Go to **Settings** > **Applications**.
2. Select the target application.
3. In the **Entrances** section, click the entrance you want to configure.

![Set entrance](/images/one/settings-entrance.png#bordered){width=80%}

## Configure endpoints

The **Endpoint settings** section controls the external URL used to access your app. You can use the system default or customize it for a cleaner and more recognizable URL.

### Manage route ID

Every app is automatically assigned a default endpoint in the following format:

`<route-id>.<your-domain>.olares.com`

The route ID is the first part of the domain. You can replace it with a custom value to make the URL easier to read and share.

To set a custom route ID:

1. Click <i class="material-symbols-outlined">add</i> next to **Set custom route ID**.
2. Enter your preferred identifier.
3. Click **Confirm** to save your changes.

:::info Unique name
Custom route IDs must be unique within your Olares domain.
:::

**Example: Steam Headless service**

Changing the route ID from `139ebc4f0` to `steam` makes the endpoint easier to recognize.

![Set custom route ID](/images/one/settings-custom-route.png#bordered){width=60%}

| Route ID type | Route ID | URL |
|--|--|--|
| Default | `139ebc4f0`| `https://139ebc4f0.olaresdemo.olares.com`|
| Custom | `steam` |`https://steam.olaresdemo.olares.com`|

![Use custom route ID for Steam](/images/one/settings-entrance-steam.png#bordered){width=75%}

### Bind a custom domain

Bind a third-party domain to access the app using your own URL.

Before you start, make sure you have:

- A domain name you control.
- An HTTPS certificate and matching private key for that domain.
- Set the authentication level to **Public** or **Internal**.

#### Step 1: Add the domain information

1. Click <i class="material-symbols-outlined">add</i> next to **Set custom domain**.
2. In the pop-up window, enter your domain name, paste the HTTPS certificate, and paste the private key.
3. Click **Confirm**.
![Set custom domain](/images/one/settings-entrance-custom-domain.png#bordered){width=60%}

#### Step 2: Activate the third-party domain

Adding the domain does not enable it immediately. Activate it to apply the changes.

1. Click **Activation** next to **Activate third-party domain**.
    ![Activate custom domain](/images/one/settings-activate-domain.png#bordered){width=60%}
2. In the pop-up window, follow the instructions to create a CNAME record with your domain hosting provider.
3. Return to Olares and wait for activation to complete. You can monitor the activation **status** in the interface.

Once activated successfully, users can access the application using the custom domain.

## Manage access policies

Access policies define who can access your app and how they authenticate.

### Authentication level

Authentication level sets the overall access level for the app:

- **Public**: Anyone can access the app without logging in.
- **Private**: Users must log in with an Olares account.
- **Internal**: Accessible without login when connected via VPN. External access requires authentication.

### Authentication mode

Specify how users verify their identity:

- **System**: Inherits the system-wide authentication rules defined on the My Olares page.
- **One factor**: Requires the Olares login password.
- **Two factor**: Requires the Olares login password plus a second verification code.
- **None**: No authentication required.

:::warning Caution
Use **None** carefully, especially for apps exposed to the public internet.
:::

### Add sub-policies

You can create fine-grained access rules for specific paths.

To add a sub-policy:

1. Click <i class="material-symbols-outlined">chevron_forward</i> next to **Manage sub policies**.
2. Click **Add sub policy**.
3. Enter the path.
4. Select the required authentication mode.
5. Save the configuration.

Sub-policies allow you to protect sensitive areas of your application while keeping other sections publicly accessible.

## Resources

- [Entrance concept](/developer/concepts/network.md#entrance): Learn more about the technical background.
- [Activate custom domain name](/manual/olares/settings/custom-app-domain.md#custom-domain-name): Learn how to activate a custom domain name in detail.