---
noindex: true
outline: [2, 3]
description: 在 Olares 主机上运行的所有 olares-cli 命令的字母顺序参考。包含语法、结构，以及指向各命令详情页（含选项和示例）的链接。
---
# 命令参考

本页是在 Olares 主机上运行的所有 `olares-cli` 命令的字母顺序参考。你可以在这里查找某条命令的语法、选项和示例。对于以登录用户身份、或在 Olares 应用内部运行的命令，参见[搭配 AI Agent 使用](../../cli-overview.md#了解更多)。

:::warning 适用于 Olares 1.12.X
如果你的 Olares 版本是 1.12.X，请使用此版本的参考。
:::

:::info 需要 root 权限
大多数 `olares-cli` 命令都需要 root 权限。请使用 root 用户执行命令，或在命令前加上 `sudo`。
:::

:::info 在 WSL 中使用 olares-cli
如果通过 WSL（Windows Subsystem for Linux）方式安装了 Olares，请在 WSL 环境中运行 `olares-cli`。在 PowerShell 中：

```powershell
wsl -d Ubuntu
```
:::

## 语法
Olares 命令行工具使用如下语法：

> `olares-cli 命令 [子命令] [参数] [选项]`

其中：
- `命令`：指定要执行的主要操作，例如 `olares-cli install`。
- `子命令`：进一步指定命令的具体任务，适用于支持子操作的命令。例如 `wizard` 或 `component`。
- `参数`：指定命令的目标资源或输入数据，通常是 ID、名称或文件路径。例如，在 `olares-cli user activate <Olares ID> [选项]` 中，`<Olares ID>` 就是该命令的参数。
- `选项`：可选参数，用于修改命令的行为。包括标志（flags）和带参数的选项。

通过 Olares 命令行工具，你可以临时覆盖某些 Olares 默认设置。每个选项仅对当前执行的命令生效。

例如，在执行 `olares-cli download wizard` 时使用 `--base-dir` 选项，只会影响向导的下载过程，而不会改变其他命令（如“安装”阶段）的基础目录。

如需查看任何命令的详细帮助信息，请运行 `olares-cli help`。

## 可用命令列表

| 操作 | 语法   | 说明   |
|--|--|--|
| `backups` | `olares-cli backups <子命令> [选项]`  | 管理备份相关操作。 |
| `change-ip` | `olares-cli change-ip [选项]` | 修改 Olares OS 的 IP 地址。 |
| `disk` | `olares-cli disk <子命令>` | 管理 Olares 系统存储资源。 |
| `download` | `olares-cli download <子命令> [选项]` | 下载指定资源。 |
| `gpu` | `olares-cli gpu <子命令> [选项]` | 管理 GPU 相关的操作。 |
| `info` | `olares-cli info [选项]`  | 显示已下载的 Olares OS 的常规信息。|
| `install` | `olares-cli install [选项]` | 部署 Olares 的系统级和用户级组件。|
| `logs` | `olares-cli logs [选项]` | 收集 Olares 系统组件的日志，用于调试和故障排查。 |
| `node`  | `olares-cli node <子命令> [选项]` | 管理节点相关的操作。 |
| `osinfo` | `olares-cli osinfo <子命令> [选项]` | 显示当前设备的操作系统信息。 |
| `precheck`| `olares-cli precheck [选项]` | 检查系统环境是否满足 Olares 安装要求。|
| `prepare` | `olares-cli prepare [选项]` | 为安装过程准备环境，包括设置 Olares 的基础服务和配置。 |
| `release` | `olares-cli release [选项]` | 打包 Olares 安装资源以供分发或部署。|
| `start` | `olares-cli start [选项]` | 启动 Olares 服务和组件。 |
| `status` | `olares-cli status [选项]` | 查询本机 olaresd 守护进程，输出当前 Olares 系统状态。 |
| `stop` | `olares-cli stop [选项]` | 停止 Olares 服务和组件。 |
| `uninstall` | `olares-cli uninstall [选项]` | 完全卸载 Olares，或将安装回滚到特定阶段。 |
| `upgrade` | `olares-cli upgrade <子命令> [选项]` | 升级 Olares，检查升级准备情况与兼容性。 |
| `user` | `olares-cli user <子命令> [选项]`| 管理 Olares 用户。 |