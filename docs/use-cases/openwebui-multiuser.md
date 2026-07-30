---
outline: deep
description: Share Open WebUI with other users on your Olares device by configuring public access, adding users, and sharing local models.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, multi-user, shared access, local LLM
app_version: "1.0.38"
doc_version: "2.0"
doc_updated: "2026-07-30"
---

# Set up multi-user access in Open WebUI

Configure additional user accounts in Open WebUI to share your local models and AI resources. This configuration enables multiple people to use the same Open WebUI instance on your Olares device using their own separate accounts.

## Learning objectives

In this guide, you will learn how to:

- Change the Open WebUI entrance to Public.
- Add additional user accounts.
- Share configured models with specific users or all users.
- Verify user access to shared models.

## Prerequisites

Before you begin, ensure you have the following in place:

- [Open WebUI](openwebui.md) installed and configured with at least one connected local model.
- Administrator privileges for the Open WebUI instance.

## Change entrance to Public

By default, Open WebUI restricts access to the Olares owner. To grant access to other users, change the application entrance type to Public.

1. Open Olares Settings, and then go to **Applications** > **Open WebUI** > **Entrances** > **Open WebUI**.
2. Change the **Authentication level** from **Private** to **Public**, and then click **Submit**.
   
   ![Entrance public](/images/manual/use-cases/openwebui-entrance-public.png#bordered){width=70%}

:::warning Security notice
Setting the entrance to **Public** exposes Open WebUI directly to the internet. The Open WebUI account system becomes your primary protection. Ensure the admin account and all sub-user accounts use strong passwords.
:::

## Add additional users

Create individual accounts for the people sharing your Olares device.

1. In Open WebUI, click your profile icon and select **Admin Panel**.
2. On the **Users** tab, click **Add User** in the upper-right corner.
3. In the **Add User** window, select the role and specify the user details.

   ![Add user in Open WebUI](/images/manual/use-cases/openwebui-add-user1.png#bordered)

4. Click **Save**. The new user can now sign in with these credentials you specified.

## Share local models

Models added by the admin remain private by default. Other users do not see or interact with them until you explicitly share them.

1. In Open WebUI, click your profile icon and select **Admin Panel**.
2. Select the **Settings** tab, and then choose **Models** from the left sidebar.
3. Click <span class="material-symbols-outlined">edit</span> on the right of the model you want to share.
4. Click **Access** in the upper-right corner.
5. Select one of the following access control options:

   - **Public**: All authenticated users can access this model.
   - **Private**: Only specified users can use this model.
   - Click **Add Access** to add users to the access list.

   ![Model access list](/images/manual/use-cases/openwebui-model-access-list1.png#bordered)

6. Close the **Access Control** window.
7. Scroll down and click **Save & Update**.

## Verify access

To ensure your configuration works correctly, test the new user experience.

1. Sign in to Open WebUI using the newly created user account.
2. In the chat area, check the model drop-down list. You should see the shared model.

   ![User model dropdown](/images/manual/use-cases/openwebui-subuser-model-dropdown1.png#bordered)
