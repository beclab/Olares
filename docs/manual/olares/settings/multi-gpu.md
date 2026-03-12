---
outline: deep
description: Configure GPU modes, reassign apps between GPUs, and manage GPU access on Olares with multiple GPUs.
---
# Manage GPU resources for multiple GPUs

This guide explains how to manage GPU modes and app allocation when Olares has multiple GPUs. Each GPU is configured separately.

## Open GPU settings

This page lists the available GPUs, including each GPU's model, node, total VRAM, and current mode. You can also select a GPU and change its mode.

1. Go to **Settings** > **GPU**.
   ![GPU overview multiple-GPU](/images/manual/olares/gpu-overview.png#bordered){width=90%}

2. Click the GPU you want to configure.
3. Choose a mode from the **GPU mode** dropdown.

:::tip Can't find the target app?
If the app still does not appear on the current GPU after you refresh the list, it may already be assigned to another GPU on the same node or on another node.

Check other GPUs to see where the app is currently assigned. If needed, use **Switch GPU** to reassign it to the target GPU.
:::

## Time slicing

:::info
Time slicing is currently not supported on DGX Spark.
:::

![Time slicing multi GPU](/images/manual/olares/gpu-time-slicing-multi.png#bordered){width=90%}

### Add an app

In most cases, apps are automatically assigned to an available GPU and appear in the list after GPU scheduling is complete.

If the target app does not appear on the current GPU:

1. Wait a few seconds.
2. In the **Pin application** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the app still does not appear on the current GPU, check other GPUs to see whether it has been assigned elsewhere.
4. If the system still does not complete the assignment automatically, click **Bind app** to add it manually.

### Reassign an app to another GPU

1. In the **Pin application** section, find the app.
2. Click <i class="material-symbols-outlined">repeat</i>.
3. Select the target GPU and click **Confirm**.

The app continues running during this process.

### Remove an app's access to this GPU

1. If the app is assigned only to this GPU, stop it first.
    - Go to **Market** > **My Olares**, then select **Stop** from the dropdown list.
    - Or go to **Settings** > **Applications**, select the app, then click **Stop**.
2. Return to **Settings** > **GPU**.
3. In the **Pin application** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

:::tip Unbind from multiple GPUs
If the app is still assigned to other GPUs on the same node, you can remove it from the current GPU without stopping it.
:::

## Memory slicing

![Multi-GPU memory slicing](/images/manual/olares/gpu-mem-slicing-multi.png#bordered){width=90%}

### Adjust VRAM allocation

1. In the **Allocate VRAM** section, find the target app.
2. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
3. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB, then click **Confirm**.

:::warning
The total VRAM allocated to all apps on the current GPU cannot exceed that GPU's total VRAM.

If the value is lower than the app's minimum requirement, **Confirm** is disabled.
:::

### Add an app

In most cases, apps are automatically assigned to an available GPU and appear in the list after GPU scheduling is complete.

If the target app does not appear on the current GPU:

1. Wait a few seconds.
2. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the app still does not appear on the current GPU, check other GPUs to see whether it has been assigned elsewhere.
4. If the system still does not complete the assignment automatically, click **Bind app** to add it manually.

### Reassign an app to another GPU

1. In the **Allocate VRAM** section, find the app and click <i class="material-symbols-outlined">repeat</i>.
2. Choose the target GPU.
3. Click **Confirm**.

### Remove an app's VRAM allocation from this GPU

1. If the app is assigned only to this GPU, stop it first.
    - Go to **Market** > **My Olares**, then select **Stop** from the dropdown list.
    - Or go to **Settings** > **Applications**, select the app, then click **Stop**.
2. Return to **Settings** > **GPU**.
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

:::tip
If the app also has VRAM allocated on other GPUs on the same node, you can remove the VRAM allocation from the current GPU without stopping it.
:::

## App exclusive

![Multi-GPU app exclusive](/images/manual/olares/gpu-app-exclusive-multi.png#bordered){width=90%}

### Reassign the exclusive app to another GPU

1. In the **Select exclusive app** section, click <i class="material-symbols-outlined">repeat</i>.
2. Choose the target GPU.
3. Click **Confirm**.

### Set the exclusive app

In most cases, Olares automatically selects an available running app for exclusive access after GPU scheduling is complete.

If no app appears on the current GPU:

1. Wait a few seconds.
2. In the **Select exclusive app** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the app still does not appear on the current GPU, check other GPUs to see whether it has been assigned elsewhere.
4. If the system still does not complete the selection automatically, click **Bind app** to set it manually.

### Remove an app's exclusive access to this GPU

1. If the app is assigned only to this GPU, stop it first.
    - Go to **Market** > **My Olares**, then select **Stop** from the dropdown list.
    - Or go to **Settings** > **Applications**, select the app, then click **Stop**.
2. Return to **Settings** > **GPU**.
3. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

:::tip
If the app is also assigned to other GPUs on the same node, you can remove it from the current GPU without stopping it.
:::

## Learn more

- [Understand GPU management](./gpu-resource.md)
- [Monitor GPU usage in Olares](../resources-usage.md)
- [Manage GPU resources with one GPU](./single-gpu.md)