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

There are two primary ways to add new skills to your OpenClaw: 
- Install skills from ClawHub, the package manager for OpenClaw.
- Install skills manually via local upload.

### Install from ClawHub

Installing skills via ClawHub automatically handles the necessary package dependencies.

1. Open the OpenClaw CLI.
2. Enter the following command:

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

8. Open the Control UI, go to the **Skills** page and find **mcporter**:

    - If it is enabled, click **Disable**, and then click **Enable** again to force the system to save the configuration.
    - If it is disabled, click **Enable**.

    ![Enable skill](/images/manual/use-cases/enable-skill.png#bordered)

9. Click **Save** in the upper-right corner. The system validates the config and restarts automatically to apply the changes.

### Upload skills

1. Open the Files app from the Launchpad, and then go to **Application** > **Data** > **clawdbot** > **config**.
2. Create a new folder named `skills`.
3. Upload your skill package such as an extracted zip file into this `skills` folder.
4. Install required package dependencies if there is anyone missing.

## Install missing dependencies

If a skill is blocked or unusable, you need to identify and install its missing dependencies.

<Tabs>
<template #Fix-via-Control-UI>

1. Go to the **Skills** page in the Control UI, and then expand **BUILT-IN SKILLS** to see why the skill is unavailable and what dependency is missing.

    ![Identify missing dependency from Control UI](/images/manual/use-cases/identify-missing-dependency.png#bordered)

2. Some missing dependencies can be installed directly from the Control UI by clicking the install prompt on the right.

    ![Install missing dependency from Control UI](/images/manual/use-cases/install-missing-package-ui.png#bordered)
3. When the installation of missing components is completed, restart the OpenClaw container for the changes to take effect:

    a. Open Control Hub from Desktop.
    
    b. Click **clawdbot** under **Deployments**, and then click **Restart**.

4. Verify the installation:

    a. Open the Control UI.
    
    b. Go to the **Skills** page. The skill should now be tagged with **eligible**. 
    
    c. Configure required API keys if there is any, and then the agent will be able to use the skill.
</template>
<template #Fix-via-OpenClaw-CLI>

1. Open OpenClaw CLI and run the following command:
    
    ```bash
    openclaw skills info {skill_name}
    ```
    ![Install missing dependency from CLI](/images/manual/use-cases/missing-dependency-cli.png#bordered){width=70%}

2. Use `npm` or `brew` to install the dependency manually. For details information about the installation requirements, see the `skills.md` file. 

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
</template>
</Tabs>

## Install plug-ins

1. In the OpenClaw CLI, check the list of compatible plug-ins by entering the following command:

    ```bash
    openclaw plugins list
    ```

2. Find the target plug-in name in the **Name** column, and then install by entering the following command:

    ```bash
    openclaw plugins install {Name}
    ```
    For example, to install BlueBubbles, enter the following command:

    ```bash
    openclaw plugins install @openclaw/bluebubbles
    ```

3. When the installation is completed, close OpenClaw CLI and open it again to load the new plug-in.

4. Verify by checking the plugin status:

    ```bash
    openclaw plugins list
    ```

    Now the status of the plug-in is **loaded**.

5. Open the Control UI, go to **Config** > **Plugins**, and then find **@openclaw/bluebubbles** on the **All** tab:

    - If it is enabled, turn off the toggle switch, and then turn on again to force the system to explicitly save the configuration.
    - If it is disabled, turn on the toggle switch.

    ![Toggle on plugin](/images/manual/use-cases/toggle-plugin.png#bordered)

6. Click **Save** in the upper-right corner. The system validates the config and restarts automatically to apply the changes.
    ::: tip Manual restart
    If you need to restart OpenClaw manually, do not use the OpenClaw CLI. Use one of the following methods:
    - **Restart the app from Settings or Market**: 
        - Open **Settings**, go to **Applications** > **OpenClaw**, click **Stop**, and then click **Resume**.
        - Open **Market**, go to **My Olares**, find **OpenClaw**, click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the operation button, select **Stop**, and then select **Resume**.
    - **Restart the container**: Open **Control Hub**, click `clawdbot` under **Deployments**, and then click **Restart**.
    :::