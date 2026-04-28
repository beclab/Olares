---
outline: [2, 3]
description: Troubleshoot the issue where your Olares One experiences a black screen, and you cannot access the BIOS setup menu.
---

# Cannot enter BIOS (Black screen)

Use this guide if your Olares One shows no display during startup, and you cannot access the BIOS setup screen.

## Condition

When you power on Olares One and attempt to enter the BIOS setup, the screen remains black, preventing the boot menu from displaying.

## Cause

This issue typically occurs if the `Primary Display` setting in the BIOS was changed from `Auto` (dedicated GPU mode) to `HG` (hybrid graphics mode).

When `Primary Display` is set to `HG`, the video output during the early boot stage might be sent to a different display interface (e.g., integrated graphics) that your monitor does not detect. The system might actually enter BIOS successfully, but your monitor screen stays black, making it impossible to see the BIOS interface.

## Solution

Follow these steps to blindly reset the BIOS to factory defaults. This restores the `Primary Display` setting to its default value `Auto` (the dedicated GPU mode).

:::warning
- These steps require physical removal of storage devices. Ensure that you power off the device and disconnect the power cable before proceeding.   
- The screen will remain black until Step 3. Use the keyboard Caps Lock indicator to confirm BIOS entry.
:::

### Step 1: Prepare the device

1. Power off Olares One.
2. Disconnect the power cable.
3. Disconnect all external storage drives, such as USB flash drives or external hard drives.
4. Open the device enclosure and temporarily remove the internal NVMe SSD.
5. Connect a keyboard and monitor to the device.
6. Reconnect the power cable.

### Step 2: Perform a blind reset

1. Power on Olares One, and then immediately press and hold the **Delete** key for about 20 seconds. 
2. Verify that the system is receiving keyboard inputs by pressing the **Caps Lock** key: If the Caps Lock indicator light on your keyboard turns on and off, you have successfully entered the BIOS menu.
3. Press the **F9** key, and then press **Enter**. This shortcut loads the factory default BIOS settings.
4. Press the **F10** key, and then press **Enter**. This shortcut saves the configuration and prompts the device to restart.

### Step 3: Verify the fix

1. Wait for the device to restart. The Olares One logo should appear, indicating that the normal display output is restored.
2. Power off the device and disconnect the power cable.
3. Reinstall your internal NVMe SSD and reconnect any external drives.
4. Reconnect the power cable and power on the device to resume normal BIOS operations.
