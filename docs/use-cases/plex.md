---
outline: [2, 3]
description: Organize personal videos with Plex on Olares, play them in a browser or on your phone and computer, and enable hardware-accelerated transcoding.
head:
  - - meta
    - name: keywords
      content: Olares, Plex Media Server, personal media library, video streaming, NVIDIA, hardware transcoding, Plex Pass
app_version: "0.0.14"
doc_version: "1.2"
doc_updated: "2026-11-16"
---

# Organize and play your personal videos with Plex

Plex Media Server is a media library and streaming app for your videos and music. On Olares, you can store your collection on your own device and play it in a browser or a Plex app.

## Learning objectives

In this guide, you will learn how to:

- Install Plex, choose an acceleration mode, and connect your Plex account.
- Organize media folders on Olares and choose library types for personal videos, movies, and music.
- Select playback quality and subtitles in a browser, and check whether hardware transcoding is active.
- Add the local HTTPS address in Plex so a browser or TV on the same LAN can open your library.
- Connect a Plex client on your phone or computer on another network through LarePass VPN.

## Prerequisites

- Olares 1.12.7 or later.
- A Plex account. A free account is enough to connect the server and create libraries.
- Hardware-accelerated transcoding requires a compatible NVIDIA GPU and an active Plex Pass on the server owner's account.

## Install Plex Media Server

