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

`https://{routeID}.{OlaresDomainName}`

For convenience, Olares uses easy-to-remember route IDs for pre-installed system applications.
For community applications, you can quickly obtain a simple and memorable URL by changing the route ID. To use Jellyfin as an example:

1. Open the Settings app, and select **Application** from the left sidebar.
2. Click **Jellyfin** on the right to view application details.
3. Go to **Entrances** > **Set up endpoint**. You can see the default Route ID for Jellyfin, which is a combination of numbers and letters.
4. Next to **Set custom Route ID**, click <i class="material-symbols-outlined">add</i>.
5. Enter a route ID that is more memorable and recognizable. For example, `jellyfin`.

   ![Custom route ID](/images/manual/olares/custom-route-id.png#bordered)
6. Click **Confirm**.

Now, you will be able to access Jellyfin from your new URL: `https://jellyfin.bob.olares.com`.

### Custom domain name
Instead of using the default Olares domain, you can use your own domain name to access your applications, making them more professional and easier to remember. To use Affine as an example:
:::info
Only applications with the authentication level set to **Internal** or **Public** support custom third-party domains.
:::
1. Open the Settings app, and select **Application** from the left sidebar.
2. Click Affine on the right to view application details.
3. Go to **Entrances** > **Set up endpoint**.
4. Next to **Set custom domain**, click <i class="material-symbols-outlined">add</i>.

   ![Set third-party domain](/images/manual/olares/set-custom-domain.png#bordered)
5. Enter your custom domain, for example, `hello.coffee`, and click **Confirm**.
6. Click **Activation**, and follow the instructions to add a CNAME record on your domain hosting site. Then click **Confirm**.

   ![Activate third-party domain](/images/manual/olares/activate-custom-domain.png#bordered)
   At this stage, the custom domain status will display as "Waiting for CNAME Activation". This means you need to wait for the DNS changes to propagate. The propagation time typically ranges from a few minutes to 48 hours, depending on your domain provider.

   Olares will periodically check if the DNS record is correctly configured. Once the CNAME record is verified, the custom domain status will automatically update to "Activated". After activation, you can access Affine using the new URL: `hello.coffee`.
:::tip
To allow public access to your custom domain without login, update the access policies as below:
1. Navigate to **Settings** > **Application**, and click the target application.
2. Click **Entrance**, then under **Create access policies**, set **Authentication level** to **Public**.
3. Click **Submit** to apply changes.
   
   ![Set auth level to public](/images/manual/olares/set-auth-level-to-public.png){width=50%}
:::
