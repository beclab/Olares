---
outline: [2, 3]
description: Use n8n on Olares to build self-hosted workflow automations with visual nodes, credentials, NocoDB records, webhooks, and reusable workflow JSON.
head:
  - - meta
    - name: keywords
      content: Olares, n8n, workflow automation, self-hosted, no-code, low-code, NocoDB, webhooks, integrations
app_version: "1.0.16"
doc_version: "1.0"
doc_updated: "2026-06-04"
---

# Automate workflows with n8n

n8n is a self-hostable workflow automation tool built around a visual node editor. You can connect apps and APIs, transform data, run JavaScript when a workflow needs custom logic, and trigger automations from schedules, webhooks, forms, or events.

On Olares, n8n runs on your own device, so your workflow definitions, credentials, and execution history stay under your control.

This guide uses release monitoring as a practical example. The same patterns apply to many automations: calling an API, transforming data, saving results to another app, and receiving real-time events through webhooks.

## Learning objectives

In this guide, you will learn how to:

- Install n8n from Market.
- Create the first n8n owner account.
- Build a workflow that calls an API and transforms the response.
- Save workflow results to another app, using NocoDB as an example.
- Receive real-time events through webhooks, using GitHub releases as an example.
- Download and import workflows as JSON.

## Install and set up n8n

