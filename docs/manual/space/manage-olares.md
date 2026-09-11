---
outline: [2, 3]
description: Check Olares status, hosts, cloud-service quotas, and traffic usage from Settings or the Olares Space website.
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, system status, active hosts, backup storage, traffic usage, Dashboard
---
# Check Olares status and Olares Space usage

Use Olares Space when you need a remote view of your Olares status or want to check usage for cloud-assisted services such as backup storage and reverse-proxy traffic. For detailed, real-time CPU, memory, disk, network, pod, GPU, and app metrics, use [Dashboard](../olares/resources-usage.md) instead.

The information available depends on where you open it:

| Location | Use it for |
|---|---|
| **Settings** > your avatar > **Olares Space** | The recommended sign-in entry and a quick summary of the linked account, plan, backup storage, and traffic usage |
| [Olares Space](https://www.olares.com/space) | Remote Olares status, active hosts, recent traffic, and billing-cycle usage |
| **Dashboard** in Olares | Detailed local resource and application diagnostics |

## Connect Olares Space in Settings

Connect Olares Space from its dedicated page in Olares Settings. After the connection succeeds, you can view usage on this page, and Olares Space also appears under **Settings** > **Integrations**.

:::warning Use the Olares Space page to connect
Do not start from **Settings** > **Integrations** > **Add account** > **Olares Space**. Account binding is not supported from that entry yet.
:::

1. Open **Settings** in Olares.
2. Click your avatar in the upper-left corner.
3. Select the **Olares Space** card.
4. When the QR code appears, scan it with the LarePass app on your mobile device.
5. Wait for the page to show your Olares Space account and usage information.

## Check the summary in Settings

1. Open **Settings** in Olares.
2. Click your avatar in the upper-left corner.
3. Select the **Olares Space** card.
4. Review the linked account, subscribed plan, backup storage, reverse-proxy service, and traffic usage shown on the page.

Use this view for a quick quota check. Open the Olares Space website when you need status details, a time range, or billing-cycle history.

## Check resource usage in Olares Space

The Olares Space overview provides a remote summary of CPU, memory, and disk usage. Use it to check whether a system that you cannot access locally is online and nearing a resource limit.

1. On the **Olares** page, select the **Overview** tab.

   ![Olares page, overview tab](/images/how-to/space/olares_page_overview.png#bordered)

2. Locate the **Resource Monitor** section. It displays the real-time usage of CPU, memory, and disk.

   | Metric | Description |
   | ------ | ----------- |
   | **CPU (Cores)** | Current CPU usage in cores and the total available cores. |
   | **Memory (GB)** | Current memory usage in GB and the total available memory. |
   | **Disk (GB)** | Current disk usage in GB and the total available disk space. |

## Check active hosts

Check which hosts are currently running in your Olares cluster and their status.

1. On the **Olares** page, select the **Overview** tab.
2. Locate the **Active hosts** section. It displays the hosts currently running in the Olares cluster.

## Check recent traffic usage

Check recent traffic usage to spot sudden increases and avoid exceeding your plan's limit.

:::info
For self-hosted Olares users, monitor traffic statistics for the reverse proxy service. If you exceed the monthly quota, speed is throttled to 5 Mbps. The free alternative is to use [LarePass VPN](../larepass/private-network.md) or host your own FRP server.
:::

1. On the **Olares** page, select the **Usage statistics** tab.

   ![Olares Space traffic usage](/images/how-to/space/olares_usage_statistics1.png#bordered)

2. Locate the **Traffic Usage** section. By default, it shows traffic used by all users over the last 12 hours.
3. To change the time range, select one from the **Last 12 hours** drop-down list.
4. To view traffic for a specific user, select the account from the **All Users** drop-down list.

## Check billing-cycle traffic usage

Review your monthly traffic usage to see how much data you have consumed in the current billing cycle.

1. From the left navigation pane, select **Usage & billing**.

   ![Olares Space traffic details](/images/one/olares-space-traffic-usage.png#bordered)

2. On the **Usage** tab, locate the **Traffic details** section. By default, the traffic details of the latest billing cycle are displayed.

   - **Progress bar**: Shows how much data you have consumed against your plan's limit. For example, 0.05 GB/2.0 GB.
   - **Daily chart**: A bar chart displaying your data usage day by day, helping you spot sudden increases in activity.

3. To view a previous billing period, select one from the date range drop-down list.

   ![Olares Space traffic by month](/images/one/space-traffic-filter.png#bordered)

4. To view traffic for a specific user, select the account from the **All Users** drop-down list.
