---
outline: [2, 3]
description: 了解如何在 Olares 中管理应用入口，包括配置端点和设置访问策略。
---

# 管理应用入口

入口定义了用户访问 Olares 应用的方式。详情参见[入口](../../../developer/concepts/network.md#入口)概念。

入口管理包括两部分：

* **端点配置**：定义应用的网络地址和路由配置。
* **访问策略**：控制访问应用所需的身份验证方式。

## 进入入口管理

1. 前往**设置** > **应用**。
2. 点击目标应用。
3. 在**入口**下，点击目标入口。

   ![应用入口](/images/zh/manual/olares/app-entrance1.png#bordered){width=90%}

## 端点配置

**端点配置**面板用于自定义如何通过专属 URL 从外部访问应用。

![端点配置面板](/images/zh/manual/olares/app-entrance-endpoint-panel.png#bordered){width=70%}

可配置项包括：

- **端点**：用于访问应用的域名。点击 <i class="material-symbols-outlined">content_copy</i> 可复制该 URL。

- **默认路由 ID**：系统为应用路由分配的默认标识符。在此示例中，Jellyfin 的默认路由 ID 为 `7e89d2a1`。

- **设置自定义路由 ID**：点击 <i class="material-symbols-outlined">add</i> 可使用自定义路由 ID 替换默认路由 ID。例如，将其设置为 `jellyfin` 后，应用既可以通过 `https://7e89d2a1.alexmiles.olares.com` 访问，也可以通过 `https://jellyfin.alexmiles.olares.com` 访问。详细步骤请参阅[自定义路由 ID](custom-app-domain.md#自定义路由-id)。

- **设置自定义域名**：点击 <i class="material-symbols-outlined">add</i> 可为该应用添加自定义域名，例如 `app.yourdomain.com`。在域名生效前，你还需要完成相应的 DNS 配置。详细步骤请参阅[自定义域名](custom-app-domain.md#自定义域名)。

## 访问策略

访问策略用于控制谁可以访问应用，以及访问时需要使用哪种认证方式。

![访问策略面板](/images/zh/manual/olares/app-entrance-access-policy-panel.png#bordered){width=70%}

可配置项包括：

* **认证级别**：设置应用的身份验证要求：

   * **公开**：任何人都可以访问，无需登录。
   * **私有**：需要用户登录才可访问。
   * **内部**：如果通过 VPN 访问应用，则无需登录。

* **认证模式**：指定验证用户身份的方法：

   * **跟随系统**：继承**我的 Olares**页面中设置的系统身份验证规则。
   * **单因素**：仅需要 Olares 登录密码。
   * **双因素**：需要 Olares 密码以及二次验证码。
   * **无**：无需任何身份验证即可访问。

* **管理子策略**：使用**正则表达式**为应用中的特定路径设置更细粒度的访问规则。

  1. 点击 <i class="material-symbols-outlined">chevron_forward</i> 打开**管理子策略**页面。
  2. 点击**添加子策略**，然后在**受影响 URL** 中输入目标路径，并设置**认证模式**。
  3. 点击**提交**。