---
outline: [2, 3]
description: 排查暂停应用后内存未释放的问题。
---

# 暂停应用后内存不足或未释放

当你在设置中暂停应用后系统内存仍未释放，或因内存不足导致安装、恢复失败时，可参考本指南。

## 适用情况

- 应用市场提示内存不足，无法安装或恢复应用。
- 你在设置中手动暂停了某个应用，但内存占用未下降。

## 原因

在 Olares 中，部分应用（如 Ollama 及多数 LLM 应用）以[共享应用](../../developer/concepts/application.md#共享应用)形式部署，为集群内所有用户提供集中资源或服务。

这类应用的资源占用主要来自服务端。在 Olares OS 1.12.4 及更早版本中，从设置中暂停共享应用只会停止其面向用户的客户端，系统侧服务默认不会停止，因此其占用的内存不会被释放。

## 解决方案：通过 Market 释放内存

通过应用市场暂停应用及其服务端（而不仅从设置中暂停），然后视需要在仪表盘和控制面板中确认内存已释放。

:::info
本指南适用于你具备管理员权限且当前系统无子账户的情况。
:::

### 步骤 1：在仪表盘中查看当前内存

打开仪表盘，查看当前内存占用情况。

![仪表盘内存概览](/images/manual/help/ts-release-memory-dashboard-memory.png#bordered)

### 步骤 2：在应用市场中查看应用内存需求

在应用市场中打开要安装或恢复的应用，查看其内存需求。若已用内存与所需内存之和超过总内存约 90%，安装或恢复可能会失败。

![应用内存需求](/images/manual/help/ts-release-memory-market-requirements.png#bordered)

### 步骤 3：在仪表盘中查看不同应用的内存占用

1. 回到仪表盘，进入**应用**页面。
2. 在右上角使用排序选项，按内存占用排序。
3. 查看列表，确认哪些应用占用内存最多，并记下可关闭或暂时停止以释放内存的应用。

    ![仪表盘应用内存占用排序](/images/manual/help/ts-release-memory-dashboard-app-usage.png#bordered)

### 步骤 4：在应用市场中暂停应用

在应用市场中暂停应用可以同时停止共享服务端并释放其内存。

1. 打开应用市场，进入**我的 Olares**。
2. 找到要暂停的应用，点击应用操作按钮旁的 <i class="material-symbols-outlined">keyboard_arrow_down</i>，选择**暂停**。
    ![我的 Olares 选择要暂停的应用](/images/manual/help/ts-release-memory-market-my-olares.png#bordered)

3. 暂停应用时若出现弹窗，勾选**同时停止共享服务器（影响所有用户）**。

    ![停止 C/S 应用弹窗及关闭服务端选项](/images/manual/help/ts-release-memory-market-cs-dialog.png#bordered)

:::tip 卸载共享应用
若要卸载共享应用，勾选**同时卸载共享服务（影响所有用户）**，以便完全移除服务端。
:::

### 步骤 5：确认内存已释放

暂停应用后，等待约 5 分钟，再次打开 Dashboard 查看当前内存占用。

如果占用内存最多的应用已不再排在前面，但总内存占用仍然较高：

1. 打开控制面板，进入**命名空间**页面。
2. 点击**内存用量**列，按内存占用对命名空间排序。
3. 查找名称以 `server-shared` 结尾的命名空间，表示该共享应用的服务端仍在运行并占用内存。

![控制面板按内存排序的命名空间](/images/manual/help/ts-release-memory-controlhub-namespaces.png#bordered)

4. 对仍在占用内存的共享应用：

    a. 前往**应用市场** > **我的 Olares**，先恢复该应用再暂停。
    
    b. 在暂停弹窗中勾选**同时停止共享服务器（影响所有用户）**，使服务端完全停止并释放内存。
