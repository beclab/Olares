---
description: 在 LXC 容器环境中安装配置 Olares 的完整步骤，包括容器设置、系统要求和激活方法。
---
# 在 Linux 容器（LXC）中安装 Olares
LXC 是一种轻量级的虚拟化技术，可以在隔离的容器中运行应用程序。在 PVE 环境下使用 LXC 部署 Olares 能够避免完整虚拟机的额外开销，提供了一种更高效的部署方式。

::: warning 不适用于生产环境
该部署方式当前仍有功能限制，建议仅用于开发或测试环境。
:::

<!--@include: ./reusables.md{36,41}-->

## 系统要求
请确保设备满足以下配置要求：

- CPU：4 核及以上
- 内存：不少于 8GB 可用内存
- 存储：建议使用 SSD，且可用磁盘空间不少于 64GB
- 支持的系统版本：
  - PVE: 8.2.2
  - LXC 容器系统：Debian 12（既有 LXC 环境）

:::info 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台上安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::

## 准备工作

- 在 PVE 主机上创建用于存储镜像和软件包的工作目录。
  
   ```bash
   mkdir -p /root/.olares/images /root/.olares/pkg
   ```
-  `debian-12-standard_12.7-1_amd64.tar.zst` 的容器模板（CT），可从 PVE [镜像仓库](http://download.proxmox.com/images/system/)下载。

## 配置 LXC 环境

::: tip 安装至已有 LXC 容器
如果你想要在 PVE 中已有 LXC 容器上安装 Olares，请直接到第二步更新 LXC 配置。要记得更新对应的容器 ID。
:::

1. 使用以下命令创建 LXC 容器：

   ::: tip 指定唯一容器 ID
   要创建容器，必须分配一个唯一的**容器 ID**。此处以 `16553` 为例，你可以将其替换为任何可用的数字 ID，并在所有相关命令和配置中更新此 ID。
   :::

   ```bash{2}
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

2. 修改 LXC 配置。
   
   a. 打开配置文件:

   ```bash
   nano /etc/pve/lxc/16553.conf
   ```
   
   b. 复制并粘贴以下配置到文件中:
      
      ```bash
      # 基础配置
      arch: amd64
      cores: 4
      hostname: olares
      memory: 10240
      net0: name=eth0,bridge=vmbr0,firewall=1,hwaddr=BC:24:11:13:05:7C,ip=dhcp,ip6=dhcp,type=veth
      ostype: debian
      rootfs: local-lvm:vm-16553-disk-0,size=80G

      # 存储配置
      mp0: /root/.olares/images,mp=/root/.olares/images
      mp1: /root/.olares/pkg,mp=/root/.olares/pkg

      # 权限配置
      lxc.apparmor.profile: unconfined
      lxc.cgroup.devices.allow: a
      lxc.cap.drop:
      lxc.mount.auto: "proc sys cgroup:mixed"
      ```
   
   c. 保存并退出编辑界面。

3. 在 PVE 主机上启用 IP 虚拟服务器 （IPVS) 模块：

   ```bash
   sudo modprobe ip_vs
   sudo modprobe ip_vs_rr
   sudo modprobe ip_vs_wrr
   sudo modprobe ip_vs_sh
   sudo modprobe overlay
   ```
4. 启动并配置 LXC 容器。

   ```bash 
   # 启动容器
   pct start 16553

   # 进入容器
   pct enter 16553

   # 创建缺失的目录
   mkdir -p /lib/modules

   # 更新 PATH 环境变量
   echo 'export PATH="/usr/local/bin:$PATH"' >> /root/.bashrc
   source ~/.bashrc
      
   # 退出 LXC
   exit
   ```

5. 将 PVE 依赖项复制到 LXC 容器。
   
   ```bash
   # 将内核配置从 PVE 主机复制到 LXC 容器
   pct push 16553 /boot/config-$(uname -r) /boot/config-$(uname -r)
   
   # 打包并复制内核模块目录
   tar cvf /lib/modules/6.8.4-2-pve.tar.gz /lib/modules/6.8.4-2-pve
   pct push 16553 /lib/modules/6.8.4-2-pve.tar.gz /lib/modules/6.8.4-2-pve.tar.gz
   
   # 在 LXC 容器内解压内核模块文件
   pct enter 16553
   cd /lib/modules
   tar xvf /lib/modules/6.8.4-2-pve.tar.gz -C /
   ```

## 安装 Olares

在 LXC 容器中运行以下安装命令：

<!--@include: ./reusables.md{4,28}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{30,34}-->