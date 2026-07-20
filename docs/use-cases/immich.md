---
outline: deep
description: Learn how to install and use Immich on Olares to back up, organize, and share your photos and videos with built-in AI-powered search and face recognition.
head:
  - - meta
    - name: keywords
      content: Olares, Immich, photo backup, self-hosted photos, photo management, face recognition, smart search
app_version: "1.0.15"
doc_version: "1.0"
doc_updated: "2026-03-26"
---

# Manage photos with Immich

Immich is an open-source, self-hosted photo and video backup solution. It supports automatic backup, original-quality storage, and timeline browsing. With built-in machine learning models, Immich can automatically recognize people, places, and objects in your photos, making photo management smarter and efficient.

Running Immich on Olares gives you a Google Photos experience while keeping full control of your data. Combined with its mobile app, it is ideal for individuals or families building a private media library.

## Learning objectives

In this guide, you will learn how to:
- Install Immich and set up the admin account.
- Populate your library via web uploads, mobile backups, and external imports.
- Browse, edit, and manage your photo timeline.
- Use AI-powered smart search and face recognition.
- Share photos and albums with others locally and publicly.

## Install Immich

1. Open Olares Market and search for "Immich".

   ![Search for Immich from Market](/images/manual/use-cases/immich.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Set up the admin account

The first user registered becomes the administrator, responsible for managing the instance and adding other users.

1. Open Immich from Launchpad, and then click **Getting Started**.

   ![Immich welcome page](/images/manual/use-cases/immich-welcome.png#bordered)

2. On the **Admin Registration** page, set the admin email, password, and user name.

   ![Register the admin account](/images/manual/use-cases/immich-admin-registration.png#bordered){width=50%}

3. Log in with the new credentials. You land on the **Photos** page, which displays all photos in a timeline view.
<!--4. Open **Settings**, go to **Applications** > **Immich**, and then ensure the **Authentication level** is set to **Internal**. This allows devices on your local network to access Immich without additional authentication.

   ![Change authentication level to Internal](/images/manual/use-cases/immich-auth-level.png#bordered){width=50%}-->

## Populate your library

Build your library by bringing photos into Immich from multiple sources.

### Upload from computer

Add photos that are stored on your computer.

1. On the **Photos** page of the Immich web UI, click the upload area in the center, or click **Upload** in the upper-right corner. 

   ![Upload photos to Immich](/images/manual/use-cases/immich-upload-photos.png#bordered)
2. Select the photos to upload.
   
   Once uploaded, they are automatically sorted by date on the photos timeline.

   ![Photos uploaded to Immich](/images/manual/use-cases/immich-photos-uploaded.png#bordered)

### Upload from mobile device

Use the Immich mobile app to upload the photos on your mobile device to the Immich server, creating a private backup.

1. Install the Immich mobile app.
   - iOS: Download "Immich" from the App Store.
   - Android: Download the APK from the [Immich GitHub Releases](https://github.com/immich-app/immich/releases) page or install from [Google Play](https://play.google.com/store/apps/details?id=app.alextran.immich).

2. Open LarePass on your mobile device and enable the VPN to ensure a secure connection to your Olares.

      ![Enable LarePass VPN](/images/manual/use-cases/alex-larepass-vpn-mobile.png#bordered)

3. Open the Immich mobile app, and then enter your Immich server URL, admin email, and password to log in. The photos you uploaded from the web UI are displayed.

   :::tip Obtain Immich server URL
   Open Settings, go to **Applications** > **Immich** > **Entrances** > **Immich** > **Endpoint settings**, and then copy the Endpoint address. This is your Immich server address.
      ![Obtain Immich server URL](/images/manual/use-cases/immich-endpoint.png#bordered){width=70%}   
   :::

4. Tap <i class="material-symbols-outlined">backup</i> in the upper-right corner.

   ![Start mobile backup](/images/manual/use-cases/immich-mobile-backup.png#bordered){width=90%}

5. In **Backup Albums**, tap **Select**, and then select the folders to upload.

   ![Select files for mobile backup](/images/manual/use-cases/immich-mobile-backup-select-file.png#bordered){width=90%}

6. Toggle on **Enable Backup**. 

   ![Enable backup on mobile](/images/manual/use-cases/immich-enable-backup.png#bordered){width=90%}

   Once finished, the uploaded photos are displayed in the photos timeline in the Immich web UI. Any new photos added to the selected folder on mobile will be automatically synced to the web UI.

### Upload from Olares Files

If you have photos stored in the Olares Files app, you can configure Immich to scan that folder as an external library. Immich will then index and display them without moving the actual files. 

1. In the Immich web UI, click your user avatar in the upper-right corner, and then select **Administration**.
2. Select **External Libraries** > **Create Library**.
3. Set the **Owner** to your admin account, and then click **Create**.
4. In the **Folders** area, click **Add**.

   ![Create an external library](/images/manual/use-cases/immich-external-libraries.png#bordered)

5. Enter the import path, which is case-sensitive, and then click **Add**. For example,

   ```text
   /home/Pictures
   ```
6. Click **Scan**. Immich displays photos under this directory in the photos timeline.

   ![Scan the external library](/images/manual/use-cases/immich-scan-library.png#bordered)

### Import from NAS

If you have a large collection of photos stored on a local NAS device, you can mount your NAS directly in Olares and map these directories into Immich without duplicating files. 

For detailed steps on setting up an SMB connection and mapping paths, see [Import photos from NAS](./immich-import-from-nas.md).

## Browse and manage photos   

With your library populated, you can now interact with the photos.

### View photos

1. Click the target photo on the timeline to open the preview.
2. To view the details of the photo, click <i class="material-symbols-outlined">info</i> in the upper-right corner.

   Metadata such as capture date, camera model, and file format are displayed on the right.

   ![View photo details and metadata](/images/manual/use-cases/immich-photo-details.png#bordered)

### Edit photos

1. Open a photo preview, and then click <i class="material-symbols-outlined">tune</i> in the upper-right corner.
2. Use the **Editor** panel to crop, rotate, or mirror the photo.

   :::info 
   Immich supports non-destructive editing of photos. This means that any edits you make to an asset do not modify the original file, but instead create a new version of the asset with the edits applied.
   :::
3. Click **Save**.
4. To revert to the original at any time, click **Reset changes** in the same panel.

### Favorite photos

1. Hover over the target photo on the timeline, and then click  <i class="material-symbols-outlined">check_circle</i> in the upper left corner to select it.
2. Click <i class="material-symbols-outlined">favorite</i>. The photo is added to **Favorites** in the sidebar.

   ![Favorite a photo](/images/manual/use-cases/immich-favorite-photo.png#bordered)

### Delete photos

1. Hover over the target photo on the timeline, and then click <i class="material-symbols-outlined">check_circle</i> in the upper left corner to select it.
2. Click <i class="material-symbols-outlined">more_vert</i> on the top right, and then select **Delete**. The deleted photo is moved to **Trash** in the sidebar and will be permenently deleted after 30 days.
3. To permanently delete it immediately, click **Trash** in the left sidebar, select the photo, and then click <i class="material-symbols-outlined">delete_forever</i>.

   ![Permenently delete photos](/images/manual/use-cases/immich-trash-delete.png#bordered)

### Restore photos

1. Click **Trash** in the left sidebar.
2. Hover over the photo, and then click <i class="material-symbols-outlined">check_circle</i> in the upper-left corner to select it.
3. Click **Restore**.

   ![Restore photos from Trash](/images/manual/use-cases/immich-trash-restore.png#bordered)

### Download photos

1. Hover over the target photo on the timeline, and then click <i class="material-symbols-outlined">check_circle</i> in the upper-left corner to select it.
2. Click <i class="material-symbols-outlined">more_vert</i> on the top right, and then select **Download**.

   The photo is saved to your computer. When multiple photos are selected, they are packaged into a `.zip` file for batch download.

   ![Download a photo](/images/manual/use-cases/immich-download-photo.png#bordered)

## Organize photos

Group your media into curated collections or browse your library using a traditional file explorer structure.

### Create an album

Group photos into themed collections for easier access and sharing.

1. Click **Albums** in the left sidebar, and then click the album creation area in the center of the page, or click **Create album** at the top of the page.

   ![Create a new album](/images/manual/use-cases/immich-create-album.png#bordered)

2. Set the album name and description, and then click **Select photos**.
3. To add an existing photo, select the photo, and then click **Add assets**.
4. To add a photo from your local device, click **Select from computer**.

   The new album appears in the **Albums** list in the left sidebar.

### Organize by file structure

Enable folder view to navigate your media files using the original directory hierarchy of your Olares files system.

1. Click your avatar and then click **Account Settings**.
2. Expand **Features** > **Folders**, and then enable it.

   ![Enable folders view](/images/manual/use-cases/immich-enable-folder-view.png#bordered){width=65%}
3. Click **Save**. A **Folders** node is displayed on the left sidebar.
4. Click **Folders**. You can see the files organized in a view similar to file explorer.

   ![Folders view](/images/manual/use-cases/immich-folders-view.png#bordered)

## Search photos

Immich uses built-in AI models to analyze your image content, providing a flexible search experience.

### Search by context

Search for people, places, and objects using natural language without relying on the keywords in the metadata of the photo files.

1. At the top of the page, type a search word in the search bar. For example, `halloween`.

   ![Smart search for photos](/images/manual/use-cases/immich-smart-search.png#bordered)

2. Immich identifies relevant images based on visual content even if they have no manual tags.

   ![Smart search result](/images/manual/use-cases/immich-smart-search-result.png#bordered)

### Search by face

Immich recognizes faces in your photos and videos and groups them together into **People** on the **Explore** page. You can assign names to these people and search for them. 

1. Click **Explore** in the left sidebar to see automatically detected faces grouped by person.

   ![Explore face recognition results](/images/manual/use-cases/immich-face-recognition.png#bordered)

2. Click a face group, and then enter a name in the **Add a name** field to label this face.

   ![Name a face group](/images/manual/use-cases/immich-face-group-name.png#bordered)

3. Once named, you can search for that person directly in the search bar.

### Search by location

Immich clusters your media based on GPS metadata, so you can search by navigating to specific regions on a global map.

1. Select **Map** in the left sidebar to view your photos plotted globally.
2. Use the zoom controls or scroll to a specific country or city.
3. Select a blue location cluster to view the photos associated with that area.

   ![Map view](/images/manual/use-cases/immich-map-view.png#bordered)

## Share photos

Immich supports two types of sharing: public links for external recipients and local collaborative albums for users on the same Olares cluster.

### Share with external users

Create secure public links to share photos with people outside your Olares network.

1. Open the Immich web interface, select the photos you want to share from the **Photos** page.
2. Click <i class="material-symbols-outlined">share</i> in the upper-right corner.
3. Specify settings for the share link as needed, such as URL, description, access password and expiration date.
4. Click **Create link**. Immich generates a share link and QR code that others can use to view the shared photos.

   ![Create a share link for photos](/images/manual/use-cases/immich-share-link.png#bordered){width=40%}

### Share with local members

Collaborate on shared albums with other users on your Olares instance.

1. Add a user.

   a. Click your user avatar, and then click **Administration**.

   b. Click **Users** from the left sidebar, and then click **Create user** to create accounts for family members or friends.

      ![Create user for share](/images/manual/use-cases/immich-share-create-user.png#bordered){width=40%}

   c. Enter related information for the new user account as needed, and then click **Create**.

   d. Return to the homepage by clicking immich at the upper left corner.

2. Invite the user to an album.

   a. Select the album from the left sidebar, and then click <i class="material-symbols-outlined">share</i> in the upper-right corner.

   b. In the **Options** window, click **Invite People**, select the user you want to share the album with, and then click **Add**.

   ![Set up a shared album](/images/manual/use-cases/immich-shared-album.png#bordered){width=40%}   

   c. Under **People**, assign the Editor or Viewer access for the user.

   The invited user can view photos in the shared album when logging in to Immich, and upload or download photos if granted the Editor access.

## Learn more

- [Immich documentation](https://immich.app/docs)
