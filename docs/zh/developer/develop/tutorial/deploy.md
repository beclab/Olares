---
outline: [2, 3]
description: 如何使用 Studio 将单容器 Docker 应用快速部署到 Olares。
---
# 基于 Docker 镜像部署应用
本文档介绍如何使用 Studio 将单容器 Docker 应用部署到 Olares 系统。

:::info 仅限单容器应用
此方法仅适用于通过单个容器镜像运行的应用。
:::
:::tip 推荐用于测试场景
通过 Studio 部署的应用主要面向开发与测试场景。相比市场安装的正式应用，它在版本维护和数据持久化方面存在局限。如需长期稳定使用，建议先[打包并上传应用](package-upload.md)，然后通过应用市场安装。
:::

## 前提条件
- Olares 1.12.2 及以上版本。
- 应用的容器镜像已存在，且 Olares 主机可以访问。
- 具备应用的 `docker run` 命令或 `docker-compose.yaml` 文件，用于参考端口、环境变量和挂载卷等配置信息。

## 创建并配置应用
本节以个人订阅和开支追踪应用 [Wallos](https://hub.docker.com/r/bellamy/wallos) 为例，演示如何将常见的 Docker 配置（镜像、端口、环境变量、卷）映射到 Studio 中。

**Docker 配置参考示例**
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

```yaml{5-6,7-10,12-14} [docker compose 文件]
version: '3.0'

services:
  wallos:
    container_name: wallos
    image: bellamy/wallos:latest
    ports:
      - "8282:80/tcp"
    environment:
      TZ: 'America/Toronto'
    volumes:
      - './db:/var/www/html/db'
      - './logos:/var/www/html/images/uploads/logos'
    restart: unless-stopped
```
:::
### 创建应用

1. 打开 Studio，选择**创建新应用**。
2. 输入**应用名称**，例如 `wallos`，然后点击**确认**。
3. 选择**将自己的容器部署到 Olares 上**。
   ![将自己的容器部署到 Olares 上](/images/manual/olares/studio-port-your-own-container-to-olares.png#bordered)

### 配置镜像、端口和实例规格
这些字段定义了应用的核心组件。参考 `docker run` 命令中的镜像名和 `-p` 参数，或 `docker-compose.yaml` 文件中的 `image:` 和 `ports:` 字段进行填写。
1.  在**容器镜像**字段中，粘贴镜像名称，例如 `bellamy/wallos:latest`。
2.  在**容器端口**字段中，参考 `主机端口:容器端口` 格式的映射（如 `8282:80`），填写冒号后的 `80`。
   :::tip 仅需填写容器端口
    端口映射的标准格式为 `主机端口:容器端口`。冒号后的是应用在内部监听的“容器端口”，冒号前的是供外部访问的“主机端口”。由于 Studio 会自动管理外部路由，你只需填写容器端口即可。
   :::
3.  在**实例规格**区域，设置应用所需的最低 CPU 和内存要求。例如：
   * **CPU**: 2 core
   * **Memory**: 1 Gi
     ![部署 Wallos](/images/manual/olares/studio-deploy-wallos.png#bordered)

### 添加环境变量
环境变量主要用于向应用传递配置信息，对应 Docker 示例中的 `-e` 参数或 `environment` 字段。
1. 向下滚动至**环境变量**区域，点击**添加**。
2. 参照下图示例，填写时区配置：
   - **键**： `TZ`
   - **值**：`America/Toronto`
3. 点击**提交**。如需添加更多变量，重复此过程。
   ![添加环境变量](/images/manual/olares/studio-add-environment-variables.png#bordered)

### 添加存储卷
存储卷用于将 Olares 设备的物理存储映射到容器内部，这是确保数据持久化的关键步骤，对应 Docker 示例中的 `-v` 参数或 `volumes` 字段。

:::info 理解主机路径
主机路径是指数据在 Olares 系统中的实际存储位置，Studio 提供了三种预设的前缀路径：

- `/app/data`：应用数据目录。数据可跨节点访问，且卸载应用时**不会**删除。在文件管理器中显示为 `/Data/studio`。
- `/app/cache`：应用缓存目录。数据存在节点本地磁盘，卸载应用时会自动删除。在文件管理器中显示为 `/Cache/<device-name>/studio`。
- `/app/Home`：用户数据目录。主要用于读取外部文件，数据不会被删除。
:::
:::info 主机路径规则
- 输入的主机路径必须以 `/` 开头。
- Studio 会自动补全路径前缀。例如，应用名为 `test`，当设置主机路径为 `/app/data/folder1` 时，在文件管理器中的实际路径为 `/Data/studio/test/folder1`。
:::

本应用需要依次挂载两个存储卷：
1. 添加数据库卷。此类数据涉及高频 I/O 读写且无需永久保存。将其映射至 `/app/cache` 以便在应用卸载时自动清理。

   a. 点击**存储卷**旁的**添加**。

   b. **主机路径**选择 `/app/cache`，并输入 `/db`。

   c. **容器路径**输入 `/var/www/html/db`。

   d. 点击**提交**。
2. 添加 Logo 卷。此类数据为用户上传内容，需持久化保存，即使重装应用也不应丢失。将其映射至 `/app/data`。

   a. 点击**存储卷**旁的**添加**。
   
   b. **主机路径**选择 `/app/data`，并输入 `/logos`。
   
   c. **容器路径**输入 `/var/www/html/images/uploads/logos`。
   
   d. 点击**提交**。
![添加存储卷](/images/manual/olares/studio-add-storage-volumes.png#bordered)

添加完成后，可在文件管理器中确认挂载路径。
![在文件管理器中确认挂载路径](/images/manual/olares/studio-check-mounted-path-in-files.png#bordered)

### 可选：配置 GPU 或数据库中间件
如果应用依赖 GPU，需要在**实例规格**下启用 **GPU** 选项并选择 GPU 厂商。
![启用 GPU](/images/manual/olares/studio-enable-GPU.png#bordered)

如果应用需要 Postgres 或 Redis 数据库，在**实例规格**下启用相应选项。
![启用数据库](/images/manual/olares/studio-enable-databases.png#bordered)

启用数据库后，Studio 会提供一组动态变量。你必须在应用的**环境变量**中添加这些变量，应用才能连接到数据库。
- **Postgres 变量**

| 变量名            | 说明             |
|----------------|----------------|
| `$(PG_USER)`   | PostgreSQL 用户名 |
| `$(PG_DBNAME)` | 数据库名称          |
| `$(PG_PASS)`   | Postgres 密码    |
| `$(PG_HOST)`   | Postgres 服务主机名 |
| `$(PG_PORT)`   | Postgres 服务端口  |

- **Redis 变量**

| 变量名             | 说明          |
|-----------------|-------------|
| `$(REDIS_HOST)` | Redis 服务主机名 |
| `$(REDIS_PORT)` | Redis 服务端口  |
| `$(REDIS_USER)` | Redis 用户名   |
| `$(REDIS_PASS)` | Redis 密码    |

### 生成应用项目
1. 完成所有配置后点击**创建**，系统将生成应用的项目文件。
2. 创建完成后，Studio 会自动打包并部署应用。你可以在页面底部栏查看进度状态。
3. 部署成功后，在右上角点击**预览**即可打开应用。
   ![预览 Wallos](/images/manual/olares/studio-preview-wallos.png#bordered)

## 检查包文件与测试应用
通过 Studio 部署的应用标题会自动添加 `-dev` 后缀，以便与从应用市场安装的正式版区分。
![检查部署的应用](/images/manual/olares/studio-app-with-dev-suffix.png#bordered)

你可以查看或编辑 `OlaresManifest.yaml` 等配置文件。例如，修改应用的显示名称和图标：

1. 在右上角点击 **<span class="material-symbols-outlined">box_edit</span> 编辑**打开编辑器。
2. 点击 `OlaresManifest.yaml` 查看内容。
3. 修改 `entrance` 和 `metadata` 部分的 `title` 字段。例如，将 `wallos` 改为 `Wallos`。
4. 替换 `entrance` 和 `metadata` 部分的图标地址。
   ![编辑 `OlaresManifest.yaml`](/images/manual/olares/studio-edit-olaresmanifest1.png#bordered)

5. 在右上角点击 <span class="material-symbols-outlined">save</span> 保存更改。 
6. 点击**应用**，系统将使用更新后的配置重新安装应用。

   :::info
   如果自上次部署后未检测到任何更改，点击**应用**将直接返回应用状态页，不会触发重装。
   :::
   ![修改应用图标](/images/manual/olares/studio-change-app-icon1.png#bordered)

## 卸载或删除应用
如果不再需要该应用，可执行以下操作：
1. 在右上角点击 <span class="material-symbols-outlined">more_vert</span>。
2. 选择相应操作：
   - **卸载**：仅从 Olares 系统中移除运行的实例，但在 Studio 中保留项目文件，方便继续编辑。
   - **删除**：卸载应用并从 Studio 中彻底删除项目文件。此操作不可撤销。

## 部署故障排除

### 无法安装应用
如果安装失败，请查看页面底部的错误提示。点击**查看**可打开详细日志。

### 运行时遇到问题
应用启动后，你可以在 Studio 的部署详情页进行管理（界面类似于控制面板）。如果详情未显示，刷新页面即可。
常见操作包括：
- 点击**停止**按钮或**重启**按钮后重试。这通常能解决进程卡死等临时性故障。
- 查看事件或日志以排查错误。具体方法请参考[导出问题容器日志](../../../manual/olares/controlhub/manage-container.md#导出问题容器日志)。
  ![应用部署详情](/images/manual/olares/studio-app-deployment-details.png#bordered)