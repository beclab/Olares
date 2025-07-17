---
outline: [2, 3]
description: Olares 部署阶段的技术细节，包括预检、下载、准备和安装四个阶段。详细说明每个安装阶段的具体实现。
---
# Olares 安装流程详解
本文档从四个主要阶段详细说明 Olares 的安装流程，包括各阶段的底层命令、配置和逻辑。文档适用于希望深入了解安装过程的开发者和系统管理员。

## 四个安装阶段
Olares 的安装可分为以下四个阶段：

- **预检（Precheck）**：验证系统环境是否满足 Olares 安装的所有前置条件。
- **下载（Download）**：获取安装所需的所有文件、依赖项和容器镜像。
- **准备（Prepare）**：配置操作系统和系统服务，为 Kubernetes 和 Olares 组件创建运行环境。
- **安装（Install）**：部署 Kubernetes，集成 KubeSphere，并安装 Olares 的核心服务和应用程序。

## 预检阶段

预检阶段的重点是验证系统是否满足安装 Olares 的必要条件。通过运行 `olares-cli precheck` 命令执行一系列验证检查。若在此阶段发现任何问题，需在继续安装前解决。 

关键预检项目包括：
- 检查操作系统类型、版本号以及处理器架构是否兼容
- 确保系统使用 `Systemd` 作为初始化进程
- 验证 Olares 需要暴露的多个网络端口是否可用
- 系统里是否存在与 Olares 冲突的容器运行时

下图是一个预检失败的示例：

 ![Precheck](/images/developer/install/precheck.png)

在此示例中，有两项检查失败：
- Olares 所需的端口 `9100` 已被占用。
- 系统中检测到已有容器运行时。

继续安装前必须解决这些问题。

## 下载阶段

下载阶段会下载 Olares 安装所需的 Wizard 文件、系统依赖组件和容器镜像。

### 下载 Wizard 文件

Wizard 文件是一个元数据包，包含所有 Olares 组件的下载链接和配置信息。Wizard 文件是此阶段首个被获取的文件，为后续的下载提供了关键信息。

Wizard 文件会被默认解压存储至 `$HOME/.olares/versions/<version>` 目录。

其中：
- `$HOME/.olares` 是 Olares 的基础安装目录，
- `<version>` 是 Olares 版本号，如示例中的 `1.12.0-20241215`。

:::details 脚本输出示例
```bash
➜  ~ ./install.sh
the KUBE_TYPE env var is not set, defaulting to "k3s"
olares-cli already installed and is the expected version

downloading installation wizard...

current: root
2024-12-17T18:01:19.501+0800        [Job] [Download Installation Wizard] start ...
2024-12-17T18:01:19.501+0800        [Module] GreetingsModule
Greetings, Olares
2024-12-17T18:01:19.502+0800        [A] ubuntu: Greetings success (611.77µs)
2024-12-17T18:01:19.502+0800        [Module] DownloadInstallWizard
/home/keven/.olares/versions/v1.12.0-20241215/.env
/home/keven/.olares/versions/v1.12.0-20241215/wizard/config/account/Chart.yaml
```
:::
### 下载安装所需组件与容器镜像
Wizard 下载完成后，脚本会下载 Olares 所需的所有依赖组件和容器镜像。保存路径如下：
- 依赖包：`$HOME/.olares/pkg` 
- 容器镜像：`$HOME/.olares/image`

这种存储结构支持在多个版本之间复用稳定组件，避免重复下载。

