---
description: 排查在 Olares 上运行 OpenClaw 时的常见问题，并获取常见问题的解答。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, 故障排除, 常见问题, 常见错误, 报错
app_version: "1.0.8"
doc_version: "1.2"
doc_updated: "2026-06-10"
---

# OpenClaw 常见问题

本页面整理了在 Olares 上运行 OpenClaw 时的常见问题及解决方法。

如遇到此处未列出的问题，参考[升级 OpenClaw](openclaw-upgrade.md)页面了解版本特定的更改，或[ OpenClaw 官方文档](https://docs.openclaw.ai/zh-CN)。

## 升级后智能体提示 "Missing API key" 错误且无响应

使用云端模型提供商时，智能体无法连接或调用外部 API，并报如下错误：

```text
Missing API key for the selected provider on the gateway. Configure provider
auth, then try again.
```

### 原因

从 OpenClaw V2026.06.05 开始，认证配置文件和核心结构配置从旧版 JSON 文件迁移到内部的 SQLite 数据库。

如果你近期升级了应用，但尚未执行数据库迁移工具，网关会在新数据库中查找认证配置文件，从而无法检测到云端 API 密钥。

### 解决方案

在 OpenClaw CLI 中运行自动修复工具，将旧版认证 JSON 文件迁移到 SQLite 数据库。

1. 从启动台打开 OpenClaw CLI。
2. 运行自动修复命令：

    ```bash
    openclaw doctor --fix
    ```
3. 在终端中查看输出日志。迁移成功时，会确认旧版 JSON 配置文件已导入 SQLite。

    示例输出：

    ```text
    |  Migrated auth profile JSON for ~/.openclaw/agents/main/agent/auth-profiles.json into  |
    |  SQLite (backups:                                                                      |
    |  ~/.openclaw/agents/main/agent/auth-profiles.json.sqlite-import.1781088154476.bak,     |
    |  ~/.openclaw/agents/main/agent/auth-state.json.sqlite-import.1781088154484.bak).       |
    ```

    完成后，网关将自动识别密钥并正常连接。

## 无法在 CLI 中重启 OpenClaw

如果在 OpenClaw CLI 中使用 `openclaw gateway restart` 或 `openclaw gateway stop` 等标准命令手动启动、停止或重启 OpenClaw，可能会收到类似以下报错：

- `Gateway service disabled`
- `Gateway failed to start: gateway already running (pid 1); lock timeout after 5000ms`
- `Gateway service check failed: Error: systemctl --user unavailable: spawn systemctl ENOENT`

### 原因

OpenClaw 在 Olares 中以容器化应用部署，网关作为容器主进程 `pid 1` 持续运行。此环境不使用 `systemd` 和 `systemctl` 等标准 Linux 系统和服务管理工具，因此默认的 `openclaw gateway` 命令无法生效。

### 解决方案

不要使用标准的 `openclaw gateway` 命令。通过以下方法之一重启 OpenClaw。

**方法 1：通过 OpenClaw CLI 快速重启（推荐）**

使用内置的 `restart-gateway` 脚本。此命令可安全关闭正在运行的网关，应用最新配置，并快速恢复智能体在线（通常 5 秒内完成）。

1. 打开 OpenClaw CLI。
2. 运行以下命令：

    ```bash
    restart-gateway
    ```

    终端将显示重启进度，并在网关恢复在线后进行确认：

    ```text
    gateway: restart requested
    gateway: old process gone, waiting for new one
    gateway: ready
    ```

**方法 2：从设置或应用市场重启 OpenClaw**
    
- 打开设置，进入**应用** > **OpenClaw**，点击**暂停**，然后点击**恢复**。
- 打开应用市场，进入**我的 Olares**，找到 **OpenClaw**，点击操作按钮旁边的 <i class="material-symbols-outlined">keyboard_arrow_down</i>，选择**暂停**，然后选择**恢复**。

**方法 3：重启容器**

打开控制面板，点击**部署**下的 `clawdbot`，然后点击右上角的**重启**。

## OpenClaw 在长任务期间自动停止

当你要求 OpenClaw 智能体执行耗时较长的任务（如大规模网页抓取或深度分析）时，任务可能在返回结果之前突然终止。

### 原因

默认情况下，OpenClaw 为每个任务设置最长 10 分钟的运行时限。如果任务超过此时限，系统将强制终止该任务以节省资源。

### 解决方案

按如下方式修改配置文件以延长此超时限制：
1. 打开 Control UI，进入 **Config** > **Raw**，然后找到 `agents` 部分。
2. 在 `defaults` 块中，添加或修改 `timeoutSeconds` 字段。

    例如，要设置为 1 小时，将值指定为 `3600`：

    ```json
    "agents": {
        "defaults": {
            "timeoutSeconds": 3600
        }
    }
    ```
3. 点击 **Save** 以重启网关并应用更改。

## 安装技能时出现 "Rate limit exceeded" 错误

安装技能失败并返回 `429` 错误：

```text
Downloading xurl@1.0.0 from ClawHub...
ClawHub /api/v1/download failed (429): Rate limit exceeded
```

### 原因

ClawHub 注册表因高流量暂时限制下载，以维持服务器稳定性。

### 解决方案

等待几小时后再次运行安装命令。

## 模型响应缓慢

智能体开始输出第一个响应之前有明显的延迟。

### 原因

这通常是由于 Ollama 管理系统资源和应用设置的方式导致：
- **自动卸载**：为节省资源，Ollama 默认在模型空闲时将其从内存中卸载。下次与模型交互时，需要重新加载，导致第一个响应出现明显的延迟。
- **上下文设置冲突**：如果多个应用使用同一模型但上下文设置不同，Ollama 将被迫不断卸载和重新加载模型以在不同配置之间切换。

### 解决方案

要修复此问题，尝试以下方法之一。

#### 方法 1：为模型应用启用常驻内存

通过为模型应用启用 `KEEP_ALIVE` 环境变量，使模型常驻内存。

1. 打开设置，然后进入**应用** > **{你的模型应用}** > **管理环境变量**。
2. 找到 **KEEP_ALIVE**，点击 <i class="material-symbols-outlined">edit_square</i>，将值设置为 **true**，然后点击**确认**。

    ![在设置中为模型应用启用 Keep Alive](/images/zh/manual/use-cases/keep-alive-enable.png#bordered){width=70%}

3. 点击**应用**。

#### 方法 2：统一各应用的上下文大小

为所有共享同一模型的应用设置相同的上下文大小，以减少重新加载的时间。

1. 检查正在运行的模型的当前上下文大小：

    - 在 Ollama 终端中，运行 `ollama ps` 命令。`CONTEXT` 列显示正在使用的上下文大小。

        ![在 Ollama 终端中查看模型详情](/images/manual/use-cases/ollama-ps.png#bordered)

    - 对于独立的模型应用，使用控制面板检查上下文大小：
    
        a. 在 **System** 命名空间下，找到模型应用的项目（通常名为 `{model-name}server-shared`），然后打开其 Pod 终端。

        ![在控制面板中打开 Pod 终端](/images/zh/manual/use-cases/pod-terminal-ctrl-hub.png#bordered)        
        
        b. 运行 `ollama ps` 命令。

        ![在控制面板中查看模型详情](/images/zh/manual/use-cases/ollama-ps-ctrl-hub.png#bordered)

2. 将所有应用设置为使用相同的上下文大小。

## 彻底重装 OpenClaw

如果你想卸载 OpenClaw 并重新开始，仅卸载应用是不够的。默认情况下，Olares 会保留你的应用数据（如配置和人设文件），以免丢失已有工作成果。

要在重新安装前彻底移除 OpenClaw 及其所有数据，可根据你的 Olares OS 版本执行以下相应步骤。

<Tabs>
<template #V1.12.5-及更高版本>

1. 打开应用市场，进入**我的 Olares**，然后找到 **OpenClaw**。
2. 点击应用操作按钮旁边的 <i class="material-symbols-outlined">keyboard_arrow_down</i>，然后选择**卸载**。
3. 在**卸载**窗口中，勾选**同时删除所有本地数据**。随后应用数据（在数据目录中）和缓存数据（在缓存目录中）将被永久删除，无法恢复。
    ![卸载时移除本地应用数据选项](/images/zh/manual/use-cases/uninstall-remove-local-data.png#bordered){width=65%}
4. 点击**确认**。
5. 返回应用市场并重新安装 OpenClaw。此时将以完全干净的状态安装。
</template>
<template #V1.12.4-及更早版本>

1. 打开应用市场，进入**我的 Olares**，然后找到 **OpenClaw**。
2. 点击应用操作按钮旁边的 <i class="material-symbols-outlined">keyboard_arrow_down</i>，然后选择**卸载**。
3. 卸载完成后，打开文件管理器，然后进入**应用** > **数据**。
4. 找到 `clawdbot` 文件夹，右键点击它，选择**删除**，然后点击**确认**。这将永久移除所有先前的配置和工作空间。
    ![移除 OpenClaw 应用数据](/images/zh/manual/use-cases/remove-app-data.png#bordered){width=80%}
5. 返回应用市场并重新安装 OpenClaw。此时将以完全干净的状态安装。
</template>
</Tabs>
