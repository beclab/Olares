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

## Manage app access by mode

### Time slicing

**Time slicing** is the default mode in Olares. Use this mode to let multiple apps share GPU resources.
  ![Time slicing](/images/one/gpu-time-slicing.png#bordered){width=85%}

#### Add an app

In most cases, running apps are automatically bound and appear in the list after GPU scheduling is complete.

If the target app does not appear:
1. Wait a few seconds.
2. In the **Pin application** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the app is still not bound automatically, click **Bind app**, then select the app and click **Confirm**.

#### Remove GPU access

1. Stop the app.
    - Go to **Market** > **My Olares**, then select **Stop** from the dropdown list.
    - Or go to **Settings** > **Applications**, select the app, then click **Stop**.
2. Return to **Settings** > **GPU**.
3. In the **Pin application** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

### Memory slicing

Use **Memory slicing** to run apps concurrently with strict VRAM limits.

![Memory slicing](/images/one/gpu-mem-slicing.png#bordered){width=85%}

#### Adjust VRAM allocation

1. In the **Allocate VRAM** section, find the target app.
2. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
3. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

:::warning
The total VRAM allocated to all apps cannot exceed the GPU's total VRAM.

If the value is lower than the app's minimum requirement, **Confirm** is disabled.
:::

#### Add an app and assign VRAM

In most cases, running apps are automatically bound and appear in the list after GPU scheduling is complete.

If the target app does not appear:
1. Wait a few seconds.
2. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the app is still not bound automatically, click **Bind app**, then select the app and assign VRAM.

#### Remove VRAM allocation

1. Stop the app first.
    - Go to **Market** > **My Olares**, then select **Stop** from the dropdown list.
    - Or go to **Settings** > **Applications**, select the app, then click **Stop**.
2. Return to **Settings** > **GPU**.
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

### App exclusive

Use **App exclusive** mode to dedicate a GPU entirely to one high-demand application.

![App exclusive](/images/one/gpu-app-exclusive.png#bordered){width=85%}

#### Change the exclusive app

1. Stop the current exclusive app.
    - Go to **Market** > **My Olares**, then select **Stop** from the dropdown list.
    - Or go to **Settings** > **Applications**, select the app, then click **Stop**.
2. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i> to unbind the current app.
3. Resume the new target app and make sure it is running.
    - Go to **Market** > **My Olares**, then select **Resume** from the dropdown list.
    - Or go to **Settings** > **Applications**, select the app, then click **Resume**.
4. Wait a few seconds.
5. In the **Select exclusive app** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
6. If the system still does not automatically select the new exclusive app, click **Bind app** to set it manually.

#### Set the exclusive app

In most cases, Olares automatically selects one running app for exclusive access after GPU scheduling is complete.

If no app appears:
1. Wait a few seconds.
2. In the **Select exclusive app** section, click <i class="material-symbols-outlined">sync</i> to refresh the list.
3. If the system still does not automatically select an app, click **Bind app** to set it manually.

#### Remove exclusive access

1. Stop the app first.
    - Go to **Market** > **My Olares**, then select **Stop** from the dropdown list.
    - Or go to **Settings** > **Applications**, select the app, then click **Stop**.
2. Return to **Settings** > **GPU**.
3. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

## Resources
- [Understand GPU management](/zh/manual/olares/settings/gpu-resource.md): Learn how GPU allocation and GPU modes work in Olares.
- [Manage GPU resources for multiple GPUs](/zh/manual/olares/settings/multi-gpu.md): For advanced operations when you have multiple GPUs.