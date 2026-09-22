---
outline: [2, 3]
description: Get the device terminal sign-in details and reset a forgotten Olares desktop login password.
head:
  - - meta
    - name: keywords
      content: Olares, reset password, forgot login password, olares-cli, device terminal
---

# Forgotten desktop login password

Use this guide to reset your Olares desktop login password from the host terminal.

## Condition

The Olares desktop shows "Authentication failed, incorrect password" when you try to log in.

## Cause

You have forgotten your Olares desktop login password.

## Solution

To reset your password, access the terminal of the device running Olares and run a few commands.

:::warning Administrator access required
Run these commands only from the host terminal of an Olares device you administer. The fallback command changes a cluster-wide permission used by the password-reset API.
:::

### Step 1: Get the terminal sign-in details

Use the instructions for your Olares device.

#### Olares One

The system username is `olares`. The system password is generated during activation and saved in LarePass Vault.

1. Open the LarePass mobile app, and tap **Vault**.
2. Enter your LarePass local password when prompted. If you do not know it, follow [If you forgot or have not set your LarePass local password](./ts-access-without-mnemonic.md#if-you-forgot-or-have-not-set-your-larepass-local-password).
3. Tap the filter in the top-left corner, and select **All vaults**.
4. Open the item with the terminal icon to view the system password.
5. Go to **Settings** > **System**, open the Olares One device card, and note the **Intranet IP** under **Network**.

#### Olares installed on your own device

Use the operating-system username and password configured on the device where you installed Olares. Use that device's local IP address for SSH. If you do not know the IP address or cannot use SSH, connect a monitor and keyboard and log in locally. For more terminal access options, see [Access the Olares terminal](../access-olares-terminal.md).

### Step 2: Access the device terminal

Connect to your Olares device's terminal using one of the following methods:

- **SSH**: Open a terminal on another computer on the same local network, and run `ssh <username>@<device-ip>`. For Olares One, run `ssh olares@<intranet-ip>`.
- **Local login**: Connect a monitor and keyboard directly to the device and log in.

### Step 3: Reset the password

1. Run the reset command. Replace `<username>` with the local name of the Olares ID, without the domain. For example, use `alice123` for `alice123@olares.com`.

    ```bash
    olares-cli user reset-password <username> -p <new-password>
    ```

2. Check the result:

    - If the command succeeds, continue to [Step 4](#step-4-verify-login).
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

### Step 4: Verify login

Wait about 10 seconds for the system to synchronize, then log in to your Olares desktop with the new password.
