---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/linkwarden
outline: [2, 3]
description: Save and preserve useful web pages on Olares
head:
  - - meta
    - name: keywords
      content: "Olares, Linkwarden, self-hosted, bookmarks, web archive, collections"
app_version: "0.0.5"
doc_version: "1.2"
doc_updated: "2026-11-01"
---

# Build a reading library with Linkwarden

Linkwarden saves links in collections and preserves copies of web pages for later reading. Run it on Olares to keep your bookmark library and archived content on your own device. This example collects a few pages for a home project.

## Learning objectives

- Create your Linkwarden account.
- Organize saved links with a collection and tags.
- Open a preserved page and find a saved link again.

- Prepare a source-linked brief from preserved pages with an assistant.

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

Sharing a collection changes who can read it. Leave a collection private unless you intend to share its contents. See the client and assistant connection steps below for access from other applications.

## Use your reading library from an assistant

Lares or Hermes Agent can search Linkwarden and read preserved page text through a community MCP adapter. A useful workflow is to collect sources for a project, then ask the assistant to compare the saved material and prepare a brief with source links.

### Get an API token and service address

In Linkwarden, open **Settings** and its **Access Tokens** page. Create a token for the assistant and save it privately. The token grants access as your Linkwarden account.

On a computer with Olares CLI configured for your device, run:

```bash
olares-cli settings apps get linkwarden -o json
```

In `entrances`, find the `api` entry. Use `https://` followed by its `url` as the service origin. API requests use `/api/v1/` on this origin and `Authorization: Bearer <linkwarden-token>`. From your computer, enable [LarePass VPN](../manual/larepass/private-network.md) before connecting.

### Configure a read-only MCP connection

The package provides REST. The optional community adapter is `linkwarden-mcp@0.4.0`. Configure it in the environment where the assistant runs:

| Setting | Value |
| --- | --- |
| Transport | stdio |
| Command | `npx` |
| Arguments | `-y`, `linkwarden-mcp@0.4.0` |
| `LINKWARDEN_URL` | The internal API origin, without `/api/v1/` |
| `LINKWARDEN_TOKEN` | Your Linkwarden token |
| `LINKWARDEN_READ_ONLY` | `true` |
| `LINKWARDEN_ALLOW_TOOLS` | `search_links,get_link_content,list_collections` |

Node.js and npm must be available in that environment. Save the token in the assistant's private MCP environment configuration. Read-only mode limits the adapter's tools, not the underlying token's permissions.

### Prepare a source-linked brief

After connecting, ask:

> Search my Linkwarden library for pages about the home project. Read the preserved text of relevant pages, combine duplicate URLs, and prepare a short comparison with a source link for each conclusion. Say when a page has no readable archive. Do not change bookmarks or collections.

The assistant can use `list_collections`, `search_links` and `get_link_content`. Check its conclusions against the linked sources. A saved bookmark without a completed readable archive may not provide enough text for a comparison.

## Connect a browser extension or another client

A Linkwarden client uses the same server and account as your Olares installation. Use a client that supports a custom self-hosted server address. Its address and sign-in requirements depend on whether it opens the web interface or calls the API directly. The internal API origin above serves only `/api/v1/` and does not provide the web sign-in page. Do not substitute it for the main web entrance in a client that requires browser login.

## Learn more

- [Linkwarden documentation](https://docs.linkwarden.app/)
