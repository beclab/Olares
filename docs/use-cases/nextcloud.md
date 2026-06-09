---
outline: deep
description: Learn how to use Nextcloud on Olares to mount local storage, share folders and albums, configure email notifications, and sync files from mobile and desktop clients.
head:
  - - meta
    - name: keywords
      content: Olares, Nextcloud, file management, file sharing, self-hosted cloud, SMTP, mobile sync, productivity
app_version: "1.0.15"
doc_version: "1.0"
doc_updated: "2026-06-05"
---

# Manage files and collaborate with Nextcloud

Nextcloud is an open-source, self-hosted file sync and collaboration platform. It provides file storage, sharing, and real-time collaboration, along with apps for calendars, contacts, email, and more.

Running Nextcloud on Olares lets you access Olares storage from the Nextcloud interface, share files and photo albums through public links, and keep files synced across your devices.

## Learning objectives

In this guide, you will learn how to:
- Install Nextcloud, sign in, and mount local Olares storage.
- Share folders and photo albums with public links.
- Configure SMTP for email notifications.
- Connect the Nextcloud mobile or desktop app to your server.

## Install Nextcloud

1. Open Market and search for "Nextcloud".

   ![Nextcloud](/images/manual/use-cases/nextcloud.png#bordered)

2. Click **Get**, then **Install**. 
3. When prompted, set the environment variables:

   - **NEXTCLOUD_ADMIN_USER:** Enter the administrator username.
   - **NEXTCLOUD_ADMIN_PASSWORD:** Enter the administrator password.

   ![Configure Nextcloud environment variables](/images/manual/use-cases/nextcloud-configure-env.png#bordered){width=95%}

4. Click **Confirm**, and wait for installation to complete.

## Log in to Nextcloud

After installation completes, open Nextcloud from Launchpad. On the login page, sign in with the administrator username and password you set in `NEXTCLOUD_ADMIN_USER` and `NEXTCLOUD_ADMIN_PASSWORD`.

<!-- ![Nextcloud login page](/images/manual/use-cases/nextcloud-login.png#bordered) -->

## Mount local storage

Olares exposes its home and external storage to the Nextcloud container as `/home_storage` and `/external_storage`. To browse them in Nextcloud, enable External storage support and add both paths as local storage.

### Enable External storage

1. In the Nextcloud web UI, click your avatar in the upper-right corner, and then select **Apps**.
2. Under **Featured apps**, find **External storage support** and enable it.

   ![Enable external storage support](/images/manual/use-cases/nextcloud-external-storage.png#bordered)

### Add mount points

1. Click your avatar in the upper-right corner, and then select **Administration settings**.
2. In the left sidebar under **Administration**, click **External storage**.
3. Add two local storage entries, then click <i class="material-symbols-outlined">check</i> to save them.

   | Folder name | External<br> storage | Authentication | Configuration | Available for |
   |:--|:--|:--|:--|:--|
   | Home | Local | None | `/home_storage` | All people |
   | External | Local | None | `/external_storage` | All people |

   You can change **Folder name** and **Available for** to match your needs.

   ![External storage configuration](/images/manual/use-cases/nextcloud-add-external-storage.png#bordered)

4. Check that both rows show a green check mark. This means Nextcloud can access the mounted storage.

5. Click **Files** in the top-left corner, then click **External storage**. You should see the two mounted directories in your file list.

   ![Mounted directories in file list](/images/manual/use-cases/nextcloud-mounted-directories.png#bordered)

You can now browse and edit files in these directories. Files uploaded through Nextcloud can be deleted from Nextcloud. For files managed by other Olares apps, delete them from Files instead.

## Share files and albums

Nextcloud uses different sharing dialogs for files and albums, but the basic idea is the same: share with specific collaborators, or copy a public link for external access.

### Share a file or folder

1. Click <i class="material-symbols-outlined">folder</i> in the top-left corner to open the file list.
2. Click <i class="material-symbols-outlined" style="transform: scaleX(-1)">
  person_add</i> next to the file or folder you want to share.

   ![Share a folder](/images/manual/use-cases/nextcloud-share-folder.png#bordered)

3. In the **Sharing** panel, choose how to share:

   - **Internal shares:** Enter an account or team name. You can also copy the **Internal link**, which only works for people who already have access.
   - **External shares:** Click **+** next to **Create public link**. Nextcloud creates the link and copies it to your clipboard automatically. You can adjust the link permission from the **Share link** dropdown when needed.

### Share an album

1. Click <i class="material-symbols-outlined">imagesmode</i> in the top-left corner, and then click **Albums**.
2. Open the album you want to share, and click <i class="material-symbols-outlined">share</i> in the upper-right corner.

   ![Share an album](/images/manual/use-cases/nextcloud-share-album.png#bordered)

3. Add people or groups as collaborators, or click **Copy public link** to copy a public album link.

## Configure SMTP

Set up SMTP to enable Nextcloud to send email notifications, password reset links, and sharing invitations.

1. Click your avatar in the upper-right corner, and then select **Administration settings**.
2. In the left sidebar, click **Basic settings**.
3. Under **Email server**, configure the following fields:

   | Field | Value |
   |:------|:------|
   | **Send mode** | `SMTP` |
   | **Encryption** | `None/STARTTLS` or `SSL` |
   | **From address** | Sender email |
   | **Server address** | SMTP host |
   | **Port** | Usually `587` for `None/STARTTLS`, or `465` for `SSL` |
   | **Authentication** | Check if required |
   | **Credentials** | Email and app password or authorization code |

   <!-- ![SMTP configuration](/images/manual/use-cases/nextcloud-smtp.png#bordered) -->

4. Click **Send email** to test the configuration.

## Connect mobile and desktop apps

Use the official Nextcloud client to sync files between your devices and your Olares-hosted Nextcloud server.

1. Download the Nextcloud client for your platform from [nextcloud.com/install](https://nextcloud.com/install/).
2. In the Nextcloud web UI, find your server address:

   a. In the Nextcloud web UI, click your avatar in the upper-right corner, and then select **Personal settings**.

   b. Click **Mobile & desktop**.

   c. Copy the server address shown on the page.

   ![Nextcloud mobile and desktop server address](/images/manual/use-cases/nextcloud-server-address.png#bordered){width=95%}

3. Open the Nextcloud client and click **Log in**.
   
   ![Enter Nextcloud client log in](/images/manual/use-cases/nextcloud-client-log-in.png#bordered){width=55%}

4. Enter the server address, and then click **Next**.
   
   ![Enter Nextcloud server address](/images/manual/use-cases/nextcloud-add-account.png#bordered){width=55%}

5. When the client opens a browser page, click **Log in**.

   ![Connect to your Nextcloud account](/images/manual/use-cases/nextcloud-connect-client.png#bordered)

6. On the **Account access** page, click **Grant access**.

   <!-- ![Grant account access](/images/manual/use-cases/nextcloud-grant-access.png#bordered) -->

7. After authorization succeeds, switch back to the Nextcloud client. Your files are ready to sync.

   <!-- ![Nextcloud authorization success](/images/manual/use-cases/nextcloud-mobile-auth.png#bordered){width=50%} -->

## Learn more

- [Nextcloud documentation](https://docs.nextcloud.com/server/latest/user_manual/en/): Official user manual for Nextcloud.
