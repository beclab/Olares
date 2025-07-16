---
description: Learn how to mount and access cloud storage services in Olares.
---
# Mount and use cloud storage

You can easily mount a cloud storage through the **Integration** function in Olares, and access and manage your cloud files directly in the **Files** application.

![Cloud storage](/images/manual/olares/files-cloud.png)

## Mount a cloud storage

To mount a cloud storage, connect to it in **Integrations** in LarePass or Olares Settings:

* **OAuth-based storage services**: Google Drive and Dropbox. Connect via [**LarePass** app](../../larepass/integrations.md#add-a-cloud-drive-via-oauth).
* **API credential-based services**: AWS S3 or Tencent Cloud Object Storage (COS); Connect via [LarePass app](../../larepass/integrations.md#add-a-cloud-storage-using-api-keys) or [Olares Settings](../settings/integrations.md#add-cloud-object-storage-via-api-keys).

Once connected, the cloud storage will be automatically mounted under **Cloud Drive** in **Files**.

## Access a cloud storage

Once mounted, you can access and manage files just as you would with local storage:

* **Upload / Download** files
* **Preview** supported file types
* **Rename**, **move**, or **delete** files and folders

Changes made in the Files app will sync with your remote storage provider.

## Unmount a cloud storage

You can unmount a cloud storage by removing the corresponding integration:

* [Remove integration in LarePass](../../larepass/integrations.md#disconnect-integrations)
* [Remove integration in Olares Settings](../settings/integrations.md#view-and-manage-existing-integrations)
