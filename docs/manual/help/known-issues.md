---
outline: [2, 3]
description: Known issues in Olares, including affected environments, impact, workarounds, and fix status.
---

# Olares known issues

This page lists confirmed issues that may affect Olares users. Each issue includes the affected environment, user impact, current status, and available workarounds.

:::info Status update
Olares will update this page when issues are identified, mitigated, or resolved.
:::

## Chrome real-time updates may stop when LarePass VPN is enabled

### Summary

| Item | Details |
| --- | --- |
| Affected platform | macOS |
| Affected browser | Google Chrome 148 or later |
| Trigger conditions | Accessing Olares through an `olares.com` URL in Chrome while LarePass VPN is enabled |
| Impact | WebSocket-based real-time updates may stop working |
| Status | Pending fix in an upcoming Olares release |

### Description

When this issue occurs, Olares pages still open in Chrome, but parts of the page that depend on live updates may stop refreshing.

You may notice one or more of the following symptoms:

- App status, notifications, or task progress do not update automatically.
- A page appears to be stuck until you refresh it.
- Live messages or WebSocket-powered panels stop updating.

Your data is safe. This issue affects real-time communication in the browser only. It does not indicate data loss, data corruption, account damage, app damage, or device storage issues.

### Cause

Starting in Chrome 148, Chrome enforces stricter Local Network Access checks for requests that cross from a public web origin to a local or private network address.

When LarePass VPN is enabled and you open Olares through an `olares.com` address, Chrome may treat some WebSocket connections as local network requests. For WebSocket connections, Chrome checks the `101 Switching Protocols` response. If the response is missing the header Chrome expects, the connection is blocked.

Olares does not currently include this header in the WebSocket upgrade response, which triggers the block. A fix is planned for an upcoming release.

### Workarounds

Use one of the following workarounds until you upgrade to an Olares version that includes the fix.

#### Option 1: Temporarily disable Chrome's local network access check

If you need to keep using the `olares.com` URL with LarePass VPN, you can temporarily disable Chrome's local network access check.

:::warning
This changes a browser security setting. Use it only as a temporary workaround. After you upgrade to an Olares version that includes the fix, restore this flag to **Default** or **Enabled**.
:::

1. In Chrome, open:

   ```text
   chrome://flags/#local-network-access-check
   ```

2. Set **Local Network Access Checks** to **Disabled**.
3. Click **Relaunch** to restart Chrome.
4. Open Olares again using your `olares.com` URL.

#### Option 2: Use the .local address

If your Mac and Olares are on the same local network, use the `.local` address instead of the `olares.com` address. This is the recommended workaround for LAN access.

Convert the standard Olares URL to the `.local` URL:

**Standard URL**
```text
https://<entrance_id>.<username>.olares.com
```

**Local URL**
```text
http://<entrance_id>.<username>.olares.local
```

If Chrome cannot open the `.local` address, make sure Chrome has local network access on macOS:

1. Open the Apple menu and go to **System Settings**.
2. Go to **Privacy & Security** > **Local Network**.
3. Turn on **Google Chrome** and **Google Chrome Helper**.
4. Restart Chrome and try the `.local` URL again.
