---
outline: [2, 3]
description: Set up JupyterHub on Olares to provide a multi-user Jupyter notebook environment for data science, research, and collaborative coding.
head:
  - - meta
    - name: keywords
      content: Olares, JupyterHub, Jupyter, notebook, data science, multi-user, self-hosted, Python
app_version: "1.0.3"
doc_version: "1.0"
doc_updated: "2026-04-13"
---

# Set up a multi-user notebook environment with JupyterHub

JupyterHub is an open-source, multi-user server for Jupyter notebooks. It lets you provide computational environments to multiple users, each with their own workspace, without requiring them to install anything locally. Running JupyterHub on Olares gives you a self-hosted notebook platform for data science, research, or team collaboration.

## Learning objectives

In this guide, you will learn how to:
- Install JupyterHub and set up the admin account
- Launch notebook servers and write code in JupyterLab
- Manage users as an admin

## Install JupyterHub

1. Open Market and search for "JupyterHub".

   When prompted, set environment variables:
   - **JUPYTERHUB_ADMIN_USERNAME**: Enter a username for the default admin account. You will use this to sign in later.

   <!-- ![JupyterHub in Market](/images/manual/use-cases/jupyterhub.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Set up the admin account

After installation, you need to create a password for the admin account.

1. Open JupyterHub from Launchpad. You will see the sign-in page.

   <!-- ![JupyterHub sign-in page](/images/manual/use-cases/jupyterhub-signin.png#bordered) -->

2. Click **Sign up** to go to the account creation page.

3. Enter the admin username you specified during installation and choose a password.

   <!-- ![JupyterHub sign-up page](/images/manual/use-cases/jupyterhub-signup.png#bordered) -->

4. Sign in with the admin credentials you just created.

## Use JupyterHub

After signing in, you will see the JupyterHub dashboard where you can manage servers and users.

<!-- ![JupyterHub dashboard](/images/manual/use-cases/jupyterhub-dashboard.png#bordered) -->

### Start a notebook server

1. From the dashboard, click **Start My Server**.

2. Select a notebook image. The default is `base-notebook`.

   <!-- ![Select notebook image](/images/manual/use-cases/jupyterhub-select-image.png#bordered) -->

3. Wait for the server to start. Once ready, you will be taken to the code platform.

### Configure resource limits

Each notebook server runs in its own container with pre-defined resource limits. Each notebook image profile defines its own CPU and memory allocation. The default profiles and their resource limits are:

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

   <!-- ![JupyterHub ConfigMap](/images/manual/use-cases/jupyterhub-configmap.png#bordered) -->

3. In the `data` section, find the `jupyterhub_config.py` content. Locate the `profile_list` entries, and modify the `cpu_guarantee`, `mem_guarantee`, `cpu_limit`, or `mem_limit` values for the desired profile.

4. Click **Confirm** to save the changes.

5. Return to **Deployments** > **jupyterhub**, and then click **Restart**.

Wait for the status icon to turn green, which indicates the new configuration has been loaded.

### Create a notebook

From the code platform, you can create new notebooks by selecting a kernel (for example, **Python 3**), open a terminal, or work with existing files.

<!-- ![Code platform](/images/manual/use-cases/jupyterhub-code-platform.png#bordered) -->

### Write code in JupyterLab

Click **View** > **JupyterLab** in the top navigation bar to switch to the full JupyterLab interface, which provides a more feature-rich coding environment with a file browser, multiple tabs, and extensions.

<!-- ![JupyterLab interface](/images/manual/use-cases/jupyterhub-lab.png#bordered) -->

<!-- ![JupyterLab notebook](/images/manual/use-cases/jupyterhub-lab-notebook.png#bordered) -->

### Return to the Hub

To go back to the JupyterHub dashboard from JupyterLab, click **File** > **Hub Control Panel**.

<!-- ![Hub Control Panel](/images/manual/use-cases/jupyterhub-control-panel.png#bordered) -->

## Manage users

As an admin, you can add and manage users from the JupyterHub Admin page.

### Add a new user

1. In the JupyterHub dashboard, navigate to the **Admin** page.

2. Click **Add Users** and enter the username for the new user.

   <!-- ![Add users](/images/manual/use-cases/jupyterhub-add-user.png#bordered) -->

3. The new user can now sign up with the username you created and set their own password.

### Authorize registered users

When a user signs up on their own (without being added by the admin first), they appear on a hidden authorization page. This page is not linked from the JupyterHub interface.

1. Append `/hub/authorize` to your JupyterHub URL in the browser address bar to access the authorization page, and confirm the authorization status.

   <!-- ![Authorize page](/images/manual/use-cases/jupyterhub-authorize.png#bordered) -->

2. Navigate to the **Admin** page and manually create the user there.

   Even after a user appears as authorized, they still cannot sign in and start a server until the admin creates them on the **Admin** page.

   <!-- ![Admin page - create user](/images/manual/use-cases/jupyterhub-admin-create-user.png#bordered) -->

## FAQ

### A user's server is stuck in "Pending" status

#### Cause

The cluster does not have enough CPU or memory resources to start the notebook server container.

#### Solution

- Check if other notebook servers can be stopped to free up resources.
- Adjust resource limits in the JupyterHub ConfigMap if the default allocation is too high for your cluster.

## Learn more

- [JupyterHub documentation](https://jupyterhub.readthedocs.io): Official JupyterHub docs and guides.
- [Zero to JupyterHub with Kubernetes](https://z2jh.jupyter.org): Comprehensive guide for JupyterHub on Kubernetes.
