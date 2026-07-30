---
outline: [2, 3]
title: Set up FlareSolverr with Prowlarr on Olares
description: Set up FlareSolverr with Prowlarr on Olares to bypass Cloudflare protection and access blocked indexer sites for your media stack.
head:
  - - meta
    - name: keywords
      content: Olares, FlareSolverr, Prowlarr, Cloudflare, indexer, proxy, self-hosted, setup FlareSolverr with Prowlarr, FlareSolverr Prowlarr, Prowlarr FlareSolverr, FlareSolverr Kodi, ghcr.io FlareSolverr
app_version: "1.0.4"
doc_version: "1.1"
doc_updated: "2026-07-29"
---

# Access Cloudflare-protected sites in Prowlarr with FlareSolverr

FlareSolverr is a proxy server that solves Cloudflare and DDoS-GUARD challenges on behalf of apps like Prowlarr. When an indexer site shows a CAPTCHA or JavaScript challenge, FlareSolverr uses a headless browser to complete the challenge and return the valid response.

On Olares, you can set up FlareSolverr as an indexer proxy for Prowlarr. Once configured, Prowlarr routes requests for protected indexers through FlareSolverr, so your *arr apps can keep searching and downloading content.

This guide walks you through installing FlareSolverr, connecting it to Prowlarr, and adding Cloudflare-protected indexers such as 1337x.

## Learning objectives

In this guide, you will learn how to:
- Install FlareSolverr on Olares.
- Configure FlareSolverr as an indexer proxy in Prowlarr.
- Add Cloudflare-protected indexer sites.
- Verify that FlareSolverr is solving challenges correctly.

## Prerequisites

- An Olares device that is set up and running.
- Prowlarr installed and running. See [Set up the *Arrs stack](./arrs.md) for installation steps.

## Install FlareSolverr

