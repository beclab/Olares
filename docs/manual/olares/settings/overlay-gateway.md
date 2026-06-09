---
outline: [2, 3]
description: Enable Overlay Gateway in Olares Settings so supported applications can use the local network while keeping Olares platform networking.
---

# Manage Overlay Gateway

Overlay Gateway lets supported applications use an address on the same local network as your Olares device. This is useful for apps that rely on LAN discovery or direct LAN access, such as media casting, DLNA, UDP discovery, and smart home integrations.

When Overlay Gateway is enabled, the application can use the local network while still keeping the default Olares network for platform services, DNS, and communication with other Olares components.

## Before you begin

Overlay Gateway is available only when all of the following conditions are met:

- Olares is running on a native Linux host. Windows WSL environments are not supported.
- Your Olares device is connected through a wired network.
- The application supports Overlay Gateway.
- You are the Owner if you want to enable or disable the feature.

If your Olares device switches from a wired network to Wi-Fi after Overlay Gateway has been enabled, Olares continues to work normally, but Overlay Gateway does not take effect until the wired network is restored.

## Role permissions

All users can open **Settings** > **Network** and view **Overlay Gateway**. The available controls depend on your role:

- **Owner**: Can enable or disable Overlay Gateway at the system level and for each supported application.
- **Admin**: Can view the Overlay Gateway status.
- **Member**: Can view the Overlay Gateway status. Other network settings, such as Reverse Proxy and Hosts, are hidden.

## Enabling Overlay Gateway

1. Open **Settings**, then go to **Network** > **Overlay Gateway**.
2. Check the status at the top of the page.

   If Overlay Gateway is unavailable, the system toggle is disabled and the application list is hidden. The page shows why the feature is unavailable, such as an unsupported host environment or a non-wired network connection.

3. If you are the Owner, turn on the system-level Overlay Gateway toggle.
4. In the application list, turn on Overlay Gateway for each application that needs local network access.
5. In the confirmation dialog, click **Confirm**.

If no supported applications are installed, the page shows an empty state after the system-level toggle is enabled.

## Managing application status

Overlay Gateway is enabled separately for each supported application.

When you turn Overlay Gateway on or off for a running application, Olares restarts the application so the change can take effect. The application shows a loading state until it returns to **Running**.

If the application fails to restart, it changes to **Stopped**, but the Overlay Gateway setting is still saved. You can start the application again when you are ready.

If the application is already stopped when you change its Overlay Gateway setting, Olares saves the setting without restarting the application. The application remains **Stopped**.

## When Overlay Gateway becomes unavailable

If Olares detects that Overlay Gateway is no longer available, all application-level Overlay Gateway settings are automatically turned off. For example, this can happen when the device is no longer connected through a wired network or when the host environment does not support the feature.

After Overlay Gateway becomes available again, the Owner needs to enable it again at the system level and turn it on again for the required applications.
