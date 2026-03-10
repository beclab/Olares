---
outline: [2, 3]
description: Manage and optimize GPU resources in Olares with centralized controls, supporting time-slicing, memory-slicing, and exclusive access across single or multi-GPU setups.
---
# Manage GPU usage

:::info
Only Olares admins can change GPU modes. This helps avoid conflicts and keeps GPU performance predictable for everyone.
:::

Olares lets you manage how apps use available GPUs for workloads such as AI, image and video generation, transcoding, and gaming. 

In this guide, you will learn how to:
- Choose a GPU mode based on how resources are shared.
- Change a GPU mode and manage how apps use GPU resources.

## Before you begin

Before changing GPU settings, it helps to understand how Olares manages GPU allocation.

- **Bind an app** means allocating GPU resources to a running app.
- **Unbind an app** means removing that allocation so GPU resources can be released or reassigned.

### App state and GPU allocation

Whether GPU resources can be allocated to or revoked from an app depends on its current state:

- To use GPU resources, the app must be running. If the app is stopped, resume it first.
- To fully revoke an app's GPU allocation, you may need to stop the app first.

You can change an app's state in either of these places:

- Go to **Market** > **My Olares**, then choose **Resume** or **Stop** from the dropdown list.
- Go to **Settings** > **Applications**, select the app, then click **Resume** or **Stop**.

:::info
Stopping an app pauses its workload, but it does not automatically clear its GPU allocation.

To fully release the GPU or VRAM for other workloads, you must unbind the app after stopping it.
:::

## Choose a GPU mode

Olares supports three GPU modes. Each mode determines how GPU resources are shared and what happens to running apps after you switch modes.

| GPU mode | How resources are<br> shared | After switching to this mode | Best for |
| -- | -- | -- | -- |
| **Time slicing** (Default) | Multiple apps share <br>the same GPU over <br>time. | Running apps that require a GPU are automatically assigned to share the GPU. | Running several GPU-dependent apps at the same time. |
| **Memory slicing** | Multiple apps share <br>the GPU, with fixed <br>VRAM allocations <br>for each app. | Running apps that require a GPU are automatically added and assigned the minimum VRAM required to run. | Running multiple GPU-dependent apps while strictly controlling VRAM usage. |
| **App exclusive** | One app gets full, <br>uninterrupted access<br> to the GPU. | One running app that requires a GPU is automatically selected and given exclusive access. | Heavy workloads that need maximum performance, such as large models, rendering, or high-end gaming. |

:::warning Restart notice
Changing a GPU's mode unbinds all apps currently using that GPU and restarts their containers.

After the restart, Olares automatically reallocates GPU resources to running apps based on the newly selected mode.
:::

## View GPU status

To see your GPU configuration:

