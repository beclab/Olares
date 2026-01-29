---
outline: [2, 3]
description: Learn how to install MinIO and manage object storage in Olares using the MinIO Dashboard.
---
# View MinIO data

This guide explains how to install MinIO and manage object storage using MinIO Dashboard in Olares.

## Install MinIO service

Before using object storage, install the MinIO service from Market.

1. Open Market from the Launchpad and search for "MinIO".
2. Click **Get**, then **Install**, and wait for the installation to complete.

After installation, MinIO service and its connection details will appear in the Middleware list in Control Hub.

## Install MinIO Dashboard

MinIO Dashboard depends on the MinIO service and can only be installed after MinIO is available.

1. In Market, search for "MinIO Dashboard".
2. Click **Get**, then **Install**, and wait for the installation to complete.

## Get connection information

Before connecting, obtain MinIO connection details from the Control Hub.

1. Open Control Hub from Launchpad.
2. In the left navigation pane, go to Middleware and select **Minio**.
3. On the Details panel, record the following information:
    - **User**: Used for MinIO Dashboard connection.
    - **Password**: Used for MinIO Dashboard connection.

    ![MinIO details](/images/developer/develop/middleware/mw-minio-details.png#bordered){width=60% style="margin-left:0"}

## Manage via MinIO Dashboard

MinIO Dashboard provides a graphical interface for creating buckets, browsing files, and managing permissions.

1. Open MinIO Dashboard from the Launchpad.
2. On the login screen, enter the User and Password obtained from Control Hub.

Upon successful login, you can browse buckets and manage objects directly from the interface.