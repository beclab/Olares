---
search: false
---
<!--
  自定义域名设置的复用内容块。

  Block 1 – 创建 DID:              lines 21-31
  Block 2 – 添加域名 (TXT + NS):   lines 34-73
  Block 3 – 创建新组织:             lines 76-106
  Block 4 – 添加新用户:             lines 109-127
  Block 5 – 加入已有组织:           lines 130-141

  引用方:
  - manual/best-practices/set-custom-domain.md (Blocks 1-5)
  - manual/larepass/create-org-account.md (Blocks 1, 3, 5)
  - manual/space/host-domain.md (Block 2)
  - manual/space/manage-domain.md (Block 4)
-->

<!-- Block 1: Create a DID -->
DID（去中心化标识符）是获取最终 Olares ID 之前的临时账户状态。只有在 DID 阶段，才能将自定义域名与账户关联。创建步骤如下：

1. 在 LarePass 应用中，打开账户创建页面。

2. 点击**创建账户**。

   ![LarePass 账户创建页面](/images/manual/tutorials/create-a-did1.png)

   此操作将创建一个处于 DID 阶段的 Olares 账户。在**切换账户**页面，该账户显示为"未关联 Olares ID"，标识符类似 `did:key:xxxx`。

   ![DID 阶段](/images/manual/tutorials/did-stage1.png)

<!-- Block 2: Add domain (TXT + NS) -->
以下步骤以 `space.n1.monster` 为例。

1. 在 Olares Space 中，进入**域名管理**页面，点击**域名设置**。

   ![域名管理页面中的域名设置按钮](/images/manual/tutorials/custom-domain-set-up-domain-name.png#bordered)

2. 在弹出的对话框中输入一个有效的子域名，然后点击**确认**。

   :::warning 不要使用主域名
   使用 `yourdomain.com` 这样的主域名会将所有 DNS 管理迁移到 Olares Space，且不会自动迁移已有的 DNS 记录。
   请使用 `app.yourdomain.com` 这样的子域名。
   :::

3. 添加并验证 TXT 记录，以证明域名所有权。

   a. 在**操作**列中点击**引导**。
   ![验证 TXT](/images/manual/tutorials/custom-domain-verify-txt.png#bordered)
   b. 在你的 DNS 服务商设置中，按照对话框中提供的名称和值添加一条 TXT 记录。

      ![在 DNS 服务商添加 TXT 记录](/images/manual/tutorials/custom-domain-add-txt-record.png#bordered)

   验证通过后，状态会更新为**等待添加域名 NS 记录**。
   ![域名状态更新为等待添加 NS 记录](/images/manual/tutorials/custom-domain-add-ns.png#bordered)

4. 验证域名的 NS（Name Server）记录。此操作将域名的 DNS 解析委托给 Olares 的 Cloudflare。

   a. 在**操作**列中点击**引导**。

   b. 在你的 DNS 服务商设置中，按照对话框中提供的值为子域名添加两条 NS 记录。

      ![在 DNS 服务商添加 NS 记录](/images/manual/tutorials/custom-domain-add-ns-record.png#bordered)

   验证通过后，域名状态将更新为**等待为域名申请可验证凭证**。
   ![域名状态更新为等待申请可验证凭证](/images/manual/tutorials/custom-domain-wait-vc.png#bordered)

   :::warning
   验证成功后，请勿修改 NS 记录。否则会导致自定义域名解析失败，无法正常访问。
   :::

TXT 和 NS 记录验证通过后，可以在 LarePass 中创建组织。

<!-- Block 3: Create a new organization -->
1. 在手机上打开 LarePass，在账户创建页面点击右上角的 <i class="material-symbols-outlined">display_settings</i>，进入**高级账户创建**页面。
   ![LarePass 中的高级账户创建选项](/images/manual/tutorials/custom-domain-advanced.png)

2. 前往**组织 Olares ID** > **创建新组织**。已验证的域名会自动显示在列表中。

   ![LarePass 中的组织 Olares ID 选项](/images/manual/tutorials/custom-domain-org-olares-id.png)

   ![创建新组织](/images/manual/tutorials/custom-domain-create-org.png)

3. 点击域名。
   ![选择组织对应的域名](/images/manual/tutorials/custom-domain-select-org.png)

   :::warning
   为域名创建组织后，该域名将无法从 Olares Space 中移除。
   :::

4. 输入 Olares ID 的用户名部分。例如输入 `alex`，则 Olares ID 为 `alex@space.n1.monster`。

   :::info
   用户名部分应为 1-63 个字符，仅支持小写字母和数字。
   :::
   ![以管理员身份创建 Olares ID](/images/manual/tutorials/custom-domain-create-olares-id-as-admin.png)

5. 点击**确认**。

   ![已创建具有管理员权限的 Olares ID](/images/manual/tutorials/custom-domain-admin-id-created.png)

   Olares ID 已创建，并拥有管理该域名下用户的管理员权限。

6. 点击**下一步**，进入 Olares 激活页面。
   ![在 LarePass 中发现 Olares](/images/manual/tutorials/custom-domain-discover-olares.png#bordered)

<!-- Block 4: Add a new user -->
1. 在 Olares Space 中刷新**域名管理**页面，域名状态已更新为**已分配**。
   ![Olares Space 中的域名成员列表](/images/manual/tutorials/custom-domain-view-user.png#bordered)

2. 在**操作**列中点击**查看**。

3. 点击**添加新用户**，输入成员的用户名（自定义域名前面的部分）。例如 `alice`。

   ![添加新用户对话框](/images/manual/tutorials/custom-domain-add-user.png#bordered)

4. 点击**提交**。

   ![用户已添加到组织](/images/manual/tutorials/custom-domain-user-added.png#bordered)

5. （可选）重复步骤 2 和 3 添加更多用户。
6. 将用户名和密码提供给成员。

:::tip 管理成员列表
作为组织管理员，你可以随时在**域名管理**页面管理组织成员列表。
:::

<!-- Block 5: Join an existing organization -->
1. 在账户创建页面，点击右上角的 <i class="material-symbols-outlined">display_settings</i>，进入**高级账户创建**页面。

   ![LarePass 中的高级账户创建选项](/images/manual/tutorials/custom-domain-advanced.png)

2. 前往**组织 Olares ID** > **加入已有组织**。
   ![加入已有组织](/images/manual/tutorials/custom-domain-join-org.png)

3. 输入带域名的用户名（如 `alice@space.n1.monster`）和管理员提供的密码。

   ![使用 Olares ID 和密码加入组织](/images/manual/tutorials/custom-domain-member-olares-id.png)

4. 点击**继续**。
