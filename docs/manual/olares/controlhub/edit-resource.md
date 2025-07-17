---
description: Manage Olares system resources efficiently by editing YAML configurations, modifying Pod replicas, and monitoring container status through the Control Hub interface.
---
# Edit resources via Control Hub
This guide shows you how to edit the resource of specific applications in your Olares environment.

:::warning
Modifying system resources can significantly impact system stability and performance. Only proceed with modifications under proper guidance and instructions.
:::
## Edit the YAML file

To edit the YAML file of a workload resource:

1. In Control Hub, navigate to the application's **Deployments** list, and click a resource to expand its details view.
2. In the top right corner, click **<i class="material-symbols-outlined">more_vert</i>** > **Edit YAML** to open the YAML editor.
3. Edit the YAML configuration of the workload as needed.
4. Click **OK** to save your changes and apply them.

   ![edit yaml](/images/how-to/olares/controlhub/browse/10.jpg#bordered)

## Modify Pod replicas

To modify the number of running Pod replicas:

1. In Control Hub, navigate to a Pod's resource details page, and locate the number of Pod replicas at the top.
2. Click **<i class="material-symbols-outlined">add</i>** or **<i class="material-symbols-outlined">remove</i>** to adjust the number of Pod replicas.

   ![replicas](/images/how-to/olares/controlhub/browse/09.jpg#bordered)

:::warning
Many applications in Olares do not support multi-replica mode. Increasing the number of replicas for these Pods may cause exceptions. Therefore, it's important to read the documentation thoroughly and adjust the number of replicas with caution.
:::
