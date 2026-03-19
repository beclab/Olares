---
outline: [2, 3]
description: Troubleshoot GPU-dependent apps that remain stopped after installation or resume in Memory slicing mode.
---
# GPU-dependent app remains stopped after installation or resume

Use this guide when a GPU-dependent app installs successfully but remains stopped, or remains stopped after you click **Resume**, in **Memory slicing** mode.

![App stopped after installation](/images/manual/help/ts-mem-slice-vram-app-stopped.png#bordered){width=85%}

## Condition

When the GPU mode is set to **Memory slicing** and you encounter either of the following:
- After installation, a GPU-dependent app remains in the **Stopped** state.
- After you click **Resume**, a GPU-dependent app remains in the **Stopped** state.

## Cause

Starting from Olares 1.12.5, in **Memory slicing** mode, the system automatically attempts to allocate VRAM to a GPU-dependent app when the app is installed or resumed. In earlier versions, shared GPU apps were added manually.

If most VRAM has already been allocated to other apps, the system cannot provide enough VRAM for the target app to run, so the app cannot initialize and remains in a **Stopped** state.

## Solution: Free up VRAM

### Step 1: Check the app's required VRAM

1. Go to **Market** > **My Olares**.
2. Click the card of the target app.
3. In the app details page, check the app's VRAM requirement and note it down.

![Check required VRAM for the target app](/images/manual/help/ts-mem-slice-vram-app-gpu.png#bordered){width=85%}

### Step 2: Check current VRAM availability

1. Go to **Settings** > **GPU**.
2. In the **Allocate VRAM** section, check how much VRAM has already been allocated to apps in the list.
3. Subtract the total allocated VRAM from your GPU's total VRAM capacity.

![Check current VRAM allocation](/images/manual/help/ts-mem-slice-vram-gpu-mode.png#bordered){width=90%}

In the example above, 22 GB of VRAM is currently allocated, leaving only 2 GB available, which is less than the 4 GB required by the target app.

### Step 3: Increase available VRAM

If there is not enough available VRAM, do one of the following:

#### Option A: Reduce VRAM allocated to another app

1. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">edit_square</i> next to an app's VRAM value.
2. Reduce the allocation without going below the VRAM required by that app, then click **Confirm**.

![Reduce VRAM allocation](/images/manual/help/ts-mem-slice-vram-reduce-vram.png#bordered){width=90%}

#### Option B: Release VRAM from unused apps

1. Stop an app that is not currently needed.
   - In **Market** > **My Olares**, open the dropdown menu and click **Stop**.
   - Or go to **Settings** > **Applications**, select the app, and click **Stop**.
2. Return to **Settings** > **GPU**.
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i>.
4. Click **Confirm** to remove the app's VRAM allocation completely.
5. (Optional) Repeat these steps until enough VRAM is available.

### Step 4: Resume the target app

1. Go back to the target app and click **Resume**.
   - In **Market** > **My Olares**, open the dropdown menu and click **Resume**.
   - Or go to **Settings** > **Applications**, select the app, and click **Resume**.
2. Return to **Settings** > **GPU**.
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">sync</i> to refresh the app list and verify the app status.

![Resume the target app after freeing VRAM](/images/manual/help/ts-mem-slice-vram-resume-outcome.png#bordered){width=90%}