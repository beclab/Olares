---
outline: [2, 3]
description: Set up Navidrome on Olares to build your personal music streaming server. Import your music library and stream it to any device with Subsonic-compatible apps.
head:
  - - meta
    - name: keywords
      content: Olares, Navidrome, music server, self-hosted, music streaming, Subsonic
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-03-30"
---

# Stream your music library with Navidrome

Navidrome is an open-source, self-hosted music streaming server that turns your local music collection into a personal cloud library. It supports virtually all audio formats, uses minimal resources, and is compatible with the Subsonic/Airsonic API. This lets you connect with a wide range of third-party music apps on any device.

## Learning objectives

In this guide, you will learn how to:
- Install Navidrome and set up an administrator account.
- Import music files from Olares Files.
- Stream your library to a mobile device using a Subsonic-compatible app.

## Install Navidrome

1. Open Market and search for "Navidrome".
    <!-- ![Install Navidrome](/images/manual/use-cases/navidrome.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Set up the admin account

1. Open Navidrome from Launchpad.
2. Enter a username and password to create the administrator account.

## Import music

Navidrome automatically scans the `Home/Music` directory in Olares Files.

:::tip Upload via LarePass desktop
It's recommended to upload music files through the LarePass desktop client, as it lets you track upload progress.
:::

1. Open the Files app from Launchpad.
2. Navigate to `Home/Music` and upload your music files to this directory.
3. Return to Navidrome. If the library does not update automatically, click the **Activity** icon in the top-right corner and select **Full Scan** to trigger a manual scan.

<!-- ![Upload music files](/images/manual/use-cases/navidrome-upload-music.png#bordered) -->
<!-- ![Trigger full scan](/images/manual/use-cases/navidrome-full-scan.png#bordered) -->

## Stream to your phone

You can stream your Navidrome library to your phone using any Subsonic-compatible music app. The following steps use StreamMusic as an example.

The connection steps depend on whether your phone and Olares device are on the same network.

<tabs>
<template #Use-.local-domain-(LAN)>

If your phone is on the same local network as Olares, you can connect using the `.local` domain without LarePass VPN.

:::info Windows users
On Windows, multi-level `.local` domains require additional setup. Try one of these:
- **Import hosts in LarePass**: Open the LarePass desktop app and use the built-in option to import Olares hosts to your system.
- **Use the single-level domain**: Change `https://abc123.{username}.olares.com` to `http://abc123-{username}-olares.local`.

For details, see [Access Olares services locally](../manual/best-practices/local-access.md).
:::

1. Open StreamMusic and add a new server connection to Navidrome.
2. For **Host address**, use your Navidrome URL with the `.local` domain and `http`. For example, if your Navidrome URL is:
    ```plain
    https://abc123.{username}.olares.com
    ```
    Change it to:
    ```plain
    http://abc123.{username}.olares.local
    ```
3. Enter your Navidrome username and password.
4. Tap **Login**. A success message confirms the connection.

    <!-- ![Login success](/images/manual/use-cases/navidrome-stream-music-success.png#bordered) -->

5. Return to the StreamMusic home screen. Your Navidrome library is now available for playback.

    <!-- ![StreamMusic library](/images/manual/use-cases/navidrome-stream-music-library.png#bordered) -->

</template>
<template #Use-.com-domain-(VPN)>

If your phone is not on the same network as Olares, update Navidrome's access policy and enable LarePass VPN.

1. Update Navidrome's access policy to enable direct access from external apps:

    a. In Olares, navigate to **Settings** > **Applications** > **Navidrome**.

    b. Change the **Authentication level** to **Internal** and click **Submit**.

    <!-- ![Change authentication level](/images/manual/use-cases/navidrome-authentication-level.png#bordered) -->

2. On your phone, open LarePass and enable LarePass VPN.

    ![Enable LarePass VPN on mobile](/images/manual/get-started/larepass-vpn-mobile.png#bordered)

3. Open StreamMusic and add a new server connection to Navidrome.
4. For **Host address**, enter the URL you use to access Navidrome in your browser.
5. Enter your Navidrome username and password.
6. Tap **Login**. A success message confirms the connection.

    <!-- ![Login success](/images/manual/use-cases/navidrome-stream-music-success.png#bordered) -->

7. Return to the StreamMusic home screen. Your Navidrome library is now available for playback.

    <!-- ![StreamMusic library](/images/manual/use-cases/navidrome-stream-music-library.png#bordered) -->

</template>
</tabs>

## Learn more

- [Stream media with Jellyfin](stream-media.md): Set up a video and media streaming server on Olares.
