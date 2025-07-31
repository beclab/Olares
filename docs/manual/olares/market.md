---
outline: [2, 3]
description: Complete guide to managing Olares applications - install from Market, update system and community apps, handle custom installations, and properly uninstall applications.
---

# Manage applications in Market

 Olares Market is an open and permissionless application platform. It provides one-click installation for a variety of applications and content recommendation algorithms from both Olares and third-party developers.

This guide helps users understand how to install, update, and uninstall applications through the Market. We'll also cover how to install custom applications.

## Before you begin
Before you start, it is recommended to familiarize yourself with a few concepts for Olares applications:

| Terminology                                                                             | Description                                                                                                                                                                                       |
|-----------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [System application](../concepts/application.md#system-applications)                    | Built-in applications that come pre-installed with Olares,<br/> such as Profile, Files, and Vault.                                                                                                |
| [Community application](../concepts/application.md#community-applications)              | Applications that are created and maintained by third-party<br/> developers.                                                                                                                      |
| [Shared application](../concepts/application.md#cluster-scoped-applications) | A special category of community applications on Olares<br/> designed to provide unified, shared resources or services to all <br/>users within an Olares cluster. Only one <br/>instance is allowed per cluster. |
| [Reference application](../concepts/application.md#reference-applications)              | The applications that have been granted access to specific<br/> shared applications                                                                                                                    |
| [Dependencies](../concepts/application.md#dependencies)                                 | Prerequisite applications that must already be<br/> installed before a user can access an application <br/>that requires them.                                                                              |

## Find applications
The Olares Market offers various ways to discover and browse applications.

![Market](/images/manual/olares/market-discover.png#bordered)


### Browse by categories

On **Discover** page:
* **Featured Applications**: Curated by the editorial team, showcasing trending and seasonally relevant apps.
* **Community choices**: Most loved and recommended apps by the Olares community.
* **Top apps**: Apps with the highest usage and download rates.
* **Latest apps**: Recently added applications to the market.

You can also browse applications based on their functionality:
* **Creativity**: Apps for creating and publishing digital content, from AI-generated art and 3D models to blogs and design projects.
* **Productivity**: Apps for team collaboration, project management, data organization, and building custom AI-powered agents.
* **Lifestyle**: Self-hosted applications for managing your smart home, personal photo libraries, and AI identity.
* **Fun**: Selfhosted applications for entertainment and fun such as gaming, video streaming, and connecting with people. 
* **Utilities**: Tools for system management, file sharing, data backup, and running local AI models.
* **Developer Tools** Toolchain for the software development lifecycle, including code hosting, CI/CD, observability, and database management.

### Search using keywords 

To search an app in the market:

1. Open the Market app from the Dock or Launchpad.
2. In the **Manage** sub-memu on the left, click **Search**.
2. Enter the keywords. The relevant results will appear as you type.

![Search app](/images/manual/olares/search-app.png#bordered)

## Install applications

1. Open the Market app from the Dock or Launchpad.
2. Navigate to the app you want, and click **Get**.
3. When the operation button changes to "**Install**", click it to start the installation.
4. Once finished, the button will change to "**Open**".

:::tip
To cancel an installation, hover over the operation button and click **Cancel** when it appears.
:::

### Install shared and reference applications

To ensure a shared service is running and accessible within the cluster, follow this general installation process based on the type of Shared App:

* **Headless backend service**:
    This type of shared applications typically require third-party reference applications to access its service. Take Ollama for example:
    1. The administrator installs the shared application first. This makes the core service available in the cluster.
    
    2. Members (including the administrator) install the corresponding reference application (e.g., Open WebUI or LobeChat) to access the Ollama service.

* **Complete application with built-in UI**:
    This type of shared applications can provide service to itself. Typical examples are Dify Shared and ComfyUI Shared.
    
    1. The administrator installs the shared application first. This not only launches the shared service for the cluster, but also installs the client-side interface as the reference application.
    
       ::: tip ComfyUI Launcher
       ComfyUI Shared contains a web launcher component to facilitate the management of related services and resources. The administrator needs to configure and start the service from the ComfyUI Launcher.
       :::

    2. Other members in the cluster install the same application. For these users, only the access point to the shared application is installed.

### Install custom applications

1. Prepare an Olares Application Chart file (in `.zip`, `.tgz`, `.tar`, or `.gz` format).
2. Open the Market app from the Dock or Launchpad.
3. Click **My Olares** > **Local** to see all custom applications.
4. Click **Upload custom chart** in the top right corner and select the chart file to install.

## Update applications
1. Open the Market app from the Dock or Launchpad.
2. In the left sidebar, click **Updates** under the **Manage** section. If there are available updates, a notification badge will display.
3. The **Available updates** panel will display the applications with available updates.Click **Update all** to update all applications at once, or update each application individually.

## Uninstall applications

### Uninstall from Market
1. Open the Market app from the Dock or Launchpad.
2. In the left sidebar, navigate to the **My Olares** section. Use the **Olares** or **Local** tabs to filter and fined your installed applications.
3. Click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the application's operation button, and select **Uninstall**.

### Uninstall from Launchpad
1. In Olares, click Launchpad icon in the Dock to display all installed apps.
2. Click and hold the app icon until all the apps begin to jiggle.
3. Click <i class="material-symbols-outlined">cancel</i> on the app icon to uninstall it.

## View app operation logs

The application operation log details the processes and statuses of app operations within Market, such as installation, download, update, and uninstallation. To access these logs:
 
1. Open the Market app from the Dock or Launchpad.
2. In the left sidebar, navigate to **My Olares**.
3. Click **Logs** in the top right corner. 

You can also click the <i class="material-symbols-outlined">download</i> button to download the logs.

## FAQ

### Why can't I install an application?
If you can't install an application, it might be due to:
* **Insufficient system resources**: Try freeing up system resources, or increasing your resource quota.
* **Missing dependencies**: Check the **Dependency** section on the application details page and make sure all required apps are installed.
* **Incompatible system version**: Try upgrading Olares to the latest version.
* **Shared application restrictions** (for Olares member): Install the reference app, and contact your Olares admin to install the corresponding shared application.
