---
outline: [2, 3]
description: Log in to Olares from olares-cli and manage profiles. Covers profile login, importing a refresh token, switching profiles, and where tokens are stored.
---

# Log in to Olares <Badge type="tip" text="^1.12.5" />

Before `olares-cli` can act on behalf of an Olares user, log in once to create a profile. After that, the CLI handles token refresh for you, and you only log in again when the refresh token itself becomes invalid.

## Understand profile

A profile is one Olares instance plus one user identity. Your identity is your Olares ID, such as `alice123@olares.com`. Each profile stores its own access and refresh tokens in the operating system's secure keychain.

| OS | Backend | Location |
|------|---------|----------|
| macOS | Keychain | service `olares-cli`, account = olaresId |
| Linux | AES-256-GCM file | under `~/.local/share/olares-cli/` |
| Windows | DPAPI | `HKCU\Software\OlaresCli\keychain` |

After a successful login, the CLI prints a line like `token stored via <backend> (service "olares-cli", account "<id>")`. That message tells you where your token actually landed.

A single machine can store many profiles, but only one is current at a time. Every command runs against the current profile.

## Log in for the first time

1. Start the login process. Replace `alice123@olares.com` with your own Olares ID.

   ```bash
   olares-cli profile login --olares-id alice123@olares.com
   ```

2. When the CLI prompts `password for <id>:`, type your Olares password and press Enter. The input is hidden.

3. If two-factor auth is enabled on your account, the CLI prompts again with `two-factor code for <id>:`. Enter the 6-digit code from LarePass and press Enter.

4. Verify the profile is created and logged in.

   ```bash
   olares-cli profile list
   ```

   Look for your Olares ID and a status of `logged-in (Xh Ym)`.

   Example output:

   ```text
   NAME                   OLARES-ID              STATUS
   *  laresprime@olares.com  laresprime@olares.com   logged-in (23h59m)
   ```
   The leading `*` marks the current profile.

## Next step

Now that you have a profile, install the [Agent skills](./cli-agent-skills.md) into your agent to drive Olares from natural language.
