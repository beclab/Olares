---
outline: [2, 3]
description: 了解如何在 Olares 中使用 Grafana 仪表板可视化 Prometheus 指标。
---
# 使用 Grafana 查看数据

在 Olares 中，你可以运行 Grafana，并连接到内置的 Prometheus 服务，从而可视化系统指标。本文将介绍如何安装流程，并说明如何连接数据源，以及导入仪表板。

## 安装 Grafana

在使用 Grafana 之前，需要先通过应用市场安装 Grafana。

1. 从启动台打开应用市场，搜索“Grafana”。
2. 点击**获取**，然后点击**安装**。
3. 在弹出的窗口，设置登录凭据：
   - `GF_USERNAME`：Grafana 登录用户名。
   - `GF_PASSWORD`：Grafana 登录密码。
    :::tip 记住你的登录凭据
    这是 Grafana 登录凭据，后续访问 Grafana 时会用到。
    :::
    ![设置登录凭据](/images/zh/manual/developer/mw-grafana-set-login.png#bordered){width=90% style="margin-left:0"}
4. 等待安装完成。

## 访问 Grafana

1. 从启动台打开 **Grafana**，点击 <i class="material-symbols-outlined">open_in_new</i> 在新标签页打开。
2. 在登录页面，输入安装时配置的 `GF_USERNAME` 和 `GF_PASSWORD`。

登录成功后，你会看到 Grafana 的首页。

## 添加 Prometheus 数据源

Olares 内置了 Prometheus 服务，可以收集系统指标。

要将 Grafana 连接到该服务，请按以下步骤操作：

1. 在 Grafana 左侧导航栏中，进入 **连接** > **数据源**。
2. 点击 **添加数据源**，然后选择 **Prometheus**。
3. 在 **Prometheus server URL** 字段输入：  
    ```text
    http://dashboard.<olaresid>.olares.com
    ```
    将 `<olaresid>` 替换为你的 Olares ID。
4. 点击页面底部的**保存并测试**。如果连接成功，你将看到如下提示。

    ![连接成功](/images/zh/manual/developer/mw-grafana-connect.png#bordered){width=90% style="margin-left:0"}

## 创建仪表板

如果你需要自定义指标和可视化数据，并且熟悉 PromQL，可以采用此方式。

1. 在左侧导航栏中点击**仪表板**。
2. 点击<b>+ 创建数据面板</b>，然后选择<b>+ 添加可视化</b>。
3. 选择 **prometheus** 作为数据源。
4. 根据需要配置面板、PromQL 查询和表达式。
5. 点击右上角的**保存仪表板**以便后续使用。

## 导入仪表板（推荐）

如果你不需要从零开始构建仪表板，可以直接导入现有的仪表板。

1. 访问 [Grafana 仪表板库](https://grafana.com/grafana/dashboards/)。
2. 下载所需仪表板的 `JSON` 文件。
3. 在 Grafana 中，点击右上角的 <i class="material-symbols-outlined">add_2</i>，然后选择**导入仪表板**。
4. 上传 `JSON` 文件，并选择 **prometheus** 作为数据源。
5. 点击 **Import** 完成导入。

导入的仪表盘包含预定义的面板和查询，导入后仍可根据需要进行自定义。

![导入仪表板](/images/zh/manual/developer/mw-grafana-dashboard.png#bordered){width=90%}