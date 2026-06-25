---
outline: [2, 4]
description: 学习如何在 Olares 上为 Windows 虚拟机启用 Intel 集成显卡直通，包括支持的设备、主机配置、Windows 应用设置、驱动安装和验证。
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../use-cases/windows-intel-gpu-passthrough.md)为准。
:::

# 为 Windows VM 启用 Intel 集成显卡直通

您可以在支持的设备上将 Intel 集成显卡（iGPU）从 Olares 主机直通到 Windows VM。配置完成后，Windows 可以通过 RDP 检测到 Intel 显卡设备。

:::tip 使用 Olares One
Olares One 不需要 BIOS、SR-IOV、VFIO 或 Windows 应用环境变量设置。从 Olares Market 部署的 Windows 应用默认已包含所需的 GPU 配置。

如果您使用 Olares One，请直接跳转到[在 Windows 中安装 Intel 显卡驱动](#在-windows-中安装-intel-显卡驱动)。
:::

## 支持的设备

仅当 Olares 主机满足以下条件之一时，才支持 Intel 集成显卡：

- **Olares One**：您正在使用官方 Olares One 设备。
- **自托管设备**：您正在使用带有 Intel 集成显卡的 Intel CPU 主机，运行 Ubuntu 24.04 或更高版本，并且在 BIOS 中支持并启用了 Intel VT-d / IGD。

如果主机没有 Intel iGPU，Windows VM 无法仅通过软件创建一个。

## 配置主机环境

使用此工作流程在自托管 Olares 设备上配置 BIOS 和操作系统。

在继续之前，确保您可以通过 SSH 访问 Olares 主机终端。此设置会更改 BIOS 设置、内核参数、DKMS 模块、SR-IOV 虚拟功能和 VFIO 设备绑定。配置不正确可能会影响主机启动或设备可用性。

### 配置 BIOS

重启主机并进入 BIOS/UEFI 设置。确切的菜单名称因主板厂商而异。在 **Advanced**、**Chipset**、**System Agent**、**PCIe** 或 **Graphics** 等菜单下查找这些选项：

- 启用 **Intel VT-d**。
- 如果集成显卡设备被禁用，则启用它。
- 将主显示器或初始显示输出设置为 **IGD**、**Integrated Graphics** 或类似选项。
- 保存设置并重启主机。

如果您找不到集成显卡或 IGD 选项，请确认您的 Intel CPU 包含集成显卡，然后查看主板手册以获取确切的 BIOS 选项名称。

### 验证 iGPU

通过 SSH 访问 Olares 主机终端并运行：

```bash
lspci -nn | grep VGA
```

确认输出包含 Intel 集成显卡，并且设备类型为 `VGA compatible controller`。

示例输出：

```plain
00:02.0 VGA compatible controller [0300]: Intel Corporation AlderLake-S GT1 [8086:4680] (rev 0c)
```

### 启用 IOMMU 和 Intel vGPU

此步骤安装 Intel SR-IOV DKMS 模块，将所需的 IOMMU 和 iGPU 参数写入 GRUB，并创建 vGPU 虚拟功能。

#### 安装 DKMS 模块

运行以下命令以更新主机、安装依赖项并构建 Intel SR-IOV DKMS 模块：

```bash
sudo apt update && sudo apt upgrade -y

sudo apt update && sudo apt install git sysfsutils dkms -y

sudo rm -rf /var/lib/dkms/i915-sriov-dkms*
sudo rm -rf /usr/src/i915-sriov-dkms*
rm -rf ~/i915-sriov-dkms

cd ~
git clone https://github.com/strongtz/i915-sriov-dkms.git
cd ~/i915-sriov-dkms
git checkout 2025.07.22

cp -a dkms.conf{,.bak}
KERNEL=$(uname -r)
sed -i 's/"@_PKGBASE@"/"i915-sriov-dkms"/g' dkms.conf
sed -i "s/PACKAGE_VERSION=\".*\"/PACKAGE_VERSION=\"$KERNEL\"/g" dkms.conf
sed -i 's/ -j$(nproc)//g' dkms.conf
cat dkms.conf

sudo apt install --reinstall dkms -y
sudo dkms add .
cd /usr/src/i915-sriov-dkms-$KERNEL
sudo dkms status
sudo dkms install -m i915-sriov-dkms -v $KERNEL -k $(uname -r) --force -j 1
sudo dkms status
```

#### 更新 GRUB

在 `/etc/default/grub` 中添加或修改 `GRUB_CMDLINE_LINUX_DEFAULT`。使用与您的 Ubuntu 版本匹配的命令行。

对于 Ubuntu 24.10 或更高版本，使用：

```text
GRUB_CMDLINE_LINUX_DEFAULT="quiet intel_iommu=on iommu=pt i915.enable_guc=3 i915.max_vfs=7"
```

对于 Ubuntu 24.04，添加 `i915.force_probe=7d67`：

```text
GRUB_CMDLINE_LINUX_DEFAULT="quiet intel_iommu=on iommu=pt i915.enable_guc=3 i915.max_vfs=7 i915.force_probe=7d67"
```

如果该行已包含其他参数，请保留它们并在引号内添加新参数。

在此示例中，`i915.max_vfs=7` 配置最多 7 个 Intel vGPU 虚拟功能。

更新 GRUB 配置并重启主机：

```bash
sudo update-grub && sudo reboot
```

#### 添加 SR-IOV 配置

主机重启后，将 SR-IOV 配置添加到 `/etc/sysfs.conf`。

下面的命令使用 `00:02.0`，它应与之前 `lspci` 输出中显示的 Intel iGPU PCI 地址匹配。如果您的 Intel iGPU 使用不同的 PCI 地址，请相应地更新路径。

```bash
sudo sh -c 'echo "devices/pci0000:00/0000:00:02.0/sriov_numvfs = 7" > /etc/sysfs.conf'

cat /etc/sysfs.conf

sudo reboot
```

#### 验证 vGPU 设备

主机重启后，验证 vGPU 设备是否已创建：

```bash
lspci -nn | grep VGA
```

您应该看到额外的虚拟 VGA 设备。

### 将一个 vGPU 绑定到 VFIO

将一个 vGPU（例如 `0000:00:02.1`）绑定到 VFIO 驱动：

```bash
sudo apt install -y driverctl

sudo driverctl set-override 0000:00:02.1 vfio-pci
```

在此示例中，`0000:00:02.1` 是将分配给 Windows VM 的 vGPU 设备。该覆盖在主机重启后仍然有效。

## 配置 Windows 应用

对于自托管设备，将 GPU 相关的环境变量添加到 Windows Deployment。确保 `host=` 值与您之前绑定到 `vfio-pci` 的 vGPU 设备匹配。下面的示例使用 `0000:00:02.1`。

1. 从 Launchpad 打开 Control Hub。
2. 在 **Browse** 下，从列表中选择 windows 项目。
3. 在 **Deployments** 下，选择 `windows`。
4. 在详情窗格的右上角，点击 <i class="material-symbols-outlined">edit_square</i> 编辑 YAML。
5. 在 `env` 部分下，添加或更新以下变量：

   ```yaml
   env:
     - name: GPU
       value: 'Y'
     - name: VGA
       value: 'vfio-pci,host=0000:00:02.1,multifunction=on,x-vga=on -vga virtio'
   ```

   ![Edit yaml](/images/manual/use-cases/windows-edit-yaml.png#bordered)

6. 点击 **Confirm** 保存更改，然后重启 Windows 应用。

## 在 Windows 中安装 Intel 显卡驱动

1. 使用 RDP 连接到 Windows VM。
2. 在 Windows 中打开浏览器。
3. 从 Intel 下载最新的 Intel 显卡驱动：

   ```plain
   https://www.intel.com/content/www/us/en/download/785597/intel-arc-iris-xe-graphics-windows.html
   ```

4. 安装驱动。
5. 如果提示，重启 Windows。

## 验证结果

1. 使用 RDP 重新连接到 Windows VM。
2. 打开 **Device Manager**。
3. 展开 **Display adapters**。
4. 确认 Intel 显卡设备已出现且运行正常，没有错误。

   ![Intel integrated GPU](/images/manual/use-cases/windows-intel-gpu.png#bordered)

