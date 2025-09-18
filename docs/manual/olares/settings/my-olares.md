---
outline: [2, 3]
description: Learn how to manage your Olares account, devices, security settings, and network access policies in My Olares.
---

# My Olares overview

The **My Olares** page in **Settings** serves as your central hub for managing your Olares account, connected devices, security settings, and access policies.

To access My Olares, open **Settings** and click your avatar in the top-left corner.

![My Olares](/images/manual/olares/my-olares.png#bordered)

## My hardware

View and manage your Olares hardware. You can see details such as **Model**, **Device status**, **Device Identifier**, **CPU**, and **GPU**.

Available actions are:

* **Shutdown**: Click to shut down the Olares device. You will be directed to your LarePass app to finish the operation. After shutdown, Olares device status will show `Powered Off` on LarePass. Remote operations are unavailable until the device is manually turned on.
* **Restart**: Click **Restart**. Use LarePass to scan the QR code on the page and follow the instructions. Olares device status will show `Restarting` and return to `Olares running` in about 5–8 minutes.

## Olares Space

Check your subscribed plan details and usage in Olares Space, including reverse proxy solution, backup storage, and traffic consumption. Log in to Olares Space as prompted to use this feature.

## Support and feedback

Access help resources or send feedback to the Olares team regarding your system experience.

## Change password

Update your Olares login password to enhance your account security.

## Set network access policy

Define system-level access and authentication policies to control how users connect to your Olares.

* **Network visibility**: Control how visible your Olares services are on the network. Options include:
  * **Public**: Services are accessible to anyone on the internet.
  * **Protected**: Services are public but require a login for access.
  * **Private**: Services are not exposed to the internet and are only accessible via VPN.
* **Login security requirement**: Set the authentication method for logging into Olares.
  * **Two-factor** (Recommended): Requires both your login password and a two-factor authentication code for enhanced security.
  * **One-factor**: Only requires your login password (less secure).

## View login history

Review a detailed record of all login attempts to your Olares account. Each entry displays the **Time**, **Status**, **Source IP Address**, and the **Reason for the attempt**.

## Current Version

Check the current version of your Olares. If a new version is available, go to the **Settings > Olares management** page in the LarePass mobile client to complete the system upgrade.  
For detailed steps, see [Upgrade Olares](../../larepass/manage-olares.md#upgrade-olares).

## Device management

The **Devices** section allows you to view and manage all devices authorized to access your Olares. Each entry provides details about the connected device, such as its name, operating system, connection method, device IP address, and last connection time.

Available actions are:

* **View device information**: Click a specific device to view its detailed information such as software type, last active time, location, and token details.
* **Sign out device**: Use this option to remotely disconnect a device from your Olares. The device will need to reauthenticate to access Olares again.
