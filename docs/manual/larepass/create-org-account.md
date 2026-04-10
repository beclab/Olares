---
outline: [2, 3]
description: Create an Olares ID with a custom domain. Set up your domain for the first time or join as a member using credentials from the domain admin.
---

# Create an Olares ID with a custom domain

:::tip
For a complete end-to-end walkthrough including domain setup, Olares installation, and member onboarding, see [Set up a custom domain for your Olares](../best-practices/set-custom-domain.md).
:::

In Olares, custom domains are managed through organizations. To use a custom domain, the domain owner first creates an organization, then adds members who can create their own Olares IDs under that domain.

## Create a DID

<!--@include: ../../reusables/custom-domain.md{21,27}-->

Once you have a DID:
- If you are the domain owner setting up the organization, continue to [Create a new organization](#create-a-new-organization).
- If the admin has already created the organization and added you, skip to [Join an existing organization](#join-an-existing-organization).

## Create a new organization

Before you start, make sure the custom domain has been [set up in Olares Space](../space/host-domain.md).

As the domain owner, create an organization and get an admin Olares ID under your custom domain.

<!--@include: ../../reusables/custom-domain.md{68,87}-->

After setting up the domain, you can [add members](../space/manage-domain.md) in Olares Space.

## Join an existing organization

If the domain admin has already created the organization and added you, use the Olares ID and password provided by the admin to join.

<!--@include: ../../reusables/custom-domain.md{104,113}-->

Your Olares ID is now created. You can proceed to [install and activate Olares](../get-started/install-olares.md).
