---
outline: [2, 3]
description: Install Olares on NVIDIA DGX Spark using the command-line installation script for quick deployment.
---

# Install Olares on DGX Spark via script <Badge type="warning" text="RC" />

This guide explains how to install Olares on NVIDIA DGX Spark using the command-line installation script.

:::warning RC version
DGX Spark support is currently in Release Candidate (RC). We are actively testing and will release the stable version soon.
:::

<!--@include: ./reusables.md{44,51}-->

## System requirements

- **DGX Spark**: Ensure your device has completed the [initial setup](https://docs.nvidia.com/dgx/dgx-spark/first-boot.html), with a user account created and network configured.
- **Storage**: At least 150 GB of available SSD storage on DGX Spark.
  :::warning SSD required
  The installation will fail if an HDD (mechanical hard drive) is used instead of an SSD, or if insufficient storage is available.
  :::
- **Access method**: You need access to the terminal on DGX Spark, either via:
  - Direct access: Connect a monitor, keyboard, and mouse to DGX Spark.
  - Remote access: Connect via SSH from another computer on the same network.
- **Network**: An Ethernet cable connecting DGX Spark to your router (recommended for stable connection).

## Prepare DGX Spark

Before installing Olares, you need to remove the pre-installed container runtime on DGX Spark.

On DGX Spark, open a terminal and run:

```bash
sudo apt remove docker*
sudo systemctl disable --now containerd
sudo rm -f /usr/bin/containerd
sudo nft flush ruleset
```

## Install Olares

1. Open a terminal on DGX Spark.

2. Run the following command:

<!--@include: ./reusables.md{4,36}-->

## Activate Olares

Use the Wizard URL and initial one-time password to activate. This process connects the Olares device with your Olares ID using LarePass.

1. Enter the Wizard URL in your browser. You will be directed to the welcome page. Press any key to continue.

   ![Open wizard](/images/manual/get-started/open-wizard.png#bordered)
2. Enter the one-time password and click **Continue**.

   ![Enter password](/images/manual/get-started/wizard-enter-password.png#bordered)
3. Select the system language.

   ![Select language](/images/manual/get-started/select-language.png#bordered)
4. Select a reverse proxy node that is geographically closest to your location. You can adjust this later on the [Change reverse proxy](../olares/settings/change-frp.md) page.

   ![Select FRP](/images/manual/get-started/wizard-frp.png#bordered)

5. Activate Olares using LarePass app.

   a. Open LarePass app, and tap **Scan QR code** to scan the QR code on the Wizard page and complete the activation.
   :::warning Same network required for admin users
   To avoid activation failures, ensure that both your phone and the Olares device are connected to the same network.
   :::

   ![Activate Olares](/images/manual/get-started/activate-olares.png#bordered)

   b. Reset the login password for Olares by following the on-screen instructions on LarePass.

After setup is complete, the LarePass app returns to the home screen, and the Wizard redirects you to the Olares login page.

<!--@include: ./log-in-to-olares.md-->

## Configure GPU memory for AI apps

The system uses **Memory slicing** mode for GPU resource management. When you install an AI application, Olares automatically allocates the minimum required VRAM to ensure the app can start and run properly.

You can manually adjust the VRAM allocation for each AI application based on your specific needs:

1. Open Settings from Olares, and then navigate to **GPU**.
2. In the **Allocate VRAM** section, find the target app.

    ![Memory slicing](/images/manual/get-started/install-spark-memory-slicing.png#bordered){width=70%}

3. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
4. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

<!--@include: ./reusables.md{38,42}-->
