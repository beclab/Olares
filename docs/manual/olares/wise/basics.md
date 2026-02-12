---
outline: [2, 3]
description: Get started with Wise in Olares. Learn to collect content, organize your reading library, add notes, track progress, and customize your reading experience.
---
# Wise basics

Wise helps you build a focused reading workflow on top of your personal information hub. This page walks you through the core actions you'll use every day: collecting content, organizing what to read next, and capturing your own insights.

This page focuses on saving and working with individual items. For a deeper look at feeds and subscriptions, see [Subscribe and manage feeds](./subscribe).


## Before you begin

To unlock the full potential of Wise, it is recommended to install the following apps from Olares Market:

- **Rss Subscribe**: Use it to subscribe to RSS feeds directly while browsing web pages.
- **YT-DLP**: Use it to download audio and video from supported web pages into Wise.
- **Twitter/X plugin**: Use it to save posts and download attached files from Twitter/X into Wise.

:::tip 
Wise works without these apps, but some features require them, such as in-browser subscriptions, media downloads, and Twitter/X link recognition and saving.
:::

## Build your library

Wise pulls content into your library in two ways:

- **Saved items:** Individual web pages, files, audio, and video that you capture manually. These items appear in your main **Inbox** and are automatically sorted into categories such as **Articles**.
- **Feeds:** Subscriptions to dynamic sources like websites, blogs, and podcasts. New updates appear under **Feeds**, where you can select specific entries to save to your library.

### Save items

You can save individual items to Wise in three ways: 
- Upload files
- Add items via link
- Save from browser with LarePass extension

#### Upload files

Import files directly from your computer, including PDFs, EPUBs, audio, video, and other document types. Wise automatically places each supported format into the right content folder in your library. 

1. Click <i class="material-symbols-outlined">add_circle</i> in the bottom-left menu bar, and select **Upload**.
2. Select one or more files from your local computer.
    :::tip 
    You can also drag and drop files into the Wise interface.
    :::
3. In the Upload files window, select the destination folder, then click **Confirm**.

#### Add items via link

Paste a URL to save articles, videos, or subscribe to feeds.

::: tip Handle restricted content
If a link requires login or other access control, Wise may need cookies to fetch it correctly. To configure cookies for protected sites, see **[Manage cookies for Wise](./manage-cookies)**.
:::

1. Click <i class="material-symbols-outlined">add_circle</i> in the bottom-left menu bar, and select **Add Link**.
2. Paste or type a URL.

    Wise analyzes the link and lists all the available actions:
    - **Save to library**: The content will be saved as an item in your library and added to **Inbox**. Twitter/X posts are supported when the Twitter/X plugin is installed.
    - **Subscribe to RSS feed**: If Wise detects one or more RSS feeds for the site, they will be listed here. Select the feed you want to follow, and new items from that feed will be automatically [added to **Feeds**](./subscribe).
    ![Subscribe to RSS feed](/images/manual/olares/wise-add-link-subscribe.png#bordered){width=300}
    - **Download file**: If Wise detects downloadable media (such as audio, video, or attached files in Twitter/X posts), this option will appear. Select the file you want to download to save it for offline access. 
        :::tip Install helper services
        Some downloads require helper services: 
        - [YT-DLP](https://market.olares.com/app/market.olares/ytdlp) is commonly used to download audio or video from supported pages when downloadable media is available.
        - [Twitter/X plugin](https://market.olares.com/app/market.olares/twitter) is required to download attached files from Twitter/X posts.
        :::
    ![Download files](/images/manual/olares/wise-add-link-download.png#bordered){width=300}

Newly saved items will appear under their content type.

#### Save from browser with LarePass extension

You can also save content to Wise directly from your browser using the [LarePass extension](https://www.olares.com/larepass), without opening Wise first.

1. Open the LarePass browser extension and select the "Collect" icon.
2. Under **Save to library**, review the content detected on the current page.
3. Click <i class="material-symbols-outlined">box_add</i> next to the item you want to save.
![Save content via LarePass extension](/images/manual/olares/wise-larepass-add-to-lib.png#bordered)

Items saved via LarePass are added to your Wise library and appear in the main **Inbox** folder and under the appropriate content type.

### Monitor and manage file tasks

Wise tracks background transfer tasks in two lists:

- **Download list**: Created when you add downloadable media. Wise downloads the files to Olares so you can access them offline.
- **Upload list**: Created when you upload local files into Wise. Wise tracks the upload progress and results.

To manage transfer tasks:

1. Go to **<i class="material-symbols-outlined">settings</i> Settings** > **Download list** or **Upload list**.
2. Use the tabs to filter tasks: 
    - Download list tabs: **All**, **Downloading**, **Complete**, **Failed**
    - Upload list tabs: **All**, **Uploading**, **Complete**, **Failed**.
3. Review the task list and status.
4. You can:
   - Click <i class="material-symbols-outlined">folder_open</i> to locate the transferred file in Files.
   - Click <i class="material-symbols-outlined">do_not_disturb_on</i> to remove it from the list.

## Use reading tools

Wise provides several tools to enhance your reading experience and help you keep track of what matters.

![Wise reading toolbar](/images/manual/olares/wise-reading-toolbar.png#bordered)

### Track reading progress

Wise uses green dot indicators on article covers to help you track unread content. When you open an article, it's automatically marked as read.
![Wise unseen content](/images/manual/olares/wise-unseen-content.png#bordered){width=600}

In the reader toolbar, you can manually toggle between **<i class="material-symbols-outlined">playlist_add_check</i>Seen** or **<i class="material-symbols-outlined">playlist_remove</i>Unseen** to maintain your reading progress.

### Capture notes

You can add private notes to any content in your library:

1. While browsing, click <i class="material-symbols-outlined" style="font-variation-settings: 'wght' 200;">right_panel_open</i> to open the **Info** panel.
2. Type your thoughts in the **Note** section.
3. Click **Save**.

You can edit or delete notes from the same panel at any time.

## Use tags

Tags allow you to add flexible labels to your content for easy retrieval later.

1. On the list page, click <i class="material-symbols-outlined" style="font-variation-settings: 'wght' 200;">sell</i> on the content card to add tags to it.
2. Select an existing tag, or type a new name to create one.

![Tags](/images/manual/olares/wise-tags.png#bordered){width=600}

You can manage all your tags in **<i class="material-symbols-outlined">settings</i> Settings** > **Tags**.

::: tip
Tags become even more powerful when combined with filtered views. See [Organize your knowledge with filters](./filter) to build tag-based views such as "AI articles" or "Design inspiration".
:::

## Search your library

Once you've collected your content in Wise, you can search for particular content themes or entries using aggregated search in Olares.

1. Click <i class="material-symbols-outlined">search</i> in the Dock to open the search window.
2. Specify the search scope to Wise, and enter the keywords to search.
![Search in Wise](/images/manual/tutorials/wise-search.png#bordered)

## Customize appearance

By default, Wise follows your system's light/dark theme settings. You can override this to set your preferred appearance:

1. Click the <i class="material-symbols-outlined">settings</i> in the bottom left corner and select **Preferences**.
2. Under **Theme**, choose your preferred system theme:
   - Light mode
   - Dark mode

Your choice applies to the Wise interface and reader.