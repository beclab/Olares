---
outline: [2, 3]
description: Learn how to securely back up your Olares ID mnemonic phrase for account recovery and protection.
---

# Back up mnemonic phrase
A mnemonic phrase is a sequence of 12 words that serves as the sole method to recover your Olares ID. If you lose your devices or need to import your Olares account, you can use the phrase to regain your access to Olares.

:::warning
In a decentralized system, you are responsible for your own security.

Keep your mnemonic phrase safe and secret, and never share it with anyone. It is the only way to recover your Olares account.
:::
## Set up local password
When exporting or backing up your mnemonic phrase for the first time, you may be prompted to set a local password for LarePass. This password is only used to unlock LarePass services on the current device.

After setting up, you can choose to enable biometric unlock for more secure and convenient access using face recognition or fingerprint.
![Set up local password](/images/manual/get-started/set-up-local-password.png)

:::info
* All Olares IDs on the same device share one local password in the LarePass app.
* Local passwords for LarePass apps installed on different devices are independent and stored separately.
:::

## Reveal and back up your mnemonic phrase
1. Open the LarePass app, and go to **Settings** > **LarePass settings** > **Safety** page.
2. Tap **Mnemonic phase**.
3. Read the risk warning and click **Start**. 
4. Click to view the 12-word mnemonic phrase, and enter your local password for identity verification.
5. Securely record the mnemonic phrase and store it in a safe place.
6. Click **Next**.
   :::warning
   While clicking **Copy** will save the mnemonic phrase to your clipboard, this poses a security risk. For maximum security, we strongly recommend backing up your mnemonic phrase offline.
   :::
7. Verify your backup by arranging the mnemonic phrase in the correct order.
8. Click **Completed**. 
   If arranged correctly, you have successfully backed up your mnemonic phrase.

   ![Back up mnemonic phrase](/images/manual/get-started/backup-mnemonic-phrase.png)

## FAQs
### What happens if I lose my mnemonic phrase?
Losing your mnemonic phrase will result in serious consequences:

* You will lose ownership of your digital identity (DID) and Olares ID.
* You won't be able to access data stored in Vault.

To prevent this, we strongly recommend taking these precautions:

* **Offline backup**: Write down the 12-word mnemonic phrase and store it securely, such as in a safe.
* **Multi-device backup**: Use LarePass's Vault to encrypt and save your mnemonic phrase on multiple devices. You will lose your mnemonic phrase only if all these devices are lost.

### I've activated Olares, why do I get a password error when trying to view my mnemonic phrase in LarePass?
If you encounter a password error, it may be because you haven't set a local password. Open the LarePass app, go to **Settings** > **LarePass settings** > **Security**, and set a local password. Then try the backup process again.