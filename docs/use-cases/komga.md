---
outline: deep
description: Learn how to install, configure, and use Komga on Olares to manage your digital media library.
head:
  - - meta
    - name: keywords
      content: Olares, Komga, media server, digital library, comics, manga
app_version: "1.0.7"
doc_version: "1.0"
doc_updated: "2026-03-20"
---

# Build your digital library with Komga

Komga is a specialized, open-source media server designed to manage your digital collection of comics, manga, Bandes Dessinées (BD), magazines, and e-books.

This guide shows you how to install Komga on Olares, organize your media files for automatic scanning, and use the built-in reader and metadata editors to enhance your digital reading experience.

## Learning objectives

In this guide, you will learn how to:
- Install Komga and set up an administrator account.
- Prepare and organize media files in the Olares Files app for automatic detection.
- Create and configure libraries to categorize content.
- Read books and refine their metadata.

## Install Komga

1. Open Olares Market and search for "Komga".

    ![Search for Komga from Market](/images/manual/use-cases/install-komga.png#bordered){width=90%}

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Set up the admin account

When opening Komga for the first time, you must register an account. This is the administrator account that gives you full control over the server settings, libraries, and user management.

1. Open Komga.

    ![Open Komga](/images/manual/use-cases/open-komga.png#bordered){width=50%}

2. Enter the email and password to create your primary administrator account.
3. (Optional) From the **Translation** list, select your preferred language.
4. (Optional) From the **Theme** list, select your preferred background color mode.
5. Click **CREATE USER ACCOUNT**. The **Welcome to Komga** page appears.

    ![Komga welcome page](/images/manual/use-cases/komga-welcome.png#bordered)

## Prepare your media files

Komga scans a dedicated Olares directory to populate your library. You must place your media files in the designated folder for the application to detect them.

1. Open the Files app from Launchpad.
2. Navigate to `Application/Data/komga/data/`.
3. Upload your media files to this directory, or create sub-folders in it to categorize your files.

    ![Komga data directory](/images/manual/use-cases/komga-data-path.png#bordered)

## Create a library

After uploading your files, connect them to the Komga interface by creating libraries. A library is a group of books. You can create multiple libraries to separate different types of content.

1. In **Komga**, click **ADD LIBRARY** on the main screen, or click <i class="material-symbols-outlined">add_2</i> next to **Libraries** on the left sidebar. The **Add Library** window appears.

    ![Add library menu](/images/manual/use-cases/add-lib-menu.png#bordered)

2. On the **GENERAL** tab, configure the following settings, and then click **NEXT**:
    - **Name**: Specify a name for the library.
    - **Root folder**: Click **BROWSE** to choose the file location.
        - To include all media files and sub-folders into one large library, select the `/data` folder.
        - To limit the library's content to a specific category, select a sub-folder such as `/data/{sub-folder-name}`.

    ![Komga general settings](/images/manual/use-cases/komga-general.png#bordered){width=60%} 

3. On the **SCANNER** tab, configure how Komga identifies your media files, such as the scan interval and file types. Click **NEXT**.
4. On the **OPTIONS** tab, set preferences for file analysis and cover image generation. Click **NEXT**.
5. On the **METADATA** tab, select the metadata types to import, and then click **ADD**. 

    Komga automatically scans the associated directory and displays your books with cover images and titles. 

    ![Add library completed](/images/manual/use-cases/komga-library-added.png#bordered){width=90%}  

## Edit metadata

Komga automatically pulls embedded metadata from your media files, such as summaries and release dates. But you can manually refine this data to polish your library's appearance and make your library easier to browse.

1. Hover over the book cover, and then click <i class="material-symbols-outlined">edit</i>.
2. In the **Edit metadata** window, use the following tabs to customize your content:

    | Tab | What you can do |
    |:---|:----------------|
    | General |  Define how your books appear in the library list:<ul><li>Enter the **Title** and **Summary** to provide context for your collection.</li><li>Use **Number** to set the volume number of a book within a series.<br> For example, "1" for the first book, "2" for the second one.</li><li>Use **Sort Number** to specify the order in which books are displayed in the <br>series list regardless of titles or dates.</li><li>Enter the **Release Date** and **ISBN** to track collection history.</li></ul>|
    | Authors | Add contributors by role, such as **Writers**, **Pencillers**, and **Inkers**.<br>This allows you to filter libraries by specific artists.|
    | Tags | Create custom tags to group books by themes, such as "90s Aesthetics".<br>This allows you to search based on tags. |
    | Links | Click the plus icon to add external URLs, such as official series websites.<br>This allows for quick access to more information about the book. |
    | Poster | Drag and drop a new image or choose a file to replace the default book cover.<br>This allows you to personalize your digital shelf looks. |
<!--
2. On the **General** tab, define how your books appear in the library list:
    - Enter the **Title** and **Summary** to provide context for your collection.
    - Adjust the **Number** to  
    - Adjust the **Sort Number** to ensure series volumes appear in the correct order.
    - Enter the **Release Date** and **ISBN** to track your collection's history.
3. On the **Authors** tab, add contributors by role, such as **Writers**, **Pencillers**, **Inkers**, **Colorists**, and **Letterers**. This makes it easy to filter libraries by your favorite artists.
4. On the **Tags** tab, create custom tags to group books by themes, such as "90s Aesthetics" or "Summer Reading List". This makes it easy to search based on tags.
5. On the **Links** tab, click the plus icon to add external URLs, such as official series websites or fan wikis. This allows for quick access to more information about the book.
6. On the **Poster** tab, drag and drop a new image or choose a file to replace the default book cover. This ensure your digial shelf looks exactly how you want it.-->

## Read books

Once your library is organized, you can start reading using the built-in Webreader.

1. Find your target book.
2. Open the book using one of the following methods:
    - Hover over the book cover and click <i class="material-symbols-outlined">auto_stories</i>.
    - Click the book title, and then click **READ** on the book details page.

    ![Read book button](/images/manual/use-cases/komga-read-btn.png#bordered){width=90%}  

    The book is opened in the Webreader immediately.

    ![Book opened in Webreader](/images/manual/use-cases/komga-book-opened.png#bordered){width=90%}  

## Scan library files

By default, Komga scans your libraries regularly based on the interval you set during library creation. However, if you add, rename, or move media files and want them to appear immediately, you can trigger a manual scan.

1. Find the target library from the left sidebar.
2. Click <i class="material-symbols-outlined">more_vert</i> next to it, and then select **Scan library files**. The library refreshes and pulls in any new or changed content.

    ![Scan library files](/images/manual/use-cases/komga-scan-lib-files.png#bordered){width=50%}

<!--## Manage users

You can share your libraries with others by adding individual accounts with specific access levels.

1. Go to **Server** > **Users**, and then click <i class="material-symbols-outlined">add_2</i>.

    ![Add users](/images/manual/use-cases/komga-add-user.png#bordered){width=60%}

2. In the **Add User** window, enter the email and password for the new user.
3. Select the appropriate **Roles** to define what the user can do.

    | Role | Permission |
    |:-----|:-----------|
    | Administrator | Grants the user full control over server settings, libraries, and user<br> management. |
    | File download | Allows the user to save the media files to their local device for<br> offline reading. |
    | Page streaming | Allows the user to read books directly through the Webreader or<br> compatible apps. |
    | Kobo Sync | Allows the user to synchronize their reading progress, bookmarks, <br>and libraries with Kobo devices. |
    | KOReader Sync | Allows the user to synchronize their reading progress, bookmarks, <br>and libraries with the KOReader application. |

4. Click **ADD**. The newly added user is displayed on the **Users** page.

    ![User added](/images/manual/use-cases/komga-user-added.png#bordered){width=60%}

5. Click the **Edit restrictions** icon next to the new user.

    ![Edit restrictions](/images/manual/use-cases/komga-edit-restrictions.png#bordered){width=60%}

    a. On the **SHARED LIBRARIES** tab, select which libraries the user can access.

    ![Library restrictions](/images/manual/use-cases/komga-shared-lib.png#bordered){width=60%}

    b. On the **CONTENT RESTRICTIONS** tab, specify the content to access based on age ratings or labels.

    ![Content restrictions](/images/manual/use-cases/komga-content-restrictions.png#bordered){width=60%}

    c. Click **SAVE CHANGES**.
--> 

## FAQs

### How to perform a clean uninstallation?

If you uninstall Komga and want to remove all remaining data:
1. Open the Files app, and go to **Application** > **Data**
2. Delete the `komga` folder. This removes all database configurations and cached metadata.

### Why can't I use the "Import" function in Komga?

The Import feature is designed to move or copy files from a location outside your library (such as the `Downloads` folder) into an existing series. In the Olares environment, the file picker is restricted to the `/data` directory for security and privacy. Since your libraries already reside in this directory, there are no external files for the system to detect, making the import tool unusable.

To add new files, use the Files app to upload your media files to: `Application/Data/komga/data/`.

Once uploaded, Komga detects them during the next scheduled scan, or you can trigger a manual scan by selecting **Scan library files**.
