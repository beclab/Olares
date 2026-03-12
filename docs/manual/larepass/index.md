---
description: User documentation for LarePass. Learn how to access and manage Olares through the LarePass client, including account management, file synchronization, device management, system upgrade, password management, and content collection.
outline: [2, 3]
---

# LarePass documentation

LarePass is the official cross-platform client software for Olares. It acts as a secure bridge between users and their Olares systems, enabling seamless access, identity management, file synchronization, and secure data workflows across all your devices, whether you're on mobile, desktop, or browser.

![LarePass](/images/manual/larepass/larepass.png)

## Key features
- Account and identity management
- Secure file access and sync
- Device and network management
- Password and secret management
- Knowledge collection

## Download LarePass

### iOS
Visit the [App Store product page](https://apps.apple.com/us/app/larepass/id6448082605) to download LarePass.

### Android
Visit the [Google Play product page](https://play.google.com/store/apps/details?id=com.terminus.termipass), or download the latest APK directly from the [LarePass website](https://www.olares.com/larepass).

### macOS & Windows
Download the latest desktop client from the [LarePass website](https://www.olares.com/larepass).

<!--### Chrome extension

The LarePass extension allows you to collect content and manage passwords directly from your browser. It currently supports Google Chrome only and must be installed manually.

:::warning Keep the extension folder
Your browser loads the extension from the folder you select. If you delete, move, or rename that folder, the extension will stop working.  
Extract the ZIP file to a permanent location, such as a folder under your user directory, rather than a temporary directory.
:::

1. Visit the [LarePass website](https://www.olares.com/larepass) and download the extension ZIP file.
2. Extract the ZIP file to a permanent folder on your computer.
3. In Chrome, go to `chrome://extensions/`.
4. Enable **Developer mode** in the top-right corner.
5. Click **Load unpacked** and select the extracted extension folder.

:::tip Quick access
After installation, click the puzzle icon in your browser toolbar and pin the LarePass extension for one-click access.
:::
-->
## Set up account 
- On mobile devices, you can [create an Olares ID](/manual/larepass/create-account.md#create-an-olares-id) directly in the app.
- On the desktop client<!-- or Chrome extension-->, you must [import an Olares account](/manual/larepass/create-account.md#import-an-account).

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
