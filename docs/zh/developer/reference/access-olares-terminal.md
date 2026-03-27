---
outline: [2, 3]
description: 通过控制面板或 SSH 访问 Olares 主机终端。
---
# 访问 Olares 主机终端

部分开发和运维任务需要在 Olares 主机上运行命令，例如检查磁盘、验证主机状态或更新主机级配置。由于 Olares 主机通常不直接连接显示器或键盘，终端访问主要通过远程方式进行。

你可以通过以下两种方式访问主机终端：

- **控制面板终端**：控制面板提供的网页终端，可直接以 `root` 身份访问，适合快速执行命令或处理临时任务。
- **Secure Shell (SSH)**：远程连接和管理主机的标准协议，更适合自动化或进阶操作。

:::tip Olares One 用户
如果你使用的是 Olares One，可查看 [SSH 访问 Olares One](/zh/one/access-terminal-ssh.md)。
:::

## 通过控制面板访问

如需在不配置 SSH 客户端的情况下快速访问，可使用控制面板内置的网页终端。

1. 打开控制面板。
2. 在左侧边栏的**终端**部分，点击 **Olares**。
  ![终端](/images/zh/manual/developer/controlhub-terminal.png#bordered)

:::tip 以 `root` 权限运行
通过控制面板打开的终端默认以 `root` 身份运行。请勿在命令前使用 `sudo`。
:::

## 通过 SSH 访问

SSH 会通过网络建立加密会话，用于从电脑远程连接到 Olares 主机并执行命令。

### 通过局域网连接

连接前，需确保已掌握以下信息：
- Olares 主机本地 IP 地址。
- Olares 主机用户名和密码。

1. 打开电脑上的终端。
2. 使用以下格式运行 SSH 命令：

   ```bash
   ssh <username>@<host_ip_address>
   ```

   示例：
   ```bash
   ssh olares@192.168.31.155
   ```
3. 根据提示输入主机密码。

### 从不同网络连接

如果你的电脑与 Olares 主机不在同一局域网，先启用 LarePass VPN，以建立到主机的安全连接。

连接前，请确保：

- 用于建立 VPN 连接的电脑已安装并登录 LarePass 桌面端。
- 已获取 Olares 主机用户名和密码。

1. 在 Olares 中，前往**设置** > **VPN**，并启用**允许通过 VPN 进行 SSH 连接**。
  ![允许通过 VPN 进行 SSH 连接](/images/zh/manual/developer/access-terminal-allow-vpn.png#bordered){width=90%}

2. 打开 LarePass 桌面客户端，点击左上角头像，打开**专用网络连接**开关。
3. 在 Olares 中，前往**设置** > **VPN** > **查看 VPN 连接状态**，找到主机条目，并记下以 `100.64` 开头的 IP 地址。
  ![查看 Tailscale IP](/images/zh/manual/developer/access-terminal-tailscale-ip.png#bordered){width=90%}

4. 打开电脑上的终端。 
5. 使用以下格式运行 SSH 命令：

    ```bash
    ssh <username>@<tailscale_ip_address>
    ```
    
    示例：
    ```bash
    ssh olares@100.64.0.1
    ```
6. 根据提示输入主机密码。