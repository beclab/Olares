---
outline: [2, 3]
description: Manage and optimize GPU resources in Olares with centralized controls, supporting time-slicing, exclusive access, and VRAM-slicing across single or multi-node setups.
---
# Manage GPU usage

:::info
Only Olares admins can change GPU modes. This helps avoid conflicts and keeps GPU performance predictable for everyone.
:::

Olares lets you manage your graphics cards, or GPUs, to speed up tasks like AI, image and video generation, and gaming. You can control how your applications use these resources from Olares **Settings** page.

This guide explains:
- How to choose the right GPU mode.
- How to configure GPU modes step by step.

## Before you begin

Before configuring GPUs, it helps to understand how Olares manages GPU bindings.

### Application state and binding

GPU binding depends on the state of an application.

- **To bind an app:** The application must be running. If an app is stopped, resume it before assigning it to a GPU.
- **To unbind an app:** An application cannot be removed from a GPU while it is running. Stop the application first, then unbind it.

:::info
Stopping an application does not automatically remove its GPU binding. 

A stopped app remains listed in the GPU app list until you explicitly unbind it.
:::

### Automatic binding

When switching GPU modes, Olares automatically binds currently running AI applications to GPU resources.

The binding behavior depends on the selected mode:

- **Time slicing:** Automatically binds all running AI apps so they can share available GPU resources.
- **Memory slicing:** Automatically binds all running AI apps and assigns each app the minimum VRAM required to run.
- **App exclusive:** Automatically selects one running AI app and binds it to a GPU for exclusive use.

## Choose the right GPU mode

Use the table below to choose a GPU mode based on how you want applications to share GPU resources.

| GPU mode | How GPU resources are shared | Use scenario |
| :--- | :--- | :--- |
| **Time slicing** (Default) | Multiple apps share the GPU by<br> taking turns using compute and VRAM. | Running several lightweight AI apps at the same time. |
| **App exclusive** | One app has full, uninterrupted<br> access to the GPU. | Heavy workloads that require maximum performance, such as LLMs or high-end gaming. |
| **Memory slicing** | VRAM is divided into fixed allocations<br> so multiple apps can run concurrently. | Running multiple AI apps while controlling how much GPU memory each app can use. |

## View GPU status

To see your GPUs and their current configuration:

