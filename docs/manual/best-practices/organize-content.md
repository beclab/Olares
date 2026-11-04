---
outline: [2, 3]
noindex: true
description: Build a knowledge hub with Wise on Olares. Collect web pages, manage PDFs and e-books, subscribe to RSS feeds, and organize your content library.
head:
  - - meta
    - name: keywords
      content: Olares, Wise, knowledge hub, LarePass extension, collect content, RSS feeds, PDF, EPUB
---
# Build your knowledge hub with Wise

Managing information across different sources and devices can be a challenge. You might find yourself using multiple tools to bookmark articles, track RSS feeds, or manage documents, only to end up with fragmented workflows.
	
Wise helps you collect and organize content from the web and your devices in one library on Olares.
	
This tutorial guides you on how to leverage Wise and LarePass to collect, organize, and access content across platforms.

## Objectives

By the end of this tutorial, you will learn how to:

- Use LarePass Chrome browser extension or mobile client to collect content directly from the web while you browse.
- Organize your existing files, such as PDFs and EPUBs, by uploading them into Wise.
- Stay up-to-date with your favorite blogs, podcasts, or video playlists by subscribing to RSS feeds.
- Quickly locate and retrieve any piece of content from your curated information hub.

## Before you begin

Before you begin, make sure that:
	
- Your Olares device is activated and actively running.
- Your Olares ID is [secured with mnemonic phrases](/manual/larepass/back-up-mnemonics.md).
- The LarePass app is installed on your phone.

## Install the LarePass browser extension

The LarePass browser extension is the core tool for content discovery and collection. 

::: tip Support for Chrome only
The LarePass extension is currently only available for Chrome.
:::

<tabs>
<template #Install-from-Chrome-Web-Store>

1. Search for LarePass in the Chrome Web Store.

2. Open the details page, and click **Add to Chrome** to install.

3. Log into the LarePass extension by importing your Olares ID:

   a. Open the LarePass extension, and click **Import an account**.

   b. Import your Olares ID using the corresponding mnemonics.

   c. Enter your Olares password to complete the login.
</template>
<template #Install-offline>

1. Visit https://www.olares.com/larepass to manually download the installation file of the LarePass extension and unzip it.

2. In the URL bar, enter `chrome://extensions/` to access the extension management page.

3. Enable **Developer mode** in the upper right corner.

4. Click **Load unpacked**, and select the unzipped LarePass extension folder to finish installing.

5. Log into the LarePass extension by importing your Olares account:

   a. Open the LarePass extension, and click **Import an account**.

   b. Import your Olares ID using the corresponding mnemonics.

   c. Enter your Olares password to complete the login.
</template>
</tabs>

:::tip Quick access
After installation, pin the LarePass extension in Chrome extension menu for quick and easy access.
:::

Once connected to your Olares, LarePass can send content you choose to save to your Wise library.

## Collect web content

You can collect online content, including web articles, videos, podcasts, using the LarePass browser extension, or the mobile client.

### Collect via the LarePass extension

::: tip Uploading cookies for better experience
Some websites restrict access for anonymous users. In such cases, you can upload your cookies to Olares for better experience.

1. Log into the website, open the LarePass extension.
2. Go to **Collect** > **Cookie**, and click **Upload**. You can hover over cookies to view details. If you don't want to upload a specific cookie item, click **X** to unselect it.
:::

To collect web content using the LarePass extension:

