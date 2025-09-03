---
outline: [2, 3]
description: 在 Olares 中安全共享 Vault 项目。了解团队角色权限设置，管理共享 Vault 访问权限，实现团队成员间的安全协作。
---

# 共享 Vault 项目

共享 Vault 是在同一 Olares 集群内组织、管理和安全共享数据的有效方法。无论是管理家庭账户还是企业级数据，团队 Vault 都能让你兼顾安全性与协作。

## 了解团队角色

所有 Olares 集群的管理员和成员都会自动加入 Vault 的 **My Team** 团队。Olares 超级管理员会默认成为团队所有者，其他用户则为团队成员。

所有者可导航至**团队** > **成员**页面查看所有成员：

![Vault 团队](/images/zh/manual/olares/vault-team.png#bordered)

| 角色     | 所有者                                 | 成员       |
|----------|-------------------------------------|------------|
| 权限     | - 创建、删除、编辑共享 Vault<br/>- 分配成员的读/写权限 | 根据所分配的权限访问或编辑共享 Vault 项目 |

## 设置团队访问

要设置团队访问，你需要首先创建共享 Vault，在其中添加项目，然后将 Vault 共享给指定成员。

### 创建共享 Vault

共享 Vault 仅可由所有者创建。

![创建 team vault](/images/zh/manual/olares/create-team-vault.png#bordered)

1.  导航到**团队** > **Vaults** 页面。
2.  点击右上角的 <i class="material-symbols-outlined">add</i> 按钮，并输入 Vault 名称。
3.  点击**保存**。

新创建的团队 Vault 初始为空，你可以继续为此 Vault [添加新项目](vault-items.md#添加)，或直接从[外部导入](vault-items.md#导入)。

### 共享 Vault 给成员

默认情况下，所有者会自动获得新创建共享 Vault 的访问和管理权限。团队成员则需要所有者授予访问权限：

1.  导航到**团队** > **Vaults** 页面。
2.  在团队 Vault 列表里点击要共享的 Vault。
3.  在右侧详情页右上角，点击 **+** 按钮以添加你想共享的成员。
4.  在权限设置下拉框里，为该成员设置**读/写**（Editable）或**只读**（ReadOnly）权限。
5.  点击**保存**完成设置。

![Vault 共享](/images/zh/manual/olares/vault-share.png#bordered)

该成员现在可以在 **Team Vault** 类别下查看共享的 Vault。如需移除特定成员的共享权限，所有者可以点击权限下拉框旁的 <i class="material-symbols-outlined">delete</i> 按钮。

### 删除共享 Vault

所有者可以根据需要删除共享 Vault：

:::warning 警告
删除共享 Vault 将永久移除所有相关数据。请务必在确认删除前仔细检查。
:::

1.  导航到**团队** > **Vault** 页面。
2.  在团队 Vault 列表里点击要删除的项目。
3.  在右侧详情页右上角的 <i class="material-symbols-outlined">more_horiz</i> 菜单中，点击**删除**。
4.  在弹出对话框中输入 `DELETE` 以确认删除。