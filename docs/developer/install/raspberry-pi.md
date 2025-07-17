---
description: Complete guide to installing Olares on Raspberry Pi including hardware requirements, installation process, and system activation steps.
---
# Install Olares on Raspberry Pi
This guide explains how to install Olares on a Raspberry Pi.

:::warning Not recommended for production use
Currently, Olares on Raspberry Pi has certain limitations. We recommend using it only for development or testing purposes.
:::
<!--@include: ./reusables.md{41,47}-->

## System requirements
Make sure your Raspbian device meets the following requirements.
- Hardware: Raspberry Pi 4B or Raspberry Pi 5 with 8GB memory
- Operating system: Raspbian 12
- Storage: 64GB (SSD recommended)

## Set up system environment
1. Configure the Raspbian environment to enable necessary features:

      ```bash
      echo "$(head -1 /boot/firmware/cmdline.txt) cgroup_enable=cpuset cgroup_enable=memory cgroup_memory=1" | sudo tee /boot/firmware/cmdline.txt

      echo "kernel=kernel8.img" | sudo tee -a /boot/firmware/config.txt
      ```

2. Reboot your Raspbian device to apply the change.

   ```bash
   sudo reboot
   ```
## Install Olares
Run the following command:

<!--@include: ./reusables.md{4,33}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{35,39}-->