---
outline: [2, 3]
description: Learn how to access the Olares One host terminal for command-line usage via local physical login.
head:
  - - meta
    - name: keywords
      content: Olares One, terminal, local login
---

# Access Olares One terminal physically

If network access or SSH is unavailable, you can log in to the Olares One device physically using a monitor and keyboard.

## Prerequisites

- A monitor and keyword connected to Olares One
- Olares One is powered on

## Step 1: Prepare your SSH password

The default SSH password is `olares`. However, right after you activate Olares, you will be prompted to reset the SSH password on the LarePass app, and the password is automatically generated and saved to your Vault.

Identify your password based on your activation status.

### Unactivated system

If your have not activated Olares OS yet, use the default password `olares`.

### Activated system

If you have already activated your Olares OS, obtain the saved SSH password from the LarePass mobile app. 

:::info Same network required
Your device and Olares One must be on the same local network.
:::

1. Open the LarePass mobile app, and then tap the **Vault** tab.
2. When prompted, enter your local password to unlock.
3. In the top-left corner, tap **Vault** to open the side navigation, and then tap **All vaults** to display all saved items.
4. Find the item with the <span class="material-symbols-outlined">terminal</span> icon and tap it to reveal the password.
       
    ![Check saved SSH password in Vault](/images/one/ssh-check-password-in-vault.png#bordered)

## Step 2: Log in

:::info
For security, characters will not appear on the screen as you type.
:::

1. In the text-based login prompt displayed on your connected monitor, type the username `olares`, and then press **Enter**.

    ```text
    olares login:
    ```

2. Type the SSH password obtained in Step 1, and then press **Enter**.
