---
outline: [2, 3]
description: 当设备上的应用市场缺少官网所列应用时的排查方法。
---

# 应用市场中缺少应用

当官方应用市场网站上的应用在当前设备的应用市场中不显示，且设备已持续运行超过 2 小时仍未见恢复时，可参考本指南。

## 适用情况

- 官方 [Olares 应用市场](https://market.olares.com)上列出的部分应用在当前设备的应用市场中不显示。
- 设备已持续运行超过 2 小时，这些应用仍然缺失。

## 原因

系统在初次配置时可能未完成完整应用列表的下载。重启图表仓库和应用市场相关部署可以重新拉取应用目录。

## 解决方案

在控制面板中重启图表仓库和应用市场服务，等待服务就绪后再刷新应用市场。

1. 打开控制面板，进入 **浏览** > **系统** > **os-framework**。
2. 在**部署**列表中找到 **chartrepo-deployment**，点击**重启**，等待其状态变为 **Running**。

    ![重启 chartrepo](/images/zh/manual/help/ts-missing-apps-restart-chart.png#bordered)

3. 在同一列表中找到 **market-deployment**，点击**重启**，同样等待其状态变为 **Running**。

    ![重启应用市场](/images/zh/manual/help/ts-missing-apps-restart-market.png#bordered)

4. 等待 2 到 3 分钟，让服务完成初始化。
5. 刷新应用市场。
