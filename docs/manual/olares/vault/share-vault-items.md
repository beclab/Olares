---
outline: [2, 3]
description: Learn how to share vault items securely in Olares, manage team access permissions, create shared vaults, and efficiently collaborate while maintaining data security.
---

# Share vault items
A shared vault is an effective method for organizing, managing, and securely sharing data among users within the same Olares cluster. Whether you're managing family accounts or enterprise-level data, Team Vault provides the perfect balance of security and collaboration.

## Understand team roles
:::info
The Olares admin automatically becomes the shared vault owner.
:::
| Role                             | Owner | Administrator | Member |
|----------------------------------|-------|---------------|--------|
| Add, suspend, reactivate members	 | ✅     | ✖️            | ✖️     |
| Appoint administrators           | ✅     | ✖️            | ✖️     |
| Create shared Vault items        | ✅     | ✅️            | ✖️     |
| Assign read/write permissions    | ✅     | ✅             | ✖️     |

## Get started with team access
### Confirm membership
All administrators and users of an Olares cluster are automatically included in one vault team. However, for security reasons, each new member must be verified before accessing team vault items.

1. In Vault, navigate to the page **My Team** > **Invites**.
2. Click on the member's account name to view the invitation code.

   ![Invite members](/images/manual/olares/invite-members.png#bordered)
3. Send the invitation code to the corresponding member.
   :::tip
   For members, navigate to the page **Invites** > **My team** in Vault to accept invitation.
   :::
4. After the member confirms the invitation, return to the invitation page and click **Add member**.

### Set administrator
1. In Vault, navigate to the page **My team** > **Members**.
2. Select a member from the member list. 
3. Click <i class="material-symbols-outlined">more_horiz</i> in the top right corner, and select **Make admin**.
4. To remove administrator privileges, select **Remove member**.

### Suspend members
1. In Vault, navigate to the page **My team** > **Members**.
2. Select a member from the member list.
3. Click <i class="material-symbols-outlined">more_horiz</i> in the top right corner, and select **Suspend**.
4. To reactivate a member, select **Unsuspend**.

:::info
Suspended members retain their role but won't receive updates or make changes. Reactivation requires reverification for security.
:::

## Work with shared vaults
Shared vaults are designed for sharing data among multiple Olares users. By default, they must be created by Olares admin.
### Create a shared vault
1. In Vault, navigate to the page **My team** > **Vaults**.
2. Click <i class="material-symbols-outlined">add</i> in the top right corner, and enter vault name.
3. Click **Save**.

### Edit shared vault permissions
1. In Vault, navigate to the page **My team** > **Vaults**. 
2. Select the shared vault to edit permissions. You can add or remove members and set read/write permissions.
3. Click **Save**.

### Delete a shared vault
:::warning
Deleting a shared vault permanently removes all associated data. Always double-check before confirming deletion.
:::
1. In Vault, navigate to the page **My team** > **Vaults**.
2. Select the shared vault to view vault details.
3. Click <i class="material-symbols-outlined">more_horiz</i> in the top right corner, and select **Delete**.
4. In the popup dialog, enter `DELETE` to confirm deletion.
