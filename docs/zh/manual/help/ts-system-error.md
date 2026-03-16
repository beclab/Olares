---
outline: [2,3]
description: 当 LarePass 的系统区域显示“系统错误”时，诊断并收集相关信息。
---
# # LarePass 显示“系统错误”

当 LarePass 移动端的**系统**部分显示“系统错误”时，请参考本指南进行排查。该提示可能由多种底层原因触发，请先按照以下步骤收集诊断信息，然后将结果提供给 Olares 团队。
 ![系统错误](/images/zh/manual/help/ts-sys-err.png#bordered){width=90%}

## 适用情况

- LarePass 移动端的**系统**部分显示“系统错误”。
- LarePass 无法获取 Olares 设备的系统状态。
- Olares 桌面可能无法访问。

## 原因

“系统错误”提示可能由不同的底层问题触发。常见原因之一是 Olares 设备上的一个或多个系统 Pod 未能正常运行。发生这种情况时，LarePass 无法获取整体系统状态，因此会显示“系统错误”。某些情况下，同样的底层问题也可能导致**控制面板**无法访问。

## 解决方案

通过命令行定位未正常运行的 Pod，查看其错误详情，然后将这些信息提供给 Olares 团队。这有助于缩小可能原因范围，加快故障排查。

### 步骤 1：进入命令行

- 如果仍然可以在浏览器中打开 Olares，请按照[方案 A](#方案-a-从-olares-桌面进入命令行) 操作。
- 如果无法打开 Olares 桌面，请按照[方案 B](#方案-b-通过-ssh-进入命令行) 操作。

#### 方案 A：从 Olares 桌面进入命令行

1. 打开浏览器，登录你的 Olares 桌面：

    ```text
    https://desktop.<your-olaresID>.olares.cn
    ```

2. 打开控制面板，在左侧边栏的**终端**部分，点击 **Olares**。
    ![打开终端](/images/zh/manual/help/ts-sys-err-terminal.png#bordered){width=90%}

#### 方案 B：通过 SSH 进入命令行

如果 Olares 桌面已无法访问，请使用此方法。

:::warning
要通过 SSH 连接，确保你的电脑和 Olares 设备连接到同一个局域网。否则，SSH 连接会失败。
:::

1. 如果你已经知道内网 IP，请直接进入下一步。否则，可通过以下任一方式获取：

    <Tabs>
    <template #通过-LarePass-应用>

    1. 打开 LarePass 应用，进入**设置** > **系统**，打开 **Olares 管理**页面。
    2. 点击 Olares 设备卡片。
    3. 向下滚动到**网络**部分，记录**内网 IP**。

    </template>

    <template #通过显示器查看>

    1. 将 Olares 设备连接显示器和键盘。
    2. 打开终端，并运行 `ifconfig`。
    3. 找到当前正在使用的网络接口，通常是 `enp3s0`（有线）或 `wlo1`（无线）。IP 地址显示在 `inet` 后面。

    </template>

    </Tabs>

2. 运行以下命令，将 `<local_ip_address>` 替换为上一步获取到的内网 IP：

    ```bash
    ssh olares@<local_ip_address>
    ```

3. 如果系统提示确认连接，输入 `yes` 并按回车键。
4. 当系统提示输入 SSH 密码时，输入密码。
    :::tip
    如果没有修改过密码，默认 SSH 密码为 `olares`。
    :::

### 步骤2: 定位有问题的 Pod

1. 运行以下命令，查看所有命名空间下的 Pod 状态：
    ```bash
    kubectl get pods -A
    ```
2. 查看 **STATUS** 列，找到状态不是 `Running` 的 Pod。
3. 准确记录每个异常 Pod 的 **NAMESPACE**（第一列）和 **NAME**（第二列）。
    ![定位异常 Pod](/images/zh/manual/help/ts-sys-err-pod-crash.png#bordered){width=90%}

### 步骤 3：查看 Pod 错误信息

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

### 步骤 4：联系技术支持

请在 [Olares GitHub 仓库](https://github.com/beclab/Olares/issues)提交 Issue，并提供以下信息：

- `kubectl describe pod <pod-name> -n <namespace>` 命令的输出结果。
- 错误信息的截图（如有）。

这些信息将帮助我们的团队更快排查并解决问题。