:::details **脚本输出示例**
```bash
downloading installation packages...

current: root
2024-12-17T19:41:36.847+0800        [Job] [Download Installation Package] start ...
2024-12-17T19:41:36.847+0800        [Module] GreetingsModule
Greetings, Olares
2024-12-17T19:41:36.848+0800        [A] ubuntu: Greetings success (512.711µs)
2024-12-17T19:41:36.848+0800        [Module] GenerateOlaresUninstallScript
2024-12-17T19:41:36.879+0800        [A] LocalHost: GenerateOlaresUninstallScript success (31.279866ms)
2024-12-17T19:41:36.879+0800        [Module] PackageDownloadModule
2024-12-17T19:41:36.879+0800        checking local cache ...
2024-12-17T19:41:44.614+0800        5 out of 177 files need to be downloaded
2024-12-17T19:41:44.615+0800        (1/5) downloading package olaresd, file: olaresd-v0.0.50.tar.gz
2024-12-17T19:41:51.814+0800        (2/5) downloading image calico/kube-controllers:v3.23.2, file: 521564c4b60ae73c78899b7b40ae655e.tar.gz
...
```
:::
## 准备阶段

准备阶段配置操作系统环境，以支持 Kubernetes、容器镜像和 Olares 系统服务。

此阶段包括以下主要任务：
- 配置系统
- 配置容器运行时
- 安装系统守护进程

### 配置系统

安装脚本会配置 Linux 环境以满足 Olares 的要求，这些配置包括：

- 调整 DNS、NTP 和 SSH 服务，确保网络功能和时间同步正常。
- 通过 `apt` 安装基本依赖（如 curl、net-tools、gcc、make）。

:::details 脚本输出示例
```bash
preparing installation environment...

current: root
2024-12-17T19:46:39.517+0800        [Job] [Prepare the System Environment] start ...
2024-12-17T19:46:39.517+0800        [Module] PreCheckOs
2024-12-17T19:46:39.517+0800        [A] LocalHost: PreCheckSupport success (29.999µs)
2024-12-17T19:46:39.517+0800        [A] LocalHost: PreCheckPortsBindable success (144.035µs)
2024-12-17T19:46:39.517+0800        [A] LocalHost: PreCheckNoConflictingContainerd success (31.009µs)
2024-12-17T19:46:39.517+0800        [A] ubuntu: PatchAppArmor skipped (7.677µs)
2024-12-17T19:46:39.517+0800        [A] ubuntu: RaspbianCheck success (5.796µs)
2024-12-17T19:46:39.517+0800        [A] ubuntu: CorrectHostname success (5.363µs)
nameserver
nameserver
2024-12-17T19:46:41.921+0800        [A] ubuntu: DisableLocalDNS success (2.40336625s)
2024-12-17T19:46:41.921+0800        [INFO] installing and configuring OS dependencies ...
2024-12-17T19:46:41.921+0800        [Module] InstallDeps
Hit:1 http://security.ubuntu.com/ubuntu jammy-security InRelease
Hit:2 https://download.docker.com/linux/ubuntu jammy InRelease
Hit:3 http://hk.archive.ubuntu.com/ubuntu jammy InRelease
...
```
:::
### 配置容器运行时
容器运行时是运行容器化应用程序的关键组件。在这一步中，安装脚本将：
- 安装并启动之前下载的依赖组件
- 在系统上安装 containerd 并启动服务
- 将下载好的容器镜像导入至 containerd

:::details 脚本输出示例
```bash
2024-12-17T19:47:37.510+0800        [Module] InstallContainerModule(k3s)
2024-12-17T19:47:37.518+0800        [A] ubuntu: ZfsMountReset skipped (7.321811ms)
2024-12-17T19:47:37.525+0800        [A] ubuntu: CreateZfsMount skipped (7.322591ms)
2024-12-17T19:47:38.188+0800        [A] ubuntu: SyncContainerd success (662.643982ms)
2024-12-17T19:47:38.368+0800        [A] ubuntu: SyncCrictlBinaries success (179.758334ms)
2024-12-17T19:47:38.399+0800        [A] ubuntu: GenerateContainerdService success (31.410118ms)
2024-12-17T19:47:38.451+0800        [A] ubuntu: GenerateContainerdConfig success (52.047108ms)
2024-12-17T19:47:38.505+0800        [A] ubuntu: GenerateCrictlConfig success (53.760209ms)
2024-12-17T19:47:38.857+0800        [A] ubuntu: EnableContainerd success (352.128078ms)
2024-12-17T19:47:38.857+0800        [Module] PreloadImages
2024-12-17T19:47:41.665+0800        (1/145) imported image: rancher/mirrored-pause:3.6, time: 194.363948ms
...
```
:::
### 安装系统守护进程
接着是安装并启动 Olares 系统守护进程 olaresd，用于监控系统并自动执行维护任务。

