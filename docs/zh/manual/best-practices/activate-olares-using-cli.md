---
outline: [2, 3]
description: 如何使用 Olares CLI 安装和激活 Olares 设备。
---

# 使用 Olares CLI 激活 Olares 设备

使用 Olares CLI 工具激活全新或未初始化的 Olares 设备（例如 Olares One）。

## 学习目标

通过本教程，你将学习：

- 获取并运行适合当前系统版本的 Olares CLI 工具。
- 获取用于远程访问的 Fast Reverse Proxy (FRP) 主机地址。
- 在新设备上安装 Olares。
- 运行激活命令以配置设备。

## 前提条件

开始之前，确保满足以下要求：

- 可以通过键盘和屏幕直接操作 Olares 设备，或者通过 SSH 连接设备。
- 设备已连接互联网，用于下载安装包、查询 FRP 服务器和完成激活。
- 可以以 root 用户身份运行命令，或者在命令前添加 `sudo`。
- 已使用 LarePass 应用注册了 Olares ID，并[备份了 12 个单词的助记词](/zh/manual/larepass/back-up-mnemonics.md)。

    ![快速创建](/images/zh/manual/tutorials/create-olares-id1.png)

## 步骤 1：准备 CLI 工具

根据你的 Olares 版本确定 CLI 准备步骤。

<Tabs>
<template #Olares-v1.12.6-and-later>

v1.12.6 及更高版本的系统内置了具备激活功能的 CLI，无需额外下载。直接跳到步骤 2。
</template>
<template #Olares-v1.12.5-and-earlier>

:::warning 使用独立的 CLI 工具
v1.12.5 及更早版本系统内置的 `olares-cli` 不具备激活功能。你需要下载独立的每日构建版 CLI 来执行激活。使用该工具前，确保阅读以下要求：
- **不要覆盖系统文件**：系统内置的 `olares-cli`、`olaresd` 和集群版本之间存在严格的版本对应关系。切勿将下载的独立 CLI 移动或复制到系统 `/usr/bin/olares-cli` 文件上进行覆盖。这样做会破坏版本链，影响未来的系统升级。
- **执行路径差异**：运行 `./olares-cli` 来执行下载到当前目录的独立版本。不要直接运行 `olares-cli`，因为那样会调用缺少激活功能的系统内置版本。
:::

1. 下载独立的 Olares CLI 安装包。

    ```bash
    curl -sSOL https://cdn.olares.com/common/olares-cli-amd64.8cbdc32.tar.gz
    ```

2. 解压已下载的文件。

    ```bash
    tar xzf olares-cli-amd64.8cbdc32.tar.gz
    ```

3. 为解压出的二进制文件添加可执行权限。

    ```bash
    chmod +x olares-cli
    ```
</template>
</Tabs>

## 步骤 2：获取 FRP 列表

查找可用的 FRP 主机，以启用对设备的远程访问。

1. 根据你的 Olares 版本运行以下命令，将 `{olares-id}` 替换为你已注册的 Olares ID。

    <Tabs>
    <template #Olares-v1.12.6-and-later>

    ```bash
    olares-cli wizard frp {olares-id}
    ```

    **示例**：

     ```bash
    olares-cli wizard frp alice2026@olares.com
    ```
    </template>
    <template #Olares-v1.12.5-and-earlier>

    ```bash
    ./olares-cli wizard frp {olares-id}
    ```

    **示例**：

    ```bash
    ./olares-cli wizard frp alice2026@olares.com
    ```
    </template>
    </Tabs>

2. 从输出列表中选择一个主机地址，并将其记录下来供激活步骤使用。例如，`bb.hongkong.frp.olares.com`。

## 步骤 3：安装 Olares

:::info 关于 Olares One 硬件
全新开箱的 Olares One 设备是处于未安装状态的。在尝试激活之前，必须首先运行安装命令来设置 Olares。
:::

1. 以 root 身份运行安装命令。

    ```bash
    sudo olares-cli install
    ```
2. 当提示输入域名时，输入 `olares.com`。
3. 当提示输入 Olares ID 时，输入你在 LarePass 应用中已注册的 ID。
4. 等待安装过程完成。终端会输出一个本地网关地址和一个默认密码。记录下这些信息，用于激活步骤使用。

    **示例**：
    - **Wizard URL**：本地网关地址，例如 `http://192.168.31.127:30180`。
    - **密码**：Olares 的默认登录密码。

    ![Wizard URL](/images/manual/get-started/wizard-url-and-login-password1.png)

## 步骤 4：激活 Olares

运行激活命令以配置和保护设备。此过程会将 Olares ID 与设备关联，并配置网络隧道和凭证。

1. 根据以下必需参数，准备激活命令：

    | 参数 | 说明 |
    |:----------|:------------|
    | `olares-id` | 在 LarePass 中创建的 Olares ID，例如 `alice2026@olares.com`。 |
    | `mnemonic` | Olares ID 对应的 12 个助记词。 |
    | `password` | 步骤 3 中生成的 Olares 默认登录密码。 |
    | `reset-password` | 用于替换默认密码的新登录密码。 |
    | `authurl` | 步骤 3 中的 Wizard URL。 |
    | `vault` | 步骤 3 中的 Wizard URL，后接 `/server`。 |
    | `bfl` | 步骤 3 中的 Wizard URL。 |
    | `host` | 步骤 2 中选择的 FRP 主机地址。 |
    | `enable-tunnel` | 设置为 `true` 以启用隧道模式。 |

2. 根据你的 Olares 版本，将以下命令中的占位符替换为实际值，然后运行。

    <Tabs>
    <template #Olares-v1.12.6-and-later>

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
    
    如果 Olares ID 是 `alice2026@olares.com`，Wizard URL 是 `http://192.168.31.127:30180`，选择的 FRP 主机是 `bb.hongkong.frp.olares.com`，则运行：

    ```bash
    sudo olares-cli wizard activate alice2026@olares.com \
    --mnemonic "abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef" \
    --password="b8Ln6qbz" \
    --reset-password="Abw1234@" \
    --authurl=http://192.168.31.127:30180 \
    --vault=http://192.168.31.127:30180/server \
    --bfl=http://192.168.31.127:30180 \
    --host=bb.hongkong.frp.olares.com \
    --enable-tunnel=true
    ```
    </template>
    <template #Olares-v1.12.5-and-earlier>

    ```bash
    sudo ./olares-cli wizard activate {olares-id} \
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
    
    如果 Olares ID 是 `alice2026@olares.com`，Wizard URL 是 `http://192.168.31.127:30180`，选择的 FRP 主机是 `bb.hongkong.frp.olares.com`，则运行：

    ```bash
    sudo ./olares-cli wizard activate alice2026@olares.com \
    --mnemonic "abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef abcdef" \
    --password="b8Ln6qbz" \
    --reset-password="Abw1234@" \
    --authurl=http://192.168.31.127:30180 \
    --vault=http://192.168.31.127:30180/server \
    --bfl=http://192.168.31.127:30180 \
    --host=bb.hongkong.frp.olares.com \
    --enable-tunnel=true
    ```
    </template>
    </Tabs>

3. 等待终端显示激活成功完成的提示。

## 后续步骤

现在你可以使用 Olares ID 和在 `reset-password` 中指定的登录密码登录 Olares。

## 了解更多

- [创建 Olares ID](/zh/manual/get-started/create-olares-id.md)
- [Olares CLI 命令行工具](/zh/developer/install/cli/olares-cli.md)
