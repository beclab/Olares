---
outline: [2, 3]
description: Step-by-step guide to setting up a custom domain for your Olares environment. Learn how to add and verify domains, create organizations, configure member access, and create Olares IDs under your domain.
---

# Set up a custom domain for your Olares

By default, when you create an account in LarePass, you get an Olares ID with the `olares.com` domain. This means you access your Olares services through URLs like `desktop.{your-username}.olares.com`. While this works out of the box, you might want to use your own domain instead.

:::warning Start with an unactivated Olares device
Set up the custom-domain Olares ID before you activate Olares. To complete this tutorial, you need a new device, a factory-reset device, or another device that you can activate from scratch.

If your device is already activated with an `olares.com` Olares ID, you cannot switch that activated device to a custom domain in place. [Restore Olares to factory settings](../larepass/manage-olares.md#restore-olares-to-factory-settings) first, then follow this tutorial from the beginning.
:::

## Learning objectives
In this tutorial, you will learn how to:
- Add and verify your custom domain in Olares Space
- Create an organization and an Olares ID under your domain in LarePass
- Install and activate Olares with your custom domain on an unactivated device
- Add members to your organization in Olares Space

## How custom domains work in Olares
Custom domains in Olares are managed through organizations. The custom-domain flow has two parts:

1. Create an Olares ID under your domain while your account is still in the DID stage.
2. Activate a new or factory-reset Olares device with that custom-domain Olares ID.

Whether you're an individual user or representing a company, you'll need to set up an organization first. The required actions depend on your role:

| Step                                                                                  | Organization admin | Organization member |
|---------------------------------------------------------------------------------------|--------------------|---------------------|
| Prepare a new or factory-reset Olares device that has not been activated               | ✅                  | ✅                   |
| Create a DID in LarePass                                                              | ✅                  | ✅                   |
| Add custom domain to Olares Space                                                     | ✅                  |                     |
| Create organization for the domain &<br> create an Olares ID as the admin in LarePass  | ✅                  |                     |
| Add new user to the organization in Olares Space                                      | ✅                  |                     |
| Join the organization & create an Olares ID in LarePass                               |                    | ✅                   |
| Install and activate Olares with the custom-domain Olares ID                          | ✅                  | ✅                   |

If you are joining an organization, you can skip to [Join an organization as a member](#join-an-organization-as-a-member).

## Prerequisites

Make sure you have:
- A valid domain name from a domain registrar.
- LarePass app installed on your phone. You will use LarePass to sign in to Olares Space and to associate your custom domain with an Olares ID.
- A new or factory-reset Olares device that has not been activated yet. If you want to reuse hardware that has already been activated, [restore Olares to factory settings](../larepass/manage-olares.md#restore-olares-to-factory-settings) before you continue.

## Step 1: Create a DID

<!--@include: ../../reusables/custom-domain.md#custom-domain-create-did-->

## Step 2: Add your domain

The following steps use `space.n1.monster` as an example custom domain.

1. Open [Olares Space](https://space.olares.com/) in your browser and scan the QR code with LarePass to log in.

   ![LarePass QR code scanner](/images/manual/tutorials/scan-qr-code1.png)

<!--@include: ../../reusables/custom-domain.md#custom-domain-add-domain-steps-->

## Step 3: Create an organization

In LarePass, create an organization for your domain and an Olares ID with admin privileges.

:::warning Check your account and device before continuing
Create the organization from an account in the DID stage, and make sure the Olares device you plan to use is not already activated. If you are reusing an activated device, stop here and restore it to factory settings first.
:::

<!--@include: ../../reusables/custom-domain.md#custom-domain-create-organization-->

6. Tap **Next** to navigate to the Olares activation page.
   ![Discover Olares in LarePass](/images/manual/tutorials/custom-domain-discover-olares.png#bordered)

## Step 4: Install and activate Olares

<!--@include: ../../reusables/custom-domain.md#custom-domain-install-and-activate-olares-->

## Step 5: Add members

As the admin, add members to your organization in Olares Space.

<!--@include: ../../reusables/custom-domain.md#custom-domain-add-user-->

   The member will use these credentials to create their Olares ID in LarePass.

7. (Optional) If the member will use your existing Olares instance instead of installing on a separate device, you also need to create the user on your Olares and allocate resources. See [Manage your team](../olares/settings/manage-team.md) for details. 

## Join an organization as a member

In the LarePass app, tap **Create an account** to start the account creation flow.

<!--@include: ../../reusables/custom-domain.md#custom-domain-join-organization-->

Your custom Olares ID is now created. Next, to activate Olares:

- **On a separate device**: Follow [Step 4](#step-4-install-and-activate-olares) to install and activate Olares on a new or factory-reset device.
- **On the same device as the admin**: Get the activation wizard URL and one-time password from the admin, then scan the wizard QR code in LarePass to activate. See [Activate Olares](../get-started/join-olares.md) for details. 


## Learn more

- [Olares account](../../developer/concepts/account.md): How DIDs, Olares IDs, and organizations work.
- [Install Olares](../get-started/install-olares.md): Installation options for different platforms and environments.
- [Manage your team](../olares/settings/manage-team.md): Create and manage user accounts within an Olares instance.