:::details 脚本输出示例
```bash
024-12-17T19:52:31.862+0800        [A] ubuntu: GenerateOlaresdEnv success (23.829684ms)
2024-12-17T19:52:31.862+0800        template OlaresdService result: [Unit]
Description=olaresd
After=network.target
StartLimitIntervalSec=0

[Service]
User=root
EnvironmentFile=/etc/systemd/system/olaresd.service.env
ExecStart=/usr/local/bin/olaresd
RestartSec=10s
LimitNOFILE=40000
Restart=always

[Install]
WantedBy=multi-user.target

2024-12-17T19:52:31.885+0800        [A] ubuntu: GenerateOlaresdService success (23.050958ms)
2024-12-17T19:52:32.033+0800        [A] ubuntu: EnableOlaresdService success (147.987242ms)
...
```
:::
## 安装阶段
安装阶段将所有组件整合起来，并完成最终环境配置。

在此阶段，脚本会执行以下关键任务：
- 部署 Kubernetes。
- 集成 KubeSphere，实现云原生管理与可观测性。
- 配置 Olares 账户。
- 部署并启动内置应用和服务。

### 部署 Kubernetes

Kubernetes 是 Olares 系统的核心调度组件。在此步骤中，安装脚本会执行以下操作：
1. 启动 etcd 数据库。  
2. 启动并配置 K3s。  
3. 安装用于集群网络通信的容器网络界面（CNI） 插件。  
4. 将 `kubeconfig` 文件复制到当前用户目录，以便通过 `kubectl` 与集群进行交互。

由于 K3s 轻量且易用，Olares 默认使用它作为 Kubernetes 的发行版。如果你有高级或自定义的配置需求，也可以选择安装完整的 Kubernetes。

在 macOS 环境下，脚本会使用 minikube 部署集群，并跳过上述步骤。

