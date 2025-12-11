---
outline: [2, 3]
description: Learn how to manage content subscriptions in Olares, including adding RSS feeds, importing OPML files, managing subscriptions with LarePass, and organizing your content sources efficiently.
---

# Subscribe and manage feeds

Wise offers flexible subscription options to help you follow your favorite content sources, including articles, blogs, podcasts, and video channels. This guide explains how to add and manage your subscriptions.

## Understand feeds

Subscribing to a feed creates a direct link to a content source. New entries will appear in the **Feeds** list automatically as they are published.

Please note the difference between viewing and keeping:

- **Feeds are for discovery**: You can open and read any entry from the feed list, but these items are temporary. If the feed changes or you unsubscribe, they may no longer be available.
- **Your library is for ownership**: If you find an entry you want to keep or read later, you must **Save to library**. Only saved items become part of your library and can be managed with status, tags, notes, and filters.

## Add subscriptions

You can subscribe to feeds from within Wise or directly from your browser.

::: tip Handle restricted content
If a link requires login or other access control, Wise may need cookies to fetch it correctly. To configure cookies for protected sites, see **[Manage cookies for Wise](./manage-cookies)**.
:::

### Add feeds via link

If you have a website or feed URL, you can subscribe using **Add Link**.

1. Click <i class="material-symbols-outlined">add_circle</i> in the bottom-left menu bar and choose **Add Link**.
2. Paste a website or feed URL.
3. After Wise analyzes the link, look for the **Subscribe to RSS feed** section in the actions list:
   - If one or more feeds are detected, they are listed there.
   - Select the feed you want to follow.

![Subscribe to RSS feed](/images/manual/olares/wise-add-link-subscribe.png#bordered){width=300}

The selected feed will be added to **Feeds**, and new entries from it will be fetched automatically.

### Import feeds from OPML files

If you already use another RSS reader, or someone shares a curated feed list, you can import those subscriptions using an OPML file.

1. Click <i class="material-symbols-outlined">settings</i> in the bottom left corner and select **Preferences**.
2. Under **Import/Export**, click **Import feeds from OPML**.
3. Select your OPML file and open it.

After import, the feeds will appear under <i class="material-symbols-outlined">settings</i> > **RSS feeds**.

:::info
Only new feed subscriptions are imported. Historical content from your previous reader is **not** transferred.
:::

### Use LarePass browser extension

The LarePass browser extension lets you subscribe to feeds while you browse, without switching back to Wise.

:::tip Install LarePass browser extension
LarePass extension is currently available for Chrome browsers only. You can download it from the [official page](https://www.olares.com/larepass).
:::

1. Open the LarePass browser extension and select **<i class="material-symbols-outlined">inventory_2</i>**.
2. Check the **RSS** section:
   - If feeds are detected on the current page, they will be listed there.
3. Click <i class="material-symbols-outlined">bookmark_add</i> next to the feed you want to subscribe to.
![Subscribe to RSS feed via LarePass](/images/manual/olares/wise-larepass-collect.png#bordered)

## Manage subscriptions

You can view, rename, or organize your subscriptions via the central management page.

1. Click <i class="material-symbols-outlined">settings</i> in the bottom left corner.
2. Select **RSS feeds**.
3. Perform the following operations as needed:
    - View feed details and descriptions.
    - Search for specific feeds.
    - Copy RSS links.
    - Edit feed names and descriptions.
    - Remove RSS links.

Changes here are reflected in the **Feeds** section of the Wise sidebar.

## Unsubscribe

If you no longer want to follow a source, you can unsubscribe from content sources.

:::warning
When you unsubscribe from a feed, all entries from that feed are removed from Wise **unless** they have been explicitly saved to your library (for example, by keeping their status as **Inbox** or **Read Later**). This action cannot be undone.
:::

You can remove a feed subscription in several places:

- From the **Info** panel of an entry belonging to that feed.
    ![Unsubscribe via Info panel](/images/manual/olares/wise-info-unsub.png#bordered){width=500}
- From the reader toolbar while viewing an entry.
    ![Unsubscribe via toolbar](/images/manual/olares/wise-toolbar-unsub.png#bordered){width=500}
- From <i class="material-symbols-outlined">settings</i> > **RSS feeds** in Wise.
    ![Unsubscribe in RSS feeds](/images/manual/olares/wise-unsub-rss-feeds.png#bordered){width=500}