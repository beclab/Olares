---
outline: [2, 3]
description: Learn how to use the LarePass app to check for and install system updates on your Olares One.
head:
  - - meta
    - name: keywords
      content: Olares One, Update Olares
---
# Update OS <Badge type="tip" text="15 min" />

Use the LarePass app to keep your Olares One up to date with the latest features and security patches.

:::warning Admin access required
Only an Olares administrator can perform system updates. Please note that updates apply to the entire Olares cluster and affect all members.
:::

## Learning objectives

By the end of this tutorial, you will learn how to:
- Check whether a new Olares OS version is available in LarePass.
- Choose between **Download only** and **Download and upgrade**.
- Install an update and confirm the system restarts successfully.

## Prerequisites

Make sure that:
- Your Olares One is powered on and connected to a network.
- Your phone can access Olares One through LarePass.
- You are signed in with an account that has administrator permissions.

## Check for updates

1. Open LarePass on your phone and go to **Settings**.
2. In the **My Olares** card, tap **System** to enter the **Olares management** page.
3. Tap **System update**.
4. If a new version appears in the **New version** field, tap **Upgrade**.
![Check for available version](/images/manual/larepass/check-version.png#bordered)

## Select upgrade method

A dialog will appear asking how you want to handle the update package. Choose the method that fits your schedule:

- **Download only**: Olares downloads the package in the background. You can continue using the system and install the update manually later.

- **Download and upgrade**: Olares downloads the package and prompts you to restart and install it when the download finishes.

    ![Choose upgrade method](/images/manual/larepass/olares-upgrade1.png#bordered)

## Install and restart

1. Start installation based on your selection:
   - **Download only**: After the download completes, return to **System update** and tap **Upgrade now**. Confirm the restart prompt to begin installation.
   - **Download and upgrade**: After the download completes, confirm the restart prompt to begin installation.
2. Wait for the update and restart to finish. When Olares starts up again, you'll see a success message and can continue using it.