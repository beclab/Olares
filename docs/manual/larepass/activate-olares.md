---
outline: [2, 3]
description: Learn how to activate Olares for the first time, reactivate it after reinstallation, and complete secure two-factor login using the LarePass mobile app.
---

# Activate and log in to Olares

Olares uses your Olares ID and the LarePass mobile app to provide a secure and seamless authentication experience. This document guides you on how to activate Olares and complete two-factor verification during login using LarePass.

:::warning Same network required for admin users
To avoid activation failures, ensure that both your phone and the Olares device are connected to the **same network** when activating as an admin user.  
For member users, the same network requirement does **not** apply.  
:::

## Activate after one-line script installation

If you [installed Olares via the one-line script](../get-started/install-olares.md) and completed the initial setup in the wizard:

![Activate Olares](/images/manual/larepass/activate-olares1.png#bordered)

1. Open the LarePass app.
2. Tap **Scan QR code** to scan the QR code on the Wizard page. 
3. Follow the on-screen instructions on LarePass to reset the login password for Olares. 

After successful activation, the LarePass app returns to the home screen, and the Wizard redirects you to the login page.

## Activate after ISO installation

If you installed Olares via ISO on PVE or are using an Olares hardware device with ISO pre-installed:

<!--@include: ../get-started/install-and-activate-olares.md{9,25}-->

## Activate Olares using Bluetooth

Use this method if LarePass cannot find your Olares device. This can happen if Olares is not on a wired network or if your phone is on a different network.

By using Bluetooth, you can connect Olares directly to your phone's current Wi-Fi network and continue the activation process.
![Bluetooth network](/images/manual/larepass/bluetooth-network.png#bordered)

1. On the **Olares not found** page, tap **Bluetooth network setup**. LarePass will use your phone's Bluetooth to scan for the nearby Olares device.
2. When your device appears in the list, tap **Network setup**.
3. Select the Wi-Fi network your phone is currently connected to. If the network is password-protected, enter the password and tap **Confirm**.
4. Olares will begin connecting to the Wi-Fi network. Once the process is complete, a success message will appear. If you return to the Bluetooth network setup page, you'll see that Olares' IP address has changed to your phone's Wi-Fi subnet.
5. Go back to the device scan page and tap **Discover nearby Olares** to find your device and proceed with activation.

## Reactivate Olares after reinstallation

If you have reinstalled Olares, the original instance becomes unavailable. You can reactivate the new installation using your existing Olares ID without having to create a new one. 

Select the reactivation method based on how you reinstalled Olares:

<Tabs>
<template #Script-reinstallation>

If you used the one-line script to reinstall Olares and have already completed the initial setup wizard, follow these steps to reactivate with your existing Olares ID:

<!--@include: ../larepass/activate-olares.md{18,25}-->
</template>
<template #ISO-or-Docker-reinstallation>

If you reinstalled Olares using an ISO file or a Docker image, follow these steps to reactivate with your existing Olares ID:

1. Open the LarePass app on your phone. The error message "No active Olares found" appears.

    ![No active Olares found](/images/manual/larepass/no-active-olares-found.png#bordered)
  
2. Tap **Learn more** next to the message.
3. Select **Reactivate**.

    ![Reactivate Olares](/images/manual/larepass/reactivate-olares.png#bordered)

<!--@include: ../get-started/install-and-activate-olares.md{10,20}-->

8. Follow the on-screen instructions to reset the login password for Olares, and then tap **Complete**.

    ![Reset password](/images/manual/larepass/docker-reset-password.png#bordered)

Once the activation is completed, LarePass displays the desktop address of your Olares device, such as https://desktop.marvin123.olares.com.
</template>
</Tabs>

## Two-factor verification with LarePass

When you log in to Olares, you will be prompted to complete the two-factor verification. You can either confirm the login directly in LarePass app or manually enter a 6-digit verification code.

- **To confirm login on LarePass**:

  1. Open the login notification on your phone.
  2. In the message, click **Confirm** to complete the login process. 
    ![2FA](/images/manual/larepass/second-confirmation.png#bordered)

- **To manually enter the verification code**:
  1. On the Wizard page, select **Verify using one time password from LarePass**.
  2. Get the 6-digit code from one of the following:
      - **On your phone**: Open LarePass app and go to **Settings**. In the **My Olares** card, tap the authenticator to view the one-time verification code.
      ![OTP](/images/manual/larepass/otp-larepass1.png#bordered){width=95%}
      
      - **On your computer**: Open LarePass desktop client and go to **Vault**. Use the one-time verification code from the first item on the list.
      ![OTP desktop](/images/manual/larepass/otp-larepass-desktop1.png#bordered){width=95%}

  3. Return to your Wizard page and enter the code to complete the login.

::: tip Note
The verification code is time-sensitive. Ensure you enter it before it expires. If it expires, a new code will be generated automatically.
:::

After successful verification, you'll be redirected to the Olares desktop.