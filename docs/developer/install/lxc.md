---
description: Instructions for installing Olares on Linux Containers (LXC) including container setup, system requirements, and activation steps.
---
# Install Olares on LXC
LXC (Linux Containers) is a lightweight virtualization method that runs applications in isolated containers. When used on PVE, it enables an efficient way to deploy Olares without the overhead of a full virtual machine.

:::warning Not recommended for production use
Currently, Olares on LXC has certain limitations. We recommend using it only for development or testing purposes.
:::

<!--@include: ./reusables.md{41,47}-->

## System requirements
Make sure your device meets the following requirements.

- CPU: At least 4 cores
- RAM: At least 8GB of available memory
- Storage: At least 64GB of available space (SSD recommended)
- Supported systems:
    - PVE 8.2.2
    - Linux container: Debian 12 (for existing LXC containers on PVE)

:::info Version compatibility
While the specific versions are confirmed to work, the process may still work on other versions. Adjustments may be necessary depending on your environment. If you meet any issues with these platforms, feel free to raise an issue on [GitHub](https://github.com/beclab/Olares/issues/new).
:::

## Prerequisites

-  Working directories for storing images and packages on the PVE host. You can set it using the following command:

   ``` bash
   mkdir -p /root/.olares/images /root/.olares/pkg
   ```
- The container template (CT) for `debian-12-standard_12.7-1_amd64.tar.zst`. Download it from the [PVE image repository](http://download.proxmox.com/images/system/).

## Configure the LXC environment

::: tip Install on existing LXC
To install Olares on an existing LXC container, skip to step 2 directly. Make sure you use the corresponding container ID.
:::

1. Create the LXC container using the following script:

   ::: tip Unique container ID
   To create a container, you need to assign it a unique container ID. In this guide, we use `16553`, but you can replace it with any available numeric ID. Make sure to update all commands and configurations accordingly.
   :::

   ``` bash{2}
   export ROOTPASS=123456 
   pct create 16553 /var/lib/vz/template/cache/debian-12-standard_12.7-1_amd64.tar.zst \
   --hostname olares \
   --ostype ubuntu \
   --cores 4 \
   --memory 10240 \
   --swap 0 \
   --net0 name=eth0,bridge=vmbr0,firewall=1,ip=dhcp,ip6=dhcp,type=veth \
   --rootfs local-lvm:80 \
   --unprivileged 0 \
   --ignore-unpack-errors \
   --mp0 "/root/.olares/images,mp=/root/.olares/images" \
   --mp1 "/root/.olares/pkg,mp=/root/.olares/pkg" \
   --password="$ROOTPASS"
   ``` 

2. Modify the LXC configuration.

   a. Open the configuration file using the following command:

   ``` bash
   nano /etc/pve/lxc/16553.conf
   ```

   b. Copy and paste the following configurations into the file:

   ``` bash
   arch: amd64
   cores: 4
   hostname: olares
   memory: 10240
   net0: name=eth0,bridge=vmbr0,firewall=1,hwaddr=BC:24:11:13:05:7C,ip=dhcp,ip6=dhcp,type=veth
   ostype: debian
   rootfs: local-lvm:vm-16553-disk-0,size=80G

   # Storage config
   mp0: /root/.olares/images,mp=/root/.olares/images
   mp1: /root/.olares/pkg,mp=/root/.olares/pkg

   # Permision config 
   lxc.apparmor.profile: unconfined
   lxc.cgroup.devices.allow: a
   lxc.cap.drop:
   lxc.mount.auto: "proc sys cgroup:mixed"
   ```

   c. Save and close the file.

3. Enable IP Virtual Server (IPVS) modules on the PVE host:

   ``` bash
   sudo modprobe ip_vs
   sudo modprobe ip_vs_rr
   sudo modprobe ip_vs_wrr
   sudo modprobe ip_vs_sh
   sudo modprobe overlay
   ```

4. Start the LXC container, make initial configurations, and exit:

   ```bash
   # Start the container
   pct start 16553

   # Enter the container
   pct enter 16553

   # Create missing directories
   mkdir -p /lib/modules

   # Update PATH environment variable
   echo 'export PATH="/usr/local/bin:$PATH"' >> /root/.bashrc
   source ~/.bashrc
   
   # exit LXC
   exit
   ```

5. Copy PVE dependencies to the LXC container:

   ``` bash
   # Copy kernel config from PVE host to LXC container
   pct push 16553 /boot/config-$(uname -r) /boot/config-$(uname -r)

   # Package and copy kernel modules directory
   tar cvf /lib/modules/6.8.4-2-pve.tar.gz /lib/modules/6.8.4-2-pve
   pct push 16553 /lib/modules/6.8.4-2-pve.tar.gz /lib/modules/6.8.4-2-pve.tar.gz

   # Extract the archive inside the container
   pct enter 16553
   cd /lib/modules
   tar xvf /lib/modules/6.8.4-2-pve.tar.gz -C /
   ```

## Install on LXC

Run the following installation command inside the LXC container:

<!--@include: ./reusables.md{4,33}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{35,39}-->