---
description: 本页面记录了使用 Olares One 时可能遇到的已知问题和异常行为，以及相应的解决方案或变通方法。
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/known-issues.md)。
:::

# Olares One 已知问题

使用本页面识别和排查 Olares One 的已知问题。我们会定期更新此列表，提供临时变通方法和永久修复方案。

## Olares One 初始设置在 9% 处失败

Olares One 在初始设置过程中失败，安装停在约 9% 处，并提示你卸载或重新安装。

启动时，系统在签发安全证书之前会执行异步 NTP 时间同步。虽然这通常瞬间完成，但偶尔延迟可能导致证书被签发为未来时间戳。如果设备尚未从默认出厂时区 UTC+8 更新，这种情况尤其常见，最终导致激活失败。

### 变通方法

卸载未完成的安装，然后重新激活设备。

#### 步骤 1：尝试 SSH 连接

如果你尚未将显示器和键盘连接到 Olares 设备，请先尝试此方法。

1. 从 LarePass 应用上的 **Activate Olares** 页面获取 Olares One 的本地 **IP** 地址。
    ![在 Activate Olares 上显示的 IP 地址](/images/one/obtain-ip-from-install.png#bordered)

2. 在你的电脑上打开终端。
3. 输入以下命令，将 `<local_ip_address>` 替换为上述本地 IP 地址，然后按 **Enter**：

    ```bash
    ssh olares@<local_ip_address>
    ```
4. 按提示输入默认 SSH 密码 `olares`，然后按 **Enter**。
5. 如果连接成功，跳转到[步骤 3](#步骤-3-运行卸载命令)。

#### 步骤 2：本地登录

当 SSH 访问不可用时，使用显示器和键盘本地登录设备。

1. 将显示器和键盘连接到 Olares One。屏幕上会自动显示基于文本的登录提示：

    ```text
    olares login:
    ```

2. 输入用户名 `olares`，按 **Enter**。
3. 按提示输入默认 SSH 密码 `olares`，然后按 **Enter**。

#### 步骤 3：运行卸载命令

1. 登录后，输入以下命令并按 **Enter**。此命令将删除所有已安装的组件和数据，将设备恢复到未激活状态。

    ```bash
    sudo olares-cli uninstall
    ```
2. 等待卸载完成。

#### 步骤 4：使用 LarePass 重新安装并激活

:::tip 重新安装前
为确保时间同步准确，让设备保持通电几分钟，以便自动校准内部时间。
:::

1. 根据你的网络设置发现并关联你的 Olares One。

    <tabs>
    <template #通过有线-LAN-设置>

    a. 确保 Olares One 通过网线连接到路由器。

    b. 在 LarePass 应用中，点击 **Discover nearby Olares**。
    ![发现附近的 Olares](/images/one/discover-nearby-olares.png#bordered){width=90%}

    </template>

    <template #通过-Wi-Fi-（蓝牙）设置>
    如果没有有线连接，可以使用蓝牙配置 Wi-Fi 凭据。

    a. 在 LarePass 应用中，点击 **Discover nearby Olares**。

    b. 点击底部的 **Bluetooth network setup**。

    c. 从蓝牙列表中选择你的设备，点击 **Network setup**。

    d. 按照提示将 Olares One 连接到你的手机当前使用的 Wi-Fi 网络。

    e. 连接完成后，返回主屏幕，再次点击 **Discover nearby Olares**。

    </template>
    </tabs>

2. 从可用设备列表中找到你的 Olares One，然后点击 **Install now**。安装现在应该能够顺利进行并成功完成。

## Thunderbolt 5 端口在 Windows 设备管理器中显示为 USB4

在双启动方案中安装 Windows 11 时，Thunderbolt 5 端口在设备管理器中显示为 "USB4(TM) Host Router" 或 "USB4 Root Router"。这是正常行为，不影响功能。

   ![设备管理器](/images/one/device-manager.png#bordered){width=50%}

此显示行为是 Intel 和 Microsoft 的设计。从 Windows 11 开始，Thunderbolt 驱动已内置到操作系统中，并通过 `Usb4HostRouter.sys` 驱动统一管理，该驱动统一处理 Thunderbolt 3、4 和 5 设备。设备管理器中的 USB4 标签反映了这种统一架构，并不表示端口以降速运行。

当连接 Thunderbolt 5 设备时，系统仍会协商完整的 80 Gbps 带宽，Thunderbolt 5 扩展坞也能正常工作。你可以在 USB4 设备属性页面的 **Current bandwidth (down/up)** 字段中验证连接速度。

更多信息请参阅：
- [Why Are There No Thunderbolt™ Drivers for Intel® NUC Products Using Windows 11](https://www.intel.com/content/www/us/en/support/articles/000094522/intel-nuc.html)
- [Introduction to the USB4 connection manager in Windows](https://learn.microsoft.com/en-us/windows-hardware/design/component-guidelines/usb4-intro-to-connection-manager)
