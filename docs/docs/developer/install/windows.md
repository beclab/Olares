---
outline: [2, 3]
description: Guide to installing Olares on Windows using WSL (Windows Subsystem for Linux) including setup requirements and activation steps.
---
# Install Olares on Windows via the script
This guide explains how to install Olares on Windows (WSL 2) using the provided installation script.

:::warning Not recommended for production use
Currently, Olares on Windows has certain limitations including:
- Lack of distributed storage support
- Inability to add local nodes.

We recommend using it only for development or testing purposes.
:::
<!--@include: ./reusables.md{41,47}-->

## System compatibility
Make sure your Windows meets the following requirements.
- CPU: At least 4 cores
- RAM: At least 16GB of available memory
- Storage: At least 64GB of available space (SSD recommended)
- Supported systems:
    - Windows 10 or 11
    - Linux (on WSL 2): Ubuntu 20.04 LTS or later; Debian 11 or later
## Set up system environment
1. Enable the required Windows features for virtualization.

   a. Open **Control Panel**, then go to **Programs** > **Programs and Features** > **Turn Windows features on or off**.

   b. In the **Windows Features** window, check:
    - **Hyper-V** (not required for Windows 10 Home and Windows 11 Home)
    - **Windows Subsystem for Linux**
    - **Virtual Machine Platform**

   c. Click **OK** and restart your computer when prompted.

2. Set the execution policy for the current user.

   a. Open PowerShell as administrator, then run the following command:
    ```powershell
    Set-ExecutionPolicy -ExecutionPolicy Unrestricted -Scope CurrentUser
    ```
   b. When prompted to check whether to change the execution policy, type `A` and press **Enter** to confirm.

    ```powershell{5}
    Execution Policy Change
    The execution policy helps protect you from scripts that you do not trust. Changing the execution policy might expose
    you to the security risks described in the about_Execution_Policies help topic at
    https:/go.microsoft.com/fwlink/?LinkID-135170. Do you want to change the execution policy?
    [Y] Yes [A] Yes to All [N] No [L] No to All [S] Suspend [?] Help (default is "N"):
    ```

## Install Olares
1. Click https://windows.olares.sh to download the installation script `publicInstall.latest.ps1`.

