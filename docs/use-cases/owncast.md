---
outline: [2, 3]
description: Set up Owncast on Olares as a self-hosted live streaming server and stream with OBS Studio while keeping full control over your content and audience.
head:
  - - meta
    - name: keywords
      content: Olares, Owncast, live streaming, self-hosted, OBS Studio, RTMP, broadcast
app_version: "1.0.7"
doc_version: "1.0"
doc_updated: "2026-03-30"
---

# Host live streams with Owncast

Owncast is an open-source, self-hosted live streaming and chat server. It works with popular broadcasting software like OBS Studio. You get full ownership over your content, chat moderation, and audience data.

Running Owncast on Olares keeps your streaming infrastructure private and self-contained, with no dependence on third-party platforms.

## Learning objectives

In this guide, you will learn how to:
- Install and configure Owncast on Olares.
- Set up OBS Studio with video and audio sources.
- Connect OBS to your Owncast server for RTMP streaming.
- Go live and share your stream with viewers.

## Prerequisites

- [OBS Studio](https://obsproject.com/) installed on your computer.
- LarePass desktop app installed on your computer. Required for [enabling VPN access](../manual/larepass/private-network.md#enable-vpn-on-larepass) when streaming from outside your local network.

## Install Owncast

1. Open Market and search for "Owncast".

   <!-- ![Owncast](/images/manual/use-cases/owncast.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Configure Owncast

After installation, access the admin dashboard to get the streaming information needed for OBS.

1. Open Owncast from Launchpad.
2. In your browser's address bar, add `/admin` to the end of the URL to open the admin dashboard. For example:
   ```plain
   https://abc123.username.olares.com/admin
   ```
3. Log in with the default credentials:
   - **Username**: `admin`
   - **Password**: `abc123`

   <!-- ![Owncast admin dashboard](/images/manual/use-cases/owncast-admin.png#bordered) -->

4. Note the following information from the admin dashboard. You will need this information when configuring OBS.
   - **Streaming URL**: The RTMP server address (e.g., `rtmp://abc123.username.olares.com:1935/live`).
   - **Stream Key**: The key used to authenticate your broadcast.

   <!-- ![Owncast streaming info](/images/manual/use-cases/owncast-stream-info.png#bordered) -->

:::warning
Change the default admin password and stream key before going live to prevent unauthorized access.
:::

## Set up OBS Studio

### Grant system permissions (macOS)

On macOS, OBS requires explicit permissions to capture your screen, microphone, and camera.

1. Open **System Settings** > **Privacy & Security**.
2. Enable the following permissions for OBS:
   - **Screen & System Audio Recording**
   - **Microphone**
   - **Camera**

<!-- ![OBS macOS permissions](/images/manual/use-cases/owncast-obs-permissions.png#bordered) -->

### Optimize output settings

Adjust the output settings in OBS for a smoother streaming experience.

1. In OBS, go to **Settings** > **Output**.
2. Set **Output Mode** to **Advanced**.
3. Under the **Streaming** tab, configure the following:
   - **Audio Encoder**: Select **FFmpeg AAC**.
   - **Video Encoder**: Select a hardware encoder if available (e.g., **Apple VT H264 Hardware Encoder** on macOS).
   - **Rate Control**: Select **CBR**.
   - **Bitrate**: `10000 Kbps` (adjust based on your upload speed).
   - **Keyframe Interval**: `0` (auto).
   - **Profile**: **high**.
   - **B-frames**: Check **Use B-frames** if available.
4. Click **Apply**.

<!-- ![OBS output settings](/images/manual/use-cases/owncast-obs-output.png#bordered) -->

### Add video and audio sources

Before streaming, you need at least one video source in OBS. The following are common examples. Choose the source types that fit your streaming needs.

1. In the OBS **Sources** panel, click **+** and select **Screen Capture** (or **Display Capture**). Choose the display you want to share and click **OK**.

   <!-- ![OBS screen capture](/images/manual/use-cases/owncast-obs-screen-capture.png#bordered) -->

2. Click **+** again and select **Video Capture Device**. Choose your camera from the device list and click **OK**.

   <!-- ![OBS video capture](/images/manual/use-cases/owncast-obs-video-capture.png#bordered) -->

3. Arrange and resize the sources in the preview area as needed.

## Connect OBS to Owncast

The connection steps depend on whether your computer and Olares are on the same local network.

<tabs>
<template #Use-.local-domain-(LAN)>

If your computer is on the same local network as Olares, you can stream using the `.local` domain without LarePass VPN.

:::info Windows users
On Windows, multi-level `.local` domains require additional setup. Try one of these:
- **Import hosts in LarePass**: Open the LarePass desktop app and use the built-in option to import Olares hosts to your system.
- **Use the single-level domain**: Change `rtmp://abc123.{username}.olares.local:1935/live` to `rtmp://abc123-{username}-olares.local:1935/live`.

For details, see [Access Olares services locally](../manual/best-practices/local-access.md).
:::

1. In OBS, go to **Settings** > **Stream**.
2. For **Service**, select **Custom**.
3. For **Server**, paste the Streaming URL from your Owncast admin dashboard and replace `.com` with `.local` in the domain. For example, if the Streaming URL is:
   ```plain
   rtmp://abc123.username.olares.com:1935/live
   ```
   
   Change it to:
   ```plain
   rtmp://abc123.username.olares.local:1935/live
   ```
4. For **Stream Key**, enter the stream key from your Owncast admin dashboard.
5. Click **Apply**.

<!-- ![OBS stream settings](/images/manual/use-cases/owncast-obs-stream-lan.png#bordered) -->

</template>

<template #Use-.com-domain-(VPN)>

If your computer is not on the same local network as Olares, enable LarePass VPN to stream over a secure connection.

1. Open the LarePass desktop app and click your avatar in the top-left corner to open the user menu. Toggle on the switch for **VPN connection**.

   Once enabled, make sure the connection status is either **Intranet** (LAN) or **P2P** (outside LAN).

   ![Enable LarePass VPN](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

2. In OBS, go to **Settings** > **Stream**.
3. For **Service**, select **Custom**.
4. For **Server**, paste the Streaming URL from your Owncast admin dashboard.
5. For **Stream Key**, enter the stream key from your Owncast admin dashboard.
6. Click **Apply**.

<!-- ![OBS stream settings](/images/manual/use-cases/owncast-obs-stream-vpn.png#bordered) -->

</template>
</tabs>

## Start streaming

1. In OBS, click **Start Streaming**.
2. Open the Owncast viewer page (the main Owncast URL without `/admin`) to verify your stream is live.

<!-- ![Owncast live stream](/images/manual/use-cases/owncast-live.png#bordered) -->

:::tip Share your stream
The Owncast viewer page is publicly accessible. Share the URL with your audience so they can watch and chat in real time.
:::

## FAQ

### Why is there a ~10-second delay in my stream?

RTMP streaming typically has a latency of around 10 seconds. This is expected behavior for all RTMP-based setups and is not specific to Owncast or Olares.

## Learn more

- [Owncast documentation](https://owncast.online/docs/): Advanced server configuration, chat customization, and Fediverse integration.
- [OBS Studio wiki](https://obsproject.com/wiki/): Guides for scenes, filters, and streaming optimization.
