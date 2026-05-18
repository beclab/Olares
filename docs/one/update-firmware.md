---
outline: [2, 3]
description: Learn how to manage the BIOS on your Olares One, including checking firmware versions, downloading firmware update packages, executing updates, unlocking advanced mode, and troubleshooting issues such as black screen.
head:
  - - meta
    - name: keywords
      content: Olares One, firmware update, Embedded Controller (EC), BIOS
---

# Manage BIOS

This document explains how to manage the BIOS settings on your Olares One device, including how to check the current firmware versions, download and perform firmware updates, unlock advanced settings, and troubleshoot the black screen issue caused by display configurations.

## Check firmware versions

Before proceeding with any updates, check your current system firmware versions to determine if an update is necessary.

1. Power on your Olares One or restart it if it is already running.
2. When the Olares boot logo appears, immediately press and hold the **F7** key to enter the boot menu.
3. Select **Enter Setup** to access the BIOS.
4. On the **Main** tab, check the following:

    - **System BIOS Version**: This is the current BIOS version.
    - **EC FW Version**: This is the current Embedded Controller (EC) version.

    ![Check current firmware versions in BIOS](/images/one/check-firmware-versions-in-bios.png#bordered)

5. Press **ESC**, and then select **Yes** to exit the BIOS without saving.

## Download firmware updates

Review the following changelogs for features or fixes included in each update.

If your current versions are older than the ones listed below, download the corresponding update packages to proceed.

### BIOS versions

| Version | Release date | Changelog |
|:--------|:-------------|:----------|
| [1.01 (Download)](http://cdn.olares.com/common/OlaresOne_BIOS_1.01.zip) | 2025-12-04 | <ul><li>Fix the issue where SSDs unexpectedly disconnect by disabling ASPM and L-state power management for SSD1 and SSD2.</li></ul> |
| 1.00 | 2025-11-28 | <ul><li>Update version naming convention.</li></ul> |
| C400 | 2025-11-05 | <ul><li>Hide advanced BIOS options by default.</li><li>Remove MCU version display.</li><li>Fix the issue where memory tests report errors by enabling SAGV.</li></ul> |

### EC versions

| Version | Release date | Changelog |
|:--------|:-------------|:----------|
| [1.02 (Download)](http://cdn.olares.com/common/OlaresOne_EC_1.02.zip) | 2026-01-19 | <ul><li>Fix the issue where the keyboard fails to wake the system from sleep mode.</li></ul> |
| 1.01 | 2026-01-13 | <ul><li>Add support for Wake-on-LAN (WOL).</li><li>Disable the white breathing LED indicator during the sleep mode.</li></ul> |
| 1.00 | 2025-12-01 | <ul><li>Enable the white breathing LED indicator during the sleep mode.</li></ul> |
| C3.00 | 2025-11-25 | <ul><li>Fix the issue where the fan fails to spin after waking from the sleep mode.</li></ul> |

## Update firmware

Use the following instructions to manually flash the BIOS or EC firmware on your Olares One.

### Prerequisites

- A USB flash drive formatted to `FAT32`.
- A monitor and a USB keyboard connected to your Olares One.
- The EC or BIOS update package, downloaded to your computer.

### Update the BIOS

:::warning Important
Do not disconnect the power supply or turn off the device during the BIOS update process. Doing so might permanently damage the system.
:::

1. Extract the downloaded BIOS update package.
2. Copy the resulting folder (e.g., `AGBOX4_BIOS_101` and `EFI`) to the root directory of your USB drive.
3. Connect the USB drive to your Olares One.
4. Power on the device or restart it if it is already running.
5. When the Olares logo appears, immediately press and hold the **F7** key to enter the boot menu.
6. Select your USB drive from the list, and then press **Enter**.

    ![Select USB boot device](/images/one/select-usb-boot1.png#bordered)

7. When the EFI startup countdown screen appears (`Press ESC in 5 seconds to skip startup.nsh`), immediately press **Enter** to access the command shell.

    ![UEFI shell startup](/images/one/uefi-shell-startup.png#bordered)

8. Run the following commands one by one to navigate to the AFU directory and start the flash script:

    ```bash
    cd AGBOX4_BIOS_101
    cd AFU
    FlashAFU.nsh
    ```
    ![Run BIOS flash script](/images/one/bios-flash-commands-101.png#bordered)

9. Wait for the script execution to finish.

    The system will automatically reboot and display a blue **Flash Update** progress screen.

    ![BIOS flash progress screen](/images/one/bios-update-progress.png#bordered)

10. Once the flash update reaches 100%, the **ME FW Update** starts automatically. Wait for this process to complete.

    ![BIOS flash progress screen - ME FW Update](/images/one/bios-update-progress-me.png#bordered)

11. When the **ME FW Update** finishes, the system will automatically reboot two times. Wait until the normal `olares login` prompt appears.

    :::info
    During the reboot, the system will perform a comprehensive hardware self-test. This process takes approximately 2 to 3 minutes. The screen might remain black during this time.
    :::

12. Verify the BIOS version.

    a. Restart Olares One manually.
    
    b. When the Olares logo appears, immediately press and hold **F7** to enter the boot menu.

    c. Select **Enter Setup** to access the BIOS.

    ![Enter setup for BIOS](/images/one/enter-setup.png#bordered)  

    d. On the **Main** tab, verify that the **System BIOS Version** displays `1.01` (or your target version) to confirm the update was successful.

    ![Verify BIOS version](/images/one/enter-setup-bios1.png#bordered)

### Update the EC firmware

1. Extract the downloaded EC update package.
2. Copy the resulting folders (e.g., `AGBOX4_EC_01_02` and `EFI`) to the root directory of your USB drive.
3. Connect the USB drive to your Olares One.
4. Power on the device or restart it if it is already running.
5. When the Olares logo appears, immediately press and hold the **F7** key to enter the boot menu.
6. Select your USB drive from the list, and then press **Enter**.

    ![Select USB boot device](/images/one/select-usb-boot.png#bordered)

7. When the EFI startup countdown screen appears (`Press ESC in 5 seconds to skip startup.nsh`), immediately press **Enter** to access the command shell.

    ![UEFI shell startup](/images/one/uefi-shell-startup-ec.png#bordered)

8. Enter the following command, and then press **Enter** to navigate to the EC directory:

    ```bash
    cd AGBOX4_EC_01_02
    ```

    ![Navigate to EC directory](/images/one/ec-cd-command.png#bordered)

9. Enter the following command, and then press **Enter** to execute the update tool:

    ```bash
    ECFlashTool.efi AGBOX4_EC_01_02.bin
    ```

    ![Run EC flash tool](/images/one/ec-flash-command.png#bordered)

    The system will display the progress as it erases and programs the flash memory. Wait for the update process to complete. Once finished, the device restarts automatically. 
    
    ![EC update progress](/images/one/ec-update-progress.png#bordered)

10. When the Olares logo appears, immediately press and hold **F7** to enter the boot menu.
11. Select **Enter Setup** to access the BIOS.

    ![Enter setup for BIOS](/images/one/enter-setup-bios.png#bordered)

12. On the **Main** tab, verify that the **EC FW Version** displays `1.02` (or your target version) to confirm the update was successful.

    ![Verify EC version in BIOS](/images/one/verify-ec-version.png#bordered)

##  Unlock advanced settings

Some advanced BIOS options are hidden by default to prevent accidental configuration changes that could impact system stability.

If you need to perform deep hardware configuration, you can unlock the hidden advanced settings:
1. Access the BIOS, and then go to the **Advanced** tab.

    ![Default Advanced settings in BIOS](/images/one/bios-advanced-default.png#bordered)

2. Press Ctrl + Right Arrow on your keyboard. The hidden configuration options, such as **RC ACPI Settings** and **PCIE Configuration**, will appear on the screen.

    ![Full Advanced settings in BIOS](/images/one/bios-advanced-full.png#bordered)

    :::warning Do not change the Primary Display
    Ensure that the **Primary Display** setting remains set to **Discrete GPU**. 

    Do not change this setting to **HG** (Hybrid Graphic). Selecting **HG** will route the video output to an inactive interface, resulting in a completely black screen on startup.

    If you accidentally changed this setting and lose display output, refer to the [Troubleshooting](#cannot-enter-bios-black-screen) section to perform a blind reset.
    :::

## Troubleshooting

### Cannot enter BIOS (Black screen)

When you power on Olares One and attempt to enter BIOS, the monitor screen remains black, preventing the BIOS setup interface from displaying.

This issue typically occurs if the `Primary Display` setting in the BIOS was changed from `Discrete GPU` to `HG` (Hybrid Graphics). When `Primary Display` is set to `HG`, the video output during the early boot stage might be sent to a different display interface (e.g., integrated graphics) that your monitor does not detect. The system might actually enter BIOS successfully, but your monitor screen stays black, making it impossible to see the BIOS setup interface.

#### Solution

Follow these steps to blindly reset the BIOS to factory defaults. This restores the `Primary Display` setting to its default value.

:::warning
- These steps require physical removal of storage devices. Ensure that you power off the device and disconnect the power cable before proceeding.   
- The screen will remain black until Step 3. Use the keyboard Caps Lock indicator to confirm BIOS entry.
:::

##### Step 1: Prepare the device

1. Power off Olares One.
2. Disconnect the power cable.
3. Disconnect all external storage drives, such as USB flash drives or external hard drives.
4. Open the device enclosure and temporarily remove the internal NVMe SSD.
5. Connect a keyboard and monitor to the device.
6. Reconnect the power cable.

##### Step 2: Perform a blind reset

1. Power on Olares One, and then immediately press and hold the **Delete** key for about 20 seconds. 
2. Verify that the system is receiving keyboard inputs by pressing the **Caps Lock** key: If the Caps Lock indicator light on your keyboard turns on and off, you have successfully entered BIOS.
3. Press the **F9** key, and then press **Enter**. This shortcut loads the factory default BIOS settings.
4. Press the **F10** key, and then press **Enter**. This shortcut saves the configuration and prompts the device to restart.

##### Step 3: Verify the fix

1. Wait for the device to restart. The Olares One logo should appear, indicating that the normal display output is restored.
2. Power off the device and disconnect the power cable.
3. Reinstall your internal NVMe SSD and reconnect any external drives.
4. Reconnect the power cable and power on the device to resume normal BIOS operations.
