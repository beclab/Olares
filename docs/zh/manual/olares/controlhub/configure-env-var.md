---
outline: [2, 3]
description: 了解如何在 Control Hub 中定位和修改应用环境变量，用于调试、更新或配置更改。
---

# 配置环境变量

通过控制面板，你可以查看和修改应用的环境变量，以满足调试、功能配置或临时调整的需求。

## 开始之前

在进行修改前，先确认目标变量应在何处配置。

| 对比项 | 系统级变量 | 应用级变量 |
|:---|:---|:---|
| 配置位置 | 设置 | 控制面板 |
| 作用范围 | 所有引用了该变量的应用共享 | 仅作用于单个应用资源 |
| 涵盖内容 | Olares 预置的常用变量，如 API 密钥、邮件服务器设置等 | 系统级变量未覆盖的应用特定参数 |
| 应用升级后是否保留 | 是 | 否 |

先查阅 [设置系统级环境变量](/zh/manual/olares/settings/developer.md#设置系统级环境变量)，确认你需要的变量是否已存在。如果没有，使用以下步骤在控制面板中直接配置。

## 确定变量的存储位置

应用中的变量可以存储在不同类型的 Kubernetes 资源中。了解目标变量存储在哪个资源中，决定了你如何修改它以及是否需要重启。

| 资源类型 | 典型内容 | 修改后是否需要重启 |
|:---|:---|:---|
| 部署 | 直接的键值对 | 自动重启 |
| 配置字典 | 配置数据、启动参数、配置文件 | 需手动重启 |
| 保密字典 | 敏感数据，如密码、令牌、凭证 | 需手动重启 |

### 常规应用

要确定变量存储在哪里，在 YAML 编辑器中打开部署，查看 `spec` > `containers` 部分。注入方式会告诉你变量的来源：

- `env`：变量直接在部署中以键值对的形式定义。
- `envFrom` 配合 `configMapRef`：变量存储在引用的 Configmap 中。
- `valueFrom` 配合 `secretKeyRef`：变量存储在引用的 Secret 中。

例如，在下面的 YAML 中，`envFrom` 引用了 Configmap `lobechat-config`，而 `env` 直接定义了 `PGID`、`PUID` 和 `TZ`。

```yaml{9-18}
spec:
  containers:
    - name: lobechat
      image: docker.io/beclab/lobehub-lobehub:2.1.18
      ports:
        - name: http
          containerPort: 3210
          protocol: TCP
      envFrom:
        - configMapRef:
            name: lobechat-config
      env:
        - name: PGID
          value: '1000'
        - name: PUID
          value: '1000'
        - name: TZ
          value: Etc/UTC
```

### C/S 架构应用

部分应用采用客户端/服务器（C/S）架构，例如 Ollama。它们的变量分布在两个不同的命名空间中：用户命名空间用于客户端资源，系统命名空间用于服务端资源。

对于 C/S 应用，环境变量通常存储在跨两个命名空间的 Configmap 中。根据你想要修改的内容导航到正确的命名空间：

- 要修改外部访问行为，前往用户命名空间，查找名称中包含 `nginx` 或 `sidecar` 的 Configmap。
- 要修改应用核心参数，前往系统命名空间，查找名称中包含 `env`、`config` 或类似标识的 Configmap。

## 修改部署中的变量

此方法适用于直接调整工作负载。下面的示例更改 Jellyfin 的时区，使媒体库显示本地时间戳而不是 UTC。

1. 在控制面板的**浏览**面板中，选择 Jellyfin 项目。

2. 在**部署**下，点击 **jellyfin**，然后点击 <i class="material-symbols-outlined">edit_square</i>。

    <!--![浏览到 Jellyfin 的部署实例](/images/zh/manual/olares/jellyfin-env-var.png#bordered)-->

3. 在 YAML 编辑器中，找到 `containers` 部分，定位 jellyfin 的 `env` 字段，然后更改 `TZ` 的值：

    ```yaml
    env:
      - name: PGID
        value: '1000'
      - name: PUID
        value: '1000'
      - name: UMASK
        value: '002'
      - name: TZ
        value: Asia/Shanghai   # 原为 Etc/UTC
    ```
4. 点击 **Confirm**。Pod 将自动重启以使更改生效。

## 修改配置字典中的变量

此方法用于添加第三方 API 密钥、修改启动参数或更新配置文件。

### 常规应用 {#modify-standard-apps}

以下示例向 DeerFlow 的配置中添加 Tavily API 密钥，以启用网络搜索。

1. 在控制面板的**浏览**面板中，选择 DeerFlow 项目。
2. 在**配置字典**下，点击 `deerflow-config`。
    ![浏览到 DeerFlow 的配置字典](/images/zh/manual/use-cases/deerflow-configmap.png#bordered)

3. 在资源详情页，点击右上角的 <i class="material-symbols-outlined">edit_square</i> 打开 YAML 编辑器。
4. 在 `data` 部分下添加以下键值对：
   ```yaml
   SEARCH_API: tavily
   TAVILY_API_KEY: tvly-xxx # 你的 Tavily API Key
   ```
   ![配置 Tavily](/images/zh/manual/use-cases/deerflow-configure-tavily.png#bordered)
5. 点击 **Confirm** 保存更改。
6. 返回**部署** > **deerflow**，然后点击**重启**。

   ![重启 DeerFlow](/images/zh/manual/use-cases/deerflow-restart.png#bordered)

7. 在确认对话框中输入 `deerflow`，然后点击 **Confirm**。

等待状态图标变为绿色，表示新配置已加载。

### C/S 架构应用 {#modify-cs-apps}

根据你要修改的内容，你可能需要在用户命名空间、系统命名空间或两者中修改变量。

#### 在用户命名空间中修改客户端设置

以下步骤将 Ollama 的代理读取超时从 `300s` 改为 `600s`。

1. 在控制面板的**浏览**面板中，选择 Ollama 项目。

2. 在**部署**下，点击 **ollamav2**，然后点击 <i class="material-symbols-outlined">edit_square</i>。
   ![在控制面板中定位 Ollama 实例](/images/zh/manual/use-cases/locate-ollama-instance.png#bordered)

3. 在 YAML 编辑器中，找到 `containers` 部分，定位并检查 `env` 字段。这里的配置引用了 `nginx.conf`。

   ![编辑 Ollama 实例的 YAML](/images/zh/manual/use-cases/edit-yaml-ollama.png#bordered)

4. 点击 **Cancel** 关闭编辑器。

5. 展开**配置字典**资源组，点击 `nginx-config`，然后点击右上角的 <i class="material-symbols-outlined">edit_square</i>。

   ![定位 Nginx 配置实例](/images/zh/manual/use-cases/locate-nginx-config.png#bordered)

6. 在 YAML 编辑器中，找到 `data` 部分，定位 `nginx.conf` 键，然后将 `proxy_read_timeout` 的值从 `300s` 改为 `600s`。

   ![编辑 Nginx 配置](/images/zh/manual/use-cases/edit-nginx-conf.png#bordered)

7. 点击 **Confirm**。

8. 返回**部署** > **ollamav2**，然后点击**重启**以使更改生效。

等待状态图标变为绿色，表示新配置已加载。

#### 在系统命名空间中修改服务端设置

以下步骤将 Ollama 的最大加载模型数从 `3` 改为 `5`。

1. 在控制面板的**浏览**面板中，向下滚动并点击 **System** 展开系统部分。

2. 选择 **ollamaserver-shared**，然后在**部署**下点击 **ollama**，再点击 <i class="material-symbols-outlined">edit_square</i>。

   ![在系统命名空间中定位 Ollama](/images/zh/manual/use-cases/locate-ollama-sys-namespace.png#bordered)

3. 在 YAML 编辑器中，找到 `containers` 部分，检查 `envFrom` 字段。这里的配置引用了 `ollama-env`。

   ![编辑 YAML](/images/zh/manual/use-cases/edit-yaml-envfrom.png#bordered)

4. 点击 **Cancel** 关闭编辑器。

5. 返回**配置字典**资源组，点击 `ollama-env` 实例，然后点击右上角的 <i class="material-symbols-outlined">edit_square</i>。

   ![编辑 Ollama 环境变量](/images/zh/manual/use-cases/edit-ollama-env.png#bordered)

6. 在 YAML 编辑器中，找到 `data` 部分，然后将 `OLLAMA_MAX_LOADED_MODELS` 的值从 `3` 改为 `5`。

   ![编辑 Ollama 变量](/images/zh/manual/use-cases/modify-var-ollama.png#bordered)

7. 点击 **Confirm**。

8. 返回**部署** > **ollama**，然后点击**重启**以使更改生效。

等待状态图标变为绿色，表示新配置已加载。

## 修改保密字典中的变量

修改保密字典的流程与修改配置字典相同。根据你的应用类型，遵循[常规应用](#modify-standard-apps)或 [C/S 架构应用](#modify-cs-apps)中的步骤。

:::info
当你打开保密字典的 YAML 编辑器时，`data` 字段下的所有值都必须是 Base64 编码的。
:::

## 常见问题

### 修改配置字典或保密字典后，更改未生效

修改配置字典或保密字典后，关联的工作负载不会自动重新加载配置。你必须重启工作负载才能读取新的值。

使用以下任一方法重启工作负载：

- **在控制面板中重启**
  
  进入应用所在命名空间的**部署**资源组，点击目标工作负载，然后点击**重启**。

- **通过应用市场或设置重启**
  
  如果不确定应该重启哪个工作负载，可以停止再恢复该应用：
   - 进入**应用市场** > **我的 Olares**，点击应用操作按钮旁边的 <i class="material-symbols-outlined">keyboard_arrow_down</i>，选择**暂停**，然后再点击**恢复**。
   - 进入**设置** > **应用**，点击应用，点击**暂停**，然后点击**恢复**。

这两种方法都会应用并加载配置字典或保密字典中的最新配置。
