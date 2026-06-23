---
outline: deep
description: 在 Olares 上部署 Falco，实时监控 Linux 内核事件，检测主机、容器和 Kubernetes 工作负载中的运行时安全威胁。
head:
  - - meta
    - name: keywords
      content: Olares, Falco, runtime security, eBPF, Kubernetes, container security, threat detection, Falcosidekick
app_version: "1.0.11"
doc_version: "1.0"
doc_updated: "2026-04-23"
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../use-cases/falco.md)为准。
:::

# 使用 Falco 监控运行时安全

Falco 是一个基于 eBPF 的开源云原生运行时安全工具。它实时监控 Linux 内核事件，当检测到主机、容器或 Kubernetes 工作负载中的可疑行为时触发告警。

在 Olares 上，Falco 作为共享应用运行。代理在每个节点上收集事件，中央 Falcosidekick UI 将所有内容汇集到一个地方。

当你希望在 Olares 上安装 Falco 并查看来自主机、容器和 Kubernetes 工作负载的运行时安全告警时，请使用本指南。

## 学习目标

在本指南中，你将学习如何：
- 在 Falcosidekick UI 中查看安全告警。
- 配置事件保留、检测规则、输出通道和插件。
- 排查常见插件问题。

## Prerequisites

- **需要管理员权限**：Falco 采用客户端/服务器架构运行，只有管理员才能安装或配置它。如果你是普通用户，请要求你的管理员先安装 Falco 共享应用。

## Falco 在 Olares 上的工作原理

Falco 使用分布式收集和集中显示，因此你可以从一个 UI 监控每个节点。

### 组件

| 组件 | 类型 | 角色 |
|:----------|:-----|:-----|
| `falco-agent` | DaemonSet | 在每个节点上运行，捕获匹配 Falco 规则的内核事件。 |
| `falco-plugin-installer` | DaemonSet | 用于安装插件和规则的工具箱。附带 `falcoctl`。 |
| `falco-sidekick` | Service | 接收来自所有 `falco-agent` 实例的 HTTP 输出。 |
| `webui` | Service | 提供仪表板和事件视图。 |

### 事件流

1. 每个节点上的 `falco-agent` 在本地捕获内核事件。
2. 当事件匹配规则时，`http_output` 将其转发到 `falco-sidekick`。
3. `falco-sidekick` 将事件写入 Redis。
4. Falcosidekick UI 从 Redis 读取数据并渲染仪表板。

## 安装 Falco

