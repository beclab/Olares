---
outline: [2, 3]
description: Plan a workout routine, record completed sets, and follow your body weight and nutrition in wger on Olares.
head:
  - - meta
    - name: keywords
      content: "Olares, wger, workout journal, body weight, nutrition, self-hosted"
app_version: "0.0.3"
doc_version: "1.0"
doc_updated: "2026-10-30"
---

# Keep a workout and nutrition journal with wger

wger is a fitness journal that brings workout routines, training logs, body measurements, and nutrition into one browser application. On Olares, you can keep your plans and records on your own device and return to them between sessions.

## Learning objectives

- Build a routine with a training day and an exercise.
- Record completed sets and follow your weight over time.
- Add an ingredient and log it in a nutrition diary.

## Install and sign in

1. Open Olares Market and search for **wger**.
2. Click **Get**, then **Install**.
3. Set **ADMIN_PASSWORD** to a password of at least 12 characters. Keep it for your first sign-in.
4. When installation finishes, open wger from Launchpad.
5. Sign in with username `admin` and the password you supplied during installation.

The installation password initializes an empty database. Updating the app keeps existing accounts and records. Public registration and email delivery are disabled in this package.

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

## Keep using your journal

Account and journal records are stored in PostgreSQL. Uploaded media and application state persist in the app's storage. Keep this data when updating the application.

This package supports the browser workflow described here. PowerSync mobile offline synchronization is not included.
