---
outline: [2, 3]
description: Reset the Olares desktop login password from the host terminal.
---

# Forgotten desktop login password

Use this guide to reset your Olares desktop login password from the host terminal.

## Condition

The Olares desktop shows "Authentication failed, incorrect password" when you try to log in.

## Cause

You have forgotten your Olares desktop login password.

## Solution

To reset your password, access the host terminal of your Olares device and run a few commands.

:::info
You need the following information about your Olares device:
- The local IP address
- Your device's username and password
:::

### Step 1: Access the host terminal

Connect to your Olares device's terminal using one of the following methods:

- **SSH**: Open a terminal on another computer on the same local network, and run `ssh <username>@<device-ip>`.
- **Local login**: Connect a monitor and keyboard directly to the device and log in.

### Step 2: Reset the password

1. Enable the reset permission:

    ```bash
    kubectl patch clusterrole backend:auth-provider --type='json' -p='[{"op": "add", "path": "/rules/0/nonResourceURLs/-", "value": "/cli/api/reset/*"}]'
    ```

2. Run the reset command:

    ```bash
    olares-cli user reset-password <olares-id> -p <new-password>
    ```

    For example, to reset the password for user "alice123" to "NewSecurePassword456!":

    ```bash
    olares-cli user reset-password alice123 -p NewSecurePassword456!
    ```

3. Confirm the result. You should see:

    ```text
    Password for user '<olares-id>' reset successfully
    ```

### Step 3: Verify login

Wait about 10 seconds for the system to synchronize, then log in to your Olares desktop with the new password.
