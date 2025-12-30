---
outline: [2, 3]
description: Manage and optimize GPU resources in Olares with centralized controls, supporting time-slicing, exclusive access, and VRAM-slicing across single or multi-node setups.
---
# Manage GPU usage
:::info
Only Olares admins can change GPU modes. This helps avoid conflicts and keeps GPU performance predictable for everyone.
:::

Olares lets you manage your graphics cards, or GPUs, to speed up tasks like AI, image and video generation, and gaming. You can control how your applications use these resources from Olares Settings page.

This guide explains:
- How to choose the right GPU mode.
- How to configure GPU modes step by step.

## Choose the right GPU mode

Use the table below to pick a mode based on your workload.

| GPU mode | Definition | Use scenario |
| :--- | :--- | :--- |
| **Time slicing** | Multiple apps share one GPU<br> by taking turns using compute<br> and VRAM. | General workloads and experimentation with multiple lightweight apps. |
| **App exclusive** | One app gets full, uninterrupted<br> access to the compute and VRAM<br> of a single GPU. | Heavy workloads (e.g., LLMs, high-end gaming) requiring maximum stability. |
| **Memory slicing** | The GPU's VRAM is divided into<br> fixed quotas; apps run concurrently<br> within their limits. | Running specific apps simultaneously while strictly limiting their memory usage. |

## View GPU status

To see your GPUs and their current configuration:

1. Go to **Settings** > **GPU**.
2. Review the list to see each GPU's model, node, total VRAM, and current mode.
   ![GPU overview](/images/manual/olares/gpu-overview.png#bordered)
3. Click a GPU to open its details page.  

:::tip
If you have only one GPU, Olares may open the GPU details page directly.
:::

## Configure GPU mode

Follow these steps to change how a GPU is used:

1. Go to **Settings** > **GPU**.
2. Click the GPU you want to configure.
3. Choose a mode from the **GPU mode** dropdown.

:::warning Restart notice 
Changing a GPU's mode will unbind apps from that GPU and restart their containers.

After restart, if any GPU is in **Time slicing** mode, unbound apps may be automatically scheduled and bound to it.
:::

### Time slicing

**Time slicing** is the default mode in Olares. Use this mode to allow multiple applications to share resources.

Apps without a specific GPU binding may be automatically scheduled onto GPUs in **Time slicing** mode.

![Time slicing](/images/manual/olares/gpu-time-slicing.png#bordered)

#### Bind app
To assign an app to this GPU:
1. In **Pin application** section, click **Bind App**.
2. Choose your target application and click **Confirm**.

#### Switch GPU
If your system has more than one GPU, you can move an assigned app to a different GPU:
1. In **Pin application** section, find the app you want to move.
2. Click <i class="material-symbols-outlined">repeat</i>, then choose the target GPU and click **Confirm**.
:::info Same-node limitation
An application can use multiple GPUs only if they are located on the same node. If you switch an app to a GPU on a different node, the app is moved and bound only to the target GPU.
:::

#### Unbind app
To remove an app from this GPU:
1. In **Pin application** section, find the app you want to remove.
2. Click <i class="material-symbols-outlined">link_off</i>, then **Confirm**.

### App exclusive

Use **App exclusive** mode to dedicate a GPU entirely to one high-demand application.

![App exclusive](/images/manual/olares/gpu-app-exclusive.png#bordered)

#### Bind app
To give an app exclusive access:
1. In **Select exclusive app** section, click **Bind App**.
2. Select your target application and click **Confirm**.

#### Switch app
To replace the current exclusive app with a new one:
1. In **Select exclusive app** section, click **Switch App**.
2. Choose the new application and confirm.

The previous app is unbound, and the new app takes over exclusive access.

#### Switch GPU
If your system has more than one GPU, you can move the exclusive app to a different GPU:
1. In **Select exclusive app** section, click <i class="material-symbols-outlined">repeat</i>.
2. Choose the target GPU and confirm.

:::info Same-node limitation
An application can use multiple GPUs only if they are located on the same node. If you switch an app to a GPU on a different node, the app is moved and bound only to the target GPU.
:::

#### Unbind app
To remove the exclusive binding:
1. In **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>.
2. Click **Confirm**.

### Memory slicing

Use **Memory slicing** to run apps concurrently with strict VRAM limits.

![Memory slicing](/images/manual/olares/gpu-memory-slicing.png#bordered)

#### Bind app and allocate VRAM

1. In **Allocate VRAM** section, click **Bind App**.
2. Select your target application, assign it a specific amount of VRAM in GB, and click **Confirm**.
3. Repeat for other apps as needed.

The total of all VRAM limits must not exceed the GPU total VRAM.

## Learn more
- [Monitor GPU usage in Olares](../resources-usage.md)