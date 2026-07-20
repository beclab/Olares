---
description: Comprehensive guide to managing teams in Olares, including creating accounts, assigning roles, setting permissions, and maintaining efficient team collaboration within your Olares cluster.
---
# Manage your team
As an administrator, you can create and manage users in your team while ensuring system security and resource efficiency.

:::tip Note on role permissions
As an administrator, Super Admin and Admin share most system management permissions, but only the **Super Admin** can create or remove Admin accounts. Admins can only create and manage Member accounts. See [Roles and permissions](roles-permissions.md) for more information.
:::

![Manage users](/images/manual/olares/manage-users.png#bordered)

## Before you begin
Ensure that:

* You have Super Admin or Admin privileges
* Your system has sufficient available resources
* The new users have created their Olares IDs

:::info
When creating a new user in Olares, make sure the domain part of their Olares ID matches yours.
:::
## Create a new user

1. Navigate to the page **Settings** > **Users**.
2. Click **Create account**.
3. In the dialog, fill in the required fields.
   - **Olares ID**: Enter the local name only.
   - **Role**: Choose **Member** or **Admin** (only Super Admin can assign this role).
   - **CPU**: Allocate CPU cores (minimum 1 core)
   - **Memory**: Allocate memory (minimum 3GB)
4. Click **Save**.
   Once created, you will see activation credentials for the specific Olares ID:
   - Activation wizard URL
   - One-time password
5. Share activation credentials with the new user.

You can verify whether they have completed the activation in the **Users** page.

:::tip
New users can activate their account through the wizard without installing Olares locally.
For detailed instructions, see [Activate Olares](../../get-started/activate-olares).
:::

## Remove a user
:::warning
Ensure users backup important data before deletion - some data cannot be recovered.
:::

1. Navigate to the page **Settings** > **Users**.
2. Click the user you want to delete to view its account details.
3. Scroll to the bottom, and click **Delete user**.
4. In the dialog, click **OK** to confirm.

## Manage resource quotas
You can adjust the allocated resources for users in your Olares cluster.

1. Navigate to the page **Settings** > **Users**.
2. Click the user you need to adjust resource quotas.
3. In the **Account info** page, scroll to the bottom and click **Modify limits**.
4. In the dialog, adjust CPU and memory quotas.
5. Click **OK** to apply changes.

## Reset passwords
1. Navigate to the page **Settings** > **Users**.
2. Click the user whose password you need to reset.
3. In the **Account info** page, scroll down and click **Reset password**. The new password will be generated immediately.
4. Share the new password with the user.

## FAQ
### Why can't I create a new user?

* You must be a Super Admin to create a new Admin. You must be a Super Admin or Admin to create a new Member.
* Ensure the new user has obtained their Olares ID. See [Create an Olares ID](../../get-started/create-olares-id).
* The name you entered is correctly spelled.
* There are sufficient system resources to allocate.

