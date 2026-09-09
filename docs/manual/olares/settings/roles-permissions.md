---
outline: [2, 3]
description: Compare the permissions of Member, Admin, and Super Admin roles in an Olares cluster.
head:
  - - meta
    - name: keywords
      content: Olares, roles, permissions, Super Admin, Admin, Member, team access
---

# Roles and permissions

Olares uses three built-in roles to control who can manage the cluster, shared resources, and team members. Use this reference before assigning a role or asking an administrator to perform an action.

## Role types

- **Super Admin**: The first user who activates Olares. A Super Admin has full system control and can create or manage Admin and Member accounts.
- **Admin**: An administrator created by a Super Admin. An Admin can manage the system and Member accounts, but cannot create, manage, or delete another Admin.
- **Member**: A regular user created by a Super Admin or Admin. A Member can use personal and permitted shared resources, but cannot perform cluster-wide administrative tasks.

## Permission matrix

| Permission | Member | Admin | Super Admin |
|---|:---:|:---:|:---:|
| Use system apps | ✅ | ✅ | ✅ |
| Use LarePass VPN for private access | ✅ | ✅ | ✅ |
| Connect an Olares Space account | ✅ | ✅ | ✅ |
| Customize personal app entrances | ✅ | ✅ | ✅ |
| Install regular apps from Market | ✅ | ✅ | ✅ |
| Access shared Vault items when permission is granted | ✅ | ✅ | ✅ |
| View basic system status in Control Hub | ✅ | ✅ | ✅ |
| Manage Vault teams and shared Vaults | ❌ | ✅ | ✅ |
| Install and manage shared applications | ❌ | ✅ | ✅ |
| Monitor and manage system resources | ❌ | ✅ | ✅ |
| Configure accelerator usage modes | ❌ | ✅ | ✅ |
| Update Olares | ❌ | ✅ | ✅ |
| Create, edit, and delete Members | ❌ | ✅ | ✅ |
| Create, edit, and delete Admins | ❌ | ❌ | ✅ |

For steps to add a user, change resource limits, reset a member password, or remove a user, see [Create and manage team members](manage-team.md).
