---
description: Find answers to common questions about Olares.
---
# FAQs

## What license is Olares using?

Olares consists of a series of projects using a hierarchical authorization approach. The basic principles are:

- Projects running on blockchain use Apache 2.0, such as [Snowinning Protocol](https://github.com/beclab/terminusdid-contract-system).
- Projects related to protocols use Apache 2.0, such as [r4](https://github.com/beclab/r4).
- Projects around Olares and LarePass use the [Olares License](https://github.com/beclab/Olares/blob/main/LICENSE.md).
- For third-party applications running on Olares, it is up to the developer to decide whether they want them open source or not and choose the license accordingly.

For more details, visit our projects on [GitHub](https://github.com/beclab).

## Is the Olares License an open source license?

Olares's choice of license for its major projects is inspired by [fair code](https://faircode.io/). The [Olares License](https://github.com/beclab/Olares/blob/main/LICENSE.md) also follows these principles:

> - Is generally free to use and can be distributed by anybody
> - Has its source code openly available
> - Can be extended by anybody in public and private communities
> - Is commercially restricted by its authors

## Why can't I restore my account if the mnemonics goes missing?

From 1Password’s MasterKey to crypto wallet’s mnemonic phrase, for more than ten years, the problem of mnemonic storage has not been well solved.

The mnemonic phrase of Olares will be encrypted and stored on all devices that install LarePass. Generally, you only lose the mnemonic phrase if you lose all the devices with LarePass installed at the same time.

Safety is the most important principle in designing our system. We will continue to improve it in the future to provide you with a better solution that balances convenience and safety.

## Is there a difference between Olares and the current operating systems running on NAS?

At the inception of Olares (formerly Terminus), the market already had excellent NAS operating systems such as [Synology](https://www.synology.com/en-global/dsm/packages), [CasaOS](https://github.com/IceWhaleTech/CasaOS), and [Umbrel](https://github.com/getumbrel/umbrel). They have indeed inspired us.

But we do think the operating system running on Edge should be able to:

- Orchestrate resources for multiple hardware
- Manage applications in sandboxes

This is difficult to achieve with the above-mentioned NAS operating systems built on Docker Compose.

Meanwhile, Olares aims to provide a one-stop self-hosted solution, which goes beyond the scope of general NAS operating systems.

## Do I need to pay for Olares?

When you're self-hosting, you can essentially use Olares for free.

But for the following two features, we may introduce reasonable charges due to the cost (currently both are provided for free):

- **Backup**

  We provide 10G of free backup space for each Olares ID on Olares Space. When the archive size exceeds this limit, we will charge you a certain fee to cover the cloud provider fee.

- **Fast Reverse Proxy (FRP)**

   Accessing Olares locally or via VPN is essentially free. However, if you’re providing external services like hosting a blog, traffic will be forwarded to a Fast Reverse Proxy (FRP) server before reaching Olares. In this case:

   - If you use your own FRP service, Olares does not impose any charges.
   - If you opt to use the default FRP service from Olares, we offer a free monthly traffic allowance of 2GB. This is usually sufficient for users who do not provide external services through Olares. Additional charges may apply if your usage exceeds this limit.

## When are other languages available?

Right now we only support English and Simplified Chinese.

In fact, we have completed i18n replacement in all front-end projects. However, we lack the experience in maintaining translation resources for a fast iterating project through the open source community. We are still learning.

## What are the differences among the different "passwords"?

Olares does have various passwords to ensure its security, including:

- Private key
- The password of LarePass:
    - On mobile phones, biometrics can be used for login
    - On computers and browser plug-ins, manual input is required
- Password for first activation of Olares
- Password for Olares login
- Second verification code when logging in to Olares

Don't panic! For daily use, what you need to enter is the two-step verification code when logging in to Olares.

## How to deploy multi-user applications?

It depends on whether you want to provide external service or simply let internal Olares users use it.

- To provide services to the public, you can select the Entrance to access the application as **Public**. This allows the application to manage its own user registration and authentication.
- To provide internal access only, you can deploy the Cluster-scoped version of such products on Olares.

For Gitlab, we provide two versions of porting: [Gitlab Pure](https://github.com/beclab/apps/tree/main/gitlabpure) and [Gitlab Fusion](https://github.com/RLovelett/gitlab-fusion).

## How can I reactivate Olares with the same Olares ID?

If you've reinstalled Olares, the Olares instance you originally activated will no longer be accessible. To reactivate Olares using the same Olares ID:

::: tip Install with the same Olares ID
During the Olares installation, ensure that you have entered the exact same domain and Olares ID that you used previously. 
:::

![Reactivate](/images/manual/help/reactivate.png)

  1. Open LarePass on your phone and enter your previous account. You should see a red prompt on the top saying "No active Olares found".
  2. Tap **Learn more** > **Reactivate** to enter the QR scan screen.
  3. Tap **Scan QR code** to scan the QR code on the wizard page and activate Olares.
