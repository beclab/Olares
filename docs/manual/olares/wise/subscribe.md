---
outline: [2, 3]
description: Subscribe to RSS feeds in Wise, choose content to save, and manage or remove your subscriptions.
head:
  - - meta
    - name: keywords
      content: Olares, Wise, RSS feeds, subscribe feeds, OPML import, LarePass extension
---

# Subscribe and manage feeds <Badge type="tip" text="Olares 1.12.8+" />

Subscribe to websites, blogs, podcasts, and other sources in Wise to see their new content in one place. After subscribing, browse **Feeds** and add anything you want to keep to **Inbox** or **Read Later**. This guide covers both steps, along with managing and removing subscriptions.

## Prerequisites

Set up the components for the features you plan to use:

- **Fetch feed updates**: Search for "RSS Subscribe" in Market and install the plugin.
- **Subscribe from your browser**: [Install the LarePass browser extension](../../install-larepass-browser-extension) and connect it to your Olares.

## Add subscriptions

Choose the method that fits where you found the source.

::: tip Handle restricted content
If a link requires login or other access control, Wise may need cookies to fetch it correctly. To configure cookies for protected sites, see **[Manage cookies for Wise](./manage-cookies)**.
:::

### Add a feed with a link

If you have a website or feed URL, you can subscribe using **Add Link**.

1. Click <i class="material-symbols-outlined">add_circle</i> in the bottom-left menu bar and choose **Add Link**.
2. Paste a website or feed URL. If one or more feeds are detected, they will be listed under the **Subscribe to RSS feed** section.
3. Click <i class="material-symbols-outlined">bookmark_add</i> next to the feed you want to subscribe to.

    ![Subscribe to RSS feed](/images/manual/olares/wise-add-link-subscribe1.png#bordered){width=400}

### Import subscriptions from OPML

:::info
OPML files contain subscription lists, not articles or reading history. After an import, Wise fetches the latest available content from those feeds.
:::

1. Click <i class="material-symbols-outlined">settings</i> in the bottom left corner and select **Preferences**.
2. Under **Import/Export**, click **Import feeds from OPML**.
3. Select your OPML file and open it.

### Subscribe with the LarePass extension

The [LarePass browser extension](https://www.olares.com/larepass) lets you subscribe to feeds while you browse, without switching back to Wise.

1. Open the LarePass browser extension and click the "Collect" icon.
2. If one or more feeds are detected, they will be listed under the **Subscribe to RSS feed** section.
3. Click <i class="material-symbols-outlined">bookmark_add</i> next to the feed you want to subscribe to.

    ![Subscribe to RSS feed via LarePass](/images/manual/olares/wise-larepass-collect2.png#bordered)

## Browse and save

Subscribing brings new content into **Feeds**, but does not save it to **Inbox** or **Read Later**. Content left only in **Feeds** may disappear if the source changes or you unsubscribe, so save anything you want to keep:

1. Open **Feeds** to browse new content. You can open an item directly from there.
2. For content you want to keep, select **Inbox** or **Read Later** in either place:
   - **On the feed card**: Use the actions beside the item. This example shows **Inbox**.

     ![Add a feed item to Inbox from its card](/images/manual/olares/wise-add-to-inbox.png#bordered)

   - **In the reader**: Use the toolbar while viewing the item. This example shows **Read Later**.

     ![Add a feed item to Read Later from the reader](/images/manual/olares/wise-add-to-read-later.png#bordered)

3. Check the relevant view or tab to find the saved item.

## Manage feed settings

Open <i class="material-symbols-outlined">settings</i> **Settings** > **RSS feeds** to find your subscriptions. Use the icons in the **Operations** column to:

- **Refresh** a feed to check for new content.
- **Copy** its RSS link.
- **Edit** its details and update options.
- **Remove** it from your subscriptions. See [Unsubscribe from a feed](#unsubscribe-from-a-feed).

![Manage subscribed feeds in Wise](/images/manual/olares/wise-rss-feeds.png#bordered)

To edit a feed:

1. Click its edit icon to open **Edit RSS feed**.
2. Update the title or description, adjust content and download options, or set the update interval. Then click **Confirm**.

To export your subscriptions as an OPML file:

1. Open **Settings** > **Preferences**.
2. Under **Import/Export**, click **Export feeds to OPML**.

The file contains your feed list, not saved articles or reading history.

## Unsubscribe from a feed

If you no longer want to follow a source, you can unsubscribe from its feed.

:::warning Before you unsubscribe

- **Saved items stay**: Items you saved to your library remain after you unsubscribe. Other items from the feed are removed.
- **Associated files**: The confirmation dialog includes an **Also delete associated files** option. Review it before confirming.

:::

You can remove a feed subscription in several places:

- From the **Info** panel of an item from that feed.
    ![Unsubscribe via Info panel](/images/manual/olares/wise-info-unsub.png#bordered)
- From the reader toolbar while viewing an item.
    ![Unsubscribe via toolbar](/images/manual/olares/wise-toolbar-unsub.png#bordered)
- From <i class="material-symbols-outlined">settings</i> > **RSS feeds** in Wise.
    ![Unsubscribe in RSS feeds](/images/manual/olares/wise-unsub-rss-feeds.png#bordered)
