---
description: Learn how to manage repositories, view downloaded images, and export system logs for troubleshooting.
---

# Developer resources

The **Developer** page in Olares **Settings** is designed for developers and advanced users to manage core system resources and diagnose issues. Key functions include:

* **Repository Management**
* **Image Management**
* **Export System Logs**
* **Set System Environment Variables**

## Repository management

**Repository management** is where you maintain the source repositories for Olares to download essential system images and other software packages. You can view existing repositories, add new ones, and manage endpoints to optimize Olares' package retrieval performance.

On the repository list page, you can view the name of the repository, number of related images, and image size for each repository.

![Repo management](/images/manual/olares/repo-management.png#bordered)

### Add a new repository

Follow these steps to add a new repository:

1. Navigate to **Settings** > **Developer** > **Repository management**. 
2. Click the **+ Add repository** button in the top-right corner. 
3. In the pop-up dialog, fill in the following information:
    * **Repository Name**: Enter a unique name for the repository, such as `docker.io` or `quay.io`.
    * **Starting endpoint**: Enter the initial URL for the repository.
4.  Click **Confirm** to complete the addition.

### Manage repository endpoints

You can reorder a repository's access endpoints to optimize its access speed and stability.

![Endpoint management](/images/manual/olares/repo-endpoint-management.png#bordered)

1.  On the **Repository management** page, click the <i class="material-symbols-outlined">table_edit</i> button to the right of the target repository.
2.  On the **Endpoint management** page, you can:
    * **Reorder**: Use the up and down arrows to sort the endpoints. Olares will prioritize the endpoints higher on the list.
    * **Delete**: Click the <i class="material-symbols-outlined">delete</i> button to delete an endpoint you no longer need.

## Image management

The **Image Management** page provides a comprehensive view of all downloaded and cached application and software package images on your Olares system.


![Image management](/images/manual/olares/image-management.png#bordered)

## Export system logs

Logs record the operational status of various system components. When troubleshooting Olares issues, system logs can provide crucial diagnostic information. To download system logs:

1.  Navigate to **Settings** > **Developer** > **Logs**.
2.  Click **Collect** to generate the log file. The log will automatically be saved to the `/Home/pod_logs` directory.
3.  Click **Open** to open the logs directory in a new window.

   ![Generate logs](/images/manual/olares/export-log.png#bordered)

4.  Right-click the generated log file and select **Download** to save it to your local machine.

   ![Download logs](/images/manual/download-logs.png#bordered){width=70%}
Once downloaded, you can attach the log file to a GitHub feedback post and share it with the Olares team to help them locate the issue faster.

## Set system environment variables

System environment variables, prefixed with `OLARES_SYSTEM_`, let you centrally manage important Olares configurations such as SMTP and API keys, making setup and maintenance simpler.

You can view and adjust these variables at any time to optimize network access and improve overall system performance.

:::tip Note

- System environment variables are predefined. You can edit their values, but you **cannot** add or delete variables or change their types.

- If a variable is marked as `required: true` and has no value, the system will prompt you to configure it before installing any application that depends on it.
::: 

### Automatic configuration

When you activate Olares, the system automatically configures certain environment variables based on your Olares ID. This helps Olares connect to the network region closest to you for faster downloads and smoother performance.

- If your Olares ID ends with `.cn` (for example, `user@olares.cn`), the variables will use `.cn` service URLs.

- If your Olares ID ends with `.com` (for example, `user@olares.com`), the variables will use `.com` service URLs.

Olares automatically sets the following variables to ensure optimal network access:

- `OLARES_SYSTEM_REMOTE_SERVICE` – Handles component interface calls.

- `OLARES_SYSTEM_CDN_SERVICE` – Manages downloads and content delivery.

These variables take effect automatically after activation, ensuring Olares always uses the fastest and most suitable network connections.

### Modify variables

To manually change system environment variables:

1.  Navigate to **Settings** > **Developer** > **System Environment Variables**.

2.  Locate the variable you want to modify.

3.  Click the <i class="material-symbols-outlined">edit_square</i> icon to change its value. You cannot edit grayed-out variables.

4.  Enter the new value and click **Confirm** to apply the changes.
    ![Set Sys Env Var](/images/manual/olares/sys-env-var.png#bordered)