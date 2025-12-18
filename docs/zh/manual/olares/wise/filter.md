---
description: 了解如何使用 Wise 强大的过滤系统来整理你的内容库。创建基于标签的视图、添加自定义过滤视图，并使用高级查询参数来高效管理内容。
---
# 使用过滤视图管理内容

随着 Wise 内容库日益庞大，仅靠浏览文件夹来查找特定内容可能会变得困难。你可以使用过滤视图，根据自定义规则（例如特定标签、内容类型或日期）将相关内容分组，归类到动态的“智能文件夹”中。

与普通文件夹不同，这些视图会自动更新。一旦创建完成，只要新内容符合过滤条件，就会自动出现在对应视图中，无需你手动操作。

:::info
只有来自订阅源的内容，或状态为<i class="material-symbols-outlined">inbox</i> **收件箱** 或 <i class="material-symbols-outlined">schedule</i> **稍后阅读**的内容，才能用于过滤视图。
:::

本文将介绍如何：

- 基于已有标签和订阅源构建过滤视图。
- 使用查询语法创建自定义视图。
- 将常用视图固定到侧边栏。
- 了解更多高级过滤选项。

## 基于标签或订阅源快速创建视图

创建视图最快的方式，就是基于已有标签或订阅源来生成。

1. 在 Wise 中，点击左下角的 <i class="material-symbols-outlined">settings</i>。
2. 选择**标签**或 **RSS 订阅源**。
3. 找到你想使用的标签或订阅源，点击**管理视图...**。
4. 选择操作：
   - 添加到已有视图：将此标签/订阅源加入已有视图。
   - 创建新视图：以此标签/订阅源为基础，新建一个独立视图，将所有相关内容集中展示。
      a. 输入视图名称。  
      b. 点击**确认** 保存视图。

![添加过滤视图](/images/zh/manual/olares/wise-add-filtered-view.png#bordered) 

## 创建自定义过滤视图

如果你追求更加灵活的操作，可以使用查询语法创建视图，用更精确的条件来分组内容。

创建自定义筛选视图的方法如下：

1. 在 Wise 中，点击左下角的 <i class="material-symbols-outlined">settings</i>。
2. 选择**过滤视图**。
3. 在右上角点击**添加视图**。
4. 为视图命名（例如：`AI trends`）。
5. 在**查询**输入框中填写过滤表达式。例如：
   - `tag:AI`：匹配所有带有`AI`标签的内容。
![创建新视图](/images/zh/manual/olares/wise-create-views.png#bordered){width=60%}
   
6. 点击**确认**保存视图。

新视图会立即将所有符合条件的内容包含进来。

## 将视图固定到侧边栏

你可以将常用的筛选视图固定到侧边栏，以便一键访问。

1. 在 Wise 中，点击左下角的 <i class="material-symbols-outlined">settings</i>。
2. 选择**过滤视图**。
3. 通过以下任一方式固定目标视图：
   - 在列表中找到目标视图，点击**操作**列中的 <i class="material-symbols-outlined">keep</i>。
   - 点击目标视图并打开，在视图页头部点击 <i class="material-symbols-outlined">keyboard_arrow_down</i>，选择**固定至边栏**。
  ![固定至边栏](/images/zh/manual/olares/wise-pin-sidebar.png#bordered){width=50%}

固定后的视图会显示在侧边栏，与其他导航项并列展示。

## 常用筛选条件与语法

若想了解所有可用筛选条件及其用法，请参考：

- [过滤视图示例](filter-examples)
- [过滤语法参考](filter-syntax-guide.md)