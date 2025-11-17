---
outline: [2, 3]
description: Deploy a single-container Docker app to Olares using Studio.
---
# Deploy an app from Docker image
This guide explains how to deploy a single-container Docker app to Olares using Studio.

:::info For single-container apps
This method supports apps that run from a single container image. For multi-container apps (for example, a web service plus a separate database), use the workflow in the [developer documentation](../../../developer/develop/tutorial/index.md) instead.
:::
:::tip Recommended for testing
Studio-created deployments are best suited for development, testing, or temporary use. Upgrades and long-term data persistence can be limited compared to installing a packaged app from the Market. For production use, consider [packaging and uploading the app](package-upload.md) and installing it via the Market.
:::

## Prerequisites
- Olares version 1.12.2 or later.
- A container image for the app exists and is accessible from the Olares host.
- The app's `docker run` command or `docker-compose.yaml` is available to reference configuration (ports, environment variables, volumes).

## Step 1: Create an app
1. Open Studio and select **Create a new application**.
2. Enter an **App name**, for example: `test`, and click **Confirm**.
   :::info
   Use only lowercase letters and numbers.
   :::
    
3. Select **Port your own container to Olares**.
   ![Port your own container to Olares](/images/manual/olares/studio-port-your-own-container-to-olares.png#bordered)
## Step 2: Configure deployment information
The following examples show how to map common Docker settings (image, ports, environment variables, volumes) into Studio.

### Option 1: Deploy a simple app

Use this if the app does not require environment variables, volumes, databases, or GPUs.

This example uses [LibreSpeed](https://hub.docker.com/r/linuxserver/librespeed), a lightweight speed test app.

The corresponding `docker run` command and the `docker compose` file are as follows:
::: code-group
```docker [docker run command]
docker run -d \
  --name=librespeed \
  -p 80:80 \
  lscr.io/linuxserver/librespeed:latest
```

```yaml [docker compose]
---
services:
  librespeed:
    image: lscr.io/linuxserver/librespeed:latest
    container_name: librespeed
    ports:
      - "80:80"
```
:::
1. For the **Image** field, paste the image name: `lscr.io/linuxserver/librespeed:latest`.
2. For the **Port** field, from a `HOST:CONTAINER` mapping like `80:80`, enter the container port only: `80`.
   :::tip Container port only
   When a port is defined as `HOST:CONTAINER` in Docker, Studio manages the host port automatically. Enter only the container port (the number after the colon).
   :::
3. For **Instance Specifications**, enter the minimum CPU and memory requirements. For example:
   - **CPU**: 1 core
   - **Memory**: 128 Mi
4. Click **Create** to generate the app project.
   ![Deploy Librespeed](/images/manual/olares/studio-deploy-librespeed.png#bordered)

### Option 2: Deploy an app with environment variables and storage
Use this if the app requires environment variables or persistent storage.

This example uses [Wallos](https://hub.docker.com/r/bellamy/wallos), a personal subscription and expense tracker.

The corresponding `docker run` command and the `docker compose` file are as follows:
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
    # Volumes store your data between container upgrades
    volumes:
      - './db:/var/www/html/db'
      - './logos:/var/www/html/images/uploads/logos'
    restart: unless-stopped
```
:::
1. For the **Image** field, paste the image name: `bellamy/wallos:latest`.
2. For the **Port** field, from a `HOST:CONTAINER` mapping like `8282:80`, enter the container port only: `80`.
3. For **Instance Specifications**, enter the minimum CPU and memory requirements. For example:
   - **CPU**: 2 core
   - **Memory**: 1 G
     ![Deploy Wallos](/images/manual/olares/studio-deploy-wallos.png#bordered)

4. Add environment variables:
   1. Scroll down to **Environment Variables**, and click **Add**.
   2. In this example, enter the key-value pair:
      - **key**: `TZ`
      - **value**: `America/Toronto`
   3. Click **Submit**. Repeat this process for any other variables.
   ![Add environment variables](/images/manual/olares/studio-add-environment-variables.png#bordered)

5. Add storage volumes. This app requires two volumes.
   1. Review the host path options and rules.
      :::info Host path options
      The host path is where Olares stores the data, and the mount path is the path inside the container. Olares provides three managed host path prefixes:

      - `/app/data`: App data directory. Data can be accessed across nodes and is not deleted when the app is uninstalled. Appears under `/Data/studio` in Files.
      - `/app/cache`: App cache directory. Data is stored in the node's local disk and is deleted when the app is uninstalled. Appears under `/Cache/<device-name>/studio` in Files.
      - `/app/Home`: User data directory. Mainly used for reading external user files. Data is not deleted.
        :::
        :::info Host path rules
      - The host path you enter *must* start with `/`.
      - Studio automatically prefixes the full path with the app name. If the app name is `test` and you set host path `/app/data/folder1`, the actual path becomes` /Data/studio/test/folder1` in Files.
        :::
   2. Click **Add** next to **Storage Volume**.
   3. Configure the database volume. This data is for high-frequency I/O and does not need to be saved permanently. Map it to `/app/cache` so it will be automatically deleted when the app is uninstalled.
      - **Host path**: Select `/app/cache`, then enter `/db`.
      - **Mount path**: Enter `/var/www/html/db`.
   4. Click **Submit**.
   5. Click **Add** again to add the logo volume. This is user-uploaded data that should be persistent and reusable, even if the app is reinstalled. Map it to `/app/data`.
      - **Host path**: Select `/app/data`, then enter `/logos`.
      - **Mount path**: Enter `/var/www/html/images/uploads/logos`
      ![Add storage volumes](/images/manual/olares/studio-add-storage-volume.png#bordered)

6. Click **Create** to generate the app project.


You can check Files to verify the mounted paths.
![Check mounted path in Files](/images/manual/olares/studio-check-mounted-path-in-files.png#bordered)

### Option 3: Deploy an app with GPU, Postgres or Redis support
For more complex apps, Studio can also configure GPU access and connect to Postgres or Redis databases.

If your app needs GPU, enable the **GPU** option under **Instance Specifications** and select the GPU vendor.
![Enable GPU](/images/manual/olares/studio-enable-GPU.png#bordered)

If your app needs Postgres or Redis, enable it under **Instance Specifications**.
![Enable databases](/images/manual/olares/studio-enable-databases.png#bordered)

When enabled, Studio provides dynamic variables. You must use these variables in the **Environment Variables** section for your app to connect to the database.
- **Postgres variables:**

| Variables    | Description           |
|--------------|-----------------------|
| $(PG_USER)   | PostgreSQL username   |
| $(PG_DBNAME) | Database name         |
| $(PG_PASS)   | Postgres Password     |
| $(PG_HOST)   | Postgres service host |
| $(PG_PORT)   | Postgres service port |

- **Redis variables:**

| Variables     | Description        |
|---------------|--------------------|
| $(REDIS_HOST) | Redis service host |
| $(REDIS_PORT) | Redis service port |
| $(REDIS_USER) | Redis username     |
| $(REDIS_PASS) | Redis password     |

## Step 3: Review the package files

After creation, Studio generates the package files for your app, and then automatically deploys the app.

You can click on files like `OlaresManifest.yaml` to review and make changes.
For example, to change the app's display name:

1. Click <span class="material-symbols-outlined">box_edit</span> in the top-right to open the editor.
2. Click `OlaresManifest.yaml` to view the content.
   ![Edit `OlaresManifest.yaml`](/images/manual/olares/studio-edit-olaresmanifest.png#bordered)

3. Change the `title` field under `entrance` and `metadata`. For example, change `test` to `speedtest`.
4. Click <span class="material-symbols-outlined">save</span> in the top-right to save changes.

## Step 4: Deploy and test the app
After making changes, you can reinstall the app to test it.

1. Click **Apply** in the top-right corner to install the app. You can check the status at the bottom, and click in the bottom-right corner to view the details.
   ![Check app status](/images/manual/olares/studio-check-app-status.png#bordered)

2. Watch the deployment details. The interface of this page is similar to Control Hub, where you can check status, events, and logs of the app just deployed. If details don't appear, refresh the page.
   ![App deployment details](/images/manual/olares/studio-app-deployment-details.png#bordered)

3. Open the app. You can:
   - Click **Preview** in Studio.
   - Launch it from the Launchpad.
      :::info
      Apps deployed from Studio include a `-dev` suffix in the title to distinguish them from Market installations.
      :::
   ![App with dev suffix](/images/manual/olares/studio-app-with-dev-suffix.png#bordered)

## Step 5: Modify and redeploy
If you need to modify the app, for example, to update the logo of the app:
1. Click <span class="material-symbols-outlined">box_edit</span> in the top-right to open the editor.
2. Click `OlaresManifest.yaml`, and replace the default icon image address under `entrance` and `metadata`.
3. Click <span class="material-symbols-outlined">save</span> in the top-right to save changes.
4. Click **Apply** again to reinstall with the updated package. When it finished, you can check out the app with a new logo.
   :::info
   If no changes are detected since the last deployment, clicking **Apply** will simply return to the app's status page without reinstalling.
   :::
![Change app icon](/images/manual/olares/studio-change-app-icon.png#bordered)

## Uninstall or delete the app
If you no longer need the app, you can remove it.
1. Click <span class="material-symbols-outlined">more_vert</span> in the top-right corner.
2. You can choose to:
   - **Uninstall**: Removes the running app from Olares, but keeps the project in Studio so you can continue editing the package.
   - **Delete**: Uninstalls the app and removes the project from Studio. This action is irreversible.

## Troubleshoot a deployment

### Cannot install the app
If installation fails, review the error at the bottom of the page and click **View** to expand details.

### Run into issues when the app is running
Once running, you can manage the app from its deployment details page in Studio. You can:
- Use the **Stop** and **Restart** controls to retry. This action can often resolve runtime issues like a frozen process.
- Check events or logs to investigate runtime errors. See [Export container logs for troubleshooting](../controlhub/manage-container.md#export-container-logs-for-troubleshooting) for details.