---
outline: [2, 3]
description: Understand Vault basics in Olares. Learn to set up vaults, manage vault items, organize sensitive data with tags, and protect your information with local passwords and encryption.
---

# Vault basics
This guide will walk you through the fundamentals of using Vault, from setting up your first vault to organizing your sensitive data efficiently.

## Understand Vault components

### Types of vault
Olares Vault offers two main types of vaults for users:

* Main vault (**My vault**): Automatically created upon account activation, serving as the user's private vault. It's encrypted with the user's mnemonic for maximum security.
* Shared vault (**Team vault**): These are collaborative vaults within Olares, allowing secure information sharing among team members or family.

### Vault items
Think of vault items as individual secure containers for your sensitive information. Each vault item consists of several components:

* **Name**: A title for easy identification
* **Tags**: For organization and quick filtering
* **Fields**: Various data types to store different kinds of information
* **History**: Tracks changes made to the item
* **Attachments**: For adding related files
* **Expiration**: Set an expiry date for time-sensitive information

Currently, Vault supports various default types for a vault item:
- Website/App
- Computer
- Credit card
- Bank account
- Wi-Fi Password
- Passport
- Authenticator
- Document

### Fields
Fields are the core component of Vault Items, allowing storage of various data types, including:

* Username
* Password
* Mnemonic
* Email address
* URL
* Date and month
* Credit card number
* Phone number
* PIN
* Plain text
* One-time password (OTP)

## Protect Vault with password
When use Vault in Olares for the first time, you will be prompted to set a local password. This password should not be the same as your Olares login password.

1. Set the local password for Vault.
2. Import the Olares ID that's already linked to your Olares server using mnemonic phrase. 

![Vault password](/images/manual/olares/vault-local-password.png)

:::info
If you don't know where to find the mnemonic phrase, see [back up mnemonic phrase](/manual/larepass/back-up-mnemonics).


## Manage vault items
:::tip
Keep your vault organized from the start by using descriptive names and relevant tags. This becomes especially valuable as your collection of secure items grows.
:::

### Add
To add a vault item:
<tabs>
<template #Olares>

1. Open the Vault app from the Dock or Launchpad.
2. Click <i class="material-symbols-outlined">add</i> in the top right corner.
3. Select a type, such as **Website / App** for storing account details, and click **Create**.

   ![Add vault item](/images/manual/olares/add-vault-item.png#bordered)
4. Fill in the required fields like item name, Username, Password, and URL.

   ![Fill item fields](/images/manual/olares/fill-item-fields.png#bordered)
5. Click **Save** to create the new vault item.
</template>
<template #LarePass-desktop-or-mobile>

1. Open LarePass on your device, and navigate to the **Vault** page within the app.
2. Click <i class="material-symbols-outlined">add</i> in the top right corner.
3. Select a type, such as **Website / App** for storing account details, and click **Create**.
4. Fill in the required fields like item name, Username, Password, and URL.
5. Click **Save** to create the new vault item.
</template>

<template #LarePass-browser-extension>

:::info
The LarePass browser extension is currently only available for Google Chrome. Visit the [official page](https://larepass.olares.com) to download the extension.
:::
:::tip
To easily access LarePass, you can pin the extension to your toolbar.
:::
1. Click the LarePass icon in your toolbar or extension menu to open LarePass on the right side of your browser window.
2. Navigate to the **Vault** page within the extension.
3. Click <i class="material-symbols-outlined">add</i> in the top right corner.
4. Select a type, such as **Website / App** for storing account details, and click **Create**.
5. Fill in the required fields like item name, Username, Password, and URL. The URL field is automatically populated with the current web address.
6. Click **Save** to create the new vault item.
</template>
</tabs>

### Import

You can import passwords from other popular password applications like 1Password and BitWarden. Supported formats include `text/plain`, `.csv`, `.pls`, `.set`, `.pbes2`, `.1pux`, and `.json`.

:::tip LarePass Desktop only
The import process must be completed using the LarePass desktop client.
:::

![Vault import](/images/manual/olares/vault-import.png)
1.  Open LarePass and click your avatar in the top-left corner.
2.  Navigate to **Settings** > **Import**, and click the <i class="material-symbols-outlined">upload_file</i> button.
3.  Select the file you want to import, such as `bitwarden_export.json`.
4.  Choose the destination Vault. The default is **My Vault**, but you can also select a team Vault you have created.
5.  Click **Confirm** to complete the import.

The imported items will be automatically tagged with the import file's name, making them easy to find and manage in your Vault.

### Edit
:::info
The LarePass browser extension does not support editing vault items. For full editing capabilities, please use the Vault app in Olares, the mobile, or desktop versions of LarePass.
:::
In the edit mode, you can:
- Update required fields.
- Add tags to items for easy organization and filtering.
- Set an expiration time.
- Add file attachments. Each file must not exceed 1 MB.
- View and restore a history item. Vault keeps up to 10 records for each item. When this limit is reached, older entries are discarded to make way for new ones.

To edit a vault item:
1. In Vault, select the vault item you need to edit. 
2. In its details window or page, click <i class="material-symbols-outlined">edit_note</i> in the top right corner to enter edit mode.
3. Make the necessary changes to the item's details.
4. Click **Save**.

### Favorite
Important items can be marked as favorites for quick access.
<tabs>
<template #Olares>

1. In Vault, click the vault item to open its details window on the right.
2. Click <i class="material-symbols-outlined">star_border</i> in the top right corner to mark this item as favorite.
</template>
<template #LarePass-desktop-or-mobile>

1. Open LarePass on your device, and navigate to the **Vault** page within the app.
2. Click the vault item to navigate to its details page.
3. Click <i class="material-symbols-outlined">star_border</i> in the top right corner to mark this item as favorite.
</template>
</tabs>



## Filter vault items
You can use quick filters or search box to locate the vault item you need.
### Quick filters
* By Main Vault & Shared Vault: Select **My Vault** or **Team Vault** to quickly find items within.
* By Tags: Click on tag names to locate tagged vault items easily.
* By Favorites: Click on **Favorites** to list all your favorite items.
* By Recently Used: Click on **Recently used** to display your recent items.
* By attachment: Click on **Attachment** to display all items with attached files.

### Keyword search
Click <i class="material-symbols-outlined">search</i> to directly search for target items using keywords.