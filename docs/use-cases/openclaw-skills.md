---
outline: [2, 3]
description: Learn how to manage, install, and troubleshoot skills and plugins for OpenClaw.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, install skills, install plugins
app_version: "1.0.36"
doc_version: "1.3"
doc_updated: "2026-09-04"
---

# Manage OpenClaw skills and plugins

OpenClaw can be extended using skills and plugins:
- Skills add new capabilities to the AI. For example, managing Model Context Protocol servers.
- Plugins extend the system to support additional channels or community features. For example, adding iMessage via BlueBubbles.

:::info Why manual installation required
To protect your device, OpenClaw runs in a restricted, non-root environment without administrative privileges. This prevents the agent from modifying your system or self-installing software.
:::

## Learning objectives

In this guide, you will learn how to:
- Understand where skills are loaded from and why some are blocked in Olares.
- Install skills using the configuration wizard, ClawHub, or manual upload.
- Identify and install missing dependencies for a skill.
- Install and enable plugins to extend OpenClaw with new channels or features.

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

You can use the built-in configuration wizard to install default or officially supported skills and their missing dependencies.

1. Open the OpenClaw CLI.
2. Enter the following command to start the wizard:

    ```bash
    openclaw config
    ```
3. Follow the prompts to configure your installation. Use the arrow keys to navigate and press **Enter** to confirm.

    | Settings | Option |
    |:---------|:-------|
    | Where will the Gateway run | Local (this machine) |
    | What do you want to configure | Skills | 
    | Install missing skill dependencies | Select as needed |

4. Wait for the installation to finish and review the installed summary.
5. When prompted **What do you want to configure**, select **Done**.

    The message `Configuration updated` appears, indicating the setup is finished.

### Install from ClawHub

Use the ClawHub CLI to search and install skills from [ClawHub](https://clawhub.ai/). Installing skills via ClawHub automatically handles the necessary package dependencies.

:::tip Prerequisite
Ensure that the [clawhub skill is installed](#install-via-openclaw-config). This skill enables the openclaw skills commands like `list`, `search`, and `install`, which allows you to install more ClawHub-backed skills.
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

4. Install the target skill using its skill ID.
    
    For example, to install this calendar skill, run the following command:

    ```bash
    openclaw skills install caldav-calendar
    ```

    The terminal displays a security audit report for the skill.

5. Review the security audit results, and then type `y` and press **Enter** to proceed.
6. Wait for the terminal to indicate that the skill is installed, and then verify by running the following command:

    ```bash
    openclaw skills list
    ```

    The status of **caldav-calendar** is **ready**, indicating the installation is successful.

7. Open the **Control UI**, click the user account in the lower left, and then select **Settings**.
8. Select **Agents** from the left sidebar, and then click the **Skills** tab.
9. Search the skill you just installed, and you will see it is enabled.

### Upload skills

1. Open the Files app from the Launchpad, and then go to **Data** > **clawdbot** > **config**.
2. Create a new folder named `skills`.
3. Upload your skill package, such as an extracted zip file, into this `skills` folder.
4. Install required package dependencies if any are missing.

## Install missing dependencies

If a skill is blocked or unusable, you need to identify and install its missing dependencies.

1. Open the OpenClaw CLI and run the following command:

    ```bash
    openclaw skills check
    ```
    
    The terminal lists all unavailable skills and shows their missing requirements in parentheses.

2. Use `npm` or `brew` to install the dependency manually. For detailed information about the installation requirements, see the `skills.md` file.

    - Example: The `gh-issues` skill requires `gh` to be installed. 
    - Run the following command to install it:
        ```bash
        npm i -g gh
        ```
3. When the installation of missing components is completed, restart the OpenClaw container for the changes to take effect:

    a. Open Control Hub from the Launchpad.
    
    b. Click **clawdbot** under **Deployments**, and then click **Restart**.

4. Verify the installation:

    a. Open the Control UI, click the user account in the lower left, and then select **Settings**.
    
    b. Select **Agents** from the left sidebar, and then click the **Skills** tab. 
    
    c. Search the skill. It should now be tagged with **eligible**. 
    
    d. Configure required API keys if any are required, and then the agent will be able to use the skill.

## Install plugins

1. In the OpenClaw CLI, check the list of compatible plugins by entering the following command:

    ```bash
    openclaw plugins list
    ```

2. Find the target plugin name in the **Name** column, and then install it by entering the following command:

    ```bash
    openclaw plugins install {Name}
    ```
    For example, to install LLM Task, enter the following command:

    ```bash
    openclaw plugins install @openclaw/llm-task
    ```

    :::warning Blocked plugin installation
    If your installation fails with a `Plugin "{Name}" installation blocked` error, you can bypass this security restriction by appending `--dangerously-force-unsafe-install` to your command. Only bypass this protection if you are certain the plugin is safe and comes from a trusted source.
    
    For example:
    ```bash
    openclaw plugins install @openclaw/nextcloud-talk --dangerously-force-unsafe-install
    ```
    :::

3. When the installation is completed, restart the gateway to load the new plugin.
4. Verify by checking the plugin status:

    ```bash
    openclaw plugins list
    ```

    Now the status of the plugin is **enabled**.

5. Open the Control UI, click the user account in the lower left, and then select **Settings**.
6. Select **Automation** from the left sidebar, and then click the **Plugins** tab.
7. Find **LLM Task** and click it to expand its panel:

    - If it is enabled, turn off the toggle switch, and then turn it on again to force the system to explicitly save the configuration.
    - If it is disabled, turn on the toggle switch.

    The system validates the config and applies the change automatically.

    :::tip Manual restart
    If you need to restart OpenClaw manually, do not use the OpenClaw CLI. Use one of the following methods:
    - **Restart the app from Settings or Market**:
        - Open **Settings**, go to **Applications** > **OpenClaw**, click **Stop**, and then click **Resume**.
        - Open **Market**, go to **My Olares**, find **OpenClaw**, click <i class="material-symbols-outlined">keyboard_arrow_down</i> next to the operation button, select **Stop**, and then select **Resume**.
    - **Restart the container**: Open **Control Hub**, click `clawdbot` under **Deployments**, and then click **Restart**.
    :::
