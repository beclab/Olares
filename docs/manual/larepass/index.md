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
- [Collect content via LarePass extension](/manual/olares/wise/basics.md#save-from-browser-with-larepass-extension)
- [Subscribe to RSS feeds](/manual/olares/wise/subscribe.md#use-larepass-browser-extension)

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

Download the latest version for your platform from the [LarePass website](https://www.olares.com/larepass).

- **iOS**: App Store
- **Android**: Google Play or direct download from the website
- **macOS and Windows**: Desktop client from the website

### Chrome extension

The LarePass extension allows you to collect content and manage passwords directly from your browser. It currently supports Google Chrome only and must be installed manually.

:::warning Keep the extension folder
Your browser loads the extension from the folder you select. If you delete, move, or rename that folder, the extension will stop working.  
Extract the ZIP file to a permanent location, such as a folder under your user directory, rather than a temporary directory.
:::

1. Visit the [LarePass website](https://www.olares.com/larepass) and download the extension ZIP file.
2. Extract the ZIP file to a permanent folder on your computer.
3. In your browser, go to `chrome://extensions/`.
4. Enable **Developer mode** in the top-right corner.
5. Click **Load unpacked** and select the extracted extension folder.

#### Sign in to the extension

1. Click the LarePass icon in your browser toolbar.
2. Select **Import an account**.
3. Enter your mnemonics and password to complete setup.

:::tip Quick access
After installation, click the puzzle icon in your browser toolbar and pin the LarePass extension for one-click access.
:::