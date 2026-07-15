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

First, select the GPU mode that best matches your task needs. The following table summarizes how each mode handles resources.

| Workload needs | Recommended mode | How it works | Example |
| :--- | :--- | :--- | :--- |
| Multiple lightweight apps sharing resources | Time slicing (Default) | Apps take turns using the GPU's compute cores and VRAM. | Running several lightweight AI services or image tools simultaneously. |
| Single high-performance app requiring maximum power | App exclusive | One app gets full, uninterrupted access to a single GPU. | Fine-tuning large language models (LLMs) or high-end 3D rendering. |
| Multiple apps with strict VRAM limits | Memory slicing | The GPU's VRAM is divided into fixed quotas for concurrent apps. | Running different, VRAM-constrained AI models for multiple users. |

## 2. View GPU status

Before making changes, check your current GPU resource allocation.

1. Open the Settings app from the Dock or Launchpad on Olares, and then select **GPU**.
2. On the GPU details page, review the GPU's model, node, total VRAM, current working mode, and bound applications.
    ![Single GPU details](/images/one/single-gpu-details.png#bordered){width=70%}
3. If you have multiple GPUs, click a specific GPU from the list to view its details.
    ![Multiple GPU list](/images/one/multiple-gpu-list.png#bordered){width=70%}

## 3. Configure GPU mode

After choosing a GPU mode based on your task, follow the steps below to configure it.

### Share GPU among multiple apps (Time slicing) <Badge type="tip" text="Olares One only" />

To run multiple lightweight applications, the "Time slicing" mode is suitable.

This is the system's default mode. Applications without a specific binding are automatically scheduled to GPUs in this mode.

:::tip Switching to "Time slicing" from "App exclusive"
If you are switching to this mode from the **App Exclusive** mode:
- The previously exclusive application will be unbound and remain **Running**.
- However, other applications that were stopped due to exclusive mode will not resume automatically. You must manually resume these applications from **Market** > **My Olares** or from **Settings** > **Applications**.
:::

1. On the GPU details page, select **Time slicing** from the **GPU mode** list.
2. In the **Switch GPU mode** window, click **Confirm**. 
3. Bind an application to this GPU:

    a. In the **Pin application** section, click **Bind app**.

    b. Select the target application and click **Confirm**.

    c. Repeat the same steps to bind additional apps. 
    
    After binding, the applications will always use this GPU.
4. (Optional) To move a bound application to another GPU on the same node:

    a. Click <i class="material-symbols-outlined">repeat</i> next to the app.
    ![Switch GPU](/images/one/switch-gpu.png#bordered){width=70%}
    
    b. In the **Switch GPU** window, select a new GPU, and then click **Confirm**.
    
    c. In the **Unbind app** window, review your switch details again, and then click **Confirm**. 
    
    The app is immediately unbound from the current GPU, and it is automatically migrated to the new GPU as a bound app.

5. (Optional) To remove an app binding:

    a. Stop the app from **Settings** > **Applications** or from **Market** > **My Olares**.
    
    b. Return to the **Pin application** section and click <i class="material-symbols-outlined">link_off</i> next to the app. 

    c. In the **Unbind app** window, click **Confirm**.
    
    After unbinding, the application is removed from the bound app list automatically.

### Dedicate a GPU to a high‑performance app (App exclusive)

To run an application that exclusively occupies the entire GPU, use the "App exclusive" mode.

:::tip Switching to "App exclusive"
When switching to the **App Exclusive** mode from other modes ("Time slicing" or "Memory slicing"), all other applications currently running on this GPU will be stopped automatically to ensure the exclusive application receives full resources.
:::

1. On the target GPU details page, select **App exclusive** from the **GPU mode** list.
2. In the **Switch GPU mode** window, click **Confirm**.
3. In the **Select exclusive app** section, click **Bind app**.
    ![Select exclusive app](/images/one/bind-exclusive-app.png#bordered){width=70%}
4. Select the target application you want to grant exclusive access to and click **Confirm**.
5. (Optional) To move the exclusive application to another GPU on the same node:

    a. Click <i class="material-symbols-outlined">repeat</i> next to the app.
    ![Switch GPU](/images/one/switch-gpu-exclusive.png#bordered){width=70%}
    
    b. In the **Switch GPU** window, select a new GPU, and then click **Confirm**.

    :::tip Switch GPU unavailable
    In the **Switch GPU** window, if the new GPU you want to switch to is not available, it is because the current app has already been bound to it. Switch to that GPU and check the bound app list.
    ![Switch GPU not available](/images/one/switch-gpu-disabled.png#bordered){width=70%}    
    :::
    
    c. In the **Unbind app** window, review your switch details again, and then click **Confirm**. 
    
    The app is immediately unbound from the current GPU, and it is automatically migrated to the new GPU as a bound app.

6. (Optional) To remove the exclusive binding:

    a. Stop the app from **Settings** > **Applications** or from **Market** > **My Olares**.
    
    b. Return to the **Select exclusive app** section and click <i class="material-symbols-outlined">link_off</i> next to the app. 

    c. In the **Unbind app** window, click **Confirm**.
    
    After unbinding, the application is removed from the bound app list automatically.

### Run multiple apps with specific VRAM limits (Memory slicing)

To run multiple applications concurrently with precise control over each application's VRAM usage, use the "Memory slicing" mode.

:::tip Switching to "Memory slicing" from other modes
When switching to the **Memory slicing** mode from other modes ("Time slicing" or "App exclusive"):
- The system will automatically assign a minimum VRAM quota to each currently running application. You need to manually check and adjust these quotas to meet the actual needs of the applications.
- If switching from the "App exclusive" mode, the previously exclusive application will be unbound, and other applications that were stopped will not resume automatically. You must manually resume them from **Market** > **My Olares** or from **Settings** > **Applications**.
:::

1. On the target GPU details page, select **Memory slicing** from the **GPU mode** list.
2. In the **Switch GPU mode** window, click **Confirm**.
3. In the **Allocate VRAM** section, click **Bind app**.
4. In the **Bind app** window, select the target application, assign an appropriate VRAM quota to it in GB, and then click **Confirm**.

    :::tip
    The sum of all allocated VRAM quotas must not exceed the GPU's total VRAM.
    :::
5. Repeat the same allocation for other applications that need VRAM limits.
6. (Optional) To edit the VRAM quota for an app, click <i class="material-symbols-outlined">edit_square</i> next to the app, update the value, and then click **Confirm**.
7. (Optional) To move the application to another GPU on the same node:

    a. Click <i class="material-symbols-outlined">repeat</i> next to the app.
    
    b. In the **Switch GPU** window, select a new GPU, and then click **Confirm**.

    :::tip Switch GPU unavailable
    In the **Switch GPU** window, if the new GPU you want to switch to is not available, it is because the current app has already been bound to it. Switch to that GPU and check the bound app list.
    ![Switch GPU not available](/images/one/switch-gpu-disabled.png#bordered){width=70%}    
    :::
    
    c. In the **Unbind app** window, review your switch details again, and then click **Confirm**. 
    
    The app is immediately unbound from the current GPU, and it is automatically migrated to the new GPU as a bound app.

8. (Optional) To remove an application's VRAM binding:

    a. Stop the application from **Market** > **My Olares** or from **Settings** > **Applications**.
    
    b. Return to the **Allocate VRAM** section, click **Unbind** next to the app, and then click **Confirm**.

## FAQs

### What happens to my other applications when switching to "App exclusive" mode?

When switching from other modes ("Time slicing" or "Memory slicing") to "App exclusive", all other applications currently running on this GPU will be stopped to ensure the exclusive application receives full GPU resources.

### When switching back from "App exclusive" to another mode, will the previously stopped applications resume automatically?

No.

When switching from the "App exclusive" mode to another mode:
- The exclusive application will be unbound and remain **Running**.
- However, other applications that were stopped due to exclusive mode will not resume automatically, so they remain **Stopped**.

Therefore, you must manually resume these apps one by one from **Market** > **My Olares** or from **Settings** > **Applications**.

### How are application VRAM quotas set when switching to "Memory slicing" mode?

When switching from other modes to the **Memory slicing**, the system automatically assigns a minimum VRAM quota to each currently running application. However, this quota might not be sufficient for the application's actual runtime needs. You must manually check and adjust each application's VRAM quota to ensure they can run properly.

Note that the sum of all quotas must not exceed the GPU's total VRAM. So before any mode switch, it's recommended to check the current applications running on the GPU and plan your VRAM allocation in advance.
