---
outline: [2, 3]
description: Learn how to manage repositories, view downloaded images, export system logs, and configure system-level environment variables.
head:
  - - meta
    - name: keywords
      content: Olares OS, advanced settings, repo management, image management, system logs, system-level environment variables
---

# Advanced settings

The **Advanced** page in **Settings** is designed for developers and advanced users to manage core system resources and diagnose issues. 

## Manage repositories

The **Repository management** page allows you to view the source repositories that Olares uses to download system images and software packages. You can also configure mirror endpoints to optimize download speeds and stability.

### View repositories

1. Go to **Settings** > **Advanced** > **Repository management**. 
2. In the repository list, you can view the name of the repository, number of related images, and image size for each repository.

    ![Repo management](/images/manual/olares/repo-management1.png#bordered){width=65%}

### Manage repository mirrors

Manage mirror endpoints for repositories to improve access speed and stability.

1. On the **Repository management** page, find the target repo, and then click <i class="material-symbols-outlined">chevron_forward</i> in the **Action** column.
2. On the **Mirror management** page, you can perform the following actions:

    ![Mirror management](/images/manual/olares/mirror-management.png#bordered){width=68%}
    
    - To re-order the mirror endpoints, click <i class="material-symbols-outlined">keyboard_control_key</i> or <i class="material-symbols-outlined">keyboard_arrow_down</i>. Olares prioritizes endpoints higher on the list.
    - To delete an endpoint you no longer need, click <i class="material-symbols-outlined">delete</i>.
    - To add a new mirror endpoint, click **Add mirror**, enter the mirror URL, and then click **Confirm**.

## Manage images

The **Image management** page provides a comprehensive view of all downloaded and cached application and software package images on your Olares system. You can filter or search to quickly find specific images.

![Image management](/images/manual/olares/image-management1.png#bordered){width=65%}

## Export system logs

System logs record the operational status of various system components and provide crucial diagnostic information for troubleshooting. 

To export system logs:
1. Go to **Settings** > **Advanced** > **Logs**.
2. Click **Collect** to generate the log file. The log is automatically saved to the `/Home/pod_logs` directory.
3. Click **Open** to open the logs directory in a new window.

   ![Generate logs](/images/manual/olares/export-log1.png#bordered){width=65%}

4.  Right-click the generated log file and select **Download** to save it to your local machine.

    :::tip
    You can attach this log file to a GitHub issue to help the Olares team locate and resolve your problem faster.
    :::

## Set system environment variables

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

    g. Click **Apply** at the bottom of **System environment varuables**.

3. To modify a variable:

    a. Find the target variable from the list, and then click <i class="material-symbols-outlined">edit_square</i>. Variables without this icon are locked and cannot be changed.

    b. In the **Edit environment variable** window, update the variable's value. 

    c. Click **Confirm** to save your changes.

    d. Click **Apply** at the bottom of **System environment varuables**.

4. To delete a variable:

    a. Locate a user-level variable and click <i class="material-symbols-outlined">delete</i>.

    b. Click **Confirm**.

    c. Click **Apply** at the bottom of **System environment varuables**.
