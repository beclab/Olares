---
outline: [2, 3]
description: Centralize third-party service connections, SMB network storage credentials, and website cookies in Olares. Extend file storage capabilities and ensure smooth access to subscriptions and automated tasks.
---

# Manage integrations in Settings

The **Integrations** section in **Settings** centralizes all services and authentication credentials connected to Olares. You can perform two main types of operations here:

- **Connect services** – Link cloud storage services such as **Olares Space**, **AWS S3**, and **Tencent Cloud COS** to extend Olares’ storage capabilities.
- **Manage cookies** – Store, import, and delete website cookies to support Olares’ access to subscription-based or restricted content.
- **Manage SMB accounts** – Store and manage usernames and passwords for SMB network shares, making it easier to reference them when mounting SMB shares in Files.

## View and manage connected services

Follow these steps to view or manage your integrations:

1. Open **Settings** from the Dock or Launcher.
2. From the left sidebar, select **Integrations > Link your accounts and data**.
3. Manage the existing integrations:
   - View the list of authorized services and click a card to check its status or manage settings.
   - On the **Account settings** page, click **Delete** to remove an integration.

## Add cloud object storage via API keys

Olares supports manual configuration of AWS S3 and Tencent Cloud COS using API credentials:

1. Navigate to **Settings** > **Integration** and click the **+ Add Account** button in the top-right corner.
2. Select **AWS S3** or **Tencent COS**, then click **Confirm**.
3. In the mount dialog box, fill in the required details:
   - Access Key
   - Secret Key
   - Region
   - Bucket name
4. Click **Next**. You will see a success message if the credentials are valid.

Your connected cloud storage will now appear under the **Cloud storage** section in Files.

Alternatively, you can configure this direction directly within [LarePass](../../larepass/integrations.md#add-a-cloud-storage-using-api-keys).

:::tip Integrations that require LarePass
OAuth-based integrations and **Olares Space** connections must be completed through the **LarePass** app.  
See the [LarePass integrations guide](../../larepass/integrations.md) for details.
:::

## Manage Cookies

Manage cookies under **Settings > Integrations > Cookie Management**.

The Cookie Management page displays all saved cookies, grouped by domain.  
Each entry shows the cookie name, value, expiration date, and associated domain.

![Cookie Management](/images/manual/olares/cookie-management.png#bordered)

You can perform the following actions:

- **Upload cookie** – Paste cookie content into the import dialog. Supported formats include **Netscape**, **JSON**, and **Header String**.  
   :::tip Note  
   (First-time users can only import cookies manually. You can also use the [LarePass extension](../../larepass/manage-knowledge.md#collect-content-via-larepass-extension) to upload cookies from the browser.)
   :::
- **Edit cookie** – Click a specific cookie entry to modify its value in the editor below.
- **Delete cookie** – Remove invalid or expired cookies. You can delete cookies individually or in bulk.

:::warning Keep cookies up to date
When a cookie expires, the system will highlight it in red.  
Expired or missing cookies may cause some subscription or web automation tasks to fail. It’s recommended to check and update cookies periodically.
:::

## Manage SMB accounts

SMB account management allows you to centrally store usernames and passwords for accessing SMB network shares. You can view, add, or delete accounts here for easy reference when mounting SMB shares in Files.

1. To view SMB accounts, go to **Settings** > **Integration** > **SMB account management**. All saved SMB accounts are listed.
2. To add an SMB account:
   
   a. In **SMB account management**, click **Add account**.

   b. In the **Add account** window, enter the username and password required to access the SMB share.

   c. Click **Confirm**. The information is saved and displayed as a new record in the SMB account list.

3. To delete an account, click <i class="material-symbols-outlined">delete</i> on the right of the account record, and then click **Confirm**. The account and its credentials are removed.

:::tip Use SMB accounts to mount shares
- SMB credentials are stored locally and are not uploaded to the cloud. It is recommended to periodically check and update saved credentials to keep them in sync with the server.
- Saved SMB accounts do not automatically mount shared folders, and you cannot select them directly in Files. To mount an SMB share, manually enter the server address along with the username and password saved here in Files. For more information, see [Mount SMB shares](../files/mount-SMB.md).
:::
