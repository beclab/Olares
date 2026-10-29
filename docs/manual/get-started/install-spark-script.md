---
outline: [2, 3]
description: Install Olares on NVIDIA DGX Spark using the installation script, stop and restart it, or uninstall it safely.
head:
  - - meta
    - name: keywords
      content: Olares, DGX Spark script, NVIDIA DGX Spark, command line install, GPU memory
---

# Install Olares on DGX Spark via script

This guide explains how to install Olares on NVIDIA DGX Spark using the command-line installation script. If you already installed Olares this way, you can also [stop and restart it](#stop-and-start-olares) or [uninstall it](#uninstall-olares).

## Prerequisites

- **Olares device**: NVIDIA DGX Spark with at least 150 GB of available SSD storage.
<!--@include: ./reusables.md#larepass-prerequisite-->
- **Initial setup**: DGX Spark has completed the [initial setup](https://docs.nvidia.com/dgx/dgx-spark/first-boot.html), including user account and network configuration.
- **Terminal access**: Use either method:
  - Direct access: Connect a monitor, keyboard, and mouse to DGX Spark.
  - Remote access: Connect via SSH from another computer on the same network.
- **Network**: An Ethernet connection to your router is recommended.

## Prepare DGX Spark

Before installing Olares, you need to remove the pre-installed container runtime on DGX Spark.

Olares installs and manages its own containerd runtime. If the system already has Docker or containerd installed, it will cause conflicts that prevent Olares from functioning properly.

:::warning Impact of removing container runtime
The following steps will:
- Uninstall Docker packages
- Stop and disable the existing containerd service
- Clear existing network rules
:::

On DGX Spark, open a terminal and run:

```bash
sudo apt remove docker*
sudo systemctl disable --now containerd
sudo rm -f /usr/bin/containerd
sudo nft flush ruleset
```

## Install Olares

1. In the same terminal, run the following command:

<!--@include: ./reusables.md#install-script-command-->

<!--@include: ./reusables.md#root-password-tip-->

<!--@include: ./reusables.md#installation-error-tip-->

<!--@include: ./reusables.md#prepare-wizard-heading-->

<!--@include: ./reusables.md#prepare-wizard-details-->

## Activate Olares

Use the Wizard URL and initial one-time password to activate. This process connects the Olares device with your Olares ID using LarePass.

1. Enter the Wizard URL in your browser. You will be directed to the welcome page. Press any key to continue.

   ![Open wizard](/images/manual/get-started/open-wizard.png#bordered)
2. Enter the one-time password and click **Continue**.

   ![Enter password](/images/manual/get-started/wizard-enter-password1.png#bordered)
3. Select the system language. Olares supports English, Simplified Chinese, German, Spanish, Italian, French, and Japanese. The Wizard switches to the selected language immediately, and Olares Desktop uses the same language after activation.

   ![Select language](/images/manual/get-started/select-language.png#bordered)
4. Select a reverse proxy node that is geographically closest to your location. You can adjust this later on the [Change reverse proxy](../olares/settings/change-frp.md) page.

   ![Select FRP](/images/manual/get-started/wizard-frp.png#bordered)

5. Activate Olares using LarePass app.

   :::warning Same network required for admin users
   To avoid activation failures, ensure that both your phone and the Olares device are connected to the same network.
   :::
   a. Open LarePass app, and tap **Scan QR code** to scan the QR code on the Wizard page and complete the activation.

   ![Activate Olares](/images/manual/get-started/activate-olares1.png#bordered)

   b. Reset the login password for Olares by following the on-screen instructions on LarePass.

After setup is complete, the LarePass app returns to the home screen, and the Wizard redirects you to the Olares login page.

<!--@include: ./log-in-to-olares.md-->

## Configure GPU memory for AI apps

DGX Spark features a unified memory architecture where the CPU and GPU share 128 GB of LPDDR5x memory. Unlike traditional GPUs with dedicated VRAM, DGX Spark does not distinguish between system memory and GPU memory.

On DGX Spark, Olares uses **Memory slicing** mode by default for GPU resource management. When you install an AI application, Olares automatically allocates the minimum required memory to ensure the app can start and run properly.

If needed, you can manually adjust the memory allocation for each AI application:

1. Open Settings from Olares, and then navigate to **GPU**.
2. In the **Allocate VRAM** section, find the target app.

    ![Memory slicing](/images/manual/get-started/install-spark-memory-slicing.png#bordered){width=70%}

3. Click <i class="material-symbols-outlined">edit_square</i> next to the VRAM value.
4. In the **Edit VRAM allocation** dialog, enter the desired VRAM amount in GB and click **Confirm**.

## Stop and start Olares

Stop Olares when you want to pause its services and apps while keeping your data. DGX Spark stays powered on, but Olares Desktop and its apps are unavailable until you start Olares again.

Run the following commands in a terminal on DGX Spark, either directly or through SSH.

1. Save your work and wait for running tasks, such as model training or file transfers, to finish.
2. Stop Olares:

   ```bash
   sudo olares-cli stop
   ```

   Wait for the command to finish and report `StopOlares execute successfully` without errors. Stopping does not delete installed apps, configuration, or user data, and does not disable automatic startup when DGX Spark reboots.

3. When you want to use Olares again, start it:

   ```bash
   sudo olares-cli start
   ```

   Wait for the command to finish, then open Olares Desktop and check that you can sign in and access your apps.

## Uninstall Olares

Uninstall Olares when you want to remove it from DGX Spark. The `--all` option also removes dependencies installed during the preparation phase, including the Olares container runtime.

:::warning Back up before uninstalling
Uninstalling deletes Olares configuration, installed apps, and user data. This also applies to `uninstall` without `--all`; omitting that option does not preserve your data.

While Olares is still running, [back up your files and supported app data](../olares/settings/backup.md) to external storage or cloud storage. Export other app data using each app's own backup tools, and check that your backups completed successfully. If Olares is stopped, start it first. To keep your installation and data, use `stop` instead.
:::

1. Open a terminal on DGX Spark, either directly or through SSH, and run:

   ```bash
   sudo olares-cli uninstall --all
   ```

2. Wait for the command to finish and report `Uninstall Olares execute successfully`. Check the output for errors before considering the uninstall complete.

Uninstalling does not restore Docker or the previous containerd setup removed in [Prepare DGX Spark](#prepare-dgx-spark). If you need them again, reinstall and configure them separately.

<!--@include: ./reusables.md#protect-olares-id-->

<!--@include: ./reusables.md#installation-troubleshooting-tip-->
