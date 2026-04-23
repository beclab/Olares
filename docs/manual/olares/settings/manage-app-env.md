---
outline: [2, 3]
description: Learn how to manage application environment variables in Olares
---

# Manage application environment variables

Application environment variables define how an application runs in its container, including its configuration, connected services, and other runtime settings.

To view or modify an application's environment variables:

1. Go to **Settings** > **Application**.
2. Select the target application.
3. Under **Environment variables**, click **Manage environment variables**.

![Application environment variables](/images/manual/olares/manage-app-env1.png#bordered){width=90%}

## Application environment variable types

| Environment variable type | Description | Edit permissions |
|:--|:--|:--|
| **Referenced system environment variables**   | System environment variables<br> referenced by the application, such<br> as `OLARES_USER_USERNAME`.<br>If the application is configured to<br> use system environment variables, <br>they appear here. | **Read-only**. To modify them, go to **Settings** > **Advanced** > **System environment variables**. |
| **Application-specific variables** | Environment variables defined <br>specifically for the application. | **Editable**. You can modify them directly on this page without editing YAML files in Control Hub. |

## Learn more 

[System environment variables](developer.md#set-system-environment-variables)