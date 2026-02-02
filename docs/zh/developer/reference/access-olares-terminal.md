---
outline: [2, 3]
description: 通过 SSH 或 Control Hub 访问 Olares 主机终端的完整指南。
---
# 访问 Olares 主机终端

某些开发和运维任务需要在 Olares 主机上运行命令，例如检查磁盘、验证主机状态或更新主机级配置。由于 Olares 主机通常以无显示器、无键盘的方式部署，因此终端访问主要通过远程方式进行。

你可以通过以下两种方式访问主机终端：

- **Secure Shell（SSH）**：标准的远程管理。
- **控制面板**：通过 Olares 网页端直接获取 root 权限。

## 方式一：通过 SSH 访问

SSH 是从远程开发设备操作 Olares 主机的标准协议，可在网络上建立安全的终端会话。

### 前提条件

连接之前，请确保满足以下条件：

- 与 Olares 主机建立网络连接：  
  - 在大多数情况下，你的电脑和 Olares 主机位于同一局域网内。  
  - 如果需要从局域网外部连接，请先配置 VPN 访问。详见[通过 LarePass VPN 连接](#可选-通过-larepass-vpn-连接)。
- Olares 主机 IP 地址，通常为 `192.168.x.x`。
- 有效的登录凭证。

### 通过局域网连接

1. 打开计算机上的终端。
2. 使用以下格式运行 SSH 命令：

   ```bash
   ssh <username>@<host_ip_address>
   ```

   示例：
   ```bash
   ssh olares@192.168.31.155
   ```
3. 根据提示输入主机的登录密码。

### 可选：通过 LarePass VPN 连接

如果你的电脑与 Olares 主机不在同一局域网，执行 SSH 命令之前先启用 LarePass VPN。

1. 在 Olares 中，打开设置应用。
2. 启用**允许通过 VPN 进行 SSH 连接**。 
3. 打开 LarePass 桌面客户端，点击左上角头像打开用户菜单。 
4. 打开**专用网络连接**开关。 
5. 打开计算机上的终端。 
6. 使用以下格式运行 SSH 命令：

    ```bash
    ssh <username>@<host_ip_address>
    ```
    
    示例：
    ```bash
    ssh olares@192.168.31.155
    ```
7. 根据提示输入主机密码。

## 方式二：通过控制面板访问

你可以直接通过 Olares 控制面板打开主机终端。

1. 打开控制面板。
2. 在左侧边栏的**终端**部分，点击 **Olares**。

:::info `root` 权限
通过控制面板打开的终端默认以 `root` 身份运行。

请勿在命令前使用 `sudo`。例如，直接运行 `apt update`，而不是 `sudo apt update`。
:::