---
description: Configure system-level environment variables in Olares to apply common settings globally for all applications, such as third-party API keys.
head:
  - - meta
    - name: keywords
      content: Olares, system environment variables, OLARES_USER, API keys, global settings
---
# Set system environment variables

Starting from Olares version 1.12.2, you can configure system-level environment variables for applications. This allows you to apply common settings globally without having to configure them for each application individually.

The variables are divided into two main categories:

| Category | Description | Permissions |
|:-------- |:----------- |:----------- |
| System config | Predefined variables required<br> for core system operations. <br><br>Example: `OLARES_SYSTEM_CDN_SERVICE` | You cannot delete them. The system locks and grays out some variables to ensure stability, but you can edit the values of others. |
| User information | Custom or predefined user-level<br> variables, such as third-party API keys.<br><br>Example: `OLARES_USER_CUSTOM_OPENAI_APIKEY`| You can add, edit, and delete them. The system automatically applies the `OLARES_USER_` prefix to any custom keys you create. |

To manage environment variables:

1. Go to **Settings** > **Advanced** > **System environment variables**.

    ![System system-level environment variables](/images/manual/olares/sys-env-var1.png#bordered){width=65%}

2. To add a new custom variable:

    a. Click **Add environment variables**.

    b. In the **Key** field,  enter your custom key name which is appended to the `OLARES_USER_` prefix.

    c. In the **Value** field, fill in the value.

    d. From the **Type** list, select the data type.

    e. Provide an optional description.

    f. Click **Save**.

    g. Click **Apply** at the bottom of **System environment variables**.

3. To modify a variable:

    a. Find the target variable from the list, and then click <i class="material-symbols-outlined">edit_square</i>. Variables without this icon are locked and cannot be changed.

    b. In the **Edit environment variable** window, update the variable's value.

    c. Click **Confirm** to save your changes.

    d. Click **Apply** at the bottom of **System environment variables**.

4. To delete a variable:

    a. Locate a user-level variable and click <i class="material-symbols-outlined">delete</i>.

    b. Click **Confirm**.

    c. Click **Apply** at the bottom of **System environment variables**.
