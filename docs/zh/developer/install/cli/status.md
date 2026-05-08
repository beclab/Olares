# `status`

## 命令说明

`status` 命令通过调用本机 olaresd 守护进程的 `/system/status` HTTP 接口，输出当前 Olares 系统的状态。

该接口绑定在 `127.0.0.1:18088`，仅接受本地回环流量，因此 `status` 命令必须与 olaresd 运行在同一台机器（通常是主节点）上。

```bash
olares-cli status [选项]
```

默认输出按以下分组展示，便于人工阅读：

- **Olares**：安装生命周期、版本、用户名、关键时间戳。
- **System**：主机的硬件和操作系统信息。
- **Network**：有线/Wi-Fi 连接状态、内/外网 IP 地址。
- **Install / Uninstall**：正在进行的安装或卸载进度。
- **Upgrade**：正在进行的升级进度（包括下载阶段和安装阶段）。
- **Logs collection**：最近一次日志收集任务的状态。
- **Pressures**：节点上当前激活的 kubelet 节点压力条件（若有）。
- **Other**：FRP、容器模式等其他信息。

加上 `--json` 可以输出 olaresd 返回的原始 JSON，便于脚本化处理或与 `jq` 等工具配合使用。

## 选项

| 选项         | 用途                                                                                          | 是否必需 | 默认值                   |
|--------------|----------------------------------------------------------------------------------------------|----------|--------------------------|
| `--endpoint` | 本机 olaresd 守护进程的基础 URL。仅当 olaresd 监听在非默认地址时才需要修改。                 | 否       | `http://127.0.0.1:18088` |
| `--json`     | 直接输出 olaresd 返回的原始 JSON（即响应中的 `data` 字段），适合配合 `jq` 等工具使用。       | 否       | `false`                  |
| `--timeout`  | 等待 olaresd 响应的最长时间。                                                                | 否       | `5s`                     |
| `--help`     | 显示命令帮助。                                                                               | 否       | 无                       |

## 使用示例

```bash
# 默认输出：分组的人工可读报表
olares-cli status

# 原始 JSON 输出，原样转发自 olaresd
olares-cli status --json | jq

# 指定守护进程地址并延长超时时间
olares-cli status --endpoint http://127.0.0.1:18088 --timeout 10s
```

## 字段参考

下表列出 olaresd 返回的字段（即 JSON 响应中 `data` 对象的字段），以及它们在分组输出中显示的标签。

### Olares

