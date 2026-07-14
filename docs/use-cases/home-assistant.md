---
outline: [2, 3]
description: Install Home Assistant on Olares to run your smart home hub without a separate Raspberry Pi or Docker setup. Connect your local network and add devices in a few clicks.
head:
  - - meta
    - name: keywords
      content: Olares, Home Assistant, home assistant on olares, home assistant docker, home assistant raspberry pi, smart home, home automation, HACS
app_version: "1.0.23"
doc_version: "1.0"
doc_updated: "2026-07-03"
---

# Build your smart home hub with Home Assistant

Home Assistant is an open-source home automation platform that brings together your smart home devices.

This guide walks you through installing Home Assistant on Olares, connecting it to your local network, and adding smart home devices, both automatically and manually.

## Learning objectives

In this guide, you will learn how to:

- Install Home Assistant from the Olares Market and set up your administrator account.
- Connect Home Assistant to your local network, and then discover and add nearby smart home devices.
- Manually add a device and monitor its live feed on your dashboard, using a Dahua IP camera as an example.

## Install Home Assistant

1. Open Market, and search for "Home Assistant".

   ![Install Home Assistant](/images/manual/use-cases/home-assistant.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Set up your account

Set up your local administrator account to begin using your Home Assistant dashboard.

1. Open Home Assistant from the Launchpad.
2. On the welcome screen, select your preferred language, and then select **Create my smart home**.

   ![Open Home Assistant](/images/manual/use-cases/home-assistant-welcome.png#bordered)

3. Enter the required profile information, including the username and password, and then click **Create account**.
4. Follow the remaining on-screen prompts to complete the setup.
5. Log in to the Home Assistant dashboard using your new credentials.

## Discover and add local devices

By default, Home Assistant runs isolated from your local network and cannot automatically discover smart home devices. Olares can expose Home Assistant to your local network through an overlay gateway, giving it a local IP address. Once Home Assistant is on the local network, it can discover DLNA media renderers, UPnP devices, HomeKit accessories, and other local services.

### Enable the overlay gateway

1. Open Olares Settings, and then go to **Network** > **Overlay gateway**.
2. Ensure the system-level **Enable overlay gateway** option is enabled. If you can't enable it, ask the Super Admin to enable it first.
3. Under **Applications**, find **Home Assistant**, and then enable overlay gateway for it.
4. In the confirmation dialog, click **Confirm**. Olares assigns a local IP address to Home Assistant.

   ![Enable overlay gateway for Home Assistant](/images/manual/use-cases/ha-enable-overlay-gateway.png#bordered){width=90%}

### Configure Home Assistant network adapters

After enabling the overlay gateway, configure Home Assistant to listen on the local network adapter.

1. Open Home Assistant, and then go to **Settings** > **System** > **Network** > **Network adapter**.
2. Clear the checkbox for **Autoconfigure**.
3. Select **Adapter: eth0** and **Adapter: net1**.

   ![Home Assistant network adapter settings](/images/manual/use-cases/home-assistant-network-adapter.png#bordered)

4. Click **Save**.
5. Go back to **Settings** > **System**, and then click <i class="material-symbols-outlined">power_settings_new</i> in the upper-right corner.
6. Select **Restart Home Assistant**, and then confirm to restart.

   After the restart, Home Assistant listens on your local network and can discover nearby devices.

### Add a discovered device

1. In Home Assistant, go to **Settings** > **Devices & services**.

   The **Discovered** panel appears, listing devices found on your local network. For example:
   - **DLNA Digital Media Renderer**: Smart TVs, set-top boxes, or speakers that can play media streams from Home Assistant.
   - **UPnP/IGD**: Routers or gateways that expose network information.
   - **Synology DSM**: NAS devices on the local network.
   - **HomeKit accessory**: HomeKit-compatible devices such as lights, plugs, or sensors.
   - **Internet Printing Protocol (IPP)**: network printers.

   ![Discovered devices](/images/manual/use-cases/ha-discovered2.png#bordered)

2. Click **Add** on the device you want to add, and then follow the on-screen prompts to finish pairing. Some devices might ask for a PIN code.
3. Assign the device to an area, and then click **Finish**. After pairing, you can find and control the device on your **Overview** dashboard.

## Add devices manually

You can also add a device manually, whether it doesn't appear in the **Discovered** panel or you prefer to set it up yourself. This section uses a Dahua IP camera as an example.

Before you start, make sure that:

- You have a Dahua IP camera powered on and connected to the same local network as your Olares device.
- You have the administrator username and password for the camera.

### Prepare your Dahua camera

To allow Home Assistant to discover and communicate with the camera, you must locate its network details via the Dahua web interface.

#### Obtain the camera IP address

1. Download the device discovery tool according to your operating system:

   - Windows: Go to the [Dahua Support website](https://support.dahuasecurity.com/en/toolsDownloadDetails?IsDpValue=Q93jdSLr94chjRuQ1y%2FcQQ%3D%3D) and download the **ConfigTool**.
   - macOS: Open the App Store and install the **CCTV Super Tool**. This guide uses the CCTV Super Tool.

2. Open CCTV Super Tool, and then click **Scan LAN**.
3. When prompted to allow the application to find devices on your local networks, select **Allow**.
4. Click **Scan LAN** again. The camera device is discovered and listed.

   ![Discover device](/images/manual/use-cases/home-assistant-discover-device.png#bordered)

5. Locate your camera in the discovered device list, and then note down the IP address. For example, `192.168.50.43`.

#### Get the camera ports

1. From the discovered device list, click your device, and then select **Open device website**.
2. Enter the username and password. The default is usually `admin` for both, which you must change on your first login.
3. Go to **Network** > **Port**.
4. Note down the HTTP port (usually `80`) and RTSP port (usually `554`).

### Add the camera to Home Assistant

Integrate your camera using one of the following methods:
- **Generic Camera (RTSP) integration** for a quick, basic video feed.
- **HACS integration** for deeper device control and advanced features.

<Tabs>
<template #RTSP-integration>

The Generic Camera integration uses the camera's Real-Time Streaming Protocol (RTSP) stream URL to display video.

#### Step 1: Add the Generic Camera integration

1. In Home Assistant, go to **Settings** > **Devices & services**.
2. Click **Add integration**.
3. Search for "Generic Camera" and select it.
4. In the **Stream source URL** field, construct and enter your RTSP address. Dahua cameras usually use the following format:

    ```
    rtsp://{username}:{password}@{camera_ip}:{rtsp_port}/cam/realmonitor?channel=1&subtype=1
    ```

    Where:
    - `username`: The camera's web interface login username.
    - `password`: The camera's web interface login password.
    - `camera_ip`: The camera's IP address.
    - `rtsp_port`: The RTSP port number of the camera (usually `554`).
    - `subtype=1`: The stream quality subtype. Use `0` for the main (high-resolution) stream, or `1` for the sub (low-resolution) stream.

    For example:

    ```
    rtsp://admin:12345Olares@192.168.50.43:554/cam/realmonitor?channel=1&subtype=1
    ```

   ![Generic Camera settings](/images/manual/use-cases/home-assistant-generic-camera.png#bordered)

5. Keep the remaining default settings, and then click **Submit**.
6. Wait for the preview to load.
7. After confirming the video feed works, click **Submit**.

#### Step 2: Add the camera to your dashboard

1. Select **Overview** from the left sidebar.
2. Click <i class="material-symbols-outlined">edit</i> in the upper-right corner.
3. From the **Favorite entities** list, select your camera device, and then click **Save**.

   The live camera feed now appears on your dashboard in the **Favorites** section.

   ![Generic Camera added to dashboard](/images/manual/use-cases/home-assistant-dashboard-fav.png#bordered)

4. Click the camera feed to open it in a separate window and view the real-time stream.

</template>
<template #HACS-integration>

The Home Assistant Community Store (HACS) allows you to download a dedicated, community-built Dahua integration for expanded functionality.

#### Step 1: Download the HACS plugin

1. Open your web browser and go to the [official GitHub repository of HACS](https://github.com/hacs/integration).
2. Select **Releases** on the right side of the page.
3. Locate the **Assets** section, and then download the latest `hacs.zip` file.
4. Extract the downloaded `.zip` file on your local computer to access the `hacs` folder.

#### Step 2: Upload HACS to Olares

Use Olares Files to place the plugin in the correct system folder so Home Assistant reads it.

1. Open the Files app from the Launchpad.
2. Go to **Application** > **Data** > **homeassistant**.
3. Create a new folder by clicking <i class="material-symbols-outlined">create_new_folder</i> in the upper-right corner.

   ![New folder for Home Assistant in Files](/images/manual/use-cases/home-assistant-new-folder.png#bordered)

4. Enter `custom_components` as the folder name, and then click **Create**.
5. Double-click the newly created folder **custom_components**.
6. Click <i class="material-symbols-outlined">drive_folder_upload</i> in the upper-right corner, select **Upload folder**, and then upload the extracted `hacs` folder from your local computer.

#### Step 3: Restart Home Assistant

Restart Home Assistant for it to detect the newly uploaded `custom_components` folder.

1. In Home Assistant, select **Settings** from the left sidebar, and then select **System**.
2. Click <i class="material-symbols-outlined">power_settings_new</i> in the upper-right corner.

   ![Restart Home Assistant](/images/manual/use-cases/home-assistant-restart.png#bordered)

3. Select **Restart Home Assistant**, and then click **Restart**. Wait for the restart to complete.

#### Step 4: Authorize and install HACS

1. Go back to **Settings**, and then select **Devices & services**.
2. Click **Add integration**.
3. Search for **HACS**, and then select it from the list.

   ![Add HACS to Home Assistant](/images/manual/use-cases/home-assistant-add-hacs.png#bordered)

4. Read the notices, select all the acknowledgment checkboxes, and then click **Submit**.
5. In the **Wait for device activation** window, copy the provided authorization key, and then click the displayed link, such as https://github.com/login/device.
6. Sign in with your GitHub account.
7. Paste the authorization key you copied, and then click **Authorize hacs**.
8. Return to Home Assistant. HACS now appears on the left sidebar and in the **Integrations** list.

   ![HACS added to Home Assistant](/images/manual/use-cases/home-assistant-hacs-added.png#bordered)

#### Step 5: Install and configure the Dahua integration

1. Select **HACS** from the left sidebar, and then search for "dahua".

   ![Search for product in HACS](/images/manual/use-cases/home-assistant-hacs-search.png#bordered)

2. Select the target device from the list, and then click **Download**.
3. Go to **Settings** > **System**, and then restart Home Assistant again to apply the new integration.
4. After restarting, select **Overview** from the left sidebar.
5. Click <i class="material-symbols-outlined">add</i> in the upper-right corner, and then select **Add device**.
6. Search for the brand name "Dahua", and then select it from the result.
7. In the **Add Dahua Camera** window, configure the device settings using the information you noted down earlier:

    - **Username**: Enter the camera's web interface login username.
    - **Password**: Enter the camera's web interface login password.
    - **Address**: Enter the camera's IP address, that is, `192.168.50.43`.
    - **Port**: Enter the HTTP port number, that is, `80`.
    - **RTSP Port**: Enter the RTSP port number, that is, `554`.

8. Keep the remaining settings as default, and then click **Submit**.
9. Specify a name for the device, and then click **Submit**.
10. Assign it to an area such as **Front door**, and then click **Finish**.

#### Step 6: Monitor your device on the dashboard

1. Select **Overview** from the left sidebar, and then locate your camera by its assigned area in the **Areas** section.
2. Click the area to view the live feed and device controls.

   ![Home Assistant dashboard, area section](/images/manual/use-cases/home-assistant-dashboard-areas.png#bordered)

</template>
</Tabs>
