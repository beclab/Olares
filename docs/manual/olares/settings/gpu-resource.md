---
outline: [2, 3]
description: Manage and optimize GPU resources in Olares with centralized controls, supporting time-slicing, exclusive access, and VRAM-slicing across single or multi-node setups.
---
# Manage GPU usage
:::info
Only Olares admins can change GPU modes. This helps avoid conflicts and keeps GPU performance predictable for everyone.
:::

Olares helps you manage your graphics cards, or GPUs, to speed up tasks like AI, image and video generation, and gaming. You can control how your applications use these resources from Olares Settings page.

This guide explains:
- How to choose the right GPU mode.
- The basic rules you need to know before assigning GPUs.
- How to configure GPU modes step by step.

## Choose the right GPU mode

If you are not sure which mode to use, start with this section.

### Time slicing
**Best for**: Sharing one GPU among several applications.

- Default mode for GPU assignment in Olares.
- Multiple apps can share the same GPU by taking turns.
- At any moment, one app uses the GPU's compute and VRAM.
- Suitable for general workloads and experimentation with many lightweight apps.
- Apps without a specific GPU binding are automatically scheduled onto GPUs that are in Time slicing mode.

### App exclusive
**Best for**: Giving one demanding app maximum and stable performance.

- One app gets full access to the compute and VRAM of a single GPU.
- No other app can use that GPU while the exclusive binding is active.
- Ideal for heavy workloads such as large language models or high-end gaming.

### Memory slicing
**Best for**: Running specific apps at the same time while limiting how much GPU memory each one can use.

- The GPU's VRAM is divided into fixed quotas.
- Each app is restricted to its assigned VRAM quota.
- Apps run concurrently instead of taking turns.
- The sum of all quotas cannot be more than the GPU's physical VRAM.

## Understand GPU assignment rules

Before assigning GPUs, understand these rules first.

### GPU and nodes

In Olares, each GPU belongs to a specific node managed by Olares. You can see which node a GPU belongs to in **Settings** > **GPU**.

Follow these two rules to avoid errors:

- **Same node only**: An application can use multiple GPUs only if they are on the same node.
- **No mixing nodes**: You cannot combine GPUs from different nodes for a single application.

### Switch GPU

Switching an app to another GPU always restarts the app.  
The result depends on where the target GPU is located:

| Target GPU location | Result after restart |
| --- | --- |
| Same node | The app may end up bound to both GPUs. |
| Different node | The app is moved and is bound only to the target GPU. |

## Hardware requirements

Olares supports NVIDIA GPUs only. The GPU must be Turing or newer. This includes Turing, Ampere, Ada Lovelace, and Blackwell.

### Check your GPU type

Check your GPU with one of these options:

| What you know | What to do |
| --- | --- |
| GTX or RTX 16 series or newer | Your card is likely supported. |
| Other NVIDIA model | Check the [compatible GPU table](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus). |
| Unknown model | Run `lspci \| grep -i nvidia` and then check<br> the compatible GPU table. |

:::warning AI performance
Even if your GPU is supported, some AI apps can fail if VRAM is too small. Always check the VRAM requirements of your workloads.
:::

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
3. Choose a mode from the **GPU Mode** dropdown.

:::warning Restart notice 
Changing a GPU's mode will unbind apps from that GPU and restart their containers.

After restart, if any GPU is in **Time slicing** mode, unbound apps may be automatically scheduled and bound to it.
:::

### Time slicing

![Time slicing](/images/manual/olares/gpu-time-slicing.png#bordered)

In **Time slicing** mode, you can perform the following actions.

<Tabs>
<template #Bind-app>

1. In **Pin application** section, click **Bind App**.
2. Choose your target application and click **Confirm**.
</template>
<template #Switch-GPU>

1. In **Pin application** section, find the app you want to move.
2. Click <i class="material-symbols-outlined">repeat</i>, then choose the target GPU and confirm.

| Target GPU location | Result |
| --- | --- |
| Same node | The app may use multiple GPUs after restart. |
| Different node | The app is moved and bound only to the target GPU. |
</template>
<template #Unbind-app>

1. In **Pin application** section, find the app you want to remove.
2. Click <i class="material-symbols-outlined">link_off</i>, then **Confirm**.
</template>
</Tabs>

### App exclusive

![App exclusive](/images/manual/olares/gpu-app-exclusive.png#bordered)

In **App exclusive** mode, you can perform the following actions.

<Tabs>
<template #Bind-app>

1. In **Select exclusive app** section, click **Bind App**.
2. Select your target application and click **Confirm**.

The app now has exclusive access to this GPU.
</template>
<template #Switch-app>

1. In **Select exclusive app** section, click **Switch App**.
2. Choose the new application and confirm.

The old app is unbound from this GPU. The new app becomes the exclusive app.
</template>
<template #Switch-GPU>

1. In **Select exclusive app** section, click <i class="material-symbols-outlined">repeat</i>.
2. Choose the target GPU and confirm.
</template>
<template #Unbind-app>

1. In **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>.
2. Click **Confirm**.
</template>
</Tabs>

### Memory slicing

![Memory slicing](/images/manual/olares/gpu-memory-slicing.png#bordered)

In **Memory slicing** mode, you can bind apps and set VRAM limits:

1. In **Allocate VRAM** section, click **Bind App**.
2. Select your target application, assign it a specific amount of VRAM in GB, and click **Confirm**.
3. Repeat for other apps as needed.

The total of all VRAM limits must not exceed the GPU total VRAM.

## Learn more
- [Monitor GPU usage in Olares](../resources-usage.md)