1. Open Market and search for "FlareSolverr".
   ![Install FlareSolverr](/images/manual/use-cases/install-flaresolverr.png#bordered){width=90%}

2. Click **Get**, then **Install**, and wait for installation to complete.


:::info
FlareSolverr runs as a background service, so it does not appear on Launchpad.

You can find it in **Settings** > **Applications** or **Market** > **My Olares**.
:::

In **Settings** > **Applications**, look for FlareSolverr in the app list.

![FlareSolverr in Settings](/images/manual/use-cases/flaresolverr-installed-in-settings.png#bordered){width=60%}

Or open **Market** > **My Olares** to see FlareSolverr among your installed apps.

![FlareSolverr in My Olares](/images/manual/use-cases/flaresolverr-installed-in-market.png#bordered){width=60%}

## Configure FlareSolverr in Prowlarr

### Get the FlareSolverr API address

1. Open Settings, then navigate to **Applications** > **FlareSolverr**.
2. Under **Entrances**, click **FlareSolverr**.
3. Under **Endpoint settings**, locate the API endpoint and click <i class="material-symbols-outlined">content_copy</i> to copy the address.
   ![FlareSolverr entrance](/images/manual/use-cases/flaresolverr-endpoint.png#bordered){width=80%}

### Add FlareSolverr as an indexer proxy

1. Open Prowlarr from Launchpad.
   :::tip First-time setup
   The first time you open Prowlarr, the **Authentication Required** dialog appears. Select an authentication method, set a username and password, then click **Save** to continue.
   :::
2. Go to **Settings** > **Indexers**.
3. Under **Indexer Proxies**, click <i class="material-symbols-outlined">add_2</i>, then select **FlareSolverr**.

   ![Add FlareSolverr proxy](/images/manual/use-cases/prowlarr-add-flaresolverr.png#bordered){width=80%}

4. Configure the proxy settings:
    - **Tags**: Enter a lowercase tag name such as `flaresolverr`. Prowlarr uses this tag to decide which indexers should route requests through FlareSolverr.
    - **Host**: Paste the FlareSolverr API address you copied earlier.
5. Click the gear icon and set **Request Timeout** to `180` seconds.
6. Click **Test** to verify the connection, then click **Save**.
   ![FlareSolverr proxy settings](/images/manual/use-cases/prowlarr-flaresolverr-settings.png#bordered){width=80%}

### Add a Cloudflare-protected indexer

This example uses 1337x, a popular indexer site protected by Cloudflare.

1. In Prowlarr, click **Indexers** > **Add Indexer** and search for "1337x".
2. Select **1337x**, then set **Base URL** to `1337x.to`.
3. In the **Tags** field at the bottom, enter the same tag you assigned to the FlareSolverr proxy (e.g., `flaresolverr`).
   ![1337x indexer settings](/images/manual/use-cases/prowlarr-1337x-tags.png#bordered){width=80%}

4. Click **Test**.

:::info
The challenge-solving process might take some time and does not always succeed on the first attempt. Try a few times if the initial test fails.
:::

### Verify FlareSolverr is working

You can check FlareSolverr's logs to confirm it is receiving and solving Cloudflare challenges.

1. Open Control Hub and select the **FlareSolverr** project from the sidebar.
2. Under **Deployments**, click the running pod of **flaresolverr**, then expand the container to view its logs.

   ![FlareSolverr container logs](/images/manual/use-cases/flaresolverr-logs.png#bordered){width=80%}

3. Click the play button to stream real-time logs.
4. Go back to Prowlarr and click **Test** on the 1337x indexer. You should see the incoming request appear in FlareSolverr's logs.
5. Look for `Challenge solved` in the logs. This confirms FlareSolverr has bypassed Cloudflare protection.
   ![FlareSolverr challenge solved](/images/manual/use-cases/flaresolverr-challenge-solved.png#bordered){width=80%}

6. Search for content in Prowlarr's search bar. If results appear, FlareSolverr is working correctly.
   ![Prowlarr search results](/images/manual/use-cases/prowlarr-search-results.png#bordered){width=80%}

## Use FlareSolverr with your media stack

FlareSolverr is most commonly used with Prowlarr, but the same proxy setup benefits other parts of a self-hosted media workflow:

- **Prowlarr** sends indexer queries through FlareSolverr when an indexer tag matches the proxy tag.
- **Sonarr**, **Radarr**, and **Lidarr** rely on Prowlarr's indexers, so they indirectly benefit from FlareSolverr without extra configuration.
- Media centers such as **Kodi**, **Jellyfin**, and **Plex** use the files that Sonarr and Radarr grab, which means a working FlareSolverr setup helps keep your entire media stack fed with new content.

If you run into Cloudflare-protected indexers in other *arr apps, make sure the indexer is managed through Prowlarr and tagged correctly.

## FAQ

### Prowlarr test fails but FlareSolverr logs show "Challenge solved"

If FlareSolverr successfully solves the Cloudflare challenge but Prowlarr still reports the site as blocked, you can force-save the indexer. In some cases, the indexer returns search results even when the test fails.

To force-save:

1. In the indexer settings, configure all other parameters as needed.
2. Uncheck the **Enable** checkbox, then click **Save**. The indexer is now added to the list in a disabled state.
3. On the **Indexers** page, locate the indexer in the list and click the wrench icon on the far right to edit it. 
   ![Edit indexer from Indexers list](/images/manual/use-cases/prowlarr-cloudflare-edit-indexer.png#bordered){width=80%}
   
4. Select the **Enable** checkbox.
5. Click **Save**.
6. If Prowlarr keeps the edit page open and displays a warning (e.g., `Unable to access 1337x.to, blocked by CloudFlare Protection`), the **Save** button will change to a red warning icon. Click this icon again to save the indexer despite the failed connection check.

After re-enabling, try searching in Prowlarr. If results appear, the indexer is working despite the failed test.

### FlareSolverr is not receiving requests from Prowlarr

Check the following:

- The **Host** field in Prowlarr's FlareSolverr proxy settings matches the API endpoint you copied from Olares Settings.
- The indexer **Tags** field contains the same tag assigned to the FlareSolverr proxy.
- The FlareSolverr pod is running in Control Hub.

### How do I update FlareSolverr

When a new version of FlareSolverr is published to the Olares Market, open Market, go to **My Olares**, and click **Update** next to FlareSolverr. Olares pulls the latest `ghcr.io/flaresolverr/flaresolverr:latest` image automatically.

## Learn more

- [Stream your media with Jellyfin](jellyfin.md): Host and stream movies, shows, and music from your Olares device.
