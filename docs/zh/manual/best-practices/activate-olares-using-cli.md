---
outline: [2, 3]
description: 如何使用 Olares CLI 安装和激活 Olares 设备。
---

# 使用 Olares CLI 激活 Olares 设备

本教程介绍如何使用 Olares CLI 工具激活 Olares 设备（例如 Olares One）。以下步骤假定设备是全新开箱的，尚未安装或激活。

## 目标

通过本教程，你将学习：
- 下载并解压 Olares CLI 工具。
- 在新设备上安装 Olares。
- 获取用于远程访问的 Fast Reverse Proxy (FRP) 主机地址。
- 运行激活命令以配置设备。

## 前提条件

- **设备连接**：已通过键盘和屏幕直连 Olares 设备，或已通过 SSH 登录到该设备。
- **root 权限**：确保拥有 root 权限。所有安装和激活命令都需要在设备上使用 `sudo` 或 root 权限执行。
- **LarePass 准备**：已使用 LarePass 应用注册了 Olares ID。

    ![快速创建](/images/zh/manual/tutorials/create-olares-id.png)

## 步骤 1：下载并解压 CLI 工具

1. 下载 Olares CLI 安装包。

    ```bash
    curl -sSOL https://cdn.olares.com/common/olares-cli-amd64.8cbdc32.tar.gz
    ```

2. 解压已下载的文件。

    ```bash
    tar xzf olares-cli-amd64.8cbdc32.tar.gz
    ```

## 步骤 2：获取 FRP 列表

查找可用的 FRP 主机，以启用对设备的远程访问。

1. 运行以下命令，将 `{olares-id}` 替换为已注册的 Olares ID。

    ```bash
    olares-cli wizard frp {olares-id}
    ```

    **示例**：

    ```bash
    olares-cli wizard frp alice123@olares.com
    ```

2. 从输出列表中选择一个主机地址，并保存以备激活步骤使用。

## 步骤 3：安装 Olares

:::info 关于 Olares One 硬件
全新开箱的 Olares One 设备是处于未安装状态的。在尝试激活之前，必须首先运行安装命令来设置 Olares。
:::

1. 运行安装命令。

    ```bash
    sudo olares-cli install
    ```

2. 等待安装过程完成。终端会输出一个本地网关地址和一个默认密码。保存这些信息，用于激活步骤。

    **示例**：
    - **Wizard URL**：本地网关地址，例如 `http://192.168.50.123:30180`。
    - **密码**：Olares 的默认登录密码。

    ![Wizard URL](/images/manual/get-started/wizard-url-and-login-password.png)

## 步骤 4：激活 Olares

运行激活命令以配置和保护设备。此过程会将 Olares ID 与设备关联，并配置网络隧道和凭证。

1. 根据以下必需参数准备激活命令：

    | 参数 | 说明 |
    |:----------|:------------|
    | `olares-id` | Olares 生态系统中的唯一标识符。注册后可在 LarePass 应用中找到。 |
    | `mnemonic` | LarePass 应用中的助记词。 |
    | `password` | 步骤 3 中生成的 Olares 默认登录密码。 |
    | `reset-password` | 为 Olares 指定一个新的登录密码。 |
    | `authurl` | 输入步骤 3 中生成的 `wizard-url`。 |
    | `vault` | 输入 `wizard-url` 并加 `/server`后缀。 |
    | `bfl` | 输入步骤 3 中生成的 `wizard-url`。 |
    | `host` | 输入步骤 2 中选择的 FRP 主机地址。 |
    | `enable-tunnel` | 设置为 `true` 以启用隧道模式进行激活。 |

2. 将以下命令中的占位符替换为实际值，然后运行。

    ```bash
    sudo olares-cli wizard activate {olares-id} \
    --mnemonic "{mnemonic}" \
    --password="{password}" \
    --reset-password="{reset-password}" \
    --authurl={authurl} \
    --vault={vault} \
    --bfl={bfl} \
    --host={host} \
    --enable-tunnel=true
    ```

    **示例**：
    
    如果 Olares ID 是 `alice123@olares.com`，Wizard URL 是 `http://192.168.50.123:30180`，选择的 FRP 主机是 `bb.hongkong.frp.olares.com`，则运行：

    ```bash
    sudo olares-cli wizard activate alice123@olares.com \
    --mnemonic "abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef" \
    --password="v0kSmyVN" \
    --reset-password="Ab1234@" \
    --authurl=http://192.168.50.123:30180 \
    --vault=http://192.168.50.123:30180/server \
    --bfl=http://192.168.50.123:30180 \
    --host=bb.hongkong.frp.olares.com \
    --enable-tunnel=true
    ```

## 后续步骤

现在你可以使用 Olares ID 和在 `reset-password` 中指定的登录密码登录 Olares。

## 了解更多

- [创建 Olares ID](/zh/manual/get-started/create-olares-id.md)
- [Olares CLI](/../zh/developer/install/cli/olares-cli.md)