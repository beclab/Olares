---
outline: [2, 3]
description: Learn how to add users to Olares One and set their roles and resource limits.
---

# Create users <Badge text="5 min"/>

On Olares One, you can create multiple user accounts to share the device securely. Each user has their own space, applications, and resource limits.

Only Super admins and Admins can create and manage user accounts.

## Learning objectives

By the end of this tutorial, you will learn how to:

- Create a new user account on Olares One.
- Assign roles and allocate CPU and memory resources.
- Share activation credentials with new users.
- Modify or remove user accounts when needed.

## Before you begin

User permissions and resource usage depend on the assigned role.

| Role | Create | Remove | Resources |
|--|--|--|--|
| **Super admin** | Admin and Members | Admin and Members | Use all resources |
| **Admin** | Members | Members | Use allocated resources |
| **Members** | — | — | Use allocated resources |

## Prerequisites

To create a new user, make sure the following conditions are met:
- **Permissions**: You are logged in as **Super admin** or **Admin**.
- **Olares ID**: The new user has a valid Olares ID that is not already activated on another Olares device.
- **Domain matching**: The domain part of the new user's Olares ID matches the Olares One server owner.
- **Resources**: Your Olares One has sufficient available CPU and memory resources.

## Create a new user

1. Go to **Settings** > **Users**.
2. Click **Create account**.
3. In the dialog, fill in the required information:

   - **Olares ID**: Enter the local name only.
   - **Role**: Choose **Members** or **Admin**.
   - **CPU**: Allocate CPU cores. Minimum 1 core.
   - **Memory**: Allocate memory. Minimum 3 GB.

4. Click **Save**.

## Share activation credentials

Once the account is created, the system generates a temporary activation wizard URL and a one-time password.

1. Click **Copy** to copy the activation credentials displayed on the screen.
2. Share them securely with the new user.
3. The user can open the URL in their browser to activate their account and set up their own password.

:::tip Remote activation 
New users do not need physical access to the Olares One device. They can complete the setup entirely using the activation URL.
:::

You can check whether the user has completed activation on the **Users** page.
![View user lists](/images/one/settings-create-users.png#bordered){width=85%}

## Manage existing users

After users are created, you can adjust their resource limits or remove access as needed.

1. Go to **Settings** > **Users**.
2. Select a user to open the **Account info** page.

![Manage users](/images/one/settings-manage-user.png#bordered){width=90%}

### Adjust resource limits

1. Click **Modify limits**.
2. Adjust CPU and memory values, then click **OK**.

### Reset password

:::tip Forgot your password
If a user forgets the password, a higher-level role can reset it. Super admins can reset passwords for Admins and Members. Admins can reset passwords for Members.
:::

1. Click **Reset password**.
2. Share the generated password with the user.

### Remove a user

:::warning
Deleting a user permanently removes their data. Proceed with caution.
:::

1. Click **Delete user**.
2. Click **OK** to confirm the deletion.

## Resources

- [Roles and permissions](/manual/olares/settings/roles-permissions.md): Learn more about roles and corresponding permissions in Olares.