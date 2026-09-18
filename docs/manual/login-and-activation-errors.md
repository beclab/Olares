---
outline: [2, 3]
description: Look up Olares login and activation error messages, understand what they indicate, and choose the next action.
head:
  - - meta
    - name: keywords
      content: Olares, login error, activation error, authentication failed, MFA binding error, DID binding error, invalid jws
---

# Login and activation error messages

Use this reference when Olares or LarePass displays a specific error during activation or sign-in. Match the complete message whenever possible. If the message differs, record it exactly instead of assuming it has the same cause.

## Activation errors

### `MFA binding error`

The request to bind multi-factor authentication (MFA) timed out. Check that both the device running LarePass and the Olares device have stable internet access, then retry the activation step.

### `DID binding error`

The request to the identity binding service timed out during activation. Check the network connection on both devices, then retry.

### `Invalid jws, timestamp is out of range`

The time on the device running LarePass differs from the Olares host by more than the accepted range. Enable automatic date and time on both devices, confirm that the time zone and clock are correct, then retry.

### `Resolve name error`

The Olares host could not resolve or reach the Olares identity service. Check its internet connection and DNS configuration, then retry. If the error persists, record the time and the DNS servers in use before collecting diagnostics.

## Login errors

### `Authentication failed, incorrect password`

The password does not match the account. Check capitalization and typing. If you no longer know the password, follow [Forgotten desktop login password](help/ts-forget-login-password.md).

### `Authentication failed, user not found`

Olares cannot find the username you entered. Check the local name and domain of the Olares ID. For a team member, ask an administrator to confirm that the account exists under **Settings** > **Users**.

### `Authentication failed, failed to query user from lldap service`

Olares could not retrieve the user record from its internal identity service. Retry once. If the error persists, ask an administrator to confirm that the user exists, record the time of the attempt, and [collect diagnostic information](collect-diagnostic-information.md).

### `too many failed login attempts, retry again later after 5 minutes`

Olares temporarily blocked sign-in after repeated failures. Stop retrying, wait at least five minutes, then enter the confirmed password once.

### `Authentication failed, disk space is full`

The system disk has no usable free space, so the authentication service cannot complete sign-in. Follow [Free up disk space](free-up-disk-space.md), then try again.

### `Authentication failed, lldap service is unavailable`

The internal identity service is unavailable. Ask an administrator to check whether LarePass also shows **System error**, then follow ["System error" in LarePass](help/ts-system-error.md).

### `Authentication failed, citus service is unavailable`

The internal database service is unavailable. Ask an administrator to check the system state, then follow ["System error" in LarePass](help/ts-system-error.md).

## If the error is not listed

Record the complete message, the time and time zone, the installed Olares version, and whether the error occurred during activation or login. Then follow [Collect diagnostic information](collect-diagnostic-information.md). Do not post full logs, Olares IDs, IP addresses, domains, or account details in a public issue.
