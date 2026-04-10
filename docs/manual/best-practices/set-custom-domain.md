---
outline: [2, 3]
description: Step-by-step guide to setting up a custom domain for your Olares environment. Learn how to add and verify domains, create organizations, configure member access, and create Olares IDs under your domain.
---

# Set up a custom domain for your Olares

By default, when you create an account in LarePass, you get an Olares ID with the `olares.com` domain. This means you access your Olares services through URLs like `desktop.{your-username}.olares.com`. While this works out of the box, you might want to use your own domain instead.

## Learning objectives
In this tutorial, you will learn how to:
- Add and verify your custom domain in Olares Space
- Create an organization to manage your custom domain
- Configure member access for your organization
- Create an Olares ID under your custom domain
- Install and activate Olares with your Olares ID

## How custom domains work in Olares
Custom domains in Olares are managed through organizations. Whether you're an individual user or representing a company, you'll need to set up an organization first. The required actions depend on your role:

| Step                                                                              | Organization admin | Organization member |
|-----------------------------------------------------------------------------------|--------------------|---------------------|
| Create a DID in LarePass                                                          | ✅                  | ✅                   |
| Add custom domain to Olares Space                                                 | ✅                  |                     |
| Create organization for the domain &<br> create an Olares ID as the admin in LarePass | ✅                  |                     |
| Add new user to the organization in Olares Space                                  | ✅                  |                     |
| Join the organization & create an Olares ID in LarePass                           |                    | ✅                   |
| Set up Olares                                                                     | ✅                  | ✅                   |

If you are joining an organization, you can skip to [Join an organization as a member](#join-an-organization-as-a-member).

## Prerequisites

Make sure you have:
- A valid domain name from a domain registrar.
- LarePass app installed on your phone. You will use LarePass to sign in to Olares Space and to associate your custom domain with an Olares ID.

:::info
If you have previously installed and activated an Olares instance on your device and want to reuse the same hardware with a custom domain, [perform a factory reset](../larepass/manage-olares.md#restore-olares-to-factory-settings) first, then [create a new account](../larepass/create-account.md) to follow this tutorial.
:::

## Step 1: Create a DID

<!--@include: ../../reusables/custom-domain.md{21,27}-->

## Step 2: Add your domain to Olares Space

Open [Olares Space](https://space.olares.com/) in your browser and scan the QR code with LarePass to log in.

![LarePass QR code scanner](/images/manual/tutorials/scan-qr-code1.png)

<!--@include: ../../reusables/custom-domain.md{30,65}-->

## Step 3: Create an organization for the domain

This step links your domain to an organization in Olares and requests the Verifiable Credential (VC) for the domain.

::: tip Verifiable Credential
A Verifiable Credential (VC) is a digital certificate that confirms your organization's ownership of the domain.
:::

<!--@include: ../../reusables/custom-domain.md{68,87}-->

## Step 4: Install and activate Olares

Now you can install and activate Olares with your Olares ID.

The installation steps are similar to the standard process. The following example uses Linux. For other systems or a clean install with an ISO image, refer to the [installation guide](../get-started/install-olares.md).
<!--@include: ../get-started/install-and-activate-olares.md{5,7}-->

1. In the terminal, run the following script to start the installation:

   ```bash
   export PREINSTALL=1 &&
   curl -sSfL https://olares.sh | bash -
   ```
   This runs a partial installation (prepare phase only) without proceeding to full setup.

<!--@include: ../get-started/install-and-activate-olares.md{9,13}-->

<!--@include: ../get-started/install-and-activate-olares.md{20,20}-->

Once activation is complete, LarePass will display the desktop address of your Olares device with the custom domain, such as `https://desktop.admin123.olares.hellocoffee.online`.

## Step 5: Add a new user within the same organization

<!--@include: ../../reusables/custom-domain.md{90,101}-->

(Optional) If the member will join the admin's existing Olares instance instead of installing on a separate device, go to **Settings** > **Users** in the admin's Olares and create a user account for the member. Share the activation wizard URL and one-time password with the member. See [Manage your team](../olares/settings/manage-team.md) for details.

## Join an organization as a member

In the LarePass app, tap **Create an account** to start the account creation flow.

<!--@include: ../../reusables/custom-domain.md{104,113}-->

Your custom Olares ID is now created. Next, activate Olares:

- **Separate device**: Follow [Step 4](#step-4-install-and-activate-olares) to install and activate Olares on a new device.
- **Same device as the admin**: Get the activation wizard URL and one-time password from the admin, then scan the wizard QR code in LarePass to activate. See [Activate Olares](../get-started/join-olares.md) for details. 


## Learn more

- [Olares account](../../developer/concepts/account.md): How DIDs, Olares IDs, and organizations work.
- [Install Olares](../get-started/install-olares.md): Installation options for different platforms and environments.
- [Manage your team](../olares/settings/manage-team.md): Create and manage user accounts within an Olares instance.
