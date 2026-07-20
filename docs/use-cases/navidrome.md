---
outline: [2, 3]
description: Set up Navidrome on Olares to build a private music streaming server, organize your music library, and play it from a mobile Subsonic-compatible client.
head:
  - - meta
    - name: keywords
      content: Olares, Navidrome, music streaming, self-hosted music server, Subsonic, mobile music client
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-05-26"
---

# Stream your music library with Navidrome

Navidrome is an open-source, self-hosted music streaming server that turns your local music collection into a personal cloud library. It supports virtually all audio formats, uses minimal resources, and is compatible with the Subsonic/Airsonic API. This lets you connect with a wide range of third-party music apps on any device.

## Learning objectives

In this guide, you will learn how to:
- Install Navidrome and create the administrator account.
- Import music files to the Navidrome library folder from Olares Files.
- Stream your library to a mobile device using a Subsonic-compatible app.

## Prerequisites

- LarePass installed on your phone and signed in with your Olares ID.
- Music files ready to upload to Olares.
- A mobile Subsonic-compatible music client. This guide uses [Stream Music](https://music.aqzscn.cn/en/docs/versions/latest) as an example.

## Install Navidrome

1. Open Market and search for "Navidrome".

   ![Navidrome](/images/manual/use-cases/navidrome.png#bordered){width=95%}

2. Click **Get**, then **Install**, and wait for installation to complete.

## Set up the administrator account

When you open Navidrome for the first time, create the administrator account used to manage the server.

1. Open Navidrome from Launchpad.

2. Follow the page prompts to create an administrator username and password.

   ![Create the Navidrome administrator account](/images/manual/use-cases/navidrome-create-admin.png#bordered){width=95%}

After signing in, you should see an empty Navidrome library. The library will populate after you upload music and Navidrome scans the folder.

## Add music to Navidrome

Navidrome scans the `Home/Music` folder in Olares Files. Files placed in this folder appear in your Navidrome library after scanning.

1. Open Files from Launchpad.

2. Go to **Home** > **Music**.

3. Upload your music files to this folder.

   ![Upload music to Home/Music](/images/manual/use-cases/navidrome-upload-music.png#bordered){width=95%}

:::tip Organize by album
For cleaner browsing, organize files into album folders before uploading them. Navidrome can still identify single tracks, but album folders make the library easier to scan and maintain.
:::

After the scan finishes, Navidrome displays your songs, albums, and artists in the library.

## Scan the music library

Navidrome scans the music folder automatically. If newly uploaded songs do not appear, run a manual full scan.

1. Return to Navidrome.

2. In the upper-right corner, click **Activity**.

3. Click the **Full Scan** icon.

   ![Run a full scan in Navidrome](/images/manual/use-cases/navidrome-full-scan.png#bordered){width=95%}

4. Wait for the scan to complete, then refresh the library view.

## Connect a mobile music client

To stream from your phone, allow client access to Navidrome, enable LarePass VPN, and sign in from a Subsonic-compatible client.

1. Update Navidrome's access policy and copy its endpoint:

   a. Open Settings, then go to **Applications** > **Navidrome**.

   b. Under **Entrances**, click **Navidrome**.

   c. Set **Authentication level** to **Internal**, then click **Submit**.

   d. Under **Endpoint settings**, copy the URL displayed in **Endpoint**.

   ![Set Navidrome authentication level to Internal](/images/manual/use-cases/alex-navidrome-endpoint.png#bordered){width=95%}

2. Enable LarePass VPN on your phone.

   ![Enable LarePass VPN on mobile](/images/manual/get-started/larepass-vpn-mobile.png#bordered){width=95%}

3. Open Stream Music on your phone, then choose the option to connect to Navidrome.

   ![Connect Stream Music to Navidrome](/images/manual/use-cases/navidrome-music-stream-connect.png#bordered){width=95%}

4. In the login page, enter your info:
   - **Host address**: The Navidrome endpoint you copied from Olares Settings.
   - **Username**: Your Navidrome username.
   - **Password**: Your Navidrome password.
   
   ![Log in to Navidrome](/images/manual/use-cases/navidrome-log-in.png#bordered){width=95%}

5. Tap **Login**.

When the app shows a login success message, return to the home page. Your Navidrome library should appear in the mobile client.

## FAQs

### Can I add lyrics with external `.lrc` files?

Navidrome on Olares is most reliable with embedded synchronized lyrics. External `.lrc` files might not appear in the Navidrome web interface. Some Subsonic-compatible clients might support them depending on the Navidrome and client versions.

To make lyrics work consistently, embed synchronized lyrics into the audio files before uploading them to `Home/Music`.

## Learn more

- [Build your private media server with Jellyfin](stream-media.md): Stream movies, shows, and music from Olares with Jellyfin.
- [Build your digital library with Komga](komga.md): Manage comics, manga, magazines, and e-books on Olares.
