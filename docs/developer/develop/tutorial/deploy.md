---
outline: [2, 3]
description: Deploy a single-container Docker app to Olares using Studio.
---
# Deploy an app from Docker image
This guide explains how to deploy a single-container Docker app to Olares using Studio.

:::info For single-container apps
This method supports apps that run from a single container image.
:::
:::tip Recommended for testing
Studio-created deployments are best suited for development, testing, or temporary use. Upgrades and long-term data persistence can be limited compared to installing a packaged app from the Market. For production use, consider [packaging and uploading the app](package-upload.md) and installing it via the Market.
:::

## Prerequisites
- Olares version 1.12.2 or later.
- A container image for the app exists and is accessible from the Olares host.
- The app's `docker run` command or `docker-compose.yaml` is available to reference configuration (ports, environment variables, volumes).

## Create and configure your app
The following uses [Wallos](https://hub.docker.com/r/bellamy/wallos), a personal subscription and expense tracker, to show you how to map common Docker settings (image, ports, environment variables, volumes) into Studio.

**Docker examples**
::: code-group
```docker{3-6,8} [docker run command]
docker run -d \
  --name wallos \
  -v /path/to/config/wallos/db:/var/www/html/db \
  -v /path/to/config/wallos/logos:/var/www/html/images/uploads/logos \
  -e TZ=America/Toronto \
  -p 8282:80 \
  --restart unless-stopped \
  bellamy/wallos:latest
```

```yaml{5-6,7-10,12-14} [docker compose]
version: '3.0'

services:
  wallos:
    container_name: wallos
    image: bellamy/wallos:latest
    ports:
      - "8282:80/tcp"
    environment:
      TZ: 'America/Toronto'
    volumes:
      - './db:/var/www/html/db'
      - './logos:/var/www/html/images/uploads/logos'
    restart: unless-stopped
```
:::
### Create an app

1. Open Studio and select **Create a new application**.
2. Enter an **App name**, for example: `wallos`, and click **Confirm**.
3. Select **Port your own container to Olares**.
   ![Port your own container to Olares](/images/manual/olares/studio-port-your-own-container-to-olares.png#bordered)

### Configure image, port, and instance spec
These fields define the app's core components. You can find this information as the main image name and the `-p` flag in a `docker run` command, or under the `image:` and `ports:` keys in a `docker-compose.yaml` file.
1. For the **Image** field, paste the image name: `bellamy/wallos:latest`.
2. For the **Port** field, from a `HOST:CONTAINER` mapping like `8282:80`, enter the Container Port only: `80`.
   :::tip Container port only
   A port mapping is defined as `HOST:CONTAINER`. The Container Port (after the colon) is the internal port the app listens on. The Host Port (before the colon) is the external port you access. Studio manages external access automatically, so you only need to enter the Container Port.
   :::
3. For **Instance Specifications**, enter the minimum CPU and memory requirements. For example:
   - **CPU**: 2 core
   - **Memory**: 1 Gi
     ![Deploy Wallos](/images/manual/olares/studio-deploy-wallos.png#bordered)

### Add environment variables
Environment variables are used to pass configuration settings to your app. In the Docker examples, these are defined using the `-e` flag or in the `environment:` section.
1. Scroll down to **Environment Variables**, and click **Add**.
2. In this example, enter the key-value pair:
   - **key**: `TZ`
   - **value**: `America/Toronto`
3. Click **Submit**. Repeat this process for any other variables.
   ![Add environment variables](/images/manual/olares/studio-add-environment-variables.png#bordered)

### Add storage volumes
Volumes connect storage on your Olares device to a path inside the app's container, which is essential for saving data permanently. These are defined using the `-v` flag or in the `volumes:` section.

:::info Host path options
The host path is where Olares stores the data, and the mount path is the path inside the container. Studio provides three managed host path prefixes:

- `/app/data`: App data directory. Data can be accessed across nodes and is not deleted when the app is uninstalled. Appears under `/Data/studio` in Files.
- `/app/cache`: App cache directory. Data is stored in the node's local disk and is deleted when the app is uninstalled. Appears under `/Cache/<device-name>/studio` in Files.
- `/app/Home`: User data directory. Mainly used for reading external user files. Data is not deleted.
  :::
  :::info Host path rules
- The host path you enter *must* start with `/`.
- Studio automatically prefixes the full path with the app name. If the app name is `test` and you set host path `/app/data/folder1`, the actual path becomes `/Data/studio/test/folder1` in Files.
  :::

This app requires two volumes. You will add them one by one.
1. Add the database volume. This data is for high-frequency I/O and does not need to be saved permanently. Map it to `/app/cache` so it will be automatically deleted when the app is uninstalled.

   a. Click **Add** next to **Storage Volume**. 

   b. For **Host path**, select `/app/cache`, then enter `/db`.

   c. For **Mount path**, enter `/var/www/html/db`. 

   d. Click **Submit**.
2. Add the logo volume. This is user-uploaded data that should be persistent and reusable, even if the app is reinstalled. Map it to `/app/data`. 

   a. Click **Add** next to **Storage Volume**. 

   b. For **Host path**, select `/app/data`, then enter `/logos`. 

   c. For **Mount path**, enter `/var/www/html/images/uploads/logos`.

   d. Click **Submit**.
![Add volumes](/images/manual/olares/studio-add-storage-volumes.png#bordered)

You can check Files later to verify the mounted paths.
![Check mounted path in Files](/images/manual/olares/studio-check-mounted-path-in-files.png#bordered)

### Optional: Configure GPU or database middleware
If your app needs GPU, enable the **GPU** option under **Instance Specifications** and select the GPU vendor.
![Enable GPU](/images/manual/olares/studio-enable-GPU.png#bordered)

If your app needs Postgres or Redis, enable it under **Instance Specifications**.
![Enable databases](/images/manual/olares/studio-enable-databases.png#bordered)

When enabled, Studio provides dynamic variables. You must use these variables in the **Environment Variables** section for your app to connect to the database.
- **Postgres variables**

| Variables      | Description           |
|----------------|-----------------------|
| `$(PG_USER)`   | PostgreSQL username   |
| `$(PG_DBNAME)` | Database name         |
| `$(PG_PASS)`   | Postgres Password     |
| `$(PG_HOST)`   | Postgres service host |
| `$(PG_PORT)`   | Postgres service port |

- **Redis variables**

| Variables       | Description        |
|-----------------|--------------------|
| `$(REDIS_HOST)` | Redis service host |
| `$(REDIS_PORT)` | Redis service port |
| `$(REDIS_USER)` | Redis username     |
| `$(REDIS_PASS)` | Redis password     |

### Generate the app project
1. Once all your configurations are set, click **Create**. This generates the app's project files.
2. After creation, Studio generates the package files for your app, and then automatically deploys the app. You can check the status in the bottom bar.
3. When the app is successfully deployed, click **Preview** in the top-right corner to launch it.
   ![Preveiw Wallos](/images/manual/olares/studio-preview-wallos.png#bordered)

## Review the package files and test the app
Apps deployed from Studio include a `-dev` suffix in the title to distinguish them from Market installations.
![Check deployed app](/images/manual/olares/studio-app-with-dev-suffix.png#bordered)

You can click on files like `OlaresManifest.yaml` to review and make changes. For example, to change the app's display name and logo:

1. Click **<span class="material-symbols-outlined">box_edit</span>Edit** in the top-right to open the editor.
2. Click `OlaresManifest.yaml` to view the content.
3. Change the `title` field under `entrance` and `metadata`. For example, change `wallos` to `Wallos`.
4. Replace the default icon image address under `entrance` and `metadata`.
   ![Edit `OlaresManifest.yaml`](/images/manual/olares/studio-edit-olaresmanifest1.png#bordered)

5. Click <span class="material-symbols-outlined">save</span> in the top-right to save changes. 
6. Click **Apply** to reinstall with the updated package.

   :::info
   If no changes are detected since the last deployment, clicking **Apply** will simply return to the app's status page without reinstalling.
   :::
   ![Change app icon](/images/manual/olares/studio-change-app-icon1.png#bordered)

## Uninstall or delete the app
If you no longer need the app, you can remove it.
1. Click <span class="material-symbols-outlined">more_vert</span> in the top-right corner.
2. You can choose to:
   - **Uninstall**: Removes the running app from Olares, but keeps the project in Studio so you can continue editing the package.
   - **Delete**: Uninstalls the app and removes the project from Studio. This action is irreversible.

## Troubleshoot a deployment

### Cannot install the app
If installation fails, review the error at the bottom of the page and click **View** to check details.

### Run into issues when the app is running
Once running, you can manage the app from its deployment details page in Studio. The interface of this page is similar to Control Hub. If details don't appear, refresh the page.
You can:
- Use the **Stop** or **Restart** controls to retry. This action can often resolve runtime issues like a frozen process.
- Check events or logs to investigate runtime errors. See [Export container logs for troubleshooting](../../../manual/olares/controlhub/manage-container.md#export-container-logs-for-troubleshooting) for details.
  ![App deployment details](/images/manual/olares/studio-app-deployment-details.png#bordered)