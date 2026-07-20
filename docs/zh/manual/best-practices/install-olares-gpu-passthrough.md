---
outline: [2, 3]
title: 在启用显卡直通的 PVE 上安装 Olares
description: 在 Proxmox VE（PVE）中配置 GPU 直通，并在启用 GPU 加速的虚拟机中安装 Olares 的详细教程。
---

# 在启用显卡直通的 PVE 上安装 Olares

Proxmox 虚拟环境（PVE）中的 GPU 直通允许虚拟机（VM）直接访问物理 GPU，从而启用 AI 模型推理、图形渲染等需要硬件加速的计算任务。

:::warning 不适用于生产环境
该部署方式当前仍有功能限制，仅用于开发或测试环境。
:::

## 学习目标

完成本教程后，你将能够：

- 在 PVE 主机上启用 GPU 直通；
- 创建一台带有 NVIDIA GPU 直通的 PVE 虚拟机；
- 通过官方 ISO 镜像安装 Olares 并验证 GPU 被识别。

## 前提条件

开始前，请准备：

- **CPU**：4 核及以上，并在 BIOS 中启用 IOMMU
  - Intel：`VT-d`
  - AMD：`AMD-Vi`/`IOMMU`
- **GPU**：支持 GPU 直通的 NVIDIA GPU
- **内存**：推荐 16 GB 及以上
- **存储**：不少于 200 GB 的 SSD 可用磁盘空间（HDD 可能导致安装失败）
- **PVE 版本**：8.3.2
- **Olares ISO 镜像**：[官方镜像](https://cdn.olares.cn/olares-v1.12.6-amd64.iso)

## 在 PVE 中配置 GPU 直通

要在 Olares 中使用 GPU 加速，先在 PVE 主机上启用 GPU 直通。

### 启用 IOMMU

**IOMMU（输入输出内存管理单元）** 是一种硬件功能，允许操作系统控制外设访问内存的方式。直通机制需要这一控制。

1. 在 PVE Shell 中打开 GRUB 配置文件：

   ```bash
   nano /etc/default/grub
   ```

2. 找到这一行：

   ```plain
   GRUB_CMDLINE_LINUX_DEFAULT="quiet"
   ```

   根据 CPU 类型替换为对应内容：

   ::: code-group
   ```bash [Intel]
   GRUB_CMDLINE_LINUX_DEFAULT="quiet intel_iommu=on iommu=pt"
   ```
   ```bash [AMD]
   GRUB_CMDLINE_LINUX_DEFAULT="quiet amd_iommu=on iommu=pt"
   ```
   :::

3. 保存并关闭文件，然后更新 GRUB 并重启主机：

   ```bash
   update-grub
   reboot
   ```

4. 重启后，检查 IOMMU 是否已启用：

   ```bash
   dmesg | grep -e DMAR -e IOMMU
   ```

   如果成功启用，你将看到类似输出：

   ::: code-group
   ```plain [Intel]
   [0.061644] DMAR: IOMMU enabled
   ...
   [0.408103] DMAR: Intel(R) Virtualization Technology for Directed I/O
   ```
   ```plain [AMD]
   [1.219719] AMD-Vi: Found IOMMU at 0000:00:00.2 cap 0x40
   ```
   :::

### 添加 VFIO 模块

**Virtual Function I/O (VFIO)** 使虚拟机能够直接访问 PCI 设备（如 GPU）。

1. 在 PVE 主机上打开 `modules` 文件：

   ```bash
   nano /etc/modules
   ```

2. 在文件末尾添加以下行：

   ```plain
   vfio
   vfio_iommu_type1
   vfio_pci
   vfio_virqfd
   ```

3. 保存并关闭文件。

### 屏蔽主机 GPU 驱动

屏蔽 PVE 主机的默认 GPU 驱动，让 GPU 专用于 `vfio-pci`。

1. 创建黑名单配置：

   ```bash
   nano /etc/modprobe.d/blacklist.conf
   ```

2. 添加以下行以屏蔽 NVIDIA 驱动：

   ```plain
   blacklist nouveau
   blacklist nvidia
   blacklist nvidiafb
   blacklist nvidia_drm
   blacklist nvidia_modeset
   ```

3. 保存并关闭文件。

### 将 GPU 绑定到 VFIO

1. 查找 GPU 的 PCI 地址：

   ```bash
   lspci | grep NVIDIA
   ```

   示例输出：

   ```plain
   01:00.0 VGA compatible controller: NVIDIA Corporation AD106 [GeForce RTX 4060 Ti] (rev a1)
   01:00.1 Audio device: NVIDIA Corporation AD106M High Definition Audio Controller (rev a1)
   ```

   本例中，GPU 的 PCI 地址为 `01:00`，包含两个功能。

2. 获取 GPU 的 PCI 标识符：

   ```bash
   lspci -n -s 01:00
   ```

   示例输出：

   ```plain
   01:00.0 0300: 10de:2803 (rev a1)
   01:00.1 0403: 10de:22bd (rev a1)
   ```

   本例中 GPU 的 ID 为 `10de:2803` 和 `10de:22bd`。

3. 将 ID 绑定到 VFIO（请替换为你自己的 ID）：

   ```bash
   echo "options vfio-pci ids=10de:2803,10de:22bd" > /etc/modprobe.d/vfio.conf
   ```

4. 更新 `initramfs` 并重启：

   ```bash
   update-initramfs -u
   reboot
   ```

5. 重启后，检查 GPU 是否正在使用 `vfio-pci` 驱动：

   ```bash
   lspci -v
   ```

   你应该会看到类似输出：

   ```plain
   01:00.0 VGA compatible controller: NVIDIA Corporation AD106 [GeForce RTX 4060 Ti] (rev a1) (prog-if 00 [VGA controller])
   Subsystem: Gigabyte Technology Co., Ltd AD106 [GeForce RTX 4060 Ti]
   Flags: fast devsel, IRQ 255, IOMMU group 11
   ...
   Kernel driver in use: vfio-pci
   ```

## 设置虚拟机并安装 Olares

启用 GPU 直通后，即可在 PVE 中安装 Olares。

### 创建和配置虚拟机

1. 将下载的 Olares 官方 ISO 上传到你的 PVE 存储（例如 `local`）：

   1. 在 PVE Web 界面中，选择目标存储（例如 `local`）。
   2. 点击 **ISO 镜像** > **上传**。
   3. 点击 **选择文件**，选择 Olares ISO 文件，然后点击 **上传**。

2. 点击 **创建虚拟机**。

3. 配置以下虚拟机参数：

   - **操作系统**：
     - `ISO 镜像`：选择下载的 Olares 官方镜像。
   - **系统**：
     - `BIOS`：选择 OVMF（UEFI）。
     - `EFI 存储`：选择一个存储位置（如本地 LVM 或目录），用于保存 UEFI 固件变量。
     - `预注册密钥`：**取消勾选**以禁用安全启动。
   - **磁盘**：
     - `磁盘大小 (GiB)`：不少于 200 GB。
   - **CPU**：
     - `核心`：4 核及以上。
   - **内存**：
     - `内存 (MiB)`：不少于 8 GB。

   下图为 PVE 中虚拟机硬件的示例配置。

   ![PVE 虚拟机硬件配置示例](/images/zh/manual/tutorials/pve-hardware-cn.png#bordered)

4. 点击 **完成**。**先不要启动虚拟机**。

### 将 GPU 绑定到虚拟机

1. 在 PVE 界面中，选择你的虚拟机，然后转到 **硬件** > **添加** > **PCI 设备**。

   ![添加 PCI 设备](/images/zh/manual/tutorials/pve-add-pci-cn.png#bordered)

2. 选择 **原始设备**，根据 PCI 地址（例如 `01:00`）选择你的 GPU。

3. 在右下角，选中 **高级** 选项，并勾选 **PCI-Express**。

4. 点击 **添加**。

   ![添加 GPU PCI 设备](/images/zh/manual/tutorials/pve-add-pci-gpu-cn.png#bordered){width=70%}

现在你的虚拟机已准备好使用 GPU 直通。

### 安装 Olares

通过 ISO 镜像安装 Olares：

1. 选择并启动你刚创建的虚拟机。

2. 从启动菜单中选择 **Install Olares to Hard Disk**，并按 **Enter** 确认。

3. 在 Olares System Installer 界面，选择安装磁盘：

   1. 查看可用磁盘列表（例如 `sda 200G QEMU HARDDISK`）。
   2. 输入 `/dev/` 加上第一个磁盘的名称来选择目标磁盘（例如 `/dev/sda`）。
   3. 当磁盘警告出现时，输入 `yes` 继续。

   :::info 忽略 NVIDIA 显卡驱动警告
   安装过程中会出现 NVIDIA 显卡驱动相关警告。按 **Enter** 忽略。
   :::

4. 安装完成后，你会看到以下信息：

   ```plain
   Installation completed successfully!
   ```

5. 在 PVE 界面中重启虚拟机。

### 验证安装和 GPU 直通

虚拟机重启后，会进入 Ubuntu 系统。

1. 使用默认账号登录 Ubuntu：

   - 用户名：`olares`
   - 密码：`olares`

2. 运行以下命令确认 Olares 是否安装成功：

   ```bash
   sudo olares-check
   ```

   如果输出如下，则安装成功：

   ```plain
   ...
   check Olaresd:  success
   check Containerd:  success
   ```

3. 使用 NVIDIA 系统管理界面工具检查 GPU 是否已被直通并被 Olares 识别：

   ```bash
   nvidia-smi
   ```

   如果 GPU 直通设置成功，此命令会显示一张表格，列出 NVIDIA GPU 的详细信息，包括名称、驱动程序版本和内存使用情况。

<!--@include: ../get-started/install-and-activate-olares.md-->

<!--@include: ../get-started/log-in-to-olares.md-->
