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

### EC versions

| Version | Release date | Release notes | Download |
| :--- | :--- | :--- | :--- |
| **1.02** | YYYY-MM-DD | Optimized fan control logic<br>Improved power management | [Download package]() |
| **1.01** | YYYY-MM-DD | Initial system release | [Download package]() |

### BIOS versions

| Version | Release date | Release notes | Download |
| :--- | :--- | :--- | :--- |
| **1.02** | YYYY-MM-DD | Added support for [Feature]<br>Fixed an issue where [Bug] | [Download package]() |
| **1.01** | YYYY-MM-DD | Initial system release | [Download package]() |

## Prerequisites

- A USB flash drive formatted to **FAT32**.
- A monitor and a USB keyboard connected to your Olares One.
- The EC or BIOS update packages were downloaded.

## Update the EC firmware

1. Extract the downloaded EC update package.
2. Copy the resulting folders (e.g. `AGBOX4_EC_01_02` and `EFI`) to the root directory of your FAT32-formatted USB drive.
3. Connect the USB drive to your Olares One.
4. Power on the device or restart it if it is already running.
5. When the Olares logo appears, immediately press and hold the **F7** key to enter the boot menu.
6. Select your USB drive from the list (typically prefixed with `UEFI:`) and press **Enter**.

    ![Select USB boot device](/images/one/select-usb-boot.png#bordered)

7. When the EFI startup countdown screen appears (`Press ESC in 5 seconds to skip startup.nsh`), immediately press **Enter** to access the command shell.

    ![UEFI shell startup](/images/one/uefi-shell-startup.png#bordered)

8. Enter the following commands, and then press **Enter** to navigate to the EC directory:

    ```bash
    cd AGBOX4_EC_01_02
    ```

    ![Navigate to EC directory](/images/one/ec-flash-commands.png#bordered)

9. Enter the following command, and then press **Enter** to execute the update tool:

    ```bash
    ECFlashTool.efi AGBOX4_EC_01_02.bin
    ```

    ![Run EC flash tool](/images/one/ec-flash-commands.png#bordered)

    The system will display the progress as it erases and programs the flash memory. Wait for the update process to complete. 
    
    ![EC update progress](/images/one/ec-update-progress.png#bordered)

10. Once finished, restart the device.
11. When the Olares logo appears, immediately press and hold **F7** to enter the boot menu.
12. Select **Enter Setup** to access the BIOS.
13. On the **Main** tab, verify that the **EC FW Version** displays `1.02` (or your target version) to confirm the update was successful.

    ![Verify EC version in BIOS](/images/one/verify-ec-version.png#bordered)

## Update the BIOS

:::warning Important
Do not disconnect the power supply or turn off the device during the BIOS update process. Doing so might permanently damage the system.
:::

1. Extract the downloaded BIOS update package.
2. Copy the resulting folder (e.g., `AGBOX4_BIOS_103`) to the root directory of your FAT32 formatted USB drive.
3. Connect the USB drive to your Olares One.
4. Power on the device or restart it if it is already running.
5. When the Olares logo appears, immediately press and hold the **F7** key to enter the boot menu.
6. Select your USB drive from the list and press **Enter**.
7. When the EFI startup countdown screen appears, immediately press **Enter** to access the command shell.
8. Run the following commands one by one to navigate to the AFU directory and start the flash script:

    ```bash
    cd AGBOX4_BIOS_103
    cd AFU
    FlashAFU.nsh
    ```
    ![Run BIOS flash script](/images/one/bios-flash-commands.png#bordered)
9. Wait for the script execution to finish. The system will automatically reboot and display a blue **Flash Update** progress screen.
    ![BIOS flash progress screen](/images/one/bios-update-progress.png#bordered)
10. Once the flash update reaches 100%, the system will automatically reboot again. 

    :::info
    During this final reboot, the system will perform a comprehensive hardware self-test. This process takes approximately **2 to 3 minutes**. The screen might remain black temporarily. Wait until the normal OS login prompt appears.
    :::
