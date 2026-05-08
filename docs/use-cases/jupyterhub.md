---
outline: [2, 3]
description: Set up JupyterHub on Olares to provide a multi-user Jupyter notebook environment for data science, research, and collaborative coding.
head:
  - - meta
    - name: keywords
      content: Olares, JupyterHub, Jupyter, notebook, data science, multi-user, self-hosted, Python
app_version: "1.0.5"
doc_version: "1.0"
doc_updated: "2026-05-08"
---

# Set up a multi-user notebook environment with JupyterHub

JupyterHub is an open-source, multi-user server for Jupyter notebooks. It lets you provide computational environments to multiple users, each with their own workspace, without requiring them to install anything locally. Running JupyterHub on Olares gives you a self-hosted notebook platform for data science, research, or team collaboration.

## Learning objectives

In this guide, you will learn how to:
- Install JupyterHub and set up the admin account.
- Launch notebook servers and write code in JupyterLab.
- Add and manage JupyterHub users.
- Customize notebook profiles as an admin.

## Install JupyterHub

1. Open Market and search for "JupyterHub".

   ![JupyterHub in Market](/images/manual/use-cases/jupyterhub.png#bordered){width=90%}

2. Click **Get**, then **Install**.

3. When prompted, set the following environment variable:

   - **JUPYTERHUB_ADMIN_USERNAME**: Enter a username for the default admin account. After installation, use this same username to sign up and create the admin password.
   
4. Wait for installation to complete.

## Set up the admin account

After installation, the admin username is reserved, but no password has been created yet. To activate the admin account, sign up with the same username you entered during installation.

1. Open JupyterHub from Launchpad. You will see the sign-in page.

   ![JupyterHub sign-in page](/images/manual/use-cases/jupyterhub-signin.png#bordered){width=40%}

2. Click **Sign up** to go to the account creation page.

3. Enter the admin username you specified during installation and set a password, then click **Create User**.    
   
   For example, if you set `JUPYTERHUB_ADMIN_USERNAME` to `olares`, enter `olares` as the username here.

   ![JupyterHub sign-up page](/images/manual/use-cases/jupyterhub-signup.png#bordered){width=40%}

4. Return to the sign-in page and sign in with the admin username and password. 

## Start and use a notebook server

After signing in, you will see the JupyterHub dashboard. From here, you can start your own notebook server and access the admin page.

![JupyterHub dashboard](/images/manual/use-cases/jupyterhub-dashboard.png#bordered){width=90%}

### Start a notebook server

1. From the dashboard, click **Start My Server**.

2. Select a notebook profile, then click **Start**. 

   The default option is **Base Environment**. It provides a minimal Python-only Jupyter environment and is suitable for most users.

   ![Select notebook profile](/images/manual/use-cases/jupyterhub-select-profile.png#bordered){width=90%}

   :::info
   When you start a notebook server for the first time with a selected profile, JupyterHub needs to pull the corresponding notebook image. This may take several minutes depending on the image size and your network connection.

   If you want to use a different notebook image, ask an admin to [customize the notebook profile](#optional-customize-notebook-profiles).
   :::

3. Wait for the server to start. After the server starts, the Jupyter Notebook interface opens.

### Create a notebook

From the Jupyter Notebook interface, you can create new notebooks, open a terminal, upload files, and work with existing files.

To create a notebook:

1. In the Jupyter Notebook interface, click **New**, then select **Notebook**.

   ![Jupyter Notebook interface](/images/manual/use-cases/jupyterhub-notebook-interface.png#bordered){width=90%}

2. Select a kernel, then click **Select**.

   ![Select kernel](/images/manual/use-cases/jupyterhub-new-notebook.png#bordered){width=90%}

A new notebook opens in the current workspace.

### Open JupyterLab and write code

You can also work in JupyterLab, which provides a more feature-rich coding environment with a file browser, multiple tabs, terminals, notebooks, and extensions.

1. In the Jupyter Notebook interface, click **View** > **Open JupyterLab** in the top navigation bar.

2. In the Launcher, click **Python 3 (ipykernel)** under **Notebook**.

   ![JupyterLab launcher](/images/manual/use-cases/jupyterhub-lab.png#bordered){width=90%}

3. After the notebook opens, enter your code in a cell, then click <i class="material-symbols-outlined">play_arrow</i> or press **Shift + Enter** to run it.

   ![Write code in JupyterLab](/images/manual/use-cases/jupyterhub-lab-notebook.png#bordered){width=90%}

### Return to the Hub

To go back to the JupyterHub dashboard from JupyterLab, click **File** > **Hub Control Panel**.

![Hub Control Panel](/images/manual/use-cases/jupyterhub-control-panel.png#bordered){width=90%}

## Manage users

As an admin, you can control who is allowed to use JupyterHub.

The recommended workflow is to create the username on the **Admin** page first. The user can then sign up with that exact username and set their own password.

### Add a new user

1. In the JupyterHub dashboard, go to **Admin** > **Add Users**.

2. Enter the username for the new user and click **Add Users**.

   ![Add users](/images/manual/use-cases/jupyterhub-add-user.png#bordered){width=90%}

3. Ask the user to open JupyterHub, click **Sign up**, and create a password with the exact username you added.

After signing in, the user can start their own notebook server.

### Add self-registered users

If a user signs up before the admin adds their username, they may appear on the hidden authorization page. However, appearing as authorized on this page does not automatically allow them to sign in and start a notebook server.

To allow the user to use JupyterHub:

1. In the browser address bar, append `/hub/authorize` to your JupyterHub URL.

   ![Authorize page](/images/manual/use-cases/jupyterhub-authorize.png#bordered){width=90%}

2. Check the username shown on the authorization page.

3. Return to the JupyterHub dashboard and go to **Admin** > **Add Users**.

4. Add the same username shown on the authorization page.

After the user is added on the **Admin** page, they can sign in with the account they registered.

![Admin page - create user](/images/manual/use-cases/jupyterhub-admin-create-user.png#bordered){width=90%}

## Optional: Customize notebook profiles

Each notebook profile defines the image and resource limits used when a user starts a notebook server. Most users do not need to change these settings.

As an admin, you can customize profiles from the JupyterHub ConfigMap.

:::warning GPU acceleration is not supported
JupyterHub on Olares currently does not support GPU acceleration for notebook servers. Use CPU-based notebook images only. Do not use CUDA-enabled image tags, such as tags prefixed with `cuda12-` or `cuda-`.
:::

The default profiles and their resource limits are:

| Profile | CPU (guarantee / limit) | Memory (guarantee / limit) |
|:--------|:------------------------|:---------------------------|
| Base Environment | 0.1 / 1 | 1 GB / 1 GB |
| Minimal Environment | 0.2 / 1 | 1 GB / 1 GB |
| Scientific Computing | 0.5 / 1 | 1 GB / 2 GB |
| Data Science (Python + R) | 1 / 1 | 2 GB / 2 GB |
| Deep Learning (TensorFlow) | 2 / 4 | 2 GB / 4 GB |
| Big Data (PySpark) | 2 / 4 | 4 GB / 8 GB |
| All Spark (Complete) | 2 / 4 | 4 GB / 8 GB |
| R Environment | 1 / 2 | 2 GB / 4 GB |

To customize a notebook profile:

1. In Control Hub, go to **Browse**, then select the JupyterHub project.

2. Under **Configmaps**, select `jupyterhub-config`, then click <i class="material-symbols-outlined">edit_square</i> in the top-right of the details panel to open the YAML editor.

   ![JupyterHub ConfigMap](/images/manual/use-cases/jupyterhub-configmap.png#bordered){width=90%}

3. In the YAML editor, find `c.KubeSpawner.profile_list`.

4. Locate the profile you want to modify, then update the values under `kubespawner_override`.

   - To change the notebook image, modify `image`. The value must be a full image address.
   - To change resource limits, modify `cpu_guarantee`, `mem_guarantee`, `cpu_limit`, or `mem_limit`.

   A profile entry contains fields like these:

   ```python
   'kubespawner_override': {
       'image': 'docker.io/beclab/jupyter-base-notebook:notebook-7.0.6',
       'cpu_guarantee': 0.1,
       'mem_guarantee': '1G',
       'cpu_limit': 1,
       'mem_limit': '1G',
   }
   ```

5. Click **Confirm** to save the changes.

6. Return to **Deployments** > **jupyterhub**, and then click **Restart** in the right panel.

Wait until the status icon turns green. The updated profile settings apply when users start a new notebook server for that profile.

## FAQ

### Why is my notebook server stuck in "Pending" status?

A notebook server may stay in **Pending** when the cluster does not have enough CPU or memory resources to start the server container.

As an admin, you can:

- Stop unused notebook servers to free up resources.
- If the selected notebook profile requires more resources than your cluster can provide, [customize the notebook profile](#optional-customize-notebook-profiles), then ask the user to start the server again.

### How do I use a different notebook image?

To use a different notebook image, an admin needs to update the `image` value in the JupyterHub ConfigMap. For detailed steps, see [Customize notebook profiles](#optional-customize-notebook-profiles).

Use CPU-based notebook images only, because GPU acceleration is not currently supported for notebook servers.

### Can I use GPU acceleration?

JupyterHub on Olares currently does not support GPU acceleration for notebook servers. The default notebook profiles use CPU-based images.

Do not use CUDA-enabled image tags, such as tags prefixed with `cuda12-` or `cuda-`. For more information, see [CUDA-enabled variants](https://jupyter-docker-stacks.readthedocs.io/en/latest/using/selecting.html#cuda-enabled-variants) in the Jupyter Docker Stacks documentation.

## Learn more

- [JupyterHub documentation](https://jupyterhub.readthedocs.io): Official JupyterHub docs and guides.
- [Zero to JupyterHub with Kubernetes](https://z2jh.jupyter.org): Comprehensive guide for JupyterHub on Kubernetes.