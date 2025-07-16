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

::: tip Nvidia GPU only
Currently, only Nvidia GPU is supported.
:::

## Understand GPU allocation modes

Olares supports three GPU allocation modes. Choosing the right mode helps optimize performance based on your needs.

### Time slicing 

In this mode, the GPU's processing power is shared among multiple applications.  

* Acts as a default resource pool. Any application not explicitly assigned to a specific GPU will automatically use a time-slicing GPU if available.

* Suitable for General-purpose use and running multiple lightweight applications.

### App exclusive

In this mode, the entire GPU processing power and memory is dedicated to a single application. 

* Best for intensive, performance-critical applications like AI-generated imagery or high-performance gaming servers.
* Large memory demands may limit availability for other tasks.

### Memory slicing
In this mode, GPU memory (VRAM) is partitioned into fixed, dedicated amounts for specific applications.

* Ideal for running multiple GPU-intensive applications simultaneously, each with guaranteed VRAM allocation.
* Prevents memory conflicts between applications running on the same GPU.

## View GPU status

To view your GPU status:

1. Navigate to **Settings > GPU**. The GPU list shows each GPU’s model, associated node, total VRAM, and current GPU mode.
2. Click on a specific GPU to visit its details.

::: tip Note
If your Olares only contains one GPU, navigating to the GPU section will take you directly to the GPU details page. If you have multiple GPUs, you will see a list first.
:::

## Configure GPU mode

On the **GPU details** page, select your desired mode from the **GPU mode** dropdown. Depending on your selected mode, different follow-up options apply.

* **Time slicing**：   
  1. Select this mode from the GPU mode dropdown.
  2. In the **Application pinning** section, click **+Add an application** button to manually pin an application to this specific GPU in a multi-GPU setup.

:::tip Note
No manual pinning is required if you only have one GPU in your cluster.
:::
  
* **App exclusive**
  1. Select this mode from the GPU mode dropdown.
  2. In the **Select exclusive app** dropbox, choose your target application.
  3. Click **Confirm**.
    ![App exclusive](/images/manual/olares/gpu-app-exclusive.png#bordered)

* **Memory slicing**
    1. Select this mode from the dropdown.
    2. In the **Allocate VRAM** section, click **Add application**. 
    3. Select your target application and assign it a specific amount of VRAM (in GB).
    4. Repeat for other applications and click **Confirm**.
       ![VRAM slicing](/images/manual/olares/gpu-memory-slicing.png#bordered)
     
  ::: tip Note
  You can't assign a VRAM that's larger than the total VRAM.
   :::

## Learn more
- [Monitor GPU usage in Olares](../resources-usage.md)