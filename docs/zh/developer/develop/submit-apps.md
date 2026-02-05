---
outline: [2, 3]
description: 学习如何向 Olares 应用市场提交应用。
---
# 提交 Olares 应用

本文介绍如何通过创建指向 `beclab/apps:main` 的拉取请求（Pull Request，PR），将新的 Olares 应用提交到默认索引。

Terminus-Gitbot 会基于 PR 标题、文件范围与所有权规则对 PR 进行校验，并自动关闭不符合要求的 PR。

## 前提条件

在提交之前，请确保你的应用已在 Olares 上完成充分测试。

推荐流程：

- 使用 [Studio](/zh/developer/develop/tutorial/develop.md) 开发容器，在真实的在线环境中测试和调试。
- [通过应用市场安装应用](/zh/developer/develop/tutorial/package-upload.md)，从用户视角测试安装与升级流程。

## 提交新应用

### 第一步：添加 OAC 到你的 Fork 仓库

1. Fork [官方仓库](https://github.com/beclab/apps) `beclab/apps`。
2. 在你的 Fork 仓库中新建目录，并将你的 [Olares 应用图表（OAC）](/zh/developer/develop/package/chart.md) 放在其中。
3. 在 OAC 根目录中创建一个 [`owners` 文件](/zh/developer/develop/distribute-index.md#开始之前)（无扩展名），并确保其中包含你的 GitHub 用户名。
    :::info 目录命名规范
    目录名即你的 OAC 目录名（图表文件夹名）。Terminus-Gitbot 会用它来解析 PR 标题并校验文件范围。目录名必须满足以下要求：
    - 只包含小写字母与数字
    - 不包含连字符（`-`）
    - 长度不超过 30 个字符
    :::

### 第二步：创建 Draft PR

创建一个指向 `beclab/apps:main` 分支的 Draft PR。

Terminus-Gitbot 会同时校验你的 PR 元数据（例如标题和文件作用范围）以及 Chart 内容（例如 OAC 根目录中是否包含必需文件）。请在继续之前确认你已完成第一步。

为通过 Terminus-Gitbot 的自动校验，你的 PR 必须严格遵循以下规则：


1. **标题格式**：标题必须必须清楚表明提交意图，并严格遵循以下格式：
```text
[PR 类型][图表文件夹名][版本] 标题内容
```
    
| 字段 | 说明 |
|--|--|
| PR 类型 | <ul><li>**NEW**：提交新应用。<br></li><li>**UPDATE**：更新已成功合并的应用。<br></li><li>**REMOVE**：移除已成功合并的应用。<br></li><li>**SUSPEND**：暂停已成功合并的应用在应用商店的分发。</li></ul> |
| 图表文件夹名 | 你的 OAC 目录。必须符合上述命名规范。 |
| 版本 | 应用的 Chart 版本，必须与以下字段一致：<br><ul><li>`Chart.yaml` 中的 `version`</li><br><li>`OlaresManifest.yaml` 的 `metadata` 下的 `version`</li></ul> |
| 标题内容 | 对 PR 的简要说明。 |

2. **文件范围**：PR 只能新增或修改 PR 标题中声明的图表文件夹下的内容。
3. **无重复 PR**：确保针对该图表文件夹没有其他 Open 或 Draft 状态的 PR。
4. **目录结构正确（针对新应用）：**: 
    - `beclab/apps:main`中不存在同名文件夹。
    - 你的图表文件夹中不包含[控制文件](/zh/developer/develop/manage-apps.md#控制文件)（`.suspend` 或 `.remove`）。

:::tip Draft PR 可随时修改
在 Draft 阶段，你可以继续提交 Commit 来调整文件。
:::

确认无误后，点击 **Ready for review**。

### 第三步：等待 Terminus-Gitbot 校验

提交后，Terminus-Gitbot 会自动校验 PR：
- 若 PR 通过所有检查，Terminus-Gitbot 会自动将 PR 合并到 `beclab/apps:main`。
- 稍等片刻，你的应用将出现在 Olares 应用市场。

## 跟踪 PR 状态

### 类型标签

当 PR 被标记为 `NEW`、`UPDATE`、`REMOVE` 或 `SUSPEND`时，表示标题中的 PR 类型已被识别。
:::warning 请勿更改类型
- 标记后，请勿更改 PR 类型。
- 如果类型不符合预期，请关闭该 PR 并新建一个 PR。
:::

### 状态标签

- `waiting to submit`：发现问题，需要在合并前继续修改。你可以继续提交 Commit，Terminus-Gitbot 会重新检查并更新状态。
- `waiting to merge`：通过所有检查，PR 已进入自动合并队列。请勿提交新的 Commit 或手动干预。
- `merged`：PR 已合并到 `beclab/apps:main`。
- `closed`：PR 无效或存在无法修复的问题。请勿重新打开，需修复问题后提交新的 PR。