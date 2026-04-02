---
outline: [2, 3]
description: Set up NocoDB on Olares as a no-code database platform. Create tables, import data, configure SMTP, manage team access, and integrate with n8n for workflow automation.
head:
  - - meta
    - name: keywords
      content: Olares, NocoDB, no-code database, Airtable alternative, self-hosted, spreadsheet, workflow automation, n8n
app_version: "1.0.9"
doc_version: "1.0"
doc_updated: "2026-03-27"
---

# Turn your database into a web app with NocoDB

NocoDB is an open-source no-code database platform that turns any database into a smart spreadsheet interface, similar to Airtable. It provides a rich web UI for managing your data visually, along with a full REST API. Running NocoDB on Olares gives you a self-hosted, privacy-first alternative to cloud-based spreadsheet tools.

## Learning objectives

In this guide, you will learn how to:
- Install NocoDB and create your admin account
- Create tables and import data from external sources
- Configure SMTP for email notifications and password recovery
- Invite team members and manage permissions
- Connect NocoDB with n8n for workflow automation

## Install NocoDB

1. Open Market and search for "NocoDB".
2. Click **Get**, then click **Install**, and wait for the installation to complete.

   :::info
   NocoDB is publicly accessible by default.
   :::

<!-- ![Install NocoDB from Market](/images/manual/use-cases/nocodb-install-market.png#bordered) -->

## Register your account

1. Open NocoDB from Launchpad.
2. Enter your email and password, then click **Sign Up**.

   The first registered user automatically becomes the **Super Admin** and can manage team member permissions.

<!-- ![NocoDB registration page](/images/manual/use-cases/nocodb-register.png#bordered) -->

## Create tables and import data

You can either create tables manually or import existing data.

1. Open your target project in NocoDB.
2. To create a new table, click **+** and add fields as needed.
3. To import data, click **+** > **Import**, and select from the supported formats:
   - Airtable
   - CSV
   - Excel
   - JSON

<!-- ![Create a new table in NocoDB](/images/manual/use-cases/nocodb-create-table.png#bordered) -->

<!-- ![Import data into NocoDB](/images/manual/use-cases/nocodb-import-data.png#bordered) -->

## Configure SMTP

Setting up SMTP enables email-related features such as password recovery. Without SMTP configured, you won't be able to reset your NocoDB account password via email.

1. Click the account icon in the bottom-left corner and go to **Account Settings**.
2. Under **Email Configuration**, click **Edit**.
3. Select **SMTP** and fill in the required fields. For example, to use Outlook:

   | Field | Value |
   |:------|:------|
   | **From Address** | Your full Outlook email address |
   | **SMTP Server** | `smtp.office365.com` |
   | **Username** | Your full Outlook email address |
   | **Password** | Your app password |
   | **Port** | `587` (STARTTLS) or `465` (SSL) |
   | **Encryption** | STARTTLS or SSL/TLS (match the port) |

4. Click **Save** to apply the SMTP settings.

<!-- ![NocoDB account settings](/images/manual/use-cases/nocodb-account-settings.png#bordered) -->

<!-- ![Configure SMTP in NocoDB](/images/manual/use-cases/nocodb-smtp-config.png#bordered) -->

:::tip
For other email providers, refer to their documentation for the correct SMTP server, port, and encryption settings.
:::

## Invite team members

1. Click the team icon in the top-left corner and go to **Team & Settings**.
2. In the user management page, click **Invite User** in the top-right corner.
3. Enter the team member's email address, set the appropriate permission level, and click **Invite**.

   Invited members can sign in through the link in the invitation email.

<!-- ![Invite team members in NocoDB](/images/manual/use-cases/nocodb-invite-member.png#bordered) -->

## Automate workflows with n8n

You can connect NocoDB to n8n, an open-source workflow automation tool, to automate data entry and other repetitive tasks.

### Install n8n

1. Open Market and search for "n8n".
2. Click **Get**, then click **Install**.
3. Open n8n and complete the initial registration.

### Create an API token in NocoDB

1. In NocoDB, click the account icon in the bottom-left corner and go to **Account Settings**.
2. Navigate to the **Tokens** tab.
3. Click **Create**, name your token, and click **Save**.
4. Copy the generated token for later use.

### Connect n8n to NocoDB

1. In n8n, create a new workflow.
2. Add a trigger node as the first step (for example, a schedule or webhook trigger).
3. Add a NocoDB node to the workflow:

   a. In the node search box, type "NocoDB" and select it.

   b. Click **Create New Credential**.

   c. Paste the API token you copied from NocoDB.

   d. For the **Host**, copy and paste NocoDB's URL from your browser, and click **Save**.

4. Configure the NocoDB node by selecting the operation type (e.g., create, read, update, or delete rows) and the target table.
5. Click **Test** to verify the workflow runs correctly.
6. Once verified, activate the workflow.

<!-- ![Create API token in NocoDB](/images/manual/use-cases/nocodb-create-token.png#bordered) -->

<!-- ![Configure NocoDB node in n8n](/images/manual/use-cases/nocodb-n8n-config.png#bordered) -->

<!-- ![Test and activate n8n workflow](/images/manual/use-cases/nocodb-n8n-workflow.png#bordered) -->

## Learn more
- [NocoDB documentation](https://docs.nocodb.com/): Official documentation for NocoDB features and API reference.
