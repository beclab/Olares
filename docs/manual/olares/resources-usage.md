---
outline: [2, 3]
description: Monitor your Olares system and application performance with comprehensive dashboards tracking CPU, memory, and disk usage for optimal resource management.
---

# Monitor system and application status
The Dashboard app in Olares provides a centralized and intuitive view of your system's status, offering deep insights without requiring extensive technical expertise. From the main dashboard, you can now monitor key resource usage patterns and access detailed real-time metrics across your cluster.

## Access monitoring dashboards

View your system's status through these specialized dashboards:
- **Overview**: Shows current resource usage and system health.
- **Applications**: Displays running applications and their status.

## Overview

The Overview page provides a comprehensive, at-a-glance view of your Olares system's health and resource utilization. It is divided into key sections to help you monitor performance effectively:

- Cluster's physical resources
- User resource usage
- Usage ranking

### View physical resources

Monitor the fundamental resource metrics directly from the **Overview** dashboard:

- CPU utilization
- Memory consumption
- Disk usage
- Pod status
- GPU usage
- Network status

![Dashboard overview](/images/manual/olares/dashboard-overview.png#bordered)

### Access detailed resource panels

For a deeper analysis of any resource, simply click on its card in the **Overview** dashboard. This navigates you to a dedicated detail panel with comprehensive monitoring data and metrics.

#### CPU panel

The CPU details panel provides an in-depth view of your Olares cluster's CPU performance and health. To access it, click on the **CPU core** card from the Overview dashboard.

This panel displays:
- A real-time CPU utilization graph.
- Node-specific CPU specifications (model, cores, threads).
- Breakdown of utilization rate (User, System, I/O wait).
- Current CPU temperature.
- Average load over 1, 5, and 15 minutes.

#### Memory panel

The Memory details panel offers a clear, in-depth view of your Olares cluster's memory usage and allocation. To access it, click on the **Memory** card from the **Overview** dashboard.

This panel allows you to switch between **Physical memory** and **Swap** views using the dropdown menu.

- When **Physical memory** is selected, it displays:
  - A real-time memory utilization graph.
  - A memory usage breakdown showing Reserved, Used, Buffer, Cache, and Available memory, along with total memory and utilization rate.

- When **Swap** is selected, it displays:

  - A real-time swap usage graph.
  - Numerical swap in/out rates.
  - Swap space summary (Total, Used, and Utilization rate).

#### Disk panel 

The Disk panel offers a comprehensive view of your storage devices. Use this panel to monitor disk health, track storage consumption, and analyze space allocation in your Olares cluster. To access it, click on the Disk card from the **Overview** dashboard.

This panel displays:

- Overall storage status: Disk name, storage status, and a usage bar showing used, and available space.

- Detailed information: Key device specifications like Total capacity, model, serial number, interface protocol, temperature, power-on time, and write volume.

For a detailed usage breakdown for a specific storage device, click **Occupancy analysis** in the top right corner. 

The storage usage popup displays:

- A list of file systems (partitions) on the disk.
- For each file system, you can view storage metrics like total capacity, used space, available space, usage rate, and mount point.


Here's the summarized documentation for the Pods details panel:

#### Pods details panel
The Pods panel offers a dynamic view of your application's deployment status  To access it, click the Pods card on the **Overview** dashboard.

This panel displays:

Pod Count Graphs: Real-time graphs showing the number of running pods over time for different nodes within the cluster. (e.g., "Olares" and "Olares2").

#### GPU panel

The **GPU** panel provides in-depth information about your GPUs within the cluster. Use this panel to effectively monitor GPU health, resource allocation, and performance across your cluster. 

To access the GPU panel, click the GPU card on the **Overview** dashboard.

![GPU details list](/images/0329bd.png#bordered)

This panel includes two tabs: 

- **Graphics management tab**: view a list of all detected GPUs, including their GPU ID, model, GPU mode (e.g., Memory Slicing), host node, health status, computing power usage, VRAM usage rate, and power draw.
  
  For a more granular view of a specific GPU, click **View details** next to its entry

- **Task management tab**: Monitor tasks currently utilizing your GPUs. It provides insights into task name, status, GPU Mode, host node, computing power usage, and VRAM used, along with available operations.

#### Network panel

The Network panel provides comprehensive insights into your network interfaces. Use this panel to monitor network connectivity, traffic flow, and configuration for optimal system performance.

The Network panel displays:

- Network port information: Details for each network port (e.g., wlo1), including its usage status, real-time upload and download speeds, and connection status.

- IP configuration: Information on IP acquisition method (e.g., DHCP), owned node, and network configuration.

- IPv4 and IPv6 details: Comprehensive details for both IPv4 and IPv6, including address, subnet mask, gateway address, DNS, and network status.

![Network details](/images/0328fd.png#bordered)


### Access physical resources

For deeper analysis, click **More details** to view comprehensive monitoring data for the past 7 days.

Use the dropdown menu in the top right to change the time range, or click <i class="material-symbols-outlined">refresh</i> to update monitoring data.

The following metrics help you maintain optimal system performance:

| Metric           | Description                        | Impact                                        |
|------------------|------------------------------------|-----------------------------------------------|
| CPU usage        | Percentage of CPU resources used   | Prolonged spikes can slow down the system     |
| Memory usage     | Percentage of memory in use        | Impacts application performance and stability |
| Average CPU load | Average number of active processes | High load indicates system overload           |
| Disk usage       | Percentage of disk space used      | Crucial for data reliability, prevent overuse |
| Inode usage      | Percentage of inodes used          | Exhaustion prevents new file creation         |
| Disk throughput  | Data transfer rate (MB/s)          | Important for large file transfers            |
| IOPS             | Input/Output Operations Per Second | Critical for small file or random data access |
| Network traffic  | Network usage (Mbps)               | Reflects network speed and quality            |
| Pod status       | Count of pods by state             | Reflects application health                   |

![Physical resource monitoring](/images/manual/olares/physical-resource-monitoring.png#bordered)
### Check resource quota
You can view your resource quota allocated by the Olares admin.

![Resource quota](/images/manual/olares/resource-quota.png#bordered)

:::warning
When your resource quota runs low, you may experience:

* Slower system performance.
* Inability to install new applications.
* Automatic suspension of resource-intensive applications.
:::


### Track application performance
The **Usage ranking** section displays the top 5 applications consuming CPU and memory resources. To access the complete list of application resource usage, click **More**.

![Usage ranking](/images/manual/olares/usage-ranking.png#bordered)

## Applications

The **Applications** dashboard helps you monitor resource usage patterns across your applications through various sorting and filtering options.

Use the dropdown menu in the upper right corner to sort applications based on their resource consumption:
- CPU usage
- Memory usage
- Inbound traffic
- Outbound traffic

![Applications](/images/manual/olares/applications.png#bordered)

Toggle between ascending and descending order to identify which applications are consuming the most or least resources.

For applications supporting multiple entrances (such as Wordpress), you can click icons to switch between different entrance types and view their corresponding resource metrics.
![Multiple entrances](/images/manual/olares/multiple-entrances.png){width=40%}
:::tip
* When your application list grows large, quickly locate specific applications by typing their names in the search box at the top of the page.
* Regularly checking resource consumption patterns helps you identify applications that might need optimization or attention.
:::