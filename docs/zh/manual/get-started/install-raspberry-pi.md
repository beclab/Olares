---
description: 使用树莓派安装配置 Olares 的完整步骤，包括硬件要求、安装命令和系统设置。
---
# 在树莓派上安装 Olares
本文介绍如何在树莓派设备上安装 Olares。

::: warning 不适用于生产环境
该部署方式当前仍有功能限制，建议仅用于开发或测试环境。
:::

## 系统要求
请确保树莓派设备满足以下要求：
- 硬件：树莓派 4B 或树莓派 5（8GB 内存版本）
- 操作系统：Raspbian 12
- 存储：不少于 150GB 的可用磁盘空间，需使用 SSD 硬盘安装。
:::warning 注意
使用 HDD（机械硬盘）可能会导致安装失败。
:::

## 配置环境
1. 配置 Raspbian 环境，启用必要功能：

      ```bash
      echo "$(head -1 /boot/firmware/cmdline.txt) cgroup_enable=cpuset cgroup_enable=memory cgroup_memory=1" | sudo tee /boot/firmware/cmdline.txt

      echo "kernel=kernel8.img" | sudo tee -a /boot/firmware/config.txt
      ```

2. 重启树莓派使配置生效。

   ```bash
   sudo reboot
   ```
## 安装 Olares
执行以下命令:

<!--@include: ./reusables.md{4,21}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{33,37}-->