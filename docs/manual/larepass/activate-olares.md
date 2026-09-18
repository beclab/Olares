---
outline: [2, 3]
description: Reactivate Olares with your existing Olares ID after reinstalling Olares, using the LarePass mobile app.
head:
  - - meta
    - name: keywords
      content: Olares, reactivate Olares, LarePass, reinstallation, Olares ID
---

# Reactivate Olares

If you have reinstalled Olares, the original instance becomes unavailable. You can reactivate the new installation using your existing Olares ID without having to create a new one.

:::warning Same network required
During reactivation, ensure that both your phone and the Olares device are connected to the same network so that LarePass can discover the device.
:::

Select the reactivation method based on how you reinstalled Olares:

<Tabs>
<template #Script-reinstallation>

If you used the one-line script to reinstall Olares and have already completed the initial setup wizard, follow these steps to reactivate with your existing Olares ID:

![Activate Olares](/images/manual/larepass/activate-olares1.png#bordered)

1. Open the LarePass app.
2. Tap **Scan QR code** to scan the QR code on the Wizard page.
3. Follow the on-screen instructions on LarePass to reset the login password for Olares.

After successful activation, the LarePass app returns to the home screen, and the Wizard redirects you to the login page.
</template>
<template #ISO-or-Docker-reinstallation>

If you reinstalled Olares using an ISO file or a Docker image, follow these steps to reactivate with your existing Olares ID:

1. Open the LarePass app on your phone. The error message "No active Olares found" appears.

    ![No active Olares found](/images/manual/larepass/no-active-olares-found.png#bordered)

2. Tap **Learn more** next to the message.
3. Select **Reactivate**.

    ![Reactivate Olares](/images/manual/larepass/reactivate-olares.png#bordered)

4. On your Olares activation page, tap **Discover nearby Olares**. LarePass will list the detected Olares instances in the same network.
5. Select the target Olares instance from the list and tap **Install now**.
6. When the installation completes, click **Activate now**.
7. In the **Select a reverse proxy** dialog, select a node that is closer to your geographical location. The installer will then configure HTTPs certificate and DNS for Olares.

    :::tip Note
    - You can change this setting later on the [Change reverse proxy](../olares/settings/change-frp.md) page in Olares.
    - If your Olares device is connected to a public IP network, this step will be skipped automatically.
    :::

8. Follow the on-screen instructions to reset the login password for Olares, and then tap **Complete**.

    ![Reset password](/images/manual/larepass/docker-reset-password.png#bordered)

Once the activation is completed, LarePass displays the desktop address of your Olares device, such as https://desktop.marvin123.olares.com.
</template>
</Tabs>
