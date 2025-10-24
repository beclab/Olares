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

![My Hardware](/images/manual/olares/my-hardware.png#bordered)

Available actions are:

- **Shutdown**

  Powers off the Olares device. You’ll be handed off to the **LarePass** app to confirm.  
  After shutdown, the device status in LarePass shows **Powered off**.  
  Remote operations are unavailable until you manually turn the device back on.

- **Restart**  
  Reboots the device with confirmation in **LarePass**.  
  The status changes to **Restarting** and returns to **Olares running** in about **5–8 minutes**.

<a id="reset-ssh"></a>
- **Reset SSH Password** <Badge type="tip" text="Olares One Only" />  
  
  Take the following steps to change the default SSH password:
  1. On the **My hardware** page, click the **Reset SSH Password** button.
  2. In the dialog, enter a new SSH password that meets all strength requirements, then click **OK**.
  3. Open the LarePass app and scan the QR code shown on the screen.
  4. Click **Confirm** on LarePass to finish.
- **Power mode** <Badge type="tip" text="Olares One Only" />
   
  Toggles Olares One’s performance profile.

    - **Silent mode** – Limits CPU and GPU power for quiet operation, suitable for everyday workloads.
    - **Performance mode** – Enables maximum CPU and GPU performance for demanding tasks such as AI inference or gaming.
  

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
