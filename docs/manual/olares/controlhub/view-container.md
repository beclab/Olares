---
description: Troubleshooting application issues status by examining the container staus or exporting logs 
---
# Examine container status

The Pods page provides a comprehensive view of all Pods in your Olares environment, allowing you to manage them at the smallest granularity offered by Kubernetes.

This guide shows you how to check the status and export logs of your containers. 

## View container status

Click on a Pod in the list takes you to the Pod details page, where you can:
- View container logs.
- Access the container environment.
- View container ports and environment variables.
- Open the Pod's YAML configuration in a read-only view.
  :::tip
  You cannot edit the YAML configuration directly from this view. The YAML is managed by Olares through workload templates and webhooks.
  :::
  ![pod detail](/images/how-to/olares/controlhub/pods/02.jpg#bordered)

## Export container logs for troubleshooting

To effectively diagnose and resolve issues, you may need to examine detailed logs from your containers.

![pod detail](/images/manual/olares/controlhub-export-log.png)

1. In the Browse column, navigate to your application, then go to **Deployments** > **Containers**.
2. Locate the container that's experiencing issues (with an orange dot).
3. Click the <i class="material-symbols-outlined">article</i> button next to the container.
4. In the pop-up log window, you have the following options to manage the logs:
   ![Log operations](/images/manual/olares/controlhub-log.png)
   - Click the <i class="material-symbols-outlined">download_2</i> button to download the entire log file.
   - Click the <i class="material-symbols-outlined">autorenew</i> button to refresh and see the latest log entries.
   - Click the <i class="material-symbols-outlined">play_pause</i> button to start or pause the log updating in real-time.
