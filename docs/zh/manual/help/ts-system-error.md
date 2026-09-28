---
outline: [2,3]
description: 当 LarePass 显示“系统错误”时，通过 Ticket 收集日志并获取帮助，也可使用终端进一步检查系统 Pod。
head:
  - - meta
    - name: keywords
      content: Olares, LarePass, 系统错误, Pod 状态, SSH, 故障排查, kubectl
---
# LarePass 显示“系统错误”

当 LarePass 移动端的**系统**部分显示“系统错误”时，参考本指南进行排查。你可以通过 Ticket 收集日志并联系 Olares 团队，也可以按下方的进阶排查步骤进一步定位原因。

:::warning 不要卸载 LarePass 或 Olares OS
不要卸载 LarePass，也不要进入“系统错误”页面卸载 Olares OS 或恢复出厂设置。卸载 Olares OS 或恢复出厂设置可能清除数据。
:::

 ![系统错误](/images/zh/manual/help/ts-sys-err.png#bordered){width=90%}

## 适用情况

- LarePass 移动端的**系统**部分显示“系统错误”。
- Olares 桌面可能无法访问。

## 原因

该提示表示 LarePass 未能获取到健康的系统状态。一个或多个系统 Pod 可能处于异常状态，可以通过系统日志和 Pod 事件进一步定位原因。

## 解决方案

### 收集日志并获取帮助

Olares 1.12.7 及以上版本可通过 Ticket 应用提交支持工单。在**系统日志**中点击**采集日志**，自动收集并附加日志。前提条件和操作步骤见[通过 Ticket 应用提交](request-technical-support.md#通过-ticket-应用提交)。

如果无法使用 Ticket 应用，但可以访问设备终端，可运行 Olares Space 提供的命令，自动收集、上传日志并创建工单。操作步骤见[通过 Olares CLI 自动创建工单](../space/tickets.md#通过-olares-cli-自动创建工单)。

在工单中附上“系统错误”截图、出现时间和时区、Olares 版本，并说明近期是否更新或重启过设备。

### 进阶排查（可选）

通过设备终端检查系统 Pod（运行系统组件的容器组）的状态和错误事件，帮助 Olares 团队定位故障组件。

:::info
本文以 Olares One 为例。如果你在自己的设备上安装 Olares，排查步骤基本相同，只是访问终端的方式可能存在差异。
:::

#### 步骤 1：尝试访问 Olares 桌面

如果你仍然可以访问 Olares 桌面，打开控制面板并使用 Olares 内置终端。

1. 打开浏览器，登录你的 Olares 桌面：

    ```text
    https://desktop.<username>.olares.cn
    ```

2. 打开控制面板。
3. 在左侧边栏的**终端**部分，点击 **Olares**。
    ![打开终端](/images/zh/manual/help/ts-sys-err-terminal.png#bordered){width=90%}

如果你可以成功访问终端，跳转至[步骤 4](#步骤-4-检查系统-pod-状态)。

#### 步骤 2: 尝试 SSH 连接

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

      ![在 Vault 中查看保存的 SSH 密码](/images/zh/manual/olares/ssh-check-password-in-vault1.png#bordered)

3. 在你的电脑上打开终端，通过 SSH 连接设备。

    a. 输入以下命令，将 `<local_ip_address>` 替换为此前获取的内网 IP：

      ```bash
      ssh olares@<local_ip_address>
      ```
    b. 根据提示输入 SSH 密码。

如果连接成功，跳转至[步骤 4](#步骤-4-检查系统-pod-状态)。

#### 步骤 3: 本地登录设备

如果无法通过 SSH 访问，使用显示器和键盘在本地登录设备。

1. 将显示器和键盘连接至 Olares One。屏幕上会有一行文字提示登录。

   ```text
   olares login:
   ```

2. 输入用户名 `olares` 并按回车键。
3. 输入[步骤 2](#步骤-2-尝试-ssh-连接) 中获取的设备登录密码并按回车键。

#### 步骤 4: 检查系统 Pod 状态

1. 运行以下命令，查看所有命名空间下的 Pod 状态：
    ```bash
    kubectl get pods -A
    ```
2. 查看 **STATUS** 和 **RESTARTS**。重点检查 `CrashLoopBackOff`、`Error`、`ImagePullBackOff` 等错误状态，或长时间处于 `Pending` 的 Pod。任务 Pod 显示 `Completed` 表示已成功完成。
3. 记录显示错误或重启次数持续增加的 Pod 对应的 **NAMESPACE** 和 **NAME**。如果没有，直接跳到[步骤 6](#步骤-6-记录结果并收集日志)。
    ![定位异常 Pod](/images/zh/manual/help/ts-sys-err-pod-crash.png#bordered){width=90%}

#### 步骤 5：查看 Pod 错误信息

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

#### 步骤 6：提供排查结果 {#步骤-6-记录结果并收集日志}

将以下信息补充到支持工单中：

- 异常 Pod 的 **NAMESPACE**、**NAME**、**STATUS** 和 **RESTARTS**
- **Events** 部分中的错误行
- 报错时间和时区
- 当前 Olares 版本，以及错误是否出现在更新或重启之后

如果所有 Pod 状态正常，且重启次数保持稳定，也请在工单中说明，帮助 Olares 团队缩小排查范围。
