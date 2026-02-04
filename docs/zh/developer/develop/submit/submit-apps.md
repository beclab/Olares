---
outline: [2, 3]
description: 学习如何向 Olares 应用市场提交应用。
---
# 提交 Olares 应用

本文介绍如何通过创建指向 `beclab/apps:main` 的拉取请求（Pull Request，PR），将新的 Olares 应用提交到默认索引。

GitBot 会基于 PR 标题、文件范围与所有权规则对 PR 进行校验，并可能自动关闭无效 PR。

## 前提条件

在提交之前，请确保你的应用已在 Olares 上完成充分测试。

推荐流程：

- 使用 [Studio](/zh/developer/develop/tutorial/develop.md) 开发容器，在真实的在线环境中测试和调试。
- [通过应用市场安装应用](/zh/developer/develop/tutorial/package-upload.md)，从用户视角测试安装与升级流程。

## 提交新应用

### 第一步：添加 OAC 到你的 Fork 仓库

1. Fork [官方仓库](https://github.com/beclab/apps) `beclab/apps`。
2. 在你的 Fork 仓库中新建目录，并将你的 [Olares 应用图表（OAC）](/zh/developer/develop/package/chart.md) 放在其中。
3. 确保 OAC 根目录包含 `owners` 文件，其中写入你的 GitHub 用户名。

    :::info 目录命名规范
    目录名即你的 OAC 目录名（图表文件夹名）。GitBot 会用它来解析 PR 标题并校验文件范围。目录名必须：
    - 只包含小写字母与数字
    - 不包含连字符（`-`）
    - 长度不超过 30 个字符
    :::

### 第二步：创建 Draft PR

创建一个指向 `beclab/apps:main` 分支的 Draft PR，并按以下格式设置 PR 标题：

```text
[PR 类型][图表文件夹名][版本] 标题内容
```
    
| 字段 | 说明 |
|--|--|
| PR 类型 | **NEW**：提交新应用。 <br>**UPDATE**：更新已成功合并的应用。<br>**REMOVE**：移除已成功合并的应用。<br>**SUSPEND**：暂停已成功合并的应用在应用商店的分发。 |
| 图表文件夹名 | 你的 OAC 目录。必须符合上述命名规范。 |
| 版本 | 应用的 Chart 版本，必须与以下字段一致：<br>- `Chart.yaml` 中的 `version`<br>- `OlaresManifest.yaml` 的 `metadata` 下的 `version` |
| 标题内容 | 对 PR 的简要说明。 |

### 第三步：校验 PR

提交 PR 进行审核前，请检查以下内容：

1. **标题有效**：标题仅包含一个 PR 类型、一个图表文件夹名和一个版本号。
2. **修改范围合规**：PR 只添加或修改标题中声明的图表文件夹下的内容。
3. **无重复 PR**：同一个图表文件夹下没有其他处于 Open 或 Draft 状态的 PR。
4. **你是所有者**：你的 GitHub 用户名已写入 `owners` `文件中。
5. **提交新应用时**：
    - `beclab/apps:main` 中不存在同名文件夹。
    - 图表文件夹不包含 `.suspend` 或 `.remove` 文件。

在 Draft 阶段，你可以继续提交 Commit 来调整文件。

准备就绪后，点击 **Ready for review**。

### 第四步：等待 GitBot

提交后，GitBot 会自动校验 PR：
- 若 PR 通过所有检查，GitBot 会自动将 PR 合并到 `beclab/apps:main`。
- 稍等片刻，你的应用将出现在 Olares 应用市场。

## 跟踪 PR 状态

### 类型标签

当 PR 被标记为 `NEW`、`UPDATE`、`REMOVE` 或 `SUSPEND`时，表示标题中的 PR 类型已被识别。
- 标记后，请勿更改 PR 类型。
- 如果类型不符合预期，请关闭该 PR 并新建一个 PR。

### 状态标签

- `waiting to submit`：发现问题，需要在合并前继续修改。你可以继续提交 Commit，GitBot 会重新检查并更新状态。
- `waiting to merge`：通过所有检查，PR 已进入自动合并队列。请勿提交新的 Commit 或手动干预。
- `merged`：PR 已合并到 `beclab/apps:main`。
- `closed`：PR 无效或存在无法修复的问题。请勿重新打开，需修复问题后提交新的 PR。

## 邀请协作者

你可以通过两种方式协作开发应用：
-  **在 `owners` 文件中添加（推荐）**：将其他开发者的 GitHub 用户名添加到 OAC 根目录的 `owners` 文件中。列表中的每个所有者都可以独立 Fork 仓库并提交改动。
- **作为 Fork 仓库协作者**：将其他人添加为你 Fork 仓库的协作者。这种情况下，必须由你创建 PR，但协作者可以向你的 PR 分支提交代码。