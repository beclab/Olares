---
outline: [2, 3]
description: Learn how to migrate Xinference from the v2 architecture to the new shared app architecture in Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, Xinference, migration, shared app, Olares 1.12.6
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-07-13"
---

# Migrate Xinference to the new architecture

Xinference is a shared application on Olares for deploying and serving models. Olares 1.12.6 updates the shared app architecture, so you cannot update Xinference in place. This guide shows how to reinstall Xinference and re-download your models after upgrading to Olares 1.12.6.

## Before you begin

Previously, Xinference stored all models as local files in its own Cache directory.

Olares 1.12.6 introduces the [Common directory](/manual/olares/files/files-common.md) for managing shared AI models across applications. In the new architecture, Xinference stores models differently based on their source:

- **Models downloaded from Hugging Face** are stored in **Application** > **Common** > **huggingface**, following the official Hugging Face cache structure.
- **Models downloaded from other sources** are stored in **Application** > **Data** > **xinferencesv3**.

Therefore, your existing models cannot be migrated automatically. You must reinstall the app and re-download your models.

## Reinstall Xinference and re-download models

1. Uninstall the previously installed Xinference app. When prompted, do not select **Also remove all local data**.
2. Install the new Xinference app.

   a. Open Market and search for "Xinference".

   b. Click the app card to open the app details page.

   c. Check the **Information** panel. The **Compatibility** field shows `Olares >=1.12.6-0` for the new version.

   d. Click **Get**, then **Install**, and wait for the installation to finish.

3. Open the new Xinference app, and then re-download all models you need. The system will store them to the correct locations automatically.

## Learn more

- [Shared applications](../manual/olares/market/shared-apps.md): Understand the new shared app architecture.
