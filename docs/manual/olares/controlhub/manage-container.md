---
description: Learn how to manage containers and troubleshoot application issues in Olares' Control Hub. This guide covers how to view pod details, check container status, and export container logs for diagnostics. 
---

# Manage containers

When an application encounters issues such as CrashLoopBack, slow startup, or sudden resource spikes, container-level inspection is often the most direct way to troubleshoot. The pod detail page in Control Hub aggregates information of all containers within a specific pod, providing events, visual charts, and logs to help you quickly identify and resolve problems.


## View Pod details

To view the containers in a Pod:

1. In the left navigation bar, click **Browse**, and select the target namespace from the first column.
2. In the second column, expand **Deployment**, **StatefulSet**, or **DaemonSet**, then select the desired workload and drill down to its **Pods** list.
3. Click on the target pod.
4. In the third column, view the detailed pod information:
   ![containers](/images/manual/olares/controlhub-pods.png#bordered)


   | Section       | Content                                                                                                                   |
   |---------------|---------------------------------------------------------------------------------------------------------------------------|
   | **Info**      | Basic data such as pod status, restart count, <br/>IP, node, QoS, and images                                              |
   | **Containers**| Real-time CPU/memory usage charts, image versions, status<br/> with options to export logs and access the container shell |
   | **Volumes**   | Mounted persistent volumes, paths, and access modes                                                                       |
   | **Environment Variables** | All injected variables with expand-to-view support                                                                        |
   | **Events**    | Recent one-hour event logs for scheduling, probes, networking, etc.                                                       |

   :::tip
   You cannot directly edit pod YAML from this view. YAML configurations are managed by Olares via workload templates and webhooks.
   :::

## View container status

For each container listed within a pod, you can:

- Click the entry to view detailed container information, including status, image info, image pull policy, ports, and container metadata.
- View container ports and environment variables.
- Access the container’s command-line shell.
- View or export container logs in real time.

![container detail](/images/how-to/olares/controlhub/pods/02.jpg#bordered)

## Export container logs for troubleshooting

Exporting container logs, along with **Events** and resource charts, helps quickly identify issues such as CrashLoopBack, probe failures, or OOM (Out of Memory) errors.

![Log operation](/images/manual/olares/controlhub-export-log.png)

1. In the left navigation bar, click **Browse**, and select the namespace for the problem application.
2. In the second column, click to expand **Deployments** > **Target deployment** > **Pods**.
3. In the Pods detail page, locate the abnormal container (typically marked with an orange status indicator), and click the <i class="material-symbols-outlined">article</i> button.
4. In the pop-up log window, you can:
   ![Log operation](/images/manual/olares/controlhub-log.png)
    - Click <i class="material-symbols-outlined">download_2</i> to download the complete log file.
    - Click <i class="material-symbols-outlined">autorenew</i> to refresh and view the latest log entries.
    - Click <i class="material-symbols-outlined">play_pause</i> to start or pause real-time log updates.  
