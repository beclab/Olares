---
outline: deep
description: Share Open WebUI with other users on your Olares device by configuring public access, adding sub-users, and sharing local models.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, multi-user, shared access, local LLM
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-05-13"
---

# Set up multi-user access

You can add sub-users to Open WebUI and share your local models with them. This lets multiple people on your Olares device use the same Open WebUI instance with their own accounts.

## Prerequisites

- [Open WebUI installed and configured](openwebui.md) on Olares
- At least one model backend configured (Ollama or a model app)
- Admin privileges on the Open WebUI instance

## Change entrance to Public

By default, Open WebUI is accessible only to the Olares owner. To let other users access it, you need to change its entrance type to Public.

1. Open Olares Settings and navigate to **Applications** > **Open WebUI**.
2. In **Entrances**, switch the access type to **Public**.
   <!-- ![Entrance public](/images/manual/use-cases/openwebui/entrance-public.png#bordered) -->

:::warning Security notice
When you set the entrance to Public, other users can access Open WebUI directly from the public internet. Open WebUI's account system is your only protection. Use a strong password for the admin account, and ensure all sub-users also use strong passwords.
:::

## Add sub-users

1. In Open WebUI, click your **profile icon** in the bottom-left corner and select **Admin Panel**.
2. Navigate to **Users**.
3. Click **Add User** and fill in the user details.
   <!-- ![User management](/images/manual/use-cases/openwebui/user-management.png#bordered) -->
4. Save the new user. The user can now sign in with the credentials you created.

## Share local models

Models added by the admin are private by default. Other users cannot see or use them until you share them.

1. In the Admin Panel, navigate to **Settings** > **Models**.
2. Click the edit icon next to the model you want to share.
   <!-- ![Model visibility public](/images/manual/use-cases/openwebui/model-visibility-public.png#bordered) -->
3. Change the access control to one of the following:
   - **Public**: All users can use this model.
   - **Private with access list**: Only specified users can use this model. Add users to the access list.
   <!-- ![Model access list](/images/manual/use-cases/openwebui/model-access-list.png#bordered) -->
4. Save your changes.

## Verify access

1. Sign in to Open WebUI as a sub-user.
2. Start a new chat.
3. Open the model dropdown. You should see the shared models.
   <!-- ![User model dropdown](/images/manual/use-cases/openwebui/user-model-dropdown.png#bordered) -->
