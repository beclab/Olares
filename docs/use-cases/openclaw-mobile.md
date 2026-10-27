---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/openclaw-mobile
outline: [2, 3]
title: Connect using OpenClaw mobile clients
description: Pair the OpenClaw Android or iOS app with your Olares Gateway using a QR code or manual connection settings over LarePass VPN.
app_version: "1.0.47"
doc_version: "1.0"
doc_updated: "2026-09-30"
---

# Connect using OpenClaw mobile clients

Use the OpenClaw mobile app to chat with the agent running on Olares. Pair the phone once, then use the same Gateway address at home and away with LarePass VPN enabled.

## Prerequisites

- Complete [OpenClaw setup](openclaw.md) on Olares, including configuring a model for chat.
- Install LarePass on the phone and sign in with the Olares account that owns this OpenClaw instance.
- Upgrade OpenClaw to Chart version **1.0.47 or later**.

::: warning Important
Keep LarePass VPN enabled on the phone whether you use local Wi-Fi or mobile data. On the local network, LarePass automatically uses an Intranet connection. OpenClaw tokens and setup codes do not replace Olares entrance authentication.
:::

## Confirm the Gateway address and Auth Level

1. In Olares, open **Settings** > **Applications** > **OpenClaw**.
2. Under **Entrances**, open **OpenClaw Gateway** and copy its domain. Keep its authentication level set to **Internal**.

The Gateway entrance is hidden from the Launchpad. Its address differs from **OpenClaw CLI**, **Control UI**, and the Olares Desktop address. If you use a cloned app, select that instance's Gateway entrance.

In the examples below, replace `efa2f8ec2.yourolaresid.olares.com` with that domain. The connection URL is `wss://efa2f8ec2.yourolaresid.olares.com`, using port **443** and **TLS**. The internal service port `18789` is not the port to enter when using the Olares entrance.

Install and connect the client

<Tabs>
<template #iOS>

### Install the iOS client

