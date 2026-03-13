---
outline: [2,3]
description: Learn how GPU allocation works in Olares, including binding and unbinding, how app state affects GPU actions, and how time slicing, memory slicing, and app exclusive differ from one another.
---
# Understand GPU management

:::info
Only Olares admins can change GPU modes. This helps avoid conflicts and keeps GPU performance predictable for everyone.
:::

Olares lets you manage how apps use available GPUs for workloads such as AI, image and video generation, transcoding, and gaming.

In this guide, you will learn:
- How GPU allocation works in Olares.
- How app state affects available GPU actions.
- How GPU modes differ and when to use each one.

## How GPU allocation works

In Olares, giving an app access to GPU resources is called binding. Unbinding removes that access so the GPU can be released or reassigned.

Whether you can bind or unbind an app depends mainly on whether the app is running or stopped.

| App state | Bind (Give access) | Unbind (Remove access) |
| -- | -- | -- |
| **Running** | Supported | Not supported. You need to stop the app first. <sup>1, 2</sup> |
| **Stopped** | Not supported. You need to resume <br>the app first. | Supported |

1. Stopping an app pauses its workload, but it does not automatically remove its GPU allocation. To fully release the GPU or VRAM for other workloads, you must explicitly unbind the app after stopping it.
2. Multi-GPU exception: If an app is allocated to multiple GPUs on the same node, you can remove its access from one GPU while it remains running on the others.

You can check whether an app is running or stopped in either of these places:

- **Market** > **My Olares**: The current status is displayed on the app's card. 
- **Settings** > **Applications**: The current status is shown in the app list. 
- **Launchpad**: A stopped app is marked with an orange dot next to its name.

## GPU modes and when to use them

Olares supports three GPU modes. Each mode determines how GPU resources are shared and what happens to running apps after you switch modes.

:::info DGX Spark support
On DGX Spark, you can use **Memory slicing** and **App exclusive** to manage GPU resources.
:::

| GPU mode | How resources are<br> shared | After switching to this mode | Best for |
| -- | -- | -- | -- |
| **Time slicing** (Default) | Multiple apps share <br>the same GPU over <br>time. | Running apps that require a GPU are automatically assigned to share the GPU. | Running several GPU-dependent apps at the same time. |
| **Memory slicing** | Multiple apps share <br>the GPU, with fixed <br>VRAM allocations <br>for each app. | Running apps that require a GPU are automatically added and assigned the minimum VRAM required to run. | Running multiple GPU-dependent apps while strictly controlling VRAM usage. |
| **App exclusive** | One app gets full, <br>uninterrupted access<br> to the GPU. | One running app that requires a GPU is automatically selected and given exclusive access. | Heavy workloads that need maximum performance, such as large models, rendering, or high-end gaming. |

:::warning App interruption notice
Changing a GPU's mode reallocates hardware resources. Depending on the mode you choose, apps that are currently using the GPU may be paused automatically.

After switching modes, check the state of your apps and manually resume them if necessary.
:::

## Next steps

- [Manage GPU resources for a single GPU](./single-gpu.md)
- [Manage GPU resources for multiple GPUs](./multi-gpu.md)