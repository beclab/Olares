---
outline: [2, 3] 
description: 了解如何在 Olares 中管理应用环境变量。
---

# 管理应用环境变量

应用环境变量用于定义应用在容器中的运行方式，包括应用配置、服务连接以及其他运行时设置。

## 查看环境变量

1. 前往**设置** > **应用**。
2. 在列表中选择目标应用。
3. 在**环境变量**部分，点击**管理环境变量**。

![应用环境变量](/images/zh/manual/olares/manage-app-env1.png#bordered){width=90%}

## 理解环境变量类型

**管理环境变量**页面显示了当前应用正在使用的环境变量。

![应用环境变量列表](/images/zh/manual/olares/manage-app-env-var-list.png#bordered){width=90%}

根据应用的配置方式，你可能会看到以下两类变量：

- **应用专属变量**：应用本身定义的变量。这些变量在页面中正常显示，可直接在当前页面编辑。
- **引用的系统环境变量**：应用引用的系统全局变量。只有当应用使用了这些变量时，它们才会显示在此页面上。这类变量在当前页面呈置灰状态，且无法直接编辑。若要修改此类变量，可参考[设置系统级环境变量](/zh/manual/olares/settings/developer.md#设置系统级环境变量)。

## 编辑应用专属变量

如需修改应用专属变量：

1. 找到要编辑的变量。
2. 点击右侧的 <i class="material-symbols-outlined">edit_square</i>。
3. 在弹出的对话框中更新变量值，然后点击**确认**。
4. 点击**应用**。

![编辑应用环境变量](/images/zh/manual/olares/manage-app-env-edit-var.png#bordered){width=90%}

应用会自动重启，新配置随后生效。

## 了解更多

- [系统环境变量](developer.md#设置系统级环境变量)