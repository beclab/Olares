---
outline: [2, 3]
description: Keep a workout and nutrition journal in wger on Olares, and use its MCP integration to query records and log body weight from Lares or Hermes Agent.
head:
  - - meta
    - name: keywords
      content: "Olares, wger, workout journal, body weight, nutrition, MCP, Lares, Hermes Agent, self-hosted"
app_version: "0.0.10"
doc_version: "1.2"
doc_updated: "2026-10-31"
---

# Keep a workout and nutrition journal with wger

wger is a fitness journal that brings workout routines, training logs, body measurements, and nutrition into one browser application. On Olares, you can keep your plans and records on your own device and return to them between sessions.

The MCP integration lets Lares or Hermes Agent read your journal and add records through natural-language requests.

:::info Package version
Client synchronization and MCP require wger Chart 0.0.10 or later.
:::

## Learning objectives

- Build a routine with a training day and an exercise.
- Record completed sets and follow your weight over time.
- Add an ingredient and log it in a nutrition diary.
- Connect a desktop or mobile client to your Olares journal.
- Connect an AI assistant to wger, summarize your records, and save a body-weight entry.

## Install and sign in

1. Open Olares Market and search for **wger**.
2. Click **Get**, then **Install**.
3. Set **ADMIN_PASSWORD** to a password of at least 12 characters. Keep it for your first sign-in.
4. When installation finishes, open wger from Launchpad.
5. Sign in with username `admin` and the password you supplied during installation.

The installation password initializes an empty database. Updating the app keeps existing accounts and records. Public registration and email delivery are disabled in this package.

## Connect a desktop or mobile client

Use the browser on your computer, or connect an official wger client to the same Olares instance. Sign in with the same wger account on each device to use the same journal.

### Download a client

