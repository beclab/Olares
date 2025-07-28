---
description: Understand Olares team roles and permissions. Learn about administrator responsibilities, user access levels, and effective team management structures.
---
# User roles and permissions

Olares supports multi-user operations, allowing multiple users to access the system simultaneously. Each user can securely access resources based on their assigned role and permissions.

## Role types
Olares has two default user roles:

- **Super Admin**: The first user to activate and log into Olares. Has full, unrestricted control of the system and can create other Admin and Member accounts.
- **Admin**: Created by the Super Admin. Has nearly the same system management permissions as the Super Admin. **Can only create and manage Members**, not other Admin accounts.
- **Member**: Standard users created by Super Admin or Admin, with limited system resources and access permissions.

This structure ensures organizations can scale Olares management securely, with multiple Admins sharing responsibility while the Super Admin retains ultimate authority.

## Role permissions

| Permission Area | Member | Admin | Super Admin |
|-----------------|--------|-------|-------------|
| Use system apps (Files, Vault, Wise, Profile, Dashboard, Control Hub) | ✅ | ✅ | ✅ |
| Enable VPN for private entrances | ✅ | ✅ | ✅ |
| Connect to Olares Space | ✅ | ✅ | ✅ |
| Customize app entrances | ✅ | ✅ | ✅ |
| Install regular apps from Market | ✅ | ✅ | ✅ |
| Access shared vaults with assigned permissions | ✅ | ✅ | ✅ |
| View basic system status in Control Hub | ✅ | ✅ | ✅ |
| Manage Vault teams & shared vaults | ❌ | ✅ | ✅ |
| Install and manage shared apps | ❌ | ✅ | ✅ |
| Monitor and manage system resources | ❌ | ✅ | ✅ |
| Set GPU usage modes | ❌ | ✅ | ✅ |
| Update Olares versions | ❌ | ✅ | ✅ |
| Create, edit, and delete Members | ❌ | ✅ | ✅ |
| Create, edit, and delete Admins | ❌ |❌| ✅ |


