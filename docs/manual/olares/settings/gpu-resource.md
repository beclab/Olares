---
outline: deep
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

Before managing GPU resources, it helps to understand:
- What binding and unbinding mean.
- How an app's state affects available actions.
- Where to resume or stop an app.
- How GPU modes differ.

### Understand GPU allocation

In Olares, allocating GPU resources to an app is called "binding", while revoking that access so the GPU can be released or reassigned is called "unbinding".

Whether you can bind or unbind an app depends mainly on whether it is running or stopped. Use the table below as a quick reference:

| App state | Bind (Allocate GPU access) | Unbind (Revoke GPU access) |
| -- | -- | -- |
| **Running** | Supported | Stop the app first.* |
| **Stopped** | Resume the app first. | Supported |

*\*Multi-GPU exception: If an app is allocated to multiple GPUs on the same node, you can revoke its access from one GPU while it remains running on the others.*

:::info
Stopping an app pauses its workload, but it does not automatically clear its GPU allocation.

To fully release the GPU or VRAM for other workloads, you must explicitly unbind the app after stopping it.
:::

### Check or change an app's state

You can check whether an app is **Running** or **Stopped**, and change its state, in either of these places:

- **Market** > **My Olares**: The current status is displayed on the app's card. Click the dropdown menu to select **Stop** or **Resume**.
- **Settings** > **Applications**: The current status is shown in the app list. Select the app, then click **Stop** or **Resume**.

### Understand GPU modes

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

## Single-GPU setup

Use this section if your Olares device has only one GPU.

### Open GPU settings

1. Go to **Settings** > **GPU**.
2. Olares opens the GPU details page directly.
3. Choose a mode from the **GPU mode** dropdown.

### Time slicing

![Time slicing](/images/manual/olares/gpu-time-slicing-single.png#bordered){width=90%}

#### Manually add an app

In most cases, running apps are automatically bound and appear in the list.

If the target app does not appear:
1. Make sure the app is running.
2. Reload the GPU page in your browser.
3. In the **Pin application** section, click **Bind app**.
4. Select the app and click **Confirm**.

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

If the value is lower than the app's minimum requirement, **Confirm** is disabled.
:::

#### Manually add an app

In most cases, running apps are automatically bound and appear in the list.

If the target app does not appear:
1. Make sure the app is running.
2. Reload the GPU page in your browser.
3. In the **Allocate VRAM** section, click **Bind app**.
4. Select the app, assign VRAM, and click **Confirm**.

#### Revoke an app's VRAM allocation

1. Stop the app first.
2. Return to the GPU page.
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

### App exclusive

![App exclusive](/images/manual/olares/gpu-app-exclusive-single.png#bordered){width=90%}

#### Change the exclusive app

1. Make sure the new target app is running.
2. In the **Select exclusive app** section, click **Switch app**.
3. Choose the new app and click **Confirm**.

#### Manually set the exclusive app

In most cases, Olares automatically selects one running app for exclusive access.

If no app appears:
1. Make sure the target app is running.
2. Reload the GPU page in your browser.
3. In the **Select exclusive app** section, click **Bind app**.
4. Select the app and click **Confirm**.

#### Revoke an app's exclusive access

1. Stop the app first.
2. Return to the GPU page.
3. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

## Multi-GPU setup

Use this section if your system has multiple GPUs.

### Open GPU settings

1. Go to **Settings** > **GPU**.
2. Review the list to see each GPU's model, node, total VRAM, and current mode.
3. Click the GPU you want to configure.
4. Choose a mode from the **GPU mode** dropdown.

![GPU overview multiple-GPU](/images/manual/olares/gpu-overview.png#bordered){width=90%}

:::info Multi-GPU scheduling
When multiple GPUs are available, Olares may distribute running apps across different GPUs.

After switching GPU modes, running apps may be allocated to different GPUs.

You can also reassign an app to a different GPU or fully revoke its GPU access.
:::

:::tip Can't find the target app?
If the target app does not appear as an available option when you click **Bind app**, it may already be assigned to another GPU on the same node or on another node.

Check other GPUs to find where the app is currently assigned, then use **Switch GPU** to reassign it to the target GPU.
:::

### Time slicing

![Time slicing multi GPU](/images/manual/olares/gpu-time-slicing-multi.png#bordered){width=90%}

#### Manually add an app

If the target app is not assigned to the current GPU:

1. In the **Pin application** section, click **Bind app**.
2. Select the app and click **Confirm**.

#### Reassign an app to another GPU

1. In the **Pin application** section, find the app.
2. Click <i class="material-symbols-outlined">repeat</i>.
3. Select the target GPU and click **Confirm**.

The app continues running during this process.

#### Revoke an app's access to this GPU

1. If the app is bound only to this GPU, stop it first.
2. Return to **Settings** > **GPU**.
3. In the **Pin application** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

:::tip Unbind from multiple GPUs
If the app is still allocated to other GPUs on the same node, you can revoke its access from the current GPU without stopping it.
:::

### Memory slicing

![Multi-GPU memory slicing](/images/manual/olares/gpu-mem-slicing-multi.png#bordered){width=90%}

#### Manually add an app

If the target app is not assigned to the current GPU:

1. In the **Allocate VRAM** section, click **Bind app**.
2. Select the app and click **Confirm**.

#### Adjust VRAM allocation

1. In the **Allocate VRAM** section, find the target app.
2. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
3. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

:::warning
The total VRAM allocated to all apps on the current GPU cannot exceed that GPU's total VRAM.

If the value is lower than the app's minimum requirement, **Confirm** is disabled.
:::

#### Reassign an app to another GPU

1. In the **Allocate VRAM** section, find the app and click <i class="material-symbols-outlined">repeat</i>.
2. Choose the target GPU.
3. Click **Confirm**.

#### Revoke an app's VRAM allocation on this GPU

1. If the app is bound only to this GPU, stop it first.
2. Return to **Settings** > **GPU**.
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

:::tip
If the app has also been allocated VRAM on other GPUs on the same node, you can revoke the VRAM allocation from the current GPU without stopping it.
:::

### App exclusive

![Multi-GPU app exclusive](/images/manual/olares/gpu-app-exclusive-multi.png#bordered){width=90%}

#### Manually set the exclusive app

If no app is selected on the current GPU:
1. In the **Select exclusive app** section, click **Bind app**.
2. Select the app and click **Confirm**.

#### Reassign the exclusive app to another GPU

1. In the **Select exclusive app** section, click <i class="material-symbols-outlined">repeat</i>.
2. Choose the target GPU.
3. Click **Confirm**.

#### Revoke an app's exclusive access to this GPU

1. If the app is bound only to this GPU, stop it first.
2. Return to **Settings** > **GPU**.
3. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>, then click **Confirm**.

:::tip
If the app has also been allocated resources on other GPUs on the same node, you can remove it from the current GPU without stopping it.
:::

## Learn more
- [Monitor GPU usage in Olares](../resources-usage.md)