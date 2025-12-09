---
outline: [2, 3]
description: Get started with Wise in Olares. Learn to collect content, organize your reading library, add notes, track progress, and customize your reading experience.
---
# Wise basics

Wise helps you build a focused reading workflow on top of your personal information hub. This page walks you through the core actions you'll use every day: collecting content, organizing what to read next, and capturing your own insights.

:::tip Install companion apps
Some advanced features in Wise rely on companion apps in Olares:

- **[Rss Subscribe](https://market.laresprime.olares.com/app/market.olares/rssubscribe)**: Use it to subscribe to RSS feeds directly while browsing web pages.
- **[YT-DLP](https://market.laresprime.olares.com/app/market.olares/ytdlp)**: Use it to download audio and video from supported web pages into Wise.

Wise works without these apps, but in-browser subscription and media download will be unavailable until you install them.
:::

## Build your library

Wise pulls content into your library in two ways:

- **Saved items**: Individual pages, files, audio, or video you capture. These show up under main **Inbox** folder and content types (**Articles**, **Videos**, **Audios**, **Books**, **PDFs**), and can be marked as **Inbox** or **Read Later**.
- **Feeds**: Subscriptions to websites, blogs, podcasts, or channels. New entries appear under **Feeds**, and you can choose which ones to save into your library.

This page focuses on saving and working with individual items. For a deeper look at feeds and subscriptions, see [Subscribe and manage feeds](./subscribe).

### Save items

You can save individual items to Wise in three ways: by uploading files, adding links from within Wise, or saving directly from your browser with the LarePass extension.

#### Upload files

1. Click <i class="material-symbols-outlined">add_circle</i> in the bottom-left menu bar.  
2. Choose **Upload**.
3. Select one or more files from your local computer.

Wise supports various kinds of local files, including PDFs, EPUBs, audio, video, and other document types. Supported formats are organized into the appropriate content type in your library (for example, PDFs under **PDFs**, EPUBs under **Books**, media under **Videos** or **Audios**).

#### Add links

::: tip Handle restricted content
If a link requires login or other access control, Wise may need cookies to fetch it correctly. To configure cookies for protected sites, see **[Manage cookies for Wise](./manage-cookies)**.
:::

1. Click <i class="material-symbols-outlined">add_circle</i> in the bottom-left menu bar.  
2. Choose **Add Link**.
3. Paste or type a URL.

Wise analyzes the link and lists all actions available:
- **Save to library**: The content will be saved as an item in your library and added to **Inbox**.
- **Subscribe to RSS feed**: If Wise detects one or more RSS feeds for the site, they will be listed here. Select the feed you want to follow, and new items from that feed will be automatically added to **Feeds**(see [Subscribe and manage feeds](./subscribe)).
![Subscribe to RSS feed](/images/manual/olares/wise-add-link-subscribe.png#bordered){width=300}
- **Download file**: If Wise detects videos or other downloadable files on the page, this option will appear. Select the file you want to download to save it for offline access (requires **[YT-DLP](https://market.laresprime.olares.com/app/market.olares/ytdlp)**).
![Download files](/images/manual/olares/wise-add-link-download.png#bordered){width=300}

Newly saved items will appear under their content type. You can then decide whether their status should be **Inbox** or **Read Later**.

#### Save from browser with LarePass extension

You can also save content to Wise directly from your browser using the LarePass extension, without opening Wise first.

:::tip Install LarePass browser extension
LarePass extension is currently available for Chrome browsers only. You can download it from the [official page](https://www.olares.com/larepass).
:::

To save content from a web page into Wise:

1. Open the **LarePass** browser extension and select **<i class="material-symbols-outlined">inventory_2</i>**.
2. Under **Save to library**, review the content detected on the current page.
3. Click <i class="material-symbols-outlined">box_add</i> next to the item you want to save.
![Save content via LarePass extension](/images/manual/olares/wise-larepass-add-to-lib.png#bordered)

Items saved via LarePass are added to your Wise library and appear in the main **Inbox** folder and under the appropriate content type.

### Follow sources with feeds

If you want ongoing updates instead of saving one page at a time, subscribe to a feed:

- Paste a feed URL into **Add Link**.
- Use the **LarePass** browser extension while browsing to detect feeds on the current page and subscribe with one click.
- Import RSS feeds from OPML files.

Once subscribed, new entries appear under **Feeds**. From there, you can open items and save the ones you care about into your library, namely the folder **Inbox**.

For full details on managing subscriptions, importing OPML files, and using LarePass browser extension, see [Subscribe and manage feeds](./subscribe).

### Download audio and video

When you add new audio or video content, Wise automatically creates download tasks and save media files to Olares. This:

- Ensures your media is available offline.
- Protects your library if the original source is removed. 
- Makes it faster to open and play items.

To view all download tasks:

1. Go to **<i class="material-symbols-outlined">settings</i> Settings > Transmission > Download**.
2. Review the list of media downloads and their status.
3. You can:

    - Click <i class="material-symbols-outlined">folder_open</i> to locate a downloaded file in **Files**.
    - Click <i class="material-symbols-outlined">do_not_disturb_on</i> to remove it from the list.

## Manage item status

Every item in Wise has a status that can be either **Inbox** or **Read Later**. You can change an item's status using the icons in the list view or the reader view:

- **<i class="material-symbols-outlined">inbox</i> Inbox**: Indicates the item is active or "to do."
- **<i class="material-symbols-outlined">schedule</i> Read Later**: Indicates the item is saved for future reading.

![Wise content status](/images/manual/olares/wise-item-status.png#bordered)
![Wise content status in list](/images/manual/olares/wise-list-status.png#bordered)

**When you click <i class="material-symbols-outlined">inbox</i>**: The item appears in the main **Inbox** folder and the **Inbox** tab of its specific content type (e.g., **Articles** > **Inbox**). The **Read Later** icon will be greyed out.

**When you click <i class="material-symbols-outlined">schedule</i>**: The item is removed from the main **Inbox** folder and moves to the **Read Later** tab of its specific content type (e.g., **Articles** > **Read Later**).

::: tip 
Changing an item's status instantly updates its location across all folders. You can switch between **Inbox** and **Read Later** at any time. 
:::

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

### Use tags

Tags allow you to add flexible labels to your content for easy retrieval later.

1. On the list page, click <i class="material-symbols-outlined" style="font-variation-settings: 'wght' 200;">sell</i> on the content card to add tags to it.
2. Select an existing tag, or type a new name to create one.

![Tags](/images/manual/olares/wise-tags.png#bordered){width=600}

You can manage all your tags in **<i class="material-symbols-outlined">settings</i> Settings > Tags**.

::: tip
Tags become even more powerful when combined with filtered views. See [Organize your knowledge with filters](./filter) to build tag-based views such as "AI articles" or "Design inspiration".
:::

## Search your library

Once you've collected your content in Wise, you can search for particular content themes or entries using aggregated search in Olares.

1. Click <i class="material-symbols-outlined">search</i>in the Dock to open the search window.
2. Specify the search scope to Wise, and enter the keywords to search.
![Search in Wise](/images/manual/tutorials/wise-search.png#bordered)

## Customize appearance

By default, Wise follows your system's light/dark theme settings. You can override this to set your preferred appearance:

1. Click the <i class="material-symbols-outlined">settings</i> in the bottom left corner and select **Preferences**.
2. Under **Theme**, choose your preferred system theme:
   - Light mode
   - Dark mode

Your choice applies to the Wise interface and reader.