:::details 脚本输出示例
```bash
[certs] Generating "ca" certificate and key
[certs] admin-ubuntu serving cert is signed for DNS names [etcd etcd.kube-system etcd.kube-system.svc etcd.kube-system.svc.cluster.local lb.kubesphere.local localhost ubuntu] and IPs [127.0.0.1 ::1 192.168.1.16]
[certs] member-ubuntu serving cert is signed for DNS names [etcd etcd.kube-system etcd.kube-system.svc etcd.kube-system.svc.cluster.local lb.kubesphere.local localhost ubuntu] and IPs [127.0.0.1 ::1 192.168.1.16]
[certs] node-ubuntu serving cert is signed for DNS names [etcd etcd.kube-system etcd.kube-system.svc etcd.kube-system.svc.cluster.local lb.kubesphere.local localhost ubuntu] and IPs [127.0.0.1 ::1 192.168.1.16]
2024-12-17T19:52:36.957+0800        [A] LocalHost: GenerateETCDCerts success (263.237575ms)
2024-12-17T19:52:37.263+0800        [A] ubuntu: SyncCertsFile success (306.213676ms)
2024-12-17T19:52:37.264+0800        [A] ubuntu: SyncCertsFileToMaster skipped (20.351µs)
2024-12-17T19:52:37.264+0800        [Module] InstallETCDBinaryModule
2024-12-17T19:52:37.698+0800        [A] ubuntu: InstallETCDBinary success (434.014395ms)
2024-12-17T19:52:37.728+0800        [A] ubuntu: GenerateETCDService success (30.732882ms)
2024-12-17T19:52:37.728+0800        [A] ubuntu: GenerateAccessAddress success (23.491µs)
2024-12-17T19:52:37.728+0800        [Module] ETCDConfigureModule
2024-12-17T19:52:37.728+0800        [A] ubuntu: ExistETCDHealthCheck skipped (9.903µs)
2024-12-17T19:52:37.753+0800        [A] ubuntu: GenerateETCDConfig success (24.125665ms)
2024-12-17T19:52:37.773+0800        [A] ubuntu: AllRefreshETCDConfig success (20.321235ms)
2024-12-17T19:52:40.048+0800        [A] ubuntu: RestartETCD success (2.274541565s)
2024-12-17T19:52:40.068+0800        [A] ubuntu: AllETCDNodeHealthCheck success (20.251062ms)
2024-12-17T19:52:40.094+0800        [A] ubuntu: RefreshETCDConfigToExist success (26.207599ms)
2024-12-17T19:52:40.129+0800        [A] ubuntu: AllETCDNodeHealthCheck success (34.462881ms)
2024-12-17T19:52:40.129+0800        [Module] ETCDBackupModule
2024-12-17T19:52:40.185+0800        [A] ubuntu: BackupETCD success (56.639923ms)
2024-12-17T19:52:40.230+0800        [A] ubuntu: GenerateBackupETCDService success (44.727929ms)
2024-12-17T19:52:40.273+0800        [A] ubuntu: GenerateBackupETCDTimer success (42.839457ms)
2024-12-17T19:52:40.396+0800        [A] ubuntu: EnableBackupETCDService success (122.621074ms)
2024-12-17T19:52:40.396+0800        [Module] InstallKubeBinariesModule
2024-12-17T19:52:41.188+0800        [A] ubuntu: SyncKubeBinary(k3s) success (791.866964ms)
2024-12-17T19:52:41.218+0800        [A] ubuntu: GenerateK3sKillAllScript success (30.442837ms)
2024-12-17T19:52:41.253+0800        [A] ubuntu: GenerateK3sUninstallScript success (34.802683ms)
2024-12-17T19:52:41.268+0800        [A] ubuntu: ChmodScript(k3s) success (14.640733ms)
2024-12-17T19:52:41.268+0800        [Module] K3sInitClusterModule
2024-12-17T19:52:41.334+0800        [A] ubuntu: GenerateK3sService success (66.556896ms)
2024-12-17T19:52:41.379+0800        [A] ubuntu: GenerateK3sServiceEnv success (44.492752ms)
2024-12-17T19:52:41.414+0800        [A] ubuntu: GenerateK3sRegistryConfig success (34.814475ms)
2024-12-17T19:52:46.511+0800        [A] ubuntu: EnableK3sService success (5.097800474s)
2024-12-17T19:52:46.572+0800        [A] ubuntu: CopyKubeConfig success (60.33887ms)
...
```
:::
### 集成 KubeSphere

Olares 会基于 Kubernetes 安装 KubeSphere，以增强系统的管理和可观测性。主要特性包括：

- 系统监控与告警
- 资源和工作区管理
- 命名空间和自定义资源定义（CRD）管理

### 配置账户
Olares ID 在 LarePass 应用中创建。在这一步中，安装脚本会提示输入以下信息，以便后续使用 LarePass 激活 Olares：

- **Olares 域名**: Olares 提供默认域名 `olares.com` 和 `olares.cn` 。如果你已将自定义域名添加到 Olares Space，也可以在此步骤中输入自定义域名。 
- **Olares ID**: 输入 Olares ID 中的用户名部分。

完成此步骤后，系统会为你创建一个用于登录 Olares 的账户，并完成相关的访问和权限配置。  

