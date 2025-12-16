---
outline: [2, 3]
description: Manage and optimize GPU resources in Olares with centralized controls, supporting time-slicing, exclusive access, and VRAM-slicing across single or multi-node setups.
---
# Manage GPU usage
:::info
Only Olares admin can configure GPU usage mode. This ensures optimal resource management across the system and prevents conflicts between users' resource needs.
:::

Olares allows you to harness the full power of your GPUs to accelerate demanding tasks such as large language models, image and video generation, and gaming. Whether your GPUs are on a single node or spread across multiple nodes, you can manage them conveniently from one centralized interface.

This guide helps you understand and configure GPU allocation modes to maximize hardware performance.

## Multi-GPU and nodes

Before configuring your GPUs, it is critical to understand how Olares handles applications across different physical nodes.

- **Same node:** You **can** assign multiple GPUs to a single application if all those GPUs are located on the **same physical node**.
- **Cross-node:** You **cannot** assign multiple GPUs to a single application if the GPUs are spread across **different nodes**.

:::warning Configuration failure
If you attempt to bind an application to GPUs located on different nodes, the configuration will fail. 

Always verify which node a GPU belongs to in the **Settings > GPU** list before assigning it.
:::

:::tip Switch vs. bind across nodes
- When you **switch** an app from one GPU to another on the **same node**, the app can end up bound to **both GPUs** (multi-GPU on a single node).
- When you **switch** an app to a GPU on a **different node**, the app is **moved** so it remains bound **only** to the target GPU (no cross-node multi-GPU).
:::

## GPU compatibility

Olares supports **only Nvidia GPUs** of **Turing architecture or later** (Turing, Ampere, Ada Lovelace, and Blackwell).

- **Quick check:** GTX/RTX **16 series and newer** consumer cards are supported.
- **For other models:** Cross-check with the [compatible GPU table](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus).
- **Unknown model:** Run `lspci | grep -i nvidia` to query the GPU architecture code and determine compatibility.

:::warning AI performance
Even if your GPU architecture is supported, **low VRAM capacity may cause AI applications to fail**. Ensure your GPU has enough memory for your workloads.
:::

## Understand GPU allocation modes

Olares supports three GPU allocation modes. Choosing the right mode helps optimize performance based on your needs.

### Time Slicing

In this mode, a GPU can be bound to multiple applications and rotates execution in time slices.

- **Shared access**: At any instant, only one application uses all available compute and VRAM of the GPU.
- **Auto-scheduling:** Applications without allocated GPU resources automatically join the time-sliced GPU queue.
- **Queuing:** Other apps enter a wait queue; their VRAM contents (e.g., CUDA context, etc.) may be temporarily swapped out to system memory.

:::info Default GPU allocation
By default, GPUs run in **Time Slicing** mode. Applications without explicitly allocated GPU resources are automatically scheduled onto GPUs in Time Slicing mode and will be automatically bound to them.
:::


### App Exclusive

In this mode, the entire GPU is allocated to a single application.

- **Dedicated resources:** The bound application has exclusive access to all compute and VRAM of the bound GPU.
- **Performance:** Eliminates cross-app contention and scheduling overhead, providing the best possible performance for heavy workloads.

### Memory Slicing

In this mode, VRAM of the GPU is partitioned into fixed quotas for multiple designated applications.

- **Fixed quotas:** You manually set a VRAM limit (in GB) for each app.
- **Concurrency:** Apps run concurrently, restricted to their assigned quota.
- **Limit:** The total allocated quota cannot exceed the physical VRAM of the GPU (no oversubscription).

:::tip Multi-GPU allocation
All three modes support assigning multiple GPUs to the same application, **provided they are on the same node**. Olares assigns these GPUs to the application container; whether the application actually utilizes multiple GPUs depends on the application's own framework support.
:::

## View GPU status

To view your GPU inventory and status:

1. Navigate to **Settings** > **GPU**. The GPU list shows each GPU's model, associated node, total VRAM, and current GPU mode.
2. Click on a specific GPU to access its details page.
  ![GPU overview](/images/manual/olares/gpu-overview.png#bordered)

::: tip Note
If your Olares only has one GPU, navigating to the GPU section will take you directly to the GPU details page.
:::

## Configure GPU mode

On the GPU details page, select your desired mode from the **GPU Mode** dropdown. 

### Time Slicing

![Time slicing](/images/manual/olares/gpu-time-slicing.png#bordered)

1. Select **Time Slicing** from the dropdown.
2. In the **Pin application** section, you can bind applications and manage existing bindings.
   - **Bind an app**: 

     Click **Bind App** to manually pin an application to this GPU in Time Slicing mode.  
     If you don't bind anything manually, the scheduler can still assign apps automatically to this GPU because Time Slicing is the default mode.

   - **Switch GPU**:  
     Click <i class="material-symbols-outlined">repeat</i> beside an app to switch it to another GPU. Behavior depends on the target GPU's node:

     - **Same-node target:**  
       The application container restarts, and after restart the app is bound to **both** the original GPU and the new target GPU (multi-GPU on one node).

     - **Different-node target:**  
       The application container restarts, and the app is **moved** so that it is now bound **only** to the target GPU. It no longer appears under the original GPU.
   - **Unbind app**:  
     Click <i class="material-symbols-outlined">link_off</i> to unbind the app from the current GPU.

     - Since Time Slicing is the default mode, an app with no bound GPU may be automatically scheduled back onto **any** GPU that is in Time Slicing mode and automatically bound to it (including this GPU).

### App Exclusive

![App exclusive](/images/manual/olares/gpu-app-exclusive.png#bordered)
1. Select **App Exclusive** from the dropdown.
2. In the Select exclusive app section, click **Bind App** to add your target application and click **Confirm**.  
   This binds the selected app to this GPU with exclusive access.
3. Once bound, you can perform the following in the Select exclusive app
 section:

   - **Switch App**:
     Click **Switch App** to replace the current exclusive app with a different one. The old app is unbound from this GPU; the new one becomes the exclusive one.
   - **Switch GPU**:
     Click <i class="material-symbols-outlined">repeat</i> to move this exclusive binding to another GPU.  
   - **Unbind**:  
     Click <i class="material-symbols-outlined">link_off</i> to unbind the app and release the GPU completely.

### Memory Slicing
![VRAM slicing](/images/manual/olares/gpu-memory-slicing.png#bordered)
1. Select **Memory Slicing** from the dropdown.
2. In the Allocate VRAM section, click **Bind App**.
3. Select your target application, assign it a specific amount of VRAM in GB, and click **Confirm**.
4. Repeat **Bind App** for other applications if needed, as long as the total quota does not exceed the GPU's physical VRAM.

:::info Mode switching
Changing a GPU's mode (for example, from **Time Slicing** to **App Exclusive**) will unbind apps from that GPU and restart their containers. 

After restart, apps follow the normal scheduling rules: if there is any GPU running in **Time Slicing** mode, unbound apps may be automatically scheduled and bound to a Time Slicing GPU.
:::

## Learn more
- [Monitor GPU usage in Olares](../resources-usage.md)