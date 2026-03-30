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

Immich is an open-source, self-hosted photo and video backup solution. It supports automatic backup, original-quality storage, and timeline browsing. With built-in machine learning models, Immich can automatically recognize people, places, and objects in your photos, making photo management smarter and more efficient.

Running Immich on Olares gives you a Google Photos-like experience while keeping full control of your data. Combined with native iOS and Android apps, it is ideal for individuals or families building a private photo library.

## Learning objectives

In this guide, you will learn how to:
- Install Immich and set up the admin account.
- Upload, organize, and browse photos with albums, favorites, and smart search.
- Use face recognition and map view to explore your photo library.
- Sync photos from your PC and mobile devices.
- Share photos and albums with others.
- Import photos from a NAS device.

## Install Immich

1. Open Olares Market and search for "Immich".

   <!-- ![Search for Immich from Market](/images/manual/use-cases/immich.png#bordered) -->

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Set up the admin account

1. Open Immich from Launchpad. The welcome page appears.

   ![Immich welcome page](/images/manual/use-cases/immich-welcome.png#bordered)

2. Click **Getting Started**.
3. In the **Admin Registration** step, set the admin email, password, and name. The first registered user becomes the administrator, responsible for administrative tasks, and additional users will be created by admin.

   ![Register the admin account](/images/manual/use-cases/immich-admin-registration.png#bordered){width=50%}

4. Log in with the email and password you just registered. You are landed on the **Photos** page by default.
5. In Olares, navigate to **Settings** > **Applications** > **Immich** and ensure the **Authentication level** is set to **Internal**. This allows devices on your local network to access Immich without additional authentication.

   ![Change authentication level to Internal](/images/manual/use-cases/immich-auth-level.png#bordered){width=50%}

## Manage photos

### Upload photos

1. On the **Photos** page, click the upload area in the center of the page, or click **Upload** in the upper-right corner. 

   ![Upload photos to Immich](/images/manual/use-cases/immich-upload-photos.png#bordered)
2. Select the photo files you want to upload.
   Once uploaded, photos are automatically sorted by date on the main timeline.

   ![Photos uploaded to Immich](/images/manual/use-cases/immich-photos-uploaded.png#bordered)

### View photos and details

To view a photo and its details:

1. On the **Photos** page, locate and click the target photo. The photo preview is opened.
2. Click <i class="material-symbols-outlined">info</i> in the upper-right corner. The **Info** panel is opened on the right, displaying the photo's metadata such as capture date, camera model, and file format.

   ![View photo details and metadata](/images/manual/use-cases/immich-photo-details.png#bordered)

### Favorite photos

To mark a photo as a favorite:
1. On the **Photos** page, hover over the photo, and then click  <i class="material-symbols-outlined">check_circle</i> in the upper left corner to select it.
2. Click <i class="material-symbols-outlined">favorite</i> on the top right. The photo is marked as a favorite and it is added to the **Favorites** in the left sidebar.

   ![Favorite a photo](/images/manual/use-cases/immich-favorite-photo.png#bordered)

### Delete and restore photos

To delete or restore a photo:
1. On the **Photos** page, hover over the photo, and then click <i class="material-symbols-outlined">check_circle</i> in the upper left corner to select it.
2. Click <i class="material-symbols-outlined">more_vert</i> on the top right, and then select **Delete**. The deleted photo is moved to **Trash** in the left sidebar and will be permenently deleted after 30 days.
3. To permanently delete it immediately, click **Trash** in the left sidebar, select the photo, and then click <i class="material-symbols-outlined">delete_forever</i>.

   ![Permenently delete photos](/images/manual/use-cases/immich-trash-delete.png#bordered)

### Restore photos

To restore a deleted photo:
1. Click **Trash** in the left sidebar.
2. Hover over the photo, and then click <i class="material-symbols-outlined">check_circle</i> in the upper left corner to select it.
3. Click **Restore** on the top right.

   ![Restore photos from Trash](/images/manual/use-cases/immich-trash-restore.png#bordered)

### Download photos

To download photos:
1. On the **Photos** page, hover over the target photo, and then click <i class="material-symbols-outlined">check_circle</i> in the upper left corner to select it.
2. Click <i class="material-symbols-outlined">more_vert</i> on the top right, and then select **Download**. The photo is saved to your computer. When multiple photos are selected, Immich packages them into a ZIP file for batch download.

   ![Download a photo](/images/manual/use-cases/immich-download-photo.png#bordered)

## Organize with albums

### Create an album

1. Click **Albums** in the left sidebar, and then click the album creation area in the center of the page, or click **Create album** at the top of the page.

   ![Create a new album](/images/manual/use-cases/immich-create-album.png#bordered)

2. Set the album name and description, and then click **Select photos**.
3. To add an existing photo, select the photo, and then click **Add assets**.
4. To add a photo from your local device, click **Select from computer**.

   The new album appears in the **Albums** list in the left sidebar.

:::tip Quick album creation
You can also select multiple photos from the main timeline and click <i class="material-symbols-outlined">add_2</i> in the upper-right corner to create an album directly.
:::

### Smart search

Immich uses a CLIP-based machine learning model to analyze image content, so you can search based on context. Tha means you can search for people, places, and objects using natural language without relying on metadata of the photo files.

1. At the top of the page, type a search word in the search field. For example, `halloween`.

   ![Smart search for photos](/images/manual/use-cases/immich-smart-search.png#bordered)

2. Results are displayed at the very first or first ones.

   ![Smart search result](/images/manual/use-cases/immich-smart-search-result.png#bordered)

### Face recognition

Immich recognizes faces in your photos and videos and groups them together into people. You can then assign names to these people and search for them. The list of people is shown in the Explore page. Upon clicking on a person, a list of assets that contain their face will be shown.

1. Click **Explore** in the left sidebar to see automatically detected faces grouped by person.

   ![Explore face recognition results](/images/manual/use-cases/immich-face-recognition.png#bordered)

2. Click a face group to view all photos containing that person.
3. Type a name in the **Add a name** field to label this face.

   ![Name a face group](/images/manual/use-cases/immich-face-group-name.png#bordered)

4. To hide the face of the person from the Explore page and the people details page, click <i class="material-symbols-outlined">more_vert</i> on the top right, and then select **Hide person**.

   ![Hide person](/images/manual/use-cases/immich-face-group-hide.png#bordered)

5. To show hidden people again:

   a. On the Explorer page, click **View All**.

   ![View all Immich people](/images/manual/use-cases/immich-face-group-details.png#bordered)
      
   b. On the people details page, click **Show & hide people**.

   c. Click <i class="material-symbols-outlined">visibility_off</i> on the face group.

   d. Click **Done**.

## View photos on the map

Click **Map** in the left sidebar to see all geotagged photos plotted on a map. Immich uses reverse geocoding to convert GPS coordinates into readable location names such as city, state, and country.

- Photos taken with a phone usually include GPS data and appear on the map automatically.
- Photos without location data, such as those from a standalone camera, can be geotagged manually. Open the photo details and enter a location name in the address field.

![View photos on the map](/images/manual/use-cases/immich-map-view.png#bordered)

## Sync photos

### Sync from PC using external libraries

To sync photos from your PC to Immich, first upload them to Files on Olares, then configure Immich to scan that folder as an external library. The following steps use the Pictures folder as an example.

1. Upload your photos to the **Pictures** folder in Files.
2. In Immich, click your user avatar in the upper-right corner and select **Administration**.
3. In the left sidebar, select **External Libraries**, and then click **Create Library**.
4. Set the **Owner** to your admin account and click **Create**. The **New External Library** page opens.

   ![Create an external library](/images/manual/use-cases/immich-external-libraries.png#bordered)
5. In the **Folders** area, click **Add**.
6. Eenter the import path, and then click **Add**. The path is case-sensitive.

   ```text
   /home/Pictures
   ```
7. Click **Scan** in the upper right to start importing photos.

   ![Scan the external library](/images/manual/use-cases/immich-scan-library.png#bordered)

   Once complete, the imported photos appear in the main timeline on the **Photos** page.

8. To configure automatic periodic scanning:

   a. Click **Settings** from the left sidebar, and then expand the **External Library** panel.

   ![Configure library scan settings](/images/manual/use-cases/immich-scan-settings.png#bordered)

   b. To set real-time scanning, expand **Library watching**, and then enable it. Immich will watch for changed files automatically.

   c. To set scheduled scanning, expand **Periodic Scanning**, and then select your preferred scan interval from the **Cron expression presets**.

   d. Click **Save**.

### Sync from mobile devices

1. Download the Immich mobile app.
   - iOS: Search and install "Immich" from the App Store.
   - Android: Download the APK from the [Immich GitHub Releases](https://github.com/immich-app/immich/releases) page or install from [Google Play](https://play.google.com/store/apps/details?id=app.alextran.immich).

2. Log in to the server.

   a. Open LarePass on your mobile device and enable LarePass VPN.

      <!--![Enable LarePass VPN](/images/manual/get-started/larepass-vpn-mobile.png#bordered)-->

   b. Open the Immich app, and then enter your Immich server URL, admin email, and password to log in. 
   
3. If this is your first time using the app please make sure to choose a backup album so that the timeline can populate photos and videos in it.

   a. Tap <i class="material-symbols-outlined">backup</i> in the upper-right corner to open the backup screen. 

   b. Tap **Select** to choose the albums to back up. You can enable **Sync albums** to keep them continuously synced. Scroll down and tap **Start Backup**.

   <!--![Start mobile backup](/images/manual/use-cases/immich-mobile-backup.png#bordered)-->

   Once the backup is completed, the photos appear on the Immich server. From now on, newly taken photos are automatically synced each time you open the Immich app.

   <!--![Photos synced from mobile](/images/manual/use-cases/immich-mobile-sync-result.png#bordered)-->

## Share photos

### Share individual photos

1. Open the Immich web interface, select the photos you want to share from the **Photos** page.
2. Click <i class="material-symbols-outlined">share</i> in the upper-right corner.
3. Specify settings for the share link as needed, such as URL, description, access password and expiration date.
4. Click **Create link**. Immich generates a share link and QR code that others can use to view the shared photos.

   ![Create a share link](/images/manual/use-cases/immich-share-link.png#bordered){width=50%}

### Share albums

1. Click your user avatar and go to **Administration**. Use **Create User** to create accounts for family members or friends.
2. Click **Sharing** in the left sidebar, then click **Create album** in the upper-right corner to create a shared album.
3. Open the shared album and click the share icon to set an access password and configure user permissions.

   <!-- ![Set up a shared album](/images/manual/use-cases/immich-shared-album.png#bordered) -->

Recipients can view photos in the shared album through the link, and upload or download photos if granted the appropriate permissions.

## Import photos from NAS

If you have photos stored on a NAS device, you can mount the NAS shared folder in Olares and import them into Immich as an external library.

### Prerequisites for NAS import

- Immich is updated to Chart version 1.0.15 or later.
- The Olares device and NAS are on the same local network.
- The NAS shared folder has LAN access permissions enabled.

:::info
The steps below use Synology NAS as an example. The process might differ for other NAS brands.
:::

### Mount the NAS shared folder

1. On your NAS, create a shared folder and make sure the **Hide this shared folder** option is unchecked.

   <!-- ![Create a shared folder on NAS](/images/manual/use-cases/immich-nas-shared-folder.png#bordered) -->

2. In Olares, open Files and navigate to **External**.

3. Click **Connect to Server** and enter the NAS IP address in SMB format, for example `//192.168.1.100`.

   <!-- ![Connect to NAS server in Files](/images/manual/use-cases/immich-files-connect-server.png#bordered) -->

4. Choose to mount the entire shared folder or a specific subdirectory, then log in with your NAS credentials.

   <!-- ![Mount the NAS folder](/images/manual/use-cases/immich-files-mount-folder.png#bordered) -->

### Add to Immich external library

1. In Immich, go to **Administration** > **External Libraries**. Create a new library or use an existing one.

   <!-- ![Add library in Immich](/images/manual/use-cases/immich-add-nas-library.png#bordered) -->

2. Add the import path. The path format is `/external_storage/` followed by the directory name you mounted in Files. For example:

   ```text
   /external_storage/temp/test/
   ```

   <!-- ![Set the NAS library path](/images/manual/use-cases/immich-add-library-path.png#bordered) -->

3. Click **Scan** in the upper-right corner to start scanning.

   <!-- ![Scan NAS library](/images/manual/use-cases/immich-scan-nas-library.png#bordered) -->

   :::tip Scanning large folders
   If the folder contains many files, scanning might take a while and consume significant NAS disk I/O. You can pause some tasks in the **Jobs** queue to speed up processing.
   :::

4. Once the scan is complete, the NAS photos appear in the Immich timeline.

   <!-- ![NAS photos in timeline](/images/manual/use-cases/immich-nas-photos-timeline.png#bordered) -->

## Learn more

- [Immich documentation](https://immich.app/docs)
