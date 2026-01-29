---
outline: [2, 3]
description: Learn how to view and manage Redis data in Olares using CLI or Bytebase.
---
# View Redis data

Redis is available by default in Olares. This guide walks you through accessing it from Olares.

## Get connection information

Before connecting, obtain Redis connection details from the Control Hub.

1. Open Control Hub from Launchpad.
2. In the left navigation pane, go to Middleware and select **Redis**.
3. On the Details panel, record the following information:
    - **Host**: Used for Bytebase connection.
    - **Password**: Used for both CLI and Bytebase.

    ![Redis details](/public/images/developer/develop/middleware/mw-redis-details.png#bordered){width=60% style="margin-left:0"}

## Access via CLI

You can use the Olares terminal to access the database container.

1. In Control Hub, open the Olares terminal at the bottom of the left navigation pane.
2. Enter the Redis container. The container name is fixed.

    ```bash
    kubectl exec -it -n os-platform kvrocks-0 -- sh 
    ```
3. Connect to the Redis database:

    ```bash
    redis-cli -p 6666 -a <your password from control-hub>
    ```

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

### Create a Redis instance

1. Log in to Bytebase with your admin account.
2. In the left navigation pane, select **Instances**, then click **+ Add Instance**.
3. Choose **Redis** as the database type.
4. Fill in the connection fields using values from Control Hub:
    - **Host or Socket**: Enter the `Host` address and do not include the port.
    - **Port**: Keep the default, usually `6379`.
    - **Username**: Leave it empty.
    - **Password**: Enter the `Password`.
5. Click **Test Connection** to verify connectivity, then click **Create**.

Creating an instance in Bytebase does not create a new database. Once the instance is created, you can use the corresponding client tools to inspect and manage data.