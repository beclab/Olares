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
- Adjust notebook resource limits as an admin.

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

2. Select a notebook image, then click **Start**. 

   The default option is **Base Environment**. It uses the `base-notebook` image and is suitable for most users.

   ![Select notebook image](/images/manual/use-cases/jupyterhub-select-image.png#bordered){width=90%}

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

## Optional: Adjust notebook resource limits

Each notebook server runs in its own container with predefined CPU and memory limits. Most users do not need to change these settings.

As an admin, you can adjust the resource limits for each notebook profile from the JupyterHub ConfigMap. If you are not sure which values to use, keep the defaults. Increasing resource limits may cause notebook servers to stay in **Pending** if your Olares cluster does not have enough available CPU or memory.

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

To adjust resource limits for a profile:

1. In Control Hub, select the JupyterHub project from the Browse panel.

2. Under **Configmaps**, click `jupyterhub-config`, and then click <i class="material-symbols-outlined">edit_square</i> in the top-right to open the YAML editor.

   ![JupyterHub ConfigMap](/images/manual/use-cases/jupyterhub-configmap.png#bordered){width=90%}

3. In the `data` section, find the `jupyterhub_config.py` content. Locate the `profile_list` entries, and modify the `cpu_guarantee`, `mem_guarantee`, `cpu_limit`, or `mem_limit` values for the desired profile.

4. Click **Confirm** to save the changes.

5. Return to **Deployments** > **jupyterhub**, and then click **Restart** in the right panel.

Wait until the status icon turns green. The updated limits apply to new or restarted notebook servers.

## FAQ

### A user's server is stuck in "Pending" status

#### Cause

The cluster does not have enough CPU or memory resources to start the notebook server container.

#### Solution

- As a user, contact the JupyterHub admin.
- As an admin:
  - Stop unused notebook servers to free up resources.
  - If the selected notebook profile requires more resources than your cluster can provide, [adjust the notebook resource limits](#optional-adjust-notebook-resource-limits), then ask the user to start the server again.

## Learn more

- [JupyterHub documentation](https://jupyterhub.readthedocs.io): Official JupyterHub docs and guides.
- [Zero to JupyterHub with Kubernetes](https://z2jh.jupyter.org): Comprehensive guide for JupyterHub on Kubernetes.