1. Go to **Settings** > **GPU**.
2. Review the list to see each GPU's model, node, total VRAM, and current mode.
   ![GPU overview](/images/manual/olares/gpu-overview.png#bordered){width=90%}

3. Click a GPU to open its details page.  

:::tip
If you have only one GPU, Olares opens the GPU details page directly.
:::

## Configure GPU mode

Follow these steps to change how a GPU is used:

1. Go to **Settings** > **GPU**.
2. Click the GPU you want to configure.
3. Choose a mode from the **GPU mode** dropdown.

:::warning Restart notice 
Changing a GPU's mode will unbind apps from that GPU and restart their containers.

After restarting, Olares automatically binds currently running AI applications to available GPU resources.
:::

## Single-GPU setup

Use this section if your Olares device has only one GPU.

### Time slicing

![Time slicing](/images/manual/olares/gpu-time-slicing-single.png#bordered){width=90%}

#### Bind an app

To assign an app to this GPU:
1. Resume the stopped application.
   - Go to **Market** > **My Olares**, then choose **Resume** from the dropdown list.
   - Go to **Settings** > **Applications**, select the target app, then click **Resume**.
2. Go to **Settings** > **GPU**. In the **Pin applications** section, click **Bind app**.
3. Select the application and click **Confirm**.

:::info
The **Bind app** option appears only when there are running AI applications that are not already bound to the GPU.
:::

#### Unbind an app

To remove an app from this GPU:
1. Stop the application.
   - Go to **Market** > **My Olares**, then choose **Stop** from the dropdown list.
   - Go to **Settings** > **Applications**, select the target app, then click **Stop**.
2. Go back to **Settings** > **GPU**. In **Pin application** section, click <i class="material-symbols-outlined">link_off</i>, then **Confirm**.

### Memory slicing

![Memory slicing](/images/manual/olares/gpu-mem-slicing-single.png#bordered){width=90%}

#### Adjust VRAM allocation

1. In the **Allocate VRAM** section, find the target application.
2. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
3. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

:::warning
The total VRAM allocated to all applications cannot exceed the GPU's total VRAM.

If the value is lower than the application's minimum requirement, **Confirm** will be disabled.
:::

#### Unbind an app

To remove an app from this GPU:
1. Stop the application.
   - Go to **Market** > **My Olares**, then choose **Stop** from the dropdown list.
   - Go to **Settings** > **Applications**, select the target app, then click **Stop**.
2. Go back to **Settings** > **GPU**. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>, then **Confirm**.

### App exclusive

![App exclusive](/images/manual/olares/gpu-app-exclusive-single.png#bordered){width=90%}

#### Switch app

To replace the current exclusive app with a new one:
1. In the **Select exclusive app** section, click **Switch app**.
2. Choose the new application and click **Confirm**.

#### Unbind an app

To remove an app from this GPU:
1. Stop the application.
   - Go to **Market** > **My Olares**, then choose **Stop** from the dropdown list.
   - Go to **Settings** > **Applications**, select the target app, then click **Stop**.
2. Go back to **Settings** > **GPU**. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>, then **Confirm**.

## Multi-GPU setup

Use this section if your system has multiple GPUs. 

:::info Multi-GPU scheduling

When multiple GPUs are available, Olares may distribute running applications across different GPUs.

After switching GPU modes, running applications may appear on different GPUs.

You can manage bindings using:

- **Switch GPU** to move an app to another GPU.
- **Bind an app** to add a running app to the GPU.
- **Unbind an app** to remove the app from a GPU.

Stop the application only if you are removing it from its last bound GPU.
:::

### Time slicing

![Time slicing multi GPU](/images/manual/olares/gpu-time-slicing-multi.png#bordered){width=90%}

#### Bind an app

If no app is bound to the target GPU:
1. In the **Pin applications** section, click **Bind app**.
2. Select the application and click **Confirm**.

#### Switch GPU

1. In the **Pin applications** section, find the app.
2. Click <i class="material-symbols-outlined">repeat</i>.
3. Select the target GPU and click **Confirm**.

The application continues running during this process.

#### Unbind an app

To remove an application from a GPU:
1. If the app is only bound to a single GPU, stop the application first.
   - Go to **Market** > **My Olares**, then choose **Stop** from the dropdown list.
   - Go to **Settings** > **Applications**, select the target app, then click **Stop**.
2. Go back to **Settings** > **GPU**. In **Pin application** section, click <i class="material-symbols-outlined">link_off</i>, then **Confirm**.

:::tip Unbind from multiple GPUs
If the application is still bound to other GPUs on the same node, you can remove it from the current GPU without stopping it.
:::

### Memory slicing

![Multi-GPU memory slicing](/images/manual/olares/gpu-mem-slicing-multi.png#bordered){width=90%}

#### Adjust VRAM allocation

1. In the **Allocate VRAM** section, find the target application.
2. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
3. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

:::warning
The total VRAM allocated to all applications cannot exceed the GPU's total VRAM.

If the value is lower than the application's minimum requirement, **Confirm** will be disabled.
:::

#### Bind an app

If no app is bound to the target GPU:
1. In the **Allocate VRAM** section, click **Bind app**.
2. Select the application and click **Confirm**.

#### Switch GPU

To move an application to another GPU:
1. Find the app in the application list and click <i class="material-symbols-outlined">repeat</i> next to it.
2. Choose the target GPU.
3. Click **Confirm**.

#### Unbind an app

To remove an application from a GPU:
1. If the app is only bound to a single GPU, stop the application first.
   - Go to **Market** > **My Olares**, then choose **Stop** from the dropdown list.
   - Go to **Settings** > **Applications**, select the target app, then click **Stop**.
2. Go back to **Settings** > **GPU**. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>, then **Confirm**.

:::tip Unbind from multiple GPUs
If the application is still bound to other GPUs on the same node, you can remove it from the current GPU without stopping it.
:::

### App exclusive

![Multi-GPU app exclusive](/images/manual/olares/gpu-app-exclusive-multi.png#bordered){width=90%}

#### Bind an app

If no app is bound to the target GPU:
1. Click **Bind app** in the **Select exclusive app** section.
2. Choose the target app, then click **Confirm**.

#### Switch GPU

To move the exclusive app to another GPU:
1. Click <i class="material-symbols-outlined">repeat</i>.
2. Choose the target GPU.
3. Click **Confirm**.

#### Unbind an app

To remove an application from a GPU:
1. If the app is only bound to a single GPU, stop the application first.
   - Go to **Market** > **My Olares**, then choose **Stop** from the dropdown list.
   - Go to **Settings** > **Applications**, select the target app, then click **Stop**.
2. Go back to **Settings** > **GPU**. In the **Select exclusive app** section, click <i class="material-symbols-outlined">link_off</i>, then **Confirm**.

:::tip Unbind from multiple GPUs
If the application is still bound to other GPUs on the same node, you can remove it from the current GPU without stopping it.
:::

## Learn more
- [Monitor GPU usage in Olares](../resources-usage.md)