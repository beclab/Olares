---
outline: [2, 3]
description: Learn how to add users to Olares One, assign roles and resource limits, and manage existing accounts.
---

# Create and manage users <Badge text="5 min"/>

On Olares One, you can create multiple user accounts to share the device securely. Each user has their own space, applications, and resource limits.

## Before you begin

User permissions and resource usage depend on the assigned role.

|  | **Super admin** | **Admin** | **Members** |
|--|--|--|--|
| Create | Admin and Members | Members | — |
| Remove | Admin and Members | Members | — |
| Resources | Use all resources | Use allocated resources | Use allocated resources |

## Prerequisites

**Hardware**<br>
- Your Olares One has sufficient available CPU and memory resources.

**User permissions**<br>
- You are logged in as **Super admin** or **Admin**.

**Olares ID**<br>
- The new user has a valid Olares ID that is not already activated on another Olares device.
- The domain part of the new user's Olares ID matches the current domain.

## Create a new user

1. Go to **Settings** > **Users**.
2. Click **Create account**.
3. In the dialog, fill in the required information:

   - **Olares ID**: Enter only the username (the part before `@`).
   - **Role**: Choose **Members** or **Admin**.
   - **CPU**: Allocate CPU cores. Minimum 1 core.
   - **Memory**: Allocate memory. Minimum 3 GB.

4. Click **Save**.

   Once the account is created, the system generates a temporary activation wizard URL and a one-time password.

5. Click **Copy** to copy the activation credentials, and share them with the user.

:::tip Remote activation
New users do not need physical access to Olares One. They can complete the setup entirely using the activation URL. Share [Join an Olares](/manual/get-started/join-olares) with them for the full steps.
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
If a user forgets the password, a higher-level role can reset it. Super admins can reset passwords for Admins and Members. Admins can reset passwords for Members.

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
