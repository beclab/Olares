---
outline: [2, 3]
description: Learn how to manage accelerator resources on Olares One, including GPU modes, app resume, and resource release.
head:
  - - meta
    - name: keywords
      content: Accelerator, GPU, Time slicing, Exclusive, Memory slicing, Olares One
---
# Manage accelerator resources <Badge text="10 min"/>

Olares One uses accelerator resources for local LLMs, image and video generation, transcoding, and gaming. The built-in NVIDIA GPU is the main accelerator resource. Supported external GPUs can also appear in **Settings** > **Accelerator**.

In most cases, Olares automatically assigns a suitable GPU when you install a GPU-dependent app. You only need to manage accelerator resources manually when you want to check GPU usage, change how the GPU is shared, resume a stopped app, or release resources for another app.

## Learning objectives

By the end of this tutorial, you will learn how to:

- Check GPU usage in **Settings**.
- Understand the three GPU sharing modes.
- Change a GPU mode.
- Resume a stopped app and assign resources.
- Remove an app from a GPU.

## Understand GPU sharing modes

GPU modes control how apps share the GPU. For most users, the main choice is whether to let several apps share the GPU or give one app dedicated access. Not every GPU supports all three modes. The mode dropdown only shows the modes available for that GPU.

| Mode | How it works | When to use it |
| -- | -- | -- |
| **Time slicing** | Several apps share one GPU by<br> taking turns. Each app can see<br> the full GPU, but workloads run<br> in slices. | Use this for general workloads or when you want several GPU apps available at the same time. |
| **Memory slicing** | Several apps share one GPU with<br> fixed VRAM allocations. | Use this when you want to run multiple apps and control how much VRAM each app can use. |
| **Exclusive** | One app gets full access to the<br> GPU. Other apps cannot bind to<br> it until it is released. | Use this for heavy workloads that need maximum performance, such as large models, rendering, or high-end gaming. |

:::info
In **Time slicing** mode, Olares also checks node system memory. The current estimate is the app's memory requirement plus the GPU's dedicated VRAM. Launch is blocked if adding that amount would bring node memory usage to 90% or more. See [Manage accelerator resources](/manual/olares/settings/gpu-resource.md#gpu-modes) for an example.
:::

## View accelerator resources

1. Go to **Settings** > **Accelerator**.
2. Review the resource card. It shows the node name, resource type, dedicated VRAM, GPU model, current mode, and apps currently assigned to the resource.

   ![Accelerator overview](/images/manual/olares/settings-gpu-info.png#bordered)

## Change a GPU mode

:::warning
Switching modes stops all apps currently using this GPU and removes their assigned GPU resources. To use these apps again, reassign resources and launch them.
:::

1. Go to **Settings** > **Accelerator**, then click **Manage** on the resource card.
2. On the **Manage node** page, find the target GPU.
3. Open the mode dropdown and select **Time slicing**, **Memory slicing**, or **Exclusive**.

   ![Select mode](/images/manual/olares/settings-gpu-dropdown.png#bordered)

## Resume an app and assign resources

You can resume a stopped app and manually assign resources from either **Settings** > **Applications** or **Market** > **My Olares**.

1. Find the target app and click **Resume**.
2. Review the app's requirements at the top of the launch dialog.

   The dialog shows the required resource type, whether the app is **Single-GPU** or **Multi-GPU**, and the minimum VRAM required to run the app.

3. Under **Select GPUs**, choose a resource. If the app cannot be launched, check the message in the launch dialog:
   - If the available quota is too low, expand **Assigned apps** under the GPU, then click **Remove** for an assigned app to free resources.
   - If the GPU's total dedicated VRAM is too low, select another GPU. Removing apps will not increase its total capacity.
   - In **Time slicing** mode, the node memory check may also block launch. Retry later or select another node.

   ![Resume an app](/images/manual/olares/settings-gpu-resume-app.png#bordered){width=70%}

4. If you select a **Memory slicing** GPU, **Quota for current app** defaults to the app's minimum required VRAM. To change it, enter a new amount directly in the quota field.

   :::info Reallocate VRAM
   You cannot adjust VRAM after the app is assigned. To change it later, remove the app from the GPU and resume it again.
   :::

5. Click **Launch**.

## Remove an app from a GPU

Removing an app from a GPU stops the app and releases the GPU resources it was using. Use this workflow when you want to free resources from **Settings**, even if you are not launching another app right now.

To use the app again later, assign resources and launch it.

:::warning
If an app is bound to multiple GPUs or GPUs across multiple nodes, removing the app from one GPU releases all GPU resources assigned to that app and stops the app.
:::

1. Go to **Settings** > **Accelerator**, then click **Manage** on the resource card.
2. Under **Assigned apps**, find the target app, then click **Remove**.

   ![Remove an app](/images/manual/olares/settings-gpu-remove-app.png#bordered)

3. In the confirmation dialog, click **Remove**.

:::info
Resource release may take a short time. If no resource can be assigned to an app immediately after removal, refresh the resource list and check again.
:::

## Resources

- [Manage accelerator resources](/manual/olares/settings/gpu-resource.md): Learn how accelerator allocation works in Olares.
