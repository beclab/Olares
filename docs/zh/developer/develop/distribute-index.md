---
outline: [2, 3]
description: 了解 Olares 中应用分发的整体机制。
---
# 分发 Olares 应用

Olares 的应用分发基于开放标准和自动化校验机制。如果你的应用已经打包为Olares 应用图表（OAC），即可发布到 Olares 应用市场，让用户能够轻松获取和安装。

本指南将带你了解 Olares 应用分发的生命周期，理解应用市场的索引机制，如何发布、维护与推广应用。

## 开始之前

在分发应用之前，先来了解以下核心概念：

- **[Olares Application Chart（OAC）](/zh/developer/develop/package/chart.md)**  
  用于描述 Olares 应用的打包格式，包含元数据、所有权、版本信息以及安装配置。

- **应用索引**  
  向 Olares 应用市场提供应用元数据的服务。Olares 提供默认的公共索引服务，你也可以部署自己的索引服务。

- **Terminus-Gitbot**  
  自动验证系统，负责校验提交的应用并执行分发规则。

- **Owners 文件（`owners`）**  
  OAC 根目录下的一个文件，用于验证所有权和权限。该文件没有后缀名。
    ```text
    owners:
    - <your-github-username>
    - <collaborator1-username>
    - <collaborator2-username>
    ```
- **控制文件**
    位于 OAC 根目录中的特殊空文件，用于控制应用的分发状态：
    - `.suspend`：暂停应用分发
    - `.remove`：从应用市场移除应用

    详情请参阅[管理应用生命周期](/zh/developer/develop/manage-apps.md#控制文件)。

## 所有权与协作

若要以团队形式协作：
- （推荐）将所有维护者添加到 `owners` 文件中。列表中的每个所有者都可以独立 Fork 仓库并提交针对该应用的更改。
- 将团队成员添加为你 Fork 仓库的协作者，成员可以直接向你的 PR 分支提交代码。

## 应用分发流程

### 1. 准备应用包

分发应用之前，必须将其打包为 Olares 应用图表（OAC）。

在此阶段，开发者通常需要完成以下工作：
- 在 Olares 主机上开发并测试应用。
- 验证应用的安装和升级行为。
- 完善 Chart 的元数据和目录结构。

详情请参阅[Olares 应用图表（OAC）](/zh/developer/develop/package/chart.md)。

### 2. 提交应用到默认索引

Olares 应用市场从 Git 仓库索引应用。 要将应用发布到默认的公共索引，开发者需要向官方仓库提交一个拉取请求（Pull Request，PR）。

提交过程中：
- PR 标题用于声明操作类型。
- Terminus-Gitbot 验证文件范围、所有权和版本规则。
- 通过校验的 PR 会自动合并，无需人工审核。

详情请参阅[提交 Olares 应用](/zh/developer/develop/submit-apps.md)。

### 3. 自动校验与索引

当 PR 提交后，Terminus-Gitbot 会自动执行校验，确保提交内容符合分发规则。

如果所有检查通过，PR 将自动合并。 稍后，应用应用就会出现在 Olares 应用市场。

### 4. 管理应用生命周期

应用发布后，开发者仍可以通过提交 PR 的方式来管理其生命周期。

生命周期操作包括：
- 发布新版本。
- 暂停应用分发。
- 从应用市场永久移除应用。

这些操作通过 PR 类型以及 OAC 中的控制文件来实现。

详情请参阅[管理应用生命周期](/zh/developer/develop/manage-apps.md)。

### 5. 优化应用市场展示

应用发布后，你可以通过添加图标、宣传图和头图来优化应用其在 Olares 应用市场中的展示效果。

详情请参阅[推广你的应用](/zh/developer/develop/promote-apps.md)。


### 6. (可选) 发布付费应用
Olares 应用市场同样支持付费应用分发。付费应用需要额外的身份注册、定价配置以及许可证管理。

详情请参阅[发布付费应用](/zh/developer/develop/paid-apps.md)。