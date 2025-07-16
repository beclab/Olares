---
outline: [2, 3]
---

# 如何开始开发一个应用

## 下载并安装 DevBox

1. 从 Olares 应用市场安装 [DevBox](https://market.olares.com/app/devbox)。
2. 在 Olares 桌面的 LaunchPad 中找到 DevBox 图标。
3. 点击图标启动应用程序。

  ![main screen](/images/developer/develop/tutorial/create/home.jpg)

## 创建应用

点击 **Create a new application**，从模版创建一个空白的 Olares 应用。
- 在 **App Name** 一栏，输入你的应用名称。
- 设置 **App type** 为 **app**。
- 修改 APP 入口的端口。
- **Image** 一栏填写应用将要推送的镜像仓库的镜像名称和 tag。

![create app](/images/developer/develop/tutorial/create/create.jpg)

## 设置应用配置

创建应用程序后，你可以在 **Files** 选项卡下看到 DevBox 生成的 Olares Application Chart 文件。你可以根据需要添加、删除或重命名各种配置文件。

![upload icon](/images/developer/develop/tutorial/create/add-file.jpg)

### Chart.yaml

`Chart.yaml `文件是 Helm Chart 规范所必须的文件之一。其中包含了应用的名称和 Chart Version，你可以[在此](https://helm.sh/docs/topics/charts/)了解更多。我们暂时先不用修改默认创建的`Chart.yaml`。

### OlaresManifest.yaml
在 `OlaresManifest.yaml` 文件中，你可以更改许多配置。例如：
- 更改应用的标题、图标和其他详细信息
- 添加系统中间件
- 获取系统目录访问的权限
- 更改应用程序的所需的资源限制

#### 添加系统的[数据库集群需求](../../package/manifest.md#middleware)

![config app](/images/developer/develop/tutorial/create/olares-manifest.jpg)

此处我们需要申请一个 PostgreSQL 的数据库，在配置文件中添加以下内容。
```Yaml
middleware:
  postgres:
    username: postgres
    databases:
    - name: db
      distributed: false
```

申请时，需要定义你的数据库访问用户名。也可以自定以密码（只需要添加一个 password 申明），也可以由系统生成随机密码。这里需要设置你的 APP 需要的 database name。另外，还可以选择申请一个分布式数据库。如果选在分布式数据库，系统会为你创建一个[citus](https://github.com/citusdata/citus)数据库

完成配置后，可在你的 deployment 中引用对应的数据库配置。例如，在容器的环境变量中引用。
```yaml
- env:
    - name: DB_PORT
      value: "{{ .Values.postgres.port }}"
    - name: DB_NAME
      value: "{{ .Values.postgres.databases.demo }}"
    - name: DB_USER
      value: "{{ .Values.postgres.username }}"
    - name: DB_HOST
      value: "{{ .Values.postgres.host }}"
    - name: DB_PWD
      value: "{{ .Values.postgres.password }}"
```
  - `.Values.postgres.username`：对应申请 PostgreSQL 中的 username。
  - `.Values.postgres.databases.demo`：对应申请中的 database name。
  - `.Values.postgres.password`：对应申请中的 password
  - `.Values.postgres.host`：系统为 APP 指定的数据库服务地址
  - `.Values.postgres.port`：系统为 APP 指定的数据库服务的端口

::: warning
这些参数不应该被硬编码，它们必须引用系统传入的变量，并且系统会随机化配置中的数据库信息。
:::

#### 申请系统的[文件系统访问权限](../../package/manifest.md#permission)

为了能在 Olares 中读取和保存文件，我们需要在 `Permissions` 一项中，配置所需的文件目录。`OlaresManifest.yaml`提供了三个位置的文件目录，分别是：
- `appData`: 申请应用独立数据云存储空间。
- `appCache`: 给应用申请节点本地磁盘（一般为 SSD 磁盘）数据缓存空间。
- `userData`: 申请用户的数据目录访问权限。可列举需要访问的目录列表。

完成上述配置后，就可以在你的 deployment 中引用这些配置。

```yaml
volumes:
  - hostPath:
      path: "{{ .Values.userspace.appCache }}/demo"
      type: DirectoryOrCreate
    name: appcache
  - hostPath:
      path: "{{ .Values.userspace.appData }}/demo"
      type: DirectoryOrCreate
    name: appdata
```
  - `.Values.userspace.appCache` 对应 appCache 目录
  - `.Values.userspace.appData` 对应 appData 目录
  - `.Values.userspace.userData` 对应 userData 目录

### deployment.yaml

`templates` 文件夹中的 `deployment.yaml` 详细描述了应用的部署配置。

如果你的应用分为前后端两个不同的容器。你可以在 templates 的部署文件中，添加多个容器。DevBox 将识别这些不同的容器并将它们分别绑定到不同的开发容器。例如，
```yaml
containers:
  # 前端容器
  - env:
      - name: PGID
        value: "1000"
      - name: PUID
        value: "1000"
      - name: TZ
        value: Etc/UTC
    image: bytetrade/demo-app:0.0.1
    name: demo
    ports:
      - containerPort: 8080
    resources:
      limits:
        cpu: "1"
        memory: 2000Mi
      requests:
        cpu: 50m
        memory: 1000Mi
    volumeMounts:
      - mountPath: /appcache
        name: appcache

  # Server 端容器
  - env:
      - name: DB_PORT
        value: "{{ .Values.postgres.port }}"
      - name: DB_NAME
        value: "{{ .Values.postgres.databases.demo }}"
      - name: DB_USER
        value: "{{ .Values.postgres.username }}"
      - name: DB_HOST
        value: "{{ .Values.postgres.host }}"
      - name: DB_PWD
        value: "{{ .Values.postgres.password }}"
      - name: PGID
        value: "1000"
      - name: PUID
        value: "1000"
      - name: TZ
        value: Etc/UTC
    image: bytetrade/demo-server:0.0.1
    name: server
    ports:
      - containerPort: 9000
    resources:
      limits:
        cpu: "1"
        memory: 1000Mi
      requests:
        cpu: 50m
        memory: 500Mi
    volumeMounts:
      - mountPath: /appcache
        name: appcache
      - mountPath: /appdata
        name: appdata
```

## 绑定容器
配置完上述信息后，你需要在 **Containers** 页面为这个开发中的应用，绑定开发容器，进行代码开发。

![containers](/images/developer/develop/tutorial/create/bind.jpg)

你可以为绑定的开发容器设置一个指定的开发环境，目前 DevBox 支持 NodeJS、Golang、python 三种开发容器。让我们给 demo 前端容器绑定一个 NodeJS 开发容器，给 Server 容器绑定一个 Golang 的开发容器。

这里选择了创建一个新的开发容器。如果之前已经创建过未绑定的开发容器，这里也可以选择一个已有的容器进行绑定。

![bind container](/images/developer/develop/tutorial/create/bind-2.jpg)

## 安装应用
完成开发容器绑定后，就可以点击右上角的安装将当前应用安装到系统中了。等待安装状态从 `Processing` 变为 `Running` 即表示应用已安装完成，可进入正式的代码开发流程。

![installing](/images/developer/develop/tutorial/create/installing.jpg)

此时，再次进入 Containers 页面，可以看到开发容器上的 **Open IDE** 按钮激活。点击即可进入开发环境进行代码开发。

![processing](/images/developer/develop/tutorial/create/success.jpg)
