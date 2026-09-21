---
outline: [2, 3]
description: Identify the credentials used by Olares, understand what they protect, and find the right recovery option.
---
# Find the right Olares credential

Olares uses separate credentials for account sign-in, LarePass, identity recovery, and host access. They protect different parts of the system and are not interchangeable.

:::warning
If you lose access to every copy of your mnemonic phrase, you cannot recover your Olares ID.
:::

## Common prompts

| Where it appears                  | Credential                                              | What it is for                                                   |
| --------------------------------- | ------------------------------------------------------- | ---------------------------------------------------------------- |
| Activation Wizard                 | Wizard one-time password                                | Opens the [activation flow](../get-started/join-olares.md#step-2-activate-your-account) |
| Olares sign-in page               | [Olares login password](#olares-login-password)         | Signs you in to Olares and protected applications                |
| After entering the login password | Olares login verification                               | Confirms [sign-in](../get-started/join-olares.md#step-3-log-in-to-olares) with a second factor |
| LarePass                          | [LarePass local password](#larepass-local-password)     | Unlocks protected functions in the current LarePass installation |
| Account import or recovery        | [Olares ID mnemonic phrase](#olares-id-mnemonic-phrase) | Imports an existing Olares ID                                    |
| SSH or physical host login        | [Host or SSH password](#host-or-ssh-password)           | Accesses the operating system and Olares host terminal           |

## Credential details

### Olares login password

Use this password to sign in to Olares and applications protected by its sign-in system. It does not unlock LarePass or the host operating system.

- **Where it is kept:** On your Olares device, in protected form within its authentication service.
- **If you forget it:**
  - **As a team member**, ask a team administrator to [reset your password in Settings](../olares/settings/manage-team.md#reset-passwords).
  - **As an Olares administrator**, [reset it from the terminal](./ts-forget-login-password.md).
- **To change it:** [Change your password in Settings](../password-and-devices.md#change-password).

### LarePass local password

Use this password to unlock protected functions in LarePass, including viewing the mnemonic phrase. It applies to all Olares IDs in the current installation.

- **Where it is kept:** In the current LarePass installation. Each installation has its own local password.

- **If you forget it:**
  - **With biometric unlock:** On mobile, open **Settings** > **Security** to view the password.
  - **Otherwise:** Uninstall and reinstall LarePass, import your Olares ID with its mnemonic phrase, and set a new local password.

### Olares ID mnemonic phrase

Use this 12-word phrase to import your Olares ID into another LarePass installation. Olares cannot reset or reproduce it.

- **Where it is kept:** LarePass stores an encrypted copy locally. Keep another complete copy in a secure location you control.

See [Back up mnemonic phrase](../larepass/back-up-mnemonics.md) and [Manage accounts in LarePass](../larepass/manage-accounts.md#import-an-account).

### Host or SSH password

Use this password to access the operating system on the device running Olares, including its host terminal.

- **Where it is kept:** The host operating system manages the account password. On Olares One, the SSH password generated after activation is also saved in Vault.
- **If you forget it on Olares One:** Check the saved password in LarePass Vault.

See [Access Olares One via SSH](../../one/access-terminal-ssh.md).

The host password and Olares login password are independent. Resetting one does not change the other.
