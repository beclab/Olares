---
outline: deep
description: Configure GPU modes and manage app access when Olares has one GPU.
---
# Manage GPU resources for a single GPU

This guide explains how to manage GPU modes and app access when Olares has one GPU.

Before you start, read [Understand GPU management](./gpu-resource.md) for an overview of how GPU management works in Olares.

:::info GPU scheduling delay
After you switch GPU modes or resume an app, the app-to-GPU binding status may not appear in the UI immediately.

If a running app does not appear in a section, wait a few seconds, then click <i class="material-symbols-outlined">sync</i> in that section to refresh the list. In most cases, Olares binds the app automatically after GPU scheduling is complete.

The **Bind app** button appears only when there are apps waiting to be bound.
:::

## Open GPU settings

This page shows GPU details and lets you change the GPU mode.

1. Go to **Settings** > **GPU**. 
2. Choose a mode from the **GPU mode** dropdown.

## Time slicing

![Time slicing](/images/manual/olares/gpu-time-slicing-single.png#bordered){width=90%}

### Add an app

In most cases, running apps are automatically bound and appear in the list after GPU scheduling is complete.

If the target app does not appear:

1. Wait a few seconds.
2. In the **Pin application** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the app is still not bound automatically, click **Bind app** to add it manually.

### Remove an app's GPU access

1. Stop the app first.
2. Return to **Settings** > **GPU**.
3. In the **Pin application** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

## Memory slicing <Badge type="tip" text="DGX Spark supported" />

![Memory slicing](/images/manual/olares/gpu-mem-slicing-single.png#bordered){width=90%}

### Adjust VRAM allocation

1. In the **Allocate VRAM** section, find the target app.
2. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
3. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

:::warning
The total VRAM allocated to all apps cannot exceed the GPU's total VRAM.

If the value is lower than the app's minimum requirement, **Confirm** is disabled.
:::

### Add an app and assign VRAM

In most cases, running apps are automatically bound and appear in the list after GPU scheduling is complete.

If the target app does not appear:
1. Wait a few seconds.
2. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the app is still not bound automatically, click **Bind app**, then select the app and assign VRAM.

### Remove an app's VRAM allocation

1. Stop the app first.
2. Return to **Settings** > **GPU**.
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

## App exclusive <Badge type="tip" text="DGX Spark supported" />

![App exclusive](/images/manual/olares/gpu-app-exclusive-single.png#bordered){width=90%}

### Change the exclusive app

1. Stop the current exclusive app.
2. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i> to unbind the current app.
3. Resume the new target app and make sure it is running.
4. Wait a few seconds.
5. In the **Select exclusive app** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
6. If the system still does not automatically select the new exclusive app, click **Bind app** to set it manually.

### Set the exclusive app

In most cases, Olares automatically selects one running app for exclusive access after GPU scheduling is complete.

If no app appears:
1. Wait a few seconds.
2. In the **Select exclusive app** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the system still does not automatically select an app, click **Bind app** to set it manually.


### Remove exclusive access from an app

1. Stop the app first.
2. Return to **Settings** > **GPU**.
3. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

## Learn more

- [Understand GPU management](./gpu-resource.md)
- [Monitor GPU usage in Olares](../resources-usage.md)
- [Manage GPU resources for multiple GPUs](./multi-gpu.md)