---
outline: [2, 3]
description: Configure domain settings in Olares Space with email invitation rules and member management. Administrate organizational Olares IDs for seamless team collaboration.
---

# Manage your domain

You must [add a domain](host-domain.md#add-your-domain) before managing it. Once your domain is set up, you can configure email invitation rules for organization members and invite them to create their own organizational **Olares ID** via email.

## Set email invitation rules

Most companies use a standard domain suffix for their team members' emails, like `A@myteam.com` for person A or `B@myteam.com` for person B. However, sometimes, team members might use emails in different domains. To accommodate these scenarios, Olares provides two types of rules for adding organization members' emails:

![alt text](/images/how-to/space/set_rule.jpg#bordered)

- **Fixed email suffix**: Use this option when all team members share the same email domain. Enter your organization's domain suffix (example: @company.com). Any email matching this suffix can be associated with your organization's Olares ID.

- **Specified email address**: Use this option if your organization doesn't have a corporate email suffix. You need to manually add the email address for each member of the organization.

:::info NOTE
- Currently, only Gmail is supported for both rule types.
- Emails that have been used to create organizational Olares IDs will appear in the member list and cannot be deleted.
- Emails that are manually added and have not been used to create the Olares ID appear as "unbound" and can be removed.
:::

## Manage members

After setting email rules, you can add or remove members under your organization.

![alt text](/images/how-to/space/management_members.jpg#bordered)

### Add a member

To add a member:

1. On the domain management page, add members to the organization by entering their email address.
2. Notify the corresponding users to use their email addresses to [create an organizational Olares ID](host-domain.md#create-an-org-olares-id).

### Remove a member

You can remove email addresses that haven't been used to create an Olares ID. Once an email address is associated with an organization's Olares ID, it cannot be removed.