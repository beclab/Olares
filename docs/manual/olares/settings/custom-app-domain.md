---
outline: [2, 3]
description: Personalize how your Olares applications are accessed by setting a custom route ID or your own domain name.
head:
  - - meta
    - name: keywords
      content: Olares, Settings, custom domain, custom route ID, application URL, CNAME, HTTPS certificate
---

# Customize application URLs

Use this guide when you want to make your Olares applications easier to access. Olares provides two ways to personalize application URLs:

- **Custom route ID**: Replace the auto-generated route ID in your Olares URL with a memorable name.
- **Custom domain name**: Use your own domain name instead of the default Olares domain.

## Prerequisites

To use a custom domain name, you need:

- A registered domain name that you control.
- Access to your domain's DNS settings to create a CNAME record.
- A valid HTTPS certificate and its RSA private key in PEM format.

## Set a custom route ID

A [route ID](../../../developer/concepts/network.md#route-id) is the part of an Olares application URL that identifies the app, such as `7e89d2a1` in `https://7e89d2a1.laresprime.olares.com`. By default, Olares assigns community applications a random route ID made of numbers and letters, which is hard to remember.

You can get a simpler URL by setting a custom route ID. Using Jellyfin as an example:
1. On Olares, open Settings, then go to **Application** > **Jellyfin**.
2. Under **Entrances**, click **Jellyfin**.
3. Under **Endpoint settings**, next to **Set custom route ID**, click <i class="material-symbols-outlined">add</i>.
4. Enter a route ID that is more memorable and recognizable. For example, `jellyfin`.
5. Click **Confirm**.

   ![Custom route ID](/images/manual/olares/custom-route-id1.png#bordered){width=90%}

Now, you can access Jellyfin from both URLs: `https://7e89d2a1.laresprime.olares.com` or `https://jellyfin.laresprime.olares.com`.

## Set a custom domain name
:::warning Custom domains disable Olares authentication
Apps with custom domains do not support Olares authentication. Only application entrances with **Authentication level** set to **Public** support custom domains.
:::

You can assign a custom domain to a single application. To use the same custom domain across all Olares services, set it up in Olares Space before activating your device. See [Set up a custom domain for your Olares](../../best-practices/set-custom-domain.md).

For example, change Jellyfin's access URL from `https://7e89d2a1.laresprime.olares.com` to `https://media.n1.monster`.

1. On Olares, open **Settings**, then go to **Application** > **Jellyfin**.
2. Under **Entrances**, click **Jellyfin**.
3. Under **Access policies**, set **Authentication level** to **Public**, then click **Submit** to apply changes.
4. Under **Endpoint settings**, next to **Set custom domain**, click <i class="material-symbols-outlined">add</i>.
5. In the **Set custom domain** pop-up, enter your custom domain, and paste the valid HTTPS certificate and private key. Both files should start with a `-----BEGIN-----` line and end with a `-----END-----` line.

   ![Submit custom domain](/images/manual/olares/settings-add-custom-domain.png#bordered)

6. Click **Confirm** to submit.

7. Click the **Activation** button to open the activation instruction pop-up.

8. Follow the instructions in the pop-up to add a CNAME record with your domain registrar. In this example, add a CNAME record with Name set to `media` and Value set to `laresprime.olares.com`.
   ![Add CNAME record](/images/manual/olares/settings-add-cname.png#bordered)

9. Click **Confirm** to close the activation pop-up. The custom domain status should display "Wait for CNAME to be activated".

DNS propagation typically takes a few minutes or hours, depending on your domain registrar.

Once the CNAME record is verified, the custom domain status will automatically update to "Activated".

Now, you can access Jellyfin from both URLs: the original Olares URL `https://7e89d2a1.laresprime.olares.com` and the custom domain `https://media.n1.monster`.

![Custom domain activated](/images/manual/olares/settings-custom-domain-activated.png#bordered)
