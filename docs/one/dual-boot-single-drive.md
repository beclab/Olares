---
outline: [2,3]
description: Install Windows and Olares OS on a single SSD using a dual-boot configuration.
head:
  - - meta
    - name: keywords
      content: Dual-boot, Single SSD, Windows, Ubuntu, Olares OS, Partitioning
---

# Dual-boot Windows on a single SSD <Badge type="tip" text="45 min" />

Run both Windows and Olares OS on Olares One without adding a second drive.

This configuration installs Windows first, then sets up Ubuntu Linux alongside it, and finally deploys Olares OS.

## Prerequisites
**Hardware**<br>
- Olares One connected to Power.
- Wired keyboard and mouse connected to Olares One.
- (Recommended) An Ethernet cable connecting Olares One to your router.
- A USB flash drive contains Windows installation media.
- A USB flash drive contains Ubuntu Server or Desktop (24.04 LTS or newer) installation media.

**Network**<br>
- A reliable internet connection.
- Your mobile phone connected to the same network.

**Software**<br>
- [You have downloaded the LarePass app and created an Olares ID](first-boot.md#step-1-power-on-and-install-larepass).

:::danger Backup required
Partitioning a drive carries a risk of data loss. If you have important data on your current drive, backup everything to an external source before proceeding.
:::

## Step 1: Install Windows
1. Insert the Windows USB boot drive into a USB port on Olares One.
2. Power on Olares One or restart if it is already running.
3. Immediately press and hold the **Delete** key repeatedly until the BIOS setup screen appears.
4. Navigate to the **Boot** tab using the arrow keys on your keyboard.
5. Locate **Boot Option #1** or **Boot Override** and select your USB flash drive.
6. Press **Enter** to boot from the USB drive immediately. 
7. Follow the on-screen prompts to begin the Windows installation. 
8. When the installation finishes and the system restarts, unplug the Windows USB drive.

The system will boot into Windows automatically.

## Step 2: Create a partition for Ubuntu

1. Once Windows is running, right-click the **Start** button and select **Disk Management**.
2. Right-click your main `C:` partition and select **Shrink Volume**.
3. Enter the amount of space to free up for Olares. Minimum 150 GB is required.
    :::tip
   Avoid splitting the drive into two equal sizes. Using distinct sizes makes it easier to identify the correct partition later during the Ubuntu installation.
   :::
4. Click **Shrink** to create a block of unallocated space.

## Step 3: Install Ubuntu

Olares OS runs on top of a Linux kernel. You will install Ubuntu to serve as the host system.

1. Insert the Ubuntu USB drive and restart Olares One.
2. Press **Delete** repeatedly to enter BIOS, then select the USB drive as the boot device.
3. Follow the installer prompts until you reach the **Installation type** screen.
4. Select **Install Ubuntu alongside Windows Boot Manager**.
    :::tip
    If this option does not appear, select the manual installation option to manually assign the unallocated space to Ubuntu.
    :::
5. When the installation finishes and the system restarts, unplug the Ubuntu USB drive.

The system will boot into Ubuntu automatically.

## Step 4: Install Olares OS

1. Log in to Ubuntu using your username and password.
2. Open a terminal window if using Ubuntu Desktop or use the command line directly. 
3. Run the official installation script:
   ```bash
   curl -fsSL https://olares.sh | bash
   ```

4. At the end of the installation process, you will be prompted to enter your domain name and Olares ID.

    ![Enter domain name and Olares ID](/images/manual/get-started/enter-olares-id.png)

    For example, if your full Olares ID is `alice123@olares.com`:
    
    - **Domain name**: Press `Enter` to use the default domain name or type `olares.com`.
    - **Olares ID**: Enter the prefix of your Olares ID. In this example, enter `alice123`.

Upon completion of the installation, the initial system information, including the Wizard URL and the initial login password, will appear on the screen. You will need them later in the activation stage.
![Wizard URL](/images/manual/get-started/wizard-url-and-login-password.png)

## Step 5: Activate Olares OS

1. Enter the Wizard URL in the browser on Ubuntu, or use a browser on another computer connected to the same network.
   ![Open wizard](/images/manual/get-started/open-wizard.png#bordered)
2. Enter the one-time password and click **Continue**.

   ![Enter password](/images/manual/get-started/wizard-enter-password.png#bordered)
3. Select the system language.

   ![Select language](/images/manual/get-started/select-language.png#bordered)
<<<<<<< HEAD
4. Select a reverse proxy node that is geographically closest to your location.
=======
4. Select a reverse proxy node that is geographically closest to your location. You can adjust this later on the [Change reverse proxy](../olares/settings/change-frp.md) page.
>>>>>>> 5972b65b (add single-drive setup & update dual-drive setup)

   ![Select FRP](/images/manual/get-started/wizard-frp.png#bordered)

5. Activate Olares using LarePass app.

   a. Open LarePass app, and tap **Scan QR code** to scan the QR code on the Wizard page and complete the activation.

   ![Activate Olares](/images/manual/get-started/activate-olares.png#bordered)

   b. Reset the login password.


After setup is complete, the LarePass app returns to the home screen, and the browser redirects you to the Olares login page.

## Switch between operating systems

You switch between Windows and Olares using the BIOS boot priority.
### Switch to Olares OS
1. Restart Olares One.
2. Press **Delete** repeatedly to enter BIOS.
3. Go to the **Boot** tab.
4. Set **Boot Override** to Ubuntu.
5. Press **F4** to confirm.

### Switch to Windows
1. Restart Olares One.
2. Press **Delete** repeatedly to enter BIOS.
3. Go to the **Boot** tab.
4. Set **Boot Override** to Windows.
5. Press **F4** to confirm.

## Troubleshooting

### An error occurs during installation

If an error occurs during installation, use the following command to uninstall first:

```bash
olares-cli uninstall --all
```

After uninstalling, retry the installation by running the original installation command.

### The Ubuntu interface is lagging
When you first boot into the Ubuntu desktop, the interface might feel laggy. This is because the GPU drivers for Ubuntu are not yet installed.

Proceed with the Olares OS installation. The installation script automatically installs the necessary GPU drivers, which will resolve the lag.


## Resources
- [Install NVIDIA drivers on Windows](install-nvidia-driver.md)