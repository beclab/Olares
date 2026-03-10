---
outline: [2, 3]
description: Learn about the three GPU modes in Olares and how to switch between them to match different workloads.
head:
  - - meta
    - name: keywords
      content: Olares One, NVIDIA DGX Spark, GPU, Time slicing, App exclusive, Memory slicing, GPU management
---

# Manage GPU resources

Graphics Processing Units (GPUs) can significantly accelerate tasks such as AI processing, image generation, and 3D rendering. Olares allows you to flexibly manage your GPU resources to ensure every application gets the exact performance it needs. 

As an administrator, you can balance resource usage and application performance by assigning specific working modes to your GPUs.

## 1. Choose a GPU mode

First, select the GPU mode that best matches your task needs. The following table summarizes how each mode handles resources:

| Workload needs | Recommended mode | How it works | Example |
| :--- | :--- | :--- | :--- |
| Multiple light-weight apps sharing resources | Time slicing (Default) | Apps take turns using the GPU's compute cores and VRAM. | Running several light-weight AI services or image tools simultaneously. |
| Single high-performance app requiring maximum power | App exclusive | One app gets full, uninterrupted access to a single GPU. | Fine-tuning large language models (LLMs) or high-end 3D rendering. |
| Multiple apps with strict VRAM limits | Memory slicing | The GPU's VRAM is divided into fixed quotas for concurrent apps. | Running different, VRAM-constrained AI models for multiple users. |

## 2. View GPU status

Before making changes, check your current GPU resource allocation.

1. Go to **Settings** > **GPU**.
2. On the GPU details page, review the GPU's model, node, total VRAM, current gpu working mode, and bound applications.
3. If you have multiple GPUs, click a GPU in the list to view its details.

## 3. Configure GPU mode

After choosing a GPU mode based on your task, follow the steps below to configure it. 

:::info
Changing a GPU's mode will automatically restart all application containers currently associated with this GPU.
:::

### Share GPU among multiple apps (Time slicing)

To run multiple light-weight applications, the "Time slicing" mode is suitable.

This is the system's default mode. Applications without a specific binding are automatically scheduled to GPUs in this mode.

:::tip Switching to "Time slicing" from "App exclusive"
If switching to this mode from the **App Exclusive** mode:
- The previously exclusive application will be unbound and remain **Running**.
- However, other applications that were stopped due to exclusive mode will not resume automatically. You must manually resume these applications from **Market** > **My Olares** or from **Settings** > **Applications**.
:::

1. On the GPU details page, select **Time slicing** from the **GPU mode** list.
2. In the **Switch GPU mode** window, click **Confirm**. 
3. (Optional) To bind a specific application to this GPU:

    a. In the **Pin application** section, click **Bind app**.

    b. Select the target application and click **Confirm**. 
    
    After binding, the application will always use this GPU.
4. (Optional) To move a bound application to another GPU on the same node, click <i class="material-symbols-outlined">repeat</i> and confirm. The application will be migrated to the new target GPU.

### Dedicate a GPU to a high‑performance app (App exclusive)

To run an application that exclusively occupies the entire GPU, use the "App exclusive" mode.

:::tip Switching to "App exclusive"
When switching to the **App Exclusive** mode from other modes ("Time slicing" or "Memory slicing"), all other applications currently running on this GPU will be stopped automatically to ensure the exclusive application receives full resources.
:::

1. On the GPU details page, select **App exclusive** from the **GPU mode** list.
2. In the **Switch GPU mode** window, click **Confirm**.
3. In the **Select exclusive app** section, click **Bind app**.
4. Select the target application you want to grant exclusive access to and click **Confirm**.
5. (Optional) To replace the exclusive application:

    a. First, go to **Market** > **My Olares** or **Settings** > **Applications** to resume the new application to the **Running** state.

    b. Return to the **Select exclusive app** section, click **Switch app**, select the new application, and then click **Confirm**. 
    
    The original exclusive application will be automatically unbound and removed from the list.

6. (Optional) To remove the exclusive binding, click **Unbind**. After unbinding, the application will be stopped automatically and you need to manually resume it.
7. (Optional) To move the exclusive application to another GPU on the same node, click <i class="material-symbols-outlined">repeat</i> and confirm.

    :::info
    An application can use multiple GPUs only if they are located on the same node. If you switch the application to a GPU on a different node, the application will be moved out of the original node and bound only to that target GPU.
    :::

### Run multiple apps with strict VRAM limits (Memory slicing)

To run multiple applications concurrently with precise control over each application's VRAM usage, use the "Memory slicing" mode.

:::tip Switching to "Memory slicing" from other modes
When switching to the **Memory slicing** mode from other modes ("Time slicing" or "App exclusive"):
- The system will automatically assign a minimum VRAM quota to each currently running application. You need to manually check and adjust these quotas to meet the actual needs of the applications.
- If switching from the "App exclusive" mode, the previously exclusive application will be unbound, and other applications that were stopped will not resume automatically. You must manually resume them from **Market** > **My Olares** or from **Settings** > **Applications**.
:::

1. On the GPU details page, select **Memory slicing** from the **GPU mode** list.
2. In the **Switch GPU mode** window, click **Confirm**.
3. In the **Allocate VRAM** section, click **Bind app**.
4. Select the target application and assign an appropriate VRAM quota in GB.

    :::tip
    The sum of all allocated VRAM quotas must not exceed the GPU's total VRAM.
    :::
5. Click **Confirm** and repeat for other applications that need VRAM limits.
6. (Optional) To remove an application's VRAM binding:

    a. First, manually stop the application from **Market** > **My Olares** or from **Settings** > **Applications**.
    
    b. Return to the **Allocate VRAM** section, click **Unbind** next to the app, and then click **Confirm**.

## FAQs

### What happens to my other applications when switching to "App exclusive" mode?

When switching from other modes ("Time slicing" or "Memory slicing") to "App exclusive", all other applications currently running on this GPU will be stopped to ensure the exclusive application receives full GPU resources.

### When switching back from "App exclusive" to another mode, will the previously stopped applications resume automatically?

No.

When switching from the "App exclusive" mode to another mode:
- The exclusive application will be unbound and remain **Running**.
- However, other applications that were stopped due to exclusive mode will not resume automatically, so they remain **Stopped**.

Therefore, you need to manually resume these apps one by one from **Market** > **My Olares** or from **Settings** > **Applications**.

### How are application VRAM quotas set when switching to "Memory slicing" mode?

When switching from other modes to the **Memory slicing**, the system automatically assigns a minimum VRAM quota to each currently running application. However, this quota might not be sufficient for the application's actual runtime needs. 

Therefore, you need to manually check and adjust each application's VRAM quota to ensure they can run properly.

Note that the sum of all quotas must not exceed the GPU's total VRAM. So before any mode switch, it's recommended to check the current applications running on the GPU and plan accordingly in advance.
