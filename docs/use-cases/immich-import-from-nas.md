---
outline: deep
description: Learn how to mount a NAS shared folder in Olares and import photos into Immich as an external library.
head:
  - - meta
    - name: keywords
      content: Olares, Immich, photo backup, self-hosted photos, photo management, face recognition, smart search
app_version: "1.0.15"
doc_version: "1.0"
doc_updated: "2026-04-09"
---

# Import photos from NAS

If you have photos stored on a NAS device, you can mount the NAS shared folder in Olares and import them into Immich as an external library.

:::info
This tutorial uses Synology NAS as the example. The process might differ for other NAS brands.
:::

## Prerequisites

- Immich is updated to V1.0.15 or later.
- The Olares device and the NAS are on the same local network.
- Shared folder configurations:
  - The shared folder is configured to allow read/write access over the local network (SMB).
  - The **Hide this shared folder in "My Network Places"** option is not selected for the shared folder.

## Step 1: Mount the NAS shared folder to Olares

1. Open Files, click **External**, and then click **Connect to server**.

   ![Connect to server in Files](/images/manual/use-cases/immich-connect-server.png#bordered)

2. In the **Connect to server** window:

    a. Enter the NAS IP address in SMB format (e.g., `//192.168.50.156/`), and then click **Confirm**.

    b. Enter your NAS user name and password, and then click **Confirm**.

      ![Connect to server in Files](/images/manual/use-cases/immich-nas-login.png#bordered){width=60%}
  
    c. Select the folder to mount (`/CZ-test` in this case), and then click **Confirm**.

      ![Connect to server in Files](/images/manual/use-cases/immich-nas-select-share.png#bordered){width=60%}

      Once connected, the shared folder will appear in the **External** directory.

      ![NAS mounted to Files](/images/manual/use-cases/immich-nas-mounted.png#bordered)

## Step 2: Add the folder to an Immich external library

1. Open Immich, click your user avatar in the upper-right corner, and then select **Administration**.
2. Click **External Libraries** from the left sidebar. 
3. Create a new library or use an existing one.
4. In the **Folders** area, click **Add**.
5. Enter the import path. The path format is `/external_storage/` followed by the directory name you mounted in Files. In this case, it is:

   ```text
   /external_storage/CZ-test
   ```
   ![NAS mounted to Files](/images/manual/use-cases/immich-add-nas-folder.png#bordered)

6. Click **Add**.
7. Click **Scan** in the upper-right corner to start scanning. 

    Once the scan finishes, photos from the NAS appear in the photos timeline.

    :::tip Scanning large folders
    If the folder contains many files, scanning might take a while and consume significant NAS disk I/O. You can go to **Administration** > **Job Queues** and pause some tasks in it to speed up processing.
    :::




