---
outline: [2, 3]
description: Learn how to manage application environment variables in Olares.
---

# Manage application environment variables

Application environment variables define how an application runs in its container, including its configuration, connected services, and other runtime settings. 

## View environment variables

1. Go to **Settings** > **Applications**.
2. Select the target application from the list.
3. Under **Environment variables**, click **Manage environment variables**.

![Application environment variables](/images/manual/olares/manage-app-env1.png#bordered){width=90%}

## Understand variable types

The **Manage environment variables** page shows the environment variables currently used by the application.

![Application environment variable list](/images/manual/olares/manage-app-env-var-list.png#bordered){width=90%}

You may see two types of variables depending on the application's configuration:

- **Application-specific variables**: Variables defined for the application itself. They display normally and can be edited directly on this page.
- **Referenced system environment variables**: System environment variables referenced by the application. These variables appear on this page only if the application uses them. They appear dimmed and cannot be edited directly. To modify them, see [Set system environment variables](/manual/olares/settings/developer.md#set-system-environment-variables)

## Edit an application-specific variable

To modify an application-specific variable:

1. Find the variable you want to edit.
2. Click <i class="material-symbols-outlined">edit_square</i> on the right.
3. In the dialog, update the value and click **Confirm**.
4. Click **Apply**.

![Edit application environment variables](/images/manual/olares/manage-app-env-edit-var.png#bordered){width=90%}

The application restarts automatically to apply the new configuration.

## Learn more

[System environment variables](developer.md#set-system-environment-variables)