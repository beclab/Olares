---
outline: [2, 3]
---

# 提交应用

## 详细步骤

**基本流程**

1. 在 Olares 上测试你的应用，根据[指南](../package/chart.md)创建 Olares 应用 Chart。
2. Fork 官方 [Olares 应用市场仓库](https://github.com/beclab/apps)。添加应用 Chart。创建 PR 到 `beclab/apps:main`。
3. 等待 **GitBot** 检查你的 PR。如果需要，修改 PR 直至通过。
4. PR 合并后，你的应用程序就可以启动了。


### 1. 创建并测试 Olares 应用

在提交应用程序之前，请确保它已在你的 Olares 上经过完整测试。
- 使用 DevBox 的开发容器在真实的在线环境中测试和调试你的应用。 [了解有关 DevBox 的更多信息](../tutorial/studio)。
- 使用 Market 应用程序中的[自定义安装](../../../manual/olares/market/market#install-custom-applications) 进行用户测试。

### 2. 提交应用
应用提交需要通过 **Pull Request** 完成。具体步骤如下：
- 从官方 [Olares 应用市场仓库](https://github.com/beclab/apps)创建 fork，并在你的 fork 仓库中添加应用 Chart。
- 创建一个指向 `beclab/apps:main` 分支的 Draft PR。
- 请按照模板要求编辑 PR 标题和内容。
  - **PR 标题**必须符合以下格式：[PR 类型][文件夹名][版本]标题内容
  - `PR 类型`包括：
    - NEW: 提交新应用
    - UPDATE: 更新已成功合并的应用
    - REMOVE: 移除已成功合并的应用
    - SUSPEND: 暂停已成功合并的应用在应用商店的分发
  - `文件夹名`是你的 Olares 应用 Chart 名称。必须符合 [chart 规范](../package/chart.md)中的命名要求。
  - `版本`指你的应用 chart 的` Chart 版本`，需要与 `Chart.yaml` 中的 `version` 字段和 `OlaresManifest.yaml` 元数据部分保持一致。
- 为防止你的 PR 被错误解析或关闭，请遵守以下规则：
  - PR 标题只能包含一个 `PR 类型`、`文件夹名`和`版本`。
  - 你的 `PR 类型`必须是预定义类型之一。
  - PR 只能添加或修改PR标题中声明的`文件夹名`下的内容。
  - 同一`文件夹名`同时只能存在一个 Open PR 或 Draft PR。
  - 你必须是要修改的文件夹的所有者之一。所有者列在chart中的`owners`文件中。如果你正在提交新应用，你的 GitHub 用户名应该包含在 `owners` 文件中。

- 在 Draft PR阶段，你可以持续调整 PR 内容并添加新的提交。一切就绪后，点击 **Ready for review** 按钮提交 PR 并调用 **GitBot** 进行检查。

:::info 注意
PR 的标题和内容对于 **GitBot** 来说至关重要。请按照模板规范填写。**GitBot** 可能会自动关闭任何无效的PR。
:::

### 3. 跟踪 PR 状态
- 当你的 PR 被标记为 `PR 类型`时，表示你的 PR 标题有效。请**不要在之后修改 `PR 类型`。**如果它不符合你的意图，只需关闭它并创建一个新的。

- 你可以通过状态标签跟踪PR的进度：
  - `waiting to submit`: 你的 PR 存在问题，需要在合并前进行进一步修改。
  - `closed`: 你的 PR 无效或包含无法纠正的错误。
  - `waiting to merge`: 一切进展顺利。你的PR已通过检查，现在正等待 **GitBot** 自动合并。
  - `merged`: 你的PR已被自动合并到`beclab/apps:main`中。

- 如果 **GitBot** 自动关闭了你的 PR，请**不要重新打开**它。这表示 PR 存在不可修复的问题，**GitBot** 不得不终止检查流程。你可以在进行必要修改后提交新的 PR。

- 当 PR 处于 `waiting to submit` 状态时，你可以继续提交 Commits 来修改你的 PR。提交 Commit 后，GitBot将重新检查你提交的应用 chart 文件并更新 PR 状态。

- 一旦你的 PR 通过所有检查，它将自动合并到 `beclab/apps:main` 中。应用将在一段时间后在**Olares Market**上列出。

- 如果你在提交过程中遇到任何问题，随时可以联系 Olares 团队或寻求社区帮助。

## 管理你的应用

你可以通过创建 Pull Request 到 `beclab/apps:main` 继续管理和维护你的应用。你可以升级应用、修改其可用性，或从**Olares Market**完全移除它。

管理应用的流程与提交类似。你创建特定类型的 Pull Request，GitBot 会处理剩余工作。Olares 使用应用 chart 根目录中的**特殊控制文件**来管理应用状态。这些**特殊控制文件**是带有特定后缀的空文件，如 `.suspend` 和 `.remove`。

:::info 注意
初次提交时不应包含 ".suspend" 或 ".remove" 文件。
:::

### 更新
当你需要更新已发布的应用时，你需要创建一个 `UPDATE` PR。

**请注意：**
- 每当你对应用chart进行更改时，如升级程序、更新元数据或更改所有者列表，请务必升级你的chart版本。
- 更新后的应用chart版本必须***大于***仓库中的当前版本。
- 更新的应用chart根目录中不包含 `.suspend` 或 `.remove` 文件。
- **Olares 应用市场**不提供版本回滚。如果你的应用存在任何问题，你需要提交新版本来修复它。
- 为避免潜在冲突，我们建议同步你的fork并将PR的提交变基到最新的main分支。

### 暂停
如果出于任何原因你想暂时禁用你的应用在**Olares 应用市场**的下载和安装，请提交一个`SUSPEND` PR。

**请注意：**
- 提交的应用chart版本必须与仓库中的当前版本***匹配***。
- 你提交的根目录应包含 `.suspend` 文件，且不应包含 `.remove` 文件。
- 一旦暂停PR通过检查并合并，应用商店将停止列出你的应用。
- 已经下载并安装应用的用户在暂停后可以继续使用它。

### 移除
如果出于任何原因你想从**Olares Market**移除你的应用，请提交一个 `REMOVE` PR。

**请注意：**
- 完全清空当前应用目录中的文件，并在根目录中添加一个 `.remove` 文件。
- 一旦移除PR通过检查并合并，应用商店将移除你的应用。
- 你将无法在未来重用相同的目录或应用chart名称。
- 已经下载并安装应用的用户在移除后可以继续使用它。

## 推广你的应用

通过使用组织良好的应用描述、截图和宣传图片来突出显示应用的特点和功能，可以帮助吸引Market中的新用户。截图和预览可以直观地展示用户体验，帮助你的应用脱颖而出。

要在应用详情页添加宣传图片，请在 `OlaresManifest.yaml` 文件的 `spec` 部分的 `promoteImage` 字段中包含这些资源的链接。

:::info **Olares 应用市场的资源规范**

- 应用的**图标**必须为 PNG 或 WEBP 格式，大小不超过 512 KB，尺寸为 256x256 像素。

- 强烈建议至少上传 2 张**截图**用于推广。**截图**必须为 JPEG、PNG 或 WEBP 格式，每张大小不超过 8MB，尺寸为1440x900像素。你最多可以上传 8 张**截图**。

- 如果你希望你的应用在市场中被推荐，则需要一张**特色图片**。在 `OlaresManifest.yaml` 文件的 `spec` 部分的 `featuredImage` 字段中添加此图片的链接。图片必须为 JPEG、PNG 或 WEBP 格式，大小不超过 8MB，尺寸为 1440x900 像素。
  :::

## 邀请他人协作

有两种方式可以邀请他人一起开发Olares应用：
1. 将其他开发者的 GitHub 用户名添加到 `owners` 文件中。列在`owners`中的每个开发者都可以独立 fork 仓库并提交他们的更改。
2. 将其他人添加为你 fork 仓库的协作者。在这种情况下，你作为代表创建 Pull Request，所有其他人可以共同提交到计划合并的分支。
