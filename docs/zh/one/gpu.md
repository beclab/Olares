---
outline: [2, 4]
description: Learn about the three GPU modes in Olares and how to switch between them to match different workloads.
head:
  - - meta
    - name: keywords
      content: GPU, Time slicing, App exclusive, Memory slicing, GPU management
---
# Switch GPU mode <Badge text="10 min"/>

:::info
Only Olares admins can change GPU modes. This helps avoid conflicts and keeps GPU performance predictable for everyone.
:::

Olares lets you control how applications use GPU resources for workloads like AI, image and video generation, and gaming. You can choose different GPU modes depending on how your apps should share compute and memory.

Olares also supports NVIDIA DGX Spark, with **Memory slicing** and **App exclusive** available for GPU resource management.

:::tip Multi-GPU setup
If your device has multiple GPUs, see [Manage GPU resources for multiple GPUs](/zh/manual/olares/settings/multi-gpu.md) instead.
:::

## Learning objectives

By the end of this tutorial, you will learn how to:

- Choose the right GPU mode for your workload.
- Switch GPU mode in Settings.
- Use the basic app controls in each GPU mode.

## Choose the right GPU mode

Olares provides three GPU modes, each designed for a different usage pattern.

| GPU mode | Definition | Best for |
|--|--|--|
| **Time slicing** (Default) | Multiple apps share one GPU<br> by taking turns using compute<br> and VRAM.                         | General workloads that run several lightweight apps.                              |
| **Memory slicing**         | The GPU's VRAM is divided into<br> fixed quotas, and apps run concurrently<br> within their limits. | Running specific apps simultaneously while strictly limiting their memory usage.  |
| **App exclusive**          | One app gets full, uninterrupted<br> access to the compute and VRAM<br> of a single GPU.            | Heavy workloads that require maximum stability, such as LLMs and high‑end gaming. |

## Open GPU settings

This page displays your GPU details and allows you to change its mode.

1. Go to **Settings** > **GPU**.
  ![GPU overview](/images/one/gpu-details.png#bordered){width=85%}

2. Choose a mode from the **GPU mode** dropdown.

:::warning App interruption notice
Changing a GPU mode reallocates hardware resources. Depending on the mode you choose, apps that are currently using the GPU may be paused automatically.

After switching modes, check the state of your apps and manually resume them if needed.
:::

:::info GPU scheduling delay
After you switch GPU modes or resume an app, the app-to-GPU assignment may not appear in the UI immediately.

If an app does not appear in the list, wait a few seconds, then click <i class="material-symbols-outlined">sync</i> to refresh the list. In most cases, Olares updates the assignment automatically after GPU scheduling is complete.

The **Bind app** button appears only when there are apps waiting to be bound.
:::

## Manage app access by mode

### Time slicing

**Time slicing** is the default mode in Olares. Use this mode to let multiple apps share GPU resources.
  ![Time slicing](/images/one/gpu-time-slicing.png#bordered){width=85%}

#### Add an app

1. Wait a few seconds after GPU scheduling is complete.
2. In the **Pin application** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the app still does not appear automatically, click **Bind app**, then select the app and click **Confirm**.

#### Remove GPU access

1. Stop the app.
2. In the **Pin application** section, click <i class="material-symbols-outlined">link_off</i>.
3. Click **Confirm**.

### Memory slicing <Badge type="tip" text="DGX Spark supported" />

Use **Memory slicing** to run apps concurrently with strict VRAM limits.

![Memory slicing](/images/one/gpu-mem-slicing.png#bordered){width=85%}

#### Add an app and assign VRAM

1. Wait a few seconds after GPU scheduling is complete.
2. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the app does not appear automatically, click **Bind app**, then select the app, assign VRAM, and click **Confirm**.

#### Adjust VRAM allocation

1. In the **Allocate VRAM** section, find the app.
2. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
3. Enter the new VRAM value.
4. Click **Confirm**.

:::warning
The total of all VRAM allocations must not exceed the GPU's total VRAM.
:::

#### Remove VRAM allocation

1. Stop the app.
2. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>.
3. Click **Confirm**.

### App exclusive <Badge type="tip" text="DGX Spark supported" />

Use **App exclusive** mode to dedicate a GPU entirely to one high-demand application.

![App exclusive](/images/one/gpu-app-exclusive.png#bordered){width=85%}

#### Change the exclusive app

1. Stop the current exclusive app.
2. Resume the new target app and make sure it is running.
3. Wait a few seconds after GPU scheduling is complete.
4. In the **Select exclusive app** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
5. If the system does not automatically select the new app, click **Bind app**, then select the app and click **Confirm**.

#### Set the exclusive app

1. Wait a few seconds after GPU scheduling is complete.
2. In the **Select exclusive app** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If no app is selected automatically, click **Bind app**, then select the target app and click **Confirm**.

#### Remove exclusive access

1. Stop the app.
2. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>.
3. Click **Confirm**.

## Resources
- [Understand GPU management](/zh/manual/olares/settings/gpu-resource.md): Learn how GPU allocation and GPU modes work in Olares.
- [Manage GPU resources for multiple GPUs](/zh/manual/olares/settings/multi-gpu.md): For advanced operations when you have multiple GPUs.