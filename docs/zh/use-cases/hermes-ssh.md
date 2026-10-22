---
outline: [2, 3]
description: 使用 Hermes Agent 通过 SSH 管理局域网 Linux 主机。配置密钥登录，检查单台或多台主机的磁盘空间，并排查连接问题。
head:
  - - meta
    - name: keywords
      content: Olares, Hermes Agent, AI 助手, SSH, 局域网主机管理, 密钥认证, 磁盘空间
app_version: "1.3.49"
doc_version: "1.0"
doc_updated: "2026-10-21"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/hermes-ssh.md)为准。
:::

# 使用 Hermes Agent 通过 SSH 管理局域网主机

Hermes Agent 是一个能够执行任务的 AI 助手，可以通过 SSH 在局域网主机上运行命令和脚本。你可以在对话中指定目标主机和任务，可执行的操作取决于远端账号的权限。

本教程以检查 Linux 主机的磁盘空间为例，先连接一台主机，再将任务扩展到第二台。

## 学习目标

- 配置 SSH 密钥登录，并核对目标主机的身份。
- 让 Hermes 检查单台或多台主机的磁盘空间。
- 排查 SSH 连接和权限问题。

## 前提条件

- [Hermes Agent 已在 Olares 上安装并连接模型](hermes.md#配置-hermes-agent)。
- 目标 Linux 主机已启用 SSH 服务，并允许你使用指定账号登录。
- Olares 设备能访问目标主机的 SSH 端口，默认端口为 `22`。

## 准备连接

将终端后端保持为 **Local**，与 [Hermes 设置向导](hermes.md#第-2-步-运行设置向导)中的配置一致。Hermes 会在应用容器中运行标准 `ssh` 命令，连接目标主机。无需开启 Gateway API，也无需在 Hermes 配置中固定目标地址。

下面以 Linux 主机 `ubuntu@192.168.1.50` 为例。请将用户名、IP 地址和端口替换为你自己的配置。

:::tip 检查 Olares 的网络连通性
用于打开 Hermes 的电脑能访问目标主机，并不代表 Olares 也能访问。跨网段路由、访客网络隔离或防火墙规则都可能阻止连接。
:::

## 配置密钥登录

自动化任务使用密钥登录。完成公钥授权后，Hermes 无需输入账号密码即可连接目标主机。以下步骤会生成一对专用密钥。如果你已有密钥，可参阅[使用已有私钥](#existing-ssh-key)。

### 第 1 步：在持久化目录中准备密钥

1. 从启动台打开 Hermes CLI。如果当前处于聊天界面，输入 `/exit` 返回命令行。

2. 创建 SSH 目录，并生成一对密钥：

   ```bash
   mkdir -p /opt/data/.ssh
   chmod 700 /opt/data/.ssh
   ssh-keygen -t ed25519 -C "hermes-lan" -f /opt/data/.ssh/id_ed25519
   ```

   如果已经存在同名密钥，请保留原密钥，或换一个文件名，不要覆盖。自动化任务需要使用无需交互输入口令的密钥。生成专用自动化密钥时，可在两次口令提示中直接按 **Enter**，将口令留空。

3. 设置私钥权限，并查看公钥：

   ```bash
   chmod 600 /opt/data/.ssh/id_ed25519
   cat /opt/data/.ssh/id_ed25519.pub
   ```

   复制输出的完整公钥行。私钥保留在应用中，不要粘贴到聊天消息里。

`/opt/data` 是 Hermes 的持久化数据目录，应用重启或容器重建后，目录中的密钥仍会保留。卸载应用或删除应用数据前，请先备份密钥。克隆应用使用独立的数据目录，需要单独配置 SSH 密钥。

### 第 2 步：在目标主机上添加公钥 {#authorize-ssh-key}

1. 在目标主机上，以 Hermes 将要使用的账号登录，例如 `ubuntu`。打开终端，准备该账号的 SSH 目录：

   ```bash
   mkdir -p "$HOME/.ssh"
   chmod 700 "$HOME/.ssh"
   touch "$HOME/.ssh/authorized_keys"
   chmod 600 "$HOME/.ssh/authorized_keys"
   ```

2. 用文本编辑器打开 `$HOME/.ssh/authorized_keys`，将上一步复制的公钥追加为单独一行，然后保存。保留文件中原有的公钥，并确认目录和文件归当前登录账号所有。

3. 查看目标主机的 SSH 主机密钥指纹，供首次连接时核对：

   ```bash
   ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub
   ```

### 第 3 步：验证 SSH 连接 {#check-ssh-connection}

1. 返回 Hermes CLI，先手动连接一次：

   ```bash
   ssh -i /opt/data/.ssh/id_ed25519 \
     -o IdentitiesOnly=yes \
     -o UserKnownHostsFile=/opt/data/.ssh/known_hosts \
     ubuntu@192.168.1.50 'hostname; id'
   ```

   首次连接会提示确认主机指纹。与目标主机上显示的指纹核对一致后，再输入 `yes`。如果 SSH 使用其他端口，在命令中加入 `-p 端口号`。

2. 再验证不需要交互输入的连接方式：

   ```bash
   ssh -i /opt/data/.ssh/id_ed25519 \
     -o IdentitiesOnly=yes \
     -o BatchMode=yes \
     -o StrictHostKeyChecking=yes \
     -o UserKnownHostsFile=/opt/data/.ssh/known_hosts \
     -o ConnectTimeout=10 \
     ubuntu@192.168.1.50 'hostname; id'
   ```

   命令应直接返回目标主机名和登录账号信息，不再要求输入密码。`BatchMode=yes` 会在无法自动认证时直接返回错误，避免任务停在密码提示处。

私钥和 `known_hosts` 都使用绝对路径，避免助手运行命令时因主目录不同而找不到文件。

## 让 Hermes 检查磁盘空间

打开 Dashboard，或在 Hermes CLI 中运行 `hermes` 进入 TUI 聊天界面。将以下提示词中的示例账号和地址替换为你自己的配置：

```text
通过 SSH 检查 ubuntu@192.168.1.50 的磁盘空间。
使用私钥 /opt/data/.ssh/id_ed25519 和主机密钥文件
/opt/data/.ssh/known_hosts，保留严格主机密钥校验，
使用批处理模式，连接超时设为 10 秒。
先确认目标主机名和登录账号，再运行 df -h。
汇总每个文件系统的容量、已用空间、可用空间和使用率。
只检查磁盘使用情况，不删除文件或修改设置。
遇到权限不足或连接失败时，说明原因。
```

查看命令输出和汇总结果，了解哪些文件系统的剩余空间较少。

要检查第二台主机，先在该主机上完成公钥授权和首次连接验证。下面以 `olares@192.168.1.60` 为例，两台主机使用同一私钥。你也可以使用不同的私钥，但需要为每台主机指定对应的路径。

然后在同一次对话中发送以下提示词，将两个账号和地址替换为你自己的配置：

```text
依次检查 ubuntu@192.168.1.50 和 olares@192.168.1.60 的磁盘空间。
两台主机都使用私钥 /opt/data/.ssh/id_ed25519 和主机密钥文件
/opt/data/.ssh/known_hosts。保留严格主机密钥校验，
使用批处理模式，连接超时设为 10 秒。
分别确认主机名和登录账号，再运行 df -h，不修改任何文件。
按主机名和 IP 地址汇总结果，列出每个文件系统的容量、
已用空间、可用空间和使用率。
某台主机失败时，报告错误和已完成的检查，再询问我是否继续。
```

执行其他任务时，请说明要运行的命令或脚本，以及执行顺序。如果使用脚本，还需指定脚本所在的主机和路径。

:::info 远端 sudo 权限
远端命令的权限取决于 SSH 登录账号。需要提权时，可以要求助手使用 `sudo -n`。账号必须有权执行对应命令。如果需要输入密码，`sudo -n` 会直接返回错误，不会等待输入。不要把 sudo 密码写进提示词。远端 sudo 权限与 Hermes 容器内的 sudo 权限相互独立。
:::

## 可选：使用已有私钥 {#existing-ssh-key}

你可以使用已有私钥，包括 PEM 私钥，无需生成新密钥。SSH 不要求私钥文件使用 `.pem` 扩展名。用于自动化任务的密钥必须无需交互输入口令即可使用。

1. 在 Hermes CLI 中准备持久化的 SSH 目录：

   ```bash
   mkdir -p /opt/data/.ssh
   chmod 700 /opt/data/.ssh
   ```

2. 在文件管理器中进入**数据** > **hermesagent** > **home**，显示隐藏文件，然后将私钥上传到 `.ssh` 目录，命名为 `lan-host.pem`。不要把私钥粘贴到聊天消息里。

3. 在 Hermes CLI 中设置私钥权限，并查看对应的公钥：

   ```bash
   chmod 600 /opt/data/.ssh/lan-host.pem
   ssh-keygen -y -f /opt/data/.ssh/lan-host.pem
   ```

4. 如果目标账号尚未授权此密钥，复制输出的公钥，按照[在目标主机上添加公钥](#authorize-ssh-key)中的步骤完成授权。

5. 按照[验证 SSH 连接](#check-ssh-connection)中的步骤检查连接，将命令中的 `/opt/data/.ssh/id_ed25519` 替换为 `/opt/data/.ssh/lan-host.pem`。任务提示词中的路径也需要相应替换。

## 可选：使用密码手动登录

目标主机允许密码认证时，你也可以使用密码建立手动 SSH 会话。这是密钥登录之外的另一种方式，不是让助手执行任务的前提。

1. 从启动台打开 Hermes CLI。如果当前处于聊天界面，输入 `/exit` 返回命令行。

2. 准备持久化的 SSH 目录，然后发起连接：

   ```bash
   mkdir -p /opt/data/.ssh
   chmod 700 /opt/data/.ssh
   ssh -o PubkeyAuthentication=no \
     -o PreferredAuthentications=password,keyboard-interactive \
     -o UserKnownHostsFile=/opt/data/.ssh/known_hosts \
     ubuntu@192.168.1.50
   ```

   如果使用其他 SSH 端口，加入 `-p 端口号`。首次连接时，先向目标主机管理员核对主机指纹，确认一致后再输入 `yes`。

3. 出现 `ubuntu@192.168.1.50's password:` 提示时，输入远端账号的密码并按 **Enter**。输入内容不会显示。登录成功后，可以执行该账号有权限运行的命令。完成后输入 `exit` 退出远端会话。

普通 SSH 命令不会自动回答后台任务中的密码提示。需要让助手独立执行任务时，请使用密钥登录。不要把密码写进命令、脚本或聊天提示词。

## 常见连接问题

| 现象 | 检查方法 |
|:---|:---|
| 连接超时或被拒绝 | 检查 IP、SSH 端口、服务状态、防火墙，以及 Olares 到目标主机的网络路由。开启 Olares 应用入口不能解决出站 SSH 的连通性问题。 |
| `Permission denied (publickey)` | 核对远端用户名、该账号 `authorized_keys` 中的公钥和私钥路径。私钥权限应为 `600`，`.ssh` 目录权限应为 `700`。 |
| `Host key verification failed` | 使用同一个 `known_hosts` 路径手动完成首次连接。如果已保存的指纹发生变化，先在目标主机上确认原因，再更新记录，不要关闭指纹校验。 |
| `sudo: a password is required` | 当前账号没有免密码执行该命令的权限。改用账号有权执行的只读命令，或请目标主机管理员配置所需权限。 |

## 了解更多

- [配置 Hermes Agent](hermes.md)：连接模型、使用 Olares 技能，并将助手接入其他应用。