2. Execute the script.
   
   a. Open PowerShell as administrator, then navigate to the folder where the script is located. For example, if the script is in the `Downloads` folder, run the following command:
   ```powershell
   cd C:\Users\<YourUsername>\Downloads
   ```
   
   b. Once in the correct folder, run the following command:
   ```powershell
   .\publicInstall.latest.ps1
   ```
  
   :::warning Administrator privileges required 
   Running PowerShell without administrator privileges will cause the installation to fail. See [How to make sure I am using PowerShell as administrator](#how-to-make-sure-i-am-using-powershell-as-administrator).
   :::

3. When prompted with security warning, type `R` and press **Enter** to run the script once. The installation process for Olares will start.

   ```powershell{4}
   Security warning
   Run only scripts that you trust. While scripts from the internet can be useful, this script can potentially harm your computer. If you trust this script, use the Unblock-File cmdlet to allow the script to run without this warning message. Do you want to run
   publicInstall.latest.ps1?
   [D] Do not run [R] Run once [S] Suspend [?] Help (default is "D"):
   ```

4. When prompted to select the drive to store the WSL Ubuntu distro, type the drive letter of an available disk. Ensure the selected drive has at least **80 GB** of free space.
   ```powershell{8}
   Installing Olares will create a WSL Ubuntu Distro and occupy at least 80 GB of disk space.
   Please select the drive where you want to install it.
   
   Available drives and free space:
   C:\  Free Disk: 391.07 GB
   D:\  Free Disk: 281.32 GB
   
   Please enter the drive letter (e.g., C):
   ```

5. When prompted with the firewall rules setup, type `yes` to automatically configure them, or type `no` to skip this step. <br>
   If you choose to skip, either [disable Windows Firewall Defender](#how-to-disable-windows-defender-firewall), or [manually add TCP inbound rules](#how-to-manually-set-firewall-rules).
   ```powershell{2}
   Accessing Olares requires setting up firewall rules, specifically adding TCP inbound rules for ports 80, 443, and 30180.
   Do you want to set up the firewall rules? (yes/no):
   ```
6. When promoted to confirm the IP address of Windows, type **Y** to confirm, or **R** to re-enter.

   ```powershell
   The NAT gateway (the Windows host)'s IP is 192.168.50.136. Confirm [Y] or Re-enter [R]?
   ```
   ::: tip Obtain the IPv4 address of Windows
   You can get the IPv4 address in advance by running `ipconfig` in the Windows command line.
   :::

<!--@include: ./reusables.md{7,9}-->

:::info Errors during installation?
If an error occurs during installation, use the following command to uninstall first:
```powershell
wsl --unregister ubuntu
```
After uninstalling, retry the installation by running the original installation command.
:::
<!--@include: ./reusables.md{20,33}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{35,39}-->

## FAQ

### How to make sure I am using PowerShell as administrator?
You can confirm that PowerShell is running as an administrator if you see "Administrator: Windows PowerShell" in the title bar of the PowerShell window.

![Confirm run Powershell as administrator](/images/manual/get-started/confirm-run-powershell-as-admin.png#bordered){width=70%}

If not, use one of the following methods:
- Search for "PowerShell" in the **Start** menu, right-click it, and select **Run as administrator**.
- Or press **Win** + **R**, type `powershell`, and press **Ctrl** + **Shift** + **Enter** to open PowerShell as an administrator.

### How to configure the CPU and memory for WLS?
When installing Olares in WSL, the default memory allocation is `12GB`. But you can configure the memory before Olares installation, or adjust both memory and CPU settings after installation.

**Adjust the memory setting before installation**

For example, to allocate 16GB of memory:

1. Add a user variable with the following:
   - **Variable name**: `WSL_MEMORY`
   - **Variable value**: `16`

   ![Add user variable](/images/manual/get-started/add-user-variable.png#bordered)

2. Click **OK** to apply changes.

   :::tip
   If you already have a PowerShell window open, changes to environment variables will not take effect in the current session. To ensure the updated environment variables are loaded, open a new PowerShell terminal as administrator, and then run the installation script.
   :::

**Adjust memory and CPU settings after installation**

After installation, a configuration file named `.wslconfig` will be created in the current user's home directory (`C:\Users\<YourUsername>\`). This file allows you to adjust memory and CPU settings. The default configuration looks like this:

```bash
[wsl2]
memory=12GB
swap=0GB
```

For example, to use 4 CPU cores:
1. Add the `processors` parameter to the file:
   ```bash
   [wsl2]
   memory=12GB
   processors=4
   swap=0GB
   ```
2. Save the `.wslconfig` file with your custom changes. 
3. Close all running virtual machines by running the following command in PowerShell:
   ```powershell
   wsl --shutdown
   ```
4. Restart Olares by running:
   ```powershell
   wsl -d Ubuntu
   ```
It will take a few minutes for Olares services to restart.

### How to reactivate Olares after the PC restarts?
Run the following command in PowerShell to restart the Olares service:
```powershell
wsl -d Ubuntu
```

### How to disable Windows Defender Firewall?
:::tip
You can turn on Windows Defender Firewall when the Olares installation completes.
:::
To completely disable the firewall:
1. Open **Control Panel** > **System and Security** > **Windows Defender Firewall**.
2. In the navigation pane, click **Turn Windows Defender Firewall on or off**.
3. Select **Turn off Windows Defender Firewall** for both private and public networks, then click **OK**.

   ![Turn off Windows Defender Firewall](/images/manual/get-started/disable-firewall.png#bordered)

### How to manually set firewall rules?
If you choose not to configure firewall rules during installation, follow these steps to set them manually:
1. Open **Control Panel** > **System and Security** > **Windows Defender Firewall**.

   ![Navigate to Windows Defender Firewall](/images/manual/get-started/select-firewall.png#bordered)

2. In the navigation pane, select **Advanced settings**.

   ![Select Advanced settings](/images/manual/get-started/select-advanced-settings.png#bordered)
3. In the navigation pane, right-click **Inbound Rules** and select **New Rule**.

   ![Add new rule](/images/manual/get-started/add-new-rule.png#bordered)
4. In the **New Inbound Rule Wizard**, select **Port** and click **Next**.

   ![Select Port](/images/manual/get-started/select-port.png#bordered)
5. In **Specific local ports**, enter `80`, `443`, `30180`, and click **Next**.

   ![Specify Port](/images/manual/get-started/specify-port.png#bordered)
6.  Select **Allow the connection** and click **Next**.

   ![Allow the connection](/images/manual/get-started/allow-the-connection.png#bordered)

7. Confirm the rules apply to **Domain**, **Private**, and **Public**, then click **Next**.

   ![Confirm rules](/images/manual/get-started/confirm-rules.png#bordered)
8. Provide a name for the rule and click **Finish**.

   ![Name the rule](/images/manual/get-started/name-the-rule.png#bordered)

### How to uninstall Olares?
Run the following command in PowerShell:
```powershell
wsl --unregister ubuntu
```