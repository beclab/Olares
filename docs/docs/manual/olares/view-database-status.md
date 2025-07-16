---
description: Access and monitor your Olares database status using Bytebase, managing PostgreSQL and MongoDB instances with detailed configuration guidance.
---
# View database status in Olares
You can use third-party applications to view database status in Olares.

For example, Bytebase is a multipurpose middleware for working with databases. This guide uses Bytebase to demonstrate how to access database in the middleware.

## Before you begin
Make sure that you have installed Bytebase from Market.

![bytebase](/images/how-to/olares/controlhub/middleware/07.jpg#bordered)

## Add a PostgreSQL instance

To add a PostgreSQL instance in Bytebase:

1. In **Bytebase**, click **Add Instance**, and select **PostgreSQL**.
2. Configure the instance:
    - **Instance Name**: `Olares` or others.
    - **Environment**: Choose either `PROD` or `TEST`.
    - **HOST**, **USERNAME**, **PASSWORD**: Fill according to the information in the Control Hub's **Middleware** section.

   ![alt text](/images/how-to/olares/controlhub/middleware/09.jpg#bordered)

3. Click **Create** to save the changes and connect the instance.

Now you should be able to view the details of the PostgreSQL instance you just added.

![alt text](/images/how-to/olares/controlhub/middleware/10.jpg#bordered)

## Add a MongoDB instance
To add a MongoDB instance in Bytebase:

1. In **Bytebase**, click **Add Instance**, and select **MongoDB**.
2. Configure the instance.
   - **Instance Name**: `Olares` or others.
   :::info NOTE
   Do not use duplicate instance names.
   :::
   - **Environment**: Choose either `PROD` or `TEST`.
   - **HOST**, **USERNAME**, **PASSWORD**: Fill according to the information in the Control Hub's **Middleware** section.

   ![bytebase1](/images/how-to/olares/controlhub/middleware/11.jpg#bordered)

3. Click **Create** to save and connect the instance.

Now you should be able to view the details of the MongoDB instance you just added.

![alt text](/images/how-to/olares/controlhub/middleware/12.jpg#bordered)

## Edit the database

In Bytebase, click **SQL Editor** in the upper right corner to go to the **Editor** page for further operations.

![bytebase1](/images/how-to/olares/controlhub/middleware/13.jpg#bordered)

