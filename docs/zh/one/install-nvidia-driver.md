---
outline: [2, 3]
description: Learn how to identify your NVIDIA GPU model and install the correct driver for optimal Windows performance.
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, NVIDIA drivers, GPU, graphics card，Windows
---

# Install NVIDIA drivers on Windows <Badge type="tip" text="10 min"/>

Installing the official drivers ensures the best graphics performance and stability when running Windows in your dual-boot setup.

This guide walks you through identifying your specific graphics card model and installing the correct driver.

## Learning objectives

- Identify the model of your NVIDIA graphics card.
- Download the correct driver from the official NVIDIA website.
- Install the graphics driver.

## Before you begin

- Connect to internet: The driver file is large (often 500 MB+), so a stable connection is required.
- Close applications: Save your work and close any graphic-intensive programs like games or photo editors before starting the installation. The screen might temporarily turn black or blink during the update, which can cause open applications to crash or lose unsaved data.

## Prerequisites

- Administrator access: You need admin rights to install system drivers.
- Windows OS: You must be currently booted into the Windows operating system.

## Step 1: Identify your graphics card model

Before downloading a driver, you need to know exactly which graphics card model you have. Olares One features an NVIDIA-based graphics card, so you can identify the model directly in the device settings.

1. Open Olares Settings.
2. Click **GPU** and you can find the specific model name "GeForce RTX 5090".

    ![Find GPU model](/images/one/find-gpu-model.png#bordered){width=70%}

## Step 2: Download the driver

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

## Step 3: Install the driver

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

6. Restart your computer to complete the setup.

## Next steps

Now your Windows environment is ready, you can proceed to configure your system for the Olares installation.

## Resources

- [Install Olares](../manual/get-started/install-olares.md)
- [Olares installation and activation FAQs](../manual/help/installation.md)
