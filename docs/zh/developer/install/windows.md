---
outline: [2, 3]
description: 通过 WSL 在 Windows 系统安装配置 Olares 的完整步骤，包括环境要求和激活设置。
---
# 在 Windows 上使用脚本安装 Olares
本文介绍如何在 Windows （WSL 2）上使用脚本安装 Olares。

:::warning 不适用于生产环境
Windows 版 Olares 目前存在以下限制：
- 不支持分布式存储
- 无法添加本地节点

建议仅用于开发或测试环境。
:::

<!--@include: ./reusables.md{36,41}-->

## 系统要求
Windows 设备需满足以下条件：
- CPU：4 核及以上
- 内存：不少于 16GB 可用内存
- 存储空间：建议使用 SSD，且可用磁盘空间不少于 64GB
- 支持的系统：
    - Windows 10 或 11
    - Linux（WSL 2 环境）：Ubuntu 20.04 LTS 及以上；Debian 11 及以上
## 配置系统环境
1. 启用虚拟化所需的 Windows 功能。

   a. 打开**控制面板**，依次进入**程序** > **程序和功能** > **启用或关闭 Windows 功能**。

   b. 在弹出的 **Windows 功能** 窗口中，勾选以下选项：
     - **Hyper-V**（Windows 10 家庭版和 Windows 11 家庭版无此选项）
     - **适用于 Linux 的 Windows 子系统**
     - **虚拟机平台**

   c. 点击**确定**，然后根据提示重启电脑。

2. 设置当前用户的执行策略。

   a. 以管理员身份打开 PowerShell，运行以下命令：
    ```powershell
    Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope CurrentUser
    ```
   b. 当提示是否更改执行策略时，输入 `A` 并按下 **Enter** 确认。
    
    ```powershell{5}
    Execution Policy Change
    The execution policy helps protect you from scripts that you do not trust. Changing the execution policy might expose
    you to the security risks described in the about_Execution_Policies help topic at
    https:/go.microsoft.com/fwlink/?LinkID-135170. Do you want to change the execution policy?
    [Y] Yes [A] Yes to All [N] No [L] No to All [S] Suspend [?] Help (default is "N"):
    ```

## 安装 Olares
1. 点击 https://cn.windows.olares.sh 下载安装脚本 `publicInstall.latest.ps1`。

