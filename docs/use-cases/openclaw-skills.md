---
outline: [2, 3]
description: Learn how to manage, install, and troubleshoot skills and plug-ins for OpenClaw.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, install skills, install plug-ins
---

# Manage skills and plugins

OpenClaw can be extended using skills and plugins:
- Skills add new capabilities to the AI. For example, managing Model Context Protocol servers.
- Plugins extend the system to support additional channels or community features. For example, adding iMessage via BlueBubbles.

:::info Why manual installation required
To protect your device, OpenClaw runs in a restricted, non-root environment without administrative privileges. This prevents the agent from modifying your system or self-installing software.
:::

## Understanding skills

Understanding where skills come from and how they are loaded helps you manage them effectively.

### Locations and precedence

Skills are loaded from three locations. If a skill with the same name exists in multiple locations, OpenClaw uses the one with the highest precedence, allowing you to easily customize or override built-in skills.

The order of precedence from highest to lowest is as follows:
1. Workspace skills (`Data/clawdbot/config/workspace/skills`): Per-agent skills that override all others.
2. Managed/local skills (`Data/clawdbot/config/skills`): Shared skills available to all agents on the same machine.
3. Bundled skills: Default skills shipped with your OpenClaw installation.

:::tip View all available skills
To see the complete list of skills available to your agent, including the bundled, shared, and workspace skills, run the `openclaw skills list` command in the OpenClaw CLI.
:::

### Compatibility on Olares

Not all skills can run in the Olares environment. OpenClaw actively blocks skills that cannot function correctly based on their declared requirements.

A skill might be blocked for the following reasons:
- Incompatible OS: The skill requires a different operating system (e.g., darwin for macOS), whereas Olares runs on Linux. For example, Apple ecosystem skills like apple-reminders cannot be used in Olares.
- Missing executables (`bins`): The environment lacks a required command-line tool, such as `gh` for managing GitHub issues.
- Missing configuration (`config`): A required setting in `openclaw.json` is not enabled.
- Missing environment variables (`env`): A required API key or authentication token has not been provided.

## Install skills

There are three ways to add new skills to your OpenClaw:
- Install skills via the `openclaw config` wizard. 
- Install skills from ClawHub, the package manager for OpenClaw.
- Install skills manually via local upload.

### Install via `openclaw config`

You can use the built-in configuration wizard to install default or officially supported skills. In this example, we will install the `clawhub` skill, which is required for the next method.

1. Open the OpenClaw CLI.
2. Enter the following command to start the wizard:

    ```bash
    openclaw config
    ```
3. Follow the prompts to configure your installation. Use the arrow keys to navigate and press **Enter** to confirm.

    | Settings | Option |
    |:---------|:-------|
    | Where will the Gateway run | Local (this machine) |
    | Select settings to configure | Skills |
    | Configure skills now | Yes |
    | Install missing skill dependencies | Navigate to the skill **clawhub**, press the **Space** key <br>to select it, and then press **Enter**. |
    | Preferred node manager for skill installs | npm<br>Wait for the message `Installed clawhub` to appear<br> before proceeding. |
    | Set [API_KEY] for [skill] | Select **No** for all these settings.| 

4. Finally, select **Continue** for **Select sections to configure**. The message `Configure complete` appears, indicating the setup is finished. 

### Install from ClawHub

