---
outline: [2, 3]
description: Manage Olares apps in Market. Install and update system or community apps, add custom apps, and safely uninstall software.
---

# Manage applications in Market

 Olares Market is an open and permissionless application platform. It provides one-click installation for a variety of applications and content recommendation algorithms from both Olares and third-party developers.

This guide helps users understand how to install, update, and uninstall applications through the Market. We'll also cover how to install custom applications.

## Before you begin

Before you start, it is recommended to familiarize yourself with a few concepts for Olares applications:

| Terminology | Description   |
|:------------|:--------------|
| [System application](../../../developer/concepts/application.md#system-applications)   | Built-in applications that come pre-installed with Olares,<br/> such as Profile, Files, and Vault. |
| [Community application](../../../developer/concepts/application.md#community-applications)  | Applications that are created and maintained by third-party<br/> developers.   |
| [Shared application](shared-apps.md) | A special type of community application, deployed centrally by the<br/> administrator, that provides shared resources or services to all users<br/> in a cluster. <br/><br/>Applications with a UI can be opened directly from the Launchpad.<br/> Headless backend services expose a standard API for client connections. |
| [Dependencies](../../../developer/concepts/application.md#dependencies) | Prerequisite applications that must already be<br/> installed before a user can access an application <br/>that requires them.  | 

## Find applications

The Olares Market offers various ways to discover and browse applications.

![Market](/images/manual/olares/market-discover1.png#bordered)

### Browse by categories

Upon launching the Market app, the **Discover** page serves as your central hub for exploration, organizing content into intuitive sections to guide your journey:
* **Discover Amazing Apps**: Featured applications curated by the editorial team, showcasing trending and seasonally relevant apps. Click these banners to access in-depth editorial features such as comprehensive guides, industry use cases, and detailed app comparisons to help you choose the right tools.
* **Community choices**: Most loved and recommended apps by the Olares community.
* **Top apps on Olares**: Apps with the highest usage and download rates.
* **Latest apps on Olares**: Recently added applications to the market.

You can also browse applications based on their functionality:
* **Creativity**: Apps for creating and publishing digital content, from AI-generated art and 3D models to blogs and design projects.
* **Productivity**: Apps for team collaboration, project management, data organization, and building custom AI-powered agents.
* **Fun**: Self-hosted applications for entertainment and fun such as gaming, video streaming, and connecting with people.
* **Lifestyle**: Self-hosted applications for managing your smart home, personal photo libraries, and AI identity.
* **Utilities**: Tools for system management, file sharing, data backup, and running local AI models. 
* **Developer Tools**: Toolchain for the software development lifecycle, including code hosting, CI/CD, observability, and database management.
* **AI**: Latest open-source LLMs and generative tools for text, audio, and 3D assets.

### Search using keywords 

To search an app in the market:

1. Open the Market app from the Dock or Launchpad.
2. In the **Manage** submenu on the left, click **Search**.
3. Enter the keywords. The relevant results will appear as you type.

    ![Search app](/images/manual/olares/search-app.png#bordered)

### Switch market source

You can switch market sources to speed up browsing, searching, and downloading, or to install apps exclusive to a particular source.

1.  Open **Market**, and navigate to **My Olares** > **Settings**.
2.  Under **Market sources**, click **Add source** to add a new app source. The current official sources include:
    * Global: `https://api.olares.com/market`
    * China: `https://api.olares.cn/market`
3.  Fill in the source name, URL, and description as required, then click **Confirm**.
4.  In the source list, select the target source to activate it. Wait for about 10 minutes for the store page to switch.

:::info
Applications from different installation sources will generate corresponding tabs in **My Olares** for easier application management.
:::

## Install applications

To install an application from Market:

1. Open Market from the Dock or Launchpad.
2. Find your target application, and double-click it to view its details.
3. If the application supports multiple hardware accelerators, configure your deployment in the **RESOURCES** section:

    ![Accelerator resources](/images/manual/olares/market-accelerator1.png#bordered)

    a. Select your preferred computing resource from the drop-down list. For example, **NVIDIA GPU**, **NVIDIA GB10**, or **CPU**.

    b. Review the **CPU**, **Memory**, **Required disk**, and **VRAM** requirements for the selected computing resource, and make sure your hardware meets them.

4. Click **Get**.
5. When the operation button changes to **Install**, click it.
6. If prompted, confirm your hardware accelerator choice. The installation starts.
7. (Optional) To cancel the installation, click <i class="material-symbols-outlined">close_small</i> on the right of the button.
8. When the installation finishes, the button changes to **Open**.

### Install shared applications

Shared applications are deployed centrally by the administrator, and cluster members do not need to install them themselves. To ensure a shared service is running and accessible within the cluster, follow the installation process based on the type of shared application.

::: info Manage shared applications
The administrator is responsible for upgrading, stopping, resuming, and uninstalling shared applications. These operations affect all members in the cluster, so please confirm before proceeding.
:::

#### Headless backend service

This type of shared application provides only API services without a graphical user interface. Any client that supports the corresponding API can invoke the service. Take a model instance created on **Ollama Engine Base** as an example:

1. **Administrator deploys the model**: The administrator creates a model instance from **Ollama Engine Base** in Market. Once deployed, the model instance starts as a shared service within the cluster, and a model application entry with the same name is generated on the Launchpad.
2. **Members configure and use it**:

    a. Get the access address: On the Launchpad, open the model application entry to enter the model console, and copy the **Base URL** displayed on the page.

    b. Configure the client: Install any third-party client that supports the corresponding API, such as LobeChat or Open WebUI, and enter the address above in the client's configuration settings to start using it.

#### Application with built-in UI

This type of shared application includes both a backend service and a web UI, and can provide services to users independently. Typical examples are Dify Shared and ComfyUI Shared.

1. **Administrator installs the application**: The administrator installs the shared application in Market. Once installed, the shared service starts within the cluster, and an application entry with the same name is added to the Launchpad.

    ::: tip ComfyUI Launcher
    ComfyUI Shared contains a desktop launcher component to manage ComfyUI services and related resources. The administrator needs to configure and start the service from the ComfyUI Launcher.
    :::
2. **Members use it directly**: Cluster members find the application entry on the Launchpad and click to open it directly, without installing any additional client.

### Install custom applications

To install a custom application:

1. Prepare an Olares Application Chart file (in `.zip`, `.tgz`, `.tar`, or `.gz` format).
2. Open **Market** from the Dock or Launchpad.
3. From the left sidebar, click **My Olares** > **Upload custom chart**, and select the chart file to install.

You can view all installed custom applications under the **My Olares** > **Upload** tab.

### Setting environment variables

During app installation, if an environment variable is required for the app but it either has no default value or its referenced system variable is unset, Market will display a settings pop-up:

![Set environment variables](/images/manual/olares/set-app-env-var.jpeg#bordered)

* **Custom variables**: Enter the value directly in the installation pop-up.
* **Referenced system variables**: You must first go to the **Settings > Developer > System Environment Variables** page to set the value for the corresponding variable.

After completing the environment variable setup, you can continue the installation.

## Update applications

To update an application from Market:

1. Open Market from the Dock or Launchpad.
2. In the left sidebar, click **Updates** under the **Manage** section. If there are available updates, a notification badge will display.
3. The **Available updates** panel will display the applications with available updates.Click **Update all** to update all applications at once, or update each application individually.

## Uninstall applications

Uninstall an application from Market or LaunchPad. 

<tabs>
<template #Uninstall-from-Market>

1. Open Market from Dock or Launchpad.
2. In the left sidebar, navigate to the **My Olares** section. Use the source tabs to filter and find your installed applications.
3. Click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the application's operation button, and select **Uninstall**.
4. In the **Uninstall** window, select the removal options as needed:

    - **Also remove all local data**

        - If you select this option, app data (in the Data directory) and cache data (in the Cache directory) will be permanently deleted and cannot be recovered.
        - If you do not select this option, app data (in the Data directory) will be retained and can be restored upon re-installation, while cache data (in the Cache directory) will be permanently deleted and cannot be recovered.
    - **Also uninstall the shared server (affects all users)**

        - If this is a shared application, select this option to remove it for all users in the cluster.
        - If you have uninstalled the user-facing app before removing the share app, you must re-install the user-facing app first, and then uninstall the shared application. 

5. Click **Confirm**.
</template>
<template #Uninstall-from-Launchpad>

1. In Olares, click the Launchpad icon in the Dock to display all installed apps.
2. Click and hold the app icon until all the apps begin to jiggle.
3. In the **Uninstall** window, select the removal option as needed:

    - **Also remove all local data**

        - If you select this option, app data (in the Data directory) and cache data (in the Cache directory) will be permanently deleted and cannot be recovered.
        - If you do not select this option, app data (in the Data directory) will be retained and can be restored upon re-installation, while cache data (in the Cache directory) will be permanently deleted and cannot be recovered.
    - **Also uninstall the shared server (affects all users)**

        - If this is a shared application, select this option to remove it for all users in the cluster.
        - If you have uninstalled the user-facing app before removing the share app, you must re-install the user-facing app first, and then uninstall the shared application. 

4. Click **Confirm**.
</template>
</tabs>

## View app operation logs

The application operation log details the processes and statuses of app operations within Market, such as installation, download, update, and uninstallation. To access these logs:
 
1. Open Market from Dock or Launchpad.
2. In the left sidebar, navigate to **My Olares**.
3. Click **Logs** in the top right corner. 

You can also click the <i class="material-symbols-outlined">download</i> button to download the logs.

## FAQs

### Why can't I install an application?

If you can't install an application, it might be due to:
* **Insufficient system resources**: Try freeing up system resources, or increasing your resource quota.
* **Missing dependencies**: Check the **Dependency** section on the application details page and make sure all required apps are installed.
* **Incompatible system version**: Try upgrading Olares to the latest version.
* **Shared service dependency** (for Olares members): The application requires a shared service to be running in the cluster. Contact your admin to install the shared application first before you can install it.

### Why was my application stopped?

An application is usually stopped due to one of the following reasons:
* **System auto stop**: To ensure Olares's stability, the Olares system monitors resource usage. If an application consumes excessive resources (such as CPU or memory) causing a high system load, the system might automatically pause it to prevent the entire device from freezing or crashing.
* **Manual stop**: You or an administrator might have manually stopped the application previously, and the application has not been resumed yet.

### Why can't I resume my application?

Starting an application requires reserving a specific amount of computing resources. If other running applications are already occupying most of the resources, the remaining free resources are not enough for the application you want to start.

Therefore, when you try to resume the application, you might encounter the following messages, and you need to stop other applications to free up resources.

| Error message | Description |
| :--- | :--- |
| Insufficient system CPU/memory | The physical resources of the entire system are nearly exhausted. |
| Insufficient disk space | The hard drive is full, and new data cannot be written. |
| Available CPU/memory insufficient | There are some resources left, they are less than the minimum<br> amount required by this specific application. |

### How to resume my application?

To resume your application, you need to free up some occupied resources:

1. Go to **Settings** > **Application** to view the applications that are currently **Running**.
2. Find applications that you do not need to use right now.
3. Stop each application by clicking the app and clicking **Stop**.
4. After resources are freed, go back to your target application and click **Resume** again.

### How to free up resources from unused apps?

If certain applications are not in use and you want to free up the system resources they are using, you can stop them from Market or Settings.

<tabs>
<template #Stop-apps-from-Market>

1. Open Market from Dock or Launchpad.
2. In the left sidebar, click **My Olares**. Use the source tabs to filter and find the app you want to stop.
3. Click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the application's operation button, and then select **Stop**.
</template>
<template #Stop-apps-from-Settings>

1. Go to **Settings** > **Applications**.
2. Click the target application you want to stop from the list.
3. Click **Stop**.

</template>
</tabs>

#### Free up shared application resources

To fully release resources for shared applications such as Ollama, the system-side service must be stopped. This action can only be performed by an administrator.

When the admin stops a shared application, the **Also stop the shared server (affects all users)** checkbox appears in the **Stop** window:
- To fully release the resources, this checkbox must be selected.
- Once selected, the service is disabled for the entire cluster, and other users will no longer be able to use the application.
- This checkbox only appears in multi-user clusters. In a single-user scenario, the system automatically stops both the user-facing application and the system-side service by default.

:::info Notes for Olares V1.12.4 and earlier
In Olares 1.12.4 and earlier versions, to fully release resources, you must use Market:
- Stop the app in Market and ensure that the **Also stop the shared server (affects all users)** option is selected.
- If the user-facing application has already been stopped through Settings, you must first resume it in Market, and then stop the shared application while ensuring the **Also stop the shared server (affects all users)** option is selected. 
:::

### What happens to my previously installed shared applications after upgrading to V1.12.6?

Olares 1.12.6 introduces a new V3 shared application architecture. Legacy V2 shared applications can still be started, stopped, paused, and resumed, but they cannot be upgraded directly to V3. To use the V3 version, uninstall the V2 app first, then install the V3 version. Existing data must be migrated manually.

For a full explanation of the architecture change and the migration workflow, see [Shared applications](shared-apps.md).
