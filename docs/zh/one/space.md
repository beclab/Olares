---
outline: [2,3]
description: Learn how to link your device to Olares Space and monitor your cloud traffic and backup storage usage.
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, Olares Space, monitor traffic usage, billing
---

# Monitor traffic in Olares Space <Badge type="tip" text="10 min" />

Olares Space is the cloud-based management platform for your device. While your Olares device runs locally, Olares Space allows you to monitor cloud traffic usage and manage billing from any web browser.

This guide explains how to link your Olares device to Olares Space and track your traffic usage.

:::info Usage and billing
If you use cloud services such as remote access (accessing your device from outside your home) or cloud backups, monitoring your usage is important. These services might incur charges if you exceed the limits of your subscription plan.
:::

## Learning objectives

- Link your Olares device to your Olares Space account.
- Log in to the Olares Space.
- Monitor cloud traffic consumption.

## Before you begin

Ensure you have the LarePass app installed on your mobile device and logged in with your Olares ID.

## Step 1. Link Olares Space

Before you can view statistics in Olares Space, you must authorize Olares Space to access your device's status. You do this through the LarePass mobile app.

1. Open the LarePass app on your mobile and go to **Settings** > **Integration**.

   ![LarePass Settings](/images/one/larepass-settings-integration.png#bordered)

2. Tap <i class="material-symbols-outlined">add</i> in the top right corner.

   ![Integration](/images/one/larepass-integration.png#bordered)

3. Tap **Space** from the list.

   ![Integration to add Space](/images/one/larepass-integration-add.png#bordered)

   When completed, your Olares Space account is linked to your physical device, and the Olares Space account is displayed in the  **Integration** list.

   ![Olares Space account integrated](/images/one/larepass-integration-add-space.png#bordered)   

## Step 2. Log in to Olares Space

Access the Olares Space dashboard via a web browser.

1. Go to https://space.olares.com/.

   ![Olares Space login page](/images/one/olares-space-login.png#bordered)

2. Log in to Olares Space:
    
    a. In the LarePass app, go to **Settings**, and then tap the scan icon.

      ![Scan to log in](/images/one/scan-olares-space.png#bordered)

    b. Scan the QR code on the computer screen and tap **Confirm**. Now you have logged in to Olares Space.

<!--## Step 3. Monitor system usage

Olares Space separates your local hardware monitoring from your cloud traffic usage.
### Check device health

The landing page is the **Olares** tab. Use the **Resource Monitor** section to check the real-time status of your connected device:
- CPU and memory: Shows how much computing power your device is currently using.
- Disk: Shows the total storage used on the device's hard drive.
- Active hosts: Lists the connection status and uptime of your specific device.

![Olares Space Olares tab](/images/one/olares-space-olares-tab.png#bordered)-->
## Step 3. Check traffic usage

Monitor how much data you have used for remote access to avoid exceeding the monthly limit of your subscription.

1. From the left navigation pane, click **Usage & billing**.

      ![Olares Space traffic details](/images/one/olares-space-traffic-usage.png#bordered)

2. On the **Usage** tab, check the **Traffic details** section. By default, traffic details of the latest billing cycle is displayed.

   - Progress bar: Shows exactly how much data you have consumed against your plan's limit. For example, 0.05 GB/2.0 GB.
   - Daily chart: A bar chart displaying your data usage day by day, helping you spot sudden increases in activity.

3. If you need to check usage from a previous billing cycle:

   a. Click the date range dropdown list.

      ![Olares Space traffic by month](/images/one/space-traffic-filter.png#bordered)

   b. Select a specific billing period to view the traffic history for that month. For example, 2025-12-02 ~ 2026-01-02.

## Resources

- [Olares Space plans and pricing](https://space.olares.com/plans)
- [Manage Olares](../space/manage-olares.md)
- [Olares Space billing](../space/billing.md)