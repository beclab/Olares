---
outline: [2, 3]
description: Learn how to visualize Prometheus metrics in Olares using Grafana dashboards.
---
# Use Grafana dashboards

To visualize system metrics in Olares, you can run Grafana and connect it to the built-in Prometheus service. This guide explains how to install Grafana, connect the data source, and import a standard dashboard.

## Install Grafana

Before using Grafana, install it from Olares Market.

1. Open Olares Market from Launchpad and search for "Grafana".
2. Click **Get**, then **Install**.
3. In the pop-up window, set your login credentials:
   - `GF_USERNAME`: Grafana login username.
   - `GF_PASSWORD`: Grafana login password.
    :::tip Remember your login credentials
    These are the login credentials for Grafana. You will need them if you access Grafana later.
    :::
    ![Set login credentials](/public/images/developer/develop/middleware/mw-grafana-set-login.png#bordered){width=90% style="margin-left:0"}
4. Wait for the installation to complete.

## Access Grafana

1. Open **Grafana** from Launchpad, then click <i class="material-symbols-outlined">open_in_new</i> to open it in a new tab.
2. On the login screen, enter the `GF_USERNAME` and `GF_PASSWORD` you configured during installation.

After logging in, you will see the Grafana home page.

## Add Prometheus data source

Olares runs a built-in Prometheus service that collects system metrics. 

To connect Grafana to this internal service:

1. In the Grafana left navigation pane, go to **Connections** > **Data sources**.
2. Click **Add data source**, then select **Prometheus**.
3. For the **Prometheus server URL** field, enter:  
    ```text
    http://dashboard.<olaresid>.olares.com
    ```
    Replace `<olaresid>` with your Olares ID.
4. Click **Save & test** at the bottom of the page. If the connection is successful, you will see the prompt below.
    ![Successful connection](/public/images/developer/develop/middleware/mw-grafana-connect.png#bordered){width=90% style="margin-left:0"}

## Create a dashboard

This approach is suitable when you need custom metrics and visualizations and are familiar with PromQL.
1. In the left navigation pane, click **Dashboards**.
2. Click **+ Create dashboard**, then select **+ Add visualization**.
3. Select **prometheus** as the data source.
4. Configure panels, PromQL queries, and expressions as needed.
5. Click **Save dashboard** in the top-right corner for future use.

## Import a dashboard (recommended)

If you do not need to build dashboards from scratch, you can import existing dashboards.

1. Visit the [Grafana Dashboard library](https://grafana.com/grafana/dashboards/).
2. Download the required dashboard as a `JSON` file.
3. In Grafana, click <i class="material-symbols-outlined">add_2</i> in the top-right corner and select **Import dashboard**.
4. Upload the `JSON` file, and select **prometheus** as the data source.
5. Click **Import** to complete the import.

Imported dashboards provide predefined panels and queries and can be customized after import.
![Imported dashboard](/public/images/developer/develop/middleware/mw-grafana-dashboard.png#bordered){width=90%}