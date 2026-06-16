---
outline: [2,3]
description: Manage accelerator resources in Olares, including GPU modes, app GPU requirements, app resume, and resource release.
---
# Manage accelerator resources

Olares uses accelerator resources to run workloads such as AI inference, image and video generation, transcoding, and gaming. An accelerator can be a GPU, an integrated accelerator chip, or a CPU-based fallback resource.

This guide helps you understand how Olares allocates accelerator resources, then walks through common tasks such as changing a GPU mode, resuming a stopped app with GPU resources, and removing an app from a GPU.

## Understand accelerator resources

### How allocation works

Olares treats GPUs and other acceleration devices as shared resources in the cluster. Instead of asking each app to manage hardware directly, Olares checks the app's accelerator requirement, finds matching resources, and binds the app to a resource that has enough available capacity.

For GPU-dependent apps, Olares considers several factors during allocation:

- Resource type, such as **NVIDIA GPU**.
- GPU mode, such as **Time slicing**, **Memory slicing**, or **Exclusive**.
- Total and free VRAM.
- Whether the app can bind to one GPU or multiple GPUs.
- Apps already assigned to the same resource.

When you install a new app that requires GPU resources, Olares automatically binds it to a suitable GPU if enough matching capacity is available. If the app is stopped later, or if you want to choose a different GPU or mode, you can resume the app and assign resources manually.

### Supported resource types

Olares can display the following accelerator resource types:

| Resource type | Description |
| -- | -- |
| **NVIDIA** | General NVIDIA GPUs. |
| **NVIDIA GB10** | NVIDIA GB10 SoC devices. |
| **Apple M-series** | Apple M-series chips. |
| **AMD Strix Halo** | AMD Strix Halo devices. |
| **CPU** | Fallback resource used when no accelerator device is available. |

:::info
AMD GPU and AMD APU devices are not covered in this version.
:::

### GPU modes

For supported GPU resources, the mode controls how apps share the GPU. Not every accelerator supports all three modes. The mode dropdown only shows the modes available for that specific resource.

| Mode | Description | Usage & notes |
| -- | -- | -- |
| **Time slicing** | Multiple apps share one GPU<br> by taking turns using compute<br> and VRAM. Each app can see <br>the full GPU, but workloads run<br> in slices. | <ul><li>**Best for**: Running several GPU-dependent apps at the same time.</li><li>**Note**: Launch may be blocked if the selected node's system RAM usage reaches the 90% threshold, even when GPU VRAM is enough.</li></ul> |
| **Memory slicing** | Multiple apps share one GPU<br> with fixed VRAM allocations. | <ul><li>**Best for**: Running multiple GPU-dependent apps while controlling VRAM usage. </li><li>**Note**: During launch, Olares allocates the app's minimum required VRAM by default. You can adjust the allocation.</li></ul> |
| **Exclusive** | One app gets full access to the<br> GPU. | <ul><li>**Best for**: Running heavy workloads that need maximum performance, such as large models, rendering, or high-end gaming. </li><li>**Note**: No other apps can bind to this GPU until it is released.</li></ul> |

## View accelerator resources

Go to **Settings** > **Accelerator** to view GPU and other accelerator resources across all nodes in the cluster.

The page shows each available resource with:

- Node name.
- Resource type, such as **NVIDIA GPU**.
- GPU model, such as **NVIDIA GeForce RTX 4060 Ti**.
- Current sharing mode, such as **Time slicing**, **Memory slicing**, or **Exclusive**.
- Used and total VRAM.
- Apps currently assigned to the resource.

![Accelerator overview](/images/manual/olares/settings-gpu-info.png#bordered)

## Change a GPU mode

:::warning
Switching modes stops all apps currently using this GPU and removes their assigned GPU resources. To use these apps again, reassign resources and launch them.
:::

1. Go to **Settings** > **Accelerator**, then click **Manage** on the resource card.
2. On the **Manage node** page, find the target GPU, open its mode dropdown, and select **Time slicing**, **Memory slicing**, or **Exclusive**.

   ![Select mode](/images/manual/olares/settings-gpu-dropdown.png#bordered)

## Resume an app and assign resources

You can resume a stopped app and manually assign resources from either **Settings** or **Market**.

1. Open one of the following locations:
   - **Settings** > **Applications**.
   - **Market** > **My Olares**.
2. Find the target app and click **Resume**.
3. Review the app's requirements at the top of the launch dialog, including resource type, GPU binding capability, and VRAM requirement.
4. Under **Select GPUs**, choose an **Available** resource. Olares automatically lists resources that match the app's requirements. If a resource cannot be selected, check the inline message:
   - **Insufficient free VRAM**: The GPU has enough total VRAM, but too much VRAM is already assigned to other apps. Expand **Assigned apps**, then click **Remove** for an app to free resources.
   - **Not enough VRAM**: The GPU's total VRAM is below the app's requirement. Removing apps will not help. Select another GPU.

   ![Resume an app](/images/manual/olares/settings-gpu-resume-app.png#bordered){width=70%}

5. For **Memory slicing** mode, VRAM defaults to the minimum requirement. To adjust it, click **Configure**, enter an amount, and click **Confirm**.

   ![Set VRAM in Memory slicing mode](/images/manual/olares/settings-gpu-mem-slicing-vram.png#bordered){width=70%}

   :::info Reallocate VRAM
   You cannot adjust VRAM after the app is assigned. To change it later, you must remove the app from the GPU and resume it again.
   :::

6. Click **Launch**.

## Remove an app from a GPU

Removing an app from a GPU stops the app and releases its hardware resources. Use this workflow when you want to actively free up resources from **Settings**, even if you are not launching another app right now.

To use the app again later, you must assign resources and launch it again.

:::warning
If an app is bound to multiple GPUs or GPUs across multiple nodes, removing the app from one GPU releases all GPU resources assigned to that app and stops the app.
:::

1. Go to **Settings** > **Accelerator**, then click **Manage** on the resource card.
2. Under **Assigned apps**, find the target app, then click **Remove**.

   ![Remove an app](/images/manual/olares/settings-gpu-remove-app.png#bordered)

3. In the confirmation dialog, click **Remove**.

:::info
Resource release may take a short time. If no resource can be assigned to the app immediately after removal, refresh the resource list and check again.
:::
