---
outline: [2, 3]
description: Troubleshoot GPU-dependent apps that remain stopped after installation or resume in Memory slicing mode.
---

# GPU-dependent app remains stopped after installation or resume

Use this guide when a GPU-dependent app remains in the **Stopped** state after installation or resume in **Memory slicing** mode.

## Condition

The GPU mode is set to **Memory slicing**, and you encounter either of the following:

- After installing a GPU-dependent app, it remains in the **Stopped** state.
- After clicking **Resume** on a GPU-dependent app, it remains in the **Stopped** state.

## Cause

In Olares 1.12.5, **Memory slicing** mode automatically allocates VRAM to a GPU-dependent app when it is installed or resumed. If most VRAM has already been allocated to other apps, the system cannot provide enough VRAM for the target app to run, so the app cannot initialize and remains in a **Stopped** state.

## Solution: Free up VRAM

### Step 1: Check the app's required VRAM

1. Go to **Market** > **My Olares**.
2. Click the card of the target app.
3. In the app details page, note the app's VRAM requirement.

![Check required VRAM for the target app](/images/manual/help/ts-mem-slice-vram-app-gpu.png#bordered){width=85%}

### Step 2: Check current VRAM availability

1. Go to **Settings** > **GPU**.
2. In the **Allocate VRAM** section, note the total VRAM currently allocated across all apps in the list.
3. Subtract this from your GPU's total VRAM capacity to get the available VRAM.

![Check current VRAM allocation](/images/manual/help/ts-mem-slice-vram-gpu-mode.png#bordered){width=90%}

In the example above, 22 GB of VRAM is currently allocated, leaving only 2 GB available, which is less than the 4 GB required by the target app.

### Step 3: Free up VRAM

Use one or both of the following approaches to free up enough VRAM.

#### Reduce the allocation of an active app

1. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">edit_square</i> next to an app's VRAM value.
2. Reduce the allocation without going below the VRAM required by that app, then click **Confirm**.

![Reduce VRAM allocation](/images/manual/help/ts-mem-slice-vram-reduce-vram.png#bordered){width=90%}

#### Stop unused apps and release their VRAM

1. Stop an app that is not currently needed using one of the following methods:
   - In **Market** > **My Olares**, click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the app's operation button and click **Stop**.
   - Or go to **Settings** > **Applications**, select the app, and click **Stop**.
2. Return to **Settings** > **GPU**.
3. In the **Allocate VRAM** section, click <i class="material-symbols-outlined">link_off</i> next to the stopped app, then click **Confirm**.
4. Repeat until enough VRAM is available.

### Step 4: Resume the target app

1. Resume the target app:
   - In **Market** > **My Olares**, click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the app's operation button and click **Resume**.
   - Or go to **Settings** > **Applications**, select the app, and click **Resume**.
2. Wait for the app status to change to **Running**.

![Resume the target app after freeing VRAM](/images/manual/help/ts-mem-slice-vram-resume-outcome.png#bordered){width=90%}