| Platform | Download |
| --- | --- |
| Windows | [GitHub Releases](https://github.com/wger-project/flutter/releases), select `windows-x64.zip`. |
| macOS | [GitHub Releases](https://github.com/wger-project/flutter/releases), select `wger.app.zip`. |
| Linux | [GitHub Releases](https://github.com/wger-project/flutter/releases), select `linux-x64.tar.gz`, or install from [Flathub](https://flathub.org/apps/de.wger.flutter). |
| Android | [GitHub Releases](https://github.com/wger-project/flutter/releases), select `app-release.apk`, or use [Google Play](https://play.google.com/store/apps/details?id=de.wger.flutter) or [F-Droid](https://f-droid.org/packages/de.wger.flutter/). |
| iPhone | [App Store](https://apps.apple.com/app/id6502226792). Availability depends on your store region. |

On GitHub, expand **Assets** under a release to find the packages. Windows, macOS, Linux and Android packages are available there. For iPhone, use the App Store. The `Runner.app.zip` and `Runner.xcarchive.zip` assets are development artifacts, not packages you can install directly on a regular iPhone.

### Set your server and sign in

1. Install LarePass on the computer or phone where you will run the wger client. Sign in with the Olares ID for your device and enable its VPN connection. On desktop, open the avatar menu and turn on **VPN connection**. On mobile, open **Settings** and enable VPN in **My Olares**.
2. In Olares Settings, open **Applications** > **wger** > **Entrances**. Copy the HTTPS address of **wger Mobile**. Despite its name, this entrance is used by both desktop and mobile clients. Keep its access level set to **internal**.
3. Open the wger client. On its sign-in screen, expand **ADVANCED** and set the custom server to the address you copied. Enter only the server root, such as `https://YOUR-WGER-CLIENT-HOST`, without `/api`, `/mcp` or `/ps`.
4. Sign in with your wger username and password. For the initial account, use `admin` and the password you set during installation. These are wger credentials, separate from your LarePass sign-in.

Keep LarePass VPN enabled while connecting and synchronizing. Leaving the default server unchanged connects the client to the public wger service instead of your Olares journal. Do not use the separate **wger MCP** address as the client server.

### Use the same journal across devices

After signing in, wait for the initial synchronization before opening your records. To check that you selected the intended server, create a custom measurement category in the client, then open the web journal and look for that category under the same account.

This package includes PowerSync to synchronize your journal between connected clients and Olares. Use the same server and wger account on your computer and phone. Keep LarePass VPN enabled while synchronizing, and wait for synchronization to finish before switching devices.

## Build your first routine

The screenshots use fictional demonstration records. Substitute your own routine and measurements.

1. Open **Training** and select the routine overview.
2. Click the **+** button. Enter a name, such as `Weekly strength journal`, and choose the start and end dates.
3. Keep **Fixed weekly schedule** enabled to repeat training days each week. Click **Save**.
4. Open the routine menu and select **Edit**. Click **Add training day** and name the day, such as `Full body A`.
5. Add an exercise to that day. Search the exercise catalog, select an exercise, and enter its sets, repetitions, and weight. Save your changes.

The routine view groups exercises by training day. The example below contains Squats with three sets of eight repetitions at 40 kg. These values illustrate the interface rather than a recommended workout.

![A training routine with one exercise](/images/manual/use-cases/wger-routine.jpg)

The exercise catalog is synchronized during first startup. If search has no results, check the app's network connection and allow the initial synchronization to finish.

## Record a training session

1. In the routine, click **Add training log** beside the training day.
2. Review the date and record the repetitions and weight you actually completed for each set.
3. Submit the exercise entries. Use the session section to add notes and a session impression, then submit that section as well.

Keep planned values in the routine and actual results in the log. This lets you reuse the same routine without replacing earlier training records.

## Follow your weight

1. Open **Body weight** > **Weight overview**.
2. Click **+**, enter the measurement and date, and save it.
3. Add later measurements as you take them. Use the time filters above the graph to review a shorter or longer period.

The graph shows individual values, an average, and a trend. The table below it keeps the dates and notes for each entry.

![Body weight history and trend](/images/manual/use-cases/wger-weight.jpg)

## Log nutrition

### Add an ingredient

A fresh installation may have no local ingredients. You can add the ingredients you use from their nutrition labels.

1. Open **Nutrition** > **Ingredient overview** and click **Add**.
2. Enter the ingredient name and its energy, protein, carbohydrate, and fat values per 100 g.
3. Complete any other values you have, identify the source, and save.

For a practice entry, use an explicitly fictional name such as `Demo oats (synthetic)`. The demonstration screenshot uses 380 kcal, 12 g protein, 60 g carbohydrates, and 8 g fat per 100 g. Replace these values with the actual label when recording food.

### Record a diary entry

1. Open the nutrition plan overview, add a plan, and give it a name such as `Meal journal`.
2. Open the plan and click **Add nutrition diary entry**.
3. Search for the ingredient you added, select it, and enter the amount and date.
4. Submit the entry and scroll to **Logged** to review the energy and macronutrient totals.

![Logged energy and macronutrients](/images/manual/use-cases/wger-nutrition.jpg)

The plan's scheduled meals and the **Logged** section are separate. Adding a diary entry updates the logged totals without adding a planned meal.

## Use your journal from an AI assistant

MCP gives an assistant tools for working with your wger records. The examples below use the routines, training logs, nutrition diary, and body-weight records from the browser workflow above. You can connect either Lares or Hermes Agent on the same Olares device.

Complete the [Lares setup](lares.md#get-started) or [Hermes Agent setup](hermes.md#step-2-run-the-setup-wizard) first, with a chat model available through [Router](olares-router.md). Hermes can use `default-chat`. In Lares, select the full Router model name corresponding to your default chat model if the selector only lists concrete models.

The model connection and the wger connection are separate. Router supplies the model. The following MCP connection gives the assistant access to your journal.

### Get the MCP connection details

1. In Olares Settings, open **Applications** > **wger** > **Entrances** and locate the **wger MCP** entrance. Copy its endpoint and append `/mcp` if the endpoint does not already include that path. Use this full address as `WGER_MCP_URL` in the examples below. This is a separate entrance from the wger browser page.
2. Open Files and go to **Data** > **wger** > **data** > **state** > **mcp-credentials**. Open `auth.private` and copy only the value of `MCP_STATIC_TOKEN` for use as `WGER_MCP_TOKEN`.

Keep the token in the client's connection settings. Do not paste it into a chat. The same file contains a separate wger API key that the MCP server uses internally. Do not copy that key into the examples below.

:::info Access to your journal
The preview connection uses the wger `admin` account. Anyone using this MCP token can access that account's records and the enabled creation tools. It is not a separate sign-in for each wger user. The MCP entrance uses Olares internal access and Bearer authentication. The examples use assistants running on the same Olares device.
:::

### Connect Lares

1. Open Lares and go to **Settings** > **MCP**.
2. Click **Add server** and enter the connection details:

   | Field | Value |
   | --- | --- |
   | **Server name** | `wger_lab` |
   | **Transport** | `Streamable HTTP` |
   | **MCP URL** | Your `WGER_MCP_URL` |
   | **Headers** | The JSON object below, with your MCP token substituted |

   ```json
   {
     "Authorization": "Bearer WGER_MCP_TOKEN"
   }
   ```

3. Click **Save**. Keep the server enabled. Lares loads the tools after saving. If the server reports a load error, check the MCP URL and token before continuing.
4. Start a conversation and send one of the requests in the next section. When Lares asks for permission, review the requested wger tool and its arguments before approving it.

### Connect Hermes Agent

1. In Files, go to **Data** > **hermesagent** > **home** and make a backup of `config.yaml`.
2. Add the following `wger_lab` entry under `mcp_servers`. Replace `WGER_MCP_URL` and `WGER_MCP_TOKEN` with the values you obtained above. Preserve existing servers and your model settings. If `mcp_servers` already exists, add the entry under it instead of creating a second key.

   ```yaml
   mcp_servers:
     wger_lab:
       url: "WGER_MCP_URL"
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

3. Save the file. Exit the current Hermes terminal chat with `/exit`, then run `hermes` again to load the connection. This example uses [terminal chat](hermes.md#option-1-terminal-chat), so you do not need to enable the Gateway API.
4. Send one request at a time and review the wger tool results in the conversation.

This configuration enables the read tools used below and one write tool, `log_body_weight`. To keep this Hermes connection read-only, omit `log_body_weight` from the list. The preview server also excludes update and delete tools.

### Review training and nutrition

Use either assistant. Substitute your routine name and dates in this request:

```text
Can you review my "Weekly strength journal" in wger for October 27–31,
2026? Tell me how many sets and repetitions I completed each day,
the weights I used, and my total training volume over those days.
Leave out my other routines. I'm just reviewing my journal for now.
```

You do not need to know a tool name or record ID. The assistant finds the routine and chooses the tools. Naming the routine and date range helps it keep other routines out of the summary. Compare the dates and sets with your training log in wger. Do not combine weights recorded in different units without conversion.

For nutrition, select the plan by name and ask for daily totals:

```text
How much did I eat on October 29, 30, and 31, 2026, according to
my "Meal journal" in wger? Show my calories, protein, carbs, and fat
for each day, and add up the three days for me.
```

Use the dates on which you logged food. The assistant selects the nutrition plan and reads the daily records. Open the corresponding plan's **Logged** section in wger to compare the results with the source records.

The totals reflect the food you logged. Missing diary entries do not establish that you ate too little, even if the assistant adds a comment about your goals.

### Save a body-weight entry

Start with a measurement you want to record. Include the date, unit, and a distinct note so the assistant can check for an existing entry before adding one. For example:

```text
I weighed 72.2 kg on October 31, 2026. Please add that to wger with
the note "Journal entry 2026-10-31". If I've already recorded that entry,
don't add it again. Show me the entry when you're done.
```

In Lares, approve the `log_body_weight` call only after checking its date, weight, unit, and note. Then open **Body weight** > **Weight overview** in wger and find the entry in the table. A date without a time is stored at noon in the app's time zone.

If a request appears to stall after a write, ask the assistant to read the history before retrying. The note is a way to find the entry, not a server-enforced uniqueness rule.

### Go beyond a daily summary

Once you have several training sessions and food diary entries, ask the assistant to connect the records and explain what changed. The following examples work with the read tools in the Hermes configuration above. They do not require additional write permissions. Replace the names and dates with your own.

#### Find out where your progress came from

```text
Am I making progress on squats? Look at the last three sessions in my
"Weekly strength journal" in wger, up to October 31, 2026. Compare the dates,
sets, reps, weights, and training volume. How much changed from the first
session to the last? Was it heavier weight, more reps, or more sets?
Keep other routines out of the comparison, and don't estimate my one-rep max.
```

This turns separate log entries into a progression story. Review the session rows before relying on a percentage. With fewer than three matching sessions, ask the assistant to show the available records and identify the gap rather than fill it in.

#### Compare eating records on training and other days

```text
For October 29–31, 2026, line up my "Weekly strength journal" training dates
with my "Meal journal" food diary in wger. Compare days with training records
and days without them. How many days have food entries in each group, and
what are their average daily calories and protein? Don't count an unlogged
day as zero food. I'm looking for patterns in my records, not an explanation
of what caused a weight change.
```

The assistant needs to associate training dates with nutrition entries and calculate averages for each group. A day without a training log only means there is no recorded session. A short or incomplete diary might not support a useful comparison.

#### Make a weekly achievement card

```text
Make me a short weekly achievement card from my wger records for
October 27–31, 2026. Use "Weekly strength journal" and "Meal journal".
Give me up to three small achievements backed by numbers and dates.
Don't award something the records don't support. Then suggest one fun
logging challenge for next week to make my journal more complete.
Keep it about recording habits, not increasing weights or cutting food.
Leave my original records unchanged.
```

The card is a text summary in the assistant's conversation. For example, it might recognize several logged sessions or consistent food entries, with the actual dates beside each achievement. The assistant can propose a challenge, but this request does not schedule reminders or add a challenge feature to wger.

Check the card against the session details. An assistant can calculate the comparison correctly and still miscount a total in its final summary. If something does not add up, follow up naturally:

```text
The total sets on the achievement card don't match the session table.
Please check the original records again and correct the card.
Keep only achievements supported by those records.
```

## Keep using your journal

Account and journal records are stored in PostgreSQL. Uploaded media and application state persist in the app's storage. Keep this data when updating the application.

The browser and connected clients use the same account records. Keep the independent PostgreSQL data, uploaded media, and application state when backing up or updating the application.

## Official resources

- [wger website](https://wger.de/)
- [Official documentation](https://wger.readthedocs.io/en/latest/)
- [Server source code on GitHub](https://github.com/wger-project/wger)
- [Client source code on GitHub](https://github.com/wger-project/flutter)
- [Client downloads on GitHub Releases](https://github.com/wger-project/flutter/releases)
