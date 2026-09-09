---
outline: [2, 3]
description: Manage Olares apps in Market. Install and update system or community apps, add custom apps, and safely uninstall software.
head:
  - - meta
    - name: keywords
      content: Olares, Olares Market, install apps, update apps, uninstall apps, self-hosted app store, custom app chart
---

# Manage applications in Market

 Olares Market is an open and permissionless application platform. It provides one-click installation for a variety of applications and content recommendation algorithms from both Olares and third-party developers.

This guide helps users understand how to install, update, and uninstall applications through the Market. We'll also cover how to install custom applications.

## Find applications

When you open Market, the **Discover** page presents curated sections such as featured picks, community choices, top apps, and latest releases. You can also browse apps by category, including Creativity, Productivity, Fun, Lifestyle, Utilities, Developer Tools, and AI.

![Market](/images/manual/olares/market-discover1.png#bordered)

To find a specific app, open the **Manage** submenu on the left and click **Search**. Enter the keywords, and the relevant results will appear as you type.

![Search app](/images/manual/olares/search-app.png#bordered)

## Install applications

To install an application from Market:

1. Open Market from the Dock or Launchpad, and double-click the target application to view its details.
2. If the application supports multiple hardware accelerators, select your computing resource in the **RESOURCES** section, and make sure your hardware meets the listed **CPU**, **Memory**, **Required disk**, and **VRAM** requirements.

    ![Accelerator resources](/images/manual/olares/market-accelerator1.png#bordered)

3. Click **Get**, and then click **Install**.
4. If the app requires environment variables, set them in the pop-up that appears:

    ![Set environment variables](/images/manual/olares/set-app-env-var.jpeg#bordered)

    - Enter **custom variables** directly in the pop-up.
    - For **referenced system environment variables**, first set the value on the **Settings** > **Developer** > **System Environment Variables** page, then return and continue the installation.

5. When the installation finishes, the button changes to **Open**. Click it to launch the app.

The administrator installs a shared application in Market the same way as a regular app. Once installed, an entry appears on the Launchpad, and cluster members can open it directly without installing anything themselves. For details, see [Shared applications](shared-apps.md).

## Install models

Market offers two ways to run local large language models (LLMs):

- **Pre-built model apps**: Models packaged with a validated engine combination. Install them like any other app: find the app in Market, click **Get**, and then click **Install**.
- **Custom models with Engine Base**: Olares v1.12.6 introduced Engine Base apps, template apps built on inference engines such as Ollama, vLLM, SGLang, and llama.cpp. Clone the base app into an independent model instance, and configure the model source and parameters yourself.

Model instances are shared applications. In a multi-user cluster, only the administrator installs them, and all members can use them directly without installing anything themselves.

To create a model instance from an Engine Base app:

1. In Market, search for "Engine Base" and open the base app for your preferred engine.
2. Click **Create**, select the hardware accelerator, and give the instance a unique name.

    ![Create a model instance](/images/manual/olares/llm-base-apps-create-instance2.png#bordered)

3. Configure the model source and other environment variables, and complete the creation.

When the installation finishes, open the app to enter the model console. The model files download automatically, and the service exposes its API endpoints only after the model is downloaded and ready. On the **Status** tab, **Model** shows **Ready** when the instance is ready to serve client requests.

For the full configuration reference, see [Run local LLMs with Engine Base apps](/use-cases/llm-base-apps.md).

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

## Switch market source

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

## Resources

- [Clone applications](clone-apps.md): Run multiple instances of the same app.
- [Shared applications](shared-apps.md): Understand and manage cluster-wide shared apps.
- [Application concepts](../../../developer/concepts/application.md): Learn how system, community, and shared apps work.
