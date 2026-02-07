---
outline: [2, 3]
description: Learn about the three GPU modes in Olares and how to switch between them to match different workloads.
head:
  - - meta
    - name: keywords
      content: GPU, Time slicing, App exclusive, Memory slicing, GPU management
---
# Switch GPU mode <Badge text="10 min"/>

:::info
Only Olares admins can change GPU modes. This helps avoid conflicts and keeps GPU performance predictable for everyone.
:::

Olares lets you control how applications use GPU resources for workloads like AI, image and video generation, and gaming. You can choose different GPU modes depending on how your apps should share compute and memory.

## Learning objectives

By the end of this tutorial, you will learn how to:

- Choose between the three GPU modes.
- Switch GPU mode in Settings.
- Configure app binding for **App exclusive** mode and **Memory slicing** mode.

## Choose the right GPU mode

Olares provides three GPU modes, each designed for a different usage pattern.

| GPU mode                   | Definition                                                                                          | Use scenario                                                                      |
|:---------------------------|:----------------------------------------------------------------------------------------------------|:----------------------------------------------------------------------------------|
| **Time slicing** (Default) | Multiple apps share one GPU<br> by taking turns using compute<br> and VRAM.                         | General workloads that run several lightweight apps.                              |
| **App exclusive**          | One app gets full, uninterrupted<br> access to the compute and VRAM<br> of a single GPU.            | Heavy workloads that require maximum stability, such as LLMs and high‑end gaming. |
| **Memory slicing**         | The GPU's VRAM is divided into<br> fixed quotas, and apps run concurrently<br> within their limits. | Running specific apps simultaneously while strictly limiting their memory usage.  |

## View GPU details

Go to **Settings** > **GPU** to view your GPU status and current mode.

![GPU overview](/images/one/gpu-details.png#bordered){width=85%}

:::tip 
If you have more than one GPU, Olares lists all your GPUs on this page. Just click a GPU to open its details page directly. 
:::

## Switch GPU mode

Follow these steps to change how a GPU is used:

1. Go to **Settings** > **GPU**.
2. Choose a mode from the **GPU mode** dropdown.

:::warning Restart notice
Changing a GPU's mode restarts applications currently using that GPU.
:::

### Time slicing

**Time slicing** is the default mode in Olares. Use this mode to allow multiple applications to share resources.

When a GPU is in **Time slicing** mode, GPU-enabled applications are automatically assigned to it. These applications will appear in the app list on this page.

![Time slicing](/images/one/gpu-time-slicing.png#bordered){width=85%}

### App exclusive

Use **App exclusive** mode to dedicate a GPU entirely to one high-demand application.

![App exclusive](/images/one/gpu-app-exclusive.png#bordered){width=85%}

#### Bind app
To give an app exclusive access:
1. In **Select exclusive app** section, click **Bind app**.
2. Select your target application and click **Confirm**.

#### Switch app
To replace the current exclusive app with a new one:
1. In **Select exclusive app** section, click **Switch app**.
2. Choose the new application and click **Confirm**.

The previous app is unbound, and the new app takes over exclusive access.

#### Unbind app
To remove the exclusive binding:
1. In **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>.
2. In the pop-up window, click **Confirm** to unbind the app from the GPU.

### Memory slicing

Use **Memory slicing** to run apps concurrently with strict VRAM limits.

![Memory slicing](/images/one/gpu-mem-slicing.png#bordered){width=85%}

#### Bind app and allocate VRAM

1. In **Allocate VRAM** section, click **Bind app**.
2. Select your target application, assign it a specific amount of VRAM in GB, and click **Confirm**.
   :::warning
   The total of all VRAM limits must not exceed the GPU total VRAM.
   :::
3. Repeat for other apps as needed.

## Resources
- [Manage GPU usage](/manual/olares/settings/gpu-resource.md): For advanced operations when you have multiple GPUs.