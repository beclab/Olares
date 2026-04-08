---
outline: deep
description: Learn how to mount a NAS shared folder in Olares and import photos into Immich as an external library.
head:
  - - meta
    - name: keywords
      content: Olares, Immich, photo backup, self-hosted photos, photo management, face recognition, smart search
app_version: "1.0.15"
doc_version: "1.0"
doc_updated: "2026-04-10"
---

# Import photos from NAS

If you have photos stored on a NAS device, you can mount the NAS shared folder in Olares and import them into Immich as an external library.

## Prerequisites

- Immich is updated to V1.0.15 or later.
- The Olares device and NAS are on the same local network.
- The NAS shared folder has LAN access permissions enabled.

:::info
This tutorial uses Synology NAS as the example. The process might differ for other NAS brands.
:::

## Step 1: Mount the NAS shared folder

1. On your NAS, create a shared folder and ensure the **Hide this shared folder in "My Network Places"** option is not selected.

2. In Olares, open Files, click **External**, and then click **Connect to server**.

3. In the **Connect to server** window, enter the NAS IP address in SMB format (e.g., `//192.168.1.100`.), and then click **Confirm**. 

4. Choose to mount the entire shared folder or a specific subdirectory, then log in with your NAS credentials.

## Step 2: Add to Immich external library

1. In Immich, go to **Administration** > **External Libraries**. Create a new library or use an existing one.

2. Add the import path. The path format is `/external_storage/` followed by the directory name you mounted in Files. For example:

   ```text
   /external_storage/temp/test/
   ```

3. Click Scan in the upper-right corner to start scanning.

:::tip Scanning large folders
If the folder contains many files, scanning might take a while and consume significant NAS disk I/O. You can pause some tasks in the Jobs queue to speed up processing.
:::

4. Once the scan is complete, the NAS photos appear in the Immich timeline.