| 字段             | JSON Key           | 含义                                                                                       |
|------------------|--------------------|--------------------------------------------------------------------------------------------|
| State            | `terminusState`    | 系统的高层状态，详见 [状态值列表](#状态值列表)。                                            |
| Olaresd state    | `terminusdState`   | olaresd 守护进程自身的生命周期：启动初始化时为 `initialize`，初始化完成后为 `running`。 |
| Name             | `terminusName`     | 管理员的 Olares ID，例如 `alice@olares.cn`。                                              |
| Version          | `terminusVersion`  | 已安装的 Olares 版本（语义化版本号）。                                                    |
| Olaresd version  | `olaresdVersion`   | 当前运行的 olaresd 二进制版本。可用于排查升级后的版本漂移。                              |
| Installed at     | `installedTime`    | Olares 安装完成时间（Unix 时间戳，单位秒）。                                              |
| Initialized at   | `initializedTime`  | 管理员完成初始激活的时间（Unix 时间戳，单位秒）。                                         |

### System

| 字段       | JSON Key      | 含义                                       |
|------------|---------------|--------------------------------------------|
| Device     | `device_name` | 用户友好的设备/机型名称。                  |
| Hostname   | `host_name`   | 内核报告的主机名。                          |
| OS         | `os_type` / `os_arch` / `os_info` | 操作系统类型、CPU 架构、发行版描述。 |
| OS version | `os_version`  | 操作系统版本号，例如 `22.04`。              |
| CPU        | `cpu_info`    | CPU 型号。                                  |
| Memory     | `memory`      | 物理内存总量，格式为 `<N> G`。             |
| Disk       | `disk`        | 数据分区的文件系统总容量，格式为 `<N> G`。 |
| GPU        | `gpu_info`    | 检测到的 GPU 型号（若有）。                |

### Network

| 字段        | JSON Key         | 含义                                                                |
|-------------|------------------|---------------------------------------------------------------------|
| Wired       | `wiredConnected` | 检测到有线连接时为 `yes`。                                          |
| Wi-Fi       | `wifiConnected`  | 默认路由走 Wi-Fi 时为 `yes`。                                       |
| Wi-Fi SSID  | `wifiSSID`       | 已连接 Wi-Fi 的 SSID。                                              |
| Host IP     | `hostIp`         | Olares 用于互联的内网 IPv4 地址。                                   |
| External IP | `externalIp`     | 通过外部探测获取的公网 IPv4 地址（每分钟最多刷新一次）。           |

### Install / Uninstall

| 字段          | JSON Key                | 含义                                                              |
|---------------|-------------------------|-------------------------------------------------------------------|
| Installing    | `installingState`       | 进行中的安装任务的生命周期：`in-progress`、`completed`、`failed`。 |
|               | `installingProgress`    | 当前安装步骤的描述（在分组输出中以括号形式跟随显示）。           |
| Uninstalling  | `uninstallingState`     | 进行中的卸载任务的生命周期。                                      |
|               | `uninstallingProgress`  | 当前卸载步骤的描述。                                              |

### Upgrade

| 字段           | JSON Key                       | 含义                                                                       |
|----------------|--------------------------------|----------------------------------------------------------------------------|
| Target         | `upgradingTarget`              | 进行中升级的目标版本。                                                     |
| State          | `upgradingState`               | 升级安装阶段的生命周期。                                                   |
|                | `upgradingProgress`            | 升级安装阶段的进度描述（括号显示）。                                       |
| Step           | `upgradingStep`                | 当前升级步骤的名称。                                                       |
| Last error     | `upgradingError`               | 升级安装阶段最近一次报错信息。                                             |
| Download state | `upgradingDownloadState`       | 升级下载阶段的生命周期。                                                   |
|                | `upgradingDownloadProgress`    | 升级下载阶段的进度描述（括号显示）。                                       |
| Download step  | `upgradingDownloadStep`        | 当前下载步骤的名称。                                                       |
| Download error | `upgradingDownloadError`       | 升级下载阶段最近一次报错信息。                                             |
| Retry count    | `upgradingRetryNum`            | 升级被自动重试的次数（仅当大于 0 时显示）。                                |
| Next retry at  | `upgradingNextRetryAt`         | 下一次重试的预定时间（仅当存在时显示）。                                   |

### Logs collection

| 字段       | JSON Key                 | 含义                                                                  |
|------------|--------------------------|-----------------------------------------------------------------------|
| State      | `collectingLogsState`    | 通过 olaresd 触发的最近一次日志收集任务的生命周期。                  |
|            | `collectingLogsError`    | 最近一次日志收集任务的错误信息（括号显示）。                          |

### Pressures

`pressures` 数组列出当前节点上为真的 kubelet 节点压力条件。当节点健康时，分组输出中会显示 `(none)`。

| 字段    | JSON Key  | 含义                                                                |
|---------|-----------|---------------------------------------------------------------------|
| Type    | `type`    | Kubernetes 节点条件类型，例如 `MemoryPressure`、`DiskPressure`。 |
| Message | `message` | kubelet 给出的可读说明。                                            |

### Other

| 字段           | JSON Key            | 含义                                                                |
|----------------|---------------------|---------------------------------------------------------------------|
| FRP enabled    | `frpEnable`         | FRP 反向通道是否启用（来自环境变量 `FRP_ENABLE`）。               |
| FRP server     | `defaultFrpServer`  | FRP 服务器地址（来自环境变量 `FRP_SERVER`）。                      |
| Container mode | `containerMode`     | olaresd 运行在容器内时设置（来自环境变量 `CONTAINER_MODE`）。     |

## 状态值列表

`terminusState` 字段可能取以下值。CLI 也使用同一份枚举生成描述，因此下表始终与 CLI 输出保持一致。

| 取值                  | 含义                                                            |
|-----------------------|-----------------------------------------------------------------|
| `checking`            | olaresd 还未完成首次状态探测。                                  |
| `network-not-ready`   | 未检测到可用的内网 IPv4 地址。                                  |
| `not-installed`       | 当前节点未安装 Olares。                                         |
| `installing`          | Olares 正在安装中。                                             |
| `install-failed`      | 最近一次安装失败。                                              |
| `uninitialized`       | Olares 已安装，但管理员账户尚未激活。                          |
| `initializing`        | 管理员账户正在激活中。                                          |
| `initialize-failed`   | 管理员账户激活失败。                                            |
| `terminus-running`    | Olares 运行正常。                                               |
| `restarting`          | 节点刚刚重启，状态会在短时间内稳定。                            |
| `invalid-ip-address`  | 节点 IP 已变更，需要执行 `change-ip` 恢复。                    |
| `ip-changing`         | `change-ip` 操作正在进行。                                      |
| `ip-change-failed`    | 最近一次 `change-ip` 操作失败。                                 |
| `system-error`        | 关键 Pod 未正常运行。                                           |
| `self-repairing`      | olaresd 正在尝试自动修复。                                      |
| `adding-node`         | 正在加入 worker 节点。                                          |
| `removing-node`       | 正在移除 worker 节点。                                          |
| `uninstalling`        | Olares 正在卸载中。                                             |
| `upgrading`           | 升级正在执行中。                                                |
| `disk-modifing`       | 存储布局正在调整中。                                            |
| `shutdown`            | 系统正在关机。                                                  |
