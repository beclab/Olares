---
outline: [2, 3]
description: Troubleshoot and reset a forgotten login password for your Olares desktop. 
---

# Forgot login password

Use this guide to regain access to your Olares desktop if you have forgotten your login password.

## Condition

- Unable to log in to the Olares desktop via a browser.
- The error message "Authentication failed, incorrect password" appears repeatedly.

## Cause

The local authentication credentials stored in the Olares cluster do not match the input. 

## Solution

To resolve this, you must access the Olares host terminal (via SSH or local CLI) to manually reset the login password.

:::tip Hardware prerequisites
- Your Olares device is powered on and connected to a network.
- You have a client device (such as a computer) to access the terminal.
:::

### Step 1: Access the terminal via SSH

This is the most convenient way to access your Olares host terminal. If SSH is not accessible, skip to [Step 2](#step-2-log-in-locally).

1. Get the local IP address of your Olares device.

    a. Open the LarePass app, and go to **Settings** > **System** to navigate to the **Olares management** page.
    ![Tap the System card](/images/manual/get-started/larepass-system.png#bordered)
    
    b. Tap your device card: **Selfhosted** or **Olares One**.

    c. Scroll down to the **Network** section and note down the **Intranet IP**.

2. Retrieve the SSH password.

    - Self-host Olares: The default password is `olares` unless previously changed.
    - Olares One: In LarePass, go to **Vault** > **All vaults**, find the one with the <span class="material-symbols-outlined">terminal</span> icon, and then tap it to reveal the password.
            ![Check saved SSH password in Vault for Olares One](/images/one/ssh-check-password-in-vault.png#bordered)

3. Connect to the host terminal.

    a. Open a terminal on your computer.

    b. Type the following command, replace `<local_ip_address>` with the Intranet IP, and then press **Enter**:

    ```bash
    ssh olares@<local_ip_address>
    ```

    For example,
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

2. Type the username `olares` and press **Enter**.
3. Type the same SSH password obtained in [**Step 1**](#step-1-access-the-terminal-via-ssh) and press **Enter**.

### Step 3: Reset the password

After you accessed the terminal, run the following commands to enable the reset permissions and update your password.

1. Type the following command, and then press **Enter**. This command allows the CLI to perform the reset operation.

    ```bash
    kubectl patch clusterrole backend:auth-provider --type='json' -p='[{"op": "add", "path": "/rules/0/nonResourceURLs/-", "value": "/cli/api/reset/*"}]'
    ```

2. Type the following command, and then press **Enter** to reset the password:

    ```bash
    olares-cli user reset-password <username> --p <newpassword>
    ```

    For example, reset password for the user named "alice123":

    ```bash
    olares-cli user reset-password alice123 -p NewSecurePassword456!
    ```

3. Verify the reset result. 

    When the terminal returns the following message, it means the password is reset successfully:

    ```text
    Password for user '<username>' reset successfully
    ```

### Step 4: Verify login

After you reset the password, wait about 10 seconds for the system services to synchronize the new credentials. Then return to your Olares desktop and log in with the new password.
