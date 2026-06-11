---
outline: [2, 3]
description: Learn how to enable overlay gateway in Olares Settings so supported apps can get a dedicated local IP for LAN discovery and local access.
---
# Enable overlay gateway for applications

Overlay gateway assigns a dedicated local IP to supported apps through a virtual network interface. Use it for apps that need LAN discovery or direct local access, such as screen mirroring, DLNA, device discovery, or local media streaming.

Apps can be accessed through their local IP while keeping the default Olares cluster network for system services, DNS, and platform communication.

## Before you begin

Check the following before you configure overlay gateway:

- Your Olares device must run on a native Linux host. WSL is not supported.
- Your Olares device must use a wired Ethernet connection. If you switch to Wi-Fi after enabling overlay gateway, Olares continues to work, but overlay gateway does not take effect.
- Only apps that declare overlay gateway support are listed.
- Super admin permissions are required to turn the system-level overlay gateway service on or off. After the service is enabled, Admins and Members can enable overlay gateway for their own supported apps.

If the feature is unavailable, the system switch is disabled and the app list is hidden.

:::info Availability changes
If overlay gateway becomes unavailable, such as after switching from Ethernet to Wi-Fi, app-level settings are reset. When it becomes available again, enable overlay gateway again for the apps you need.
:::

## Access permissions

All users can open **Settings** > **Network**, but available options depend on their role.

| Role | Available actions |
| -- | -- |
| Super admin | Turn the system-level overlay gateway service on or off, and enable or disable <br>overlay gateway for supported apps. |
| Admin | View the overlay gateway status, and enable or disable overlay gateway for their<br> own supported apps after the service is enabled. |
| Member | View the overlay gateway status, and enable or disable overlay gateway for their<br> own supported apps after the service is enabled. |

## Enable overlay gateway

Overlay gateway has two levels:

- The system-level service, managed by the Super admin.
- The app-level switch, managed by each user for their own supported apps.

### Turn on the system-level service

If you are the Super admin:

1. Open **Settings** and go to **Network** > **Overlay gateway**.

    ![Overlay gateway](/images/manual/olares/settings-overlay-gateway.png#bordered){width=90%}

2. Check the feature status at the top of the page.
3. If the switch is available, turn on **Overlay gateway**.

After the system-level switch is enabled, Olares shows the list of supported apps available to the current user. If no supported app is installed, the page shows an empty state.

### Enable overlay gateway for an app

After the system-level service is on:

1. Open **Settings** and go to **Network** > **Overlay gateway**.
2. Under **Applications**, find the app you want to configure and turn on the switch.
3. In the confirmation dialog, click **Confirm**.

    ![Overlay gateway for an app](/images/manual/olares/settings-app-level-overlay-gateway.png#bordered){width=90%}

If the app is running, Olares restarts it so the network change can take effect. After the app restarts, you can access it at its local IP. The app shows a loading state until it returns to **Running**.

If the app is stopped, Olares saves the overlay gateway setting without starting the app. The app remains **Stopped**.

:::info
If a running app fails to restart, its status may change to **Stopped**, but the overlay gateway setting is still saved.
:::

## Disable overlay gateway

Disable overlay gateway at the level that matches your goal:

- Use the system-level switch to disable overlay gateway for all apps and users.
- Use the app-level switch to stop one app from using its local IP.

### Turn off the system-level service

If you are the Super admin:

1. Open **Settings** and go to **Network** > **Overlay gateway**.
2. Turn off **Overlay gateway**.

This disables overlay gateway for all apps and users.

### Disable overlay gateway for an app

1. Open **Settings** and go to **Network** > **Overlay gateway**.
2. Find the app you want to configure.
3. Turn off the app-level switch.
4. In the confirmation dialog, click **Confirm**.

If the app is running, Olares restarts it so the network change can take effect. If the app is stopped, Olares saves the setting without starting the app.
