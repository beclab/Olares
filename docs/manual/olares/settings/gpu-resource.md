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

::: tip GPU support
Olares supports **only Nvidia GPUs** of **Turing architecture or later** (Turing, Ampere, Ada Lovelace, and Blackwell).

- Quick check: GTX/RTX **16 series and newer** consumer cards are supported.
- For other models, cross-check with the [compatible GPU table](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus).
- Other models: Cross-check with the [compatible GPU table](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus).
- Unknown model: Run `lspci | grep -i nvidia` to query the GPU architecture code and determine compatibility.  
  :::

:::warning AI Performance
Even if your GPU architecture is supported, **low VRAM capacity may cause AI applications to fail**. Ensure your GPU has enough memory for your workloads.
:::

## Understand GPU allocation modes

Olares supports three GPU allocation modes. Choosing the right mode helps optimize performance based on your needs.

### Time Slicing

In this mode, a GPU can be bound to multiple applications and rotates execution in time slices.

* At any instant, only one application uses all available compute and VRAM of the GPU.
* Other apps enter a wait queue; their CUDA and VRAM content will be swapped to the system memory.

### App Exclusive

In this mode, the entire GPU is allocated to a single application.

* During execution, the app can use all compute and VRAM of the bound GPU.
* No cross-app contention or scheduling overhead so that best performance is guaranteed.

### Memory Slicing
In this mode, VRAM of the GPU is partitioned into fixed quotas for multiple designated applications.

* Users need to manually set a quota for each app.
* The sum of quotas must not exceed physical VRAM of the bound GPU. Oversubscription is not supported.
* Apps with quota assigned can run concurrently, each limited to its own quota.

:::tip Multi-GPU aggregation
You can bind multiple GPUs to one application within the same cluster to gain bigger VRAM. In such scenarios, only **App Exclusive** or **Memory Slicing** modes are supported.
:::

## View GPU status

To view your GPU status:

1. Navigate to **Settings** > **GPU**. The GPU list shows each GPU’s model, associated node, total VRAM, and current GPU mode.
2. Click on a specific GPU to visit its details.

![GPU overview](/images/manual/olares/gpu-overview.png#bordered)

::: tip Note
If your Olares only has one GPU, navigating to the GPU section will take you directly to the GPU details page.
:::

## Configure GPU mode

On the **GPU details** page, select your desired mode from the **GPU mode** dropdown. Depending on your selected mode, different follow-up options apply.

* **Time Slicing**：
  1. Select this mode from the GPU mode dropdown.
  2. In the **Pin application** section, click **+Add an application** to manually pin an application to this specific GPU in a multi-GPU setup.

  ![Time slicing](/images/manual/olares/gpu-time-slicing.png#bordered)

:::tip Note
No manual pinning is required if you only have one GPU in your cluster.
:::

* **App Exclusive**
  1. Select this mode from the GPU mode dropdown.
  2. In the **Select exclusive app** dropbox, choose your target application.
  3. Click **Confirm**.
     ![App exclusive](/images/manual/olares/gpu-app-exclusive.png#bordered)

* **Memory Slicing**
  1. Select this mode from the dropdown.
  2. In the **Allocate VRAM** section, click **Add an application**.
  3. Select your target application and assign it a specific amount of VRAM in GB.
  4. Repeat for other applications and click **Confirm**.
     ![VRAM slicing](/images/manual/olares/gpu-memory-slicing.png#bordered)

## Learn more
- [Monitor GPU usage in Olares](../resources-usage.md)