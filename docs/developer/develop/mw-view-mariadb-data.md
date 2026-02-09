---
outline: [2, 3]
description: Learn how to connect to and manage MariaDB data in Olares using CLI or Bytebase.
---
# View MariaDB data

To use MariaDB in Olares, install it from Market first. This guide explains how to access and manage MariaDB data using CLI or Bytebase.

## Install MariaDB service

Before connecting, install the MariaDB service from Market.

1. Open Market from Launchpad and search for "MariaDB".
2. Click **Get**, then **Install**, and wait for the installation to complete.

Once installed, the service and its connection details will appear in the Middleware list in Control Hub.

## Get connection information

Before connecting, obtain MariaDB connection details from Control Hub.

1. Open Control Hub from Launchpad.
2. In the left navigation pane, go to Middleware and select **Mariadb**.
3. On the Details panel, record the following information:
   - **Host**: Used for Bytebase connection.
   - **User**: Used for Bytebase connection.
   - **Password**: Used for both CLI and Bytebase.

    ![MariaDB details](/images/developer/develop/middleware/mw-mariadb-details.png#bordered){width=60% style="margin-left:0"}

## Access via CLI

You can use the Olares terminal to access the MariaDB container for debugging or data operations.

1. In Control Hub, open the Olares terminal at the bottom of the left navigation pane.
2. Retrieve the Pod name for the middleware:

    ```bash
    kubectl get pods -n mariadb-middleware
    ```
3. Record the Pod name, then enter the container:

    ```bash
    kubectl exec -it -n mariadb-middleware <mariadb-pod> -- sh
    ```
4. Connect to MariaDB:

    ```bash
    mysql -u root -p
    ```
5. When prompted, enter the password you retrieved from Control Hub.

## Manage via Bytebase

Bytebase provides a graphical interface for database management and schema changes.

### Prerequisites

:::info MongoDB app required
Bytebase uses MongoDB to store its metadata. Install MongoDB before installing Bytebase.
:::

1. Open Market and search for "MongoDB".
2. Click **Get**, then **Install**, and wait until the service is running.
3. After MongoDB is installed, search for "Bytebase" in Market.
4. Click **Get**, then **Install**.

### First-time setup

When launching Bytebase for the first time, you must configure an administrator account.

:::tip 
Remember these credentials. Only the admin account can create new database instances.
:::

1. Open Bytebase from Launchpad.
2. Follow the on-screen prompts to set up your admin account with email and password.

### Create a MariaDB instance

1. Log in to Bytebase with your admin account.
2. In the left navigation pane, select **Instances**, then click **+ Add Instance**.
3. Choose **MariaDB** as the database type.
4. Fill in the connection fields using values from Control Hub:
    - **Host or Socket**: Enter the **Host** address and do not include the port.
    - **Port**: Keep the default, usually `3306`.
    - **Username**: Enter the **User** value from Control Hub.
    - **Password**: Enter the **Password** value from Control Hub.
5. Click **Test Connection** to verify connectivity, then click **Create**.

Creating an instance in Bytebase does not create a new database. Once the instance is created, you can use the corresponding client tools to inspect and manage data.