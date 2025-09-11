---
description: User documentation for LarePass. Learn how to access and manage Olares through the LarePass client, including account management, file synchronization, device management, system upgrade, password management, and content collection.
outline: [2, 3]
---

# LarePass documentation

LarePass is the official cross-platform client software for Olares. It acts as a secure bridge between users and their Olares systems, enabling seamless access, identity management, file synchronization, and secure data workflows across all your devices, whether you're on mobile, desktop, or browser.

![LarePass](/images/manual/larepass/larepass.png)


## Key features

### Account & identity management
Create and manage your Olares ID, connect integrations with other services, and back up your credentials securely.
- [Create an Olares ID](create-account.md)
- [Back up mnemonics](back-up-mnemonics.md)
- [Set or reset local password](back-up-mnemonics.md#set-up-local-password)
- [Manage integrations](integrations.md)

### Secure file access & sync
- [Manage files with LarePass](manage-files.md)

### Device & network management
Activate and manage Olares devices, and securely connect to Olares via LarePass VPN.
- [Activate your Olares device](activate-olares.md)
- [Upgrade Olares](manage-olares.md#upgrade-olares)
- [Log in to Olares with 2FA](activate-olares.md#two-factor-verification-with-larepass)
- [Manage Olares](manage-olares.md)
- [Switch networks](manage-olares.md#switch-from-wired-to-wireless-network)
- [Enable VPN for remote access](private-network.md)

### Password & secret management
Use Vault to autofill credentials, store passwords, and generate 2FA codes across devices.
- [Autofill passwords](/manual/larepass/autofill.md)
- [Generate 2FA codes](/manual/larepass/two-factor-verification.md)

### Knowledge collection
Use LarePass to collect web content and follow RSS feeds.
- [Collect content via LarePass extension](manage-knowledge.md#collect-content-via-the-larepass-extension)
- [Subscribe to RSS feeds](manage-knowledge.md#subscribe-to-rss-feeds)

---

## Download and install LarePass

Get the latest version for your device at the [LarePass website](https://larepass.olares.com).

### Install the LarePass browser extension

<tabs>
<template #Install-from-Chrome-Web-Store>

1. Search for **LarePass** in the [Chrome Web Store](https://chrome.google.com/webstore).
2. Open the details page and click **Add to Chrome**.
3. Log into the LarePass extension by importing your Olares ID:
   - Open the LarePass extension, and click **Import an account**.
   - Enter the mnemonics for your Olares ID.
   - Enter your Olares password to complete login.

</template>

<template #Install-offline>

1. Visit [larepass.olares.com](https://larepass.olares.com) and download the extension ZIP file.
2. Go to `chrome://extensions/` in your browser.
3. Enable **Developer mode** in the top-right corner.
4. Click **Load unpacked** and select the extracted LarePass extension folder.
5. Log in:
   - Open the LarePass extension, and click **Import an account**.
   - Enter the mnemonics for your Olares ID.
   - Enter your Olares password to complete login.
</template>
</tabs>

  :::tip Quick access
  After installation, pin the LarePass extension from Chrome’s extension menu for one-click access.
  :::
