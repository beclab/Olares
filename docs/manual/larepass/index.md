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

## Feature comparison

<table style="border-collapse: collapse; width: 100%; font-family: Arial, sans-serif;">
  <thead>
    <tr>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">Category</th>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">Features</th>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">Mobile</th>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">Desktop</th>
      <th style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top; background-color: #f4f4f4;">Chrome Extension</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td rowspan="4" style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: middle; font-weight: bold;">Account management</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Create Olares ID</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Import Olares ID</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Multi-account management</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">SSO login</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr>
      <td rowspan="4" style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: middle; font-weight: bold;">Device & network management</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Activate Olares</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">View resource consumption</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Remote device control</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Manage VPN connections</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr>
      <td rowspan="7" style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: middle; font-weight: bold;">Knowledge & file management</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Sync files across devices</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Manage files on Olares</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Collect webpage/video/podcast/PDF /eBook to Wise</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Download video/podcast/PDF/eBook to Files</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Add RSS feed subscription</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Immersive translation</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Backup your photos and files on phone</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td rowspan="5" style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: middle; font-weight: bold;">Secret management</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Generate, share, and autofill <br> strong passwords and passkeys</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">One-time authentication management</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Cookies Sync</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
    </tr>
    <tr>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">3rd-party SaaS account integration</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
    <tr style="background-color: #f9f9f9;">
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">Verifiable Credential (VC) card management</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">✅</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
      <td style="border: 1px solid #ddd; padding: 10px; text-align: left; vertical-align: top;">❌</td>
    </tr>
  </tbody>
</table>


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
