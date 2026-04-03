---
outline: [2, 3]
description: Set up FlareSolverr on Olares to bypass Cloudflare protection and access blocked indexer sites in Prowlarr.
head:
  - - meta
    - name: keywords
      content: Olares, FlareSolverr, Prowlarr, Cloudflare, indexer, proxy, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-03"
---

# Access Cloudflare-protected sites in Prowlarr with FlareSolverr

FlareSolverr is a proxy server that bypasses Cloudflare and DDoS-GUARD protection. Many indexer sites use Cloudflare's anti-bot measures, which can block automated access from apps like Prowlarr.

By running FlareSolverr on Olares alongside Prowlarr, you can automatically solve Cloudflare challenges and access protected indexer sites for media searches.

## Learning objectives

In this guide, you will learn how to:
- Install FlareSolverr and Prowlarr on Olares
- Configure FlareSolverr as an indexer proxy in Prowlarr
- Add Cloudflare-protected indexer sites
- Verify that FlareSolverr is solving challenges correctly

## Install FlareSolverr

1. Open Market and search for "FlareSolverr".
  <!-- ![Install FlareSolverr](/images/manual/use-cases/flaresolverr.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.


FlareSolverr runs as a background service with no user interface. You won't see an app icon on Launchpad, but you can find it in **Market** > **My Olares** or **Settings** > **Applications**.

<!-- ![FlareSolverr in My Olares and Settings](/images/manual/use-cases/flaresolverr-installed.png#bordered) -->

## Install Prowlarr

1. Open Market and search for "Prowlarr".
  <!-- ![Install Prowlarr](/images/manual/use-cases/prowlarr.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Configure FlareSolverr in Prowlarr

### Get the FlareSolverr API address

1. Open Settings, then navigate to **Applications** > **FlareSolverr**.
2. Under **Endpoint settings**, locate the API endpoint and copy the address.

   <!-- ![FlareSolverr entrance](/images/manual/use-cases/flaresolverr-entrance.png#bordered) -->

### Add FlareSolverr as an indexer proxy

1. Open Prowlarr and navigate to **Settings** > **Indexers**.
2. Under **Indexer Proxies**, click **+**, then select **FlareSolverr**.

   <!-- ![Add FlareSolverr proxy](/images/manual/use-cases/prowlarr-add-flaresolverr.png#bordered) -->

3. Configure the proxy settings:
    - **Tags**: Enter a lowercase tag name such as `flaresolverr`. This tag links specific indexers to the proxy.
    - **Host**: Paste the FlareSolverr API address you copied earlier.
4. Click the gear icon and set **Request Timeout** to `180` seconds.
5. Click **Test** to verify the connection, then click **Save**.

   <!-- ![FlareSolverr proxy settings](/images/manual/use-cases/prowlarr-flaresolverr-settings.png#bordered) -->

### Add a Cloudflare-protected indexer

This example uses 1337x, a popular indexer site protected by Cloudflare.

1. In Prowlarr, click **Indexers** > **Add Indexer** and search for "1337x".
2. Select `1337x.to` as the **Base URL**.
3. In the **Tags** field at the bottom, enter the same tag you assigned to the FlareSolverr proxy (e.g., `flaresolverr`).

   <!-- ![1337x indexer settings](/images/manual/use-cases/prowlarr-1337x-tags.png#bordered) -->

4. Click **Test**.

:::info
The challenge-solving process might take some time and does not always succeed on the first attempt. Try a few times if the initial test fails.
:::

### Verify FlareSolverr is working

You can check FlareSolverr's logs to confirm it is receiving and solving Cloudflare challenges.

1. Open Control Hub and select the **FlareSolverr** project from the sidebar.
2. Under **Deployments**, click the running pod, then expand the container to view its logs.

   <!-- ![FlareSolverr container logs](/images/manual/use-cases/flaresolverr-logs.png#bordered) -->

3. Click the play button to stream real-time logs.
4. Go back to Prowlarr and click **Test** on the 1337x indexer. You should see the incoming request appear in FlareSolverr's logs.
5. Look for `Challenge solved` in the logs. This confirms FlareSolverr has bypassed Cloudflare protection.

   <!-- ![FlareSolverr challenge solved](/images/manual/use-cases/flaresolverr-challenge-solved.png#bordered) -->

6. Search for content in Prowlarr's search bar. If results appear, FlareSolverr is working correctly.

   <!-- ![Prowlarr search results](/images/manual/use-cases/prowlarr-search-results.png#bordered) -->

## Use FlareSolverr with other indexers

When adding other indexers in Prowlarr, look for the following message:

> This site may use Cloudflare DDoS Protection, therefore Prowlarr requires FlareSolverr to access it.

<!-- ![Cloudflare warning on indexer](/images/manual/use-cases/prowlarr-cloudflare-warning.png#bordered) -->

For any indexer showing this warning, add the same FlareSolverr proxy tag (e.g., `flaresolverr`) in the indexer's **Tags** field.

## FAQ

### Prowlarr test fails but FlareSolverr logs show "Challenge solved"

If FlareSolverr successfully solves the Cloudflare challenge but Prowlarr still reports the site as blocked, you can force-save the indexer. In some cases, the indexer returns search results even when the test fails.

To force-save:

1. In the indexer settings, configure all other parameters as needed.
2. Uncheck the **Enabled** box.
3. Click **Save**, then click **Save** again to trigger a force save.
4. Open the indexer list and click the wrench icon to edit the indexer.
5. Check the **Enabled** box.
6. Click **Save**, then click **Save** again to trigger a force save.

After re-enabling, try searching in Prowlarr. If results appear, the indexer is working despite the failed test.
