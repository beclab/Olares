---
outline: [2, 3] 
description: Learn how to securely share Vault items in Olares. Understand team roles and permissions, manage shared Vault access, and enable secure collaboration among team members.
---

# Manage shared Vaults

A shared Vault is an effective way to organize, manage, and securely share sensitive data within the Olares cluster. Whether you're managing family accounts or enterprise-level data, a team Vault allows you to balance security with seamless collaboration.

## Understand team roles

All Olares administrators and members are automatically added to the **My Team** group in Vault. The Super Admin is the default owner of the Vault team, while other users become team members.

The owner can navigate to **Teams** > **Members** to view all team members and their permissions.

![Vault team](/images/manual/olares/vault-team.png#bordered)

| Role | Owner                                                                                       | Member |
|---|---------------------------------------------------------------------------------------------|---|
| Permissions | - Create, remove, and edit shared Vaults; <br/>- Assign read & write permissions to members | Access or edit shared Vault items based on assigned permissions |

## Set team access

To set team access in Vault, you first need to create a shared Vault, add items to it, and then share the Vault with specific members.

### Create a shared Vault

A shared Vault can only be created by the owner.

![Create team vault](/images/manual/olares/create-team-vault.png#bordered)

1. Navigate to the **Teams** > **Vaults**. 
2. In the team Vault page, click the <i class="material-symbols-outlined">add</i> button in the top-right corner.
3. Enter a name for the Vault and click **Save**.

At this point, this new team Vault is empty. You can [add new items](vault-items.md#add) for it or [import items from an external source](vault-items.md#import).

### Share a Vault with members

By default, the owner has automatic access and management permissions to a newly created team Vault. Team members require the owner to grant them access.

![Vault share](/images/manual/olares/vault-share.png#bordered)

1.  Navigate to **Team** > **Vaults**.
2.  In the team Vault list, click the Vault you want to share.
3.  In the details page on the right, click the **+** button to add the target member.
4.  In the permissions dropdown menu, set **Editable** (Read & Write) or **Read-only** permissions for the member.
5.  Click **Save**.

The member can now view or edit the shared Vault under the **Team Vault** category. To remove a member's access, the owner can click the <i class="material-symbols-outlined">delete</i> button next to the permission dropdown.

### Delete a shared vault

The owner can delete a shared Vault as needed.

:::warning Warning
Deleting a shared Vault will permanently remove all related data. Please check carefully before deletion.
:::

1.  Navigate to **Team** > **Vaults**.
2.  In the team Vault list, click the Vault you want to delete.
3.  In the top-right corner of the details page, click the <i class="material-symbols-outlined">more_horiz</i> menu, then select **Delete**.
4.  In the pop-up dialog, type `DELETE` to confirm the deletion.