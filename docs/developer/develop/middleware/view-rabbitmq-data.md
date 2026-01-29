---
outline: [2, 3]
description: Learn how to install RabbitMQ and manage RabbitMQ resources in Olares using RabbitMQ Dashboard.
---
# View RabbitMQ data

This guide explains how to install RabbitMQ and manage data using RabbitMQ Dashboard in Olares.

## Install RabbitMQ service

Before using RabbitMQ, install the RabbitMQ service from Market.

1. Open Market from the Launchpad and search for "RabbitMQ".
2. Click **Get**, then **Install**, and wait for the installation to complete.

After installation, RabbitMQ service and its connection details will appear in the Middleware list in Control Hub.

## Install RabbitMQ Dashboard

RabbitMQ Dashboard depends on the RabbitMQ service and can only be installed after RabbitMQ is available.

1. In Market, search for "RabbitMQ Dashboard".
2. Click **Get**, then **Install**, and wait for the installation to complete.

## Get connection information

Before connecting, obtain RabbitMQ connection details from the Control Hub.

1. Open Control Hub from Launchpad.
2. In the left navigation pane, go to Middleware and select **Rabbitmq**.
3. On the Details panel, record the following information:
    - **User**: Used for RabbitMQ Dashboard connection.
    - **Password**: Used for RabbitMQ Dashboard connection.

    ![RabbitMQ details](/images/developer/develop/middleware/mw-rabbitmq-details.png#bordered){width=60% style="margin-left:0"}

## Manage via RabbitMQ Dashboard

RabbitMQ Dashboard provides a graphical interface for viewing and managing RabbitMQ resources.

1. Open RabbitMQ Dashboard from the Launchpad.
2. On the login screen, enter the User and Password obtained from Control Hub.

Upon successful login, access the management interface to view and manage RabbitMQ resources.