Use the ClawHub CLI to search and install skills from [ClawHub](https://clawhub.ai/). Installing skills via ClawHub automatically handles the necessary package dependencies.

:::tip Prerequisite
Ensure that the [clawhub skill is installed](#install-via-openclaw-config). This skill enables the openclaw skills commands like `list`, `search`, and `install`, which allows you install more ClawHub-backed skills.
:::

1. Open the OpenClaw CLI.
2. To view the list of officially preset skills, run the following command:

    ```bash
    openclaw skills list
    ```
3. To search for a specific skill, use the `search` command.

    For example, to search for a calendar skill, run the following command:
    ```bash
    openclaw skills search Caldav Calendar
    ```

    The terminal returns the search results, displaying the skill ID at the beginning, followed by its description. In this case, the skill ID is `caldav-calendar`.

    ![Skill ID in clawhub](/images/manual/use-cases/openclaw-skill-id.png#bordered)
    
    :::warning Security recommendation
    It is highly recommended that you search for and read the details of a skill on the official ClawHub website before installing it. This ensures you are getting the correct skill and protects you from installing malicious packages.
    :::

4. Install the target skill using its skill ID.
    
    For example, to install this calendar skill, run the following command:

    ```bash
    openclaw skills install caldav-calendar
    ```

5. Wait for the terminal to indicate that the skill is installed, and then verify by running the following command:

    ```bash
    openclaw skills list
    ```

    The status of **caldav-calendar** is **ready**, indicating the installation is successful.

6. Open the **Control UI**, go to the **Skills** page, and then click the **Ready** tab. You will see the newly installed skill is enabled.

    ![Newly installed skill is enabled](/images/manual/use-cases/skill-enabled.png#bordered)

<!--2. Enter the following command:

    ```bash
    npx clawhub
    ```

3. When prompted to proceed, press Y.
4. Check the list of available skills by entering the following command:

    ```bash
    openclaw skills
    ```
    ![View skills](/images/manual/use-cases/available-skills.png#bordered)

5. Find the target skill name in the **Skill** column, and then install by entering the following command:

    ```bash
    npx clawhub install {skill_name}
    ```

    For example, to install mcporter, enter the following command:

    ```bash
    npx clawhub install mcporter
    ```

6. If prompted to **Install anyway**, select **Yes**.
7. When the installation is completed, verify by entering the following command:

    ```bash
    openclaw skills
    ```
    The status of **mcporter** is **ready**, indicating the installation is successful.

    ![Skill installed](/images/manual/use-cases/skill-installed.png#bordered)

8. Open the Control UI, go to the **Skills** page, click **BUILT-IN SKILLS** to expand the list, and then find **mcporter**:

    - If it is enabled, click **Disable**, and then click **Enable** again to force the system to save the configuration.
    - If it is disabled, click **Enable**.

    ![Enable skill](/images/manual/use-cases/enable-skill1.png#bordered)-->

### Upload skills

1. Open the Files app from the Launchpad, and then go to **Application** > **Data** > **clawdbot** > **config**.
2. Create a new folder named `skills`.
3. Upload your skill package, such as an extracted zip file, into this `skills` folder.
4. Install required package dependencies if any are missing.

## Install missing dependencies

If a skill is blocked or unusable, you need to identify and install its missing dependencies.

<!--Fix-via-Control-UI

1. Go to the **Skills** page in the Control UI, and then expand **BUILT-IN SKILLS** to see why the skill is unavailable and what dependency is missing.

    ![Identify missing dependency from Control UI](/images/manual/use-cases/identify-missing-dependency1.png#bordered)

2. Some missing dependencies can be installed directly from the Control UI by clicking the install prompt on the right.

    ![Install missing dependency from Control UI](/images/manual/use-cases/install-missing-package-ui1.png#bordered)
3. When the installation of missing components is completed, restart the OpenClaw container for the changes to take effect:

    a. Open Control Hub from Desktop.
    
    b. Click **clawdbot** under **Deployments**, and then click **Restart**.

4. Verify the installation:

    a. Open the Control UI.
    
    b. Go to the **Skills** page. The skill should now be tagged with **eligible**. 
    
    c. Configure required API keys if there is any, and then the agent will be able to use the skill.-->

1. Open the OpenClaw CLI and run the following command:

    ```bash
    openclaw skills check
    ```
    
    The terminal lists all unavailable skills and shows their missing requirements in parentheses.

    ![Install missing dependency from CLI](/images/manual/use-cases/missing-dependency-cli1.png#bordered){width=70%}

2. Use `npm` or `brew` to install the dependency manually. For detailed information about the installation requirements, see the `skills.md` file. 

    - Example: The `gh-issues` skill requires `gh` to be installed. 
    - Run the following command to install it:
        ```bash
        npm i -g gh
        ```
3. When the installation of missing components is completed, restart the OpenClaw container for the changes to take effect:

    a. Open Control Hub from Desktop.
    
    b. Click **clawdbot** under **Deployments**, and then click **Restart**.

4. Verify the installation:

    a. Open the Control UI.
    
    b. Go to the **Skills** page. The skill should now be tagged with **eligible**. 
    
    c. Configure required API keys if there is any, and then the agent will be able to use the skill.

## Install plug-ins

1. In the OpenClaw CLI, check the list of compatible plug-ins by entering the following command:

    ```bash
    openclaw plugins list
    ```

2. Find the target plug-in name in the **Name** column, and then install it by entering the following command:

    ```bash
    openclaw plugins install {Name}
    ```
    For example, to install BlueBubbles, enter the following command:

    ```bash
    openclaw plugins install @openclaw/bluebubbles
    ```

    :::warning Blocked plugin installation
    If your installation fails with a `Plugin "{Name}" installation blocked` error, you can bypass this security restriction by appending `--dangerously-force-unsafe-install` to your command. Only bypass this protection if you are certain the plugin is safe and comes from a trusted source.
    
    For example:
    ```bash
    openclaw plugins install @openclaw/nextcloud-talk --dangerously-force-unsafe-install
    ```
    :::

3. When the installation is completed, close OpenClaw CLI and open it again to load the new plug-in.

4. Verify by checking the plugin status:

    ```bash
    openclaw plugins list
    ```

    Now the status of the plug-in is **loaded**.

5. Open the Control UI, go to **Automation** > **Plugins**.
6. Find **@openclaw/bluebubbles** and click it to expand its panel:

    - If it is enabled, turn off the toggle switch, and then turn it on again to force the system to explicitly save the configuration.
    - If it is disabled, turn on the toggle switch.

    ![Toggle on plugin](/images/manual/use-cases/toggle-plugin1.png#bordered)

7. Click **Save** in the upper-right corner. The system validates the config and applies the change automatically.

    ::: tip Manual restart
    If you need to restart OpenClaw manually, do not use the OpenClaw CLI. Use one of the following methods:
    - **Restart the app from Settings or Market**: 
        - Open **Settings**, go to **Applications** > **OpenClaw**, click **Stop**, and then click **Resume**.
        - Open **Market**, go to **My Olares**, find **OpenClaw**, click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the operation button, select **Stop**, and then select **Resume**.
    - **Restart the container**: Open **Control Hub**, click `clawdbot` under **Deployments**, and then click **Restart**.
    :::