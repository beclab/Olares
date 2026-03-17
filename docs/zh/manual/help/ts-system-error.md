---
outline: [2,3]
description: 当 LarePass 的系统区域显示“系统错误”时，诊断并收集相关信息。
---
# LarePass 显示“系统错误”

当 LarePass 移动端的**系统**部分显示“系统错误”时，请参考本指南进行排查。出现此提示可能由多种底层原因导致，请先按照以下步骤收集诊断信息，然后将结果提供给 Olares 团队。

:::info
本文以 Olares One 为示例设备。如果你在自己的设备上安装 Olares，排查步骤基本相同，只是访问终端的方式可能有所不同。
:::

 ![系统错误](/images/zh/manual/help/ts-sys-err.png#bordered){width=90%}

## 适用情况

- LarePass 移动端的**系统**部分显示“系统错误”。
- Olares 桌面可能无法访问。

## 原因

“系统错误”通常意味着一个或多个系统 Pod 运行异常。发生这种情况时，LarePass 无法获取整体系统状态。

## 解决方案

按照以下步骤访问 Olares 设备终端，定位未正常运行的 Pod，查看其错误详情，并将这些信息提供给 Olares 团队。这有助于缩小可能原因范围，加快故障排查。

### 步骤 1：尝试访问 Olares 桌面

如果你仍然可以访问 Olares 桌面，打开控制面板并使用 Olares 内置终端。

1. 打开浏览器，登录你的 Olares 桌面：

    ```text
    https://desktop.<username>.olares.cn
    ```

2. 打开控制面板。
3. 在左侧边栏的**终端**部分，点击 **Olares**。
    ![打开终端](/images/zh/manual/help/ts-sys-err-terminal.png#bordered){width=90%}

如果你可以成功访问终端，跳转至[步骤 4](#步骤-4-检查系统-pod-状态)。

### 步骤 2: 尝试 SSH 连接

如果你无法访问 Olares 桌面，可以尝试 SSH 连接。

:::info 需处于同一网络
你的电脑和 Olares  One 必须连接到同一个本地网络。
:::

1. 获取 Olares One 的本地 IP 地址。

   a. 打开 LarePass 移动端，进入**设置** > **系统**，打开 **Olares 管理**页面。

   b. 点击 Olares One 设备卡片。

   c. 向下滚动至**网络**部分，记录**内网 IP**。

2. 在 Vault 中查看 SSH 密码。

   a. 在 LarePass 移动端点击 **Vault**。根据提示输入本地密码解锁。

   b. 点击左上角的 **Vault** 打开侧边导航，然后点击**所有 Vault** 显示所有已保存条目。

   c. 找到带有 <span class="material-symbols-outlined">terminal</span> 图标的条目，点击查看密码。

      ![在 Vault 中查看保存的 SSH 密码](/public/images/zh/manual/olares/ssh-check-password-in-vault1.png#bordered)

3. 在你的电脑上打开终端，通过 SSH 连接设备。

    a. 输入以下命令，将 `<local_ip_address>` 替换为此前获取的内网 IP：

      ```bash
      ssh olares@<local_ip_address>
      ```
    b. 根据提示输入 SSH 密码。

如果连接成功，跳转至[步骤 4](#步骤-4-检查系统-pod-状态)。

### 步骤 3: 本地登录设备

如果无法通过 SSH 访问，使用显示器和键盘在本地登录设备。

1. 将显示器和键盘连接至 Olares One。屏幕上会有一行文字提示登录。

   ```text
   olares login:
   ```

2. 输入用户名 `olares` 并按回车键。
3. 输入[步骤 2](#步骤-2-尝试-ssh-连接) 中获取的设备登录密码并按回车键。

### 步骤 4: 检查系统 Pod 状态

1. 运行以下命令，查看所有命名空间下的 Pod 状态：
    ```bash
    kubectl get pods -A
    ```
2. 查看 **STATUS** 列，找到状态不是 `Running` 的 Pod。
3. 准确记录每个异常 Pod 的 **NAMESPACE** 和 **NAME**。
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
    ![Pod 错误详情](/images/zh/manual/help/ts-sys-err-pod-event-detail.png#bordered){width=90%}

### 步骤 6：联系技术支持

在 [Olares GitHub 仓库](https://github.com/beclab/Olares/issues)提交 Issue，并提供以下信息：

- 每个异常 Pod 对应的 `kubectl describe pod <pod-name> -n <namespace>` 完整输出结果
- 错误信息的截图（如有）
- 错误最初出现的时间及简要说明（例如，是在更新后还是重启后出现的）

这些信息将帮助团队更快排查并解决问题。