1. Go to **Settings** > **GPU**.
2. If your device has one GPU, Olares opens its details page directly.
![GPU overview single-GPU](/images/manual/olares/gpu-time-slicing-single.png#bordered){width=90%}

3. If your device has multiple GPUs, review the list to see each GPU's model, node, total VRAM, and current mode, then click a GPU to open its details page.
![GPU overview multiple-GPU](/images/manual/olares/gpu-overview.png#bordered){width=90%}

## Change a GPU mode

Follow these steps to change how a GPU is used:

1. Go to **Settings** > **GPU**.
2. Click the GPU you want to configure.
3. Choose a mode from the **GPU mode** dropdown.

## Single-GPU setup

Use this section if your Olares device has only one GPU.

### Time slicing

![Time slicing](/images/manual/olares/gpu-time-slicing-single.png#bordered){width=90%}

#### Manually add an app

In most cases, GPU resources are allocated automatically when an app runs.

If the target app does not appear in the list, or if no app appears:
1. Refresh the page.
2. In the **Pin application** section, click **Bind app**.
3. Select the app and click **Confirm**.

:::info
**Bind app** appears only when there are running apps that require GPU resources and do not currently have GPU access.
:::

#### Revoke an app's access to this GPU

1. Stop the app first.
2. Return to **Settings** > **GPU**. 
3. In the **Pin application** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

### Memory slicing

![Memory slicing](/images/manual/olares/gpu-mem-slicing-single.png#bordered){width=90%}

#### Adjust VRAM allocation

1. In the **Allocate VRAM** section, find the target app.
2. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
3. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

:::warning
The total VRAM allocated to all apps cannot exceed the GPU's total VRAM.

If the value is lower than the application's minimum requirement, **Confirm** is disabled.
:::

#### Manually add an app

In most cases, VRAM is allocated automatically when an app runs.

If the target app does not appear in the list, or if no app appears:
1. Refresh the page.
2. In the **Allocate VRAM** section, click **Bind app**.
3. Select the app, assign VRAM, and click **Confirm**.

#### Revoke an app's VRAM allocation

1. Stop the app first.
2. Return to **Settings** > **GPU**. 
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

### App exclusive

![App exclusive](/images/manual/olares/gpu-app-exclusive-single.png#bordered){width=90%}

#### Change the exclusive app

1. In the **Select exclusive app** section, click **Switch app**.
2. Choose the new app and click **Confirm**.

#### Manually set the exclusive app

In most cases, Olares assigns an exclusive app automatically.

If no app appears:
1. Refresh the page.
2. In the **Select exclusive app** section, click **Bind app**.
3. Select the app and click **Confirm**.

#### Revoke an app's exclusive access

1. Stop the app first.
2. Return to **Settings** > **GPU**. 
3. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

## Multi-GPU setup

Use this section if your system has multiple GPUs. 
:::info Multi-GPU scheduling
When multiple GPUs are available, Olares may distribute running apps across different GPUs. 

After switching GPU modes, running apps may be allocated to different GPUs. 

You can also reassign an app to a different GPU or fully revoke its GPU access.
:::
:::tip Can't find the target app?
If the app does not appear as an available option when you click **Bind app**, it may already be assigned to a GPU on another node.

Check other GPUs in the list and find where the app is currently assigned, then use **Switch GPU** to reassign it to the target GPU.
:::

### Time slicing

![Time slicing multi GPU](/images/manual/olares/gpu-time-slicing-multi.png#bordered){width=90%}

#### Reassign an app to another GPU

1. In the **Pin application** section, find the app.
2. Click <i class="material-symbols-outlined">repeat</i>.
3. Select the target GPU and click **Confirm**.

The app continues running during this process.

#### Manually add an app

If the app does not appear on the target GPU:

1. In the **Pin application** section, click **Bind app**.
2. Select the app and click **Confirm**.

#### Revoke the app's access to this GPU

1. If the app is bound only to this GPU, stop the app first.
2. Return to **Settings** > **GPU**. 
3. In the **Pin application** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

:::tip Unbind from multiple GPUs
If the app has also been allocated resources on other GPUs on the same node, you can clear the allocation from the current GPU without stopping it.
:::

### Memory slicing

![Multi-GPU memory slicing](/images/manual/olares/gpu-mem-slicing-multi.png#bordered){width=90%}

#### Adjust VRAM allocation

1. In the **Allocate VRAM** section, find the target app.
2. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
3. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

:::warning
The total VRAM allocated to all apps on the current GPU cannot exceed that GPU's total VRAM.

If the value is lower than the app's minimum requirement, **Confirm** is disabled.
:::

#### Reassign an app to another GPU

1. In the application list, find the app and click <i class="material-symbols-outlined">repeat</i>.
2. Choose the target GPU.
3. Click **Confirm**.

#### Manually add an app

If the target app does not appear on the target GPU:
1. In the **Allocate VRAM** section, click **Bind app**.
2. Select the app and click **Confirm**.

#### Revoke an app's VRAM allocation on this GPU

1. If the app is bound only to this GPU, stop it first.
2. Return to **Settings** > **GPU**.
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

:::tip
If the app has also been allocated VRAM on other GPUs on the same node, you can revoke the VRAM allocation from the current GPU without stopping it.
:::

### App exclusive

![Multi-GPU app exclusive](/images/manual/olares/gpu-app-exclusive-multi.png#bordered){width=90%}

#### Reassign the exclusive app to another GPU

1. Click <i class="material-symbols-outlined">repeat</i>.
2. Choose the target GPU.
3. Click **Confirm**.

#### Manually set the exclusive app

If no app appears on the target GPU:
1. In the **Select exclusive app** section, click **Bind app**.
2. Select the app and click **Confirm**.

#### Revoke an app's exclusive access to this GPU

1. If the app is bound only to this GPU, stop it first.
2. Return to **Settings** > **GPU**.
3. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

:::tip
If the app has also been allocated resources on other GPUs on the same node, you can remove it from the current GPU without stopping it.
:::

## Learn more
- [Monitor GPU usage in Olares](../resources-usage.md)