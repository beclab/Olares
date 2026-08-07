---
outline: [2, 3]
description: Migrate OnlyOffice documents to the new shared app architecture after upgrading to Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, OnlyOffice, migration, shared app, Olares 1.12.6
app_version: "1.1.14"
doc_version: "1.0"
doc_updated: "2026-08-04"
---

# Migrate OnlyOffice to the new architecture

Olares 1.12.6 updates the shared app architecture, so you cannot update an existing OnlyOffice installation in place. This guide shows how to install the new version and move your documents to it.

:::warning
This guide is for an OnlyOffice installation created before the Olares 1.12.6 upgrade. If you installed OnlyOffice after upgrading, use the [OnlyOffice guide](onlyoffice.md) instead.
:::

## Migrate your documents to the new app

1. Back up your documents.

    a. Open Files, and then go to **Documents**.

    b. Select the documents you uploaded through OnlyOffice, and then download them to another location.

2. Uninstall the previously installed OnlyOffice app. When prompted, do not select **Also remove all local data**. This preserves the previous installation data needed for migration.

3. Install the new OnlyOffice app.

   a. Open Market and search for "OnlyOffice".

   b. Click the app card to open the app details page.

   c. Check the **Information** panel. The **Compatibility** field shows `Olares >=1.12.6-0` for the new version.

   d. Click **Get**, then **Install**, and wait for installation to complete.

4. Move your documents to the new location.

    a. Open Files, and then go to **Application** > **Data** > **onlyofficev3** > **documents**.

    b. Move your backup documents to this directory.

    c. Open OnlyOffice from Launchpad and check that the files appear on the home page.

Your OnlyOffice documents are now available in the new app.

## Clean up legacy data

After you confirm that the migration succeeded and the new app works as expected, you can delete the legacy data folder to free up disk space.

:::warning
Deleting the legacy data cannot be undone. Check that every document opens in the new app and keep a separate backup before continuing.
:::

Open Files, go to **Application** > **Data** > **onlyofficev2**, and delete the folder.

## Learn more

- [Edit documents with OnlyOffice](onlyoffice.md): Understand what the Olares app includes and how to use its web client.
- [Shared applications](../manual/olares/market/shared-apps.md): Learn how shared applications work on Olares.
