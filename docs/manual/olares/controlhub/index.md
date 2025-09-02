---
outline: [2, 3]
description: Understand Control Hub UI to manage applications, monitor resources, configure networking, and maintain system settings in Olares.
---

# Manage Olares with Control Hub

**Control Hub** is Olares' visual console, providing developers and operators with precise control over the cluster and its underlying Kubernetes environment. With multi-dimensional views, you can quickly browse application workloads, view resource status, and perform maintenance operations on key objects.

Use the modular entries in the side navigation bar to access the appropriate management page for your needs.

![Control Hub UI](/images/manual/olares/controlhub-ui.jpeg#bordered)

## Module overview

Control Hub modules give you clear entry points to monitor workloads, manage resources, and access the terminal for advanced maintenance operations in Olares.

### Olares

* **Browse**: A three-column workload resource navigator. Use the left column to select the namespace for applications and services, the center to expand workloads and configurations, and the right to view real-time details and monitoring charts.
* **Namespaces**: Lists all resources by Kubernetes namespace. You can sort and filter by metrics like CPU and memory to easily troubleshoot resource hotspots.
* **Pods**: Monitors the number of pods, their status, node distribution, and resource consumption in real-time through a dual-view of a pod list and resource charts.

### Resources

* **Storage**: View persistent volume claims (PVCs), PVs, and StorageClass usage and capacity trends.
* **Network**: Monitor network security policies implemented in the system and the network connectivity of each namespace.
* **Jobs**: Manage the execution status and logs of jobs and cron jobs.
* **CRDs**: Browse and manage custom resources (CRD) and their instances.

### Middleware

Quickly view a list of integrated database and caching components (such as **Postgres**, **MongoDB**, **Redis**), including their instance information, connection details, and runtime metrics.

### Terminal

Gain one-click access to the Olares host terminal to perform debugging, view logs, and modify configurations. 