1. 打开 Market，搜索 "Falco"。

  ![Market 中的 Falco](/images/manual/use-cases/falco.png#bordered){width=90%}

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 在 Falcosidekick UI 中查看告警

Falco 安装完成后，打开 Falco 应用以访问 Falcosidekick UI 并查看安全告警。

Falcosidekick UI 是在 Olares 上查看告警的默认位置。如有需要，管理员还可以将告警转发到外部系统。参见 [配置输出通道](#配置输出通道)。

### 仪表板

**Dashboard** 页面提供跨节点的告警活动实时概览。

  ![Falco 仪表板](/images/manual/use-cases/falco-dashboard.png#bordered){width=90%}

| 面板 | 显示内容 |
|:------|:--------------|
| Global statistics | 所选时间窗口的聚合告警计数。 |
| Filter bar | 按来源、优先级或标签缩小结果范围。 |
| Snapshot counters | 当前筛选器下的实时总数：`Total`、`Critical` 和 `Notice`。 |
| Pie chart | 按来源、优先级和标签分布的告警。 |
| Rule bar chart | 按规则分组的告警。有助于发现需要加入允许列表或调整阈值的嘈杂规则。 |
| Timeline by priority | 按优先级划分的时间线告警量。 |
| Timeline by source | 按来源划分的时间线告警量。 |

### 事件

**Events** 页面列出每个告警及其完整上下文。

  ![Falco 事件](/images/manual/use-cases/falco-events.png#bordered){width=90%}

| 列 | 描述 |
|:-------|:------------|
| Timestamp | 告警生成时间，例如 `2026-04-14 20:35:37`。 |
| Source | 事件来源。 |
| Hostname | 与告警关联的主机。 |
| Priority | 告警严重级别，按颜色编码。 |
| Rule | Falco 规则库中的规则名称。 |
| Output | 包含上下文变量的完整告警消息。 |
| Tags | 分类标签。 |

要详细检查告警：
1. 在 **Events** 页面，找到你要检查的告警。
2. 点击该行右侧的 **{…}**。
3. 查看详情面板。如果需要原始负载，请切换到 **JSON** 标签页。

## 配置 Falco

当你需要更改 Falco 存储、检测或转发告警的方式时，请使用本节。

:::warning 仅管理员
配置需要管理员权限。普通用户无法更改 Falco 设置。
:::


| 区域 | 控制内容 |
|:-----|:-----------------|
| [事件保留](#设置事件保留) | 告警在 Falcosidekick UI 中保留多长时间后清理。 |
| [检测规则](#管理检测规则) | 哪些行为会触发告警。 |
| [输出通道](#配置输出通道) | 告警发送到哪里（Falcosidekick UI、文件、外部系统）。 |
| [插件](#安装和使用插件) | 额外的事件源，例如 Kubernetes 审计日志。 |

### 设置事件保留

Falco 默认保留告警 72 小时。要更改告警保留时间：

1. 前往 **Settings** > **Applications** > **Falco** > **Manage environment variables**。
2. 点击 `FALCOSIDEKICK_UI_TTL` 旁边的 <i class="material-symbols-outlined">edit_square</i>。
3. 输入带单位后缀的时长，例如 `7d` 表示七天。支持的后缀包括 `s`、`m`、`h`、`d`、`w`、`M` 和 `y`。留空以无限期保留事件。
4. 点击 **Confirm**，然后点击 **Apply**。

  ![编辑 FALCOSIDEKICK_UI_TTL](/images/manual/use-cases/falco-edit-ttl.png#bordered){width=90%}

5. 可选：要验证新值是否已应用，打开 Control Hub，然后前往 **Browse** > **System** > **falcoserver-shared** > **Deployments** > **falco-central**。
   - 点击 <i class="material-symbols-outlined">edit_square</i> 打开 YAML 文件，找到 `FALCOSIDEKICK_UI_TTL`，检查其值。
   - 在右侧面板中，**Environment variables** 下，点击 **webui** 并检查 `FALCOSIDEKICK_UI_TTL` 的值。

### 管理检测规则

Falco 使用规则来决定哪些行为应该生成告警。

当你想要执行以下操作时使用本节：
- 检查当前加载了哪些规则文件。
- 添加自定义规则。
- 禁用在你环境中不相关的规则。

:::warning 需要重启
规则更改仅在重启 `falco-agent` DaemonSet 后生效。

规则名称必须唯一且完全匹配。不匹配的规则名称可能会阻止 `falco-agent` 启动。
:::

#### 了解规则格式

Falco 规则通常包括：

```yaml
- rule: Test - Terminal Shell In Container
  desc: Test rule to validate Falco custom rules pipeline
  condition: container and shell_procs and proc.name in (bash, sh, zsh)
  output: >
    TEST custom rule matched (user=%user.name command=%proc.cmdline container=%container.id image=%container.image.repository)
  priority: WARNING
  tags: [container, test]
```

| 字段 | 描述 |
|:------|:------------|
| `rule` | 唯一规则名称。 |
| `desc` | 简短描述。 |
| `condition` | 触发告警的条件。 |
| `output` | 告警消息模板。支持 `%proc.cmdline` 等字段。 |
| `priority` | 告警严重级别。其中之一：<br>`EMERGENCY`、`ALERT`、`CRITICAL`、`ERROR`、`WARNING`、`NOTICE`、<br> `INFORMATIONAL`、`DEBUG`。 |
| `tags` | 用于筛选和分组的标签。 |

#### 检查已加载的规则文件

Falco 在启动时加载规则文件。加载的确切文件集取决于你当前的配置。

例如，Falco 可能加载：

| 文件 | 用途 |
|:-----|:--------|
| `falco_rules.yaml` | Falco 提供的上游默认规则。 |
| `custom_rules.yaml` | 你为自有环境添加的自定义规则。 |
| `falco_disable_rules.yaml` | 你显式禁用的规则。 |

要检查当前加载了哪些规则文件：

1. 打开 Control Hub。
2. 前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。
3. 点击你的 pod 打开详情面板。
4. 在 **Containers** 下，点击 **falco** 旁边的 <i class="material-symbols-outlined">article</i> 打开日志。

  ![查看活动规则](/images/manual/use-cases/falco-view-rules.png#bordered){width=90%}

5. 在日志中查找 `Loading rules from:` 部分。

#### 查看默认规则

要查看默认 Falco 规则：

1. 打开 Control Hub，然后前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。
2. 点击你的 pod 打开详情面板。
3. 在 **Containers** 下，点击 **falco** 旁边的 <i class="material-symbols-outlined">terminal</i> 打开终端。
4. 运行以下命令：

    ```bash
    cat /etc/falco/falco_rules.yaml
    ```

#### 创建自定义规则

1. 打开 Control Hub，然后前往 **Browse** > **System** > **falcoserver-shared** > **Configmaps** > **falco-custom-rules**。
2. 在右侧面板中，点击 **falco-custom-rules** 旁边的 <i class="material-symbols-outlined">edit_square</i>。
3. 将 `custom_rules.yaml:` 改为 `custom_rules.yaml: |`，然后在下一行添加你的规则。

   示例：

   ```yaml
   data:
     custom_rules.yaml: |
       - rule: Test - Terminal Shell In Container
         desc: Test rule to validate the custom rule pipeline
         condition: >
           evt.type in (execve, execveat)
           and container
           and shell_procs
           and proc.name in (bash, sh, zsh)
           and k8s.ns.name exists
           and not (k8s.ns.name in ("kube-system", "falco", "falcoserver-shared"))
         output: >
           TEST custom rule matched (ns=%k8s.ns.name user=%user.name command=%proc.cmdline container=%container.id image=%container.image.repository)
         priority: WARNING
         tags: [container, test]
    ```
4. 点击 **Confirm** 保存更改。
5. 重启 **falco-agent**。

    a. 前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。

    b. 在右侧面板中，点击 <i class="material-symbols-outlined">more_vert</i>，然后选择 **Restart**。
6. 可选：验证规则是否处于活动状态。
    - 在同一页面，点击你的 pod 打开详情面板。在 **Containers** 下，点击 **falco** 旁边的 <i class="material-symbols-outlined">terminal</i>，然后运行：
      ```bash
      cat /etc/falco/rules.d/managed/custom_rules.yaml
      ```
    - 在 Falcosidekick UI 仪表板上，检查 **Rules** 下拉列表中是否有新规则。规则仅在触发后才会出现。

#### 禁用规则

1. 打开 Control Hub，然后前往 **Browse** > **System** > **falcoserver-shared** > **Configmaps** > **falco-disable-rules**。
2. 在右侧面板中，点击 **falco-disable-rules** 旁边的 <i class="material-symbols-outlined">edit_square</i> 打开 YAML 编辑器。
3. 在 `falco_disable_rules.yaml: |` 下方的行中添加你的规则。

    例如，要禁用 `Terminal shell in container` 规则：

    ```yaml
    data:
      falco_disable_rules.yaml: |
        - rule: Terminal shell in container
          override:
            enabled: replace
          enabled: false
    ```

4. 点击 **Confirm**。
5. 重启 **falco-agent**。

    a. 前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。

    b. 在右侧面板中，点击 <i class="material-symbols-outlined">more_vert</i>，然后选择 **Restart**。

6. 可选：验证规则是否已禁用。
    - 在同一页面，点击你的 pod 打开详情面板。在 **Containers** 下，点击 **falco** 旁边的 <i class="material-symbols-outlined">terminal</i>，然后运行：
      ```bash
      cat /etc/falco/rules.d/managed/falco_disable_rules.yaml
      ```
    - 在 Falcosidekick UI 仪表板上，检查 **Rules** 下拉列表。禁用的规则可能仍会出现在过去的事件中，但新事件将不再触发它，一旦历史记录过期，它就会消失。

### 配置输出通道

默认情况下，Falco 将告警发送到 Falcosidekick UI。你也可以将告警写入本地文件或转发到外部系统。

#### 发送告警到 Falcosidekick UI

默认情况下，`falco-agent` 通过 HTTP 将告警转发到 Falcosidekick，然后告警在 Falcosidekick UI 中显示。

1. 打开 Control Hub，然后前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。
2. 在右侧面板中，点击 **falco-agent** 旁边的 <i class="material-symbols-outlined">edit_square</i> 打开 YAML 编辑器。
3. 检查以下输出配置：

    示例：
    ```plain
    - '-o'
    - http_output.enabled=true
    - '-o'
    - http_output.url=http://falco-sidekick.falcoserver-shared:2801/
    ```

#### 将告警写入文件

要将告警写入本地日志文件：

1. 前往 **Settings** > **Applications** > **Falco** > **Manage environment variables**。
2. 将 `File_OUTPUT` 设为 `true`。

   ![启用文件输出](/images/manual/use-cases/falco-enable-file-output.png#bordered){width=90%}

3. 点击 **Confirm**，然后点击 **Apply**。
4. 可选：验证文件输出是否已启用。

    a. 打开 Control Hub，然后前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。

    b. 在右侧面板中，点击 **falco-agent** 旁边的 <i class="material-symbols-outlined">edit_square</i> 打开 YAML 编辑器。

    c. 检查配置是否包含 `file_output.enabled=true`。

5. 新告警将写入 Files 中 `/Data/falco/logs/` 的 `events.log`。

    :::info
    日志目录挂载在管理员环境中。只有管理员才能读取它。
    :::

#### 转发告警到外部系统

要将告警转发到 Slack、Elasticsearch、Webhook 或其他外部目的地，请直接配置 Falcosidekick。

有关支持的输出的完整列表，请参见 [Falcosidekick 文档](https://github.com/falcosecurity/falcosidekick)。

### 设置插件

Falco 插件添加额外的事件源。以下示例安装用于 Kubernetes 审计日志的 `k8saudit` 插件。

#### 安装插件

1. 打开 Control Hub，然后前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-plugin-installer**。
2. 点击你的 pod 打开详情面板。
3. 在 **Containers** 下，点击 **toolbox** 旁边的 <i class="material-symbols-outlined">terminal</i>。
4. 逐个运行以下命令以安装插件工件：

    ```bash
    falcoctl artifact install k8saudit
    falcoctl artifact install k8saudit-rules
    falcoctl artifact install json
    ```
5. 可选：验证插件是否已安装。

    a. 前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。

    b. 点击你的 pod 打开详情面板。

    c. 在 **Containers** 下，点击 **falco** 旁边的 <i class="material-symbols-outlined">article</i>。

    d. 检查 `Loading rules from:` 部分中是否出现 `k8s_audit_rules.yaml`。

#### 启用插件

<Tabs>
<template #在终端中编辑>

1. 打开 Control Hub，然后前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-plugin-installer**。
2. 点击你的 pod 打开详情面板。
3. 在 **Containers** 下，点击 **toolbox** 旁边的 <i class="material-symbols-outlined">terminal</i>。
4. 运行以下命令：
  ```bash
  cd /etc/falco/config.d/
  vi plugins.local.yaml
  ```
5. 使用以下示例配置更新文件：

    ```yaml
    plugins:
      - name: k8saudit
        library_path: /var/lib/falco/plugins/libk8saudit.so
        init_config: ""
        open_params: "http://:9765/k8s-audit"
      - name: json
        library_path: /var/lib/falco/plugins/libjson.so
        init_config: ""
    load_plugins: [k8saudit, json]
    ```
6. 保存文件并退出编辑器。

</template>

<template #在 Files 中编辑>

1. 在 Files 中，打开 `/Data/falco/plugins.local.yaml`。
2. 使用以下示例配置更新文件：

    ```yaml
    plugins:
      - name: k8saudit
        library_path: /var/lib/falco/plugins/libk8saudit.so
        init_config: ""
        open_params: "http://:9765/k8s-audit"
      - name: json
        library_path: /var/lib/falco/plugins/libjson.so
        init_config: ""
    load_plugins: [k8saudit, json]
    ```
3. 保存文件。

</template>
</Tabs>

**更新 `plugins.local.yaml` 后：**

1. 重启 **falco-agent**。

    a. 前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。

    b. 在右侧面板中，点击 <i class="material-symbols-outlined">more_vert</i>，然后选择 **Restart**。

2. 可选：验证插件是否已启用。

    a. 前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。

    b. 点击你的 pod 打开详情面板。

    c. 在 **Containers** 下，点击 **falco** 旁边的 <i class="material-symbols-outlined">article</i>。

    d. 检查日志中是否出现以下信息：
      - `Enabled event sources: k8s_audit`
      - `Opening 'k8s_audit' source with plugin 'k8saudit'`


## 故障排除

### 安装 k8saudit 规则后 falco-agent 无法启动

你可能会在日志中看到类似这样的错误：

```plain
LOAD_UNUSED_LIST (Unused list): List not referred to by any other rule/macro
Error: Plugin requirement not satisfied, must load one of: k8saudit (>= 0.7.0), k8saudit-aks (>= 0.1.0), k8saudit-eks (>= 0.4.0), k8saudit-gke (>= 0.1.0), k8saudit-ovh (>= 0.1.0)
```

**原因**

如果 `k8saudit` 规则已安装但插件未在 `plugins.local.yaml` 中成功启用，重启后 **falco-agent** 将无法启动。

**解决方案**

1. 打开 Control Hub，然后前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-plugin-installer**。
2. 点击你的 pod 打开详情面板。
3. 在 **Containers** 下，点击 **toolbox** 旁边的 <i class="material-symbols-outlined">terminal</i>。
4. 删除 `k8s_audit_rules.yaml` 文件：
    ```bash
    rm /etc/falco/rules.d/managed/k8s_audit_rules.yaml
    ```
5. 重启 **falco-agent**。

    a. 前往 **Browse** > **System** > **falcoserver-shared** > **Daemonsets** > **falco-agent**。

    b. 在右侧面板中，点击 <i class="material-symbols-outlined">more_vert</i>，然后选择 **Restart**。

如果你想继续使用 `k8saudit`，请从开头重新执行[安装和使用插件](#安装和使用插件)。

## 了解更多

- [Falco 官方文档](https://falco.org/docs/)：规则、条件和插件的完整参考。
- [Falcosidekick 文档](https://github.com/falcosecurity/falcosidekick)：支持的输出目的地和配置选项。