1. Install the official OpenClaw client from the [App Store](https://apps.apple.com/app/openclaw-ai-that-does-things/id6780396132).
2. Open LarePass and confirm that its VPN is connected, then open OpenClaw. Choose either pairing path below.

### Path 1: Automatic pairing with a QR code

1. Open **OpenClaw CLI** from the Olares Launchpad.
2. Generate a mobile setup code:

   ```bash
   openclaw qr
   ```

   New installations of Olares app version **1.0.47 or later** configure the Gateway entrance as the pairing URL. Check that the output's **Gateway** line contains the domain you copied, with `wss://`.

3. If you upgraded an existing installation, changed the entrance domain, or the displayed address is incorrect, specify the address explicitly:

   ```bash
   openclaw qr --url 'wss://efa2f8ec2.yourolaresid.olares.com'
   ```

   This command does not change your saved configuration. Upgrades preserve existing `openclaw.json` settings, so a new default does not overwrite an existing installation.

4. Open **Connections** in the mobile client and select **Scan QR to pair**.
5. Allow camera access when prompted, then scan the terminal QR code.
6. On the first connection, the client asks whether to trust the Gateway. Verify that the address matches your Olares Gateway entrance, then select **Trust and connect**. The client pairs automatically; if approval remains pending, follow [Approve a pending device](#approve-a-pending-device).
7. Open chat, send a message, and confirm a reply.

The QR code contains a short-lived pairing credential. Keep it private and generate a new one if it expires. By default, a `wss://` setup code grants node access and full Gateway operator access. Add `--limited` to request a reduced operator profile. See the [upstream QR reference](https://docs.openclaw.ai/cli/qr).

### Path 2: Manual pairing

1. In **OpenClaw CLI**, display the Gateway token:

   ```bash
   openclaw gateway auth-token --show
   ```

   Copy the token privately to the phone. It is an OpenClaw credential, not your Olares account password.

2. On your iPhone, open **Settings** > **Gateway** and enable **Use Manual Gateway** (or **Manual Host**, depending on the app version).
3. Fill in the connection details:

   | Field | Value |
   | --- | --- |
   | Host | Your Gateway entrance domain, such as `efa2f8ec2.yourolaresid.olares.com`. Do not include a scheme or path when host and port are separate fields. |
   | Port | `443` |
   | Token | The Gateway token from step 1 |
   | Password | Leave empty for the default token-based Olares deployment |
   | Connection security / TLS | **Secure (TLS)** / enabled |

   If the client provides one complete URL field, enter `wss://efa2f8ec2.yourolaresid.olares.com`. If you changed OpenClaw to password authentication, enter the configured Gateway password instead.

4. Tap **Test connection**, **Connect**, or **Save & Connect**. A **pairing required** message means the connection reached the Gateway and needs device approval.
5. Follow [Approve a pending device](#approve-a-pending-device), using `openclaw devices approve <requestId>` to approve the connection. Return to the client and reconnect, keeping the app in the foreground.
6. In **OpenClaw CLI**, list the node capability requests created after reconnection:

   ```bash
   openclaw nodes pending
   ```

7. Identify your phone and review its requested commands and capabilities. If a request is pending, approve it using the Request ID from this list:

   ```bash
   openclaw nodes approve <nodeRequestId>
   ```

   This node request ID differs from the device pairing request ID in `devices list`. Do not interchange them. If no node request is pending, continue to check the node status.

8. Check node connectivity and capabilities:

   ```bash
   openclaw nodes status
   openclaw nodes describe --node <nodeId>
   ```

   Replace `<nodeId>` with your phone's node ID from the status output. Confirm it is connected. Camera, microphone, and location features also require permission on the phone. See [Node pairing](https://docs.openclaw.ai/cli/nodes#pairing).

9. Open chat, send a message, and confirm a reply.

</template>

<template #Android>

### Install the Android client

1. Install the official OpenClaw client from [Google Play](https://play.google.com/store/apps/details?id=ai.openclaw.app). Alternatively, follow the [official installation guide](https://docs.openclaw.ai/platforms/android#install-outside-google-play) to download and verify a signed APK. Not every Gateway release includes an APK.
2. Open LarePass and confirm that its VPN is connected, then open OpenClaw. Choose either pairing path below.

### Path 1: Automatic pairing with a QR code

1. Open **OpenClaw CLI** from the Olares Launchpad.
2. Generate a mobile setup code:

   ```bash
   openclaw qr
   ```

   New installations of Olares app version **1.0.47 or later** configure the Gateway entrance as the pairing URL. Check that the output's **Gateway** line contains the domain you copied, with `wss://`.

3. If you upgraded an existing installation, changed the entrance domain, or the displayed address is incorrect, specify the address explicitly:

   ```bash
   openclaw qr --url 'wss://efa2f8ec2.yourolaresid.olares.com'
   ```

   This command does not change your saved configuration. Upgrades preserve existing `openclaw.json` settings, so a new default does not overwrite an existing installation.

4. Open **Connections** in the mobile client and select **Scan QR to pair**.
5. Allow camera access when prompted, then scan the terminal QR code.
6. On the first connection, the client asks whether to trust the Gateway. Verify that the address matches your Olares Gateway entrance, then select **Trust and connect**. The client pairs automatically; if approval remains pending, follow [Approve a pending device](#approve-a-pending-device).
7. Open chat, send a message, and confirm a reply.

The QR code contains a short-lived pairing credential. Keep it private and generate a new one if it expires. By default, a `wss://` setup code grants node access and full Gateway operator access. Add `--limited` to request a reduced operator profile. See the [upstream QR reference](https://docs.openclaw.ai/cli/qr).

### Path 2: Manual pairing

1. In **OpenClaw CLI**, display the Gateway token:

   ```bash
   openclaw gateway auth-token --show
   ```

   Copy the token privately to the phone. It is an OpenClaw credential, not your Olares account password.

2. In the Android client, choose **Set up manually**, or open **Settings** > **Gateway** > **Manual Gateway**.
3. Fill in the connection details:

   | Field | Value |
   | --- | --- |
   | Host | Your Gateway entrance domain, such as `efa2f8ec2.yourolaresid.olares.com`. Do not include a scheme or path when host and port are separate fields. |
   | Port | `443` |
   | Token | The Gateway token from step 1 |
   | Password | Leave empty for the default token-based Olares deployment |
   | Connection security / TLS | **Secure (TLS)** / enabled |

   If the client provides one complete URL field, enter `wss://efa2f8ec2.yourolaresid.olares.com`. If you changed OpenClaw to password authentication, enter the configured Gateway password instead.

4. Tap **Test connection**, **Connect**, or **Save & Connect**. A **pairing required** message means the connection reached the Gateway and needs device approval.
5. Follow [Approve a pending device](#approve-a-pending-device), using `openclaw devices approve <requestId>` to approve the connection. Return to the client and reconnect, keeping the app in the foreground.
6. In **OpenClaw CLI**, list the node capability requests created after reconnection:

   ```bash
   openclaw nodes pending
   ```

7. Identify your phone and review its requested commands and capabilities. If a request is pending, approve it using the Request ID from this list:

   ```bash
   openclaw nodes approve <nodeRequestId>
   ```

   This node request ID differs from the device pairing request ID in `devices list`. Do not interchange them. If no node request is pending, continue to check the node status.

8. Check node connectivity and capabilities:

   ```bash
   openclaw nodes status
   openclaw nodes describe --node <nodeId>
   ```

   Replace `<nodeId>` with your phone's node ID from the status output. Confirm it is connected. Camera, microphone, and location features also require permission on the phone. See [Node pairing](https://docs.openclaw.ai/cli/nodes#pairing).

9. Open chat, send a message, and confirm a reply.

</template>
</Tabs>

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
| QR advertises loopback, a container IP, or the wrong domain | Generate it again with `openclaw qr --url 'wss://efa2f8ec2.yourolaresid.olares.com'`. |
| Setup code expired or rejected | Generate a fresh QR code and scan it again. Do not use the pairing credential in the manual token field. |
| Token authentication fails | Retrieve the current token from this OpenClaw instance. Do not use the Olares password or a token from another clone. |
| Pairing command shown on the phone fails | Run `openclaw devices list` in OpenClaw CLI, identify your phone’s pending request, then run `openclaw devices approve <requestId>`. The phone may display an older requestId; **use the current requestId from CLI**. |
| `pairing required` or `unknown requestId` | Follow [Approve a pending device](#approve-a-pending-device), using the latest request for your phone. |
| Paired in CLI, but the phone still shows an old error | Reconnect with VPN enabled. If necessary, close and reopen the mobile app. |
| iOS reports `Gateway setup incomplete` | Generate a fresh QR/setup code and pair again to obtain both node and operator credentials. |

## Learn more

- [Manage application entrances](../manual/olares/settings/manage-entrance.md)
- [OpenClaw Android documentation](https://docs.openclaw.ai/platforms/android)
- [OpenClaw iOS documentation](https://docs.openclaw.ai/platforms/ios)
