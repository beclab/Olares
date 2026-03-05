---
outline: [2,3]
description: 找出 LarePass 的 System 部分显示 “System error” 时，用于诊断问题原因。
---
# # LarePass 显示`系统错误`

当 LarePass 移动端的**系统**部分显示`系统错误`时，请参考本指南进行排查。

## 适用情况

- LarePass 移动端的**系统**部分显示`系统错误`。
- Olares 设备可以正常访问，但 LarePass 无法获取系统状态。

## 原因

Olares 设备中的一个或多个系统 Pod 未正常运行。发生这种情况时，LarePass 无法获取系统状态，因此会显示`系统错误`。

## 解决方案

使用系统内置终端，定位异常 Pod 并获取其错误信息，最后将相关信息提供给我们的支持团队。

### 步骤 1：定位异常 Pod

检查系统 Pod 的运行状态，找出未能正常运行的 Pod。

1. 打开浏览器，登录你的 Olares 桌面：`https://desktop.<your-olaresID>.olares.com`。
2. 打开控制面板，在左侧边栏的**终端**部分，点击 **Olares**。
    ![打开终端](/images/zh/manual/help/ts-sys-err-terminal.png#bordered){width=90%}

3. 运行以下命令，查看所有命名空间下的 Pod 状态：
    ```bash
    kubectl get pods -A
    ```
4. 查看 **STATUS** 列，找到状态不是 `Running` 的 Pod。请准确记录该异常 Pod 的 **NAMESPACE**（第一列）和 **NAME**（第二列）。
    ![Locate problematic pod](/images/zh/manual/help/ts-sys-err-pod-crash.png#bordered){width=90%}

### 步骤 2：查看 Pod 错误信息

查看该 Pod 的详细错误信息。

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

### 步骤 3：联系技术支持

请在 [Olares GitHub 仓库](https://github.com/beclab/Olares/issues)提交 Issue，并提供以下信息：

- `kubectl describe pod <pod-name> -n <namespace>` 命令的输出结果。
- 错误信息的截图（如有）。

这些信息将帮助我们的团队更快排查并解决问题。