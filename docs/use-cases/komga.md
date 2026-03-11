---
description: Learn how to install, configure, and use Komga on Olares to manage your digital media library.
head:
  - - meta
    - name: keywords
      content: Olares, Komga, media server, digital library, comics, manga
---

# Build your digital library with Komga

Komga is a specialized, open-source media server designed to give you full control over your digital collection of comics, manga, magazines, and e-books. By installing Komga on Olares, you transform your device into a private media hub, organizing your favorite series into a beautiful, searchable library accessible from all your devices.

This guide shows you how to install Komga on Olares, organize your media files for automatic scanning, configure secure user access, and use the built-in reader and metadata editors to enhance your digital reading experience.

## Learning objectives

By the end of this guide, you are able to:
- Install and set up Komga on your Olares device.
- Prepare and organize your media files in the Olares Files app.
- Create libraries and scan for content.
- Read and edit metadata to polish your books.
- Manage user accounts and configure access permissions.

## 1. Install Komga

1. From the Olares Market, search for "Komga".

    ![Search for Komga from Market](/images/manual/use-cases/install-komga.png#bordered){width=90%}

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## 2. Set up admin account

1. Open Komga.

    ![Open Komga](/images/manual/use-cases/open-komga.png#bordered){width=50%}

2. Enter the email and password to create your primary administrator account.
3. (Optional) From the **Translation** list, select your preferred language.
4. (Optional) From the **Theme** list, select your preferred background color mode.
5. Click **CREATE USER ACCOUNT**. The **Welcome to Komga** page appears.

    ![Komga welcome page](/images/manual/use-cases/komga-welcome.png#bordered)

## 3. Prepare your media files

Komga scans a dedicated directory to populate your library. You must place your media files in the designated Olares folder for the application to detect them.

1. Open the Files app from the Dock or Launchpad on Olares.
2. Go to **Data > komga > data**.
3. Upload your media files to this directory, or create sub-folders in it to categorize your files.

    ![Komga data directory](/images/manual/use-cases/komga-data-path.png#bordered)

## 4. Create and scan a library

After uploading your files, connect them to the Komga interface by creating a library.

1. In **Komga**, click **ADD LIBRARY** on the main screen, or click <i class="material-symbols-outlined">add_2</i> next to **Libraries** from the left sidebar. The **Add Library** window appears.
2. On the **GENERAL** tab, configure the following settings, and then click **NEXT**:
    - **Name**: Specify a name for the library.
    - **Root folder**: Click **BROWSE** to choose the `/data` folder or a sub-folder you created.

    ![Komga general settings](/images/manual/use-cases/komga-general.png#bordered){width=60%} 

3. On the **SCANNER** tab, configure how Komga identifies your media files, such as the scan interval file types. Click **NEXT**.
4. On the **OPTIONS** tab, set preferences for file analysis and cover image generation. Click **NEXT**.
5. On the **METADATA** tab, select the metadata types to import. Click **ADD**. 

    Komga automatically scans the associated directory and displays your books with cover images and titles. 

    ![Add library completed](/images/manual/use-cases/komga-library-added.png#bordered){width=90%}  

6. To add new media files later, click <i class="material-symbols-outlined">more_horiz</i> next to the target library on the left sidebar, and then select **Scan library files**. The library is refreshed with the new content for read.

    ![Scan library files](/images/manual/use-cases/komga-scan-lib-files.png#bordered){width=50%}

    For example, you add a new file to the `/data` folder, and then click **Scan library files**. The newly added file is displayed in Komga.

    ![Scan and display newly added files](/images/manual/use-cases/komga-scan-new.png#bordered){width=60%}
    
## 5. Read and refine your books

Use the built-in tools to read your media and keep your metadata accurate. 

### Read books

To read a book, click the book cover to launch it in the web-based reader immediately.

### Edit metadata

To modify the details of your books:
1. Hover over the book cover, and then click <i class="material-symbols-outlined">edit</i>.
2. On the **General** tab, define how your books appear in the library list:
    - Enter the **Title** and **Summary** to provide context for your collection. 
    - Adjust the **Sort Number** to ensure series volumes appear in the correct order.
    - Enter the **Release Date** and **ISBN** to track your collection's history.
3. On the **Authors** tab, add contributors by role, such as **Writers**, **Pencillers**, **Inkers**, **Colorists**, and **Letterers**. This makes it easy to filter libraries by your favorite artists.
4. On the **Tags** tab, create custom tags to group books by themes, such as "90s Aesthetics" or "Summer Reading List". This makes it easy to search based on tags.
5. On the **Links** tab, click the plus icon to add external URLs, such as official series websites or fan wikis. This allows for quick access to more information about the book.
6. On the **Poster** tab, drag and drop a new image or choose a file to replace the default book cover. This ensure your digial shelf looks exactly how you want it.

## 6. Manage users

Share your library with family or friends by adding individual accounts with specific access levels.

1. Go to **Server** > **Users**, and then click <i class="material-symbols-outlined">add_2</i>.
2. In the **Add User** window, enter the email and password for the new user.
3. Choose the appropriate **Roles** to define what the user can do:

    - **Administrator:** Grants the user full control over server settings, libraries, and user management.
    - **File download:** Allows the user to download save the media files to their local device for offline reading.
    - **Page streaming:** Allows the user to read books directly through the web browser or compatible apps.
    - **Kobo Sync** and **KOReader Sync**: Allows the user to synchronize their reading progress, bookmarks, and libraries with Kobo devices or the KOReader application.

3. Click **ADD**.

<!--pending confirm: how to manage their access to specific libraries?--> 

## FAQ

* **Import Limitations:** The "Import" function is generally restricted because it requires files to exist outside of an established library. For the best experience on Olares, upload files directly via the **Files** app as described in step 2.
* **Clean Uninstallation:** If you uninstall Komga and wish to remove all remaining data, go to the **Files** app and delete the `komga` folder located in the **Data** directory.
