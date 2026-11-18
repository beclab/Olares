---
outline: [2, 3]
description: Keep a workout and nutrition journal in wger on Olares, and use its MCP integration to query records and log body weight from Lares or Hermes Agent.
head:
  - - meta
    - name: keywords
      content: "Olares, wger, workout journal, body weight, nutrition, MCP, Lares, Hermes Agent, self-hosted"
app_version: "0.0.10"
doc_version: "1.3"
doc_updated: "2026-11-18"
---

# Keep a workout and nutrition journal with wger

wger is a fitness journal that brings workout routines, training logs, body measurements, and nutrition into one browser application. On Olares, you can keep your plans and records on your own device and return to them between sessions.

Start in the browser to plan a workout, record what you did, and keep track of your weight and food. You can then access the same journal from a desktop or mobile client, or use Lares or Hermes Agent to query your records.

## Install wger

1. Open Market and search for "wger".
2. Click **Get**, then **Install**. When prompted, set **ADMIN_PASSWORD** to a password of at least 12 characters and save it. Wait for installation to complete.

## Use wger

Open wger from Launchpad and sign in to the browser version.

:::tip Sign-in credentials
Use `admin` as the username and the **ADMIN_PASSWORD** you set during installation as the password.
:::

### Create a workout routine

1. Open **Training** > **Routines** and click **+**.
2. In the **Add** panel, set:
   - **Name**: Enter a name for the routine.
   - **Start** and **End**: Choose the dates the routine should cover.
   - **Fixed weekly schedule**: Leave this on if the training days should repeat each week.
3. Click **Save** to open the routine in its edit view.
4. Click **Add training day**, name the day, and click **Save** in the day panel.
5. Click **Add exercise**. In the new exercise block, set:
   - **Search by exercise name**: Find and select an exercise.
   - **Sets**: Enter the number of sets you plan to do.
   - **Weight** and **Repetitions**: Enter the planned values for each set.

Return to **Training** > **Routines** and reopen the routine. Confirm that the training day, exercise, and planned sets are still there.

![A training routine with one exercise](/images/manual/use-cases/wger-routine.jpg)

### Log a training session

1. Open **Training** > **Routines** and select the routine you want to log.
2. Find the training day you completed and click <i class="material-symbols-outlined">add_chart</i> beside it to open the training log.
3. In the training log, check:
   - **Date**: Set the date you completed the session.
   - **Exercises**: Replace the prefilled weight and repetitions for each set with what you completed. Only rows with a weight or repetition value are saved.
4. Click **Submit** beneath the exercise rows to save the sets. A **Success!** message confirms the submission.
5. Check the saved session and exercise entries in the routine.
6. To record optional session details, use the separate form above **Exercises**:
   - Add notes or select a **General impression**.
   - Click that form's **Submit**.

Keep planned values in the routine and actual results in the log. This lets you reuse the same routine without replacing earlier training records.

### Track your body weight

1. Open **Body weight** > **Weight overview**.
2. Click **+** to open the **Add** panel. Set:
   - **Weight**: Enter your measurement and choose **KG** or **LB**.
   - **Date** and **Time**: Check when you took the measurement.