1. Open Market and search for "Plex Media Server".
2. Click **Get**, then **Install**.
3. In the **Choose a hardware accelerator** dialog, select a mode:
   - **CPU**: Use software transcoding.
   - **NVIDIA GPU**: Use your NVIDIA GPU for [hardware-accelerated transcoding](#use-hardware-accelerated-transcoding).

   :::info Acceleration mode
   The acceleration mode cannot be changed directly after installation.
   :::

4. Click **Confirm**. If prompted for **PLEX_CLAIM**, leave it empty and confirm to continue. You will connect your account in the next section.
5. Wait for installation to complete and the button to change to **Open**.

## Connect your Plex account

You will see two entrances for Plex Media Server in Launchpad. Use **Plex Account Setup** to connect the server to your Plex account, then use **Plex Media Server** to manage libraries and play media.

1. Open **Plex Account Setup** from Launchpad.
2. In another browser tab, open the [Plex claim page](https://www.plex.tv/claim/) and sign in to the account you want to connect.
3. Copy the temporary token that starts with `claim-`.
4. Return to Plex Account Setup, paste the token into **Claim token**, and click **Connect account**.

   :::tip Connection failed
   If the connection fails, copy a new token from the claim page and try again.
   :::

5. Once the request finishes, click **Open Plex Web**.
6. Sign in to Plex Web with the same account. On first launch, click **Got It!** on the welcome screen to open **Server Setup**.
7. Give the server a recognizable name, such as `Home media`. Clear **Allow me to access my media outside my home**, then click **Next**. This guide reaches the server through a local HTTPS address or LarePass VPN.
8. On **Media Library**, click **Next** to add your library later, then click **Done** to finish setup.
9. Open Plex Web settings, select your server, and open **General**. Check that **Server claimed by** shows the account you just connected.

## Add a video library

A Plex library lets you browse and play media stored in the folders you choose. This example uses a travel recording to create a personal video library. You can follow the same steps to add movie and music libraries.

### Upload your video

1. Open Olares Files and go to `Home/Movies`.
2. Create a subfolder, such as `Family`.
3. Upload your video to that folder. This example uses `Home/Movies/Family/Family-trip.mp4`. You can use your own filename.

Your folders might look like this:

```text
Home/
  Movies/
    Family/
      Family-trip.mp4
  Music/
```

Give each library its own folder. For example, keep personal recordings in `Family` and movie collections in `Films`.

For large files, you can [upload with LarePass Desktop](../manual/olares/files/add-edit-download.md#upload-via-larepass-desktop).

### Select the media folder in Plex

Folders in Olares Files appear under different paths in Plex. When adding a library, select the corresponding Plex path below.

| Folder in Olares | Path in Plex |
| --- | --- |
| `Home/Movies` | `/media/Movies` |
| `Home/Music` | `/media/Music` |

1. Open Plex Web settings and select the server you just connected.
2. Under **Manage**, open **Libraries** and click **Add Library**.
3. For travel recordings and other personal videos, choose **Other Videos** and name the library `Family videos`.
4. Choose the library language and click **Next**.
5. Click **Browse For Media Folder** and select `/media/Movies/Family`.
6. Click **Add**, then **Add Library**, and wait for Plex to scan the files.
7. Open `Family videos` and find the video you uploaded.

Plex has read-only access to `Home/Movies` and `Home/Music`. Use Olares Files to organize, rename, or delete source files.

### Add movie and music collections

You can use the same process for movies and music. Choose the library type and folder that match your content:

| Content | Library type | Example folder |
| --- | --- | --- |
| Travel recordings and home videos | **Other Videos** | `/media/Movies/Family` |
| Movie collection | **Movies** | `/media/Movies/Films` |
| Music collection | **Music** | `/media/Music` |

For **Movies**, include the movie title and release year in filenames to help Plex find matching posters and descriptions. Plex retrieves descriptions in the language you selected for the library.

After uploading more files, open the library's menu and select **Scan Library Files** to add them to Plex.

## Play videos and adjust playback

### Start with original quality

Local video playback is [free](https://support.plex.tv/articles/202526943-plex-free-vs-paid/). Browser playback through [LarePass VPN](#connect-from-another-network) counts as remote playback and requires a subscription.

1. Open `Family videos` and select your travel recording.
2. Play the video at its original quality.

Plex decides whether to transcode based on the file format, player capabilities, and quality setting. Starting with original quality avoids unnecessary quality loss.

### Choose a playback quality

Plex can use the following playback methods:

| Method | What it means |
| --- | --- |
| Direct Play | The player plays the original file without server-side video conversion |
| Direct Stream | The server changes the file container without re-encoding the video |
| Transcode | The server converts video or audio for the player or quality setting |

If playback buffers frequently, lower the resolution or bitrate:

1. In the player, open **Settings** > **Quality**.
2. Click **Show all** to see the available conversion options.
3. Select a quality below the source, such as **Convert to 720p HD (High)** for a 1080p video.

Plex transcodes the video to match the selected quality. To switch back, open the quality menu and select **Play Original Quality**.

### Add and select subtitles

Save the subtitle file in the same folder as the video. For English subtitles, name the files like this:

```text
Home/Movies/Family/
  Family-trip.mp4
  Family-trip.en.srt
```

`Family-trip` matches the video filename, and `en` identifies English subtitles. Plex also supports external ASS subtitles. Save text subtitles as UTF-8 to reduce the risk of garbled characters.

1. Upload the subtitle file to the video's folder through Olares Files.
2. Scan the library in Plex. If the new subtitle is not detected, refresh the video's metadata.
3. Open the video's subtitle options, select the subtitle, and start playback.

Some subtitles need to be burned into the picture, which triggers video transcoding.

## Connect through a local HTTPS address

Use this path when a TV, a Plex app, or a browser on another computer is on the same LAN as Olares. When the phone or computer is on another network, connect with [LarePass VPN](#connect-from-another-network).

Overlay gateway gives Plex Media Server a dedicated local IP. Plex Account Setup builds an HTTPS `plex.direct` address from that IP. Save the address in Plex so apps on the LAN can reach the server.

:::info Network requirements
Run Olares on a native Linux host and connect it with an Ethernet cable. Overlay gateway is unavailable on Wi-Fi and in WSL. For roles and availability, see [Manage overlay gateway](../manual/olares/settings/overlay-gateway.md).
:::

:::warning Use a trusted network
Overlay gateway places Plex directly on your local network. Enable it on a network you trust.
:::

### Get the secure local address

1. Open **Settings** and go to **Network** > **Overlay gateway**.
2. If **Enable overlay gateway** is off, turn it on. When the switch is unavailable, ask the Super admin to enable the system-level service.
3. Under **Applications**, find Plex Media Server, enable overlay gateway, and click **Confirm**.
4. If Plex is running, Olares restarts it to apply the change. Wait until the app returns to **Running** and an IP address appears.
5. Open **Plex Account Setup** from Launchpad. This entrance remains available after you connect your account.
6. In **Local connection**, click **Refresh local address**.
7. When the page shows **HTTPS certificate verified**, click **Copy address**.

The page builds the HTTPS `plex.direct` address from the current Overlay IP.

If Plex stays **Stopped**, open **Settings** > **Applications** > **Plex Media Server** and click **Resume**. Check the IP after it is **Running**.

If **Refresh local address** asks you to enable Overlay, finish the Overlay gateway steps above and refresh again after Plex is **Running**. If the secure address is still missing, check that your Plex account is connected.

### Add the address to Plex

1. Open Plex Web, go to **Settings**, and select your server.
2. Open **Network** and click **Show Advanced**.
3. Paste the copied address into **Custom server access URLs**. Keep existing addresses and separate entries with commas.
4. Click **Save Changes**.

Use the copied address as it is, without `/web/` or `/identity`. Plex Account Setup shows the address and copies it. Save that address in Plex Web before you open a client. The HTTPS address points to the local Overlay IP, so devices on this LAN can use it.

:::warning Update the saved address after Plex restarts
Overlay gateway assigns the local IP dynamically. Restarting Plex, a network change, or an app upgrade can assign a new IP. The HTTPS address includes that IP. **Custom server access URLs** keeps the previous address until you paste a new one.

After Plex returns to **Running**, open **Plex Account Setup** and click **Refresh local address**. When the IP has changed, copy the new HTTPS address and replace the previous local address in **Custom server access URLs**. Update bookmarks and any addresses you entered in a client, then open the Plex client again.
:::

### Open Plex in a browser on another computer

On another computer on the same LAN, paste the copied HTTPS address into the browser and add `/web/` at the end. Sign in with the server owner's Plex account, or with an account that can open the library.

Plex might open an authorization page on `app.plex.tv` that shows the local `plex.direct` address. Confirm that it matches the address you copied, authorize the sign-in, and return to Plex Web. Select your server and open the `Family videos` library.

Open the library on your own server for personal media. Plex's **Movies & Shows** source plays online titles, with different media and subtitles from your library.

### Play media on a TV

1. Connect the TV to the same LAN as the Plex Overlay IP.
2. Open the Plex app on the TV and sign in with the server owner's account, or with an account that can open the library.
3. After you save the custom server address, close and reopen the TV app. Select your server, such as `Home media`.
4. Open `Family videos` and play the recording. Start with original quality. When the TV cannot play the source format directly, Plex can transcode it for the player.

If the server does not appear, check the saved address and whether the account can open the library. The TV needs to reach the Overlay address on its own connection.

## Connect from another network

Use this path when your phone or computer is on a different network from Olares. Enable LarePass VPN and keep it connected during playback. A device on the same LAN can use the [local HTTPS address](#connect-through-a-local-https-address).

:::info Remote video playback requires a subscription
Remote video playback requires an active Plex Pass or Remote Watch Pass on the account you use here. Plex can also treat a VPN connection as remote. See [Plex remote playback requirements](https://support.plex.tv/articles/requirements-for-remote-playback-of-personal-media/).
:::

### Enable LarePass VPN

On the phone or computer you will use for playback, [install LarePass](../manual/larepass/index.md) and sign in with your Olares ID. Make sure it is connected to the Olares device running Plex.

<!--@include: ../reusables/larepass-vpn.md#enable-larepass-vpn-->

On iOS or macOS, allow the VPN configuration if prompted. Wait for the connection to finish before opening Plex. For details, see [LarePass connection status](../manual/larepass/private-network.md#understand-connection-status).

### Play media in a Plex client

1. Install and open the [Plex app](https://www.plex.tv/media-server-downloads/#plex-app) on your phone or computer.
2. Sign in with the Plex account you used to connect the server.
3. In the media sources or server list, select the name you assigned to the server, such as `Home media`.
4. Open the `Family videos` library and play the travel recording.

To watch in a browser on your computer, keep LarePass VPN enabled and open **Plex Media Server** from Launchpad.

## Use hardware-accelerated transcoding

Hardware-accelerated transcoding uses the GPU to decode or encode video. Plex transcodes when the player does not support the source format or you select a lower resolution or bitrate.

These steps require the NVIDIA installation mode and an active Plex Pass on the server owner's account.

### Enable hardware acceleration

1. Open Plex Web settings, select your server, and open **Transcoder**.
2. Show the advanced settings and enable **Use hardware acceleration when available** and **Use hardware-accelerated video encoding**.
3. Under **Hardware transcoding device**, select the NVIDIA GPU assigned to this instance.
4. Click **Save Changes**.

The settings take effect without restarting Plex.

### Check whether playback uses the GPU

1. Play a video and [choose a quality below the source](#choose-a-playback-quality), such as 720p for a 1080p video.
2. In Plex Web, open **Dashboard** from **Activity** and expand the current playback details.
3. Look for `(hw)` next to the video format. This marks hardware-accelerated decoding or encoding.
4. Restore original quality when you have finished checking.

:::tip Black picture while transcoding HEVC in Chrome
If a Dolby Vision or HDR10 video freezes or shows a black picture while Chrome transcodes HEVC, set **Enable HEVC video encoding** to **Never**. Leave **Use hardware acceleration when available** and **Use hardware-accelerated video encoding** enabled, then save and start playback again.

Plex can use H.264 hardware transcoding when the player needs a converted stream. HEVC files can still play at original quality. If one file still freezes, play a different video.
:::

## FAQs

### Why can't I see my uploaded video in Plex?

Check that the upload has finished and that the library uses the correct folder. For example, `Home/Movies/Family` maps to `/media/Movies/Family` in Plex. Then run **Scan Library Files**.

### Why are hardware transcoding settings missing?

Check that the server is connected to an account with an active Plex Pass. In the server's **Transcoder** settings, show the advanced settings to find the hardware acceleration options.

### Why doesn't playback show `(hw)`?

Direct Play and Direct Stream do not re-encode the video, so they do not use hardware transcoding. To check GPU use, [lower the playback quality and open Dashboard](#check-whether-playback-uses-the-gpu).

If the video is transcoding without `(hw)`, check that you selected NVIDIA mode during installation and enabled both hardware acceleration options under **Transcoder**.

### Why are subtitles missing or garbled?

Keep the subtitle and video in the same folder. Check that the subtitle filename matches the video, as in `Family-trip.mp4` and `Family-trip.en.srt`, then scan the library and refresh the video's metadata. If the text is garbled, save the subtitle as UTF-8 and upload it again. See the [Plex local subtitle guide](https://support.plex.tv/articles/200471133-adding-local-subtitles-to-your-media/) for supported formats and naming rules.

### How do I restart Plex when needed?

1. In Olares, navigate to **Settings** > **Applications** > **Plex Media Server** and click **Stop**.
2. Wait until the app has stopped, then click **Resume** to start it again.

See [Manage stopped applications](../manual/olares/market/market.md#how-to-resume-my-application) if Olares cannot resume the app because resources are in use.

### Why can't the Plex client find my server after I enable VPN?

Check that Plex is running and LarePass is connected to the Olares device hosting it. In Plex Web, open the server's **General** settings to see which account owns it. Sign in to the client with that account and reopen the server list.

## Learn more

- [Jellyfin](jellyfin.md): Set up an open-source media server on Olares as an alternative to Plex.
- [LarePass private network](../manual/larepass/private-network.md): Find the VPN controls and understand connection states on your phone or computer.
- [Plex library creation guide](https://support.plex.tv/articles/200288926-creating-libraries/): Learn about media types and folder selection.
- [Plex local subtitle guide](https://support.plex.tv/articles/200471133-adding-local-subtitles-to-your-media/): Learn about subtitle naming, language codes, and formats.
- [Plex hardware acceleration guide](https://support.plex.tv/articles/115002178853-using-hardware-accelerated-streaming/): Review Plex Pass requirements and hardware transcoding status.
