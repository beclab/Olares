---
outline: [2, 3]
description: Personalize how your Olares applications are accessed by setting a custom route ID or your own domain name.
---

# Customize application URLs

Olares provides two ways to personalize how your applications are accessed:
- Custom route ID
- Custom domain name

## Before you begin

Familiarize yourself with these concepts for Olares applications:

- [Endpoints](../../../developer/concepts/network.md#endpoints)
- [Route ID](../../../developer/concepts/network.md#route-id)

## Custom route ID

Route ID is part of the URL used to access your Olares applications in a web browser:

`https://{routeID}.{OlaresDomainName}`

Olares uses easy-to-remember route IDs for pre-installed system applications. For community applications, you can get a simpler URL by setting a custom route ID. Using Jellyfin as an example:

1. On Olares, open Settings, then go to **Application** > **Jellyfin**.
2. Under **Entrances**, click **Jellyfin**.
3. Under **Endpoint settings**, next to **Set custom route ID**, click <i class="material-symbols-outlined">add</i>.
4. Enter a route ID that is more memorable and recognizable. For example, `jellyfin`.
5. Click **Confirm**.

   ![Custom route ID](/images/manual/olares/custom-route-id1.png#bordered){width=90%}

Now, you can access Jellyfin from your new URL: `https://jellyfin.alexmiles.olares.com`.

## Custom domain name

Instead of using the default Olares domain, you can use your own domain name to access your applications. To configure a custom domain name for an app:

:::info
Only applications with the authentication level set to **Internal** or **Public** support custom third-party domains.
:::

1. On Olares, open Settings, then go to **Application** > *AppName*.
2. Under **Entrances**, click the target entrance.
3. Under **Endpoint settings**, next to **Set custom domain**, click <i class="material-symbols-outlined">add</i>.
4. In the **Third-party domain** pop-up, enter your custom domain, and click **Confirm** to submit.

   ![Submit third-party domain](/images/manual/olares/add-custom-domain.jpeg#bordered)

   ::: tip Note
   If you are using Olares Tunnel or Self-built FRP for reverse proxy, you must also upload a valid HTTPS certificate and its private key for your custom domain.
   :::

5. Click the **Activation** button to open the activation instruction pop-up.

   ![Activate third-party domain](/images/manual/olares/activate-custom-domain.jpeg#bordered)

6. Follow the instructions in the pop-up to create a CNAME record with your domain hosting provider.

   ![Add CNAME](/images/manual/olares/add-cname.jpeg#bordered)

   :::tip Disable Proxy status for Cloudflare Tunnel
   If you are using Cloudflare Tunnel, disable the **Proxy status** option next to your DNS record. This allows Olares to receive timely updates on your domain's resolution status.
   :::

7. Click **Confirm** on the activation pop-up to finish the activation.

At this stage, the custom domain status will display as "Waiting for CNAME Activation". You will need to wait for it to take effect. DNS propagation typically takes a few minutes or hours, depending on your domain provider.

Once the CNAME record is verified, the custom domain status will automatically update to "Activated".
