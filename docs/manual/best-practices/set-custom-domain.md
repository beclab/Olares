---
outline: [2, 3]
description: Step-by-step guide to setting up a custom domain for your Olares environment. Learn how to add and verify domains, create organizations, configure member access, and create Olares IDs under your domain.
---

# Set up a custom domain for your Olares

By default, when you create an account in LarePass, you get an Olares ID with the `olares.com` domain. This means you access your Olares services through URLs like `desktop.{your-username}.olares.com`. While this works out of the box, you might want to use your own domain instead.

## Learning objectives
In this tutorial, you will learn how to:
- Add and verify your custom domain in Olares Space
- Create an organization and an Olares ID under your domain in LarePass
- Install and activate Olares with your custom domain
- Add members to your organization in Olares Space

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

<!--@include: ../../reusables/custom-domain.md{21,31}-->

## Step 2: Add your domain

The following steps use `space.n1.monster` as an example custom domain.

1. Open [Olares Space](https://space.olares.com/) in your browser and scan the QR code with LarePass to log in.

   ![LarePass QR code scanner](/images/manual/tutorials/scan-qr-code1.png)

<!--@include: ../../reusables/custom-domain.md{36,73}-->

## Step 3: Create an organization

In LarePass, create an organization for your domain and an Olares ID with admin privileges.

<!--@include: ../../reusables/custom-domain.md{76,107}-->

## Step 4: Install and activate Olares

Now you can install and activate Olares with your Olares ID.

The installation steps are similar to the standard process. The following example uses Linux. For other systems, refer to the [installation guide](../get-started/install-olares.md).
<!--@include: ../get-started/install-and-activate-olares.md{5,7}-->

1. Open a terminal on the machine where you want to install Olares, and run the following command:

   ```bash
   export PREINSTALL=1 &&
   curl -sSfL https://olares.sh | bash -
   ```
   This runs a partial installation (prepare phase only) without proceeding to full setup.

<!--@include: ../get-started/install-and-activate-olares.md{9,10}-->

<!--@include: ../get-started/install-and-activate-olares.md{11,13}-->

<!--@include: ../get-started/install-and-activate-olares.md{20,20}-->

Once activation is complete, LarePass will display the desktop address of your Olares device with the custom domain, such as `https://desktop.alex.space.n1.monster`.

## Step 5: Add members

As the admin, add members to your organization in Olares Space.

<!--@include: ../../reusables/custom-domain.md{109,123}-->

   The member will use these credentials to create their Olares ID in LarePass.

7. (Optional) If the member will use your existing Olares instance instead of installing on a separate device, you also need to create the user on your Olares and allocate resources. See [Manage your team](../olares/settings/manage-team.md) for details. 

## Join an organization as a member

In the LarePass app, tap **Create an account** to start the account creation flow.

<!--@include: ../../reusables/custom-domain.md{130,141}-->

Your custom Olares ID is now created. Next, to activate Olares:

- **On a separate device**: Follow [Step 4](#step-4-install-and-activate-olares) to install and activate Olares on a new device.
- **On the same device as the admin**: Get the activation wizard URL and one-time password from the admin, then scan the wizard QR code in LarePass to activate. See [Activate Olares](../get-started/join-olares.md) for details. 


## Learn more

- [Olares account](../../developer/concepts/account.md): How DIDs, Olares IDs, and organizations work.
- [Install Olares](../get-started/install-olares.md): Installation options for different platforms and environments.
- [Manage your team](../olares/settings/manage-team.md): Create and manage user accounts within an Olares instance.