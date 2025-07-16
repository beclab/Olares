---
outline: [2, 3]
description: Connect Olares Space and third-party services to enhance functionality. Learn how to integrate, authorize, and manage connected services for seamless data synchronization.
---

# Manage integrations in Settings

The Integration section in **Settings** provides a centralized view of all third-party services connected to your Olares system. It also allows you to manually configure cloud object storage using API credentials.

OAuth-based integrations and Olares Space must be connected via the LarePass app. See the [Integration guide of LarePass](../../larepass/integrations.md) for details.


## View and manage existing integrations

1. Open **Settings** from the Dock or Launchpad.
2. Go to **Integration** from the left-hand menu. You’ll see a list of currently authorized services. 
3. Click an integration card to show its connection status and available actions.
4. In the **Account settings** page, click **Delete** to remove the integration.

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








