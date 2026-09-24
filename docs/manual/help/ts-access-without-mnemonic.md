---
outline: [2, 3]
description: Identify the credential blocking access to LarePass, Olares Desktop, or your device when no mnemonic backup is available.
head:
  - - meta
    - name: keywords
      content: Olares, LarePass, mnemonic phrase, local password, login password, host password
---

# Recover access to LarePass, Olares Desktop, or your Olares device

Use this guide if your mnemonic phrase is unavailable and you need to regain access to LarePass, Olares Desktop, or the device running Olares. Start with the section that matches what you cannot access.

:::warning Keep every existing LarePass installation
Do not uninstall LarePass or remove your Olares ID from any device where it is still available. Without a mnemonic backup, doing either may permanently prevent you from using that Olares ID in LarePass again.
:::

## If you forgot your mnemonic phrase or never backed it up

Check every phone and computer where you previously used LarePass.

If your Olares ID is available in any LarePass client, keep that client installed and follow the steps for its platform.

### On mobile

1. Go to **Settings** > **LarePass Settings** > **Security** > **Mnemonic phrase**.
2. If prompted, enter your LarePass local password. If you do not know it, follow [If you forgot or have not set your LarePass local password](#if-you-forgot-or-have-not-set-your-larepass-local-password).
3. Reveal the 12 words, write them down in order, and store them offline.
4. Complete the in-app verification to confirm that the backup is correct.

### On desktop

1. Go to **Settings** > **Account** > **Manage Account**.
2. Locate your Olares ID and follow the prompts to view its mnemonic phrase.
3. Write down the 12 words in order and store them offline.

### If the original LarePass client is unavailable

If the LarePass client on which you created the Olares ID has been uninstalled or is no longer accessible, check whether the Olares ID is available in another LarePass client. If it is not and you have no mnemonic backup, the phrase cannot be recovered.

To use the device again, uninstall Olares, reinstall it, and activate it with a new Olares ID.

:::warning Uninstalling Olares deletes data
The uninstall command removes Olares components and data from the device. Back up any files you can still access before continuing.
:::

1. Open the Olares device terminal:

   - If you can sign in to Olares Desktop, open **Control Hub**, then select **Terminal** > **Olares**.
   - Otherwise, connect through SSH or log in locally with a monitor and keyboard.

2. Uninstall Olares:

   - In the Control Hub terminal, run:

     ```bash
     olares-cli uninstall
     ```

   - In an SSH or local terminal, run:

     ```bash
     sudo olares-cli uninstall
     ```

3. Wait for the uninstall process to finish.
4. Install or open LarePass, create a new Olares ID, and follow [Install Olares](../get-started/install-olares.md) to reinstall and activate the device.
5. Back up the new mnemonic phrase immediately after activation.

## If you forgot or have not set your LarePass local password

The local password unlocks protected LarePass features on the current device. Each LarePass installation has its own local password.

### On mobile

- If LarePass asks you to create a local password, create one and follow the prompts.
- If no prompt appears, go to **Settings** > **LarePass Settings** and set a local password.
- If you forgot the local password and biometric unlock is enabled, go to **Settings** > **LarePass Settings** > **Security**, tap **Reveal local password**, and complete biometric verification.
- If biometric unlock is not enabled, the local password cannot be revealed on that device. Use another LarePass client where your Olares ID is still accessible.

### On desktop

- If LarePass asks you to create a local password, create one and follow the prompts.
- If no prompt appears, go to **Settings** > **Security** and set a local password.
- If you forgot the local password, it cannot be revealed in the desktop client.

## If you forgot your Olares login password

Follow [Forgotten desktop login password](./ts-forget-login-password.md) to access the Olares device terminal and reset the password.

## If you cannot find the 2FA code

If the sign-in notification does not appear in LarePass, enter a 2FA code instead:

1. On the Olares login page, switch to code verification.
2. Find the current six-digit code in a LarePass client that contains your Olares ID:

   - On mobile, open **Vault**, which displays 2FA codes by default. Alternatively, go to **Settings**, find the **My Olares** card, and tap the authenticator.
   - On desktop, open **Vault**. The authenticator is the first item in the list.

3. Enter the code on the Olares login page before it expires.

## If you need the username, password, or IP address for the Olares device terminal

### Olares One

The system username for an activated Olares One is `olares`. The system password is used for SSH and local terminal sign-in.

#### Reset the system password from Olares Desktop

If you can sign in to Olares Desktop but do not know the system password, reset it using either method.

- **Control Hub**:

  1. Open **Control Hub**, then select **Terminal** > **Olares**.
  2. Run:

     ```bash
     passwd olares
     ```

  3. Enter the new password twice as prompted.

- **Settings**:

  1. Open **Settings**. On the **My Olares** page, select **My hardware**.
  2. Select **Reset SSH login password**.
  3. Enter a new password that meets the strength requirements, then click **OK**.
  4. Open LarePass and scan the QR code shown on the screen.
  5. Tap **Confirm** in LarePass. The new password is saved to Vault.

#### Find the current system password in LarePass

The system password is generated during activation and saved in LarePass Vault.

If you already know the system password, skip to step 3.

1. Open **Vault** in the LarePass mobile app. If a local-password prompt blocks access, follow [If you forgot or have not set your LarePass local password](#if-you-forgot-or-have-not-set-your-larepass-local-password).
2. Tap the filter in the top-left corner and select **All vaults**. Open the item with the terminal icon to view the system password.

   ![Select All vaults in LarePass](/images/one/ssh-switch-filter.png#bordered)

   ![Find the Olares One system password in Vault](/images/one/ssh-check-password-in-vault.png#bordered)

3. In LarePass, go to **Settings** > **System**, open the Olares One device card, and note the **Intranet IP** under **Network**.
4. On a computer connected to the same local network, run:

   ```bash
   ssh olares@<intranet-ip>
   ```

5. Enter the system password.

If SSH is unavailable, connect a monitor and keyboard to Olares One and log in with the same username and password. For details, see [Access Olares One terminal via SSH](/one/access-terminal-ssh.md) or [Access Olares One terminal physically](/one/access-physical-console.md).

If you cannot open Vault and do not know the system password, keep LarePass installed and [contact support](./request-technical-support.md). Do not send support your passwords or mnemonic phrase.

### Olares installed on your own device

Use the credentials that match how Olares was installed:

- **Olares ISO installed on dedicated hardware**: Use `olares` as both the system username and password.
- **Olares installed on an existing operating system**: Use the operating-system username and password you configured on that device.

Connect over SSH or log in locally with a monitor and keyboard. For details, see [Access the Olares terminal](../access-olares-terminal.md).
