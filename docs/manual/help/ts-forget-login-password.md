---
outline: [2, 3]
description: Reset the Olares desktop login password from the host terminal.
head:
  - - meta
    - name: keywords
      content: Olares, reset password, forgot login password, olares-cli, host terminal
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

:::warning Administrator access required
Run these commands only from the host terminal of an Olares device you administer. The fallback command changes a cluster-wide permission used by the password-reset API.
:::

### Step 1: Access the host terminal

Connect to your Olares device's terminal using one of the following methods:

- **SSH**: Open a terminal on another computer on the same local network, and run `ssh <username>@<device-ip>`.
- **Local login**: Connect a monitor and keyboard directly to the device and log in.

### Step 2: Reset the password

1. Run the reset command. Replace `<username>` with the local name of the Olares ID, without the domain. For example, use `alice123` for `alice123@olares.com`.

    ```bash
    olares-cli user reset-password <username> -p <new-password>
    ```

2. Check the result:

    - If the command succeeds, continue to [Step 3](#step-3-verify-login).
    - If it returns a permission-related error for the reset API, continue with the next step.
    - For any other error, do not change cluster permissions. Record the exact error and follow [Collect diagnostic information](../collect-diagnostic-information.md).

3. For a permission-related reset error only, enable the reset API permission:

    ```bash
    kubectl patch clusterrole backend:auth-provider --type='json' -p='[{"op": "add", "path": "/rules/0/nonResourceURLs/-", "value": "/cli/api/reset/*"}]'
    ```

4. Run the reset command again:

    ```bash
    olares-cli user reset-password <username> -p <new-password>
    ```

    For example, to reset the password for user "alice123" to "NewSecurePassword456!":

    ```bash
    olares-cli user reset-password alice123 -p NewSecurePassword456!
    ```

5. Confirm the result. You should see:

    ```text
    Password for user '<username>' reset successfully
    ```

### Step 3: Verify login

Wait about 10 seconds for the system to synchronize, then log in to your Olares desktop with the new password.
