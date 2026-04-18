---
outline: [2, 3]
description: Learn how to access the Olares One host terminal for command-line usage by logging in directly with a monitor and keyboard.
head:
  - - meta
    - name: keywords
      content: Olares One, terminal, physical console, monitor, keyboard
---

# Access Olares One terminal physically

If network access or SSH is unavailable, you can log in to the Olares One device physically using a monitor and keyboard.

## Prerequisites

- Your Olares One is set up and powered on.
- A monitor and keyboard connected to Olares One.
- If Olares OS is activated, a mobile device with the LarePass app installed is required to retrieve the login password from Vault.

## Step 1: Prepare your login password

:::info Not the same as your Olares Desktop password
This password logs you in to the Olares One host system. It's different from the password you use to sign in to the Olares Desktop in your browser.
:::

The default login password is `olares`. After you activate Olares OS, you will be prompted to reset the SSH password on the LarePass app, and the new password is automatically generated and saved to your Vault.

Identify your password based on your activation status.

<tabs>
<template #Unactivated-system>

If you have not activated Olares OS yet, use the default password `olares`.

</template>
<template #Activated-system>

If you have already activated Olares OS, obtain the saved password from the LarePass mobile app.

<!--@include: ./reusables-reset-ssh.md{9,17}-->

</template>
</tabs>

## Step 2: Log in

:::info
For security, characters will not appear on the screen as you type.
:::

1. In the text-based login prompt displayed on your connected monitor, type the username `olares`, and then press **Enter**.

    ```text
    olares login:
    ```

2. Type the password obtained in Step 1, and then press **Enter**.
