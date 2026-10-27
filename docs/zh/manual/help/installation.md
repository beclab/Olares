---
outline: [2, 3]
description: 查找在支持的硬件和网络环境中安装、激活 Olares 的常见问题答案。
head:
  - - meta
    - name: keywords
      content: Olares, 安装常见问题, 激活常见问题, 系统要求, 蓝牙激活, NVIDIA GPU
---
# 安装与激活常见问题

本页用于查找安装和激活 Olares 的常见问题。激活或登录出现明确报错时，请在[登录与激活错误信息](../login-and-activation-errors.md)中查询。

## 安装

### Olares 支持哪些平台？

建议在 Linux (Ubuntu 或 Debian) 上安装 Olares，以获得最佳性能。

你也可以在以下平台安装 Olares，但仅建议用于测试，因为这些平台不支持全部功能：
* Proxmox VE
* Raspberry Pi
* macOS
* Windows

### 安装 Olares 的最低硬件要求是什么？

具体要求因平台而异。通常建议配置如下：
* **CPU**：至少 4 核，x86-64 架构 (Intel 或 AMD)。
* **内存**：至少 8 GB 可用内存。
* **存储**：至少 150 GB SSD。

详细要求参见[安装文档](../get-started/install-olares.md)。

### 可以使用机械硬盘安装 Olares 吗？

不可以。必须使用 SSD。由于机械硬盘读写速度较慢，极易导致系统初始化超时，从而造成安装失败。

### 系统支持 NVIDIA 显卡吗？

支持，但仅限 NVIDIA Turing 架构及更新版本，例如 GTX 16xx、RTX 20xx、30xx、40xx、50xx 系列及更高版本。架构更旧的显卡在 Olares 中不会被识别，依赖 GPU 的 AI 应用也无法运行。该限制源于两方面：NVIDIA 开源驱动需要 Turing 及更新架构的 GSP 模块，且 CUDA 13.x 不再支持旧架构显卡。

对于受支持的显卡，Olares 会自动完成驱动安装。Olares 也支持在单块主板上安装多块 GPU，让你能够充分利用所有显卡的算力来处理 AI 任务。

### 如果自动安装失败，如何手动安装 NVIDIA 驱动？

Olares 安装程序会自动检测并安装驱动。但如果系统此前已安装过 NVIDIA 驱动，可能会因冲突导致安装过程跳过或失败。

解决方法：
1. 完成 Olares 安装后重启机器，清除旧驱动残留。
2. 使用命令 `olares-cli gpu install` 手动触发驱动安装。
3. 安装完成后，运行 `nvidia-smi` 确认系统识别到了 GPU。

### 为什么安装失败并提示 `failed to build Kubernetes objects` 或 `Ensure CRDs are installed first`？

虽然这些错误信息指向自定义资源定义（CRD）的问题，但通常真正的原因是磁盘性能不足。

Olares 依赖 etcd 作为 Kubernetes 的后端数据库，而 etcd 对存储速度非常敏感。如果在速度较慢的磁盘（如传统机械硬盘）上安装，etcd 无法及时响应，就会导致 API 服务器在应用 CRD 时超时。

使用 SSD 安装 Olares 通常可以解决此问题。

### 如果 Olares 安装超时且未显示密码，如何找到密码？

这种情况通常发生在虚拟机（VM）中，由于系统资源不足导致安装超时。可以从安装日志文件中通过以下命令获取密码：

```bash
# 将 v1.12.2 替换为你的具体 Olares 版本号
grep password $HOME/.olares/versions/v1.12.2/logs/install.log
```

安装超时通常意味着某些服务未能正确启动。找到密码后，运行 `kubectl get pod -A` 检查所有服务的状态。

## 激活

### 通过蓝牙激活

如果 LarePass 找不到你的 Olares 设备，可以使用蓝牙激活。这通常发生在 Olares 没有连接有线网络，或者你的手机和 Olares 处于不同网络的情况下。
通过蓝牙，你可以将 Olares 直接连接到你手机当前的 Wi-Fi 网络，以便继续操作。
![蓝牙配网](/images/zh/manual/larepass/bluetooth-network.png#bordered)

1. 在**未发现 Olares** 提示页面底部，点击**蓝牙配网**选项。LarePass 将使用手机蓝牙扫描附近的 Olares 设备。
2. 设备显示后，点击**配置网络**。
3. 选择手机当前连接的 Wi-Fi 网络。如果该网络有密码保护，请输入密码并点击**确认**。
4. Olares 将开始切换网络。完成后你会看到成功消息。此时，如返回到**蓝牙配网**页面，你将看到 Olares 的 IP 地址已更改为与你手机 Wi-Fi 相同的网络。
5. 返回到设备扫描页面，点击**发现附近的 Olares**，找到你的设备并继续激活。


### 能否在非本地网络下激活 Olares？

标准激活要求 Olares 设备与客户端设备（如手机）连接到同一个本地网络。无论你是通过浏览器访问本地 IP 地址来进入激活向导，还是在 ISO 安装后使用 LarePass 应用中的“发现附近的 Olares”功能，都适用这个要求。

但如果 Olares 使用的是公网 IP 地址（例如在公有云上），则不再受这个本地网络限制。

激活之后，无论初始设置方式如何，都可以通过域名在内网和外网访问设备。

### Olares 已开机并连接到局域网，但在 LarePass 中找不到设备。怎么办？

确保手机和 Olares 设备处于同一网络。如果不在同一网络，LarePass 无法自动发现 Olares。

如果无法通过 Wi-Fi 连接，可使用 LarePass 应用中的蓝牙配网功能，将 Olares 连接到与手机相同的网络中。

详细步骤参见[通过蓝牙激活 Olares](#通过蓝牙激活)。
