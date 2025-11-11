---
outline: [2, 3]
description: Step-by-step guide to setting up Jellyfin on Olares for personal media streaming. Learn how to manage media files with LarePass, add libraries, enhance metadata with plugins, enable hardware acceleration, and securely stream your media through Olares VPN from any device.
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
   ![Install Jellyfin](../public/images/manual/use-cases/jellyfin-install.png#bordered)
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
![Sign in to Jellyfin](../public/images/manual/use-cases/jellyfin-sign-in.png#bordered)

## Add media libraries

With Jellyfin installed and running, the next step is to tell it where your media is stored.
1. In Jellyfin's sidebar, go to **Dashboard** > **Libraries** > **Libraries**.
2. Click **Add Media Library**.
![Add Media Library](../public/images/manual/use-cases/jellyfin-add-media-lib.png#bordered)
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
![Enable transcoding](../public/images/manual/use-cases/jellyfin-transcoding.png#bordered)

## Enhance experience with community plugins

You can install plugins to improve metadata, fetch better artwork, and add new features.


The process for installing plugins is the same for all. Here's an example using **Skin Manager**:
1. In the Dashboard, go to **Plugins** > **Catalog**.
![Catalog](../public/images/manual/use-cases/jellyfin-catalog.png#bordered)
2. Click the <span style="font-size: 1.1em;">&#9881;</span> icon to go to **Repositories** page, then click **+** to add a new repository.
3. Enter the **Repository Name** and **Repository URL** of the plugin, and click **Save**.
![Add plugin repository](../public/images/manual/use-cases/jellyfin-plugin-repo.png#bordered)
4. Click **Ok** to confirm the installation.
![Confirm plug installation](../public/images/manual/use-cases/jellyfin-confirm-plug.png#bordered)
5. Return to the **Catalog** tab, find your desired plugin (you may need to refresh) and click **Install**.
![Catalog plugin](../public/images/manual/use-cases/jellyfin-catalog-plug.png#bordered)
![Install plugin](../public/images/manual/use-cases/jellyfin-plug-install.png#bordered)
6. After installation, a prompt will appear to restart Jellyfin. Go to the **Dashboard** page and click **Restart**.
![Restart Jellyfin](../public/images/manual/use-cases/jellyfin-restart.png#bordered)
7. After it restarts, return to **Dashboard** > **Plugins** > **My Plugins** to confirm the plugin you installed is listed and has an **Active** status.
![Plugin status](../public/images/manual/use-cases/jellyfin-plug-status.png#bordered)


After installing plugins, you may need to enable or configure them before they take effect.
Since each plugin behaves differently, check the plugin's **GitHub repository** or **README** for setup details.

## Access your media library through Jellyfin clients

With Jellyfin set up and your libraries ready, the final step is connecting from your client devices so you can start streaming your media.

### Configure Jellyfin authentication level in Olares

To ensure secure connections between Jellyfin and your devices, configure the authentication level through Olares. 

Make sure LarePass VPN (or local network access) is enabled before starting the steps below.
If you haven’t done this yet, see [Enable VPN on LarePass](../manual/larepass/private-network.md#enable-vpn-on-larepass) for details.

1. In your Olares web interface, go to **Settings** > **Application**.
2. Find and click **Jellyfin**. 
3. Go to the **Entrances** tab.
![Entrances](../public/images/manual/use-cases/jellyfin-entrances.png#bordered)
4. Set the **Authentication level** to **Internal**. This is a required step to enable the private endpoint to accept login credentials.
![Authentication level](../public/images/manual/use-cases/jellyfin-auth-level.png#bordered)

### Connect your Jellyfin client

Assume you've already installed the official [Jellyfin client app](https://jellyfin.org/downloads/) on your device.

1. Open the Jellyfin client app on your device.
2. Click **Add Server**.
![Add server](../public/images/manual/use-cases/jellyfin-add-server.png#bordered)
3. In your Olares web interface, go to **Settings** > **Application** > **Jellyfin** > **Entrances** > **Set up endpoint**.
4. Copy the server address shown next to **Endpoint** (e.g., `https://7e89d2a1.laresprime.olares.com`)
![Endpoint](../public/images/manual/use-cases/jellyfin-endpoint.png#bordered)
5. Paste this address into the client and click **Connect**.
![Connect to server](../public/images/manual/use-cases/jellyfin-connect.png#bordered)
6. Sign in with your Jellyfin admin account.

You should now see your media libraries displayed in the app.

:::tip 
For the best experience, keep your LarePass VPN connection active when accessing Jellyfin remotely. This ensures you can always connect to your Jellyfin server securely. 
:::