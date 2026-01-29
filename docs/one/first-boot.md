---
description: Learn how to set up your Olares One for the first time, including setting up hardware, installing the client app, creating your Olares account, connecting to your device, installing & activating the system, and logging into your Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, first boot, initial setup, first time use
---

# First boot 

This guide walks you through setting up your Olares One for the first time.

## Setup overview
- You do not need to connect a monitor, keyboard, or mouse to the Olares One. The entire configuration is handled via your mobile phone.
- The LarePass app is your primary tool for initial setup, authentication, and remote access management.

## Prerequisites
**Hardware**
- Olares One connected to Power.
- (Recommended) An Ethernet cable connecting Olares One to your router.

**Network**
- A reliable internet connection.
- Your mobile phone (iOS or Android) connected to the same network.

## Step 1: Power on and install LarePass
1. Power on Olares One. The status LED turns solid white, indicating the device is powered on.
2. Search for "LarePass" in the Apple App Store or Google Play Store. Install and open the app on your mobile device.
3. Follow the on-screen instructions to create your Olares ID. This unique identifier acts as your username across the entire Olares ecosystem.
   ![Create Olares ID](/images/one/create-olares-id.png#bordered){width=90%}

## Step 2: Connect Olares One
Once your ID is ready, you need to discover and link your Olares One.

<template>
<tabs #Set-up-via-wired-LAN>

1.  Ensure your Olares One is connected to your router via Ethernet.
2. In the LarePass app, tap **Discover nearby Olares**.
   ![Discover nearby Olares](/images/one/discover-nearby-olares.png#bordered){width=90%}  

3.  Select your Olares One from the list of available devices.
</tabs>

<tabs #Set-up-via-Wi-Fi-(Bluetooth)>
If wired access is not available, use Bluetooth to configure Wi-Fi credentials.

1. In the LarePass app, tap **Discover nearby Olares**.
2. Tap **Bluetooth network setup** at the bottom.
3. Select your device from the Bluetooth list and tap **Network setup**. 
4. Follow the prompts to connect Olares One to the Wi-Fi network your phone is currently using.
5. Once connected, return to the main screen and tap **Discover nearby Olares** again to find your device.

</tabs>
</template>

## Step 3: Install & activate Olares OS

1. In the LarePass app, on the device you just found, tap **Install now**.
2. When installation finishes, tap **Activate now** to initialize the system.
3. Select a reverse proxy node that is closest to your location and tap **Confirm**. The reverse proxy node acts as a secure gateway for remote access. Choosing the nearest node ensures the fastest connection speed and best stability.
4. Set the login password for Olares.
5. Copy or note down your personal desktop URL. You need this URL to access your Olares services.

    ![Obtain URL](/images/one/obtain-url.png#bordered)

6. Tap **Got it** to close the prompt.

## Step 4: Login to Olares Desktop
1. Open a web browser on your computer and go to your desktop URL.
2. On the login page, enter your login password.
3. You will be prompted to complete two-factor verification. Open LarePass to approve the login request, or manually enter the 6-digit verification code displayed in the app.
   ::: info
   The verification code is time-sensitive. Ensure you enter it before it expires.
   :::

## Next steps
Congratulations! Your Olares One is set up and active. It is strongly recommended to complete the following steps to secure your account and optimize your experience.

### Back up mnemonic phrase
:::warning Security warning
You are responsible for your security. Never share this phrase. If you lose these 12 words, you will permanently lose access to your Olares ID and all data stored in your Vault.
:::
Your Olares ID is secured by a unique 12-word mnemonic phrase. This phrase is the sole method to recover your account if you lose your phone or need to log in on a new device.

1. Open the LarePass app and go to **Settings** > **Safety**.
2. Tap **Mnemonic phrase** and verify your identity.
3. Tap **Click to view**.
4. Enter the local password as prompted.
5. Write the 12 words onto the **Recovery Sheet**, and then store the sheet in a secure, offline location.

### Access Olares services securely
For secure remote access without complex network configuration, it is recommended to enable the LarePass VPN.

See [Access Olares services securely using LarePass VPN](access-olares-via-vpn.md).

### Reset SSH password
If you plan to connect to your Olares One via terminal (SSH), you must update the default SSH password.

1. Open Olares Settings, on the **My hardware** page, click **Reset SSH password**.
2. In the dialog, enter a new SSH password that meets all strength requirements, then click **OK**.
3. Open the LarePass app and scan the QR code shown on the screen.
4. Click **Confirm** on LarePass to finish.

For details, see [Connect to Olares One via SSH](access-terminal-ssh.md).

### Explore
Olares OS comes with pre-installed system apps. You can also browse the **Market** to download additional applications that best suit your needs.

You can continue browsing this documentation to discover more use cases and advanced configuration scenarios.

## FAQs

### Why is an Olares ID required?

The Olares ID is currently required to automate secure remote access for your device. It allows the system to configure a reverse proxy, register a subdomain, and manage HTTPS certificates on your behalf. Without this, you would need to manually handle complex network configurations such as port forwarding and DNS management to access your device from outside your home.

Unlike a centralized cloud account, the Olares ID is owned entirely by you. We never see your credentials, and we cannot recover your data if you lose your mnemonic phrase.

We understand the community's preference for flexibility. In the upcoming March update, we plan to introduce new activation options that will make the Olares ID optional if you prefer to configure your own network access.

### Can I use my own domain name?

Yes. You can use your own custom domain instead of the default `olares.com` domain. Note that setting this up currently requires the LarePass app.

For details, refer to [Set up a custom domain for your Olares](../best-practices/set-custom-domain.md).

## Resources
- [Access Olares using local network](access-olares-via-vpn.md)
- [Olares ID](../developer/concepts/olares-id.md)
- [Back up mnemonic phrase](../manual/larepass/back-up-mnemonics.md)