---
outline: [2, 3]
description: Set up Jellyfin on Olares for personal media streaming. Manage media files, add libraries, enable hardware acceleration, connect Jellyfin clients, and play media on a TV after enabling overlay gateway.
head:
  - - meta
    - name: keywords
      content: Olares, Jellyfin, media server, self-hosted streaming, overlay gateway, DLNA, TV casting
app_version: "1.0.19"
doc_version: "1.1"
doc_updated: "2026-07-01"
---

# Build your private media server with Jellyfin

Jellyfin is a powerful, open-source software media server that puts you in full control of your media. By installing it on Olares, you can turn your device into a personal streaming platform, organizing movies, shows, and music into a beautiful library accessible securely from anywhere.

## Learning objectives

In this guide, you will learn how to:
- Install and set up Jellyfin on your Olares device.
- Add and organize your media libraries.
- Optimize playback using hardware acceleration.
- Install community plugins.
- Access and stream your media securely from client devices.
- Play media on a TV and cast from Jellyfin after enabling overlay gateway.

## Prerequisites

Before you begin, make sure:
- LarePass app installed on your client devices (desktop or mobile).
- Olares ID imported into the LarePass client.

## Add media to Olares

Before setting up Jellyfin, you need to make sure your media is already available on Olares. You can add it in several ways:
- **Upload files directly**<br>
Upload your media to the `/home/Movies/` folder in Files. For better speed and progress visibility, [use the LarePass desktop client to upload](../manual/olares/files/add-edit-download.md#upload-via-larepass-desktop).
- **Mount an external drive**<br>
After you insert a USB drive to your Olares device, it will be automatically mounted and accessible. Files in it are under the `/external/` directory.
- **Mount a network share**<br> 
If your media is on a NAS or other network server, you can connect it to Olares. For detailed instructions, see [Mount SMB shares](../manual/olares/files/mount-SMB.md).

:::tip Naming conventions
Correct file naming is the key to accurate metadata and beautiful posters in Jellyfin. 
Follow Jellyfin's official guidelines for accurate metadata:
- [Movie naming conventions](https://jellyfin.org/docs/general/server/media/movies/#naming)
- [TV shows naming conventions](https://jellyfin.org/docs/general/server/media/shows/#naming)
:::
:::tip Folder organization
Keep movies and TV shows in **separate folders** for easier management and correct metadata retrieval.
:::

## Install and configure Jellyfin

With your media ready, install Jellyfin and complete its setup wizard.

### Install Jellyfin

1. Open the **Market** app in your Olares web interface.
2. Find **Jellyfin** in the **Fun** category or use the search bar.
3. Click **Get**, then **Install**.
   ![Install Jellyfin](/images/manual/use-cases/jellyfin-install.png#bordered)
4. Once the installation is finished, click **Open** to launch the setup wizard.

### Complete the initial setup

Follow the wizard prompts to finish configuring Jellyfin.
1. Select your preferred display language and click **Next**.
2. Create a username and password for your Jellyfin admin account, and click **Next**.
3. When prompted to set up your media libraries, you can skip this step for now.
4. For metadata, select your preferred language and country and click **Next**.
5. For remote access, keep the default settings (unchecked) and click **Next**. Olares VPN will handle secure remote access.
6. Click **Finish** to complete the setup wizard.
7. You will be taken to the login page. Sign in with the admin credentials you just created.
   
   ![Sign in to Jellyfin](/images/manual/use-cases/jellyfin-sign-in.png#bordered){width=90%}

## Add media libraries

With Jellyfin installed and running, the next step is to tell it where your media is stored.
1. In Jellyfin's sidebar, go to **Dashboard** > **Libraries** > **Libraries**.
2. Click **Add Media Library**.
   
   ![Add Media Library](/images/manual/use-cases/jellyfin-add-media-lib.png#bordered){width=90%}

3. Configure media library settings:
   - **Content type**: Choose the type of media (e.g., Movies, Shows, Music). For folders containing both movies and TV shows, choose **Mixed Movies and Shows**.
   - **Display name**: Enter the name to display for the library.<br>
   - **Folders**: Click + to add the path to your media.<br>
      - **Olares Files**: `/home/movies/<YourMediaFolder>`
      - **External storage**: `/external/<YourMediaFolder>`
4. Click **Ok** to save, and repeat for other media types (e.g., one for "Movies", one for "TV Shows").

Once saved, Jellyfin will automatically scan your folders and begin building your library. This process may take several minutes, depending on the size of your collection.

## Enable transcoding 

To ensure smooth playback for high-resolution videos, enable hardware acceleration. This allows Jellyfin to use your device's hardware for faster, more efficient transcoding.
1. In the Jellyfin **Dashboard** (click the ≡ icon > Dashboard), go to **Playback** > **Transcoding**.
2. Under **Hardware acceleration**, choose your preferred method based on your Olares device's hardware.
   
   ![Enable transcoding](/images/manual/use-cases/jellyfin-transcoding.png#bordered){width=90%}

## Enhance experience with community plugins

You can install plugins to improve metadata, fetch better artwork, and add new features.


The process for installing plugins is the same for all. Here's an example using **Skin Manager**:
1. In the Dashboard, go to **Plugins** > **Catalog**.

   ![Catalog](/images/manual/use-cases/jellyfin-catalog.png#bordered){width=90%}

2. Click the <span style="font-size: 1.1em;">&#9881;</span> icon to go to **Repositories** page, then click **+** to add a new repository.
3. Enter the **Repository Name** and **Repository URL** of the plugin, and click **Save**.

   ![Add plugin repository](/images/manual/use-cases/jellyfin-plugin-repo.png#bordered){width=90%}

4. Click **Ok** to confirm the installation.

   ![Confirm plug installation](/images/manual/use-cases/jellyfin-confirm-plug.png#bordered){width=90%}

5. Return to the **Catalog** tab, find your desired plugin (you may need to refresh) and click **Install**.
   
   ![Catalog plugin](/images/manual/use-cases/jellyfin-catalog-plug.png#bordered){width=90%}
   ![Install plugin](/images/manual/use-cases/jellyfin-plug-install.png#bordered){width=90%}

6. After installation, a prompt will appear to restart Jellyfin. Go to the **Dashboard** page and click **Restart**.

   ![Restart Jellyfin](/images/manual/use-cases/jellyfin-restart.png#bordered){width=90%}

7. After it restarts, return to **Dashboard** > **Plugins** > **My Plugins** to confirm the plugin you installed is listed and has an **Active** status.
   
   ![Plugin status](/images/manual/use-cases/jellyfin-plug-status.png#bordered){width=90%}

After installing plugins, you may need to enable or configure them before they take effect.
Since each plugin behaves differently, check the plugin's **GitHub repository** or **README** for setup details.

## Connect and cast to a TV with overlay gateway

Overlay gateway assigns a local IP to Jellyfin so TVs on the same local network can find it to play media directly or use DLNA casting.

:::warning Use a trusted network
Overlay gateway exposes Jellyfin directly on your local network. Enable it only in a trusted network environment when you need TV discovery or DLNA casting.
:::

:::info Network requirements
Connect your Olares device through wired Ethernet, and keep your TV on the same local network while setting up and playing media.
:::

### Enable overlay gateway for Jellyfin

1. Open **Settings**, then go to **Network** > **Overlay gateway**.

2. Ensure the system-level **Enable overlay gateway** option is turned on. If you cannot turn it on yourself, ask the Super admin to enable it first.

3. Under **Applications**, find Jellyfin and enable overlay gateway for it.

4. In the confirmation dialog, click **Confirm**.

   ![Enable overlay gateway for Jellyfin](/images/manual/use-cases/jellyfin-enable-overlay-gateway.png#bordered){width=90%}

If Jellyfin is running, Olares restarts it to apply the network change. Wait until Jellyfin returns to **Running** before connecting from the TV.

### Connect from the Jellyfin app on your TV

Use the Jellyfin app to play media directly on your TV.

1. Install the latest Jellyfin app on your TV.

2. Open the Jellyfin app on your TV, then choose the option to select or add a server. The app should automatically find Jellyfin through its overlay gateway local IP.

3. In the list, select the Jellyfin server that matches the overlay gateway IP shown for Jellyfin in **Settings**.

4. Sign in with the Jellyfin username and password you created during setup.

5. Open a media item and confirm that the video starts playing on the TV.

### Cast from Jellyfin to a TV

Start a video in Jellyfin on Olares, then send it to a DLNA-capable TV. You do not need to open the Jellyfin app on the TV before casting.

1. Open Jellyfin on Olares.

2. Click the <i class="material-symbols-outlined">menu</i> icon in the upper-left corner, then click **Dashboard** in the sidebar.

   ![Jellyfin dashboard](/images/manual/use-cases/jellyfin-dashboard.png#bordered){width=90%}

3. In the dashboard sidebar, scroll down to **Plugins**, then click **Catalog**.

4. Search for "DLNA", open the DLNA plugin result, and click **Install**.

5. After the plugin installs successfully, restart Jellyfin from Olares:

   a. Open **Settings**, then go to **Applications** > **Jellyfin**.

   b. Click **Stop**.

   c. After Jellyfin stops, click **Resume**.

6. Reopen Jellyfin on Olares.

7. Open the video you want to cast.

8. Click the <i class="material-symbols-outlined">cast</i> icon (**Play on**) in the upper-right corner. Jellyfin should automatically find your TV.

   ![Select a TV from Play on in Jellyfin](/images/manual/use-cases/jellyfin-play-on-tv.png#bordered){width=90%}

9. Select the TV and confirm that the video starts playing on the TV.

## Access your media library through Jellyfin clients
### Get the endpoint for Jellyfin

After Jellyfin is set up and your libraries are ready, you can connect from your client devices and start streaming your media.

:::info Enable LarePass VPN
Before you begin, make sure LarePass VPN is enabled.

If not, see [Enable VPN on LarePass](../manual/larepass/private-network.md#enable-vpn-on-larepass).
:::

1. On Olares, open Settings, then go to **Application** > **Jellyfin**.
2. Under **Entrances**, click **Jellyfin**.
3. Make sure that **Authentication level** is set to **Internal**. If you change the setting, click **Submit**.
4. Under **Endpoint settings**, copy the URL displayed in **Endpoint**. Use this address as the server URL in your Jellyfin client.

   ![Jellyfin endpoint](/images/manual/use-cases/lp-endpoint-jellyfin.png#bordered){width=90%}

### Connect your Jellyfin client

Assume you've already installed the official [Jellyfin client app](https://jellyfin.org/downloads/) on your device.

1. Open the Jellyfin client app on your device.
2. Click **Add Server**.

   ![Add server](/images/manual/use-cases/jellyfin-add-server.png#bordered){width=90%}

3. Paste your Jellyfin URL you just copied into the client and click **Connect**.
   
   ![Connect to server](/images/manual/use-cases/jellyfin-connect.png#bordered){width=90%}

4. Sign in with your Jellyfin admin account.

You should now see your media libraries displayed in the app.

:::tip 
For the best experience, keep your LarePass VPN connection active when accessing Jellyfin remotely. This ensures you can always connect to your Jellyfin server securely. 
:::
