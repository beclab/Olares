---
outline: [2,3]
description: Learn how to install Windows on a secondary SSD to create a dual-boot system on Olares One.
head:
  - - meta
    - name: keywords
      content: Dual-boot, Windows, NVMe SSD, BIOS, Windows installation
---

# Dual-boot Windows on a secondary SSD <Badge type="tip" text="30 min" />

For competitive gaming or Windows-exclusive software, configure a secondary NVMe SSD to create a dual-boot system.

This dual-drive setup physically isolates the operating systems. This ensures Olares OS remains stable and secure while providing full native performance for your Windows applications.

## Learning objectives

By the end of this guide, you will learn how to:
- Create a bootable Windows installation USB drive.
- Configure BIOS settings to isolate Olares OS during installation.
- Install Windows on a secondary SSD.
- Set up GRUB to detect and switch between both operating systems.

## Prerequisites

**Hardware**<br>
- A secondary NVMe M.2 SSD physically installed in Olares One.
- A USB flash drive (8 GB or larger) for Windows installation media.
- A wired keyboard and mouse connected to Olares One.
- A monitor connected to Olares One.
- An Ethernet cable connected to Olares One.
- A Microsoft account for completing the Windows setup process.

## Step 1: Create a bootable Windows USB drive

1. Download the Windows 11 ISO from the [official Microsoft website](https://www.microsoft.com/en-us/software-download/windows11).
2. Download and install [**balenaEtcher**](https://etcher.balena.io/).
3. Insert the USB flash drive into your computer.
4. Open balenaEtcher and follow these steps:

   a. Click **Flash from file** and select the ISO you downloaded.

   b. Click **Select target** and select your USB drive.

   c. Click **Flash!** to write the installer to the USB drive.

   ![balenaEtcher flashing screen](/images/one/balenaEtcher.png#bordered)

5. Wait for the flashing and validation to finish, and then safely eject the USB drive.

## Step 2: Boot into BIOS and disable Olares OS

Disable the Olares OS drive in the BIOS before you install Windows to ensure the dual-boot configuration works correctly later.

1. Insert the Windows USB boot drive into the USB port on Olares One.
2. Power on Olares One, or restart it if it is already running.
3. When the Olares logo appears, press the **Delete** key repeatedly to enter the BIOS setup.

   ![BIOS setup](/images/one/bios-setup-interface.png#bordered)

4. Go to the **Boot** tab, select the boot option for the Olares OS disk, and then press **Enter**.
5. Select **Disabled** in the popup window, and then press **Enter**.
6. Go to the **Save and Exit** tab, under **Boot Override**, select your Windows USB disk to boot directly from it, and then press **Enter**.
7. Press **F10**, and then select **Yes** to save the changes and exit.
   
   The system automatically boots from the USB drive into the Windows installation interface.

## Step 3: Install Windows

The Windows setup process guides you through configuring your new system.

1. On the **Ventoy** screen, select the ISO file such as **Win11_25H2_English_x64_v2.iso**, and then select **Boot in normal mode**. The **Windows 11 Setup** wizard launches.
2. Follow the on-screen prompts to finish the standard Windows setup. During this process, the system restarts several times.

   :::danger Verify the selected partition
   On the **Select location to install Windows 11** screen, ensure you choose the correct secondary drive or partition. Selecting the wrong partition permanently erases your Olares OS.
   :::

3. After the final configuration process finishes, the Windows desktop appears, indicating a successful installation.
4. Unplug the Windows USB drive.

## Step 4: Re-enable Olares OS in BIOS

After installing Windows, re-enable the Olares OS drive in the BIOS and set it as the primary boot device. You must boot back into Olares OS so you can configure the GRUB dual-boot menu in the next step.

1. Restart Olares One.
2. When the Olares logo appears, press the **Delete** key repeatedly to enter BIOS setup.
3. Go to the **Boot** tab.
4. Set **Boot Option #1** to the SSD that contains Olares OS, and set **Boot Option #2** to the SSD that contains Windows.

   ![BIOS boot option priorities](/images/one/bios-boot-option-priorities.png#bordered)

5. Press **F10**, and then select **Yes** to save and exit BIOS. Olares One boots into Olares OS.

## Step 5: Detect Windows and update GRUB

To choose between operating systems at startup, configure the GRUB bootloader to detect Windows.

<Tabs>
<template #Olares-not-activated>

1. Log in using the default credentials:
    * **Username**: `olares`
    * **Password**: `olares`
   
   ![Log in to Olares One](/images/one/one-terminal.png#bordered)
2. Run the following command:

   ```bash
   sudo os-prober
   ```

   If Windows has been installed successfully, you should see an entry similar to:

   ```bash
   /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi:Windows Boot Manager:Windows:efi
   ```

3. Enable GRUB to probe other operating systems and regenerate the boot menu:

   a. Create a symbolic link for GRUB configuration:
      ```bash
      sudo ln -s /boot/efi/grub /boot/grub
      ```

   b. Enable OS prober to detect Windows:
      ```bash
      sudo sed -i 's|GRUB_DISABLE_OS_PROBER=true|GRUB_DISABLE_OS_PROBER=false|' /etc/default/grub
      ```

   c. Regenerate the GRUB boot menu:
      ```bash
      sudo update-grub
      ```

   Example output:

   ```bash
   Sourcing file '/etc/default/grub'
   Generating grub configuration file ...
   Warning: os-prober will be executed to detect other bootable partitions.
   Its output will be used to detect bootable binaries on them and create new boot entries.
   Found Windows Boot Manager on /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi
   Adding boot menu entry for UEFI Firmware Settings ...
   done
   ```
</template>
<template #Olares-already-activated>

1. Obtain the system password from LarePass.

   :::info
   Right after you activate Olares, you will be prompted to reset the SSH password on the LarePass app. The password is automatically generated and saved to your Vault.
   :::

   a. Tap **Vault** in the LarePass app. When prompted, enter your local password to unlock.

   b. In the top-left corner, tap **Authenticator** to open the side navigation, and then tap **All vaults** to display all saved items.

      ![Switch Vault filter](/images/one/ssh-switch-filter.png#bordered)

   c. Find the item with the <span class="material-symbols-outlined">terminal</span> icon and tap it to reveal the password.

      ![Check saved SSH password in Vault](/images/one/ssh-check-password-in-vault.png#bordered)
   
   d. Note down the password.


2. Log in using the default credentials:
    * **Username**: `olares`
    * **Password**: The one you obtained in Step 1.
   
   ![Log in to Olares One](/images/one/one-terminal.png#bordered)
3. Run the following command:

   ```bash
   sudo os-prober
   ```

   If Windows has been installed successfully, you should see an entry similar to:

   ```bash
   /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi:Windows Boot Manager:Windows:efi
   ```

4. Enable GRUB to probe other operating systems and regenerate the boot menu:

   a. Create a symbolic link for GRUB configuration:
      ```bash
      sudo ln -s /boot/efi/grub /boot/grub
      ```

   b. Enable OS prober to detect Windows:
      ```bash
      sudo sed -i 's|GRUB_DISABLE_OS_PROBER=true|GRUB_DISABLE_OS_PROBER=false|' /etc/default/grub
      ```

   c. Regenerate the GRUB boot menu:
      ```bash
      sudo update-grub
      ```

   Example output:

   ```bash
   Sourcing file '/etc/default/grub'
   Generating grub configuration file ...
   Warning: os-prober will be executed to detect other bootable partitions.
   Its output will be used to detect bootable binaries on them and create new boot entries.
   Found Windows Boot Manager on /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi
   Adding boot menu entry for UEFI Firmware Settings ...
   done
   ```
</template>
</Tabs>

## Step 6: Switch between operating systems

Use the GRUB boot menu to choose your preferred operating system every time you start Olares One.

1. Shut down Olares One, wait a few seconds, and then power it on again. 

   The GNU GRUB dual-boot menu appears automatically.

   ![Switch systems at startup](/images/one/one-dual-boot.png#bordered)

2. Choose which operating system to boot. The system automatically executes the highlighted entry in 10 seconds.

   - **Boot Olares OS**: Select **Olares GNU/Linux**.
   - **Boot Windows**: Select **Windows Boot Manager**.

3. To switch to the other operating system while currently logged in:

   - **From Olares OS to Windows**: Run `sudo reboot` in the terminal and enter your password when prompted. When the GNU GRUB menu appears, select **Windows Boot Manager**.

      :::info
      When you type your password in the terminal, the characters remain invisible for security. Ensure you have entered the correct password and press **Enter**.
      :::

   - **From Windows to Olares OS**: Restart Windows. When the GNU GRUB menu appears, select **Olares GNU/Linux**.

## Resources

- [Install drivers on Windows](install-nvidia-driver.md)
- [Dual-boot Ubuntu on a secondary SSD](dual-boot-ubuntu-dual-drive.md)
