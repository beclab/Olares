---
outline: [2, 3]
description: Monitor your Olares system status and traffic usage in Olares Space.
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, monitor Olares, system status, resource usage, traffic usage
---
# Monitor Olares status and traffic in Olares Space

This page covers how to monitor your Olares system status and traffic usage in Olares Space.

## Before you begin

Before you can monitor your Olares in Olares Space, you must authorize Olares Space to access your system data. To do this, link your Olares Space account to the Olares device in LarePass:

1. In the LarePass app on your mobile device, go to **Settings** > **Integrations**.
2. Tap <i class="material-symbols-outlined">add</i> in the upper-right corner, and then select **Olares Space**.

## Monitor resource usage

Check CPU, memory, and disk usage to make sure your Olares has enough resources.

1. On the **Olares** page, select the **Overview** tab.

   ![Olares page, overview tab](/images/how-to/space/olares_page_overview.png#bordered)

2. Locate the **Resource Monitor** section. It displays the real-time usage of CPU, memory, and disk.

   | Metric | Description |
   | ------ | ----------- |
   | **CPU (Cores)** | Current CPU usage in cores and the total available cores. |
   | **Memory (GB)** | Current memory usage in GB and the total available memory. |
   | **Disk (GB)** | Current disk usage in GB and the total available disk space. |

## View active hosts

Check which hosts are currently running in your Olares cluster and their status.

1. On the **Olares** page, select the **Overview** tab.
2. Locate the **Active hosts** section. It displays the hosts currently running in the Olares cluster.

## Check traffic usage

Check recent traffic usage to spot sudden increases and avoid exceeding your plan's limit.

:::info
For self-hosted Olares users, monitor traffic statistics for the reverse proxy service. If you exceed the monthly quota, speed is throttled to 5 Mbps. The free alternative is to use [LarePass VPN](../larepass/private-network.md) or host your own FRP server.
:::

1. On the **Olares** page, select the **Usage statistics** tab.

   ![Olares Space traffic usage](/images/how-to/space/olares_usage_statistics1.png#bordered)

2. Locate the **Traffic Usage** section. By default, it shows traffic used by all users over the last 12 hours.
3. To change the time range, select one from the **Last 12 hours** drop-down list.
4. To view traffic for a specific user, select the account from the **All Users** drop-down list.

## View billing-cycle traffic usage

Review your monthly traffic usage to see how much data you have consumed in the current billing cycle.

1. From the left navigation pane, select **Usage & billing**.

   ![Olares Space traffic details](/images/one/olares-space-traffic-usage.png#bordered)

2. On the **Usage** tab, locate the **Traffic details** section. By default, the traffic details of the latest billing cycle are displayed.

   - **Progress bar**: Shows how much data you have consumed against your plan's limit. For example, 0.05 GB/2.0 GB.
   - **Daily chart**: A bar chart displaying your data usage day by day, helping you spot sudden increases in activity.

3. To view a previous billing period, select one from the date range drop-down list.

   ![Olares Space traffic by month](/images/one/space-traffic-filter.png#bordered)

4. To view traffic for a specific user, select the account from the **All Users** drop-down list.