2. 执行安装脚本。

   a. 以管理员身份打开 PowerShell 并导航至脚本所在文件夹。例如，如果脚本在 `Downloads` 文件夹里，则执行以下命令：
      ```powershell
      cd C:\Users\<YourUsername>\Downloads
      ```
   
   b. 进入正确的文件目录后，执行以下命令：
      ```powershell
      .\publicInstall.latest.ps1
      ```
   :::warning 需要管理员权限
   不以管理员身份运行 PowerShell 会导致安装失败。参考[如何确认 PowerShell 是否以管理员身份运行](#如何确认-powershell-是否以管理员身份运行)。
   :::

3. 出现安全提示时，输入 `R` 并按下 **Enter** 以运行脚本，开始安装 Olares。

   ```powershell{4}
   Security warning
   Run only scripts that you trust. While scripts from the internet can be useful, this script can potentially harm your computer. If you trust this script, use the Unblock-File cmdlet to allow the script to run without this warning message. Do you want to run
   publicInstall.latest.ps1?
   [D] Do not run [R] Run once [S] Suspend [?] Help (default is "D"):
   ```
4. 选择 WSL Ubuntu 的存储位置。请输入一个可用磁盘的盘符，并确保所选磁盘至少有 **80 GB** 的可用空间。
   ```powershell{8}
   Installing Olares will create a WSL Ubuntu Distro and occupy at least 80 GB of disk space.
   Please select the drive where you want to install it.
   
   Available drives and free space:
   C:\  Free Disk: 391.07 GB
   D:\  Free Disk: 281.32 GB
   
   Please enter the drive letter (e.g., C):
   ```

5. 配置防火墙规则。输入 `yes` 自动设置防火墙规则，或者输入 `no` 跳过自动设置。<br>
   如果你选择跳过，可以[暂时关闭防火墙](#如何关闭-windows-defender-防火墙)，或[手动添加 TCP 入站规则](#如何手动设置防火墙规则)。
   ```powershell{2}
   Accessing Olares requires setting up firewall rules, specifically adding TCP inbound rules for ports 80, 443, and 30180.
   Do you want to set up the firewall rules? (yes/no):
   ```
6. 确认 Windows 的 IP 地址。输入 **Y** 确认，或者 **R** 重新输入。
   
   ```powershell
   The NAT gateway (the Windows host)'s IP is 192.168.50.136. Confirm [Y] or Re-enter [R]?
   ```
   ::: tip 获取 Windows IP 地址
   你可以提前通过在 Windows 命令行里执行 `ipconfig` 获取 Windows 的 IPv4 地址。
   :::

<!--@include: ./reusables.md{7,9}-->

:::info 安装遇到报错？
如果安装过程中出现错误，请先执行以下命令卸载：
```powershell
wsl --unregister ubuntu
```
卸载完成后，重新运行安装命令进行安装。
:::
<!--@include: ./reusables.md{20,28}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{30,34}-->

## 常见问题

### 如何确认 PowerShell 是否以管理员身份运行？

如果 PowerShell 窗口的标题栏显示“管理员: Windows PowerShell”，说明已以管理员权限启动。

![使用管理员权限打开 PowerShell](/images/manual/get-started/confirm-run-powershell-as-admin.png#bordered){width=70%}

如果没有看到“管理员”标识，你可以尝试以下两种方式启动 PowerShell：
- 在开始菜单中搜索“PowerShell”，右键点击 Windows PowerShell，选择**以管理员身份运行**。
- 按下 **Win** + **R**，输入“powershell”，然后按 **Ctrl** + **Shift** + **Enter** 打开管理员模式的 PowerShell。

### 如何配置 WSL 的 CPU 和内存？
在 WSL 上安装 Olares 时，默认内存分配为 `12GB`。但是，你可以在安装之前调整分配的内存大小，或在安装完成后调整内存和 CPU。

**在安装之前调整内存**

例如分配 16GB 内存：
1. 使用如下信息添加用户变量。
   - **变量名**: `WSL_MEMORY`
   - **变量值**: `16`

     ![添加用户变量](/images/manual/get-started/add-user-variable.png#bordered)
2. 点击**确定**使变更生效。
   :::tip 提示
   如果你已经打开了一个 PowerShell 窗口，环境变量的更改不会在当前会话中生效。请务必以管理员身份打开一个新的 PowerShell 窗口，然后再运行安装脚本。
   :::

**安装完成后调整内存和 CPU**

安装完成后，系统会在用户主目录下生成一个名为 `.wslconfig` 的配置文件（路径为 `C:\Users\<你的用户名>\`）。可以通过编辑此文件调整内存和 CPU 设置。默认配置如下：
```bash
[wsl2]
memory=12GB
swap=0GB
```
例如，设置为使用 4 核 CPU：
1. 在文件中添加 `processors` 参数：
   ```bash
   [wsl2]
   memory=12GB
   processors=4
   swap=0GB
   ```

2. 保存对 `.wslconfig` 文件的修改。 
3. 在 PowerShell 中运行以下命令，关闭虚拟机： 
   ```powershell   
   wsl --shutdown
   ```   
4. 运行以下命令重启 Olares： 
   ```powershell
   wsl -d Ubuntu
   ```   
重启后，等待几分钟，Olares 服务会重新启动并生效。

### 如何在重启电脑后重新启动 Olares？
通过以下命令手动启动 Olares 服务： 
```powershell
wsl -d Ubuntu
```

### 如何关闭 Windows Defender 防火墙？
:::tip 提示
建议在完成 Olares 安装后重新启用 Windows Defender 防火墙。
:::
按照以下步骤完全关闭防火墙：
1. 打开**控制面板** > **系统和安全** > **Windows Defender 防火墙**。
2. 在左侧导航栏中，点击**启用或关闭 Windows Defender 防火墙**。
3. 选择**关闭 Windows Defender 防火墙**，分别对**专用网络**和**公用网络**进行设置，然后点击**确定**。

   ![关闭 Windows Defender Firewall](/images/manual/get-started/disable-firewall.png#bordered)

### 如何手动设置防火墙规则？
如果在安装时选择不自动配置防火墙规则，可以通过以下步骤手动添加规则：
1. 打开**控制面板** > **系统和安全** > **Windows Defender 防火墙**。

   ![进入 Windows Defender 防火墙](/images/manual/get-started/select-firewall.png#bordered)
2. 在左侧导航栏中，点击**高级设置**。

   ![选择高级设置](/images/manual/get-started/select-advanced-settings.png#bordered)
3. 在左侧导航栏中，右键点击**入站规则**，然后选择**新建规则**。

   ![添加新规则](/images/manual/get-started/add-new-rule.png#bordered)
4. 在**新建入站规则向导**中，选择**端口**，然后点击**下一步**。

   ![选择端口规则](/images/manual/get-started/select-port.png#bordered)
5. 在**特定本地端口**输入框中，输入 `80`, `443`, `30180`，然后点击**下一步**。

   ![指定端口](/images/manual/get-started/specify-port.png#bordered)
6. 选择**允许连接**，然后点击**下一步**。

   ![允许连接](/images/manual/get-started/allow-the-connection.png#bordered)
7. 确保规则适用于**域**、**专用**和**公用**网络，然后点击**下一步**。

   ![确认规则适用范围](/images/manual/get-started/confirm-rules.png#bordered)
8. 为规则提供一个名称，然后点击**完成**。

   ![命名规则](/images/manual/get-started/name-the-rule.png#bordered)

### 如何卸载 Olares？
如果需要卸载 Olares，可以在 PowerShell 中运行以下命令：
```powershell
wsl --unregister ubuntu
```