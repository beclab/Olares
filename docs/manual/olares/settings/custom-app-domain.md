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

1. On Olares, open **Settings**, then go to **Applications** > **Jellyfin**.
2. Under **Entrances**, click **Jellyfin**.
3. Under **Endpoint settings**, next to **Set custom route ID**, click <i class="material-symbols-outlined">add</i>.
4. Enter a route ID that is more memorable and recognizable. For example, `jellyfin`.
5. Click **Confirm**.

   ![Custom route ID](/images/manual/olares/custom-route-id1.png#bordered){width=90%}

Now, you can access Jellyfin from your new URL: `https://jellyfin.alexmiles.olares.com`.

## Custom domain name

Instead of using the default Olares domain, you can use your own domain name to access your applications.

:::info
Custom domains do not support Olares authentication. Only application entrances with **Authentication level** set to **Public** support custom domains.
:::

Using DeerFlow 2.0 as an example:

1. On Olares, open **Settings**, then go to **Applications** > **DeerFlow 2.0**.
2. Under **Entrances**, click **DeerFlow 2.0**.
3. Under **Access policies**, set **Authentication level** to **Public**, then click **Submit**.

   ![Set authenticationn level to public](/images/manual/olares/set-auth-level-to-pub.png#bordered){width=90%}

4. Under **Endpoint settings**, next to **Set custom domain**, click <i class="material-symbols-outlined">add</i>.
5. In the **Set custom domain** pop-up, enter your custom domain, HTTPS certificate and HTTPS private key, and click **Confirm** to submit.

   ![Submit third-party domain](/images/manual/olares/add-custom-domain2.png#bordered){width=70%}

   ::: tip Note
   If you are using Olares Tunnel or Self-built FRP for reverse proxy, you must also upload a valid HTTPS certificate and its private key for your custom domain.
   :::

6. Click **Activation** to open the activation pop-up.

   ![Activate third-party domain](/images/manual/olares/activate-custom-domain2.png#bordered){width=90%}

7. The pop-up shows the CNAME record required to activate your custom domain. Follow the instructions to add the record in your domain provider's DNS settings.


   ![Add CNAME](/images/manual/olares/add-cname1.png#bordered){width=90%}

   :::tip Disable Proxy status for Cloudflare Tunnel
   If you are using Cloudflare Tunnel, disable the **Proxy status** option next to your DNS record. This allows Olares to receive timely updates on your domain's resolution status.
   :::

8. Click **Confirm**. The pop-up closes, and **Status** displays **Wait for CNAME to be activated**.

   ![Wait for CNAME to be activated](/images/manual/olares/add-cname-status.png#bordered){width=90%}

   Olares automatically verifies the CNAME record. DNS propagation may take a few minutes to 48 hours.

   Once the record is verified, the status changes to **Activated**.
