---
outline: [2, 3]
description: Understand Vault basics in Olares. Learn to set up vaults and manage vault items.
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, Vault, password manager, store credentials, secure sensitive data
---

# Secure passwords with Vault <Badge type="tip" text="15 min" />

Vault is the dedicated password and sensitive data manager in Olares. Use it to securely store passwords, keys, digital identities, and sensitive documents.

This guide covers the fundamentals of using Vault, including setting up your first vault, importing credentials, and managing your sensitive data efficiently.

## Learning objectives

- Initialize your personal Vault for the first time.
- Create and save a secure item. This guide uses Wi-Fi password as the example.
- Edit and organize your credentials.
- Locate items using search and filters.

## Before you begin

Familiarize yourself with related Vault concepts and how Vault organizes your data.

### Vault types

Olares Vault offers two main types of vaults:

* My vault: This is your private vault that is created automatically upon account activation. The vault is encrypted with your mnemonic and is visible only to you.
* Team vaults: These are collaborative vaults that are used to share credentials securely with team members or family.

### Vault items

A vault item is a secure container for a specific piece of information. While commonly used for login credentials, vault items can also store credit cards, secure notes, passports, and network details.

## Set up Vault

When you open Vault on your device for the first time, you must initialize it for security.

1. Open the Vault app from the Launchpad.

   ![Find Vault](/images/one/find-vault-app.png#bordered)

2. Set a local password for Vault, which is used to unlock Vault on the current device only. This local Vault password acts as a second layer of defense. 

   ![Open Vault for the first time](/images/one/open-vault.png#bordered){width=35%}

   :::tip Security best practice
   Do not use the same password as your Olares login. If someone guesses your login password, this secondary password ensures your sensitive data remains locked.
   :::

3. Click **Confirm**.

4. Import your Olares ID that is already linked to your server using your mnemonic phrase.

   ![Enter mnemonic phrase](/images/one/vault-enter-mnemonic-phrase.png#bordered){width=35%}

   :::tip
   For information about how to get your mnemonic phrase in case you do not have it handy, see [Reveal and back up your mnemonic phrase](../manual/larepass/back-up-mnemonics.md#reveal-and-back-up-your-mnemonic-phrase).
   :::

5. Click **Next**.

## Add vault items

Create a new vault item to store "my company Wi-Fi password".

1. Open the Vault app from the Dock or Launchpad.
2. Click <i class="material-symbols-outlined">add</i> in the **All vaults** panel.
3. From the **Select Vault** list, choose **My Vault** for private use.

   ![Select My Vault](/images/one/select-my-vault.png#bordered){width=50%}

4. Select **WIFI Password** from the category, and then click **Confirm**. 

   ![Add vault item Wi-Fi](/images/one/select-wifi-vault.png#bordered){width=50%}

   The details panel opens on the right. It displays the default fields relevant to the category you selected.

   ![Configure vault item settings for Wi-Fi](/images/one/new-vault-item-wifi.png#bordered){width=50%}

5. Fill in details in the fields.

   ![Add vault item Wi-Fi](/images/one/fill-info-vault.png#bordered){width=50%}

6. Click **Save**. The vault item is encrypted and added to the **All vaults** list.

## Manage vault items

Keep your vault items organized and up to date.

### Edit vault items

Modify an existing vault item to update details.

1. Open the Vault app.
2. In the **All vaults** panel, click the target vault item.
3. In the details panel on the right, click <i class="material-symbols-outlined">edit_note</i>.

   ![Edit vault item Wi-Fi](/images/one/edit-vault-item.png#bordered){width=50%}

4. Make changes as needed. For example, remove the expiration date.
5. Click **Save**.

### Mark favorite vault items

Mark frequently used vault items for quick access.

1. Open the Vault app.
2. In the **All vaults** panel, click the target vault item.
3. Click <i class="material-symbols-outlined">star_border</i> in the top right corner. 

   ![Mark favourite](/images/one/mark-favourite.png#bordered){width=50%}   

   The item is marked with a star in the **All vaults** panel.

   ![Favourite vault item](/images/one/favourite-vault-item.png#bordered){width=70%} 

## Find vault items

Locate vault items quickly using keyword search or filters.

### Search

In the **All vaults** panel, click <i class="material-symbols-outlined">search</i> and enter keywords (such as `Wi-Fi` or `company`) to find specific vault items.

![Search vault item](/images/one/search-vault.png#bordered){width=70%} 

### Filter

Use the left sidebar to narrow down the list:

- Vault category: Switch between My Vault and Team Vaults to change the scope.
- Tags: Click a tag name to view all related vault items.
- Favorites: Click **Favorites** to view all starred items.
- Recently used: Click **Recently used** to view your access history.
- Attachments: Click **Attachments** to view items containing files.

## Resources

- [Import vault items](../manual/olares/vault/vault-items.md#import)
- [Manage shared vaults](../manual/olares/vault/share-vault-items.md)
- [Autofill passwords with LarePass](../manual/larepass/autofill.md)
- [Generate two-factor authentication codes](../manual/larepass/two-factor-verification.md)

<!--<template #LarePass-desktop-or-mobile>

1. Open the LarePass desktop or mobile, and then go to the **Vault** tab.
2. Click <i class="material-symbols-outlined">add</i> in the top right corner.
3. From the **Select Vault** list, choose **My Vault** for private use or **Team Vault** for a shared one.
4. Select the category of the vault item, and then click **Confirm**.
5. Specify the detailed settings for the vault item, and then click **Save**.
</template>

<template #LarePass-browser-extension>

:::info Before you begin
- The LarePass browser extension is currently only available for Google Chrome. 
- Install it from the Chrome Web Store or the [official page](https://www.olares.com/larepass). 
- For quick access, pin the extension to your browser toolbar.
:::

1. Click the LarePass icon on your browser toolbar to open the sidebar.
2. Click **Vault**.
3. Click <i class="material-symbols-outlined">add</i> in the **All vaults** panel.
4. Select the category of the vault item, and then click **Confirm**.
5. Specify the detailed settings for the vault item, and then click **Save**.
</template>
</tabs>
![Add vault item](/images/one/new-vault-item.png#bordered){width=50%}
![Configure vault item settings](/images/one/configure-vault-item.png#bordered){width=50%}-->