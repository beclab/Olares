---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/openclaw-mobile
outline: [2, 3]
title: Connect to OpenClaw from Android and iOS
description: Pair the OpenClaw Android or iOS app with your Olares Gateway using a QR code or manual connection settings over LarePass VPN.
app_version: "1.0.46"
doc_version: "1.0"
doc_updated: "2026-09-28"
---

# Connect to OpenClaw from Android and iOS

<VersionRouteSelect />

Use the OpenClaw mobile app to chat with the agent running on Olares. Pair the phone once, then use the same Gateway address at home and away with LarePass VPN enabled.

## Prerequisites

- Complete [OpenClaw setup](openclaw.md) on Olares, including configuring a model for chat.
- Install the official mobile app using the links in the upstream [Android guide](https://docs.openclaw.ai/platforms/android) or [iOS guide](https://docs.openclaw.ai/platforms/ios). Mobile app releases can differ from Gateway releases; not every Gateway release includes an Android APK.
- Install LarePass on the phone, sign in with the Olares account that owns this OpenClaw instance, and enable its VPN connection. See [Access Olares services securely](../manual/get-started/local-access.md).
- Have access to **OpenClaw CLI** on Olares for generating setup codes and approving devices.

:::important Keep LarePass VPN enabled on the phone
The **OpenClaw Gateway** entrance uses the **Internal** access policy. This skips Olares sign-in for connections through LarePass VPN. Being on the same Wi-Fi network alone does not grant that exemption.

Keep the VPN enabled on both local Wi-Fi and mobile data. On the local network, LarePass can use an Intranet connection. OpenClaw's token or setup code does not replace Olares entrance authentication.
:::

## Find the Gateway address

1. In Olares, open **Settings** > **Applications** > **OpenClaw**.
2. Under **Entrances**, open **OpenClaw Gateway** and copy its domain. Keep its authentication level set to **Internal**.

The Gateway entrance is hidden from the Launchpad. Its address differs from **OpenClaw CLI**, **Control UI**, and the Olares Desktop address. If you use a cloned app, select that instance's Gateway entrance.

In the examples below, replace `gateway.example.com` with that domain. The connection URL is `wss://gateway.example.com`, using port **443** and **TLS**. The internal service port `18789` is not the port to enter when using the Olares entrance.

## Pair with a QR code

### Generate a code in OpenClaw CLI

1. Open **OpenClaw CLI** from the Olares Launchpad.
2. Generate a mobile setup code:

   ```bash
   openclaw qr
   ```

   New installations of Olares app version **1.0.46** configure the Gateway entrance as the pairing URL. Check that the output's **Gateway** line contains the domain you copied, with `wss://`.

3. If you upgraded an existing installation, changed the entrance domain, or the displayed address is incorrect, specify the address explicitly:

   ```bash
   openclaw qr --url 'wss://gateway.example.com'
   ```

   This command does not change your saved configuration. Upgrades preserve existing `openclaw.json` settings, so a new default does not overwrite an existing installation.

:::tip Use a setup code on the same phone
If you cannot scan the terminal QR, generate a code to paste into the app:

```bash
openclaw qr --url 'wss://gateway.example.com' --setup-code-only
```

The setup code is a short-lived credential. Keep it private and generate a new one if it expires. Use a mobile setup code, not a headless-node join URL from `openclaw devices join-code`.
:::

### Scan or paste on the phone

**Android**

1. Keep LarePass VPN enabled and open OpenClaw.
2. During setup, select **Scan QR or setup code**. If you have already set up the app, open **Settings** > **Gateway** > **Add Gateway**.
3. Scan the terminal QR, or choose the option to enter a setup code and paste it.
4. Confirm that the Gateway address matches your Olares entrance, then connect.

**iOS**

1. Keep LarePass VPN enabled and open OpenClaw.
2. Open Gateway setup on first launch, or **Settings** > **Gateway** to add a Gateway.
3. Choose the QR/setup-code option, then scan the terminal QR or paste the setup code.
4. Confirm the address and connect. Allow camera access if you choose to scan.

Official apps can complete pairing automatically when the setup-code metadata matches. If approval remains pending, follow [Approve a pending device](#approve-a-pending-device).

By default, a setup code for a `wss://` endpoint grants the phone node access and full Gateway operator access. To request a reduced operator profile, generate the code with `--limited`. See the upstream [QR reference](https://docs.openclaw.ai/cli/qr).

## Pair manually

1. In **OpenClaw CLI**, display the Gateway token:

   ```bash
   openclaw gateway auth-token --show
   ```

   Copy the token privately to the phone. It is an OpenClaw credential, not your Olares account password.

2. Open manual Gateway setup:
   - **Android**: Choose **Set up manually**, or open **Settings** > **Gateway** > **Manual Gateway**.
   - **iOS**: Open **Settings** > **Gateway** and enable **Use Manual Gateway** (or **Manual Host**, depending on the app version).

3. Fill in the connection details:

   | Field | Value |
   | --- | --- |
   | Host | Your Gateway entrance domain, for example `gateway.example.com`. Do not include a scheme or path when host and port are separate fields. |
   | Port | `443` |
   | Token | The Gateway token from step 1 |
   | Password | Leave empty for the default token-based Olares deployment |
   | Connection security / TLS | **Secure (TLS)** / enabled |

   If the app provides one complete URL field instead, enter `wss://gateway.example.com`. If you deliberately changed OpenClaw to password authentication, use the configured Gateway password instead of a token.

4. Tap **Test connection**, **Connect**, or **Save & Connect**, depending on the app version. A **pairing required** message means the connection reached the Gateway and needs device approval.
5. Approve the pending device as described below, then reconnect on the phone.

## Approve a pending device

1. Leave the phone on its connection screen with LarePass VPN enabled. In **OpenClaw CLI**, run:

   ```bash
   openclaw devices list
   ```

2. In **Pending**, identify your phone by its device name and device ID. Review the requested roles and scopes. Use the current **Request** ID from this list:

   ```bash
   openclaw devices approve <requestId>
   ```

3. Return to the phone and retry the connection or confirm that you have approved it.
4. Run `openclaw devices list` again. Mobile apps use both `node` and `operator` connections; if another request appears for your phone, review and approve that current request too.

:::tip Request IDs can change
If a retry changes the requested role, scopes, or public key, the Gateway can replace a pending request with a new one. The phone may still display an earlier error and Request ID.

For `unknown requestId` or a mismatch between the phone and CLI, refresh `openclaw devices list` and use the current request for the verified device. Do not repeatedly approve an old ID or approve an unrelated device. If the phone is already under **Paired** and there are no pending requests, reconnect; another approval is unnecessary.
:::

## Verify local and remote access

1. Keep the OpenClaw app open in the foreground and check the connection:

   ```bash
   openclaw devices list
   openclaw nodes status
   ```

   The device should be paired, and its node should be connected while the app is active.

2. Send a short chat message and confirm a reply. Pairing alone does not verify the model configuration.
3. With LarePass VPN still enabled, turn off Wi-Fi and switch the phone to mobile data. Reconnect using the **same saved Gateway address**, then send another message.

You do not need to change the domain, port, or TLS setting when moving between networks. Mobile operating systems may suspend the app in the background; foreground the app when testing node capabilities.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| `Expected HTTP 101 response but was '400 Bad Request'` | The WebSocket handshake failed before pairing. First check that LarePass VPN is enabled on this phone, the host is the **Gateway** entrance, and the port is `443` with TLS. A `400` alone does not identify which proxy returned it. |
| QR advertises loopback, a container IP, or the wrong domain | Generate it again with `openclaw qr --url 'wss://gateway.example.com'`. |
| Setup code expired or rejected | Generate a fresh mobile setup code and paste it into the setup-code field, not the manual token field. |
| Token authentication fails | Retrieve the current token from this OpenClaw instance. Do not use the Olares password or a token from another clone. |
| `pairing required` or `unknown requestId` | Follow [Approve a pending device](#approve-a-pending-device), using the latest request for your phone. |
| Paired in CLI, but the phone still shows an old error | Reconnect with VPN enabled. If necessary, close and reopen the mobile app. |
| iOS reports `Gateway setup incomplete` | Generate a fresh QR/setup code and pair again to obtain both node and operator credentials. |

## Learn more

- [Manage application entrances](../manual/olares/settings/manage-entrance.md)
- [OpenClaw Android documentation](https://docs.openclaw.ai/platforms/android)
- [OpenClaw iOS documentation](https://docs.openclaw.ai/platforms/ios)
