---
outline: [2, 3]
description: Learn how to migrate SearXNG from the old architecture to the new shared app architecture in Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, SearXNG, migration, shared app, Olares 1.12.6
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-07-07"
---

# Migrate SearXNG to the new architecture

SearXNG is a shared application on Olares for privacy-focused web search. Olares 1.12.6 updates the shared app architecture, so you cannot update SearXNG in place. This guide shows how to migrate your preference settings to the new SearXNG app after upgrading to Olares 1.12.6.

## Before you begin

SearXNG does not store user data on the server. All preferences are saved in your browser's cookies, such as language, theme, enabled search engines, and plugins. This means you only need to back up your preferences hash before uninstalling the previously installed app.

:::tip Preferences hash
The preferences hash is an encoded string that contains all your SearXNG settings. Copy the hash and paste it into the new app to restore your preference settings.
:::

## Migrate your preference settings

1. Back up your preferences hash.

   a. Open the previously installed SearXNG app.
   
   b. Go to **Preferences** > **Cookies**.
   
   c. Scroll down to the **Copy preferences hash** section, copy the hash code, and save it.
   
   ![Copy SearXNG preferences hash](/images/manual/use-cases/searxng-copy-preferences-hash.png#bordered)

2. Uninstall the previously installed SearXNG app. When prompted, do not select **Also remove all local data**.
3. Install the new SearXNG app.

   a. Open Market and search for "SearXNG".
   
   b. Click the app card to open the app details page.
   
   c. Check the **Information** panel. The **Compatibility** field shows `Olares >=1.12.6-0` for the new version.
   
   d. Click **Get**, then **Install**, and wait for the installation to finish.

4. Restore your preferences.

   a. Open the new SearXNG app.
   
   b. Go to **Preferences** > **Cookies**.
   
   c. Scroll down to the **Insert copied preferences hash (without URL) to restore** section.
   
   d. Paste the hash code you saved earlier, and then click **Save**.

Your SearXNG preference settings are now migrated to the new app.

## Learn more

- [Shared applications](../manual/olares/market/shared-apps.md): Understand the new shared app architecture.
