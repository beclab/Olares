---
outline: [2, 3]
description: Customize Olares application URLs with personalized domain names and route IDs. Learn how to set up public access and manage application endpoints.
---

# Customize application domains
You can access Olares applications anytime, anywhere, whether you're accessing from local or remotely. This guide will help you:
- Personalize domain name for your applications
- Allow public access without authentication

## Before you begin
Before you start, it is recommended to familiarize yourself with a few concepts for Olares applications:

- [Endpoints](../../concepts/network.md#endpoints)
- [Route ID](../../concepts/network.md#route-id)

## Customize domain name for application

Olares provides two methods to optimize application access addresses:
* Custom route ID
* Custom domain name

### Custom route ID
Route ID is a crucial component in accessing your Olares applications. It forms part of the URL you use to reach your applications through a web browser:

`https://{routeID}.local.{OlaresDomainName}`

::: info
 `local` in a URL signifies a private access point, typically used by apps offering services for individual users, families, or teams. Public-facing apps hosting external services do not include `local` in their URLs. For more details, see [entrance](../../concepts/network.md#entrance).
:::

For convenience, Olares uses easy-to-remember route IDs for pre-installed system applications.
For community applications, you can quickly obtain a simple and memorable URL by changing the route ID. To use Jellyfin as an example:

1. Open the Settings app, and select **Application** from the left sidebar.
2. Click **Jellyfin** on the right to view application details.
3. Go to **Entrances** > **Set up endpoint**. You can see the default Route ID for Jellyfin, which is a combination of numbers and letters.
4. Next to **Set custom Route ID**, click <i class="material-symbols-outlined">add</i>.
5. Enter a route ID that is more memorable and recognizable. For example, `jellyfin`.

   ![Custom route ID](/images/manual/olares/custom-route-id.jpeg#bordered)
6. Click **Confirm**.

Now, you will be able to access Jellyfin from your new URL: `https://jellyfin.local.bob.olares.com`.

### Custom domain name
Instead of using the default Olares domain, you can use your own domain name to access your applications, making them more professional and easier to remember. To configure a custom domain name for an app:

:::info
Only applications with the authentication level set to **Internal** or **Public** support custom third-party domains.
:::
1. Open the Settings app, and select **Application** from the left sidebar.
2. Click the application you want to configure to enter its details page.
3. Go to **Entrances** > **Set up endpoint**, and click <i class="material-symbols-outlined">add</i> next to **Set custom domain**.
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
   If you are using Cloudflare Tunnel, disable the **Proxy status** option next to your DNS record. This allows Olares to receive in-time updates on your domain's resolution status.
   :::

7. Click **Confirm** on the activation popup window to finish the activation.
 
At this stage, the custom domain status will display as "Waiting for CNAME Activation". You will need to wait for it to take effect. DNS propagation typically takes a few minutes or hours, depending on your domain provider.

Once the CNAME record is verified, the custom domain status will automatically update to "Activated".
