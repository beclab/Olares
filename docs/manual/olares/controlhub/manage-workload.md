---
outline: [2, 3]
description: Learn how to use Control Hub to manage Olares workloads and configurations. This doc covers project and namespace concepts, managing different workload types, and monitoring resource usage through detailed dashboards.
---

# Manage workloads 

Workloads are resources used to manage pod replicas, responsible for overseeing multiple containerized application instances. This section guides you on how to manage pods under different workloads (applications or services) within Control Hub.

:::tip Note
Olares members can only access their own namespaces, while Olares administrators can access all user and system namespaces.
:::

## Projects and namespaces

In Olares' Browse view, **Projects** and **Namespaces** combine to provide a two-level organizational structure for resources:

![Org](/images/manual/olares/controlhub-org.jpeg#bordered)

* **Project**: Categorize namespaces by "user" or "system" to quickly locate the resources belonging to a specific user or system module.
* **Namespace**: A project is made up of multiple namespaces, which are native Kubernetes isolation units used to distinguish between different applications, components, or environments. 

| Category            | Namespace Prefix | Description                                                                                                               |
|---------------------|---|---------------------------------------------------------------------------------------------------------------------------|
| **User projects**   | `app-<olares-id>` | Community applications installed from the Market, such as `steamheadless-chenglin106`, `ollama-chenglin106`, etc.         |
|                     | `user-space-<olares-id>` | Built-in system applications such as Files, Market, Control Hub, Dashboard, Vault, etc.                               |
|                     | `user-system-<olares-id>` | User-related system services such as runtime components, schedulers, cross-application proxies, etc.                      |
| **System projects** | `System` | Cluster-level dependencies such as Kubernetes core, KubeSphere, Olares platform components, and necessary hardware drivers. |

## Manage workloads

In Olares, a **workload** represents an application or service running in the Olares cluster, responsible for managing one or more pods of an application. All workloads are deployed within a specific namespace. There are three workload types:

| Type | Typical Scenarios                                                                           | Characteristics                                                                                  |
|---|---------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| **Deployment** | Manages stateless applications.                                                             | Most common; automatically creates ReplicaSets and pods; supports horizontal scaling.            |
| **StatefulSet** | Manages stateful applications like databases,<br/> distributed storage, and cache clusters. | Ensures pods start/terminate in order; provides a fixed network identity and persistent storage. |
| **DaemonSet** | Workloads for node-level monitoring, <br/>log collection, GPU drivers, etc.                      | Runs one pod per node, which is automatically added or deleted as nodes join or leave.           |

### View workloads

To view a specific workload:

1.  In the left navigation, click **Browse**, 
2.  In the first column, click the target namespace to expand it.
2.  In the second column, click to expand the **Deployment/StatefulSet/DaemonSet** list, and then select the desired workload item.
3.  View the workload details un the third column:
    ![Workload](/images/manual/olares/controlhub-workload.png#bordered)

    | Section                   | Description                                                                                                            |
    |---------------------------|------------------------------------------------------------------------------------------------------------------------------------------|
    | **Basic information**     | Metadata such as cluster, project, creation/update time, <br/>and creator.                                                               |
    | **Pods**                  | Lists associated pods with info such as node, IP, and<br/> real-time CPU/memory charts. Click on the pods name <br/>to view the details. |
    | **Ports**                 | The container ports, protocols, and listener port numbers <br/>exposed by the pods.                                                      |
    | **Environment Variables** | A list of environment variables defined in the pod template.                                                                             |
    | **Labels**                | Configured in workload metadata for resource scheduling <br/>and filtering.                                                              |
    | **Annotations**           | Defined in `metadata.annotations`, which functions similarly <br/>to labels, allowing for flexible management by controllers.            |
    | **Events**                | Events related to the workload within the last hour, such as <br/>scheduling, restarts, or image pulls.                                  |

### Edit YAML configuration

In some advanced maintenance scenarios, you may need to directly adjust the number of pod replicas, add environment variables, modify probes, or update storage volume claims. In such cases, you can use the **Edit YAML** function for fine-grained configurations.

:::warning Proceed with caution
Directly modifying system resource YAML can affect cluster stability and performance. Operate with extreme caution and ensure you have a backup or are guided by a professional.
:::

1.  On the workload details page, click the <i class="material-symbols-outlined">edit_square</i> button in the top-left corner.
2.  Modify the configuration in the pop-up YAML editor.
3.  Click **OK** to apply the new configuration immediately.

![Edit YAML](/images/how-to/olares/controlhub/browse/10.jpg#bordered)

### Stop or restart workloads

When you need to quickly troubleshoot an application or service, free up system resources, or reload a configuration, you can use the **Stop/Start** and **Restart** functions.

| Function       | Description                                                                                                                                                                        |
|----------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Stop/Start** | Reduces the number of replicas to 0 and terminate the <br/>pods. To resume,  click the **Start** button or <br/>[adjust the number of replicas in YAML](#edit-yaml-configuration). |
| **Restart**    | Stops and then immediately starts new pods based on the<br/> original replica count, used for quickly refreshing<br/> configurations or resolving transient failures.              |

1.  On the workload details page, click the **Stop** or **Restart** button in the top-right corner.
2.  In the confirmation dialog, type the pod name as prompted and click **Confirm**.
3.  Observe the pod termination/start progress in the pods list to confirm that the status has returned to normal.

## Monitor resource usage

You can monitor the cluster resource usage through two views:

| View           | Use case                                                                                                   | Description                                                                                                                                  |
|----------------|------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| **Namespaces** | When you need to horizontally  <br/>compare source usage across<br/> multiple applications or system<br/> modules. | Summarizes key metrics like CPU, memory,<br/> and traffic for a namespace to quickly locate resource<br/> hotspots and evaluate quota usage. |
| **Pods**       | When you need to drill down into<br/> a single pod for troubleshooting or <br/>performance analysis.            | Provides real-time status, resource charts, and event logs <br/>for each pod and its containers, enabling fine-grained maintenance.          |

You can use the **Namespaces view** to first identify "who" is consuming too many resources, and then use the **Pods view** to pinpoint "which specific pod/container" is causing the issue. This helps you monitor and maintain the cluster from a global to a detailed level.

### Namespaces view

The Namespaces view aggregates key metrics like CPU, memory, and traffic by namespace, allowing you to quickly discover resource hotspots and compare quota usage for efficient performance tuning and troubleshooting.

![Namespace view](/images/manual/olares/controlhub-namespace.png#bordered)

1.  In the left navigation bar, click **Namespaces**.
2.  Use the dropdown menu at the top to switch between different users' namespaces, or use the search box for a precise search by keyword.
    
    | Column               | Description                                  |
    |----------------------|----------------------------------------------|
    | **Namespace**        | The name of the namespace.                   |
    | **CPU usage**        | Current CPU usage (Supports sorting).        |
    | **Memory usage**     | Current memory usage (Supports sorting).     |
    | **Pods**             | The number of pods running in the namespace. |
    | **Outbound traffic** | The outbound traffic rate.                   |
    | **Inbound traffic**  | The inbound traffic rate.                    |

Clicking on any namespace will take you to its resource details page:

| Section       | Description                                                                                                                    |
|---------------|--------------------------------------------------------------------------------------------------------------------------------|
| **Quota card** | The used percentage of CPU/memory quotas.                                                                                    |
| **Pods list** | List of Pods under the namespace; <br/>Supports searching by Pods name and sorting by <br/>CPU/memory/traffic.             |
| **Pod**   | View image, node, ports, and real-time resource <br/>charts of the Pods. Click the Pod name to view the <br/>container charts. |

### Pods view

The **Pods** view provides an aggregated display of statues of all pods and resource usage within the cluster.

1.  In the left navigation bar, click **Pods**.
2.  Use the dropdown menu at the top to switch between nodes, Pods status, or use the search box for a precise search by keyword.

    ![Pods view](/images/manual/olares/controlhub-pods-list.png#bordered)
    
    | Column            | Description                                                        |
    |-------------------|--------------------------------------------------------------------|
    | **Name**          | The full name of the pod.                                          |
    | **Status**        | Running/Completed/Abnormal/Error.                            |
    | **Node**          | The node where the pod is located, and its private IP.             |
    | **Pod IP**        | The IP address of the pod.                                         |
    | **Creation time** | Pod creation time. Can be sorted in ascending or descending order. |

Click on a specific pod name to drill into the [Pod details](manage-container.md#view-pod-details).

### Resource trends

Control Hub provides two distinct views to monitor real-time and historical resource trends, helping you quickly assess resource usage from different perspectives.

- **View by user**: Under the **Namespaces** > **Resources** tab, you can view the overall CPU, memory, and pod dynamics for a specific user. This allows you to quickly assess an individual user's resource footprint over a given period.

- **View by cluster**: Under the **Pods** > **Resources** tab, you can view the overall CPU, memory, storage, and pod dynamics from a cluster-wide perspective. This provides a clear, high-level overview of your cluster's resource trends and health.