3. Click **Submit** and check that the measurement appears in the table below the graph.

   ![Add a body-weight measurement](/images/manual/use-cases/wger-weight-entry.png#bordered)

4. Add later measurements as you take them. Use the filters above the graph to review your progress.

### Log food in your nutrition diary

A diary entry needs an ingredient and a nutrition plan. Create those first, then record what you ate.

**Add an ingredient**

1. Open **Nutrition** > **Ingredient overview** and click **Add**.
2. Enter the ingredient's values per 100 g:
   - **Name**: Enter the food name.
   - **Energy**, **Protein**, **Carbohydrates**, and **Fat**: Use the values on the food's nutrition label.
   - **License**: Select a license for the ingredient.
3. Click **Save**.

**Record a diary entry**

1. Open **Nutrition** > **Nutrition plans** and click **+**. In the plan panel:
   - **Description**: Enter a name for the plan.
   - **Only track calories**: Leave the default on to log food without setting up scheduled meals.
2. Submit the plan, open it, and click **Add nutrition diary entry**.
3. In the diary entry panel:
   - Search for and select the ingredient you added.
   - Enter the amount and date.
4. Click **Submit**. Under **Logged**, check that the date and nutrition totals appear.
5. Click the date to open the daily diary. Confirm that the ingredient and amount appear there.

   ![Daily nutrition diary showing the logged ingredient, amount, and nutrient values](/images/manual/use-cases/wger-nutrition-diary.png#bordered)

## Connect a desktop or mobile client (optional)

To access your journal on another device, connect a wger client to your Olares instance.

### Prerequisites

- wger Chart 0.0.10 or later is installed.
- LarePass is installed on the client device.

### Download a client

| Platform | Download |
| --- | --- |
| Windows and macOS | [wger client releases](https://github.com/wger-project/flutter/releases) |
| Linux | [wger client releases](https://github.com/wger-project/flutter/releases) or [Flathub](https://flathub.org/apps/de.wger.flutter) |
| Android | [Google Play](https://play.google.com/store/apps/details?id=de.wger.flutter), [F-Droid](https://f-droid.org/packages/de.wger.flutter/), or [wger client releases](https://github.com/wger-project/flutter/releases) |
| iPhone | [App Store](https://apps.apple.com/app/id6502226792) |

For GitHub releases, choose the installer that matches your operating system.

### Connect to your Olares journal

1. Enable LarePass VPN on the client device.
2. In Olares Settings, open **Applications** > **wger** > **Entrances** > **wger Mobile**. Keep **Authentication level** set to **Internal**, then copy the **Endpoint**.

   ![The wger Mobile endpoint in Olares Settings](/images/manual/use-cases/wger-mobile-endpoint.jpg#bordered)

3. Open the wger client's sign-in screen and expand **Advanced**. Set:
   - **Server**: Select **Self-hosted**.
   - **URL of the wger instance**: Enter the copied endpoint.
   - **Sign-in method**: Select **Username and password**.
   - **Allow self-signed certificates**: Leave this off.

   ![Advanced settings for a self-hosted wger client](/images/manual/use-cases/wger-client-advanced.jpg#bordered){width=70%}

4. Tap **Done**, then sign in with `admin` and the password you set during installation.
5. Keep LarePass VPN enabled while the client syncs. Check that the routine you created in the browser appears in the client.

## Use wger with an AI assistant (optional)

Connect an AI assistant to wger through MCP to query your journal and record body weight. The setup examples below use Lares and Hermes Agent. MCP requires wger Chart 0.0.10 or later.

:::info Which account does MCP use?
This preview connection accesses the wger `admin` account, regardless of which wger account you use in a client. To query the records created above, keep them in `admin`. Anyone with the MCP token can access that account's records and enabled creation tools.
:::

First complete the [Lares setup](lares.md#get-started) or [Hermes Agent setup](hermes.md#step-2-run-the-setup-wizard), with a chat model available through [Router](olares-router.md).

### Get the connection details

You need two values to connect an assistant:

- **MCP endpoint** (`WGER_MCP_ENDPOINT`): In Olares Settings, open **Applications** > **wger** > **Entrances** > **wger MCP** and copy the **Endpoint**.
- **MCP token** (`WGER_MCP_TOKEN`): In Files, open **Data** > **wger** > **data** > **state** > **mcp-credentials** > `auth.private` and copy `MCP_STATIC_TOKEN`. Do not use the separate wger API key in the file.

### Connect an AI agent

Choose an agent below. Both use the same queries later in this guide.

<Tabs>
<template #Lares>

1. Open Lares and go to **Settings** > **MCP**.
2. Click **Add server** and enter the connection details:

   | Field | Value |
   | --- | --- |
   | **Server name** | `wger_lab` |
   | **Transport** | `Streamable HTTP` |
   | **MCP URL** | `WGER_MCP_ENDPOINT/mcp` |
   | **Headers** | The JSON object below, with your MCP token substituted |

   ```json
   {
     "Authorization": "Bearer WGER_MCP_TOKEN"
   }
   ```

   :::tip
   Keep `Bearer` and the space before the token when replacing `WGER_MCP_TOKEN`.
   :::

3. Click **Save**. Keep the server enabled. Lares loads the tools after saving. If the server reports a load error, check the MCP URL and token before continuing.
4. Start a conversation and send one of the requests in the next section. When Lares asks for permission, review the requested wger tool and its arguments before approving it.

</template>
<template #Hermes-Agent>

1. In Files, go to **Data** > **hermesagent** > **home** and make a backup of `config.yaml`.
2. Add the following `wger_lab` entry under `mcp_servers`. Replace `WGER_MCP_ENDPOINT` and `WGER_MCP_TOKEN` with the values you obtained above, keeping `/mcp` after the endpoint. Preserve existing servers and your model settings. If `mcp_servers` already exists, add the entry under it instead of creating a second key.

   ```yaml
   mcp_servers:
     wger_lab:
       url: "WGER_MCP_ENDPOINT/mcp"
       headers:
         Authorization: "Bearer WGER_MCP_TOKEN"
       connect_timeout: 30
       timeout: 60
       tools:
         include:
           - list_routines
           - get_routine
           - search_exercises
           - get_exercise
           - list_workout_logs
           - list_workout_sessions
           - list_nutrition_plans
           - list_log_items
           - nutrition_summary
           - get_body_weight_history
           - log_body_weight
   ```

3. Save the file. If you use [terminal chat](hermes.md#option-1-terminal-chat), exit the current session with `/exit`, then run `hermes` again to load the connection.
4. Send one request at a time and review the wger tool results in the conversation.

</template>
</Tabs>

### Query your journal

After connecting an assistant, ask it to read the records you created above. Replace the bracketed names with your own:

```text
Find my "[routine name]" routine in wger. Show its training sessions
and completed sets so far. Then check "[nutrition plan name]" and show
the food entries I logged. If either has no entries, say so rather
than estimating values.
```

![Lares shows training sessions and completed sets retrieved from wger](/images/manual/use-cases/wger-lares-journal-result.png)

Compare the answer with the training log and the nutrition plan's **Logged** section in wger.

### Record your body weight

Ask the assistant to save a measurement. Replace the bracketed details with your own:

```text
I weighed [weight] [kg or lb] on [date]. Please add that to wger with
the note "[distinct note]". If I've already recorded that entry,
don't add it again. Show me the entry when you're done.
```

Verify the saved entry under **Body weight** > **Weight overview** in wger.

## Learn more

- [wger documentation](https://wger.readthedocs.io/en/latest/): Learn about routines, nutrition, and other features.
