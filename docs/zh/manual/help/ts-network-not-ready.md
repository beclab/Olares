---
outline: [2, 3]
description: 排查 Olares One 设备已连接网络但无法通过标准访问方式连接的问题。
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, Network not ready, 连接错误, SSH, Pod 状态, 网络排查
---

# 网络尚未准备好或 olares 连接错误

当你的 Olares One 设备已开机并连接至网络，但突然停止响应时，可参考本指南进行排查。

## 适用情况

- LarePass 移动端显示 **Network not ready**，访问 Olares 桌面时出现 **olares connection error**，但路由器显示设备已连接，且设备能响应网络 `ping` 命令。
- 重启设备和路由器后，问题依然存在。

## 原因

这些现象无法直接指向唯一原因。`ping` 成功只表示设备能够响应基础网络请求。系统服务未启动、部分 Pod 异常、资源不足，或访问链路故障，都可能导致 Olares 仍然无法访问。

## 解决方案

先通过可用方式进入主机，再判断问题来自 Olares 内部服务还是网络访问链路。

### 步骤 1：尝试 SSH 连接

建议优先尝试此方法，因为这是访问设备并收集诊断信息最便捷的方式。

1. 获取 Olares One 的局域网 IP 地址。

   a. 打开 LarePass 移动端，进入**设置** > **系统**，打开 **Olares 管理**页面。

   ![访问 Olares 管理](/images/zh/manual/larepass/system.png#bordered)

   b. 点击 Olares One 设备卡片。

   c. 向下滚动至**网络**部分，记录**内网 IP**。

2. 在 Vault 中查看 SSH 密码。

   a. 在 LarePass 移动端中点击 **Vault**。根据提示输入本地密码解锁。

   b. 点击左上角的 **Vault** 打开侧边导航，然后点击**所有 Vault** 显示所有已保存条目。

   c. 找到带有 <span class="material-symbols-outlined">terminal</span> 图标的条目，点击查看密码。

      ![在 Vault 中查看保存的 SSH 密码](/images/zh/manual/olares/ssh-check-password-in-vault1.png#bordered)

3. 通过 SSH 连接。

   a. 在电脑上打开终端。

   b. 输入以下命令，将 `<local_ip_address>` 替换为内网 IP，然后按回车键：

      ```bash
      ssh olares@<local_ip_address>
      ```

   c. 根据提示输入 SSH 密码，然后按回车键。

   d. 如果连接成功，直接跳转至[步骤 3](#步骤-3-检查系统-pod-状态)。

### 步骤 2：本地登录设备

如果无法通过 SSH 访问，使用显示器和键盘在本地登录设备。

1. 将显示器和键盘连接至 Olares One。屏幕上会自动显示基于文本的登录提示窗口：

   ```text
   olares login:
   ```

2. 输入用户名 `olares` 并按回车键。
3. 输入**步骤 1** 中获取的 SSH 密码并按回车键。

### 步骤 3：检查系统 Pod 状态

1. 登录成功后，运行以下命令，查看所有命名空间下的 Pod 状态：

   ```bash
   kubectl get pods -A
   ```

2. 查看 **STATUS** 列，并根据结果继续：

   - 如果 Pod 显示 `CrashLoopBackOff`、`Error`、`ImagePullBackOff` 等错误状态，或长时间处于 `Pending`，只记录对应的 **NAMESPACE**、**NAME**、**STATUS** 和 **RESTARTS**。任务 Pod 显示 `Completed` 本身不代表异常。
   - 如果没有 Pod 显示错误，且重启次数没有持续增加，问题更可能出在访问或网络链路。分别记录本地访问、Olares 域名和 LarePass 专用网络是否可用。

3. 记录检查时间、时区和当前 Olares 版本。
4. 按照[收集诊断信息](../collect-diagnostic-information.md)生成日志压缩包，并通过非公开渠道发送。

:::warning 不要公开完整日志
Pod 输出和系统日志可能包含 Olares ID、主机名、IP 地址、域名和应用元数据。公开 GitHub Issue 中可以描述现象并提供上面列出的有限字段，但不要附上完整命令输出或日志压缩包。
:::
