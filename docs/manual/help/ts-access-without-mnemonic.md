---
outline: [2, 3]
description: Recover access to Olares when your mnemonic phrase is unavailable.
---

# Cannot access Olares after forgetting the mnemonic phrase

Use this page if you have forgotten your 12-word mnemonic phrase and cannot log in to Olares. The phrase recovers your Olares ID and it is not the password for signing in to Olares or its host operating system.

:::warning
Do not uninstall LarePass or delete the Olares ID inside it until you have saved the mnemonic phrase elsewhere. Removing the app or ID could delete your only copy.
:::

## Condition

- You cannot recall the mnemonic phrase for your Olares ID and do not have a separate backup.
- You cannot sign in to Olares and need to regain access to the host or reset the Olares login password.

## Solution

To regain access, sign in to the Olares host over SSH, then reset the Olares login password if needed. If SSH is unavailable, use the physical console.

- **Olares installed on your own device:** Use the host username and password configured for that device, then go to [Step 3](#step-3-access-the-host-terminal).
- **Olares One:** Find the activation-generated SSH password in LarePass Vault using Steps 1 and 2. If you already know the password, go directly to [Step 3](#step-3-access-the-host-terminal).

### Step 1: For Olares One, open Vault in LarePass

Open LarePass on your phone.

:::info
After six incorrect password attempts, LarePass locks the account for 15 minutes and disables biometric unlock during that time. Stop trying passwords, close LarePass, and wait 15 minutes before continuing.
:::

Tap **Vault** and follow the path that matches what you see:

- **Vault opens without asking for a password:** Continue to [Step 2](#step-2-find-the-olares-one-ssh-password).

- **Vault asks for a password and you remember it:** Enter the LarePass local password, then continue to [Step 2](#step-2-find-the-olares-one-ssh-password).

- **Vault asks for a password you have forgotten:** Tap the face or fingerprint icon to try biometric unlock.

  When enabled, LarePass uses the local password stored in your phone's secure keystore to unlock Vault after biometric verification.

  ![Unlock LarePass with biometrics](/images/manual/help/olares-one-biometric-verification.png#bordered)

  - If Vault opens, continue to [Step 2](#step-2-find-the-olares-one-ssh-password).
  - If Vault does not open, keep LarePass installed. If biometric unlock was never enabled, the forgotten local password cannot be displayed or reset. If you know the Olares One host password, continue to [Step 3](#step-3-access-the-host-terminal). Otherwise, [contact support](./request-technical-support.md).

:::info
After unlocking Vault, you do not need the local password to continue. To view it for future use, update LarePass to the latest version, then go to **Settings** > **LarePass Settings** > **Safety** > **Local password** and complete biometric verification.
:::

### Step 2: Find the Olares One SSH password

1. In **Vault**, tap the filter in the top-left corner and select **All vaults**.

   ![Select All vaults in LarePass](/images/manual/help/olares-one-vault-filter.jpg#bordered)

2. Open the item with the terminal icon to view the Olares One SSH password.

   ![Find the terminal password item in Vault](/images/manual/help/olares-one-host-password.jpg#bordered)

### Step 3: Access the host terminal

Use the host account for your installation. You can connect through SSH if the host allows it, or sign in at the physical console.

**SSH**

1. In LarePass, tap **Settings**. Under **My Olares**, tap **System** and open the device card. Find the **Intranet IP** using the path for your version:

   - **LarePass 1.11.56 or later:** Tap **Node** > **Network** > **Intranet IP**.
   - **Earlier versions:** On the device card, go to **Network** > **Intranet IP**.

2. On a computer connected to the same local network, run:

   ```bash
   ssh <username>@<intranet-ip>
   ```

3. Enter the host password when prompted.

:::info Olares One SSH login
On Olares One, use `olares` as the SSH username: `ssh olares@<intranet-ip>`. The SSH password generated during activation is saved in Vault and can be found in [Step 2](#step-2-find-the-olares-one-ssh-password).
:::

**Physical console**

Connect a monitor and keyboard to the host, then sign in with its operating-system account.

Once you reach the host terminal, you have access to the device running Olares. If you can also sign in to the Olares desktop, skip Step 4 and continue to Step 5.

### Step 4: Reset the desktop login password if needed

If you can already sign in to the Olares desktop, skip this step. If you also forgot its login password, run these commands from the host terminal:

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
