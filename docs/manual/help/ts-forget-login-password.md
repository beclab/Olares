---
outline: [2, 3]
description: Reset login password for your Olares device using Olares CLI. 
---

# Forgot login password

Use this guide to regain access to your Olares desktop if you have forgotten your login password.

## Condition

You cannot log in to the Olares desktop and see the error message "Authentication failed, incorrect password".

## Cause

You have forgotten your Olares login password.

## Solution

To resolve this, you must access the Olares host terminal (via SSH or local CLI) to manually reset the login password.


### Step 1: Access the terminal via SSH

This is the most convenient way to access your Olares host terminal. If SSH is not accessible, skip to [Step 2](#step-2-log-in-locally).

:::info Same network required
Your computer and your Olares device must be on the same local network.
:::

1. Prepare your credentials based on your device type:

    | Details | Selfhosted | Olares One |
    |:--------|:-----------|:-----------|
    | Local IP address | The Intranet IP of your Olares device | The Intranet IP of your Olares device |
    | SSH username | The default `olares`, or<br> the custom one you reset previously | `olares` |
    | SSH password | The default `olares`, or<br> the custom one you reset previously | Located in LarePass Vault. <br>If your LarePass mobile client is not accessible, skip to [Step 2](#step-2-log-in-locally).|

    :::tip How to find the Olares One SSH password
    Open the LarePass mobile client, go to **Vault** > **All vaults**, find the one with the <span class="material-symbols-outlined">terminal</span> icon, and then tap it to reveal the password.
    ![Check saved SSH password in Vault for Olares One](/images/one/ssh-check-password-in-vault.png#bordered)
    :::

2. Connect to the host terminal.

    a. Open a terminal on your computer.

    b. Type the following command, replace `<local-ip-address>` with the Intranet IP, and then press **Enter**:

    ```bash
    ssh <system-username>@<local-ip-address>
    ```

    Take Olares One for example,
    ```bash
    ssh olares@192.168.11.12
    ```

    c. If prompted, type `yes` to confirm the connection, and then press **Enter**.
    
    d. When prompted, type the SSH password, and then press **Enter**.

    e. When you see the following prompt, which indicates a successful connection, skip to [Step 3](#step-3-reset-the-password). 

    ```text
    olares@olares:~$
    ```

### Step 2: Log in locally

If you cannot connect via SSH, log in to the device locally using a monitor and keyboard.

1. Connect a monitor and keyboard to your Olares device. A text-based login prompt is displayed on your screen automatically.

    ```text
    olares login:
    ```

2. Type the default username `olares` and press **Enter**.
3. Type the same password obtained in [Step 1](#step-1-access-the-terminal-via-ssh) and press **Enter**.

### Step 3: Reset the password

After you accessed the terminal, run the following commands to enable the reset permissions and update your password.

1. Type the following command, and then press **Enter**. This command allows the CLI to perform the reset operation.

    ```bash
    kubectl patch clusterrole backend:auth-provider --type='json' -p='[{"op": "add", "path": "/rules/0/nonResourceURLs/-", "value": "/cli/api/reset/*"}]'
    ```

2. Type the following command, and then press **Enter** to reset the password:

    ```bash
    olares-cli user reset-password <olares-id> -p <newpassword>
    ```

    For example, reset password for the user "alice123" to "NewSecurePassword456!":

    ```bash
    olares-cli user reset-password alice123 -p NewSecurePassword456!
    ```

3. Verify the reset result. 

    When the terminal returns the following message, it means the password is reset successfully:

    ```text
    Password for user '<olares-id>' reset successfully
    ```

### Step 4: Verify login

After you reset the password, wait about 10 seconds for the system services to synchronize the new credentials. Then return to your Olares desktop and log in with the new password.
