---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/linkwarden
outline: [2, 3]
description: Save and preserve useful web pages on Olares
head:
  - - meta
    - name: keywords
      content: "Olares, Linkwarden, self-hosted, bookmarks, web archive, collections"
app_version: "0.0.3"
doc_version: "1.0"
doc_updated: "2026-10-29"
---

# Build a reading library with Linkwarden

Linkwarden saves links in collections and preserves copies of web pages for later reading. Run it on Olares to keep your bookmark library and archived content on your own device. This example collects a few pages for a home project.

## Learning objectives

- Create your Linkwarden account.
- Organize saved links with a collection and tags.
- Open a preserved page and find a saved link again.

## Prerequisites

- Olares 1.12.7 or later.
- A browser connected to your [Olares private network](../manual/larepass/private-network.md).

## Install and sign in

1. Open Market, search for **Linkwarden**, and install it.
2. Open Linkwarden from Launchpad.
3. Select **Sign up** and create an account for this instance, then sign in.

Your Olares login protects the application entrance. The account you create inside Linkwarden belongs to this bookmark library.

## Create a collection and save links

1. Create a collection named **Home projects**.
2. Add a link to a useful page, such as `https://docs.olares.com/`.
3. Choose **Home projects**, give the link a clear title and add a tag such as `Olares`.
4. Save the link and repeat with other pages you want to keep together.

![A private reading collection in Linkwarden](/images/manual/use-cases/linkwarden-library.png#bordered)

## Read a preserved page

Use a saved link's preservation controls to open **Readable**. Select other available formats from the top bar. If your browser does not preview a PDF, download it and open the file locally. The background worker needs time to capture the page. A site that blocks automated access or requires its own login might not produce a complete copy.

Use the collection and tag filters, or search for part of the title, to return to a saved page.

## Keep your library

Bookmark and account records are stored in Olares PostgreSQL. Archived pages and search data persist in the application's private data directory. Include both database and application data in your backup plan.

Sharing a collection changes who can read it. Leave a collection private unless you intend to share its contents. Browser extensions and access from outside your Olares network require additional configuration.

## Learn more

- [Linkwarden documentation](https://docs.linkwarden.app/)
