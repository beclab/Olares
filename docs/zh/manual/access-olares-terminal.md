---
outline: [2, 3]
description: 通过 SSH 访问 Olares 主机终端。
---
# 访问 Olares 主机终端

部分开发和运维任务需要在 Olares 主机上运行命令，例如检查磁盘、验证主机状态或更新主机级配置。由于 Olares 主机通常不直接连接显示器或键盘，终端访问主要通过远程方式进行。

使用 SSH 从电脑连接到 Olares 主机。运行需要管理员权限的命令时，请使用具有 `sudo` 权限的账号。

:::tip Olares One 用户
如果你使用的是 Olares One，可查看 [SSH 访问 Olares One](/zh/one/access-terminal-ssh.md)。
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

如果你的电脑与 Olares 主机不在同一局域网，先启用 LarePass 专用网络，以建立到主机的安全连接。

连接前，请确保：

- 用于建立专用网络连接的电脑已安装并登录 LarePass 桌面端。
- 已获取 Olares 主机用户名和密码。

1. 在 Olares 中，前往**设置** > **VPN**，并启用**允许通过 VPN 进行 SSH 连接**。
  ![允许通过专用网络进行 SSH 连接](/images/zh/manual/developer/access-terminal-allow-vpn.png#bordered){width=90%}

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