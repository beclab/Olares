---
outline: [2, 3]
description: Learn how to migrate Dify Shared from the v2 architecture to the new shared app architecture in Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, Dify Shared, migration, shared app, Olares 1.12.6
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-07-07"
---

# Migrate Dify Shared to the new architecture

Dify Shared is a shared application on Olares for building AI apps, knowledge bases, and agents. Olares 1.12.6 updates the shared app architecture, so you cannot update Dify Shared in place. This guide shows you how to migrate your data to the new Dify Shared app.

## Before you migrate

:::warning Data that cannot be preserved
After migration, the following data cannot be preserved:

- **User accounts**: you must re-create all accounts.
- **Model API keys and Dify system settings**: you must reconfigure them.
- **Processed knowledge base chunks**: you must rebuild all knowledge bases.
- **App logs and conversation history**: only app configurations can be imported.
:::

## Export your Dify data

If multiple users share the Dify instance, make sure every user exports their own data before you uninstall the v2 app.

1. Export app configurations.

   a. Open Dify Studio.
   
   b. Click the expand button on an app card, then select **Export DSL**.
   
   c. Save the `.yml` file for each app.

2. Download documents from your knowledge bases.

   a. Open Files and go to **Cache** > `<your-device-name>` > `difyv2` > `volumes` > `app` > `storage` > `upload_files`.
   
   b. Right-click the files and download them.
   
   :::tip Identify files by upload time
   The file names in this folder are Dify internal names. Identify the files you need by upload time, file format, and size.
   :::
   
   Skip this step if your knowledge base uses an external data source, Notion, or a web site, or if you already have the original documents backed up.

3. Manually record your model configurations and system settings.

## Uninstall Dify Shared

Open Market, go to **My Olares**, and uninstall the v2 **Dify Shared** app.

## Install the new Dify Shared

1. Open Market and search for **Dify**.
2. Click **Get**, then **Install**, and wait for installation to complete.
3. On the app detail page, check **Info** > **Compatibility**. If it shows `Olares >=1.12.6-0`, you are installing the new version.

:::tip First launch takes time
Dify Shared needs about 10 minutes to start for the first time. Wait until the setup page appears before opening it.
:::

4. Open Dify and create the admin account.

## Import your apps

1. Open Dify Studio and select **Import DSL**.
2. Upload the `.yml` files you exported earlier.
3. If Dify detects missing plugins, install them. If you skip plugin installation or it fails, the app configuration is still imported, but you must reconfigure any plugin nodes that did not install.
4. If an app uses a knowledge base, rebuild the knowledge base first, then reconfigure the app to use it.

## Rebuild your knowledge bases

Create new knowledge bases in the new Dify Shared and re-upload your documents or reconnect your data sources.

## Learn more

- [Shared applications](../manual/olares/market/shared-apps.md): Understand the new shared app architecture.
