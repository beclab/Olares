---
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

<!--@include: ./reusables.md{41,47}-->

## System compatibility
Make sure your Mac meets the following requirements.
- Architecture: X86-64 or ARM64
- RAM: 8 GB or above (available memory)
- Storage: 90 GB or above (available disk space)
- MacOS: Monterey (12) or later

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

<!--@include: ./reusables.md{4,33}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{35,39}-->