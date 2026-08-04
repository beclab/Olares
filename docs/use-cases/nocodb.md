---
outline: [2, 3]
description: Set up NocoDB on Olares as a no-code database platform. Create tables, import data, configure SMTP, manage team access, and integrate with n8n for workflow automation.
head:
  - - meta
    - name: keywords
      content: Olares, NocoDB, no-code database, Airtable alternative, self-hosted, spreadsheet, workflow automation, n8n
app_version: "1.0.10"
doc_version: "2.0"
doc_updated: "2026-07-31"
---

# Build a self-hosted spreadsheet database with NocoDB

NocoDB is an open-source no-code database platform that turns any database into a smart spreadsheet interface, similar to Airtable. It provides a rich web UI for managing your data visually, along with a full REST API. Running NocoDB on Olares gives you a self-hosted, privacy-first alternative to cloud-based spreadsheet tools.

## Learning objectives

In this guide, you will learn how to:
- Install NocoDB and create your admin account.
- Create tables and import data from external sources.
- Configure SMTP for outgoing email.
- Invite team members and manage permissions.
- Continue to an n8n workflow that writes data to NocoDB.

## Install NocoDB

1. Open Market and search for "NocoDB".

   ![NocoDB](/images/manual/use-cases/nocodb.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Set up NocoDB

1. Open NocoDB from Launchpad.
2. Enter your email and password, then click **Sign Up**.

   The first registered user automatically becomes the Super Admin and can manage team member permissions.

   ![NocoDB registration page](/images/manual/use-cases/nocodb-register.png#bordered){width=80%}

## Create tables and import data

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

## Configure email

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

## Invite team members

1. Click the profile icon in the bottom-left corner and go to **Account Settings**.
2. In the left sidebar, expand **Users** and select **User Management**.
3. Click **Invite User** in the top-right corner.
4. Enter the team member's email address, set the appropriate access level, and click **Invite**.

   ![Invite team members in NocoDB](/images/manual/use-cases/nocodb-invite-member.png#bordered)

5. If NocoDB shows **Copy Invite URL**, copy the URL and send it to the invited member.

Invited members can use the invitation email or invite URL to sign up.

## Automate NocoDB with n8n

NocoDB works well as the data store for n8n automations. After creating a base and table, follow [Save workflow results to NocoDB](n8n.md#save-workflow-results-to-nocodb) to:

- Copy the current NocoDB endpoint from Olares Settings.
- Create a NocoDB API token.
- Add NocoDB credentials and a NocoDB node to an n8n workflow.
- Test that the workflow creates a row in your table.

## Learn more

- [Automate workflows with n8n](n8n.md): Build workflows, save results to NocoDB, receive webhooks, and reuse workflow JSON.
- [NocoDB documentation](https://docs.nocodb.com/): Official documentation for NocoDB features and API reference.
