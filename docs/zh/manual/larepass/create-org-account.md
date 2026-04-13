---
outline: [2, 3]
description: 使用自定义域名创建 Olares ID。首次设置域名创建组织，或以成员身份使用管理员提供的凭据加入组织。
---

# 使用自定义域名创建 Olares ID

:::tip
如需从域名设置到 Olares 安装的完整流程，请参阅[为 Olares 设置自定义域名](../best-practices/set-custom-domain.md)。
:::

在 Olares 中，自定义域名通过组织进行管理。域名所有者先创建组织，然后添加成员，成员即可在该域名下创建自己的 Olares ID。

## 创建 DID

<!--@include: ../../reusables/custom-domain.md{21,31}-->

创建 DID 后：
- 如果你是域名所有者，需要设置组织，请继续阅读[创建新组织](#创建新组织)。
- 如果管理员已创建组织并添加了你，请跳到[加入已有组织](#加入已有组织)。

## 创建新组织

开始之前，请确保自定义域名已在 [Olares Space 中完成设置](../space/host-domain.md)。

作为域名所有者，创建组织并获取自定义域名下的 Olares ID。

<!--@include: ../../reusables/custom-domain.md{76,103}-->

域名设置完成后，你可以在 Olares Space 中[添加成员](../space/manage-domain.md)。

## 加入已有组织

如果域名管理员已创建组织并添加了你，使用管理员提供的用户名和密码加入。

<!--@include: ../../reusables/custom-domain.md{130,141}-->

Olares ID 创建完成，可以继续[安装并激活 Olares](../get-started/install-olares.md)。
