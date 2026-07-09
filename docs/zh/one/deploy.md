---
search: false
outline: [2, 3]
description: 使用 Studio 将单容器 Docker 应用部署到 Olares。
head:
  - - meta
    - name: keywords
      content: Olares Studio, Docker, Container
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../one/deploy.md)为准。
:::

# 部署应用 <Badge type="tip" text="20 min" />

Studio 是在 Olares 上运行单容器 Docker 应用的最简单方式。你无需编写代码。

在本教程中，你将部署 [Wallos](https://hub.docker.com/r/bellamy/wallos)（一个个人订阅追踪器），并学习如何将标准 Docker 配置转换为 Studio 设置。

:::tip 推荐用于测试
Studio 创建的部署最适合开发、测试或临时使用。
:::

## 学习目标

完成本教程后，你将学会：
- 将标准 `docker run` 命令或 `docker-compose.yaml` 转换为 Olares Studio 设置。
- 配置 CPU 和内存，并添加环境变量。
- 映射存储卷，使数据持久化，或有意设置为临时数据。
- 部署后自定义应用的名称和图标。

## 前提条件

开始之前，请确保你已具备：
- Olares 1.12.2 或更高版本。
- 一个存在且可从 Olares 主机访问的应用容器镜像。
- 应用的 `docker run` 命令或 `docker-compose.yaml` 作为参考。

## 安装 Studio

1. 打开 Market 并搜索 "Studio"。
2. 点击 **Get**，然后点击 **Install**。

## 参考：Docker 配置 

使用以下任一格式作为参考。你将把相同的值复制到 Studio 字段中。

::: code-group
```docker{3-6,8} [docker run 命令]
docker run -d \
  --name wallos \
  -v /path/to/config/wallos/db:/var/www/html/db \
  -v /path/to/config/wallos/logos:/var/www/html/images/uploads/logos \
  -e TZ=America/Toronto \
  -p 8282:80 \
  --restart unless-stopped \
  bellamy/wallos:latest
```

```yaml{5-6,7-10,12-14} [docker compose]
version: '3.0'

services:
  wallos:
    container_name: wallos
    image: bellamy/wallos:latest
    ports:
      - "8282:80/tcp"
    environment:
      TZ: 'America/Toronto'
    # 卷在容器升级之间保存你的数据
    volumes:
      - './db:/var/www/html/db'
      - './logos:/var/www/html/images/uploads/logos'
    restart: unless-stopped
```
:::

## 创建并配置应用

### 创建项目

1. 打开 Studio 并选择 **Create a new application**。
2. 输入应用名称，例如：`wallos`，然后点击 **Confirm**。
3. 选择 **Port your own container to Olares**。
   ![将你自己的容器移植到 Olares](/images/manual/olares/studio-port-your-own-container-to-olares.png#bordered)

### 配置镜像和端口

这些字段定义了应用的核心运行时设置。你可以在上面的 Docker 配置中找到对应的值。

| Studio 字段 | 要输入的值 | 来源：`docker run` | 来源：`docker-compose.yaml`|
| -- | -- | -- | -- |
| Image | `bellamy/wallos:latest` | 命令中的最后一个 token | `image:` 的值|
| Port | `80` | `-p HOST:CONTAINER` 中的容器端口，即 `:` 后的值 | `ports:` 中的容器端口，即 `:` 后、`/` 前的值（如果存在）|

:::info 为什么只填容器端口
端口映射是 HOST:CONTAINER。容器端口是应用监听的内部端口。主机端口是你访问的外部端口。Studio 自动管理外部访问，所以你只需要输入容器端口。
:::
1. 在 **Image** 字段中，粘贴 `bellamy/wallos:latest`。
2. 在 **Port** 字段中，输入 `80`。

### 配置实例规格

实例规格定义了分配给该应用的 CPU 和内存。

在 **Instance Specifications** 部分，输入最低 CPU 和内存需求。例如：
   - **CPU**：2 core
   - **Memory**：1 G
     ![部署 Wallos](/images/manual/olares/studio-deploy-wallos.png#bordered)

### 添加环境变量

环境变量将配置值传递到容器中。

| Studio 字段 | 要输入的值 | 来源：`docker run` | 来源：`docker-compose.yaml`|
| -- | -- | -- | -- |
| key | `TZ` | 查找 `-e KEY=VALUE`，使用 `=` 前的文本 | 在 `environment:` 下，使用 `KEY: VALUE` 左侧的值|
| value | `America/Toronto` | 查找 `-e KEY=VALUE`，使用 `=` 后的文本 | 在 `environment:` 下，使用 `KEY: VALUE` 右侧的值|

1. 向下滚动到 **Environment Variables**，点击 **+ Add**。
2. 在本示例中，输入键值对：
   - **key**：`TZ`
   - **value**：`America/Toronto`
3. 点击 **Submit**。对其他变量重复此过程。
   ![添加环境变量](/images/manual/olares/studio-add-environment-variables.png#bordered)

### 添加存储卷

卷在重启和重新安装后保留数据。

#### 开始之前

在 Studio 中，每个卷需要填写两个字段：

1. **Mount path**：在 Docker 中，卷看起来像 `HOST:CONTAINER`。使用 `:` 后的部分作为挂载路径。

2. **Host path**：在 Studio 中，主机路径有两个输入项：
    -  Prefix：`/app/data`、`/app/cache` 或 `/app/Home`。
        | 主机路径前缀 | 说明 |
        | --- | --- |
        | `/app/data` | 持久化应用数据。数据可跨节点访问，且应用卸载时不会删除。<br>在 Files 中显示为 `/Data/studio`。 |
        | `/app/cache` | 临时应用数据。数据存储在节点本地磁盘，应用卸载时删除。<br>在 Files 中显示为 `/Cache/<device-name>/studio`。 |
        | `/app/Home` | 用户数据目录。主要用于读取外部用户文件。数据不会删除。|
    - 主机路径：输入目标文件夹，以 `/` 开头，例如 `/db` 或 `/logos`。
        :::info 主机路径规则
        Studio 会自动用应用名称作为完整路径的前缀。如果应用名称为 `test`，你设置主机路径为 `/app/data/folder1`，则在 Files 中的实际路径为 `/Data/studio/test/folder1`。
        :::

#### 为 Wallos 配置卷

Wallos 需要两个卷。逐一添加。

**卷 A：数据库**

源映射：
- 在 `docker run` 中：`/path/to/config/wallos/db:/var/www/html/db`
- 在 `docker-compose.yaml` 中：`./db:/var/www/html/db`。

此数据用于高频 I/O，不需要永久保存。将其映射到 `/app/cache`，应用卸载时将被删除。
1. 点击 **Storage Volume** 旁的 **+ Add**。
2. **Host path** 选择 `/app/cache`，然后输入 `/db`。
3. **Mount path** 输入 `/var/www/html/db`。
4. 点击 **Submit**。

**卷 B：Logo**

源映射：
- 在 `docker run` 中：`/path/to/config/wallos/logos:/var/www/html/images/uploads/logos`
- 在 `docker-compose.yaml` 中：`./logos:/var/www/html/images/uploads/logos`。

这是用户上传的数据，即使应用重新安装也应保持持久且可复用。将其映射到 `/app/data`。

1. 点击 **Storage Volume** 旁的 **+ Add**。
2. **Host path** 选择 `/app/data`，然后输入 `/logos`。
3. **Mount path** 输入 `/var/www/html/images/uploads/logos`。
4. 点击 **Submit**。
![添加卷](/images/manual/olares/studio-add-storage-volumes.png#bordered){width=90%}

你可以稍后检查 Files 以验证挂载路径。
![在 Files 中检查挂载路径](/images/manual/olares/studio-check-mounted-path-in-files.png#bordered){width=90%}

### 可选：配置 GPU 或数据库中间件

如果你的应用需要 GPU，在 **Instance Specifications** 下启用 **GPU** 选项并选择 GPU 供应商。
![启用 GPU](/images/manual/olares/studio-enable-GPU.png#bordered){width=90%}

如果你的应用需要 Postgres 或 Redis，在 **Instance Specifications** 下启用。
![启用数据库](/images/manual/olares/studio-enable-databases.png#bordered){width=90%}

启用后，Studio 会提供动态变量。你必须在 **Environment Variables** 部分使用这些变量，以便应用连接到数据库。
- **Postgres 变量：**

| 变量 | 说明 |
|--------------|-----------------------|
| $(PG_USER) | PostgreSQL 用户名 |
| $(PG_DBNAME) | 数据库名称 |
| $(PG_PASS) | Postgres 密码 |
| $(PG_HOST) | Postgres 服务主机 |
| $(PG_PORT) | Postgres 服务端口 |

- **Redis 变量：**

| 变量 | 说明 |
|---------------|--------------------|
| $(REDIS_HOST) | Redis 服务主机 |
| $(REDIS_PORT) | Redis 服务端口 |
| $(REDIS_USER) | Redis 用户名 |
| $(REDIS_PASS) | Redis 密码 |

### 部署并测试应用
1. 点击页面底部的 **Create**。Studio 将自动生成项目文件并自动部署应用。 
2. 等待 Studio 生成应用包文件。你可以在底部状态栏查看状态。
3. 应用成功部署后，点击右上角的 **Preview** 启动它。如果 **Preview** 未出现，请刷新页面。
   ![预览 wallos](/images/manual/olares/studio-preview-wallos.png#bordered)

## 管理应用

### 更新名称和图标

从 Studio 部署的应用包含 `-dev` 后缀和默认图标。你可以通过编辑清单文件来完善它。
![查看已部署应用](/images/manual/olares/studio-app-with-dev-suffix.png#bordered)

1. 在 Studio 中，点击右上角的 **<span class="material-symbols-outlined">box_edit</span>Edit** 打开编辑器。
2. 点击 `OlaresManifest.yaml` 查看内容。
3. 修改 `entrance` 和 `metadata` 下的 `title` 字段。例如，将 `wallos` 改为 `Wallos`。
4. 替换 `entrance` 和 `metadata` 下 `icon` 字段的默认图标图片地址。
   ![编辑 `OlaresManifest.yaml`](/images/manual/olares/studio-edit-olaresmanifest1.png#bordered)
5. 点击右上角的 <span class="material-symbols-outlined">save</span> 保存更改。 
6. 点击 **Apply** 以使用更新后的包重新安装。

   :::info
   如果自上次部署以来未检测到任何更改，点击 **Apply** 将直接返回应用状态页面，不会重新安装。
   :::
   ![更改应用图标](/images/manual/olares/studio-change-app-icon1.png#bordered)

### 移除应用

如果你不再需要该应用，可以将其移除。
1. 在 Studio 中，点击右上角的 <span class="material-symbols-outlined">more_vert</span>。
2. 你可以选择：
   - **Uninstall**：从 Olares 移除运行中的应用，但保留 Studio 中的项目，以便你可以继续编辑包。
   - **Delete**：卸载应用并从 Studio 移除项目。此操作不可逆。

## 故障排除

### 无法安装应用

如果安装失败，请查看页面右下角的错误信息并点击 **View** 展开详情。
![检查应用状态](/images/manual/olares/studio-check-app-status.png#bordered)

### 应用运行但无法正常工作

应用运行后，你可以从 Studio 的部署详情页面管理它。此页面界面与 Control Hub 类似。如果详情未显示，请刷新页面。你可以：

- 使用 **Stop** 和 **Restart** 控件重置进程。此操作通常可以解决运行时问题，如进程冻结。
- 检查事件或日志以调查运行时错误。详情请参阅[导出容器日志以进行故障排除](../../manual/olares/controlhub/manage-container.md#export-container-logs-for-troubleshooting)。

  ![应用部署详情](/images/manual/olares/studio-app-deployment-details.png#bordered)
