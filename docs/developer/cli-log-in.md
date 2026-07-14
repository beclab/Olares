---
outline: [2, 3]
description: Log in to Olares with olares-cli and manage profiles. Learn interactive login, profile switching and removal, and where authentication tokens are stored.
---

# Log in to Olares

Before `olares-cli` can act on behalf of an Olares user, log in once to create a profile. A profile is one Olares instance plus one user identity. After you log in, the CLI refreshes your tokens automatically, so you only log in again when the refresh token becomes invalid.

This page covers the user-mode login. It does not apply to host mode, which uses root and kubeconfig and needs no login.

## Log in for the first time

1. Run the following command to start the login. Replace `alice123@olares.com` with your own Olares ID.

   ```bash
   olares-cli profile login --olares-id alice123@olares.com
   ```

2. When the CLI prompts `password for <id>:`, type your Olares password and press Enter. The input is hidden.

3. If two-factor authentication is enabled on your Olares, the CLI prompts again with `two-factor code for <id>:`. Enter the 6-digit code from LarePass and press Enter.

4. Verify that the profile is created and logged in.

   ```bash
   olares-cli profile list
   ```

   Example output:

   ```text
      NAME                   OLARES-ID              STATUS
      laresprime@olares.com  laresprime@olares.com  logged-in
   *  alexmiles@olares.com   alexmiles@olares.com   logged-in
   ```

   The leading `*` marks the current profile.

## Manage profiles

If you work with more than one Olares instance or identity, each login adds a profile. Use these commands to move between them.

| Task | Command |
|------|---------|
| List all profiles | `olares-cli profile list` |
| Show the current identity | `olares-cli profile whoami` |
| Switch to another profile | `olares-cli profile use <name>` |
| Switch back to the previous profile | `olares-cli profile use -` |
| Remove a profile and its stored token | `olares-cli profile remove <name>` |

## Where tokens are stored

Tokens are stored automatically after a successful login. You don't need to manage them by hand. To clear them, use `olares-cli profile remove` rather than editing files directly.

| OS | Storage |
|------|---------|
| macOS | Keychain |
| Linux | AES-encrypted file under `~/.local/share/olares-cli/` |
| Windows | DPAPI |

## Next step

Install the [Agent Skills](./cli-agent-skills.md) into your agent to drive Olares from natural language.
