---
description: Comprehensive guide to managing teams in Olares, including creating accounts, assigning roles, setting permissions, and maintaining efficient team collaboration within your Olares cluster.
---
# Manage your team
As an administrator, you can create and manage team members while ensuring optimal system performance through efficient resource controls.
## Before you begin
Ensure that:

* You have Olares admin privileges
* Your system has sufficient available resources
* Team members have created their Olares IDs

    :::info
    When creating a new member account in Olares, make sure the domain part of their Olares ID matches yours.
    :::
## Create a new member

1. Navigate to the page **Settings** > **Account**.
2. Click **Create account**.
3. In the dialog, fill in the required fields.
   - **Olares ID**: Enter the local name only.
   - **CPU**: Allocate CPU cores (minimum 1 core)
   - **Memory**: Allocate memory (minimum 3GB)
4. Click **Save**.
   Once created, you will see activation credentials for the specific Olares ID:
   - Activation wizard URL
   - One-time password
5. Share activation credentials with the new member.

You can verify whether they have completed the activation in the **Accounts** page.

:::tip
New members can activate their account through the wizard without installing Olares locally.
For detailed instructions, see [Activate Olares](../../get-started/activate-olares).
:::

## Remove a member
:::warning
Ensure users backup important data before deletion - some data cannot be recovered.
:::

1. Navigate to the page **Settings** > **Account**.
2. Click the member you want to delete to view its account details.
3. Scroll to the bottom, and click **Delete user**.
4. In the dialog, click **OK** to confirm.

## Manage resource quotas
You can adjust the allocated resources for members in your Olares cluster.

1. Navigate to the page **Settings** > **Account**.
2. Click the member you need to adjust resource quotas.
3. In the **Account info** page, scroll to the bottom and click **Modify limits**.
4. In the dialog, adjust CPU and memory quotas.
5. Click **OK** to apply changes.

## Reset passwords
1. Navigate to the page **Settings** > **Account**.
2. Click the member you need to reset password fo.
3. In the **Account info** page, scroll down and click **Reset password**. The new password will be generated immediately.
4. Share the new password with the member.

## FAQ
### Why can't I create a new user?
If you're an administrator:

* Ensure the new member has obtained their Olares ID. See [Create an Olares ID](../../get-started/create-olares-id).
* The name you entered is correctly spelled.
* There are sufficient system resources to allocate.

