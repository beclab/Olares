---
outline: [2, 3]
description: Set up NocoDB on Olares as a no-code database platform. Create tables, import data, configure SMTP, manage team access, and integrate with n8n for workflow automation.
head:
  - - meta
    - name: keywords
      content: Olares, NocoDB, no-code database, Airtable alternative, self-hosted, spreadsheet, workflow automation, n8n
app_version: "1.0.10"
doc_version: "1.0"
doc_updated: "2026-06-03"
---

# Build a self-hosted spreadsheet database with NocoDB

NocoDB is an open-source no-code database platform that turns any database into a smart spreadsheet interface, similar to Airtable. It provides a rich web UI for managing your data visually, along with a full REST API. Running NocoDB on Olares gives you a self-hosted, privacy-first alternative to cloud-based spreadsheet tools.

## Learning objectives

In this guide, you will learn how to:
- Install NocoDB and create your admin account.
- Create tables and import data from external sources.
- Configure SMTP for outgoing email.
- Invite team members and manage permissions.
- Connect NocoDB with n8n for workflow automation.

## Install NocoDB

1. Open Market and search for "NocoDB".

   ![NocoDB](/images/manual/use-cases/nocodb.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Set up and use NocoDB

### Register your account

1. Open NocoDB from Launchpad.
2. Enter your email and password, then click **Sign Up**.

   The first registered user automatically becomes the **Super Admin** and can manage team member permissions.

   ![NocoDB registration page](/images/manual/use-cases/nocodb-register.png#bordered){width=80%}

### Create tables and import data

You can either create tables manually or import existing data.

1. Open the default **Getting Started** base, or select another base from the workspace menu.
2. To create a new table, use either method:
   - On the **Overview** page, click **Create New Table**.
   - In the left sidebar, click **Create New** > **Table**.

   ![Create a new table in NocoDB](/images/manual/use-cases/nocodb-create-table.png#bordered)

3. To import data, go to **Overview**, click **Import Data**, and select from the supported formats:
   - Airtable
   - CSV
   - Excel
   - JSON

   ![Import data into NocoDB](/images/manual/use-cases/nocodb-import-data.png#bordered)

### Configure SMTP

Setting up SMTP enables NocoDB to send email from your configured sender address.

1. Click the profile icon in the bottom-left corner and go to **Account Settings**.
2. On the **Configure E-mail** panel, click **Configure**.
3. Select **SMTP** and fill in the SMTP settings provided by your email service provider.

   | Field | Value |
   | :-- | :-- |
   | **From address** | Your sender email address, such as `name@example.com`.|
   | **From domain**  | The domain after `@`, such as `example.com`. |
   | **SMTP server**  | The SMTP server address from your email provider, such as<br> `smtp.example.com`.|
   | **SMTP port** | The SMTP port from your email provider. `587` for TLS, `465` for SSL,<br> or `25` for insecure connections. |
   | **Username** | Your SMTP username. This is usually your full email address. |
   | **Password** | Your SMTP password, app password, or authorization code.                        |

4. Adjust the security switches if your email provider requires it.
5. Click **Test** to check the connection, then click **Save** to apply the SMTP settings.

### Invite team members

1. Click the profile icon in the bottom-left corner and go to **Account Settings**.
2. In the left sidebar, expand **Users** and select **User Management**.
3. Click **Invite User** in the top-right corner.
4. Enter the team member's email address, set the appropriate access level, and click **Invite**.

   ![Invite team members in NocoDB](/images/manual/use-cases/nocodb-invite-member.png#bordered)

5. If NocoDB shows **Copy Invite URL**, copy the URL and send it to the invited member.

Invited members can use the invitation email or invite URL to sign up.

## Advanced: Automate workflows with n8n

You can connect NocoDB to n8n, an open-source workflow automation tool, to automate data entry. This example shows how to create a row in an existing NocoDB table from an n8n workflow.

### Install n8n

If n8n is not installed yet, add it from Market first.

1. Open Market and search for "n8n".

   ![n8n](/images/manual/use-cases/n8n.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

After installation, open n8n from Launchpad and complete the initial registration.

### Prepare NocoDB connection details

Before configuring n8n, prepare the NocoDB endpoint URL and an API token.

1. Get the endpoint URL:

   a. In Olares, go to **Settings** > **Applications** > **NocoDB** > **Entrances**.

   b. Open the **NocoDB** entrance and record the endpoint URL.

2. Create an API token:

   a. In NocoDB, click the profile icon in the bottom-left corner and go to **Account Settings**.

   b. Navigate to the **API Tokens** tab.

   c. Click **Create new token**, name your token, and click **Save**.

   d. Store the generated token somewhere secure temporarily. You will enter it when setting up the n8n credential.

:::warning Keep your API token secure
The API token can be used to access your NocoDB data. Do not share it, include it in screenshots, or commit it to public repositories. If the token is exposed, delete it and create a new one.
:::

### Configure the NocoDB node in n8n

1. In n8n, create a new workflow and add a trigger node, such as a Schedule Trigger.
2. Click <i class="material-symbols-outlined">add</i>  next to the trigger node.
3. In the search bar on the right, enter "NocoDB", then select **NocoDB**.

   ![Add NocoDB node in n8n](/images/manual/use-cases/nocodb-add-node.png#bordered){width=95%}

4. Choose the operation you want to perform. This example uses **Create a row**.

   ![NocoDB create a row](/images/manual/use-cases/nocodb-create-row.png#bordered){width=55%}

5. On the **Parameters** tab, click **Set up credential** and configure the connection:

   | Field | Value |
   | :-- | :-- |
   | **API Token** | Enter the API token you generated earlier. |
   | **Host** | Enter the NocoDB endpoint URL you prepared earlier. |

6. Click **Save**. The connection is tested automatically.

   ![Set up credential](/images/manual/use-cases/nocodb-n8n-credential.png#bordered){width=90%}

7. Configure the data parameters:

   | Field | Value |
   | :-- | :-- |
   | **Resource** | Select `Row`. |
   | **Operation** | Select `Create`. |
   | **Base Name or ID** | Select the base that contains your table. |
   | **Table Name or ID** | Select the table you want to update. |
   | **Data to Send** | Select how to map input data to NocoDB columns. |

8. Click **Execute step** to verify the node.

   ![Set up workflow](/images/manual/use-cases/nocodb-n8n-execute-step.png#bordered){width=95%}

9. After the node runs successfully, close the node configuration panel and click **Execute workflow** to test the full workflow.

   ![Test and activate n8n workflow](/images/manual/use-cases/nocodb-execute-workflow.png#bordered){width=95%}

10. Go to NocoDB and check the target table. A new row should appear if the workflow ran successfully.

   ![Verify result in NocoDB](/images/manual/use-cases/nocodb-verify-result.png#bordered){width=95%}

To let the schedule run automatically, publish the workflow after testing.

## Learn more
- [NocoDB documentation](https://docs.nocodb.com/): Official documentation for NocoDB features and API reference.