1. Open Market and search for "n8n".

   ![n8n](/images/manual/use-cases/n8n.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

3. Open n8n from Launchpad.

4. On the setup page, enter your email, first name, last name, and password. The first registered user automatically becomes the owner of the n8n instance.

   ![Create the n8n owner account](/images/manual/use-cases/n8n-create-owner.png#bordered)

5. Click **Next**. n8n opens the editor page, where you can create your first workflow.

## Build your first workflow

This example checks the latest GitHub release for n8n and extracts the fields you need for upgrade tracking. To monitor another project, replace `n8n-io/n8n` in the request URL with another `owner/repository`, such as `immich-app/immich`.

### Create the workflow

1. On the editor page, click **Start from scratch**.
2. Click **Add first step**, then in the right panel, select **Trigger manually**.
3. Click <i class="material-symbols-outlined">add</i> after the Manual Trigger node, search for "HTTP Request", and add it.
4. In the HTTP Request node, go to the **Parameters** tab and set:

   - **Method**: `GET`
   - **URL**:
     ```text
     https://api.github.com/repos/n8n-io/n8n/releases/latest
     ```
   - **Authentication**: `None`

5. Click **Execute step**. The output panel shows the API response.

   ![Execute the HTTP Request node](/images/manual/use-cases/n8n-http-request-test.png#bordered)

### Keep the release details you need

Add an Edit Fields node to turn the full API response into a compact release summary.

1. Click <i class="material-symbols-outlined">add</i> after the HTTP Request node, search for "Edit Fields", and add it.

   ![Add the Edit Fields node](/images/manual/use-cases/n8n-add-edit-fields-node.png#bordered)

2. Set **Mode** to **Manual Mapping**.
3. Click **Add Field** and add the following fields. For each field, set the name, type, and value:

   | Field | Type | Value |
   |:------|:-----|:------|
   | `app` | String | `n8n` |
   | `latest_version` | String | <code v-pre>{{ $json.tag_name }}</code> |
   | `published_at` | String | <code v-pre>{{ $json.published_at }}</code> |
   | `release_notes` | String | <code v-pre>{{ $json.html_url }}</code> |
   | `is_prerelease` | String | <code v-pre>{{ $json.prerelease }}</code> |

4. Click **Execute step**.
5. Click the node name at the top-left corner and rename it, for example `Monitor n8n releases`.

   ![Edit fields in n8n](/images/manual/use-cases/n8n-edit-fields.png#bordered)

:::tip Test before publishing
n8n supports manual executions while you build, and production executions after you publish. Test each node before publishing the workflow.
:::

### Run and inspect executions

Run the workflow once from the editor, then use the **Executions** tab to review the run history and node outputs.

1. Open the workflow you created.
2. Click **Execute workflow** at the bottom of the editor.
3. Wait for the run to finish. Successful nodes show green check marks, and the connection labels show how many items moved between nodes.
4. Click **Executions** at the top of the editor.
5. Select the target execution to inspect each node's input, output, status, and timing.

   ![Inspect an n8n execution](/images/manual/use-cases/n8n-execution-details.png#bordered)

When the workflow is ready to run automatically, replace Manual Trigger with **On a schedule**, then click **Publish**. For a release monitor, a daily or weekly schedule is usually enough.

## Save workflow results to NocoDB

This example saves the release summary to NocoDB. You can use the same pattern to write form submissions, monitoring results, webhook payloads, or API responses to a table.

n8n stores credentials separately from workflow logic, so you can reuse the same NocoDB API token across multiple workflows.

:::tip Set up NocoDB 
Make sure NocoDB is installed from Market and that you have completed the initial NocoDB setup before continuing.
:::

### Prepare NocoDB

1. Open NocoDB and create a table, for example `Release checks`.
2. Add the following columns. Use text fields to keep the first test simple:

   | Column | Type |
   |:-------|:-----|
   | `app` | Single line text |
   | `latest_version` | Single line text |
   | `published_at` | Single line text |
   | `release_notes` | URL |
   | `is_prerelease` | Single line text |

3. In Olares, go to **Settings** > **Applications** > **NocoDB** > **Entrances** > **NocoDB**, then get the NocoDB endpoint URL.
4. In NocoDB, create an API token from **Account Settings** > **API Tokens**.

:::warning Keep your API token secure
The API token can access your NocoDB data. Do not share it, include it in screenshots, or commit it to public repositories.
:::

### Add the NocoDB node

1. Return to the workflow that contains the `Monitor n8n releases` node.
2. Click <i class="material-symbols-outlined">add</i> after the `Monitor n8n releases` node, search for "NocoDB", select it and select **Create a row**.
3. Click **Set up credential** and enter the NocoDB connection details:

   | Field | Value |
   |:------|:------|
   | **API Token** | The NocoDB API token you created earlier. |
   | **Host** | The NocoDB endpoint URL from Olares Settings. |

4. Click **Save**. A connection test runs automatically.

   ![Add NocoDB credential](/images/manual/use-cases/n8n-add-nocodb-credential.png#bordered)

5. In the NocoDB node, go to the **Parameters** tab and set:

   | Setting | Value |
   |:--------|:------|
   | **Resource** | `Row` |
   | **Operation** | `Create` |
   | **Base Name or ID** | Select the base that contains `Release checks`. |
   | **Table Name or ID** | Select `Release checks`. |
   | **Data to Send** | Select the option based on your needs. |

6. Click **Execute step**. If the node succeeds, return to NocoDB and check that a new row appears in the `Release checks` table.

   ![Check NocoDB result](/images/manual/use-cases/n8n-check-nocodb-result.png#bordered)

7. Click **Execute workflow** to test the complete flow.

To record release checks automatically, replace Manual Trigger with **On a schedule**, then click **Publish** after testing.

## Receive events through webhooks

Webhook workflows let external services send events to n8n in real time. This example receives release events from a GitHub repository you own or administer, then extracts the release details.

You can only add a GitHub webhook to a repository where you have admin access. To monitor a public repository you do not control, such as `n8n-io/n8n`, use the scheduled HTTP Request workflow earlier in this guide.

### Create the webhook workflow

1. In n8n, click <i class="material-symbols-outlined">add</i> in the upper-left corner, then select **Workflow**.
2. Click **Add first step**, then in the right panel, select **On webhook call**.
3. In the Webhook node, go to the **Parameters** tab and set:

   | Setting | Value |
   |:--------|:------|
   | **HTTP Method** | `POST` |
   | **Path** | `github-release-event` |
   | **Authentication** | `None` |
   | **Respond** | `Immediately` |

4. Click **Listen for test event**, then copy the **Test URL** from the Webhook node. Use `https://` in GitHub. If n8n copies an `http://` URL, change it to `https://`.

### Add the webhook in GitHub

1. In your browser, open GitHub and go to the repository you own or administer.
2. In that repository, go to **Settings** > **Webhooks**, then click **Add webhook**.
3. Configure the webhook:

   | Setting | Value |
   |:--------|:------|
   | **Payload URL** | Paste the HTTPS test webhook URL from n8n. |
   | **Content type** | `application/json` |
   | **Secret** | Leave empty for this first test. |
   | **Which events would you like to trigger this webhook?** | Select **Let me select individual events**, then select **Releases**. |

   ![Add webhook](/images/manual/use-cases/n8n-add-webhook.png#bordered)

4. Keep **Active** enabled, then click **Add webhook**.
5. In n8n, wait for the test event. GitHub sends a `ping` event after you add the webhook.
6. To test a full release payload, publish a test release in the repository, then return to n8n and check the Webhook node output.

   ![Webhook test result](/images/manual/use-cases/n8n-webhook-test-result.png#bordered)

### Extract release fields

The initial GitHub `ping` event does not include a release object. After n8n receives a release event, add an Edit Fields node after the Webhook node to keep the payload readable.

1. Click <i class="material-symbols-outlined">add</i> after the Webhook node, search for "Edit Fields", and add it.
2. Set **Mode** to **Manual Mapping**.
3. Click **Add Field** and add the following fields:

   | Field | Type | Value |
   |:------|:-----|:------|
   | `event_action` | String | <code v-pre>{{ $json.body.action }}</code> |
   | `repository` | String | <code v-pre>{{ $json.body.repository.full_name }}</code> |
   | `release_tag` | String | <code v-pre>{{ $json.body.release.tag_name }}</code> |
   | `release_name` | String | <code v-pre>{{ $json.body.release.name }}</code> |
   | `release_url` | String | <code v-pre>{{ $json.body.release.html_url }}</code> |
   | `sender` | String | <code v-pre>{{ $json.body.sender.login }}</code> |

4. Click **Execute step** to check the extracted fields.

   ![Extract release fields](/images/manual/use-cases/n8n-extract-release-fields.png#bordered)

You can then connect the extracted fields to another node, such as NocoDB, Slack, or email.

### Publish and switch to the production URL

After the test event works, publish the workflow and update GitHub to use the production URL.

1. In n8n, click **Publish** in the upper-right corner. The production webhook starts receiving events after the workflow is published.
2. Open the Webhook node and copy the **Production URL**.
3. If the production URL starts with `http://`, change it to `https://`.
4. Return to the GitHub webhook settings and replace the test URL with the HTTPS production URL.
5. Save the webhook settings in GitHub.

   <!-- ![Configure a webhook trigger in n8n](/images/manual/use-cases/n8n-webhook-trigger.png#bordered) -->

## Manage workflows

n8n workflows can be downloaded as JSON and imported into another n8n instance. This is useful for backup, version review, and sharing workflow templates with teammates.

### Download a workflow

1. Open the workflow.
2. Click <i class="material-symbols-outlined">more_horiz</i> > **Download**.
   
   ![Download a workflow](/images/manual/use-cases/n8n-download-workflow.png#bordered)

A JSON file is downloaded to your computer.

### Import a workflow

1. Create a new workflow.
2. Click <i class="material-symbols-outlined">more_horiz</i> > **Import from File**.
   
   ![Import a workflow](/images/manual/use-cases/n8n-import-workflow.png#bordered)

3. Choose the workflow JSON file.
4. Review each credential field and reconnect credentials before publishing the workflow.

:::warning Review imported workflows
Imported workflows can contain Code nodes, HTTP calls, and webhook paths. Review every node before running workflows from an untrusted source.
:::

## FAQs

### Why is my webhook not receiving events from external services?

Check the following items:

- Make sure the webhook URL you added to the external service starts with `https://`.
- When testing, make sure n8n is still waiting after you click **Listen for test event**.
- After publishing, make sure the external service uses the **Production URL**, not the test URL.
- If the external service provides delivery logs, check the response status there.
- If the external service still cannot reach n8n, open **Settings** > **Applications** > **n8n** > **Entrances** > **n8n** and check **Authentication level**. Use **Public** only when needed because it exposes the n8n entrance to the internet.

## Learn more

- [n8n workflow documentation](https://docs.n8n.io/workflows/): Learn how workflows, nodes, templates, executions, and sharing work in n8n.
- [n8n integrations documentation](https://docs.n8n.io/integrations/): Browse built-in nodes, community nodes, credential-only nodes, and generic API integration options.
- [HTTP Request node](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.httprequest/): Configure REST API calls or import a `curl` command into n8n.
