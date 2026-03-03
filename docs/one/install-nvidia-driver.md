---
outline: [2, 3]
description: Learn how to download and use the all-in-one driver package provided by Olares to install all necessary Windows drivers, including the NVIDIA graphics driver.
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, Windows drivers, NVIDIA graphics driver, GPU
---

# Install drivers on Windows <Badge type="tip" text="15 min"/>

To ensure optimal performance and system stability when running Windows on your Olares One, Olares provides a tested all-in-one driver package. This package contains all the necessary drivers for your device, such as audio, network, and NVIDIA graphics drivers.

This guide walks you through downloading and installing the all-in-one driver package.

## Learning objectives

- Download and install all system drivers using the driver package.

## Before you begin

:::info About NVIDIA graphics driver update
This all-in-one package includes a stable, fully tested NVIDIA graphics driver specifically optimized for Olares One. To ensure system stability, avoid downloading and installing standalone graphics drivers directly from the official NVIDIA website, because that might introduce compatibility issues.
:::

- **Administrator access**: You need admin rights to install system drivers.
- **Internet connection**: The driver file is large (about 3.5 GB), so a stable connection is required.
- **Close applications**: Save your work and close any graphic-intensive programs like games or photo editors before starting the installation. The screen might temporarily turn black or blink during the update, which can cause open applications to crash or lose unsaved data.
- **Windows version**: For the best stability and driver compatibility, Windows 11 version 24H2 is recommended.

## Step 1: Download and extract the driver package

1. Download the [driver package](http://cdn.olares.com/common/AGBOX_B1_DVT_driver251125.zip).

    :::tip Browser security warning
    Because the download links use a standard HTTP connection rather than HTTPS, your web browser might flag or block it as an insecure download. If this happens, select **Keep** or **Allow** in your browser's download manager to continue the download.
    :::

    :::info Download time
    The driver package is large. Depending on your internet connection speed, the download might take a while to complete.
    :::

2. Locate the downloaded `.zip` file on your computer.
3. Right-click the file and select **Extract All**.
4. Open the extracted folder and verify that the files, including the `driver_install` application, are present.

    ```text
    Extracted_Folder/
    ├── driver/
    ├── uwp/
    ├── driver_install
    └── uwp_install
    ```

## Step 2: Install the drivers

1. In the extracted folder, locate the file `driver_install`, and then double-click it.
2. The system will automatically open a command prompt window and begin installing the driver. This process is fully automated.

    ![Install driver](/images/one/driver-install.png#bordered)

    :::tip Screen flickering
    Your screen might turn black or blink for a few seconds during installation. This is normal behavior as the system switches to the new driver.
    :::

3. When the driver is successfully installed, your device restarts automatically. 
4. After the system reboots and you log back in to Windows, a command prompt window appears. Follow the on-screen instructions and press any key until the window closes. This indicates that the installation is completely finished.

    ![System restart](/images/one/system-restart-windows.png#bordered)    

<!--## (Optional) Install the NVIDIA graphics driver

If you need the latest features, bug fixes, or specific optimizations, you can download and install the newest official graphics driver directly from NVIDIA.

### Step 1: Download the driver

1. Go to the official [NVIDIA Drivers Download](https://www.nvidia.com/en-us/drivers/) page.

    ![Nvidia drivers download page](/images/one/nvidia-drivers-download-page.png#bordered)

2. In the search field, enter `GeForce RTX 5090`. A list of matching products are displayed.

    ![Nvidia drivers search](/images/one/nvidia-driver-search.png#bordered)

3. Select the driver that matches your Windows version. For example, Windows 11.

    ![Nvidia driver select](/images/one/nvidia-win11.png#bordered)

4. Click **Find**.

    ![Nvidia find](/images/one/nvidia-find.png#bordered)

5. Review the results. You will typically see two types of drivers. Click **View** on the one that best fits your needs. For example, **NVIDIA Studio Driver**.

    - **GeForce Game Ready Driver**: Select this if you are a gamer who wants day-one support for the latest games, patches, and DLCs.
    - **NVIDIA Studio Driver**: Select this if you are a content creator (video editing, animation, design) who prioritizes stability and reliability for creative workflows.

    ![Nvidia view](/images/one/nvidia-view.png#bordered)

6. Click **Download**. The installer file (.`exe`) is automatically saved to your computer.

    ![Nvidia download](/images/one/nvidia-download.png#bordered)

### Step 2: Install the driver

1. Double-click the downloaded `.exe` file to start the installer.

    ![Launch installer](/images/one/launch-installer.png#bordered){width=65%}

2. **System Check**: 

    a. Select **NVIDIA Graphics Driver**. NVIDIA App is optional. 
    
    b. Click **AGREE AND CONTINUE**.

    ![Nvidia installer](/images/one/installer-step1.png#bordered){width=65%}

3. **Installation options**: 

    a. Select **Express (Recommended)**. This option automatically upgrades existing drivers and installs the new one with default settings. 
    
    b. Click **NEXT**. The installation begins.

    ![Installation options](/images/one/installer-step2.png#bordered){width=65%}

4. Wait for the installation to complete.

    :::tip Screen flickering
    Your screen might turn black or blink for a few seconds during installation. This is normal behavior as the system switches to the new driver.
    :::

    ![Installing](/images/one/install-wait.png#bordered){width=65%}    

5. When the installer window displays **NVIDIA Installer has finished**, click **CLOSE**.

    ![Installation finish](/images/one/nvidia-install-finish.png#bordered){width=65%}

6. Restart Windows to complete the setup.-->
