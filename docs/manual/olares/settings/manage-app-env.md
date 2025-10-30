---
outline: [2, 3]
description: Learn how to manage application environment variables in Olares
---

# Manage application environment variables

Application environment variables are key parameters used to configure the runtime behavior and settings of an application container. They determine how the application runs, which services it connects to, and what user credentials it uses.

To view or modify an application's environment variables:

1. Navigate to **Settings** > **Application**, 
2. Select the target application.
3. Enter the **Manage environment variables** page.

![Application environment variables](/images/manual/olares/manage-app-env.png#bordered)

## Application environment variable types

| Environment Variable Type                     | Description                                                                                                                                                        | Edit Permissions                                                                                                        |
|:----------------------------------------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------|:------------------------------------------------------------------------------------------------------------------------|
| **Referenced system environment variables**   | References system environment <br/>variables (e.g., **`OLARES_USER_USERNAME`**).<br/> When an application is declared to use <br/>system variables, you will see them here.       | **Read-only**. You must navigate to **Settings > Developer > System Environment Variables** page to make modifications. |
| **Application-specific variables**            | Configuration variables that are specific to<br/> the application itself.                                                                                               | **Editable**. You can modify their values directly on this page without needing to edit YAML files in Control Hub.      |


## Learn more 

[System environment variables](developer.md#set-system-environment-variables)