:::details 脚本输出示例
```bash
Enter the domain name ( olares.cn by default ):
2024-12-17T20:58:15.690+0800        using Domain Name: olares.cn

Enter the Olares ID (which you registered in the LarePass app): marvin113
2024-12-17T20:58:52.584+0800        using Olares Local Name: marvin113
2024-12-17T20:58:52.584+0800        using Olares ID: marvin113@olares.com
2024-12-17T20:58:52.584+0800        using password: 2uO5PZ2X
```
:::
### 安装系统应用
安装的最后一步会通过 Helm 部署 Olares 系统的核心服务与用户应用：
- **核心系统服务**（在 `os-system` 命名空间中）：包含备份（Velero）、存储（OpenEBS）、Redis、Nats、MinIO 等关键组件。
- **用户应用**（在`user-space-xxx` 命名空间中）：包含文件管理器、桌面、设置等系统应用。

安装过程中，日志中会显示如 `[helm] app installed success` 以及一系列 `xxx created` 的提示，表示对应的 Helm Chart 或 Kubernetes 资源安装成功。

:::details 脚本输出示例
```bash
2024-12-17T19:53:18.382+0800        [A] ubuntu: InitKsNamespace success (2.678362348s)
2024-12-17T19:53:18.382+0800        [Module] DeploySnapshotController
customresourcedefinition.apiextensions.k8s.io/volumesnapshotclasses.snapshot.storage.k8s.io created
customresourcedefinition.apiextensions.k8s.io/volumesnapshotcontents.snapshot.storage.k8s.io created
customresourcedefinition.apiextensions.k8s.io/volumesnapshots.snapshot.storage.k8s.io created
2024-12-17T19:53:18.924+0800        [helm] app installed success        {"NAME": "snapshot-controller", "LAST DEPLOYED": "Tue Dec 17 19:53:18 2024", "NAMESPACE": "kube-system", "STATUS": "deployed", "REVISION": 1}
2024-12-17T19:53:18.924+0800        [A] ubuntu: CreateSnapshotController success (541.656132ms)
2024-12-17T19:53:18.924+0800        [Module] DeployRedis
secret/redis-secret created
2024-12-17T19:53:19.057+0800        [A] ubuntu: CreateRedisSecret success (133.123121ms)
2024-12-17T19:53:19.189+0800        [A] ubuntu: BackupRedisManifests success (132.045425ms)
2024-12-17T19:53:19.339+0800        [A] ubuntu: DeployRedisHA success (149.251633ms)
local (default)   openebs.io/local   Delete          WaitForFirstConsumer   false                  31s
local (default)   openebs.io/local   Delete   WaitForFirstConsumer   false   31s
2024-12-17T19:53:19.971+0800        [helm] app installed success        {"NAME": "redis", "LAST DEPLOYED": "Tue Dec 17 19:53:19 2024", "NAMESPACE": "kubesphere-system", "STATUS": "deployed", "REVISION": 1}
...
```
:::

### 完成安装

待所有组件部署成功后，脚本会输出包含激活向导页面 URL 的汇总信息：

```bash
2024-12-17T21:00:58.086+0800        [INFO] Installation wizard is complete
2024-12-17T21:00:58.086+0800        [INFO] All done

------------------------------------------------

2024-12-17T21:00:58.086+0800        Olares is running at:
2024-12-17T21:00:58.086+0800        http://192.168.1.16:30180

2024-12-17T21:00:58.086+0800        Open your browser and visit the above address
2024-12-17T21:00:58.086+0800        with the following credentials:

2024-12-17T21:00:58.086+0800        Username: marvin113
2024-12-17T21:00:58.086+0800        Password: 2uO5PZ2X
```

此时，执行以下操作以完成安装：
1. 打开浏览器并输入提供的 URL。
2. 使用初始密码登录激活向导。
3. 按照屏幕提示完成激活流程。

完成激活后，你就可以开始使用 Olares 了。

## 了解更多

- [`olares-cli` 命令行参考](../install/cli/olares-cli.md)
- [Olares 安装概述](installation-overview.md)
- [Olares 环境变量](environment-variables.md)