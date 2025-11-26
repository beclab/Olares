---
description: Configure and use LarePass autofill features to securely manage passwords across devices, automatically save credentials, and streamline your login experience.
---
# Autofill passwords with LarePass

Password autofill eliminates the hassle of manually typing credentials while maintaining security. With LarePass, you can securely store passwords in your Vault and automatically fill them across your devices.

## Before you begin

Make sure you have LarePass mobile clients or Chrome extension installed on your device, and logged in using your Olares ID.

:::tip
For different download options of LarePass, visit the [official website](https://www.olares.com/larepass).
:::

## Enable autofill service
<tabs>
<template #Android>

1. Open LarePass, and go to **Settings** > **Autofill**.
2. Turn on Autofill, and select LarePass as your autofill provider.
3. Review and accept the security note when prompted.
</template>
<template #iOS>

Due to iOS system restrictions, you have to manually enable autofill for LarePass:

1. Open the Settings app on your iOS device.
2. Use the search feature to quickly find the autofill settings.
3. Ensure the Autofill service is on, then activate LarePass as an autofill provider.

</template>
<template #Chrome-extension>

Autofill is automatically enabled upon logging in with the browser extension.
</template>
</tabs>

## Save password
When you enter credentials in an app or website, LarePass will detect this action and prompt you to save them.
:::info
On iOS, passwords cannot be automatically saved. You can manually add a vault item or use the Chrome extension. Vault items sync across all platforms.
:::
1. Log in to an app or website.
2. When prompted, click **Save** to store your password in LarePass.
3. In the details page or window, enter a name for this vault item, and click **Save**.

## Use autofill

<tabs>
<template #Android>

1. Open an app or website where you aren't logged in.
2. Tap the username or password field.
3. In the overlay popup, tap **Autofill with LarePass**.
4. Unlock Vault to access your saved credentials.
5. Select the matching vault item to autofill your login details.
</template>
<template #iOS>

1. Open an app or website where you aren't logged in.
2. Tap the username or password field. A keyboard will slide up with a matching login, or with a **Password** option.
3. If a matching login is displayed, tap it to autofill.
4. If the **Password** option is displayed, tap it and unlock Vault to access available vault items for the login.
   :::info
   If other autofill services like iCloud Keychain are active, select **LarePass** in the provider list.
   :::
5. Select the matching vault item to autofill your login details.
</template>
<template #Chrome-extension>

1. Open a website where you aren't logged in.
2. Click the LarePass icon in the text field.
3. In the overlay popup, select the matching login to autofill your login details.
4. If no credentials are saved for this site, select **New item** to add a new vault item.
</template>
</tabs>



