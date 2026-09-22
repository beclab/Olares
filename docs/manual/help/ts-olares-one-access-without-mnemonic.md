---
outline: [2, 3]
description: Identify the password blocking access to Olares One, find the host password in LarePass, and recover access when your mnemonic phrase is unavailable.
---

# Cannot access Olares One after forgetting the mnemonic phrase

Use this page if you have forgotten your 12-word mnemonic phrase and cannot log in to Olares One. The phrase is used to recover your Olares ID. It is not the password for the Olares One host. Start in LarePass and follow the path that matches what you see.

:::warning
Do not uninstall LarePass or delete the Olares ID inside it until you have saved the mnemonic phrase elsewhere. Removing the app or ID could delete your only copy.
:::

## Condition

- You cannot recall the mnemonic phrase for your Olares ID and do not have a separate backup.
- A password prompt is blocking access to Olares One or its desktop, and you are unsure which password it requires.
- Your Olares ID is still available in LarePass on your phone.

## Cause

These credentials have different uses. Forgetting the mnemonic phrase does not change the other passwords.

| Credential | Used for |
| --- | --- |
| Mnemonic phrase | Recovering or importing your Olares ID in LarePass |
| LarePass local password | Unlocking protected functions in LarePass on your phone |
| Olares One host password | Logging in through SSH or the physical console |
| Olares login password | Signing in to the Olares desktop |

## Solution

### Step 1: Open Vault in LarePass

:::info
After six incorrect password attempts, LarePass locks the account for 15 minutes and disables biometric unlock during that time. Stop trying passwords, close LarePass, and wait 15 minutes before continuing.
:::

If you already know the Olares One host password, skip to [Step 3](#step-3-log-in-to-the-olares-one-host). Otherwise, open LarePass on your phone, tap **Vault**, and follow the path that matches what you see:

- **Vault opens without asking for a password:** Continue to [Step 2](#step-2-find-the-olares-one-host-password).

- **Vault asks for a password and you remember it:** Enter the LarePass local password, then continue to [Step 2](#step-2-find-the-olares-one-host-password).

- **Vault asks for a password you have forgotten:** Tap the face or fingerprint icon to try biometric unlock.

  When enabled, LarePass uses the local password stored in your phone's secure keystore to unlock Vault after biometric verification.

  ![Unlock LarePass with biometrics](/images/manual/help/olares-one-biometric-verification.png#bordered)

  - If Vault opens, continue to [Step 2](#step-2-find-the-olares-one-host-password).
  - If Vault does not open, keep LarePass installed. If biometric unlock was never enabled, the forgotten local password cannot be displayed or reset. If you already know the Olares One host password, continue to [Step 3](#step-3-log-in-to-the-olares-one-host). Otherwise, [contact support](./request-technical-support.md).

:::info
After unlocking Vault, you do not need the local password to continue. To view it for future use, update LarePass to the latest version, then go to **Settings** > **LarePass Settings** > **Safety** > **Local password** and complete biometric verification.
:::

### Step 2: Find the Olares One host password

1. In **Vault**, tap the filter in the top-left corner and select **All vaults**.

   ![Select All vaults in LarePass](/images/manual/help/olares-one-vault-filter.jpg#bordered)

2. Open the item with the terminal icon. It contains the Olares One host password generated during activation.

   ![Find the terminal password item in Vault](/images/manual/help/olares-one-host-password.jpg#bordered)

### Step 3: Log in to the Olares One host

Choose either method:

- **SSH:** In LarePass, tap **Settings**. Under **My Olares**, tap **System**, then open the device card and find **Intranet IP** under **Network**.

  ![Open System from LarePass Settings](/images/manual/help/olares-one-system-settings.jpg#bordered)

  ![Find the Intranet IP in device information](/images/manual/help/olares-one-intranet-ip.jpg#bordered)

  On a computer connected to the same local network, run:

  ```bash
  ssh olares@<intranet-ip>
  ```

- **Physical console:** Connect a monitor and keyboard to Olares One. At the login prompt, enter `olares` as the username.

For either method, enter the Olares One host password when prompted.

Once you reach the host terminal, you have logged in to your Olares One. If you can also sign in to the desktop, skip Step 4 and continue to Step 5.

### Step 4: Reset the desktop login password if needed

If you can already sign in to the Olares desktop, skip this step. If you also forgot its login password, run these commands from the Olares One host terminal:

```bash
kubectl patch clusterrole backend:auth-provider --type='json' \
  -p='[{"op": "add", "path": "/rules/0/nonResourceURLs/-", "value": "/cli/api/reset/*"}]'

olares-cli user reset-password <olares-id> -p '<new-password>'
```

For `<olares-id>`, use the part of your Olares ID before `@`. For example, use `alice123` for `alice123@olares.com`.

After the reset succeeds, wait about 10 seconds, then sign in to the desktop with the new password.

### Step 5: Back up the mnemonic phrase if available

If you can still unlock LarePass and reveal the phrase, go to **Settings** > **LarePass Settings** > **Safety** > **Mnemonic phrase**. Write down all 12 words in order and keep them offline. Logging in to the host or resetting the desktop password does not reveal the phrase.

If a step does not match what you see, [contact support](./request-technical-support.md). Tell them which screen rejects you, whether Vault opens, and whether you can access the host terminal. Do not send your passwords or mnemonic phrase.
