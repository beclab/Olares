---
outline: [2, 3]
description: Learn how to view and manage MongoDB data in Olares using CLI or Bytebase.
---
# View MongoDB data

To use MongoDB in Olares, install it from Market first. This guide walks you through the installation steps and shows how to access it from Olares.

## Install MongoDB service

Before connecting, install the MongoDB service from Market.
1. Open Market from the Launchpad and search for "MongoDB".
2. Click **Get**, then **Install**, and wait for the installation to complete.

Once installed, the service and its connection details will appear in the Middleware list in Control Hub.

## Get connection information

Before connecting, obtain MongoDB connection details from the Control Hub.

1. Open Control Hub from Launchpad.
2. In the left navigation pane, go to Middleware and select **Mongodb**.
3. On the Details panel, record the following information:
    - **Mongos**: The host address provided by Control Hub. Used for Bytebase connection.
    - **User**: Used for Bytebase connection.
    - **Password**: Used for both CLI and Bytebase.

    ![MongoDB details](/images/developer/develop/middleware/mw-mongodb-details.png#bordered){width=60% style="margin-left:0"}

## Access via CLI

You can use the Olares terminal to access the MongoDB container.

1. In Control Hub, open the Olares terminal at the bottom of the left navigation pane.
2. Retrieve the Pod name for the middleware:

    ```bash
    kubectl get pods -n os-platform | grep tapr-middleware
    ```
3. Record the Pod name that starts with `tapr-middleware`, then enter the container:

    ```bash
    kubectl exec -it -n os-platform <tapr-middleware-pod> -- sh
    ```
4. Connect to MongoDB using `mongosh`:

    ```bash
    mongosh "mongodb://root:<your password from controlhub>@mongodb-mongodb-headless.mongodb-middleware:27017"
    ```

## Manage via Bytebase

Bytebase provides a graphical interface for database management and schema changes.

### Install Bytebase

1. Open Market and search for "Bytebase".
2. Click **Get**, then **Install**.

### First-time setup

When launching Bytebase for the first time, you must configure an administrator account.

:::tip 
Remember these credentials. Only the admin account can create new database instances.
:::

1. Open Bytebase from Launchpad.
2. Follow the on-screen prompts to set up your admin account with email and password.

### Create a MongoDB instance

1. Log in to Bytebase with your admin account.
2. In the left navigation pane, select **Instances**, then click **+ Add Instance**.
3. Choose **MongoDB** as the database type.
4. Fill in the connection fields using values from Control Hub:
    - **Host or Socket**: Enter the `Mongos` host address and do not include the port.
    - **Port**: Keep the default, usually `27017`.
    - **Username**: Enter the `User`.
    - **Password**: Enter the `Password`.
5. Click **Test Connection** to verify connectivity, then click **Create**.

Creating an instance in Bytebase does not create a new database. Once the instance is created, you can use the corresponding client tools to inspect and manage data.