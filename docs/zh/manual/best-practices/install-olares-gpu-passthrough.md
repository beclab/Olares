---
outline: [2, 3]
description: 在 Proxmox VE（PVE）中配置 GPU 直通，并在启用 GPU 加速的虚拟机中安装 Olares 的详细教程。
---

# 在启用显卡直通的 PVE 上安装 Olares

**Proxmox 虚拟环境（PVE）** 中的 GPU 直通允许虚拟机（VM）直接访问物理 GPU ，从而启用 AI 模型推理、图形渲染等需要硬件加速的计算任务。

本教程将完整介绍如何在 PVE 主机中：
- 配置 GPU 直通；
- 通过官方 ISO 镜像安装 Olares。

这样，你可以在虚拟机中充分利用独立 GPU 的算力。

:::warning 不适用于生产环境
该部署方式当前仍有功能限制，建议仅用于开发或测试环境。
:::

## 前提条件

在开始前，请确保你的设置满足以下要求：

- CPU：4 核及以上，并在 BIOS 中启用了 IOMMU
  - Intel：`VT-d`
  - AMD：`AMD-Vi`/`IOMMU`
- GPU：支持 GPU 直通的 NVIDIA GPU
- RAM：推荐 16 GB 及以上
- 存储：不少于 200 GB的可用磁盘空间，需使用 SSD 硬盘安装。使用 HDD（机械硬盘）可能会导致安装失败。
- 支持的系统版本：PVE 8.3.2
- Olares ISO 镜像：点击[此处](https://cdn.joinolares.cn/olares-v1.12.1-amd64-cn.iso)下载 Olares 官方镜像。

## 在 PVE 中配置 GPU 直通

要在 Olares 中使用 GPU 加速计算，你需要首先在 PVE 主机上启用 GPU 直通功能。

### 启用 IOMMU

<b>IOMMU（输入输出内存管理单元）</b>是一种硬件功能，允许操作系统控制外设访问内存的方式，是直通机制的核心。

1. 在 PVE Shell 中执行以下命令以编辑 GRUB 配置文件：
        
    ```bash
    nano /etc/default/grub
    ```
    
2. 找到这一行：`GRUB_CMDLINE_LINUX_DEFAULT="quiet"`
    
    根据你的 CPU 类型修改为对应内容：
   
    ::: code-group
    ```bash [Intel]
    GRUB_CMDLINE_LINUX_DEFAULT="quiet intel_iommu=on iommu=pt"
    ```
    ```bash [AMD]
    GRUB_CMDLINE_LINUX_DEFAULT="quiet amd_iommu=on iommu=pt"
    ```
    :::

3. 保存并退出后，更新 GRUB 并重启主机：
  
    ```bash
    update-grub
    reboot
    ```
4.  重启后，在 PVE 主机上检查 IOMMU 是否已启用：

    ```bash
    dmesg | grep -e DMAR -e IOMMU
    ```
       如果成功启用，你将看到类似输出：

    ::: code-group
    ```bash [Intel]
    [0.061644] DMAR: IOMMU enabled
    ...
    [0.408103] DMAR: Intel(R) Virtualization Technology for Directed I/O
    ```
    ```bash [AMD]
    [1.219719] AMD-Vi: Found IOMMU at 0000:00:00.2 cap 0x40
    ```
    :::

### 添加 VFIO 模块

<b>Virtual Function I/O (VFIO)</b>使虚拟机能够直接访问 PCI 设备（如 GPU）。

1. 在 PVE 主机上，运行以下命令打开`modules`文件：

    ```bash
    nano /etc/modules
    ```

2. 在文件末尾添加以下行：
    
    ```
    vfio
    vfio_iommu_type1
    vfio_pci
    vfio_virqfd
    ```

3. 保存并关闭文件。

### 屏蔽主机 GPU 驱动程序

为避免 PVE 主机占用你计划直通的 GPU，建议屏蔽其默认驱动，让 GPU 专用于`vfio-pci`。

1. 在 PVE 主机上运行以下命令来创建黑名单配置：

    ```bash
    nano /etc/modprobe.d/blacklist.conf
    ```

2. 添加以下行以屏蔽 NVIDIA 驱动程序：

    ```
    blacklist nouveau 
    blacklist nvidia 
    blacklist nvidiafb
    blacklist nvidia_drm
    blacklist nvidia_modeset
    ```

3. 保存并关闭文件。

### 将 GPU 绑定到 VFIO

1. 在 PVE 主机上运行以下命令以查找你的 GPU 的 PCI 地址：

    ```bash
    lspci | grep NVIDIA
    ```

    **示例输出**：

    ```
    01:00.0 VGA compatible controller: NVIDIA Corporation AD106 [GeForce RTX 4060 Ti] (rev a1)
    01:00.1 Audio device: NVIDIA Corporation AD106M High Definition Audio Controller (rev a1)
    ```

    本例中，GPU 的 PCI 地址为`01:00`，并列出了两个功能。

2. 获取你的 GPU 的 PCI 标识符：
   
    ```bash
    lspci -n -s 01:00
    ```
    **示例输出**：

    ```
    01:00.0 0300: 10de:2803 (rev a1)
    01:00.1 0403: 10de:22bd (rev a1)
    ```

    本例中 GPU 的 ID 为`10de:2803`和`10de:22bd`。

3. 将 ID 绑定到 VFIO （请用你自己的 ID 替换）：

    ```bash
    echo "options vfio-pci ids=10de:2803,10de:22bd" > /etc/modprobe.d/vfio.conf
    ```

4. 更新`initramfs`（根文件系统）以应用所有模块和驱动程序的更改，然后重启系统：

    ```bash
    update-initramfs -u
    reboot
    ```

5. 重启后，检查 GPU 现在是否正在使用 `vfio-pci` 驱动程序：

    ```bash
    lspci -v
    ```

    你应该会看到类似输出：

    ```
    01:00.0 VGA compatible controller: NVIDIA Corporation AD106 [GeForce RTX 4060 Ti] (rev a1) (prog-if 00 [VGA controller])    
    Subsystem: Gigabyte Technology Co., Ltd AD106 [GeForce RTX 4060 Ti]
    Flags: fast devsel, IRQ 255, IOMMU group 11
    ...
    Kernel driver in use: vfio-pci
    ```


## 设置虚拟机并安装 Olares

启用 GPU 直通后，你现在可以在 PVE 中安装 Olares 了。

### 创建和配置虚拟机

本节为你介绍如何使用 Olares ISO 镜像来创建和配置虚拟机：

1. 将下载的 Olares 官方 ISO 上传到你的 PVE 存储（例如，`local`）。

    1. 在 PVE Web 界面中，选择目标存储（例如，`local`）。
    
    2. 点击**ISO 镜像** > **上传**。

    3. 点击**选择文件**，选择下载的 Olares ISO 文件，然后点击**上传**。

2. 点击**创建虚拟机**.

3. 配置以下虚拟机参数：

    - 操作系统：
        - `ISO 镜像`：选择下载的 Olares 官方镜像。
    - 系统：
        - `BIOS`：选择 OVMF（UEFI）。
        - `EFI 存储`：选择一个存储位置（如本地 LVM 或目录），用于保存 UEFI 固件变量。
        - `预注册密钥`：**取消勾选**以禁用安全启动。
    - 磁盘:
        - `磁盘大小 (GiB)`：不少于 200 GB
    - CPU:
        - `核心`：4核及以上
    - 内存:
        - `内存 (MiB)`：不少于 8 GB

4. 点击**完成**。**先不要启动虚拟机**。

    下图为 PVE 中虚拟机硬件的示例配置。
    ![Hardware](/images/zh/manual/tutorials/pve-hardware-cn.png#bordered)

### 将 GPU 绑定到虚拟机

按照以下步骤将 GPU 绑定到虚拟机：

1. 在 PVE 界面中，选择你的虚拟机，然后转到**硬件** > **添加** > **PCI 设备**。
![Add PCI](/images/zh/manual/tutorials/pve-add-pci-cn.png#bordered)

2. 选择**原始设备**，根据 PCI 地址（例如，`01:00`）选择你的 GPU。

3. 在右下角，选中**高级**选项，并勾选 **PCI-Express**。

4. 点击**添加**以保存。
![Add GPU](/images/zh/manual/tutorials/pve-add-pci-gpu-cn.png#bordered)

现在你的虚拟机已准备好使用 GPU 直通。

### 安装 Olares

虚拟机创建完成后，按照以下步骤在 PVE 上安装 ISO。

1. 选择并启动你刚创建的虚拟机。

2. 从启动菜单中，选择 **Install Olares to Hard Disk**，并按回车确认。

3. 在 Olares System Installer 界面，选择安装磁盘。

    1. 查看可用磁盘列表（例如，`sda 200G QEMU HARDDISK`）。
    
    2. 输入`/dev/`加上第一个磁盘的名称来选择目标磁盘（例如，`/dev/sda`）。
    
    3. 当屏幕上出现警告时，输入`yes`继续即可。

    ::: tip 注意
    安装过程中会出现与 NVIDIA 显卡驱动相关的警告。按**回车键**忽略即可。
    :::

4. 安装完成后，你会看到以下信息：

    ```
    Installation completed successfully!
    ```
    
5.  在 PVE 界面，选择**重启**以重启虚拟机。

### 验证安装和 GPU 直通

虚拟机重启后，将进入 Ubuntu 系统。

1. 使用默认账号登录 Ubuntu：
  - 用户名：`olares`
  - 密码：`olares`

2. 运行以下命令确认 Olares 是否安装成功：

    ```bash
    sudo olares-check
    ```

    如果运行结果如下，则说明安装成功：

    ```      
    ...
    check Olaresd:  success
    check Containerd:  success
    ```

3. 最后，使用 NVIDIA 系统管理界面工具验证 GPU 是否已成功直通并被 Olares 识别：

    ```bash
    nvidia-smi
    ```

    如果 GPU 直通设置成功，此命令将显示一张表格，列出 NVIDIA GPU 的详细信息，包括名称、驱动程序版本和内存使用情况。

## 后续步骤

Olares 现已安装完成，并在 GPU 加速模式下正常运行。
接下来你可以激活设备并登录账户。
有关详细步骤，请参阅我们的官方指南：
- [激活 Olares](../get-started/install-and-activate-olares.md)
- [登录 Olares](../get-started/log-in-to-olares.md)