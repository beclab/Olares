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
| [System application](../../developer/concepts/application.md#system-applications)                    | Built-in applications that come pre-installed with Olares,<br/> such as Profile, Files, and Vault.                                                                                                |
| [Community application](../../developer/concepts/application.md#community-applications)              | Applications that are created and maintained by third-party<br/> developers.                                                                                                                      |
| [Shared application](../../developer/concepts/application.md#cluster-scoped-applications) | A special category of community applications on Olares<br/> designed to provide unified, shared resources or services to all <br/>users within an Olares cluster. Only one <br/>instance is allowed per cluster. |
| [Reference application](../../developer/concepts/application.md#reference-applications)              | The applications that have been granted access to specific<br/> shared applications                                                                                                                    |
| [Dependencies](../../developer/concepts/application.md#dependencies)                                 | Prerequisite applications that must already be<br/> installed before a user can access an application <br/>that requires them.                                                                              |

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
* **Fun**: Self-hosted applications for entertainment and fun such as gaming, video streaming, and connecting with people. 
* **Utilities**: Tools for system management, file sharing, data backup, and running local AI models.
* **Developer Tools** Toolchain for the software development lifecycle, including code hosting, CI/CD, observability, and database management.

### Search using keywords 

To search an app in the market:

1. Open the Market app from the Dock or Launchpad.
2. In the **Manage** submenu on the left, click **Search**.
2. Enter the keywords. The relevant results will appear as you type.

![Search app](/images/manual/olares/search-app.png#bordered)

### Switch market source

You can switch market sources to speed up browsing, searching, and downloading, or to install apps exclusive to a particular source. To switch market sources:

1.  Open **Market**, and navigate to **My Olares** > **Settings** from the left sidebar.
2.  Under **Market sources**, click **Add source** to add a new app source. The current official sources include:
    * Global: `https://appstore-server-prod.bttcdn.com`
    * China: `https://appstore-china-server-prod.api.jointerminus.cn`
3.  Fill in the source name, URL, and description as required, then click **Confirm** to finish adding.
4.  In the source list, select the target source to activate it. Wait for about 10 minutes for the store page to switch.

::: tip Note
Applications from different installation sources will generate corresponding tabs in **My Olares** for easier application management.
:::

## Purchase paid applications

You can purchase paid applications directly from the Olares Market. The payment process requires coordination between your computer, the LarePass mobile app, and a cryptocurrency wallet (such as MetaMask).

### Prerequisites

Before you begin, ensure that you have completed the following tasks:
- **Cryptocurrency wallet**: You have installed and logged into a browser wallet extension (for example, MetaMask). To successfully complete the transaction, please ensure your wallet holds both of the following funds:
    - **Application fee**: Used to pay for the software license. The specific token required is determined by the app's `price.yaml` configuration (typically USDC or USDT).
    - **Network gas fee**: Used to pay for blockchain transaction costs. Note that gas fees must be paid in ETH.
- **LarePass mobile app**: You have installed the LarePass app on your phone and logged into the same account as your computer.

### Step 1: Initiate purchase and identity verification

1.  Open the Olares Market.
2.  Search for the target paid app, click **Get**, and then click **Buy**.
3.  In the **Verify payment access** popup, click **Verify now**.
4.  Open your LarePass mobile app:
    
    a. Review the request popup titled **Authorize payment access**.
    
    b. Tap **Confirm** .

    c. If the popup does not appear, return to the Market on your computer and click **Buy** again to re-initiate the request.

    ![Authorize payment access](/images/manual/olares/authorize-pay.png#bordered)

### Step 2: Confirm order

1.  After successful verification, return to the app page on your computer and click **Buy** again.
2.  The **Verify purchase request** prompt will appear on your computer. Click **Verify now**.
3.  Check your LarePass mobile app again:
    
    a. Review the request popup titled **Authorize purchase**.
    
    b. Tap **Confirm**.

    ![Authorize purchase](/images/manual/olares/authorize-purchase.png#bordered)

### Step 3: Complete payment via wallet

1. Return to the app page on your computer and click **Pay**.
2. The browser will automatically open **MetaMask** (or your connected wallet extension).
3. Review the amount in the wallet popup, and then confirm and approve the transaction.
4. Wait for the blockchain to confirm the transaction.
5. Once the payment is successfully completed, you can proceed to install the application by clicking **Install**.
6. After the payment is completed, you can verify your purchase credential from your LarePass mobile app by navigating to **Settings** > **Safety** > **VC cards**.

![Authorize purchase](/images/manual/olares/vc.png#bordered)


## Install applications

To install an application from Market:

1. Open Market from Dock or Launchpad.
2. Navigate to the app you want, and click **Get**.
3. When the operation button changes to **Install**, click it to start the installation.
4. Once finished, the button will change to **Open**.

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

## Clone applications

You can install multiple instances of the same application (also known as "clone"). This is useful for scenarios requiring different configurations or versions, such as running multiple Windows virtual machines.

To clone an application:

1. Confirm that the target application supports cloning:    
    a. Open **Market** from Dock or Launchpad, and click the application to open the details page.
    b. Check the **Required permissions** section. Only applications with the **Multi-Instance** property support cloning.
2. Click **My Olares** from the left sidebar. The list of installed applications will appear on the right.
3. Locate the target application, click the drop-down arrow next to the **Open** button, and then click **Clone**.
    ![Clone applications menu](/images/manual/olares/clone-apps.png#bordered)
4. In the **Clone App** window, configure the following information, and then click **Confirm**:
    
    a. Enter a unique name for the new instance to avoid confusion with existing applications. This name will serve as the app identifier and appear in the **Market** and the **My Olares** list.

    b. Set the display name for the shortcut icon generated on the desktop, or individually set the name for each component's icon if the application contains multiple entry points.
5. In the **Configure Environment Variables** window, complete the setup based on the application type, and then click **Confirm**. In the installed applications list, the newly created application will appear with a "Clone" tag next to it.
6. To view the detailed attributes of the newly cloned application, click the application.
7. To run the newly cloned application, click **Open**.

## Update applications
To update an application from Market:

1. Open Market from the Dock or Launchpad.
2. In the left sidebar, click **Updates** under the **Manage** section. If there are available updates, a notification badge will display.
3. The **Available updates** panel will display the applications with available updates.Click **Update all** to update all applications at once, or update each application individually.

## Uninstall applications

### Uninstall from Market

To uninstall an application from Market:

1. Open Market from Dock or Launchpad.
2. In the left sidebar, navigate to the **My Olares** section. Use the source tabs to filter and find your installed applications.
3. Click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the application's operation button, and select **Uninstall**.

### Uninstall from Launchpad

You can also uninstall an application from LaunchPad:

1. In Olares, click the Launchpad icon in the Dock to display all installed apps.
2. Click and hold the app icon until all the apps begin to jiggle.
3. Click <i class="material-symbols-outlined">cancel</i> on the app icon to uninstall it.

## Restore purchased applications

If you have previously purchased a paid application, you can restore and re-install it without additional payment, whether you simply uninstalled the app, reset your system, or migrated to a new Olares host.

### Re-install after uninstallation

If you are re-installing a paid app that was uninstalled from your current Olares system, the system preserves your license locally. 

You can find the app in the Market, click **Get**, and then click **Install** to re-install it.

### Restore after system reset or migration

If you have reset Olares OS or switched to new hardware, your local license data is cleared. You must verify your on-chain identity to retrieve your purchase rights, which will allow you to re-install the application.

1. Ensure that the current Olares OS is activated with the same **Olares ID** used for the original purchase.
2. Open the Olares Market.
3. Search for the target paid app, click **Get**, and then click **Buy**.

    :::tip Note
    The system will detect your on-chain purchase record and will not charge you again.
    :::

4. In the **Authorize payment access** popup on your computer, click **Verify now**.
    
    :::tip Note
    This step verifies your identity ownership and does not initiate a new transaction.
    :::

5. Once verified, the button will change to **Install**. Click it to start installation.

:::tip Important
- **Identity binding**: Your purchase history is bound to your **Olares ID (DID)**, not to specific hardware.
- **Identity recovery**: If you still have access to LarePass or have a backup of your mnemonic phrase, you can activate the new system with the same Olares ID and restore your purchased apps through identity verification.
- **Risk of loss**: If you have uninstalled LarePass and lost your mnemonic phrase, you will permanently lose your digital identity and all associated assets, including paid apps. In this case, you must create a new account and re-purchase the application.
- **Developer availability**: Restoring a license requires issuance via the developer's Merchant service. If the developer's Merchant service is offline or unavailable, you might be unable to restore the application.
:::

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
* **Shared application restrictions** (for Olares member): Install the reference app, and contact your Olares admin to install the corresponding shared application.

### Why can't I resume my application?

When you try to resume an application in Olares, if you receive an error message regarding insufficient CPU, memory, or disk, it means the system's current available resources cannot support running the application, and you need to close other applications to free up space.

#### Why was my application stopped?

An application is usually stopped due to one of the following reasons:
* **System auto stop**: To ensure Olares's stability, the Olares system monitors resource usage. If an application consumes excessive resources (such as CPU and memory) causing a high system load, the system might automatically pause it to prevent the entire device from freezing or crashing.
* **Manual stop**: You or an administrator might have manually stopped the application previously, and the application has not been resumed yet.

#### Why can't I resume my application now?

Starting an application requires reserving a specific amount of computing resources. If other running applications are already occupying most of the resources, the remaining "free space" is not enough for the application you want to start.

Therefore, when you try to resume the application, you might encounter the following messages:

| Error message | Description |
| :--- | :--- |
| Insufficient system CPU/memory | The physical resources of the entire system are nearly exhausted. |
| Insufficient disk space | The hard drive is full, and new data cannot be written. |
| Available CPU/memory insufficient | While there are some resources left, they are less than the minimum<br> amount (minimum request) required by this specific application. |

#### How to resume my application?

To resume your application, you need to free up some occupied resources:

1. Go to **Settings** > **Application** to view the applications that are currently **Running**.
2. Find applications that you do not need to use right now.
3. Stop each of the applications by clicking the app and clicking <i class="material-symbols-outlined">toggle_on</i> to toggle off **Running**.
4. After resources are freed, go back to your target application and click **Resume** again.
