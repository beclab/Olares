---
outline: [2, 3]
description: 了解如何在 Olares 中使用 NATS CLI 订阅和发布消息，并理解 NATS 主题的命名规则与权限模型。
---
# 使用 NATS 订阅与发布消息

本文介绍如何使用 `nats-box` CLI 工具在 Olares 集群内测试 NATS 的消息订阅与发布，并概述 NATS 的主题命名规则与权限模型。

## 获取连接信息

在建立连接之前，需要从控制面板获取 NATS 的连接信息。

1. 从启动台打开控制面板。
2. 在左侧导航栏中找到中间件，并选择 **Nats**。
3. 记录主题面板中的以下信息：
    - **主题**：目标消息主题。
    - **用户**：连接用户名。
    - **密码**：连接密码。

    ![Nats 详情](/public/images/zh/manual/developer/mw-nats-details.png#bordered){width=60% style="margin-left:0"}

## 通过 CLI 访问

`nats-box` 提供了一种便捷方式，可在集群内测试 NATS 的订阅与发布。

### 部署 `nats-box`

1. 下载示例文件 [`nats-box.yaml`](http://cdn.olares.com/common/nats-box.yaml)，并将其上传到 Olares 机器。
2. 进入 YAML 文件所在目录，部署 `nats-box`：
    ```bash
    kubectl apply -f nats-box.yaml
    ```
3. 获取 `nats-box` 的容器名称：
    ```bash
    kubectl get pods -n os-platform | grep nats-box
    ```
4. 进入 `nats-box` 容器：
    ```bash
    kubectl exec -it -n os-platform <nats-box-pod> -- sh
    ```

### 订阅消息

使用控制面板中获取的信息，包括主题、用户名和密码：

```bash
nats sub <subject-from-controlhub> --user=<user-from-controlhub> --password=<password-from-controlhub> --all
```

### 发布消息

向指定的主题发布一条消息：

```bash
nats pub <subject-from-controlhub> '{"hello":"world"}' --user=<user-from-controlhub> --password=<password-from-controlhub>
```

## 主题命名与权限参考

本节为你介绍 Olares 中使用的主题命名规范与权限模型。

### 主题结构

NATS 的主题采用三级结构，并使用英文句点（`.`）分隔：`<prefix>.<event>.<olaresId>`。

| 层级 | 名称 | 说明 |
|--|--|--|
| 第一级 |`<prefix>` | 来源标识。<br>- **系统服务**：固定为 `os`。<br> - **第三方应用**：使用对应的 `appId`。 |
| 第二级 | `<event>` | 标识事件或领域。<br>示例：`users`、`groups`、`files`、`notification`。 |
| 第三级 |`<olaresId>` | 表示用户空间的 Olares ID。| 

### 权限模型

主题的读写权限会根据应用类型而有所不同。

| 应用类型 | 权限范围 | 说明 |
|--|--|--|
| 用户空间应用| 只读 | 只能订阅包含自身 `<olaresId>` 的三级主题。 |
| 系统/集群应用| 系统级访问 | **订阅**:可订阅系统级 Subject（例如 `os.users`、`os.groups`）。<br>**写入**：可在自身空间内向二级主题发布消息。<br>**全局读取**：订阅所有二级主题的读权限需单独申请。 |