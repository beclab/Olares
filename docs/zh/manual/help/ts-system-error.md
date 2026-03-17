---
outline: [2,3]
description: 当 LarePass 的系统区域显示“系统错误”时，诊断并收集相关信息。
---
# LarePass 显示“系统错误”

当 LarePass 移动端的**系统**部分显示“系统错误”时，请参考本指南进行排查。

本文以 Olares One 为示例设备。如果你使用的是其他 Olares 设备，也可以参考相同的排障流程。

 ![系统错误](/images/zh/manual/help/ts-sys-err.png#bordered){width=90%}

## 适用情况

- LarePass 移动端的**系统**部分显示“系统错误”。
- LarePass 无法获取 Olares 设备的系统状态。
- Olares 桌面可能无法访问。

## 原因

“系统错误”提示可能由不同的底层问题触发。常见原因之一是 Olares 设备上的一个或多个系统 Pod 未能正常运行。发生这种情况时，LarePass 无法获取整体系统状态，因此会显示“系统错误”。

## 解决方案

请按照以下步骤访问 Olares 设备终端，定位未正常运行的 Pod，查看其错误详情，并将这些信息提供给 Olares 团队。这有助于缩小可能原因范围，加快故障排查。

### 步骤 1：尝试访问 Olares 桌面

如果你仍然可以访问 Olares 桌面，请打开控制面板并使用 Olares 内置终端。

1. 打开浏览器，登录你的 Olares 桌面：

    ```text
    https://desktop.<your-olares-id>.olares.cn
    ```

2. 打开控制面板。
3. 在左侧边栏的**终端**部分，点击 **Olares**。
    ![打开终端](/images/zh/manual/help/ts-sys-err-terminal.png#bordered){width=90%}

如果你可以成功访问终端，直接跳转至[步骤 4](#步骤-4-检查系统-pod-状态)。

### 步骤 2: 尝试 SSH 连接

如果你无法访问 Olares 桌面，请先尝试 SSH 连接。

:::info 需处于同一网络
你的电脑和 Olares 设备必须连接到同一个本地网络。
:::

1. 获取 Olares 设备的本地 IP 地址。如果你无法获取本地 IP 地址，请继续完成下方获取 SSH 密码的步骤，然后前往**步骤 3**。

   a. 打开 LarePass 移动端，进入**设置** > **系统**，打开 **Olares 管理**页面。

   b. 点击 Olares One 设备卡片。

   c. 向下滚动至**网络**部分，记录**内网 IP**。

2. 在 Vault 中查看 SSH 密码。

   a. 在 LarePass 移动端点击 **Vault**。根据提示输入本地密码解锁。

   b. 点击左上角的 **Vault** 打开侧边导航，然后点击**所有 Vault** 显示所有已保存条目。

   c. 找到带有 <span class="material-symbols-outlined">terminal</span> 图标的条目，点击查看密码。

      ![在 Vault 中查看保存的 SSH 密码](/public/images/zh/manual/olares/ssh-check-password-in-vault1.png#bordered)

3. 通过 SSH 连接。

   a. 在电脑上打开终端。

   b. 输入以下命令，将 `<local_ip_address>` 替换为内网 IP，然后按回车键：

      ```bash
      ssh olares@<local_ip_address>
      ```

   c. 根据提示输入 SSH 密码，然后按回车键。

如果连接成功，直接跳转至[步骤 4](#步骤-4-检查系统-pod-状态)。

如果无法通过 SSH 连接，请继续[步骤 3](#步骤-3-本地登录设备)。

### 步骤 3: 本地登录设备

使用显示器和键盘，直接在设备本地登录。

1. 将显示器和键盘连接至 Olares One。屏幕上会自动显示基于文本的登录提示窗口：

   ```text
   olares login:
   ```

2. 输入用户名 `olares` 并按回车键。
3. 输入**步骤 2** 中获取的 SSH 密码并按回车键。

### 步骤 4: 检查系统 Pod 状态

1. 运行以下命令，查看所有命名空间下的 Pod 状态：
    ```bash
    kubectl get pods -A
    ```
2. 查看 **STATUS** 列，找到状态不是 `Running` 的 Pod。
3. 准确记录每个异常 Pod 的 **NAMESPACE**（第一列）和 **NAME**（第二列）。
    ![定位异常 Pod](/images/zh/manual/help/ts-sys-err-pod-crash.png#bordered){width=90%}

### 步骤 5：查看 Pod 错误信息

1. 运行以下命令，并将 `<namespace>` 和 `<pod-name>` 替换为上一步记录的值：

    ```bash
    kubectl describe pod <pod-name> -n <namespace>
    ```

    本例中，完整命令如下：

    ```bash
    kubectl describe pod backup-66f8c76996-d7vnq -n os-framework
    ```
2. 在输出结果中向下滚动到 **Events** 部分，查看失败相关的错误信息。
    ![Locate problematic pod](/images/zh/manual/help/ts-sys-err-pod-event-detail.png#bordered){width=90%}

### 步骤 6：联系技术支持

请在 [Olares GitHub 仓库](https://github.com/beclab/Olares/issues)提交 Issue，并提供以下信息：

- `kubectl describe pod <pod-name> -n <namespace>` 命令的输出结果。
- 错误信息的截图（如有）。

这些信息将帮助我们的团队更快排查并解决问题。