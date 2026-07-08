---
outline: [2, 3]
description: Learn how to migrate OnlyOffice from the old architecture to the new shared app architecture in Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, OnlyOffice, migration, shared app, Olares 1.12.6
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-07-08"
---

# Migrate OnlyOffice to the new architecture

OnlyOffice is a shared application on Olares for document editing and collaboration. Olares 1.12.6 updates the shared app architecture, so you cannot update OnlyOffice in place. This guide shows how to migrate your documents to the new OnlyOffice app after upgrading to Olares 1.12.6.

## Before you begin

OnlyOffice currently ships with the Document Server only. The web interface is the official Node.js demo client `onlyofficeclient`. It supports uploading documents and editing them online, but it does not yet support real-time multi-user collaboration or a full account and document management system.

## Migrate your documents to the new app

1. Back up your documents.

    a. Open Files, and then go to **Documents**.
   
    b. Select the documents you uploaded through OnlyOffice, and then download them to another location.

2. Uninstall the previously installed OnlyOffice app. When prompted:

    - Do not select **Also remove all local data**.
    - Select **Also uninstall the shared server (affects all users)**.

3. Install the new OnlyOffice app.

   a. Open Market and search for "OnlyOffice".
   
   b. Click the app card to open the app details page.
   
   c. Check the **Information** panel. The **Compatibility** field shows `Olares >=1.12.6-0` for the new version.
   
   d. Click **Get**, then **Install**, and wait for the installation to finish.

4. Move your documents to the new location.

    a. Open Files, and then go to **Application** > **Data** > **onlyofficev3** > **documents**.
   
    b. Move your backup documents to this directory.
   
    c. Open OnlyOffice from the Launchpad, and verify that the files appear on the home page.

Your OnlyOffice documents are now migrated to the new app.

## Learn more

- [Shared applications](../manual/olares/market/shared-apps.md): Understand the new shared app architecture.
