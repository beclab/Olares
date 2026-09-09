---
outline: [2, 3]
description: Learn how to manage source repositories and view downloaded images on the Advanced page in Settings.
head:
  - - meta
    - name: keywords
      content: Olares OS, advanced settings, repo management, image management
---

# Advanced settings

The **Advanced** page in **Settings** is designed for developers and advanced users to manage core system resources. To export system logs for troubleshooting, see [Get technical support](../../help/request-technical-support.md). To configure system-level environment variables, see [Set system environment variables](../../manage-system-env.md).

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

