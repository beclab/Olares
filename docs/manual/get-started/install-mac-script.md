---
outline: [2, 3]
description: Step-by-step instructions for installing Olares on macOS systems including prerequisites, installation commands, and activation process.
---
# Install Olares on Mac via the script
This guide explains how to install Olares on macOS using the provided installation script.

:::warning Not for production use
Olares on Mac has certain limitations including:
- Lack of distributed storage support.
- Inability to add local nodes.

We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md#installation-troubleshooting-tip-->

## System requirements

- **CPU**: At least 4 cores.
- **RAM**: At least 8 GB of available memory.
- **Storage**: At least 150 GB of available storage.
- **macOS**: Monterey 12 or later.

## Before you begin
Ensure you have the following installed:
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [MiniKube](https://minikube.sigs.k8s.io/docs/start/?arch=%2Fmacos%2Farm64%2Fstable%2Fhomebrew)
    ::: tip
    It's recommended to install via `homebrew`.
    :::

## Set up system environment
1. In Docker Desktop, navigate to **Settings** > **Resources**, and configure as below:
    - **CPU limit**: Set to at least 4 CPUs
    - **Memory limit**: Set to at least 9 GB
    - **Virtual disk limit**: Set to at least 80 GB

   ![Update resource settings (example)](/images/manual/get-started/docker-resources-settings.png)
2. Click **Apply & restart** to implement the changes.
## Install Olares
In terminal, run the following command:

<!--@include: ./reusables.md#install-script-command-->

<!--@include: ./reusables.md#root-password-tip-->

<!--@include: ./reusables.md#installation-error-tip-->

<!--@include: ./reusables.md#prepare-wizard-heading-->

During the Wizard setup, provide the following information:
1. Check the IP address of your Mac, for example, `192.168.x.x`.

   If the automatically detected IP address is correct, press `Y` to confirm. To change it, press `R` and enter the correct address.
   ::: tip Find the IP address
   You can find the IP address of your Mac in either of the following ways:
   - GUI: Open **System Settings** (or **System Preferences**) > **Network**, then check the details of the active network connection.
   - Command line: Open Terminal and run `ipconfig getifaddr en0` for Wi-Fi, or `ipconfig getifaddr en1` for Ethernet.
   :::

2. <!--@include: ./reusables.md#prepare-wizard-details-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md#protect-olares-id-->
