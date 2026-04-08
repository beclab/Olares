---
outline: [2, 3]
description: Learn how to manually update the Embedded Controller (EC) and BIOS firmware on your Olares One using a USB drive.
head:
  - - meta
    - name: keywords
      content: Olares One, firmware update, Embedded Controller (EC), BIOS
---

# Update firmware

To ensure system stability and access the latest hardware features, you might occasionally need to update the Embedded Controller (EC) and BIOS firmware on your Olares One device.

## Before you begin

Review the following changelogs for features or fixes included in each update. Download the required update packages to your local computer before starting the update process.

:::info
To ensure system stability and security, only the latest firmware versions are available for download. Previous changelogs are provided for your reference only.
:::

### EC versions

| Version | Release date | Changelog |
|:--------|:-------------|:----------|
| [1.02 (Download)](http://cdn.olares.com/common/update_ec_1.02_tool.zip) | 2026-01-19 | <ul><li>Fix the issue where the keyboard fails to wake the system from sleep mode.</li></ul> |
| 1.01 | 2026-01-13 | <ul><li>Add support for Wake-on-LAN (WOL).</li><li>Disable the white breathing LED indicator during S3 sleep state.</li></ul> |
| 1.00 | 2025-12-01 | <ul><li>Initial MP SMT release.</li><li>Enable the white breathing LED indicator during S3 sleep state.</li></ul> |

### BIOS versions

| Version | Release date | Changelog |
|:--------|:-------------|:----------|
| [1.01 (Download)]() pending cdn address | 2025-12-04 | <ul><li>Fix the issue where SSDs unexpectedly disconnect by disabling ASPM and L-state power management for SSD1 and SSD2.</li></ul> |
| 1.00 | 2025-11-28 | <ul><li>Update version naming convention.</li></ul> |
| C400 | 2025-11-05 | <ul><li>Hide advanced BIOS options by default.</li><li>Remove MCU version display.</li><li>Fix the issue where memory tests report errors by enabling SAGV.</li></ul> |

## Prerequisites

- A USB flash drive formatted to **FAT32**.
- A monitor and a USB keyboard connected to your Olares One.
- The EC or BIOS update packages were downloaded.

## Update the EC firmware

1. Extract the downloaded EC update package.
2. Copy the resulting folders (e.g. `AGBOX4_EC_01_02` and `EFI`) to the root directory of your USB drive.
3. Connect the USB drive to your Olares One.
4. Power on the device or restart it if it is already running.
5. When the Olares logo appears, immediately press and hold the **F7** key to enter the boot menu.
6. Select your USB drive from the list, and then press **Enter**.

    ![Select USB boot device](/images/one/select-usb-boot.png#bordered)

7. When the EFI startup countdown screen appears (`Press ESC in 5 seconds to skip startup.nsh`), immediately press **Enter** to access the command shell.

    ![UEFI shell startup](/images/one/uefi-shell-startup.png#bordered)

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

## Update the BIOS

:::warning Important
Do not disconnect the power supply or turn off the device during the BIOS update process. Doing so might permanently damage the system.
:::

1. Extract the downloaded BIOS update package.
2. Copy the resulting folder (e.g., `AGBOX4_BIOS_103`) to the root directory of your USB drive.
3. Connect the USB drive to your Olares One.
4. Power on the device or restart it if it is already running.
5. When the Olares logo appears, immediately press and hold the **F7** key to enter the boot menu.
6. Select your USB drive from the list, and then press **Enter**.

    ![Select USB boot device](/images/one/select-usb-boot.png#bordered)

7. When the EFI startup countdown screen appears (`Press ESC in 5 seconds to skip startup.nsh`), immediately press **Enter** to access the command shell.

    ![UEFI shell startup](/images/one/uefi-shell-startup.png#bordered)

8. Run the following commands one by one to navigate to the AFU directory and start the flash script:

    ```bash
    cd AGBOX4_BIOS_103
    cd AFU
    FlashAFU.nsh
    ```
    ![Run BIOS flash script](/images/one/bios-flash-commands.png#bordered)

9. Wait for the script execution to finish.

    The system will automatically reboot and display a blue **Flash Update** progress screen.

    ![BIOS flash progress screen](/images/one/bios-update-progress.png#bordered)

10. Once the flash update reaches 100%, the **ME FW Update** starts automatically. Wait for this process to complete.

    ![BIOS flash progress screen - ME FW Update](/images/one/bios-update-progress-me.png#bordered)

11. When the **ME FW Update** finishes, the system will automatically reboot again. 

    :::info
    During the final reboot, the system will perform a comprehensive hardware self-test. This process takes approximately 2 to 3 minutes. The screen might remain black during this time. Wait until the normal `olares login` prompt appears.
    :::
