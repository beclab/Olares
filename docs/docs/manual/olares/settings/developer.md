---
description: Learn how to manage repositories, view downloaded images, and export system logs for troubleshooting.
---

# Developer resources

The **Developer** page in Olares **Settings** is designed for developers and advanced users to manage core system resources and diagnose issues. Key functions include:

* **Repository Management**
* **Image Management**
* **Export System Logs**

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