1. Open the content page, for example, a CNN article.
2. Open the LarePass extension. The extension will automatically detect the collectable content on the page.

   ![Collect web content](/images/manual/tutorials/wise-collect-web-content.png#bordered)
3. In **Collect** > **Page**, click <i class="material-symbols-outlined">add_box</i> next to the title to add the page to the Wise library.

Once collected, you can find the content in **Library** > **Articles** in Wise. Available media depends on the source and the download options you use.

### Collect via the LarePass mobile client

You can share web links to the LarePass mobile client for content collection.
:::info
The exact steps may vary depending on your operating system and browser. The following uses Safari as an example.
:::

1. Tap **<i class="material-symbols-outlined">ios_share</i>Share** in the browser, then either:
   - Select the LarePass icon in the sharing options, or
   - Tap **LarePass** in **Other Actions**
   
   ![Share to Wise](/images/manual/tutorials/wise-add-articles-via-share.png#bordered)

   You will be redirected to the LarePass app. LarePass will automatically detect the collectable content on the shared page and prompt whether to add it to Wise.
2. Tap **Confirm** to add it to the Wise library.

::: tip Copy URL to share
Alternatively, you can copy the URL directly and open LarePass. LarePass will automatically detect the URL in the clipboard and the collectable content as well.
:::

Once added, you can find the content in **Library** > **Articles** in Wise.

## Upload PDF/E-book content to Wise

You can upload local PDF or EPUB e-books to Wise for centralized knowledge management. This allows you to keep your reading materials, notes, and related content in one place, making it easier to organize, search, and reference them whenever needed.

1. Open **Wise**, click <i class="material-symbols-outlined">add_circle</i> in the menu bar, and select **Upload**.
2. Navigate to the directory that contains the file you wish to upload, select the file, and click **Confirm**.

View your PDFs under **Library** > **PDFs** and EPUBs under **Library** > **Books**.

![View and manage PDF](/images/manual/tutorials/wise-pdf.png#bordered)

::: tip
You can organize content with tags and record your thoughts in notes. For details, see [Get started with Wise](../olares/wise/).
:::

## Subscribe to RSS feeds
You can subscribe to podcasts, blogs, and other supported feeds in Wise. New entries appear in your feed list. To download multimedia attachments automatically, enable that option in Wise's RSS settings or for an individual feed.

### Subscribe via the LarePass extension

To subscribe an RSS feed using the LarePass extension:

1. Open the RSS page you want to subscribe to, for example, the "Paranormal Mysteries" podcast: `https://www.spreaker.com/podcast/paranormal-mysteries--2321086`.
2. Open the LarePass extension. It will automatically detect the RSS source on the page and show the **RSS** tab.

   ![Subscribe to podcast](/images/manual/tutorials/wise-sub-podcast.png#bordered)
3. In the **RSS** tab, find the correct subscription source and click <i class="material-symbols-outlined">bookmark_add</i> to subscribe.

### Manually add an RSS source

To manually add an RSS source to Wise:

1. Copy the RSS subscription link, such as [https://hnrss.org/frontpage](https://hnrss.org/frontpage), an RSS feed for HackerNews frontliners.
2. In Wise, click <i class="material-symbols-outlined">add_circle</i> in the menu bar, and select **RSS feed**. 
3. Paste the URL. Wise will automatically detect the available RSS source.

   ![Manually add RSS](/images/manual/tutorials/wise-add-rss.png#bordered){width=50%}
4. Click **Add** to subscribe.

### Follow a video playlist

If a video platform exposes a feed for a playlist, you can subscribe to it through LarePass and read its entries in Wise. Here's a YouTube example:

1. Log into your YouTube account.
2. In LarePass extension, go to **Collect** > **Cookie**, and click **Upload** to upload the cookies to Olares. Olares needs the cookies of the site to access your subscription and download videos.
   
   ::: tip Enable Auto-Sync
   Cookie data may expire. It is recommended that you enable the **Auto-Sync** feature in the Cookies page to ensure your cookies are updated automatically each time you visit the site.
   :::
3. Under the video you want to watch, click <i class="material-symbols-outlined">more_horiz</i>, select **Save**, and then click **Create New Playlist**. For example, `Save to Wise`.
4. Enter your playlist. The LarePass extension will automatically detect the available RSS feeds and show the **RSS** tab.
5. In the **RSS** tab, select the RSS feed with a link starting with `https://www.youtube.com/feeds/...` and click <i class="material-symbols-outlined">bookmark_add</i> to subscribe.

   ![Subscribe to YouTube playlist](/images/manual/tutorials/wise-youtube-rss.png#bordered)

New entries from the playlist can appear in Wise. To download supported media, check the feed's auto-download setting and the status in **Download list**.

::: tip Follow a specific channel
You can use LarePass to find a channel's feed and subscribe to it in Wise. Media availability depends on the feed and the download service.
:::

### Access your RSS content

To access your RSS content in Wise:
1. From the left-hand menu bar, navigate to **Subscriptions** > **Feeds**.
2. Select an unread RSS item and enter it to watch the video, listen to the podcast, or simply read the article. 

![Access RSS](/images/manual/tutorials/wise-access-rss.png#bordered)

## Search through collected content
<!--@include: ./wise.reusables.md#wise-aggregated-search-->
## Learn more

- [Get started with Wise](../olares/wise/)
- [Subscribe and manage feeds](../olares/wise/subscribe.md)
